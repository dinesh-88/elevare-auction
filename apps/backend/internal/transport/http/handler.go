package httptransport

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/dinesh-88/elevare-auction/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	Auth                *usecase.AuthUseCase
	Auction             *usecase.AuctionUseCase
	Lot                 *usecase.LotUseCase
	Fee                 *usecase.FeeUseCase
	Invoice             *usecase.InvoiceUseCase
	Bidder              *usecase.BidderUseCase
	Audit               *usecase.AuditUseCase
	Notification        *usecase.NotificationUseCase
	Site                *usecase.SiteUseCase
	Asset               *usecase.AssetUseCase
	Onboarding          *usecase.OnboardingUseCase
	DemoRequest         *usecase.DemoRequestUseCase
	Storage             port.ObjectStorage
	PreviewToken        string
	JWTTTL              time.Duration
	CookieSecure        bool
	AccessCookie        string
	CSRFCookie          string
	TenantAppScheme     string
	TenantAppBaseDomain string
	TenantAppPort       string
}

type lotPublicResponse struct {
	ID            uuid.UUID        `json:"id"`
	LotID         uuid.UUID        `json:"lot_id"`
	AuctionID     uuid.UUID        `json:"auction_id"`
	LotNumber     string           `json:"lot_number"`
	OrderIndex    int              `json:"order_index"`
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	StartingPrice int64            `json:"starting_price"`
	StartingMinor int64            `json:"starting_price_minor"`
	CurrentPrice  int64            `json:"current_price"`
	CurrentMinor  int64            `json:"current_price_minor"`
	MinimumNext   int64            `json:"minimum_next_bid_minor"`
	ReservePrice  *int64           `json:"reserve_price"`
	MinIncrement  int64            `json:"min_increment"`
	Increment     int64            `json:"increment_minor"`
	Metadata      json.RawMessage  `json:"metadata"`
	Status        domain.LotStatus `json:"status"`
	HammerPrice   *int64           `json:"hammer_price"`
	SaleResult    string           `json:"sale_result"`
	StartedAt     *time.Time       `json:"started_at"`
	EndedAt       *time.Time       `json:"ended_at"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

type authUserResponse struct {
	ID             uuid.UUID             `json:"id"`
	TenantID       uuid.UUID             `json:"tenant_id"`
	Email          string                `json:"email"`
	Role           domain.Role           `json:"role"`
	Status         domain.UserStatus     `json:"status"`
	EmailVerified  bool                  `json:"email_verified"`
	ApprovalStatus domain.ApprovalStatus `json:"approval_status"`
	KYCStatus      domain.KYCStatus      `json:"kyc_status"`
	CreatedAt      time.Time             `json:"created_at"`
}

type bidHistoryPublicResponse struct {
	ID            uuid.UUID `json:"id"`
	AmountMinor   int64     `json:"amount_minor"`
	CreatedAt     time.Time `json:"created_at"`
	BidderDisplay string    `json:"bidder_display"`
}

func (h *Handler) PublicTenantSignup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrganizationName       string `json:"organization_name"`
		DesiredSubdomain       string `json:"desired_subdomain"`
		AdminEmail             string `json:"admin_email"`
		Password               string `json:"password"`
		Country                string `json:"country"`
		EstimatedAuctionVolume *int   `json:"estimated_auction_volume"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	tenant, admin, err := h.Onboarding.PublicSignup(r.Context(), usecase.TenantSignupInput{
		OrganizationName:       req.OrganizationName,
		DesiredSubdomain:       req.DesiredSubdomain,
		AdminEmail:             req.AdminEmail,
		Password:               req.Password,
		Country:                req.Country,
		EstimatedAuctionVolume: req.EstimatedAuctionVolume,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"tenant_id":             tenant.ID,
		"tenant_status":         tenant.Status,
		"admin_user_id":         admin.ID,
		"admin_approval_status": admin.ApprovalStatus,
	})
}

