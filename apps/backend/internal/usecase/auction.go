package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type AuctionUseCase struct {
	Auctions    port.AuctionRepository
	Tenants     port.TenantRepository
	Lots        port.LotRepository
	Broadcaster port.Broadcaster
	Audit       port.AuditRepository
}

type CreateAuctionInput struct {
	TenantID                uuid.UUID
	Currency                string
	Title                   string
	Description             string
	Metadata                json.RawMessage
	StartTime               time.Time
	EndTime                 time.Time
	AllowManualStart        bool
	AllowManualPause        bool
	AllowManualEnd          bool
	AutoExtendEnabled       bool
	AuctionMode             domain.AuctionMode
	AutoProgressEnabled     bool
	LotDurationSeconds      int
	SnipingWindowSeconds    int
	SnipingExtensionSeconds int
	BidVisibility           domain.BidVisibility
	RegistrationPolicy      domain.AuctionRegistrationPolicy
	RequireBidderApproval   bool
}

type SaveAuctionWizardInput struct {
	CreateAuctionInput
	WizardStep int
}

func (uc *AuctionUseCase) ListPublic(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Auction, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	return uc.Auctions.ListPublic(ctx, tenantID, now)
}

func (uc *AuctionUseCase) ListAdmin(ctx context.Context, tenantID uuid.UUID) ([]domain.Auction, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	return uc.Auctions.ListAdmin(ctx, tenantID)
}

func (uc *AuctionUseCase) Get(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.Auction, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	return uc.Auctions.GetByID(ctx, tenantID, auctionID)
}

func (uc *AuctionUseCase) Create(ctx context.Context, in CreateAuctionInput) (*domain.Auction, error) {
	if err := ensureTenantScope(ctx, in.TenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionCreate); err != nil {
		return nil, ErrForbidden
	}
	if in.Title == "" || !in.EndTime.After(in.StartTime) {
		return nil, ErrValidation
	}
	tenant, err := uc.Tenants.GetByID(ctx, in.TenantID)
	if err != nil {
		return nil, err
	}
	if tenant.OnboardingStatus != domain.OnboardingCompleted {
		return nil, ErrOnboardingPending
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = strings.ToUpper(strings.TrimSpace(tenant.DefaultCurrency))
	}
	if len(currency) != 3 {
		return nil, ErrValidation
	}
	status := domain.AuctionScheduled
	if in.AllowManualStart {
		status = domain.AuctionDraft
	}
	auction := &domain.Auction{
		ID:                      uuid.New(),
		TenantID:                in.TenantID,
		Currency:                currency,
		Title:                   in.Title,
		Description:             in.Description,
		Metadata:                in.Metadata,
		StartTime:               in.StartTime,
		EndTime:                 in.EndTime,
		Status:                  status,
		AllowManualStart:        in.AllowManualStart,
		AllowManualPause:        in.AllowManualPause,
		AllowManualEnd:          in.AllowManualEnd,
		AutoExtendEnabled:       in.AutoExtendEnabled,
		AuctionMode:             in.AuctionMode,
		AutoProgressEnabled:     in.AutoProgressEnabled,
		LotDurationSeconds:      in.LotDurationSeconds,
		SnipingWindowSeconds:    in.SnipingWindowSeconds,
		SnipingExtensionSeconds: in.SnipingExtensionSeconds,
		BidVisibility:           in.BidVisibility,
		RegistrationPolicy:      normalizeRegistrationPolicy(in.RegistrationPolicy, in.RequireBidderApproval),
		RequireBidderApproval:   in.RequireBidderApproval,
		WizardStep:              5,
		AuctionStatus:           domain.AuctionPublicationPublished,
		PublishedAt:             auctionTimePtr(time.Now().UTC()),
	}
	if len(auction.Metadata) == 0 {
		auction.Metadata = []byte(`{}`)
	}
	auction.RequireBidderApproval = auction.RegistrationPolicy == domain.AuctionRegistrationApprovalRequired
	if auction.AuctionMode == "" {
		auction.AuctionMode = domain.AuctionModeTimed
	}
	if err := uc.Auctions.Create(ctx, auction); err != nil {
		return nil, err
	}
	return auction, nil
}

