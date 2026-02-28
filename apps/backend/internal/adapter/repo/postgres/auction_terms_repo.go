package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuctionTermsRepo struct {
	DB *pgxpool.Pool
}

func (r *AuctionTermsRepo) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.AuctionTerms, error) {
	rows, err := r.DB.Query(ctx, `
SELECT id, tenant_id, auction_id, version_number, content_markdown, content_hash,
       is_active, created_by_user_id, created_at
FROM auction_terms
WHERE tenant_id = $1 AND auction_id = $2
ORDER BY version_number DESC, created_at DESC`, tenantID, auctionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.AuctionTerms, 0)
	for rows.Next() {
		item, err := scanAuctionTerms(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (r *AuctionTermsRepo) GetActiveByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.AuctionTerms, error) {
	item, err := scanAuctionTerms(r.DB.QueryRow(ctx, `
SELECT id, tenant_id, auction_id, version_number, content_markdown, content_hash,
       is_active, created_by_user_id, created_at
FROM auction_terms
WHERE tenant_id = $1 AND auction_id = $2 AND is_active = true
LIMIT 1`, tenantID, auctionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return item, err
}

func (r *AuctionTermsRepo) GetAcceptance(ctx context.Context, tenantID, auctionID, userID, auctionTermsID uuid.UUID) (*domain.BidderTermsAcceptance, error) {
	item, err := scanBidderTermsAcceptance(r.DB.QueryRow(ctx, `
SELECT id, tenant_id, auction_id, user_id, auction_terms_id, terms_version_number,
       content_hash, accepted_at, accepted_ip, accepted_user_agent, created_at
FROM bidder_terms_acceptance
WHERE tenant_id = $1 AND auction_id = $2 AND user_id = $3 AND auction_terms_id = $4
LIMIT 1`, tenantID, auctionID, userID, auctionTermsID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return item, err
}

func (r *AuctionTermsRepo) CreateVersion(ctx context.Context, item *domain.AuctionTerms) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT 1 FROM auctions WHERE tenant_id = $1 AND id = $2 FOR UPDATE`, item.TenantID, item.AuctionID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(MAX(version_number), 0) + 1
FROM auction_terms
WHERE tenant_id = $1 AND auction_id = $2`, item.TenantID, item.AuctionID).Scan(&item.VersionNumber); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO auction_terms (
  id, tenant_id, auction_id, version_number, content_markdown, content_hash,
  is_active, created_by_user_id, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, now()
)`, item.ID, item.TenantID, item.AuctionID, item.VersionNumber, item.ContentMarkdown, item.ContentHash, item.IsActive, item.CreatedByUserID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	item.CreatedAt = time.Now().UTC()
	return nil
}

func (r *AuctionTermsRepo) PublishVersion(ctx context.Context, tenantID, auctionID, auctionTermsID uuid.UUID) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS(
  SELECT 1
  FROM auction_terms
  WHERE tenant_id = $1 AND auction_id = $2 AND id = $3
)`, tenantID, auctionID, auctionTermsID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return errors.New("auction terms not found")
	}
	if _, err := tx.Exec(ctx, `
UPDATE auction_terms
SET is_active = false
WHERE tenant_id = $1 AND auction_id = $2`, tenantID, auctionID); err != nil {
		return err
	}
	ct, err := tx.Exec(ctx, `
UPDATE auction_terms
SET is_active = true
WHERE tenant_id = $1 AND auction_id = $2 AND id = $3`, tenantID, auctionID, auctionTermsID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("auction terms not found")
	}
	return tx.Commit(ctx)
}

func (r *AuctionTermsRepo) CreateAcceptance(ctx context.Context, item *domain.BidderTermsAcceptance) error {
	_, err := r.DB.Exec(ctx, `
INSERT INTO bidder_terms_acceptance (
  id, tenant_id, auction_id, user_id, auction_terms_id, terms_version_number,
  content_hash, accepted_at, accepted_ip, accepted_user_agent, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, $9, $10, $11
)
ON CONFLICT (auction_id, user_id, auction_terms_id) DO NOTHING`, item.ID, item.TenantID, item.AuctionID, item.UserID, item.AuctionTermsID, item.TermsVersionNumber, item.ContentHash, item.AcceptedAt, nullIfEmpty(item.AcceptedIP), nullIfEmpty(item.AcceptedUserAgent), item.CreatedAt)
	return err
}

func scanAuctionTerms(row pgx.Row) (*domain.AuctionTerms, error) {
	var item domain.AuctionTerms
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.AuctionID,
		&item.VersionNumber,
		&item.ContentMarkdown,
		&item.ContentHash,
		&item.IsActive,
		&item.CreatedByUserID,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanBidderTermsAcceptance(row pgx.Row) (*domain.BidderTermsAcceptance, error) {
	var item domain.BidderTermsAcceptance
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.AuctionID,
		&item.UserID,
		&item.AuctionTermsID,
		&item.TermsVersionNumber,
		&item.ContentHash,
		&item.AcceptedAt,
		&item.AcceptedIP,
		&item.AcceptedUserAgent,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
