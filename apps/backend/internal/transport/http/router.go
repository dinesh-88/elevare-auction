package httptransport

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	wsadapter "github.com/dinesh-88/elevare-auction/internal/adapter/broadcaster/ws"
	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	errmiddleware "github.com/dinesh-88/elevare-auction/internal/http/middleware"
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
	r.Use(errmiddleware.ErrorMiddleware)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"https://elevareauctions.com",
			"https://www.elevareauctions.com",
		},
		AllowOriginFunc: func(_ *http.Request, origin string) bool {
			u, err := url.Parse(origin)
			if err != nil {
				return false
			}
			host := strings.ToLower(u.Hostname())
			scheme := strings.ToLower(u.Scheme)

			if (scheme == "http" || scheme == "https") &&
				(strings.HasSuffix(host, ".localhost") || host == "localhost" || host == "127.0.0.1") {
				return true
			}
			if scheme == "https" &&
				(host == "elevareauctions.com" || host == "www.elevareauctions.com" || strings.HasSuffix(host, ".elevareauctions.com")) {
				return true
			}
			return false
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Tenant-Host"},
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
	r.Post("/auth/discover-tenant", h.DiscoverTenantLogin)
	r.Post("/contact", h.SubmitDemoRequest)
	r.Post("/internal/process-email-retries", h.ProcessEmailRetries)
	r.Get("/public/institutions", h.ListPublicInstitutions)
	r.Get("/public/institutions/{slug}", h.GetPublicInstitution)
	r.Get("/public/featured-auction", h.GetPublicFeaturedAuction)

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
		r.Get("/auctions/{auctionId}/terms/active", h.GetActiveAuctionTerms)
		r.Get("/auctions/{id}/ws", AuctionWS(hub))
		r.Get("/auctions/{auctionId}/lots", h.ListLotsByAuction)
		r.Get("/lots/{lotId}", h.GetLot)
		r.Get("/lots/{lotId}/bids", h.ListLotBids)
		r.Get("/lots/{id}/ws", LotWS(hub))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/verify-email", h.VerifyEmail)
			r.Post("/invitations/accept", h.AcceptTenantInvitation)
			r.With(authMW.RequireAuth, csrfMW).Post("/logout", h.Logout)
			r.With(authMW.RequireAuth).Get("/session", h.Session)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Use(csrfMW)
			r.Post("/auctions/{auctionId}/request-to-bid", h.RequestToBid)
			r.Get("/auctions/{auctionId}/terms/status", h.GetAuctionTermsStatus)
			r.Post("/auctions/{auctionId}/terms/accept", h.AcceptAuctionTerms)
			r.Post("/lots/{lotId}/bid", h.PlaceLotBid)
			r.Get("/lots/{lotId}/financials", h.GetLotFinancials)
			r.Patch("/me", h.UpdateMe)
			r.Get("/me/winnings", h.ListMyWinningLots)
			r.Get("/me/notifications", h.ListMyNotifications)
			r.Post("/me/notifications/{id}/read", h.MarkMyNotificationRead)
			r.Post("/me/notifications/read-all", h.MarkAllMyNotificationsRead)
			r.Get("/me/notification-preferences", h.GetMyNotificationPreferences)
			r.Put("/me/notification-preferences", h.UpdateMyNotificationPreferences)
			r.Get("/invoices/{invoiceId}", h.GetInvoice)
			r.Get("/invoices/{invoiceId}/pdf", h.DownloadInvoicePDF)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Use(middleware.RequireAnyRole(domain.RoleOwner, domain.RoleAdmin, domain.RoleStaff, domain.RoleAccounting, domain.RoleViewer))
			r.Use(csrfMW)
			onboardingComplete := middleware.RequireOnboardingComplete()
			r.With(authz.Require(domain.PermissionSiteEdit)).Get("/onboarding", h.GetInstitutionOnboarding)
			r.With(authz.Require(domain.PermissionSiteEdit)).Put("/onboarding", h.SaveInstitutionOnboarding)
			r.With(authz.Require(domain.PermissionSiteEdit)).Post("/onboarding/complete", h.CompleteInstitutionOnboarding)
			r.Get("/auctions", h.ListAdminAuctions)
			r.With(authz.Require(domain.PermissionSiteEdit)).Get("/settings", h.GetTenantSettings)
			r.With(authz.Require(domain.PermissionSiteEdit)).Patch("/settings", h.UpdateTenantSettings)
			r.With(authz.Require(domain.PermissionSiteEdit)).Post("/settings", h.UpdateTenantSettings)
			r.With(authz.Require(domain.PermissionSiteEdit)).Get("/site", h.GetSite)
			r.With(authz.Require(domain.PermissionSiteEdit)).Put("/site/draft", h.UpdateSiteDraft)
			r.With(authz.Require(domain.PermissionSitePublish)).Post("/site/publish", h.PublishSite)
			r.With(onboardingComplete, authz.Require(domain.PermissionAuctionCreate)).Post("/auctions", h.CreateAuction)
			r.With(authz.Require(domain.PermissionLotCreate)).Post("/auctions/{auctionId}/lots", h.CreateLot)
			r.With(authz.Require(domain.PermissionLotReorder)).Post("/auctions/{auctionId}/reorder-lots", h.ReorderLots)
			r.With(onboardingComplete, authz.Require(domain.PermissionAuctionUpdate)).Patch("/auctions/{id}", h.UpdateAuction)
			r.With(authz.Require(domain.PermissionAuctionUpdate)).Delete("/auctions/{id}", h.DeleteAuction)
			r.With(onboardingComplete, authz.Require(domain.PermissionAuctionStart)).Patch("/auctions/{id}/start", h.StartAuction)
			r.With(authz.Require(domain.PermissionAuctionEnd)).Patch("/auctions/{id}/pause", h.PauseAuction)
			r.With(authz.Require(domain.PermissionAuctionEnd)).Patch("/auctions/{id}/end", h.EndAuction)
			r.With(onboardingComplete, authz.Require(domain.PermissionAuctionUpdate)).Put("/auctions/{auctionId}/settings", h.UpdateAuctionBidderSettings)
			r.With(authz.Require(domain.PermissionAuctionUpdate)).Get("/auctions/{auctionId}/terms", h.ListAuctionTerms)
			r.With(onboardingComplete, authz.Require(domain.PermissionAuctionUpdate)).Post("/auctions/{auctionId}/terms", h.CreateAuctionTerms)
			r.With(onboardingComplete, authz.Require(domain.PermissionAuctionUpdate)).Post("/auctions/{auctionId}/terms/{termsId}/publish", h.PublishAuctionTerms)
			r.With(authz.Require(domain.PermissionBidderApprove)).Get("/auctions/{auctionId}/bidders", h.ListAuctionBidders)
			r.With(authz.Require(domain.PermissionBidderApprove)).Post("/auctions/{auctionId}/bidders/{userId}/approve", h.ApproveAuctionBidder)
			r.With(authz.Require(domain.PermissionBidderApprove)).Post("/auctions/{auctionId}/bidders/{userId}/reject", h.RejectAuctionBidder)
			r.With(authz.Require(domain.PermissionBidderApprove)).Put("/auctions/{auctionId}/bidders/{userId}", h.UpdateAuctionBidderApproval)
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
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Get("/fees/components", h.ListFeeComponents)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Post("/fees/components", h.CreateFeeComponent)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Patch("/fees/components/{id}", h.UpdateFeeComponent)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Delete("/fees/components/{id}", h.DeleteFeeComponent)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Get("/fees/assignments", h.ListFeeAssignments)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Post("/fees/assignments", h.CreateFeeAssignment)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Delete("/fees/assignments/{id}", h.DeleteFeeAssignment)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Get("/taxes/rules", h.ListTaxRules)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Post("/taxes/rules", h.CreateTaxRule)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Patch("/taxes/rules/{id}", h.UpdateTaxRule)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Delete("/taxes/rules/{id}", h.DeleteTaxRule)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Get("/taxes/assignments", h.ListTaxAssignments)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Post("/taxes/assignments", h.CreateTaxAssignment)
			r.With(onboardingComplete, authz.Require(domain.PermissionFeesManage)).Delete("/taxes/assignments/{id}", h.DeleteTaxAssignment)
			r.With(authz.Require(domain.PermissionAuctionStart)).Post("/simulator/trigger", h.TriggerBidSimulation)
			r.With(authz.Require(domain.PermissionInvoiceMarkPaid)).Patch("/invoices/{invoiceId}/status", h.UpdateInvoiceStatus)
			r.With(authz.Require(domain.PermissionInvoiceView)).Get("/auctions/{auctionId}/settlement-summary", h.GetAuctionSettlementSummary)
			r.With(authz.Require(domain.PermissionInvoiceView)).Get("/auctions/{auctionId}/invoices", h.ListAuctionInvoices)
			r.With(authz.Require(domain.PermissionInvoiceMarkPaid)).Post("/invoices/{invoiceId}/mark-paid", h.MarkInvoicePaid)
			r.With(authz.Require(domain.PermissionInvoiceMarkPaid)).Post("/invoices/{invoiceId}/cancel", h.CancelInvoice)
			r.With(authz.Require(domain.PermissionInvoiceMarkPaid)).Get("/auctions/{auctionId}/export-settlement-csv", h.ExportAuctionSettlementCSV)
			r.With(authz.Require(domain.PermissionAuditView)).Get("/audit", h.ListAuditLogs)
			r.With(authz.Require(domain.PermissionAuditView)).Get("/audit/entity", h.ListAuditLogsByEntity)
			r.With(authz.Require(domain.PermissionAuditView)).Get("/lots/{lotId}/audit", h.ListLotAuditLogs)
			r.With(authz.Require(domain.PermissionBidderApprove)).Get("/bidders", h.ListBidders)
			r.With(authz.Require(domain.PermissionBidderApprove)).Get("/bidders/{userId}", h.GetBidderProfile)
			r.With(authz.Require(domain.PermissionBidderSuspend)).Put("/bidders/{userId}", h.UpdateBidderProfile)
			r.With(authz.Require(domain.PermissionBidderApprove)).Patch("/bidders/{id}/approve", h.ApproveBidder)
			r.With(authz.Require(domain.PermissionBidderSuspend)).Patch("/bidders/{id}/reject", h.RejectBidder)
			r.With(authz.Require(domain.PermissionStaffInvite)).Get("/staff", h.ListStaff)
			r.With(authz.Require(domain.PermissionStaffInvite)).Post("/staff/invitations", h.CreateStaffInvitation)
			r.With(authz.Require(domain.PermissionRoleManage)).Patch("/staff/{tenantUserId}", h.UpdateStaffRole)
		})
	})

	return r
}