func (h *Handler) SubmitDemoRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Institution string `json:"institution"`
		Email       string `json:"email"`
		Message     string `json:"message"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if h.DemoRequest == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	item, err := h.DemoRequest.Submit(r.Context(), usecase.SubmitDemoRequestInput{
		Name:        req.Name,
		Institution: req.Institution,
		Email:       req.Email,
		Message:     req.Message,
	})
	if err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, usecase.ErrValidation) {
			status = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"status":     "RECEIVED",
		"id":         item.ID,
		"created_at": item.CreatedAt,
	})
}

func (h *Handler) buildTenantLoginURL(match usecase.TenantLoginMatch) string {
	scheme := strings.TrimSpace(h.TenantAppScheme)
	if scheme == "" {
		scheme = "https"
	}
	baseDomain := strings.TrimSpace(h.TenantAppBaseDomain)
	if baseDomain == "" {
		baseDomain = "elevareauctions.com"
	}
	portSuffix := ""
	if p := strings.TrimSpace(h.TenantAppPort); p != "" {
		portSuffix = ":" + p
	}
	host := ""
	if match.CustomDomain != nil && strings.TrimSpace(*match.CustomDomain) != "" {
		host = strings.TrimSpace(*match.CustomDomain)
	} else {
		host = fmt.Sprintf("%s.%s", strings.TrimSpace(match.Subdomain), baseDomain)
	}
	return fmt.Sprintf("%s://%s%s/login", scheme, host, portSuffix)
}

func (h *Handler) DiscoverTenantLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	matches, err := h.Auth.DiscoverTenantsByEmail(r.Context(), req.Email)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, usecase.ErrValidation) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	items := make([]map[string]any, 0, len(matches))
	for _, m := range matches {
		host := ""
		if m.CustomDomain != nil && strings.TrimSpace(*m.CustomDomain) != "" {
			host = strings.TrimSpace(*m.CustomDomain)
		} else {
			host = fmt.Sprintf("%s.%s", strings.TrimSpace(m.Subdomain), strings.TrimSpace(h.TenantAppBaseDomain))
		}
		items = append(items, map[string]any{
			"tenant_id":   m.TenantID,
			"tenant_name": m.TenantName,
			"host":        host,
			"login_url":   h.buildTenantLoginURL(m),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ListTenantRequests(w http.ResponseWriter, r *http.Request) {
	status := domain.TenantStatus(r.URL.Query().Get("status"))
	tenants, err := h.Onboarding.ListTenantRequests(r.Context(), status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, tenants)
}

func (h *Handler) ApproveTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid tenant id", http.StatusBadRequest)
		return
	}
	if err := h.Onboarding.ApproveTenant(r.Context(), tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ACTIVE"})
}

func (h *Handler) RejectTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid tenant id", http.StatusBadRequest)
		return
	}
	if err := h.Onboarding.RejectTenant(r.Context(), tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "REJECTED"})
}

func (h *Handler) PublicConfig(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.TenantFromContext(r.Context())
	if !ok {
		http.Error(w, "tenant missing", http.StatusInternalServerError)
		return
	}
	preview := strings.TrimSpace(r.URL.Query().Get("previewToken"))
	branding := tenant.BrandingPublished
	content := tenant.ContentPublished
	if preview != "" && h.PreviewToken != "" && preview == h.PreviewToken {
		branding = tenant.BrandingDraft
		content = tenant.ContentDraft
		w.Header().Set("Cache-Control", "no-store")
	}
	if len(branding) == 0 {
		branding = tenant.Branding
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name":                   tenant.Name,
		"branding":               branding,
		"content":                content,
		"content_version":        tenant.ContentVersion,
		"content_schema_version": tenant.ContentSchemaVersion,
		"default_currency":       tenant.DefaultCurrency,
		"require_admin_approval": tenant.RequireAdminApproval,
		"require_kyc":            tenant.RequireKYC,
	})
}

func (h *Handler) ListAuctions(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctions, err := h.Auction.ListPublic(r.Context(), tenant.ID, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, auctions)
}

func (h *Handler) GetAuction(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	auction, err := h.Auction.Get(r.Context(), tenant.ID, auctionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, auction)
}

func (h *Handler) ListLotsByAuction(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	lots, err := h.Lot.ListByAuction(r.Context(), tenant.ID, auctionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := make([]lotPublicResponse, 0, len(lots))
	for i := range lots {
		resp = append(resp, toLotPublicResponse(lots[i]))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetLot(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}
	lot, err := h.Lot.GetByID(r.Context(), tenant.ID, lotID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, toLotPublicResponse(*lot))
}

func (h *Handler) GetLotFinancials(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "claims missing", http.StatusUnauthorized)
		return
	}
	if claims.Role != domain.RoleBidder {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}
	lot, err := h.Lot.GetByID(r.Context(), tenant.ID, lotID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	isAdmin := domain.HasPermission(claims.Role, domain.PermissionInvoiceView)
	isWinner := lot.WinnerUserID != nil && claims.UserID == *lot.WinnerUserID
	if !isAdmin && !isWinner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var snapshot domain.LotFinancialSnapshot
	if len(lot.FinancialData) > 0 {
		if err := json.Unmarshal(lot.FinancialData, &snapshot); err != nil {
			http.Error(w, "invalid financial snapshot", http.StatusInternalServerError)
			return
		}
	}
	var invoiceID *uuid.UUID
	if inv, err := h.Invoice.GetByLotAuthorized(r.Context(), tenant.ID, lot.ID, claims.UserID, claims.Role); err == nil {
		invoiceID = &inv.ID
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"lot_id":       lot.ID,
		"winner_user":  lot.WinnerUserID,
		"hammer_price": lot.HammerPrice,
		"fees":         snapshot.Fees,
		"total_due":    lot.TotalDue,
		"invoice_id":   invoiceID,
	})
}

func (h *Handler) ListLotBids(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}
	bids, err := h.Lot.ListBidsByLot(r.Context(), tenant.ID, lotID, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, toPublicBidHistory(bids))
}

func toPublicBidHistory(bids []domain.Bid) []bidHistoryPublicResponse {
	out := make([]bidHistoryPublicResponse, 0, len(bids))
	for _, bid := range bids {
		out = append(out, bidHistoryPublicResponse{
			ID:            bid.ID,
			AmountMinor:   bid.Amount,
			CreatedAt:     bid.CreatedAt,
			BidderDisplay: bidderAlias(bid.UserID),
		})
	}
	return out
}

func bidderAlias(userID uuid.UUID) string {
	return "Bidder " + strconv.Itoa(int(crc32.ChecksumIEEE(userID[:])%900)+100)
}

func (h *Handler) ListMyWinningLots(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "claims missing", http.StatusUnauthorized)
		return
	}
	lots, err := h.Lot.ListWonByUser(r.Context(), tenant.ID, claims.UserID, 200)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := make([]lotPublicResponse, 0, len(lots))
	for i := range lots {
		resp = append(resp, toLotPublicResponse(lots[i]))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	user, err := h.Auth.Register(r.Context(), usecase.RegisterInput{
		Tenant:   tenant,
		Email:    req.Email,
		Password: req.Password,
		Role:     domain.RoleBidder,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user_id": user.ID, "approval_status": user.ApprovalStatus})
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}
	if err := h.Auth.VerifyEmail(r.Context(), tenant.ID, uid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "verified"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	token, user, err := h.Auth.Login(r.Context(), usecase.LoginInput{TenantID: tenant.ID, Email: req.Email, Password: req.Password})
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, usecase.ErrValidation) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	h.setSessionCookies(w, token)
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": toAuthUserResponse(user)})
}

func (h *Handler) Logout(w http.ResponseWriter, _ *http.Request) {
	h.clearSessionCookies(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *Handler) Session(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "claims missing", http.StatusUnauthorized)
		return
	}
	user, err := h.Auth.CurrentUser(r.Context(), tenant.ID, claims.UserID)
	if err != nil {
		http.Error(w, "session not found", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": toAuthUserResponse(user)})
}

func (h *Handler) CreateAuction(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	var req struct {
		Currency                string               `json:"currency"`
		Title                   string               `json:"title"`
		Description             string               `json:"description"`
		Metadata                json.RawMessage      `json:"metadata"`
		StartTime               time.Time            `json:"start_time"`
		EndTime                 time.Time            `json:"end_time"`
		AllowManualStart        bool                 `json:"allow_manual_start"`
		AllowManualPause        bool                 `json:"allow_manual_pause"`
		AllowManualEnd          bool                 `json:"allow_manual_end"`
		AutoExtendEnabled       bool                 `json:"auto_extend_enabled"`
		AuctionMode             domain.AuctionMode   `json:"auction_mode"`
		AutoProgressEnabled     bool                 `json:"auto_progress_enabled"`
		LotDurationSeconds      int                  `json:"lot_duration_seconds"`
		SnipingWindowSeconds    int                  `json:"sniping_window_seconds"`
		SnipingExtensionSeconds int                  `json:"sniping_extension_seconds"`
		BidVisibility           domain.BidVisibility `json:"bid_visibility"`
		RequireBidderApproval   bool                 `json:"require_bidder_approval"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	a, err := h.Auction.Create(r.Context(), usecase.CreateAuctionInput{
		TenantID:                tenant.ID,
		Currency:                req.Currency,
		Title:                   req.Title,
		Description:             req.Description,
		Metadata:                req.Metadata,
		StartTime:               req.StartTime,
		EndTime:                 req.EndTime,
		AllowManualStart:        req.AllowManualStart,
		AllowManualPause:        req.AllowManualPause,
		AllowManualEnd:          req.AllowManualEnd,
		AutoExtendEnabled:       req.AutoExtendEnabled,
		AuctionMode:             req.AuctionMode,
		AutoProgressEnabled:     req.AutoProgressEnabled,
		LotDurationSeconds:      req.LotDurationSeconds,
		SnipingWindowSeconds:    req.SnipingWindowSeconds,
		SnipingExtensionSeconds: req.SnipingExtensionSeconds,
		BidVisibility:           req.BidVisibility,
		RequireBidderApproval:   req.RequireBidderApproval,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (h *Handler) CreateLot(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	var req struct {
		LotNumber     string          `json:"lot_number"`
		OrderIndex    int             `json:"order_index"`
		Title         string          `json:"title"`
		Description   string          `json:"description"`
		Metadata      json.RawMessage `json:"metadata"`
		StartingPrice int64           `json:"starting_price"`
		ReservePrice  *int64          `json:"reserve_price"`
		MinIncrement  int64           `json:"min_increment"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	lot, err := h.Lot.Create(r.Context(), usecase.CreateLotInput{
		TenantID:      tenant.ID,
		AuctionID:     auctionID,
		LotNumber:     req.LotNumber,
		OrderIndex:    req.OrderIndex,
		Title:         req.Title,
		Description:   req.Description,
		Metadata:      req.Metadata,
		StartingPrice: req.StartingPrice,
		ReservePrice:  req.ReservePrice,
		MinIncrement:  req.MinIncrement,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, lot)
}

func (h *Handler) UpdateLot(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}
	var req struct {
		LotNumber     string          `json:"lot_number"`
		OrderIndex    int             `json:"order_index"`
		Title         string          `json:"title"`
		Description   string          `json:"description"`
		Metadata      json.RawMessage `json:"metadata"`
		StartingPrice int64           `json:"starting_price"`
		ReservePrice  *int64          `json:"reserve_price"`
		MinIncrement  int64           `json:"min_increment"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	lot, err := h.Lot.Update(r.Context(), tenant.ID, lotID, usecase.CreateLotInput{
		TenantID:      tenant.ID,
		LotNumber:     req.LotNumber,
		OrderIndex:    req.OrderIndex,
		Title:         req.Title,
		Description:   req.Description,
		Metadata:      req.Metadata,
		StartingPrice: req.StartingPrice,
		ReservePrice:  req.ReservePrice,
		MinIncrement:  req.MinIncrement,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, lot)
}

func (h *Handler) DeleteLot(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}
	if err := h.Lot.Delete(r.Context(), tenant.ID, lotID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "DELETED"})
}

func (h *Handler) ReorderLots(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	var req struct {
		LotIDs []string `json:"lot_ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	ids := make([]uuid.UUID, 0, len(req.LotIDs))
	for _, raw := range req.LotIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			http.Error(w, "invalid lot id in list", http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}
	if err := h.Lot.Reorder(r.Context(), tenant.ID, auctionID, ids); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "REORDERED"})
}

func (h *Handler) StartLot(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}
	if err := h.Lot.Start(r.Context(), tenant.ID, lotID, time.Now().UTC()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "LIVE"})
}

func (h *Handler) EndLot(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}
	if err := h.Lot.End(r.Context(), tenant.ID, lotID, time.Now().UTC()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ENDED"})
}

func (h *Handler) ListAdminAuctions(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctions, err := h.Auction.ListAdmin(r.Context(), tenant.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, auctions)
}

func (h *Handler) UpdateAuction(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req struct {
		Currency                string               `json:"currency"`
		Title                   string               `json:"title"`
		Description             string               `json:"description"`
		Metadata                json.RawMessage      `json:"metadata"`
		StartTime               time.Time            `json:"start_time"`
		EndTime                 time.Time            `json:"end_time"`
		AllowManualStart        bool                 `json:"allow_manual_start"`
		AllowManualPause        bool                 `json:"allow_manual_pause"`
		AllowManualEnd          bool                 `json:"allow_manual_end"`
		AutoExtendEnabled       bool                 `json:"auto_extend_enabled"`
		AuctionMode             domain.AuctionMode   `json:"auction_mode"`
		AutoProgressEnabled     bool                 `json:"auto_progress_enabled"`
		LotDurationSeconds      int                  `json:"lot_duration_seconds"`
		SnipingWindowSeconds    int                  `json:"sniping_window_seconds"`
		SnipingExtensionSeconds int                  `json:"sniping_extension_seconds"`
		BidVisibility           domain.BidVisibility `json:"bid_visibility"`
		RequireBidderApproval   *bool                `json:"require_bidder_approval"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	requireBidderApproval := false
	if req.RequireBidderApproval != nil {
		requireBidderApproval = *req.RequireBidderApproval
	} else {
		current, getErr := h.Auction.Get(r.Context(), tenant.ID, auctionID)
		if getErr == nil {
			requireBidderApproval = current.RequireBidderApproval
		}
	}
	a, err := h.Auction.Update(r.Context(), usecase.CreateAuctionInput{
		TenantID:                tenant.ID,
		Currency:                req.Currency,
		Title:                   req.Title,
		Description:             req.Description,
		Metadata:                req.Metadata,
		StartTime:               req.StartTime,
		EndTime:                 req.EndTime,
		AllowManualStart:        req.AllowManualStart,
		AllowManualPause:        req.AllowManualPause,
		AllowManualEnd:          req.AllowManualEnd,
		AutoExtendEnabled:       req.AutoExtendEnabled,
		AuctionMode:             req.AuctionMode,
		AutoProgressEnabled:     req.AutoProgressEnabled,
		LotDurationSeconds:      req.LotDurationSeconds,
		SnipingWindowSeconds:    req.SnipingWindowSeconds,
		SnipingExtensionSeconds: req.SnipingExtensionSeconds,
		BidVisibility:           req.BidVisibility,
		RequireBidderApproval:   requireBidderApproval,
	}, auctionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) DeleteAuction(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.Auction.Delete(r.Context(), tenant.ID, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "DELETED"})
}

func (h *Handler) StartAuction(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.Auction.Start(r.Context(), tenant.ID, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = h.Lot.OnAuctionStart(r.Context(), tenant.ID, id, time.Now().UTC())
	writeJSON(w, http.StatusOK, map[string]string{"status": "LIVE"})
}

func (h *Handler) PauseAuction(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.Auction.Pause(r.Context(), tenant.ID, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "PAUSED"})
}

func (h *Handler) EndAuction(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.Auction.End(r.Context(), tenant.ID, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ENDED"})
}

func (h *Handler) ListBidders(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	users, err := h.Bidder.List(r.Context(), tenant.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) ListAuctionBidders(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	var statusFilter *domain.ApprovalStatus
	if raw := strings.TrimSpace(r.URL.Query().Get("status")); raw != "" {
		status := domain.ApprovalStatus(strings.ToUpper(raw))
		switch status {
		case domain.ApprovalPending, domain.ApprovalApproved, domain.ApprovalRejected:
			statusFilter = &status
		default:
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
	}
	rows, err := h.Bidder.ListAuctionBidders(r.Context(), tenant.ID, auctionID, statusFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (h *Handler) ApproveAuctionBidder(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, _ := middleware.ClaimsFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	item, err := h.Bidder.ApproveAuctionBidder(r.Context(), tenant.ID, auctionID, userID, claims.UserID, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) RejectAuctionBidder(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, _ := middleware.ClaimsFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	item, err := h.Bidder.RejectAuctionBidder(r.Context(), tenant.ID, auctionID, userID, claims.UserID, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) ApproveBidder(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.Bidder.Approve(r.Context(), tenant.ID, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "APPROVED"})
}

func (h *Handler) RejectBidder(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.Bidder.Reject(r.Context(), tenant.ID, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "REJECTED"})
}

func (h *Handler) GetBidderProfile(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	item, err := h.Bidder.GetProfile(r.Context(), tenant.ID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) UpdateBidderProfile(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	var req struct {
		Status        *domain.TenantBidderStatus    `json:"status"`
		RiskLevel     *domain.TenantBidderRiskLevel `json:"risk_level"`
		InternalNotes *string                       `json:"internal_notes"`
		Phone         *string                       `json:"phone"`
		CompanyName   *string                       `json:"company_name"`
		AddressLine1  *string                       `json:"address_line1"`
		AddressLine2  *string                       `json:"address_line2"`
		City          *string                       `json:"city"`
		PostalCode    *string                       `json:"postal_code"`
		Country       *string                       `json:"country"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	input := port.UpdateTenantBidderInput{
		InternalNotes: req.InternalNotes,
		Phone:         req.Phone,
		CompanyName:   req.CompanyName,
		AddressLine1:  req.AddressLine1,
		AddressLine2:  req.AddressLine2,
		City:          req.City,
		PostalCode:    req.PostalCode,
		Country:       req.Country,
	}
	if req.Status != nil {
		input.Status = *req.Status
	}
	if req.RiskLevel != nil {
		input.RiskLevel = *req.RiskLevel
	}
	item, err := h.Bidder.UpdateProfile(r.Context(), tenant.ID, userID, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) UpdateAuctionBidderSettings(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	var req struct {
		RequireBidderApproval bool `json:"require_bidder_approval"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := h.Auction.UpdateRequireBidderApproval(r.Context(), tenant.ID, auctionID, req.RequireBidderApproval); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"auction_id":              auctionID,
		"require_bidder_approval": req.RequireBidderApproval,
	})
}

func (h *Handler) GetTenantSettings(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	current, err := h.Onboarding.Tenants.GetByID(r.Context(), tenant.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"default_currency": current.DefaultCurrency,
	})
}

func (h *Handler) UpdateTenantSettings(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	var req struct {
		DefaultCurrency string `json:"default_currency"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	currency := strings.ToUpper(strings.TrimSpace(req.DefaultCurrency))
	if len(currency) != 3 {
		http.Error(w, "invalid currency", http.StatusBadRequest)
		return
	}
	if err := h.Onboarding.Tenants.UpdateDefaultCurrency(r.Context(), tenant.ID, currency); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"default_currency": currency})
}

func (h *Handler) GetSite(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	item, err := h.Site.Get(r.Context(), tenant.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"branding_published":     item.BrandingPublished,
		"content_published":      item.ContentPublished,
		"branding_draft":         item.BrandingDraft,
		"content_draft":          item.ContentDraft,
		"content_version":        item.ContentVersion,
		"content_schema_version": item.ContentSchemaVersion,
	})
}

func (h *Handler) UpdateSiteDraft(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	var req struct {
		Branding json.RawMessage `json:"branding"`
		Content  json.RawMessage `json:"content"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := h.Site.UpdateDraft(r.Context(), tenant.ID, usecase.SiteDraftInput{
		Branding: req.Branding,
		Content:  req.Content,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "DRAFT_SAVED"})
}

func (h *Handler) PublishSite(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	beforeVersion, afterVersion, err := h.Site.Publish(r.Context(), tenant.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":         "PUBLISHED",
		"before_version": beforeVersion,
		"after_version":  afterVersion,
	})
}

func (h *Handler) PresignAsset(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"mode":            "DIRECT_SERVER_UPLOAD",
		"upload_endpoint": "/admin/assets",
		"max_size_bytes":  10 * 1024 * 1024,
	})
}

func (h *Handler) RegisterAsset(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	var req struct {
		Kind      string `json:"kind"`
		URL       string `json:"url"`
		Width     *int   `json:"width"`
		Height    *int   `json:"height"`
		SizeBytes *int64 `json:"size_bytes"`
		MIME      string `json:"mime"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	item, err := h.Asset.Register(r.Context(), usecase.RegisterAssetInput{
		TenantID:  tenant.ID,
		Kind:      req.Kind,
		URL:       req.URL,
		Width:     req.Width,
		Height:    req.Height,
		SizeBytes: req.SizeBytes,
		MIME:      req.MIME,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) ListAssets(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	limit := parsePositiveIntOrDefault(r.URL.Query().Get("limit"), 100)
	items, err := h.Asset.List(r.Context(), tenant.ID, kind, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) PlaceLotBid(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "claims missing", http.StatusUnauthorized)
		return
	}
	if claims.Role != domain.RoleBidder {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}
	var req struct {
		Amount         *int64 `json:"amount"`
		BidAmountMinor *int64 `json:"bid_amount_minor"`
		MaxBidMinor    *int64 `json:"max_bid_minor"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	bidAmount := req.BidAmountMinor
	if bidAmount == nil {
		bidAmount = req.Amount
	}
	res, err := h.Lot.PlaceBid(r.Context(), port.PlaceLotBidTxInput{
		TenantID: tenant.ID,
		LotID:    lotID,
		UserID:   claims.UserID,
		Amount:   bidAmount,
		MaxBid:   req.MaxBidMinor,
		Now:      time.Now().UTC(),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrApprovalRequired),
			errors.Is(err, usecase.ErrPendingApproval),
			errors.Is(err, usecase.ErrNotApproved),
			errors.Is(err, usecase.ErrBidderSuspended):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"lot_id":        res.Lot.ID,
		"current_price": res.Lot.CurrentPrice,
		"ended_at":      res.Lot.EndedAt,
		"bid": map[string]any{
			"id":           res.Bid.ID,
			"amount_minor": res.Bid.Amount,
			"created_at":   res.Bid.CreatedAt,
		},
	})
}

func (h *Handler) RequestToBid(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "claims missing", http.StatusUnauthorized)
		return
	}
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	item, err := h.Bidder.RequestToBid(r.Context(), tenant.ID, auctionID, claims.UserID, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"auction_id": auctionID,
		"user_id":    claims.UserID,
		"status":     item.Status,
	})
}

func (h *Handler) ListFeeComponents(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	activeOnly := r.URL.Query().Get("active_only") != "false"
	items, err := h.Fee.ListComponents(r.Context(), tenant.ID, activeOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) CreateFeeComponent(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	var req struct {
		Name            string                    `json:"name"`
		CalculationType domain.FeeCalculationType `json:"calculation_type"`
		Value           float64                   `json:"value"`
		AppliesTo       domain.FeeAppliesTo       `json:"applies_to"`
		Taxable         bool                      `json:"taxable"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	item, err := h.Fee.CreateComponent(r.Context(), usecase.CreateFeeComponentInput{
		TenantID:        tenant.ID,
		Name:            req.Name,
		CalculationType: req.CalculationType,
		Value:           req.Value,
		AppliesTo:       req.AppliesTo,
		Taxable:         req.Taxable,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) UpdateFeeComponent(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	componentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid fee component id", http.StatusBadRequest)
		return
	}
	var req struct {
		Name            string                    `json:"name"`
		CalculationType domain.FeeCalculationType `json:"calculation_type"`
		Value           float64                   `json:"value"`
		AppliesTo       domain.FeeAppliesTo       `json:"applies_to"`
		Taxable         bool                      `json:"taxable"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	item, err := h.Fee.UpdateComponent(r.Context(), tenant.ID, componentID, usecase.CreateFeeComponentInput{
		TenantID:        tenant.ID,
		Name:            req.Name,
		CalculationType: req.CalculationType,
		Value:           req.Value,
		AppliesTo:       req.AppliesTo,
		Taxable:         req.Taxable,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteFeeComponent(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	componentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid fee component id", http.StatusBadRequest)
		return
	}
	if err := h.Fee.DeactivateComponent(r.Context(), tenant.ID, componentID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "DEACTIVATED"})
}

func (h *Handler) ListFeeAssignments(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	scopeType := domain.FeeScopeType(r.URL.Query().Get("scope_type"))
	scopeID, err := uuid.Parse(r.URL.Query().Get("scope_id"))
	if err != nil {
		http.Error(w, "invalid scope_id", http.StatusBadRequest)
		return
	}
	items, err := h.Fee.ListAssignments(r.Context(), tenant.ID, scopeType, scopeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) CreateFeeAssignment(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	var req struct {
		ScopeType      domain.FeeScopeType `json:"scope_type"`
		ScopeID        string              `json:"scope_id"`
		FeeComponentID string              `json:"fee_component_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	scopeID, err := uuid.Parse(req.ScopeID)
	if err != nil {
		http.Error(w, "invalid scope_id", http.StatusBadRequest)
		return
	}
	feeComponentID, err := uuid.Parse(req.FeeComponentID)
	if err != nil {
		http.Error(w, "invalid fee_component_id", http.StatusBadRequest)
		return
	}
	item, err := h.Fee.CreateAssignment(r.Context(), usecase.CreateFeeAssignmentInput{
		TenantID:       tenant.ID,
		ScopeType:      req.ScopeType,
		ScopeID:        scopeID,
		FeeComponentID: feeComponentID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) DeleteFeeAssignment(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	assignmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}
	if err := h.Fee.DeleteAssignment(r.Context(), tenant.ID, assignmentID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "DELETED"})
}

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	limit := parsePositiveIntOrDefault(r.URL.Query().Get("limit"), 200)
	logs, err := h.Audit.ListByTenant(r.Context(), tenant.ID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) ListAuditLogsByEntity(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	entityType := strings.TrimSpace(r.URL.Query().Get("type"))
	entityID, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	limit := parsePositiveIntOrDefault(r.URL.Query().Get("limit"), 200)
	logs, err := h.Audit.ListByEntity(r.Context(), tenant.ID, entityType, entityID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) ListLotAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}

	limit := parsePositiveIntOrDefault(r.URL.Query().Get("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	cursorCreatedAt, cursorID, err := decodeAuditCursor(strings.TrimSpace(r.URL.Query().Get("cursor")))
	if err != nil {
		http.Error(w, "invalid cursor", http.StatusBadRequest)
		return
	}

	var actorUserID *uuid.UUID
	if rawActor := strings.TrimSpace(r.URL.Query().Get("actor_user_id")); rawActor != "" {
		id, parseErr := uuid.Parse(rawActor)
		if parseErr != nil {
			http.Error(w, "invalid actor_user_id", http.StatusBadRequest)
			return
		}
		actorUserID = &id
	}

	from, err := parseISODateTime(strings.TrimSpace(r.URL.Query().Get("from")), false)
	if err != nil {
		http.Error(w, "invalid from", http.StatusBadRequest)
		return
	}
	to, err := parseISODateTime(strings.TrimSpace(r.URL.Query().Get("to")), true)
	if err != nil {
		http.Error(w, "invalid to", http.StatusBadRequest)
		return
	}

	result, err := h.Audit.ListByEntityPage(r.Context(), tenant.ID, "LOT", lotID, usecase.AuditListInput{
		Limit:           limit,
		CursorCreatedAt: cursorCreatedAt,
		CursorID:        cursorID,
		EventType:       strings.TrimSpace(r.URL.Query().Get("event_type")),
		ActorUserID:     actorUserID,
		From:            from,
		To:              to,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrForbidden):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, usecase.ErrUnauthorized):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case errors.Is(err, usecase.ErrValidation):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, row := range result.Items {
		var actorID *string
		var entityID *string
		if row.ActorUserID != nil {
			id := row.ActorUserID.String()
			actorID = &id
		}
		if row.EntityID != nil {
			id := row.EntityID.String()
			entityID = &id
		}
		items = append(items, map[string]any{
			"id":          row.ID,
			"created_at":  row.CreatedAt,
			"event_type":  row.Action,
			"entity_type": row.EntityType,
			"entity_id":   entityID,
			"actor": map[string]any{
				"user_id":      actorID,
				"display_name": auditActorDisplay(row.ActorRole, row.ActorUserID),
				"role":         auditActorRole(row.ActorRole, row.ActorUserID),
			},
			"summary":  auditSummary(row.Action, row.Metadata),
			"metadata": parseAuditMetadata(row.Metadata),
		})
	}

	var nextCursor any = nil
	if result.HasMore && len(result.Items) > 0 {
		nextCursor = encodeAuditCursor(result.Items[len(result.Items)-1].CreatedAt, result.Items[len(result.Items)-1].ID)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":       items,
		"next_cursor": nextCursor,
	})
}

func toLotPublicResponse(l domain.Lot) lotPublicResponse {
	saleResult := ""
	if l.Status == domain.LotEnded {
		if l.HammerPrice != nil && (l.ReservePrice == nil || *l.HammerPrice >= *l.ReservePrice) {
			saleResult = "SOLD"
		} else {
			saleResult = "RESERVE_NOT_MET"
		}
	}
	return lotPublicResponse{
		ID:            l.ID,
		LotID:         l.ID,
		AuctionID:     l.AuctionID,
		LotNumber:     l.LotNumber,
		OrderIndex:    l.OrderIndex,
		Title:         l.Title,
		Description:   l.Description,
		StartingPrice: l.StartingPrice,
		StartingMinor: l.StartingPrice,
		CurrentPrice:  l.CurrentPrice,
		CurrentMinor:  l.CurrentPrice,
		MinimumNext:   l.CurrentPrice + l.MinIncrement,
		ReservePrice:  l.ReservePrice,
		MinIncrement:  l.MinIncrement,
		Increment:     l.MinIncrement,
		Metadata:      l.Metadata,
		Status:        l.Status,
		HammerPrice:   l.HammerPrice,
		SaleResult:    saleResult,
		StartedAt:     l.StartedAt,
		EndedAt:       l.EndedAt,
		CreatedAt:     l.CreatedAt,
		UpdatedAt:     l.UpdatedAt,
	}
}

func parsePositiveIntOrDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val <= 0 {
		return fallback
	}
	return val
}

func encodeAuditCursor(createdAt time.Time, id uuid.UUID) string {
	payload := fmt.Sprintf("%d|%s", createdAt.UTC().UnixNano(), id.String())
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func decodeAuditCursor(raw string) (*time.Time, *uuid.UUID, error) {
	if raw == "" {
		return nil, nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, nil, err
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return nil, nil, errors.New("invalid cursor")
	}
	unixNano, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, nil, err
	}
	ts := time.Unix(0, unixNano).UTC()
	return &ts, &id, nil
}

func parseISODateTime(raw string, endOfDay bool) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		tt := t.UTC()
		return &tt, nil
	}
	day, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, err
	}
	utc := day.UTC()
	if endOfDay {
		adjusted := utc.Add(24*time.Hour - time.Nanosecond)
		return &adjusted, nil
	}
	return &utc, nil
}

func parseAuditMetadata(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil || out == nil {
		return map[string]any{}
	}
	return out
}

func auditActorDisplay(actorRole string, actorUserID *uuid.UUID) string {
	role := auditActorRole(actorRole, actorUserID)
	if role == "SYSTEM" {
		return "System"
	}
	if role == "BIDDER" && actorUserID != nil {
		return fmt.Sprintf("Bidder %03d", crc32.ChecksumIEEE([]byte(actorUserID.String()))%1000)
	}
	if actorUserID != nil {
		label := strings.ToUpper(role[:1]) + strings.ToLower(role[1:])
		return fmt.Sprintf("%s %s", label, actorUserID.String()[:8])
	}
	return "System"
}

func auditActorRole(actorRole string, actorUserID *uuid.UUID) string {
	role := strings.ToUpper(strings.TrimSpace(actorRole))
	if role != "" {
		return role
	}
	if actorUserID == nil {
		return "SYSTEM"
	}
	return "SYSTEM"
}

func auditSummary(eventType string, metadata json.RawMessage) string {
	typeLabel := strings.TrimSpace(eventType)
	if typeLabel == "" {
		return "UNKNOWN"
	}
	parsed := parseAuditMetadata(metadata)
	if amount, ok := parsed["amount_minor"]; ok {
		return fmt.Sprintf("%s (%v)", typeLabel, amount)
	}
	return typeLabel
}

func toAuthUserResponse(user *domain.User) authUserResponse {
	return authUserResponse{
		ID:             user.ID,
		TenantID:       user.TenantID,
		Email:          user.Email,
		Role:           user.Role,
		Status:         user.Status,
		EmailVerified:  user.EmailVerified,
		ApprovalStatus: user.ApprovalStatus,
		KYCStatus:      user.KYCStatus,
		CreatedAt:      user.CreatedAt,
	}
}

func (h *Handler) accessCookieName() string {
	name := strings.TrimSpace(h.AccessCookie)
	if name == "" {
		return "access_token"
	}
	return name
}

func (h *Handler) csrfCookieName() string {
	name := strings.TrimSpace(h.CSRFCookie)
	if name == "" {
		return "csrf_token"
	}
	return name
}

func (h *Handler) setSessionCookies(w http.ResponseWriter, token string) {
	ttl := h.JWTTTL
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	csrf := uuid.NewString()
	maxAge := int(ttl.Seconds())

	http.SetCookie(w, &http.Cookie{
		Name:     h.accessCookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     h.csrfCookieName(),
		Value:    csrf,
		Path:     "/",
		HttpOnly: false,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}

func (h *Handler) clearSessionCookies(w http.ResponseWriter) {
	for _, name := range []string{h.accessCookieName(), h.csrfCookieName()} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: name == h.accessCookieName(),
			Secure:   h.CookieSecure,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
	}
}
