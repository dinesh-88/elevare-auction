package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	emailadapter "github.com/dinesh-88/elevare-auction/internal/adapter/email"
	"github.com/dinesh-88/elevare-auction/internal/adapter/repo/postgres"
	"github.com/dinesh-88/elevare-auction/internal/bootstrap"
	"github.com/dinesh-88/elevare-auction/internal/transport/http"
	"github.com/dinesh-88/elevare-auction/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := httptransport.LoadConfig()
	if err := bootstrap.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	db, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer db.Close()

	auditRepo := &postgres.AuditRepo{DB: db}
	jobsRepo := &postgres.NotificationJobsRepo{DB: db}
	userRepo := &postgres.UserRepo{DB: db}

	sender := &emailadapter.ResendSender{
		APIKey: os.Getenv("RESEND_API_KEY"),
		From:   os.Getenv("RESEND_FROM_EMAIL"),
	}

	w := &worker.NotificationWorker{
		Jobs:   jobsRepo,
		Users:  userRepo,
		Sender: sender,
		Audit:  auditRepo,
		Batch:  50,
		Logger: log.Default(),
	}

	if err := w.Run(ctx, time.Second); err != nil && err != context.Canceled {
		log.Fatalf("worker stopped with error: %v", err)
	}
}
