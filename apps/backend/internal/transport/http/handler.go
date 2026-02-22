package httptransport

import (
	"encoding/json"
	"errors"
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
	Auth         *usecase.AuthUseCase
	Auction      *usecase.AuctionUseCase
	Lot          *usecase.LotUseCase
	Fee          *usecase.FeeUseCase
	Invoice      *usecase.InvoiceUseCase
	Bidder       *usecase.BidderUseCase
	Audit        *usecase.AuditUseCase
	Site         *usecase.SiteUseCase
	Asset        *usecase.AssetUseCase
	Onboarding   *usecase.OnboardingUseCase
	Storage      port.ObjectStorage
	PreviewToken string
}

type lotPublicResponse struct {
	ID            uuid.UUID        `json:"id"`
	AuctionID     uuid.UUID        `json:"auction_id"`
	LotNumber     string           `json:"lot_number"`
	OrderIndex    int              `json:"order_index"`
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	StartingPrice int64            `json:"starting_price"`
	CurrentPrice  int64            `json:"current_price"`
	ReservePrice  *int64           `json:"reserve_price"`
	MinIncrement  int64            `json:"min_increment"`
	Metadata      json.RawMessage  `json:"metadata"`
	Status        domain.LotStatus `json:"status"`
	HammerPrice   *int64           `json:"hammer_price"`
	SaleResult    string           `json:"sale_result"`
	StartedAt     *time.Time       `json:"started_at"`
	EndedAt       *time.Time       `json:"ended_at"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
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

	isAdmin := claims.Role == domain.RoleAdmin
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
	writeJSON(w, http.StatusOK, bids)
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
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
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
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
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
	if claims.Role != domain.RoleBidder && claims.Role != domain.RoleAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}
	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	res, err := h.Lot.PlaceBid(r.Context(), port.PlaceLotBidTxInput{
		TenantID: tenant.ID,
		LotID:    lotID,
		UserID:   claims.UserID,
		Amount:   req.Amount,
		Now:      time.Now().UTC(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, res)
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
		AuctionID:     l.AuctionID,
		LotNumber:     l.LotNumber,
		OrderIndex:    l.OrderIndex,
		Title:         l.Title,
		Description:   l.Description,
		StartingPrice: l.StartingPrice,
		CurrentPrice:  l.CurrentPrice,
		ReservePrice:  l.ReservePrice,
		MinIncrement:  l.MinIncrement,
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
