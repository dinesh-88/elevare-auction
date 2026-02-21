package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	wsadapter "github.com/dinesh-88/elevare-auction/internal/adapter/broadcaster/ws"
	"github.com/dinesh-88/elevare-auction/internal/adapter/repo/postgres"
	"github.com/dinesh-88/elevare-auction/internal/adapter/security"
	"github.com/dinesh-88/elevare-auction/internal/adapter/storage"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
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
	feeRepo := &postgres.FeeRepo{DB: db}

	hasher := &security.Argon2Hasher{}
	jwtSvc := &security.JWTService{Secret: []byte(cfg.JWTSecret), TTL: cfg.JWTTTL}
	hub := wsadapter.NewHub()

	authUC := &usecase.AuthUseCase{Users: userRepo, Hasher: hasher, Tokens: jwtSvc}
	auctionUC := &usecase.AuctionUseCase{Auctions: auctionRepo, Broadcaster: hub}
	lotUC := &usecase.LotUseCase{Lots: lotRepo, Broadcaster: hub}
	feeUC := &usecase.FeeUseCase{Fees: feeRepo}
	bidderUC := &usecase.BidderUseCase{Users: userRepo}
	onboardingUC := &usecase.OnboardingUseCase{Tenants: tenantRepo, Hasher: hasher}

	var objectStorage port.ObjectStorage
	serveLocalUploads := false
	switch cfg.StorageProvider {
	case "", "local":
		if err := os.MkdirAll(cfg.UploadsDir, 0o755); err != nil {
			log.Fatalf("failed to create uploads dir: %v", err)
		}
		objectStorage = &storage.LocalObjectStorage{
			BaseDir:       cfg.UploadsDir,
			BaseURLPrefix: "/uploads",
		}
		serveLocalUploads = true
	case "s3", "r2":
		if cfg.S3Bucket == "" {
			log.Fatalf("S3_BUCKET is required when STORAGE_PROVIDER=%s", cfg.StorageProvider)
		}
		s3Store, err := storage.NewS3ObjectStorage(ctx, storage.S3Config{
			Region:          cfg.S3Region,
			Bucket:          cfg.S3Bucket,
			Endpoint:        cfg.S3Endpoint,
			AccessKeyID:     cfg.S3AccessKeyID,
			SecretAccessKey: cfg.S3SecretAccessKey,
			PublicBaseURL:   cfg.S3PublicBaseURL,
			UsePathStyle:    cfg.S3UsePathStyle,
		})
		if err != nil {
			log.Fatalf("failed to initialize s3 storage: %v", err)
		}
		objectStorage = s3Store
	default:
		log.Fatalf("unsupported STORAGE_PROVIDER: %s", cfg.StorageProvider)
	}

	h := &httptransport.Handler{
		Auth:       authUC,
		Auction:    auctionUC,
		Lot:        lotUC,
		Fee:        feeUC,
		Bidder:     bidderUC,
		Onboarding: onboardingUC,
		Storage:    objectStorage,
	}
	tenantMW := middleware.TenantMiddleware{Tenants: tenantRepo}
	authMW := middleware.AuthMiddleware{Tokens: jwtSvc}

	router := httptransport.NewRouter(h, tenantMW, authMW, hub, cfg.PlatformAdminToken, cfg.UploadsDir, serveLocalUploads)
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
