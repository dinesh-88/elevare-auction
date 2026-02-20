package httptransport

import (
	"errors"
	"net/http"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/dinesh-88/elevare-auction/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	Auth       *usecase.AuthUseCase
	Auction    *usecase.AuctionUseCase
	Bidder     *usecase.BidderUseCase
	Onboarding *usecase.OnboardingUseCase
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
	writeJSON(w, http.StatusOK, map[string]any{
		"name":                   tenant.Name,
		"branding":               tenant.Branding,
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
		Title                   string               `json:"title"`
		Description             string               `json:"description"`
		StartingPrice           int64                `json:"starting_price"`
		ReservePrice            *int64               `json:"reserve_price"`
		MinIncrement            int64                `json:"min_increment"`
		StartTime               time.Time            `json:"start_time"`
		EndTime                 time.Time            `json:"end_time"`
		AllowManualStart        bool                 `json:"allow_manual_start"`
		AllowManualPause        bool                 `json:"allow_manual_pause"`
		AllowManualEnd          bool                 `json:"allow_manual_end"`
		AutoExtendEnabled       bool                 `json:"auto_extend_enabled"`
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
		Title:                   req.Title,
		Description:             req.Description,
		StartingPrice:           req.StartingPrice,
		ReservePrice:            req.ReservePrice,
		MinIncrement:            req.MinIncrement,
		StartTime:               req.StartTime,
		EndTime:                 req.EndTime,
		AllowManualStart:        req.AllowManualStart,
		AllowManualPause:        req.AllowManualPause,
		AllowManualEnd:          req.AllowManualEnd,
		AutoExtendEnabled:       req.AutoExtendEnabled,
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
		Title                   string               `json:"title"`
		Description             string               `json:"description"`
		StartingPrice           int64                `json:"starting_price"`
		ReservePrice            *int64               `json:"reserve_price"`
		MinIncrement            int64                `json:"min_increment"`
		StartTime               time.Time            `json:"start_time"`
		EndTime                 time.Time            `json:"end_time"`
		AllowManualStart        bool                 `json:"allow_manual_start"`
		AllowManualPause        bool                 `json:"allow_manual_pause"`
		AllowManualEnd          bool                 `json:"allow_manual_end"`
		AutoExtendEnabled       bool                 `json:"auto_extend_enabled"`
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
		Title:                   req.Title,
		Description:             req.Description,
		StartingPrice:           req.StartingPrice,
		ReservePrice:            req.ReservePrice,
		MinIncrement:            req.MinIncrement,
		StartTime:               req.StartTime,
		EndTime:                 req.EndTime,
		AllowManualStart:        req.AllowManualStart,
		AllowManualPause:        req.AllowManualPause,
		AllowManualEnd:          req.AllowManualEnd,
		AutoExtendEnabled:       req.AutoExtendEnabled,
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

func (h *Handler) PlaceBid(w http.ResponseWriter, r *http.Request) {
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
	auctionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	res, err := h.Auction.PlaceBid(r.Context(), port.PlaceBidTxInput{
		TenantID:  tenant.ID,
		AuctionID: auctionID,
		UserID:    claims.UserID,
		Amount:    req.Amount,
		Now:       time.Now().UTC(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}
