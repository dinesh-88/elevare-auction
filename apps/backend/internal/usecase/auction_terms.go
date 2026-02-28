package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type AuctionTermsUseCase struct {
	Terms    port.AuctionTermsRepository
	Auctions port.AuctionRepository
	Audit    port.AuditRepository
}

func hashTermsContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func (uc *AuctionTermsUseCase) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID, locale string) ([]domain.AuctionTerms, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionUpdate); err != nil {
		return nil, ErrForbidden
	}
	if !domain.IsSupportedLocale(locale) {
		return nil, ErrValidation
	}
	return uc.Terms.ListByAuction(ctx, tenantID, auctionID, domain.NormalizeLocale(locale))
}

func (uc *AuctionTermsUseCase) GetActive(ctx context.Context, tenantID, auctionID uuid.UUID, locale string) (*domain.AuctionTerms, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if !domain.IsSupportedLocale(locale) {
		return nil, ErrValidation
	}
	return uc.Terms.GetActiveByAuction(ctx, tenantID, auctionID, domain.NormalizeLocale(locale))
}

func (uc *AuctionTermsUseCase) GetBidderStatus(ctx context.Context, tenantID, auctionID, userID uuid.UUID, locale string) (*domain.AuctionTerms, *domain.BidderTermsAcceptance, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, nil, err
	}
	if !domain.IsSupportedLocale(locale) {
		return nil, nil, ErrValidation
	}
	active, err := uc.Terms.GetActiveByAuction(ctx, tenantID, auctionID, domain.NormalizeLocale(locale))
	if err != nil || active == nil {
		return active, nil, err
	}
	acceptance, err := uc.Terms.GetAcceptance(ctx, tenantID, auctionID, userID, active.ID)
	if err != nil {
		return active, nil, err
	}
	return active, acceptance, nil
}

func (uc *AuctionTermsUseCase) CreateVersion(ctx context.Context, tenantID, auctionID uuid.UUID, locale string, contentMarkdown string, publish bool) (*domain.AuctionTerms, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionUpdate); err != nil {
		return nil, ErrForbidden
	}
	if !domain.IsSupportedLocale(locale) {
		return nil, ErrValidation
	}
	trimmed := strings.TrimSpace(contentMarkdown)
	if trimmed == "" {
		return nil, ErrValidation
	}
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok || claims.UserID == uuid.Nil {
		return nil, ErrUnauthorized
	}
	if uc.Auctions != nil {
		if _, err := uc.Auctions.GetByID(ctx, tenantID, auctionID); err != nil {
			return nil, err
		}
	}
	item := &domain.AuctionTerms{
		ID:              uuid.New(),
		TenantID:        tenantID,
		AuctionID:       auctionID,
		Locale:          domain.NormalizeLocale(locale),
		ContentMarkdown: contentMarkdown,
		ContentHash:     hashTermsContent(contentMarkdown),
		IsActive:        false,
		CreatedByUserID: claims.UserID,
	}
	if err := uc.Terms.CreateVersion(ctx, item); err != nil {
		return nil, err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{
			"auction_terms_id": item.ID,
			"version_number":   item.VersionNumber,
			"content_hash":     item.ContentHash,
		})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "AUCTION_TERMS_CREATED",
			EntityType:  "AUCTION_TERMS",
			EntityID:    &item.ID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	if publish {
		if err := uc.PublishVersion(ctx, tenantID, auctionID, item.ID, locale); err != nil {
			return nil, err
		}
		item.IsActive = true
	}
	return item, nil
}

func (uc *AuctionTermsUseCase) PublishVersion(ctx context.Context, tenantID, auctionID, auctionTermsID uuid.UUID, locale string) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuctionUpdate); err != nil {
		return ErrForbidden
	}
	if !domain.IsSupportedLocale(locale) {
		return ErrValidation
	}
	if err := uc.Terms.PublishVersion(ctx, tenantID, auctionID, auctionTermsID, domain.NormalizeLocale(locale)); err != nil {
		return err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"auction_terms_id": auctionTermsID})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "AUCTION_TERMS_PUBLISHED",
			EntityType:  "AUCTION_TERMS",
			EntityID:    &auctionTermsID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return nil
}

func (uc *AuctionTermsUseCase) AcceptActive(ctx context.Context, tenantID, auctionID, userID uuid.UUID, locale string, now time.Time, acceptedIP, acceptedUserAgent string) (*domain.BidderTermsAcceptance, *domain.AuctionTerms, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, nil, err
	}
	if !domain.IsSupportedLocale(locale) {
		return nil, nil, ErrValidation
	}
	active, err := uc.Terms.GetActiveByAuction(ctx, tenantID, auctionID, domain.NormalizeLocale(locale))
	if err != nil || active == nil {
		return nil, active, ErrTermsUnavailable
	}
	existing, err := uc.Terms.GetAcceptance(ctx, tenantID, auctionID, userID, active.ID)
	if err != nil {
		return nil, active, err
	}
	if existing != nil {
		return existing, active, nil
	}
	item := &domain.BidderTermsAcceptance{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		AuctionID:          auctionID,
		UserID:             userID,
		AuctionTermsID:     active.ID,
		TermsVersionNumber: active.VersionNumber,
		ContentHash:        active.ContentHash,
		AcceptedAt:         now,
		AcceptedIP:         acceptedIP,
		AcceptedUserAgent:  acceptedUserAgent,
		CreatedAt:          now,
	}
	if err := uc.Terms.CreateAcceptance(ctx, item); err != nil {
		return nil, active, err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{
			"auction_terms_id":    active.ID,
			"version_number":      active.VersionNumber,
			"content_hash":        active.ContentHash,
			"accepted_ip":         item.AcceptedIP,
			"accepted_user_agent": item.AcceptedUserAgent,
			"accepted_at":         item.AcceptedAt,
		})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: &userID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "BIDDER_TERMS_ACCEPTED",
			EntityType:  "AUCTION_TERMS",
			EntityID:    &active.ID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return item, active, nil
}
