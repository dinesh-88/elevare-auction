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

	auditRepo := &postgres.AuditRepo{DB: db}
	tenantRepo := &postgres.TenantRepo{DB: db, Audit: auditRepo}
	userRepo := &postgres.UserRepo{DB: db}
	assetRepo := &postgres.AssetRepo{DB: db}
	demoRequestRepo := &postgres.DemoRequestRepo{DB: db}
	auctionRepo := &postgres.AuctionRepo{DB: db, Audit: auditRepo}
	lotRepo := &postgres.LotRepo{DB: db, Audit: auditRepo}
	feeRepo := &postgres.FeeRepo{DB: db}
	invoiceRepo := &postgres.InvoiceRepo{DB: db, Audit: auditRepo}
	notificationRepo := &postgres.NotificationRepo{DB: db}
	notificationJobsRepo := &postgres.NotificationJobsRepo{DB: db}
	notificationPrefsRepo := &postgres.NotificationPreferenceRepo{DB: db}
	bidderRepo := &postgres.BidderRepo{DB: db}

	hasher := &security.Argon2Hasher{}
	jwtSvc := &security.JWTService{Secret: []byte(cfg.JWTSecret), TTL: cfg.JWTTTL}
	hub := wsadapter.NewHub()

	authUC := &usecase.AuthUseCase{Users: userRepo, TenantLookup: userRepo, Hasher: hasher, Tokens: jwtSvc}
	auctionUC := &usecase.AuctionUseCase{Auctions: auctionRepo, Tenants: tenantRepo, Broadcaster: hub, Audit: auditRepo}
	notificationService := &usecase.NotificationService{
		Notifications: notificationRepo,
		Jobs:          notificationJobsRepo,
		Preferences:   notificationPrefsRepo,
		Audit:         auditRepo,
		Broadcaster:   hub,
	}
	notificationUC := &usecase.NotificationUseCase{
		Notifications: notificationRepo,
		Preferences:   notificationPrefsRepo,
	}
	lotUC := &usecase.LotUseCase{Lots: lotRepo, Auctions: auctionRepo, Bidders: bidderRepo, Broadcaster: hub, Notifier: notificationService, Audit: auditRepo}
	feeUC := &usecase.FeeUseCase{Fees: feeRepo, Audit: auditRepo}
	invoiceUC := &usecase.InvoiceUseCase{Invoices: invoiceRepo, Audit: auditRepo, Notifier: notificationService}
	bidderUC := &usecase.BidderUseCase{Users: userRepo, Bidders: bidderRepo, Auctions: auctionRepo, Audit: auditRepo}
	auditUC := &usecase.AuditUseCase{Audit: auditRepo}
	siteUC := &usecase.SiteUseCase{Tenants: tenantRepo}
	assetUC := &usecase.AssetUseCase{Assets: assetRepo}
	onboardingUC := &usecase.OnboardingUseCase{Tenants: tenantRepo, Hasher: hasher}
	demoRequestUC := &usecase.DemoRequestUseCase{Requests: demoRequestRepo}

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
		Auth:                authUC,
		Auction:             auctionUC,
		Lot:                 lotUC,
		Fee:                 feeUC,
		Invoice:             invoiceUC,
		Bidder:              bidderUC,
		Audit:               auditUC,
		Notification:        notificationUC,
		Site:                siteUC,
		Asset:               assetUC,
		Onboarding:          onboardingUC,
		DemoRequest:         demoRequestUC,
		Storage:             objectStorage,
		PreviewToken:        cfg.SitePreviewToken,
		JWTTTL:              cfg.JWTTTL,
		CookieSecure:        cfg.CookieSecure,
		AccessCookie:        cfg.AccessTokenCookie,
		CSRFCookie:          cfg.CSRFCookie,
		TenantAppScheme:     cfg.TenantAppScheme,
		TenantAppBaseDomain: cfg.TenantAppBaseDomain,
		TenantAppPort:       cfg.TenantAppPort,
	}
	tenantMW := middleware.TenantMiddleware{Tenants: tenantRepo}
	authMW := middleware.AuthMiddleware{Tokens: jwtSvc, AccessTokenCookie: cfg.AccessTokenCookie}

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
