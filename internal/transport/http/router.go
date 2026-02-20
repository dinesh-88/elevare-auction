package httptransport

import (
	"net/http"

	wsadapter "github.com/dinesh-88/elevare-auction/internal/adapter/broadcaster/ws"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(h *Handler, tenantMW middleware.TenantMiddleware, authMW middleware.AuthMiddleware, hub *wsadapter.Hub) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/swagger/openapi.yaml", ServeOpenAPISpec())
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusTemporaryRedirect)
	})
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/openapi.yaml")))

	r.Group(func(r chi.Router) {
		r.Use(tenantMW.Resolve)

		r.Get("/public/config", h.PublicConfig)
		r.Get("/auctions", h.ListAuctions)
		r.Get("/auctions/{id}", h.GetAuction)
		r.Get("/auctions/{id}/ws", AuctionWS(hub))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/verify-email", h.VerifyEmail)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Post("/auctions/{id}/bid", h.PlaceBid)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Use(middleware.RequireRole(domain.RoleAdmin))
			r.Post("/auctions", h.CreateAuction)
			r.Patch("/auctions/{id}/start", h.StartAuction)
			r.Patch("/auctions/{id}/pause", h.PauseAuction)
			r.Patch("/auctions/{id}/end", h.EndAuction)
			r.Get("/bidders", h.ListBidders)
			r.Patch("/bidders/{id}/approve", h.ApproveBidder)
		})
	})

	return r
}
