package httptransport

import (
	"net/http"
	"os"
	"strings"

	wsadapter "github.com/dinesh-88/elevare-auction/internal/adapter/broadcaster/ws"
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
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
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
		})

		r.Group(func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Post("/lots/{lotId}/bid", h.PlaceLotBid)
			r.Get("/lots/{lotId}/financials", h.GetLotFinancials)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Use(middleware.RequireRole(domain.RoleAdmin))
			r.Get("/auctions", h.ListAdminAuctions)
			r.Post("/auctions", h.CreateAuction)
			r.Post("/auctions/{auctionId}/lots", h.CreateLot)
			r.Post("/auctions/{auctionId}/reorder-lots", h.ReorderLots)
			r.Patch("/auctions/{id}", h.UpdateAuction)
			r.Delete("/auctions/{id}", h.DeleteAuction)
			r.Patch("/auctions/{id}/start", h.StartAuction)
			r.Patch("/auctions/{id}/pause", h.PauseAuction)
			r.Patch("/auctions/{id}/end", h.EndAuction)
			r.Patch("/lots/{lotId}", h.UpdateLot)
			r.Delete("/lots/{lotId}", h.DeleteLot)
			r.Post("/lots/{lotId}/start", h.StartLot)
			r.Post("/lots/{lotId}/end", h.EndLot)
			r.Post("/uploads/auction-image", h.UploadAuctionImage)
			r.Post("/uploads/lot-image", h.UploadLotImage)
			r.Get("/bidders", h.ListBidders)
			r.Patch("/bidders/{id}/approve", h.ApproveBidder)
		})
	})

	return r
}
