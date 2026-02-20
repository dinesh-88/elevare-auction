package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	wsadapter "github.com/dinesh-88/elevare-auction/internal/adapter/broadcaster/ws"
	"github.com/dinesh-88/elevare-auction/internal/adapter/repo/postgres"
	"github.com/dinesh-88/elevare-auction/internal/adapter/security"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	httptransport "github.com/dinesh-88/elevare-auction/internal/transport/http"
	"github.com/dinesh-88/elevare-auction/internal/usecase"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := httptransport.LoadConfig()
	db, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer db.Close()

	tenantRepo := &postgres.TenantRepo{DB: db}
	userRepo := &postgres.UserRepo{DB: db}
	auctionRepo := &postgres.AuctionRepo{DB: db}
	lotRepo := &postgres.LotRepo{DB: db}

	hasher := &security.Argon2Hasher{}
	jwtSvc := &security.JWTService{Secret: []byte(cfg.JWTSecret), TTL: cfg.JWTTTL}
	hub := wsadapter.NewHub()

	authUC := &usecase.AuthUseCase{Users: userRepo, Hasher: hasher, Tokens: jwtSvc}
	auctionUC := &usecase.AuctionUseCase{Auctions: auctionRepo, Broadcaster: hub}
	lotUC := &usecase.LotUseCase{Lots: lotRepo, Broadcaster: hub}
	bidderUC := &usecase.BidderUseCase{Users: userRepo}
	onboardingUC := &usecase.OnboardingUseCase{Tenants: tenantRepo, Hasher: hasher}

	h := &httptransport.Handler{Auth: authUC, Auction: auctionUC, Lot: lotUC, Bidder: bidderUC, Onboarding: onboardingUC}
	tenantMW := middleware.TenantMiddleware{Tenants: tenantRepo}
	authMW := middleware.AuthMiddleware{Tokens: jwtSvc}

	router := httptransport.NewRouter(h, tenantMW, authMW, hub, cfg.PlatformAdminToken)
	httptransport.StartScheduler(ctx, auctionUC, lotUC, 5*time.Second)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("api listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server failed: %v", err)
	}
}