func (uc *AuctionUseCase) GetWizardDraft(ctx context.Context, tenantID uuid.UUID) (*domain.Auction, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionCreate); err != nil {
		return nil, ErrForbidden
	}
	items, err := uc.Auctions.ListAdmin(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].AuctionStatus == domain.AuctionPublicationDraft {
			return &items[i], nil
		}
	}
	return nil, ErrNotFound
}

func (uc *AuctionUseCase) CreateWizardDraft(ctx context.Context, tenantID uuid.UUID) (*domain.Auction, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionCreate); err != nil {
		return nil, ErrForbidden
	}
	tenant, err := uc.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if tenant.OnboardingStatus != domain.OnboardingCompleted {
		return nil, ErrOnboardingPending
	}
	now := time.Now().UTC()
	currency := strings.ToUpper(strings.TrimSpace(tenant.DefaultCurrency))
	if currency == "" {
		currency = "USD"
	}
	auction := &domain.Auction{
		ID:                      uuid.New(),
		TenantID:                tenantID,
		Currency:                currency,
		Title:                   "",
		Description:             "",
		Metadata:                []byte(`{}`),
		StartTime:               now.Add(10 * time.Minute),
		EndTime:                 now.Add(70 * time.Minute),
		Status:                  domain.AuctionDraft,
		AllowManualStart:        true,
		AllowManualPause:        true,
		AllowManualEnd:          true,
		AutoExtendEnabled:       true,
		AuctionMode:             domain.AuctionModeTimed,
		AutoProgressEnabled:     false,
		LotDurationSeconds:      0,
		SnipingWindowSeconds:    60,
		SnipingExtensionSeconds: 120,
		BidVisibility:           domain.BidVisibilityPublic,
		RegistrationPolicy:      domain.AuctionRegistrationApprovalRequired,
		RequireBidderApproval:   true,
		WizardStep:              1,
		AuctionStatus:           domain.AuctionPublicationDraft,
	}
	if err := uc.Auctions.Create(ctx, auction); err != nil {
		return nil, err
	}
	return auction, nil
}

