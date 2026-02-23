package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepo struct {
	DB *pgxpool.Pool
}

func (r *NotificationRepo) Insert(ctx context.Context, tenantID, userID uuid.UUID, eventType domain.NotificationEventType, title, body string, link *string, metadata json.RawMessage) (uuid.UUID, error) {
	id := uuid.New()
	if len(metadata) == 0 {
		metadata = []byte(`{}`)
	}
	const q = `
INSERT INTO notifications (id, tenant_id, user_id, type, title, body, link, metadata)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	if _, err := r.DB.Exec(ctx, q, id, tenantID, userID, eventType, title, body, link, metadata); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *NotificationRepo) ListForUser(ctx context.Context, tenantID, userID uuid.UUID, limit int, cursorCreatedAt *time.Time, cursorID *uuid.UUID) ([]domain.Notification, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	args := []any{tenantID, userID}
	query := `
SELECT id, tenant_id, user_id, type, title, body, link, metadata, read_at, created_at
FROM notifications
WHERE tenant_id = $1
  AND user_id = $2`

	if cursorCreatedAt != nil && cursorID != nil {
		args = append(args, *cursorCreatedAt, *cursorID)
		query += fmt.Sprintf("\n  AND (created_at, id) < ($%d, $%d)", len(args)-1, len(args))
	}
	args = append(args, limit)
	query += fmt.Sprintf("\nORDER BY created_at DESC, id DESC\nLIMIT $%d", len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Notification, 0, limit)
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.TenantID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Link, &n.Metadata, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *NotificationRepo) MarkRead(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	const q = `
UPDATE notifications
SET read_at = COALESCE(read_at, NOW())
WHERE id = $1
  AND tenant_id = $2
  AND user_id = $3`
	ct, err := r.DB.Exec(ctx, q, notificationID, tenantID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("notification not found")
	}
	return nil
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) error {
	const q = `
UPDATE notifications
SET read_at = COALESCE(read_at, NOW())
WHERE tenant_id = $1
  AND user_id = $2`
	_, err := r.DB.Exec(ctx, q, tenantID, userID)
	return err
}

type NotificationJobsRepo struct {
	DB *pgxpool.Pool
}

func (r *NotificationJobsRepo) Enqueue(ctx context.Context, job domain.NotificationJob) (bool, error) {
	return r.enqueue(ctx, r.DB, job)
}

func (r *NotificationJobsRepo) EnqueueTx(ctx context.Context, tx pgx.Tx, job domain.NotificationJob) (bool, error) {
	return r.enqueue(ctx, tx, job)
}

func (r *NotificationJobsRepo) enqueue(ctx context.Context, execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}, job domain.NotificationJob) (bool, error) {
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	if len(job.Payload) == 0 {
		job.Payload = []byte(`{}`)
	}
	if job.Status == "" {
		job.Status = domain.NotificationJobStatusPending
	}
	if job.NextRunAt.IsZero() {
		job.NextRunAt = time.Now().UTC()
	}
	const q = `
INSERT INTO notification_jobs (
  id, tenant_id, user_id, channel, event_type, dedupe_key,
  payload, status, attempts, next_run_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
ON CONFLICT (dedupe_key) DO NOTHING`
	ct, err := execer.Exec(ctx, q,
		job.ID, job.TenantID, job.UserID, job.Channel, job.EventType, strings.TrimSpace(job.DedupeKey),
		job.Payload, job.Status, job.Attempts, job.NextRunAt,
	)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() == 1, nil
}

func (r *NotificationJobsRepo) FetchBatchForProcessing(ctx context.Context, limit int) ([]domain.NotificationJob, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const q = `
WITH picked AS (
  SELECT id
  FROM notification_jobs
  WHERE status = 'PENDING'
    AND next_run_at <= NOW()
    AND tenant_id IS NOT NULL
  ORDER BY created_at ASC
  LIMIT $1
  FOR UPDATE SKIP LOCKED
), updated AS (
  UPDATE notification_jobs j
  SET status = 'SENDING'
  FROM picked p
  WHERE j.id = p.id
  RETURNING j.id, j.tenant_id, j.user_id, j.channel, j.event_type,
            j.dedupe_key, j.payload, j.status, j.attempts,
            j.next_run_at, j.last_error, j.created_at, j.sent_at
)
SELECT id, tenant_id, user_id, channel, event_type,
       dedupe_key, payload, status, attempts,
       next_run_at, last_error, created_at, sent_at
FROM updated
ORDER BY created_at ASC`
	rows, err := tx.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]domain.NotificationJob, 0, limit)
	for rows.Next() {
		job, err := scanNotificationJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *NotificationJobsRepo) MarkSending(ctx context.Context, tenantID, jobID uuid.UUID) error {
	const q = `
UPDATE notification_jobs
SET status = 'SENDING'
WHERE id = $1
  AND tenant_id = $2`
	ct, err := r.DB.Exec(ctx, q, jobID, tenantID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("notification job not found")
	}
	return nil
}

func (r *NotificationJobsRepo) MarkSent(ctx context.Context, tenantID, jobID uuid.UUID) error {
	const q = `
UPDATE notification_jobs
SET status = 'SENT',
    sent_at = NOW(),
    last_error = NULL
WHERE id = $1
  AND tenant_id = $2`
	ct, err := r.DB.Exec(ctx, q, jobID, tenantID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("notification job not found")
	}
	return nil
}

func (r *NotificationJobsRepo) MarkFailedOrRetry(ctx context.Context, tenantID, jobID uuid.UUID, attempts int, nextRunAt time.Time, lastError string, finalFailed bool) error {
	status := domain.NotificationJobStatusPending
	if finalFailed {
		status = domain.NotificationJobStatusFailed
	}
	const q = `
UPDATE notification_jobs
SET status = $3,
    attempts = $4,
    next_run_at = $5,
    last_error = $6
WHERE id = $1
  AND tenant_id = $2`
	ct, err := r.DB.Exec(ctx, q, jobID, tenantID, status, attempts, nextRunAt, strings.TrimSpace(lastError))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("notification job not found")
	}
	return nil
}

func scanNotificationJob(row pgx.Row) (domain.NotificationJob, error) {
	var j domain.NotificationJob
	err := row.Scan(
		&j.ID,
		&j.TenantID,
		&j.UserID,
		&j.Channel,
		&j.EventType,
		&j.DedupeKey,
		&j.Payload,
		&j.Status,
		&j.Attempts,
		&j.NextRunAt,
		&j.LastError,
		&j.CreatedAt,
		&j.SentAt,
	)
	return j, err
}

type NotificationPreferenceRepo struct {
	DB *pgxpool.Pool
}

func (r *NotificationPreferenceRepo) Get(ctx context.Context, tenantID, userID uuid.UUID, eventType domain.NotificationEventType) (domain.NotificationPreference, error) {
	const q = `
SELECT tenant_id, user_id, event_type, email_enabled, in_app_enabled
FROM notification_preferences
WHERE tenant_id = $1
  AND user_id = $2
  AND event_type = $3`
	var pref domain.NotificationPreference
	err := r.DB.QueryRow(ctx, q, tenantID, userID, eventType).Scan(&pref.TenantID, &pref.UserID, &pref.EventType, &pref.EmailEnabled, &pref.InAppEnabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.NotificationPreference{
			TenantID:     tenantID,
			UserID:       userID,
			EventType:    eventType,
			EmailEnabled: true,
			InAppEnabled: true,
		}, nil
	}
	return pref, err
}

func (r *NotificationPreferenceRepo) Upsert(ctx context.Context, tenantID, userID uuid.UUID, eventType domain.NotificationEventType, emailEnabled, inAppEnabled bool) error {
	const q = `
INSERT INTO notification_preferences (tenant_id, user_id, event_type, email_enabled, in_app_enabled)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (tenant_id, user_id, event_type)
DO UPDATE SET
  email_enabled = EXCLUDED.email_enabled,
  in_app_enabled = EXCLUDED.in_app_enabled`
	_, err := r.DB.Exec(ctx, q, tenantID, userID, eventType, emailEnabled, inAppEnabled)
	return err
}

func (r *NotificationPreferenceRepo) List(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.NotificationPreference, error) {
	const q = `
SELECT tenant_id, user_id, event_type, email_enabled, in_app_enabled
FROM notification_preferences
WHERE tenant_id = $1
  AND user_id = $2`
	rows, err := r.DB.Query(ctx, q, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefs := make([]domain.NotificationPreference, 0, 3)
	for rows.Next() {
		var p domain.NotificationPreference
		if err := rows.Scan(&p.TenantID, &p.UserID, &p.EventType, &p.EmailEnabled, &p.InAppEnabled); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, rows.Err()
}
