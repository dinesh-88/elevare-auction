package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type fakeEmailOutboxRepo struct {
	items map[uuid.UUID]*domain.EmailOutbox
}

func (f *fakeEmailOutboxRepo) Create(ctx context.Context, item *domain.EmailOutbox) error {
	if f.items == nil {
		f.items = map[uuid.UUID]*domain.EmailOutbox{}
	}
	copy := *item
	f.items[item.ID] = &copy
	return nil
}

func (f *fakeEmailOutboxRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailOutbox, error) {
	if f.items == nil || f.items[id] == nil {
		return nil, errors.New("not found")
	}
	copy := *f.items[id]
	return &copy, nil
}

func (f *fakeEmailOutboxRepo) MarkSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	item := f.items[id]
	item.Status = domain.EmailOutboxSent
	item.SentAt = &sentAt
	item.LastError = nil
	item.NextAttemptAt = nil
	return nil
}

func (f *fakeEmailOutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, attempts int, nextAttemptAt *time.Time, lastError string) error {
	item := f.items[id]
	item.Status = domain.EmailOutboxFailed
	item.AttemptCount = attempts
	item.NextAttemptAt = nextAttemptAt
	item.LastError = &lastError
	item.SentAt = nil
	return nil
}

func (f *fakeEmailOutboxRepo) ListRetryable(ctx context.Context, now time.Time, limit int) ([]domain.EmailOutbox, error) {
	out := make([]domain.EmailOutbox, 0)
	for _, item := range f.items {
		if item.Status != domain.EmailOutboxFailed || item.AttemptCount >= 5 || item.NextAttemptAt == nil || item.NextAttemptAt.After(now) {
			continue
		}
		out = append(out, *item)
	}
	return out, nil
}

func (f *fakeEmailOutboxRepo) CountByRecipientSince(ctx context.Context, tenantID uuid.UUID, toEmail string, since time.Time) (int64, error) {
	var count int64
	for _, item := range f.items {
		if item.TenantID != tenantID || item.CreatedAt.Before(since) || item.ToEmail != toEmail {
			continue
		}
		count++
	}
	return count, nil
}

func (f *fakeEmailOutboxRepo) CountByTenantSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int64, error) {
	var count int64
	for _, item := range f.items {
		if item.TenantID != tenantID || item.CreatedAt.Before(since) {
			continue
		}
		count++
	}
	return count, nil
}

func (f *fakeEmailOutboxRepo) ExistsRecentByScope(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, toEmail, kind string, entityType string, entityID *uuid.UUID, since time.Time) (bool, error) {
	for _, item := range f.items {
		if item.TenantID != tenantID || item.CreatedAt.Before(since) {
			continue
		}
		if item.Type == nil || *item.Type != kind || item.EntityType == nil || *item.EntityType != entityType {
			continue
		}
		if entityID == nil || item.EntityID == nil || *item.EntityID != *entityID {
			continue
		}
		if userID != nil && *userID != uuid.Nil {
			if item.UserID == nil || *item.UserID != *userID {
				continue
			}
		} else if item.ToEmail != toEmail {
			continue
		}
		return true, nil
	}
	return false, nil
}

type fakeEmailSender struct {
	err   error
	calls int
}

func (f *fakeEmailSender) Send(ctx context.Context, msg port.EmailMessage) error {
	f.calls++
	return f.err
}

func TestEmailOutboxServiceMarksSentOnSuccess(t *testing.T) {
	repo := &fakeEmailOutboxRepo{}
	sender := &fakeEmailSender{}
	svc := &EmailOutboxService{Outbox: repo, Mailer: sender, Provider: "resend"}
	item := &domain.EmailOutbox{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		ToEmail:   "a@example.com",
		Subject:   "Hello",
		HTMLBody:  "<p>Hi</p>",
		TextBody:  "Hi",
		Status:    domain.EmailOutboxPending,
		CreatedAt: time.Now().UTC(),
	}
	if err := repo.Create(context.Background(), item); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.ProcessOutboxEmail(context.Background(), item.ID); err != nil {
		t.Fatalf("process: %v", err)
	}
	got := repo.items[item.ID]
	if got.Status != domain.EmailOutboxSent || got.SentAt == nil {
		t.Fatalf("expected sent status, got %+v", got)
	}
}