func (uc *AuctionUseCase) Update(ctx context.Context, in CreateAuctionInput, auctionID uuid.UUID) (*domain.Auction, error) {
	if err := ensureTenantScope(ctx, in.TenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionUpdate); err != nil {
		return nil, ErrForbidden
	}
	if in.Title == "" || !in.EndTime.After(in.StartTime) {
		return nil, ErrValidation
	}
	current, err := uc.Auctions.GetByID(ctx, in.TenantID, auctionID)
	if err != nil {
		return nil, err
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = current.Currency
	}
	if len(currency) != 3 {
		return nil, ErrValidation
	}
	if current.Status == domain.AuctionLive && currency != current.Currency {
		return nil, errors.New("currency cannot be changed while auction is live")
	}
	if current.AuctionStatus == domain.AuctionPublicationPublished && currency != current.Currency {
		return nil, errors.New("financial defaults cannot be changed after publish")
	}
	if current.CurrentPrice > current.StartingPrice && current.StartingPrice > 0 && currency != current.Currency {
		return nil, errors.New("currency cannot be changed after first bid")
	}
	current.Currency = currency
	current.Title = in.Title
	current.Description = in.Description
	current.Metadata = in.Metadata
	current.StartTime = in.StartTime
	current.EndTime = in.EndTime
	current.AllowManualStart = in.AllowManualStart
	current.AllowManualPause = in.AllowManualPause
	current.AllowManualEnd = in.AllowManualEnd
	current.AutoExtendEnabled = in.AutoExtendEnabled
	current.AuctionMode = in.AuctionMode
	current.AutoProgressEnabled = in.AutoProgressEnabled
	current.LotDurationSeconds = in.LotDurationSeconds
	current.SnipingWindowSeconds = in.SnipingWindowSeconds
	current.SnipingExtensionSeconds = in.SnipingExtensionSeconds
	current.BidVisibility = in.BidVisibility
	current.RegistrationPolicy = normalizeRegistrationPolicy(in.RegistrationPolicy, in.RequireBidderApproval)
	current.RequireBidderApproval = current.RegistrationPolicy == domain.AuctionRegistrationApprovalRequired
	if len(current.Metadata) == 0 {
		current.Metadata = []byte(`{}`)
	}
	if current.AuctionMode == "" {
		current.AuctionMode = domain.AuctionModeTimed
	}
	if err := uc.Auctions.Update(ctx, current); err != nil {
		return nil, err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		action := "AUCTION_REQUIRE_APPROVAL_DISABLED"
		if current.RequireBidderApproval {
			action = "AUCTION_REQUIRE_APPROVAL_ENABLED"
		}
		payload, _ := json.Marshal(map[string]any{"require_bidder_approval": current.RequireBidderApproval})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &in.TenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      action,
			EntityType:  "AUCTION",
			EntityID:    &current.ID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return current, nil
}

func (uc *AuctionUseCase) SaveWizard(ctx context.Context, auctionID uuid.UUID, in SaveAuctionWizardInput) (*domain.Auction, error) {
	if in.WizardStep < 1 || in.WizardStep > 4 {
		return nil, ErrValidation
	}
	if err := ensureTenantScope(ctx, in.TenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionCreate); err != nil {
		return nil, ErrForbidden
	}
	tenant, err := uc.Tenants.GetByID(ctx, in.TenantID)
	if err != nil {
		return nil, err
	}
	if tenant.OnboardingStatus != domain.OnboardingCompleted {
		return nil, ErrOnboardingPending
	}
	current, err := uc.Auctions.GetByID(ctx, in.TenantID, auctionID)
	if err != nil {
		return nil, err
	}
	if current.AuctionStatus != domain.AuctionPublicationDraft {
		return nil, errors.New("wizard can only edit draft auctions")
	}
	if err := validateWizardStep(in); err != nil {
		return nil, err
	}
	updated, err := uc.Update(ctx, in.CreateAuctionInput, auctionID)
	if err != nil {
		return nil, err
	}
	if updated.WizardStep < in.WizardStep {
		updated.WizardStep = in.WizardStep
	}
	updated.AuctionStatus = domain.AuctionPublicationDraft
	updated.PublishedAt = nil
	if err := uc.Auctions.Update(ctx, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

func (uc *AuctionUseCase) PublishWizard(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.Auction, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionCreate); err != nil {
		return nil, ErrForbidden
	}
	tenant, err := uc.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if tenant.OnboardingStatus != domain.OnboardingCompleted {
		return nil, ErrOnboardingPending
	}
	current, err := uc.Auctions.GetByID(ctx, tenantID, auctionID)
	if err != nil {
		return nil, err
	}
	if current.AuctionStatus != domain.AuctionPublicationDraft {
		return nil, errors.New("auction is already published")
	}
	if err := validateWizardStep(SaveAuctionWizardInput{
		CreateAuctionInput: CreateAuctionInput{
			TenantID:                current.TenantID,
			Currency:                current.Currency,
			Title:                   current.Title,
			Description:             current.Description,
			Metadata:                current.Metadata,
			StartTime:               current.StartTime,
			EndTime:                 current.EndTime,
			AllowManualStart:        current.AllowManualStart,
			AllowManualPause:        current.AllowManualPause,
			AllowManualEnd:          current.AllowManualEnd,
			AutoExtendEnabled:       current.AutoExtendEnabled,
			AuctionMode:             current.AuctionMode,
			AutoProgressEnabled:     current.AutoProgressEnabled,
			LotDurationSeconds:      current.LotDurationSeconds,
			SnipingWindowSeconds:    current.SnipingWindowSeconds,
			SnipingExtensionSeconds: current.SnipingExtensionSeconds,
			BidVisibility:           current.BidVisibility,
			RequireBidderApproval:   current.RequireBidderApproval,
			RegistrationPolicy:      current.RegistrationPolicy,
		},
		WizardStep: 4,
	}); err != nil {
		return nil, err
	}
	if uc.Lots == nil {
		return nil, errors.New("lot repository unavailable")
	}
	lots, err := uc.Lots.ListByAuction(ctx, tenantID, auctionID)
	if err != nil {
		return nil, err
	}
	if len(lots) == 0 {
		return nil, errors.New("cannot publish without at least one lot")
	}
	current.WizardStep = 5
	current.AuctionStatus = domain.AuctionPublicationPublished
	now := time.Now().UTC()
	current.PublishedAt = &now
	if current.AllowManualStart {
		current.Status = domain.AuctionDraft
	} else {
		current.Status = domain.AuctionScheduled
	}
	if err := uc.Auctions.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (uc *AuctionUseCase) UpdateRequireBidderApproval(ctx context.Context, tenantID, auctionID uuid.UUID, required bool) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionUpdate); err != nil {
		return ErrForbidden
	}
	if err := uc.Auctions.UpdateRequireBidderApproval(ctx, tenantID, auctionID, required); err != nil {
		return err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		action := "AUCTION_REQUIRE_APPROVAL_DISABLED"
		if required {
			action = "AUCTION_REQUIRE_APPROVAL_ENABLED"
		}
		payload, _ := json.Marshal(map[string]any{"require_bidder_approval": required})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      action,
			EntityType:  "AUCTION",
			EntityID:    &auctionID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return nil
}

func normalizeRegistrationPolicy(policy domain.AuctionRegistrationPolicy, requireBidderApproval bool) domain.AuctionRegistrationPolicy {
	switch policy {
	case domain.AuctionRegistrationOpen, domain.AuctionRegistrationApprovalRequired:
		return policy
	default:
		if requireBidderApproval {
			return domain.AuctionRegistrationApprovalRequired
		}
		return domain.AuctionRegistrationOpen
	}
}

func (uc *AuctionUseCase) Delete(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionUpdate); err != nil {
		return ErrForbidden
	}
	return uc.Auctions.Delete(ctx, tenantID, auctionID)
}

func (uc *AuctionUseCase) Start(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionStart); err != nil {
		return ErrForbidden
	}
	tenant, err := uc.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	if tenant.OnboardingStatus != domain.OnboardingCompleted {
		return ErrOnboardingPending
	}
	if err := uc.Auctions.UpdateStatus(ctx, tenantID, auctionID, domain.AuctionLive); err != nil {
		return err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"status": domain.AuctionLive})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "AUCTION_STARTED",
			EntityType:  "AUCTION",
			EntityID:    &auctionID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	uc.Broadcaster.BroadcastAuctionStatus(ctx, auctionID, domain.AuctionLive)
	return nil
}

