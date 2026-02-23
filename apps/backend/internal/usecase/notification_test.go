package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeNotificationRepo struct {
	inserted []domain.Notification
}

func (f *fakeNotificationRepo) Insert(ctx context.Context, tenantID, userID uuid.UUID, eventType domain.NotificationEventType, title, body string, link *string, metadata json.RawMessage) (uuid.UUID, error) {
	id := uuid.New()
	f.inserted = append(f.inserted, domain.Notification{ID: id, TenantID: tenantID, UserID: userID, Type: eventType, Title: title, Body: body, Link: link, Metadata: metadata})
	return id, nil
}
func (f *fakeNotificationRepo) ListForUser(ctx context.Context, tenantID, userID uuid.UUID, limit int, cursorCreatedAt *time.Time, cursorID *uuid.UUID) ([]domain.Notification, error) {
	return nil, nil
}
func (f *fakeNotificationRepo) MarkRead(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	return nil
}
func (f *fakeNotificationRepo) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}

type fakeNotificationJobsRepo struct {
	byDedupe map[string]domain.NotificationJob
}

func (f *fakeNotificationJobsRepo) Enqueue(ctx context.Context, job domain.NotificationJob) (bool, error) {
	if f.byDedupe == nil {
		f.byDedupe = map[string]domain.NotificationJob{}
	}
	if _, exists := f.byDedupe[job.DedupeKey]; exists {
		return false, nil
	}
	f.byDedupe[job.DedupeKey] = job
	return true, nil
}
func (f *fakeNotificationJobsRepo) EnqueueTx(ctx context.Context, tx pgx.Tx, job domain.NotificationJob) (bool, error) {
	return f.Enqueue(ctx, job)
}
func (f *fakeNotificationJobsRepo) FetchBatchForProcessing(ctx context.Context, limit int) ([]domain.NotificationJob, error) {
	return nil, nil
}
func (f *fakeNotificationJobsRepo) MarkSending(ctx context.Context, tenantID, jobID uuid.UUID) error {
	return nil
}
func (f *fakeNotificationJobsRepo) MarkSent(ctx context.Context, tenantID, jobID uuid.UUID) error {
	return nil
}
func (f *fakeNotificationJobsRepo) MarkFailedOrRetry(ctx context.Context, tenantID, jobID uuid.UUID, attempts int, nextRunAt time.Time, lastError string, finalFailed bool) error {
	return nil
}

type fakePreferenceRepo struct {
	prefs map[string]domain.NotificationPreference
}

func prefKey(tenantID, userID uuid.UUID, eventType domain.NotificationEventType) string {
	return tenantID.String() + ":" + userID.String() + ":" + string(eventType)
}

func (f *fakePreferenceRepo) Get(ctx context.Context, tenantID, userID uuid.UUID, eventType domain.NotificationEventType) (domain.NotificationPreference, error) {
	if f.prefs == nil {
		return domain.NotificationPreference{TenantID: tenantID, UserID: userID, EventType: eventType, EmailEnabled: true, InAppEnabled: true}, nil
	}
	if pref, ok := f.prefs[prefKey(tenantID, userID, eventType)]; ok {
		return pref, nil
	}
	return domain.NotificationPreference{TenantID: tenantID, UserID: userID, EventType: eventType, EmailEnabled: true, InAppEnabled: true}, nil
}
func (f *fakePreferenceRepo) Upsert(ctx context.Context, tenantID, userID uuid.UUID, eventType domain.NotificationEventType, emailEnabled, inAppEnabled bool) error {
	if f.prefs == nil {
		f.prefs = map[string]domain.NotificationPreference{}
	}
	f.prefs[prefKey(tenantID, userID, eventType)] = domain.NotificationPreference{TenantID: tenantID, UserID: userID, EventType: eventType, EmailEnabled: emailEnabled, InAppEnabled: inAppEnabled}
	return nil
}
func (f *fakePreferenceRepo) List(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.NotificationPreference, error) {
	out := make([]domain.NotificationPreference, 0)
	for _, pref := range f.prefs {
		if pref.TenantID == tenantID && pref.UserID == userID {
			out = append(out, pref)
		}
	}
	return out, nil
}

type fakeAuditSink struct{}

func (f *fakeAuditSink) Append(ctx context.Context, entry port.AuditEntry) error { return nil }
func (f *fakeAuditSink) AppendTx(ctx context.Context, tx pgx.Tx, entry port.AuditEntry) error {
	return nil
}
func (f *fakeAuditSink) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	return nil, errors.New("unused")
}
func (f *fakeAuditSink) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	return nil, errors.New("unused")
}

