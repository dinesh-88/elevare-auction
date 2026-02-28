package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	emailtpl "github.com/dinesh-88/elevare-auction/internal/email"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/dinesh-88/elevare-auction/internal/usecase"
)

type NotificationWorker struct {
	Jobs   port.NotificationJobRepository
	Users  port.UserRepository
	Sender port.EmailSender
	Audit  port.AuditRepository
	Batch  int
	Logger *log.Logger
}

func (w *NotificationWorker) Run(ctx context.Context, pollInterval time.Duration) error {
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.ProcessOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
				w.printf("notification worker tick failed: %v", err)
			}
		}
	}
}

func (w *NotificationWorker) ProcessOnce(ctx context.Context) error {
	batch := w.Batch
	if batch <= 0 {
		batch = 50
	}
	jobs, err := w.Jobs.FetchBatchForProcessing(ctx, batch)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if err := w.processJob(ctx, job); err != nil {
			w.printf("failed to process notification job %s: %v", job.ID, err)
		}
	}
	return nil
}

func (w *NotificationWorker) processJob(ctx context.Context, job domain.NotificationJob) error {
	attempts := job.Attempts + 1

	if job.Channel != domain.NotificationChannelEmail {
		if err := w.Jobs.MarkSent(ctx, job.TenantID, job.ID); err != nil {
			return err
		}
		w.appendAudit(ctx, job, "NOTIFICATION_SENT", attempts, nil)
		return nil
	}

	user, err := w.Users.GetByID(ctx, job.TenantID, job.UserID)
	if err != nil {
		return w.failOrRetry(ctx, job, attempts, "failed to load recipient")
	}

	payload := map[string]any{}
	if len(job.Payload) > 0 {
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return w.failOrRetry(ctx, job, attempts, "failed to decode payload")
		}
	}
	rendered, err := usecase.RenderEmailFull(job.EventType, user.Locale, emailtpl.BrandContext{TenantName: "Elevare"}, payload)
	if err != nil {
		return w.failOrRetry(ctx, job, attempts, err.Error())
	}
	if err := w.Sender.Send(ctx, port.EmailMessage{
		To:       user.Email,
		Subject:  rendered.Subject,
		HTMLBody: rendered.HTML,
		TextBody: rendered.Text,
	}); err != nil {
		return w.failOrRetry(ctx, job, attempts, err.Error())
	}
	if err := w.Jobs.MarkSent(ctx, job.TenantID, job.ID); err != nil {
		return err
	}
	w.appendAudit(ctx, job, "NOTIFICATION_SENT", attempts, nil)
	return nil
}

func (w *NotificationWorker) failOrRetry(ctx context.Context, job domain.NotificationJob, attempts int, message string) error {
	backoff := retryBackoff(attempts)
	finalFailed := attempts >= 5
	nextRunAt := time.Now().UTC().Add(backoff)
	if finalFailed {
		nextRunAt = time.Now().UTC()
	}
	if err := w.Jobs.MarkFailedOrRetry(ctx, job.TenantID, job.ID, attempts, nextRunAt, message, finalFailed); err != nil {
		return err
	}
	w.appendAudit(ctx, job, "NOTIFICATION_FAILED", attempts, &message)
	return nil
}

func retryBackoff(attempt int) time.Duration {
	schedule := []time.Duration{
		time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		time.Hour,
		6 * time.Hour,
	}
	if attempt <= 0 {
		return schedule[0]
	}
	if attempt > len(schedule) {
		return schedule[len(schedule)-1]
	}
	return schedule[attempt-1]
}

func (w *NotificationWorker) appendAudit(ctx context.Context, job domain.NotificationJob, action string, attempts int, errMsg *string) {
	if w.Audit == nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"job_id":     job.ID,
		"event_type": job.EventType,
		"attempts":   attempts,
		"error":      errMsg,
	})
	entry := port.AuditEntry{
		TenantID:   &job.TenantID,
		ActorRole:  "SYSTEM",
		Action:     action,
		EntityType: "NOTIFICATION_JOB",
		EntityID:   &job.ID,
		Metadata:   payload,
	}
	base := middleware.AuditContextFromContext(ctx)
	if entry.RequestID == "" {
		entry.RequestID = base.RequestID
	}
	if entry.IP == "" {
		entry.IP = base.IP
	}
	if entry.UserAgent == "" {
		entry.UserAgent = base.UserAgent
	}
	_ = w.Audit.Append(ctx, entry)
}

func (w *NotificationWorker) printf(format string, args ...any) {
	if w.Logger != nil {
		w.Logger.Printf(format, args...)
		return
	}
	log.Printf(format, args...)
}