func TestEmailOutboxServiceSchedulesRetryOnFailure(t *testing.T) {
	repo := &fakeEmailOutboxRepo{}
	sender := &fakeEmailSender{err: errors.New("boom")}
	svc := &EmailOutboxService{Outbox: repo, Mailer: sender, Provider: "resend"}
	item := &domain.EmailOutbox{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		ToEmail:   "a@example.com",
		Subject:   "Hello",
		HTMLBody:  "<p>Hi</p>",
		TextBody:  "Hi",
		Status:    domain.EmailOutboxPending,
		CreatedAt: time.Now().UTC(),
	}
	_ = repo.Create(context.Background(), item)

	if err := svc.ProcessOutboxEmail(context.Background(), item.ID); err != nil {
		t.Fatalf("process: %v", err)
	}
	got := repo.items[item.ID]
	if got.Status != domain.EmailOutboxFailed {
		t.Fatalf("expected failed status, got %s", got.Status)
	}
	if got.AttemptCount != 1 {
		t.Fatalf("expected attempts=1, got %d", got.AttemptCount)
	}
	if got.NextAttemptAt == nil {
		t.Fatal("expected retry to be scheduled")
	}
	if got.LastError == nil || *got.LastError != "boom" {
		t.Fatalf("expected last error, got %+v", got.LastError)
	}
}

func TestEmailOutboxServiceStopsAfterMaxAttempts(t *testing.T) {
	repo := &fakeEmailOutboxRepo{}
	sender := &fakeEmailSender{err: errors.New("boom")}
	svc := &EmailOutboxService{Outbox: repo, Mailer: sender, Provider: "resend"}
	item := &domain.EmailOutbox{
		ID:           uuid.New(),
		TenantID:     uuid.New(),
		ToEmail:      "a@example.com",
		Subject:      "Hello",
		HTMLBody:     "<p>Hi</p>",
		TextBody:     "Hi",
		Status:       domain.EmailOutboxFailed,
		AttemptCount: 4,
		CreatedAt:    time.Now().UTC(),
	}
	_ = repo.Create(context.Background(), item)

	if err := svc.ProcessOutboxEmail(context.Background(), item.ID); err != nil {
		t.Fatalf("process: %v", err)
	}
	got := repo.items[item.ID]
	if got.AttemptCount != 5 {
		t.Fatalf("expected attempts=5, got %d", got.AttemptCount)
	}
	if got.NextAttemptAt != nil {
		t.Fatal("expected no further retries")
	}
}

func TestEmailOutboxServiceProcessesRetries(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeEmailOutboxRepo{
		items: map[uuid.UUID]*domain.EmailOutbox{},
	}
	id := uuid.New()
	repo.items[id] = &domain.EmailOutbox{
		ID:            id,
		TenantID:      uuid.New(),
		ToEmail:       "a@example.com",
		Subject:       "Hello",
		HTMLBody:      "<p>Hi</p>",
		TextBody:      "Hi",
		Status:        domain.EmailOutboxFailed,
		AttemptCount:  1,
		NextAttemptAt: &now,
		CreatedAt:     now.Add(-time.Minute),
	}
	sender := &fakeEmailSender{}
	svc := &EmailOutboxService{Outbox: repo, Mailer: sender, Provider: "resend"}

	if err := svc.ProcessRetries(context.Background(), now, 10); err != nil {
		t.Fatalf("process retries: %v", err)
	}
	if sender.calls != 1 {
		t.Fatalf("expected one send call, got %d", sender.calls)
	}
	if repo.items[id].Status != domain.EmailOutboxSent {
		t.Fatalf("expected sent after retry, got %s", repo.items[id].Status)
	}
}

