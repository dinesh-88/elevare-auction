package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailOutboxRepo struct {
	DB *pgxpool.Pool
}

func (r *EmailOutboxRepo) Create(ctx context.Context, item *domain.EmailOutbox) error {
	const q = `
INSERT INTO email_outbox (
  id, tenant_id, user_id, to_email, subject, html_body, text_body, type, entity_type, entity_id, provider,
  status, attempt_count, last_error, next_attempt_at, created_at, sent_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,
  $12,$13,$14,$15,$16,$17
)`
	_, err := r.DB.Exec(ctx, q,
		item.ID, item.TenantID, item.UserID, item.ToEmail, item.Subject, item.HTMLBody, item.TextBody, item.Type, item.EntityType, item.EntityID, item.Provider,
		item.Status, item.AttemptCount, item.LastError, item.NextAttemptAt, item.CreatedAt, item.SentAt,
	)
	return err
}

func (r *EmailOutboxRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailOutbox, error) {
	const q = `
SELECT id, tenant_id, user_id, to_email, subject, html_body, text_body,
       type, entity_type, entity_id, provider,
       status, attempt_count, last_error, next_attempt_at, created_at, sent_at
FROM email_outbox
WHERE id = $1`
	item, err := scanEmailOutbox(r.DB.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("email outbox record not found")
	}
	return item, err
}

func (r *EmailOutboxRepo) MarkSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	const q = `
UPDATE email_outbox
SET status = 'sent',
    sent_at = $2,
    last_error = NULL,
    next_attempt_at = NULL
WHERE id = $1`
	ct, err := r.DB.Exec(ctx, q, id, sentAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("email outbox record not found")
	}
	return nil
}

func (r *EmailOutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, attempts int, nextAttemptAt *time.Time, lastError string) error {
	const q = `
UPDATE email_outbox
SET status = 'failed',
    attempt_count = $2,
    last_error = $3,
    next_attempt_at = $4,
    sent_at = NULL
WHERE id = $1`
	ct, err := r.DB.Exec(ctx, q, id, attempts, lastError, nextAttemptAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("email outbox record not found")
	}
	return nil
}

func (r *EmailOutboxRepo) ListRetryable(ctx context.Context, now time.Time, limit int) ([]domain.EmailOutbox, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	const q = `
SELECT id, tenant_id, user_id, to_email, subject, html_body, text_body,
       type, entity_type, entity_id, provider,
       status, attempt_count, last_error, next_attempt_at, created_at, sent_at
FROM email_outbox
WHERE status = 'failed'
  AND attempt_count < 5
  AND next_attempt_at IS NOT NULL
  AND next_attempt_at <= $1
ORDER BY next_attempt_at ASC, created_at ASC
LIMIT $2`
	rows, err := r.DB.Query(ctx, q, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.EmailOutbox, 0, limit)
	for rows.Next() {
		item, err := scanEmailOutbox(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (r *EmailOutboxRepo) CountByRecipientSince(ctx context.Context, tenantID uuid.UUID, toEmail string, since time.Time) (int64, error) {
	const q = `
SELECT COUNT(*)::bigint
FROM email_outbox
WHERE tenant_id = $1
  AND lower(to_email) = lower($2)
  AND created_at >= $3`
	var count int64
	err := r.DB.QueryRow(ctx, q, tenantID, toEmail, since).Scan(&count)
	return count, err
}

func (r *EmailOutboxRepo) CountByTenantSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int64, error) {
	const q = `
SELECT COUNT(*)::bigint
FROM email_outbox
WHERE tenant_id = $1
  AND created_at >= $2`
	var count int64
	err := r.DB.QueryRow(ctx, q, tenantID, since).Scan(&count)
	return count, err
}

func (r *EmailOutboxRepo) ExistsRecentByScope(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, toEmail, kind string, entityType string, entityID *uuid.UUID, since time.Time) (bool, error) {
	if strings.TrimSpace(kind) == "" || strings.TrimSpace(entityType) == "" || entityID == nil || *entityID == uuid.Nil {
		return false, nil
	}
	const qByUser = `
SELECT EXISTS(
  SELECT 1
  FROM email_outbox
  WHERE tenant_id = $1
    AND type = $2
    AND entity_type = $3
    AND entity_id = $4
    AND user_id = $5
    AND created_at >= $6
)`
	const qByEmail = `
SELECT EXISTS(
  SELECT 1
  FROM email_outbox
  WHERE tenant_id = $1
    AND type = $2
    AND entity_type = $3
    AND entity_id = $4
    AND lower(to_email) = lower($5)
    AND created_at >= $6
)`
	var exists bool
	var err error
	if userID != nil && *userID != uuid.Nil {
		err = r.DB.QueryRow(ctx, qByUser, tenantID, kind, entityType, *entityID, *userID, since).Scan(&exists)
	} else {
		err = r.DB.QueryRow(ctx, qByEmail, tenantID, kind, entityType, *entityID, toEmail, since).Scan(&exists)
	}
	return exists, err
}

func scanEmailOutbox(row pgx.Row) (*domain.EmailOutbox, error) {
	var item domain.EmailOutbox
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.UserID,
		&item.ToEmail,
		&item.Subject,
		&item.HTMLBody,
		&item.TextBody,
		&item.Type,
		&item.EntityType,
		&item.EntityID,
		&item.Provider,
		&item.Status,
		&item.AttemptCount,
		&item.LastError,
		&item.NextAttemptAt,
		&item.CreatedAt,
		&item.SentAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}
