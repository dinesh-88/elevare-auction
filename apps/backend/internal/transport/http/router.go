package httptransport

import (
	"net/http"
	"os"
	"strings"

	wsadapter "github.com/dinesh-88/elevare-auction/internal/adapter/broadcaster/ws"
	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(h *Handler, tenantMW middleware.TenantMiddleware, authMW middleware.AuthMiddleware, hub *wsadapter.Hub, platformAdminToken string, uploadsDir string, serveLocalUploads bool) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.CaptureRequestMeta)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowOriginFunc: func(_ *http.Request, origin string) bool {
			return strings.HasPrefix(origin, "http://") && strings.HasSuffix(origin, ".localhost:5173") ||
				strings.HasPrefix(origin, "https://") && strings.HasSuffix(origin, ".localhost:5173") ||
				strings.HasPrefix(origin, "http://") && strings.HasSuffix(origin, ".localhost:3000") ||
				strings.HasPrefix(origin, "https://") && strings.HasSuffix(origin, ".localhost:3000")
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))
	csrfMW := middleware.RequireCSRFCookieHeader(h.CSRFCookie)
	if serveLocalUploads {
		uploadsDir = strings.TrimSpace(uploadsDir)
		if uploadsDir != "" {
			_ = os.MkdirAll(uploadsDir, 0o755)
			r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))
		}
	}

	r.Get("/swagger/openapi.yaml", ServeOpenAPISpec())
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusTemporaryRedirect)
	})
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/openapi.yaml")))
	r.Post("/signup", h.PublicTenantSignup)

	r.Route("/platform/admin", func(r chi.Router) {
		r.Use(middleware.RequirePlatformAdmin(platformAdminToken))
		r.Get("/tenants", h.ListTenantRequests)
		r.Patch("/tenants/{id}/approve", h.ApproveTenant)
		r.Patch("/tenants/{id}/reject", h.RejectTenant)
	})

	r.Group(func(r chi.Router) {
		r.Use(tenantMW.Resolve)

		r.Get("/public/config", h.PublicConfig)
		r.Get("/auctions", h.ListAuctions)
		r.Get("/auctions/{id}", h.GetAuction)
		r.Get("/auctions/{id}/ws", AuctionWS(hub))
		r.Get("/auctions/{auctionId}/lots", h.ListLotsByAuction)
		r.Get("/lots/{lotId}", h.GetLot)
		r.Get("/lots/{lotId}/bids", h.ListLotBids)
		r.Get("/lots/{id}/ws", LotWS(hub))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/verify-email", h.VerifyEmail)
			r.With(authMW.RequireAuth, csrfMW).Post("/logout", h.Logout)
			r.With(authMW.RequireAuth).Get("/session", h.Session)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Use(csrfMW)
			r.Post("/lots/{lotId}/bid", h.PlaceLotBid)
			r.Get("/lots/{lotId}/financials", h.GetLotFinancials)
			r.Get("/me/winnings", h.ListMyWinningLots)
			r.Get("/invoices/{invoiceId}", h.GetInvoice)
			r.Get("/invoices/{invoiceId}/pdf", h.DownloadInvoicePDF)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Use(middleware.RequireAnyRole(domain.RoleOwner, domain.RoleAdmin, domain.RoleStaff, domain.RoleAccounting))
			r.Use(csrfMW)
			r.Get("/auctions", h.ListAdminAuctions)
			r.With(authz.Require(domain.PermissionSiteEdit)).Get("/settings", h.GetTenantSettings)
			r.With(authz.Require(domain.PermissionSiteEdit)).Patch("/settings", h.UpdateTenantSettings)
			r.With(authz.Require(domain.PermissionSiteEdit)).Post("/settings", h.UpdateTenantSettings)
			r.With(authz.Require(domain.PermissionSiteEdit)).Get("/site", h.GetSite)
			r.With(authz.Require(domain.PermissionSiteEdit)).Put("/site/draft", h.UpdateSiteDraft)
			r.With(authz.Require(domain.PermissionSitePublish)).Post("/site/publish", h.PublishSite)
			r.With(authz.Require(domain.PermissionAuctionCreate)).Post("/auctions", h.CreateAuction)
			r.With(authz.Require(domain.PermissionLotCreate)).Post("/auctions/{auctionId}/lots", h.CreateLot)
			r.With(authz.Require(domain.PermissionLotReorder)).Post("/auctions/{auctionId}/reorder-lots", h.ReorderLots)
			r.With(authz.Require(domain.PermissionAuctionUpdate)).Patch("/auctions/{id}", h.UpdateAuction)
			r.With(authz.Require(domain.PermissionAuctionUpdate)).Delete("/auctions/{id}", h.DeleteAuction)
			r.With(authz.Require(domain.PermissionAuctionStart)).Patch("/auctions/{id}/start", h.StartAuction)
			r.With(authz.Require(domain.PermissionAuctionEnd)).Patch("/auctions/{id}/pause", h.PauseAuction)
			r.With(authz.Require(domain.PermissionAuctionEnd)).Patch("/auctions/{id}/end", h.EndAuction)
			r.With(authz.Require(domain.PermissionLotUpdate)).Patch("/lots/{lotId}", h.UpdateLot)
			r.With(authz.Require(domain.PermissionLotUpdate)).Delete("/lots/{lotId}", h.DeleteLot)
			r.With(authz.Require(domain.PermissionLotStart)).Post("/lots/{lotId}/start", h.StartLot)
			r.With(authz.Require(domain.PermissionLotEnd)).Post("/lots/{lotId}/end", h.EndLot)
			r.With(authz.Require(domain.PermissionInvoiceMarkPaid)).Post("/lots/{lotId}/invoice/generate", h.GenerateLotInvoice)
			r.With(authz.Require(domain.PermissionSiteEdit)).Post("/uploads/auction-image", h.UploadAuctionImage)
			r.With(authz.Require(domain.PermissionSiteEdit)).Post("/uploads/lot-image", h.UploadLotImage)
			r.With(authz.Require(domain.PermissionSiteEdit)).Post("/assets/presign", h.PresignAsset)
			r.With(authz.Require(domain.PermissionSiteEdit)).Post("/assets/upload", h.UploadAsset)
			r.With(authz.Require(domain.PermissionSiteEdit)).Post("/assets", h.RegisterAsset)
			r.With(authz.Require(domain.PermissionSiteEdit)).Get("/assets", h.ListAssets)
			r.With(authz.Require(domain.PermissionFeesManage)).Get("/fees/components", h.ListFeeComponents)
			r.With(authz.Require(domain.PermissionFeesManage)).Post("/fees/components", h.CreateFeeComponent)
			r.With(authz.Require(domain.PermissionFeesManage)).Patch("/fees/components/{id}", h.UpdateFeeComponent)
			r.With(authz.Require(domain.PermissionFeesManage)).Delete("/fees/components/{id}", h.DeleteFeeComponent)
			r.With(authz.Require(domain.PermissionFeesManage)).Get("/fees/assignments", h.ListFeeAssignments)
			r.With(authz.Require(domain.PermissionFeesManage)).Post("/fees/assignments", h.CreateFeeAssignment)
			r.With(authz.Require(domain.PermissionFeesManage)).Delete("/fees/assignments/{id}", h.DeleteFeeAssignment)
			r.With(authz.Require(domain.PermissionInvoiceMarkPaid)).Patch("/invoices/{invoiceId}/status", h.UpdateInvoiceStatus)
			r.With(authz.Require(domain.PermissionAuditView)).Get("/audit", h.ListAuditLogs)
			r.With(authz.Require(domain.PermissionAuditView)).Get("/audit/entity", h.ListAuditLogsByEntity)
			r.With(authz.Require(domain.PermissionBidderApprove)).Get("/bidders", h.ListBidders)
			r.With(authz.Require(domain.PermissionBidderApprove)).Patch("/bidders/{id}/approve", h.ApproveBidder)
			r.With(authz.Require(domain.PermissionBidderSuspend)).Patch("/bidders/{id}/reject", h.RejectBidder)
		})
	})

	return r
}