func TestEmailOutboxServiceDropsNonCriticalWhenRecipientLimitExceeded(t *testing.T) {
	tenantID := uuid.New()
	now := time.Now().UTC()
	repo := &fakeEmailOutboxRepo{items: map[uuid.UUID]*domain.EmailOutbox{}}
	for i := 0; i < 3; i++ {
		id := uuid.New()
		repo.items[id] = &domain.EmailOutbox{
			ID:        id,
			TenantID:  tenantID,
			ToEmail:   "a@example.com",
			Status:    domain.EmailOutboxSent,
			CreatedAt: now.Add(-time.Minute),
		}
	}
	svc := &EmailOutboxService{Outbox: repo, Mailer: &fakeEmailSender{}, Provider: "resend"}

	item, err := svc.Queue(context.Background(), QueueEmailInput{
		TenantID: tenantID,
		ToEmail:  "a@example.com",
		Subject:  "Hello",
		HTMLBody: "<p>Hi</p>",
		Type:     string(domain.NotificationEventOutbidded),
	})
	if err != nil {
		t.Fatalf("queue: %v", err)
	}
	if item != nil {
		t.Fatal("expected silent drop when recipient rate limit is exceeded")
	}
}

func TestEmailOutboxServiceDropsNonCriticalOnCooldown(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	entityID := uuid.New()
	kind := string(domain.NotificationEventOutbidded)
	entityType := "AUCTION"
	now := time.Now().UTC()
	repo := &fakeEmailOutboxRepo{items: map[uuid.UUID]*domain.EmailOutbox{}}
	repo.items[uuid.New()] = &domain.EmailOutbox{
		ID:         uuid.New(),
		TenantID:   tenantID,
		UserID:     &userID,
		ToEmail:    "a@example.com",
		Type:       &kind,
		EntityType: &entityType,
		EntityID:   &entityID,
		Status:     domain.EmailOutboxSent,
		CreatedAt:  now.Add(-30 * time.Second),
	}
	svc := &EmailOutboxService{
		Outbox: repo,
		Mailer: &fakeEmailSender{},
		Limits: EmailRateLimits{OutbidCooldown: time.Minute},
	}

	item, err := svc.Queue(context.Background(), QueueEmailInput{
		TenantID:   tenantID,
		UserID:     &userID,
		ToEmail:    "a@example.com",
		Subject:    "Hello",
		HTMLBody:   "<p>Hi</p>",
		Type:       kind,
		EntityType: entityType,
		EntityID:   &entityID,
	})
	if err != nil {
		t.Fatalf("queue: %v", err)
	}
	if item != nil {
		t.Fatal("expected cooldown drop")
	}
}

func TestEmailOutboxServiceAllowsCriticalUnderNormalRecipientLimit(t *testing.T) {
	tenantID := uuid.New()
	now := time.Now().UTC()
	repo := &fakeEmailOutboxRepo{items: map[uuid.UUID]*domain.EmailOutbox{}}
	for i := 0; i < 3; i++ {
		id := uuid.New()
		repo.items[id] = &domain.EmailOutbox{
			ID:        id,
			TenantID:  tenantID,
			ToEmail:   "a@example.com",
			Status:    domain.EmailOutboxSent,
			CreatedAt: now.Add(-time.Minute),
		}
	}
	svc := &EmailOutboxService{
		Outbox: repo,
		Mailer: &fakeEmailSender{},
		Limits: EmailRateLimits{CriticalRecipient10Min: 10},
	}

	item, err := svc.Queue(context.Background(), QueueEmailInput{
		TenantID: tenantID,
		ToEmail:  "a@example.com",
		Subject:  "Invoice",
		HTMLBody: "<p>Invoice</p>",
		Type:     string(domain.NotificationEventInvoiceIssued),
		Critical: true,
	})
	if err != nil {
		t.Fatalf("queue: %v", err)
	}
	if item == nil {
		t.Fatal("expected critical email to be accepted under elevated cap")
	}
}
