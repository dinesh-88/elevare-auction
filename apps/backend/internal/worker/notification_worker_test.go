package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeWorkerJobsRepo struct {
	jobs            []domain.NotificationJob
	markSentCalls   int
	markFailedCalls int
	lastAttempts    int
	lastNextRunAt   time.Time
	lastFinalFailed bool
}

func (f *fakeWorkerJobsRepo) Enqueue(ctx context.Context, job domain.NotificationJob) (bool, error) {
	return false, nil
}
func (f *fakeWorkerJobsRepo) EnqueueTx(ctx context.Context, tx pgx.Tx, job domain.NotificationJob) (bool, error) {
	return false, nil
}
func (f *fakeWorkerJobsRepo) FetchBatchForProcessing(ctx context.Context, limit int) ([]domain.NotificationJob, error) {
	return f.jobs, nil
}
func (f *fakeWorkerJobsRepo) MarkSending(ctx context.Context, tenantID, jobID uuid.UUID) error {
	return nil
}
func (f *fakeWorkerJobsRepo) MarkSent(ctx context.Context, tenantID, jobID uuid.UUID) error {
	f.markSentCalls++
	return nil
}
func (f *fakeWorkerJobsRepo) MarkFailedOrRetry(ctx context.Context, tenantID, jobID uuid.UUID, attempts int, nextRunAt time.Time, lastError string, finalFailed bool) error {
	f.markFailedCalls++
	f.lastAttempts = attempts
	f.lastNextRunAt = nextRunAt
	f.lastFinalFailed = finalFailed
	return nil
}

type fakeWorkerUsersRepo struct {
	user *domain.User
}

func (f *fakeWorkerUsersRepo) Create(ctx context.Context, tenantID uuid.UUID, user *domain.User) error {
	return errors.New("unused")
}
func (f *fakeWorkerUsersRepo) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	if f.user == nil {
		return nil, errors.New("user not found")
	}
	return f.user, nil
}
func (f *fakeWorkerUsersRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	return nil, errors.New("unused")
}
func (f *fakeWorkerUsersRepo) GetBidders(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error) {
	return nil, errors.New("unused")
}
func (f *fakeWorkerUsersRepo) UpdateApproval(ctx context.Context, tenantID, userID uuid.UUID, status domain.ApprovalStatus) error {
	return errors.New("unused")
}
func (f *fakeWorkerUsersRepo) UpdateEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verified bool) error {
	return errors.New("unused")
}

type fakeEmailSender struct {
	err   error
	calls int
}

func (f *fakeEmailSender) Send(ctx context.Context, to, subject, html string) error {
	f.calls++
	return f.err
}

type fakeWorkerAuditRepo struct{}

func (f *fakeWorkerAuditRepo) Append(ctx context.Context, entry port.AuditEntry) error { return nil }
func (f *fakeWorkerAuditRepo) AppendTx(ctx context.Context, tx pgx.Tx, entry port.AuditEntry) error {
	return nil
}
func (f *fakeWorkerAuditRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	return nil, nil
}
func (f *fakeWorkerAuditRepo) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	return nil, nil
}

func TestNotificationWorker_RetryOnFailure(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	job := domain.NotificationJob{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    userID,
		Channel:   domain.NotificationChannelEmail,
		EventType: domain.NotificationEventOutbidded,
		Payload:   []byte(`{"lot_title":"Lot","amount":100,"link":"/lots/1"}`),
		Attempts:  0,
	}
	jobsRepo := &fakeWorkerJobsRepo{jobs: []domain.NotificationJob{job}}
	usersRepo := &fakeWorkerUsersRepo{user: &domain.User{ID: userID, TenantID: tenantID, Email: "a@example.com"}}
	sender := &fakeEmailSender{err: errors.New("smtp down")}
	w := &NotificationWorker{Jobs: jobsRepo, Users: usersRepo, Sender: sender, Audit: &fakeWorkerAuditRepo{}}

	if err := w.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce failed: %v", err)
	}
	if jobsRepo.markFailedCalls != 1 {
		t.Fatalf("expected 1 failed/retry update, got %d", jobsRepo.markFailedCalls)
	}
	if jobsRepo.lastAttempts != 1 {
		t.Fatalf("expected attempts=1, got %d", jobsRepo.lastAttempts)
	}
	if jobsRepo.lastFinalFailed {
		t.Fatal("expected retry state, got final failed")
	}
}

func TestNotificationWorker_MarkSentOnSuccess(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	job := domain.NotificationJob{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    userID,
		Channel:   domain.NotificationChannelEmail,
		EventType: domain.NotificationEventLotWon,
		Payload:   []byte(`{"lot_title":"Lot","hammer":100,"link":"/lots/1"}`),
		Attempts:  0,
	}
	jobsRepo := &fakeWorkerJobsRepo{jobs: []domain.NotificationJob{job}}
	usersRepo := &fakeWorkerUsersRepo{user: &domain.User{ID: userID, TenantID: tenantID, Email: "a@example.com"}}
	sender := &fakeEmailSender{}
	w := &NotificationWorker{Jobs: jobsRepo, Users: usersRepo, Sender: sender, Audit: &fakeWorkerAuditRepo{}}

	if err := w.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce failed: %v", err)
	}
	if sender.calls != 1 {
		t.Fatalf("expected sender called once, got %d", sender.calls)
	}
	if jobsRepo.markSentCalls != 1 {
		t.Fatalf("expected mark sent once, got %d", jobsRepo.markSentCalls)
	}
	if jobsRepo.markFailedCalls != 0 {
		t.Fatalf("expected no retries, got %d", jobsRepo.markFailedCalls)
	}
}

func TestRetryBackoffSchedule(t *testing.T) {
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 1, want: time.Minute},
		{attempt: 2, want: 5 * time.Minute},
		{attempt: 3, want: 15 * time.Minute},
		{attempt: 4, want: time.Hour},
		{attempt: 5, want: 6 * time.Hour},
	}
	for _, tc := range cases {
		if got := retryBackoff(tc.attempt); got != tc.want {
			t.Fatalf("attempt %d: expected %v, got %v", tc.attempt, tc.want, got)
		}
	}
}