func (uc *AuctionUseCase) Pause(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionEnd); err != nil {
		return ErrForbidden
	}
	if err := uc.Auctions.UpdateStatus(ctx, tenantID, auctionID, domain.AuctionPaused); err != nil {
		return err
	}
	uc.Broadcaster.BroadcastAuctionStatus(ctx, auctionID, domain.AuctionPaused)
	return nil
}

func (uc *AuctionUseCase) End(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionEnd); err != nil {
		return ErrForbidden
	}
	if err := uc.Auctions.UpdateStatus(ctx, tenantID, auctionID, domain.AuctionEnded); err != nil {
		return err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"status": domain.AuctionEnded})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "AUCTION_ENDED",
			EntityType:  "AUCTION",
			EntityID:    &auctionID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	uc.Broadcaster.BroadcastAuctionStatus(ctx, auctionID, domain.AuctionEnded)
	return nil
}

func (uc *AuctionUseCase) SchedulerTick(ctx context.Context, now time.Time) error {
	return uc.Auctions.RunSchedulerTick(ctx, now)
}

func validateWizardStep(in SaveAuctionWizardInput) error {
	switch in.WizardStep {
	case 1:
		if strings.TrimSpace(in.Title) == "" {
			return ErrValidation
		}
	case 2:
		if in.StartTime.IsZero() || !in.EndTime.After(in.StartTime) {
			return ErrValidation
		}
	case 3:
		if len(strings.TrimSpace(in.Currency)) != 3 {
			return ErrValidation
		}
	case 4:
		if in.AuctionMode == domain.AuctionModeSequential && in.AutoProgressEnabled && in.LotDurationSeconds <= 0 {
			return ErrValidation
		}
		if in.SnipingWindowSeconds < 0 || in.SnipingExtensionSeconds < 0 {
			return ErrValidation
		}
	default:
		return ErrValidation
	}
	return nil
}

func auctionTimePtr(v time.Time) *time.Time { return &v }