func TestNotificationEmit_DedupeEnqueueSameEvent(t *testing.T) {
	notifRepo := &fakeNotificationRepo{}
	jobRepo := &fakeNotificationJobsRepo{}
	prefRepo := &fakePreferenceRepo{}
	svc := &NotificationService{Notifications: notifRepo, Jobs: jobRepo, Preferences: prefRepo, Audit: &fakeAuditSink{}}
	tenantID := uuid.New()
	userID := uuid.New()
	event := domain.NotificationEvent{
		TenantID:         tenantID,
		EventType:        domain.NotificationEventOutbidded,
		RecipientUserIDs: []uuid.UUID{userID},
		Payload: map[string]any{
			"lot_title":  "Lot 1",
			"amount":     int64(1000),
			"bid_id":     uuid.New().String(),
			"link":       "/lots/abc",
			"dedupe_ref": "outbid:1",
		},
	}
	if err := svc.Emit(context.Background(), event); err != nil {
		t.Fatalf("first emit failed: %v", err)
	}
	if err := svc.Emit(context.Background(), event); err != nil {
		t.Fatalf("second emit failed: %v", err)
	}
	if len(jobRepo.byDedupe) != 1 {
		t.Fatalf("expected 1 deduped job, got %d", len(jobRepo.byDedupe))
	}
}

func TestNotificationEmit_DisabledEmailSkipsEmailJob(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	notifRepo := &fakeNotificationRepo{}
	jobRepo := &fakeNotificationJobsRepo{}
	prefRepo := &fakePreferenceRepo{prefs: map[string]domain.NotificationPreference{
		prefKey(tenantID, userID, domain.NotificationEventLotWon): {
			TenantID:     tenantID,
			UserID:       userID,
			EventType:    domain.NotificationEventLotWon,
			EmailEnabled: false,
			InAppEnabled: true,
		},
	}}
	svc := &NotificationService{Notifications: notifRepo, Jobs: jobRepo, Preferences: prefRepo}
	err := svc.Emit(context.Background(), domain.NotificationEvent{
		TenantID:         tenantID,
		EventType:        domain.NotificationEventLotWon,
		RecipientUserIDs: []uuid.UUID{userID},
		Payload: map[string]any{
			"lot_title": "Lot 9",
			"hammer":    int64(2500),
			"lot_id":    uuid.New().String(),
			"link":      "/lots/9",
		},
	})
	if err != nil {
		t.Fatalf("emit failed: %v", err)
	}
	if len(jobRepo.byDedupe) != 0 {
		t.Fatalf("expected no email jobs, got %d", len(jobRepo.byDedupe))
	}
	if len(notifRepo.inserted) != 1 {
		t.Fatalf("expected one in-app notification, got %d", len(notifRepo.inserted))
	}
}

func TestNotificationUseCase_TenantIsolation(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	nuc := &NotificationUseCase{Notifications: &fakeNotificationRepo{}, Preferences: &fakePreferenceRepo{}}
	ctx := middleware.WithClaims(context.Background(), &port.TokenClaims{TenantID: tenantA, UserID: uuid.New(), Role: domain.RoleBidder})
	_, err := nuc.ListForUser(ctx, tenantB, uuid.New(), 10, nil, nil)
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestRenderEmail_RequiresFieldsAndIncludesLink(t *testing.T) {
	_, _, err := RenderEmail(domain.NotificationEventInvoiceIssued, map[string]any{})
	if err == nil {
		t.Fatal("expected missing field error")
	}
	subject, html, err := RenderEmail(domain.NotificationEventInvoiceIssued, map[string]any{
		"invoice_number": "INV-001",
		"total_due":      int64(1200),
		"link":           "/invoices/1",
	})
	if err != nil {
		t.Fatalf("unexpected render error: %v", err)
	}
	if subject == "" || !strings.Contains(html, "/invoices/1") {
		t.Fatalf("expected subject and link in html, got subject=%q html=%q", subject, html)
	}
}

func TestOutbidNotificationsDoNotMentionProxy(t *testing.T) {
	subject, html, err := RenderEmail(domain.NotificationEventOutbidded, map[string]any{
		"lot_title": "Lot X",
		"amount":    int64(2000),
		"link":      "/lots/x",
	})
	if err != nil {
		t.Fatalf("unexpected render error: %v", err)
	}
	inAppTitle, inAppBody, err := renderInApp(domain.NotificationEventOutbidded, map[string]any{
		"lot_title": "Lot X",
		"amount":    int64(2000),
	})
	if err != nil {
		t.Fatalf("unexpected in-app render error: %v", err)
	}
	text := strings.ToLower(subject + " " + html + " " + inAppTitle + " " + inAppBody)
	if strings.Contains(text, "proxy") || strings.Contains(text, "auto bid") || strings.Contains(text, "autobid") {
		t.Fatalf("notification text must not mention proxy/auto bidding: %s", text)
	}
}
