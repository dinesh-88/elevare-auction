package postgres

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BidderRepo struct {
	DB *pgxpool.Pool
}

func (r *BidderRepo) EnsureTenantBidder(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantBidderProfile, error) {
	const q = `
INSERT INTO tenant_bidders (id, tenant_id, user_id, created_at, updated_at)
SELECT $3, $1, $2, now(), now()
WHERE EXISTS (
  SELECT 1 FROM users WHERE tenant_id = $1 AND id = $2 AND role = 'BIDDER'
)
ON CONFLICT (tenant_id, user_id) DO NOTHING
RETURNING id, tenant_id, user_id, phone, company_name, address_line1, address_line2, city, postal_code, country, status, risk_level, internal_notes, created_at, updated_at`
	item, err := scanTenantBidder(r.DB.QueryRow(ctx, q, tenantID, userID, uuid.New()))
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return r.GetTenantBidder(ctx, tenantID, userID)
}

func (r *BidderRepo) GetTenantBidder(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantBidderProfile, error) {
	const q = `
SELECT id, tenant_id, user_id, phone, company_name, address_line1, address_line2, city, postal_code, country, status, risk_level, internal_notes, created_at, updated_at
FROM tenant_bidders
WHERE tenant_id = $1 AND user_id = $2`
	item, err := scanTenantBidder(r.DB.QueryRow(ctx, q, tenantID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("bidder profile not found")
	}
	return item, err
}

func (r *BidderRepo) UpdateTenantBidder(ctx context.Context, tenantID, userID uuid.UUID, in port.UpdateTenantBidderInput) (*domain.TenantBidderProfile, error) {
	current, err := r.EnsureTenantBidder(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if in.Status != "" {
		current.Status = in.Status
	}
	if in.RiskLevel != "" {
		current.RiskLevel = in.RiskLevel
	}
	if in.InternalNotes != nil {
		current.InternalNotes = in.InternalNotes
	}
	if in.Phone != nil {
		current.Phone = in.Phone
	}
	if in.CompanyName != nil {
		current.CompanyName = in.CompanyName
	}
	if in.AddressLine1 != nil {
		current.AddressLine1 = in.AddressLine1
	}
	if in.AddressLine2 != nil {
		current.AddressLine2 = in.AddressLine2
	}
	if in.City != nil {
		current.City = in.City
	}
	if in.PostalCode != nil {
		current.PostalCode = in.PostalCode
	}
	if in.Country != nil {
		current.Country = in.Country
	}
	const q = `
UPDATE tenant_bidders
SET phone = $3,
    company_name = $4,
    address_line1 = $5,
    address_line2 = $6,
    city = $7,
    postal_code = $8,
    country = $9,
    status = $10,
    risk_level = $11,
    internal_notes = $12,
    updated_at = now()
WHERE tenant_id = $1 AND user_id = $2
RETURNING id, tenant_id, user_id, phone, company_name, address_line1, address_line2, city, postal_code, country, status, risk_level, internal_notes, created_at, updated_at`
	return scanTenantBidder(r.DB.QueryRow(ctx, q, tenantID, userID, current.Phone, current.CompanyName, current.AddressLine1, current.AddressLine2, current.City, current.PostalCode, current.Country, current.Status, current.RiskLevel, current.InternalNotes))
}

func (r *BidderRepo) ListUserAuctionParticipations(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.BidderAuctionParticipation, error) {
	const q = `
SELECT a.id, a.title, a.status, aba.status, a.start_time, a.end_time, aba.created_at, aba.approved_at
FROM auction_bidder_approvals aba
JOIN auctions a ON a.id = aba.auction_id AND a.tenant_id = aba.tenant_id
WHERE aba.tenant_id = $1
  AND aba.user_id = $2
ORDER BY a.end_time DESC, a.created_at DESC`
	rows, err := r.DB.Query(ctx, q, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.BidderAuctionParticipation, 0)
	for rows.Next() {
		var row domain.BidderAuctionParticipation
		if err := rows.Scan(
			&row.AuctionID,
			&row.Title,
			&row.Status,
			&row.ApprovalStatus,
			&row.StartTime,
			&row.EndTime,
			&row.RequestedAt,
			&row.ApprovedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *BidderRepo) RequestAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID, now time.Time) (*domain.AuctionBidderApproval, bool, error) {
	const q = `
INSERT INTO auction_bidder_approvals (
  id, tenant_id, auction_id, user_id, status, risk_level, deposit_required, deposit_status, created_at, updated_at
)
SELECT $4, $1, $2, $3, 'PENDING', 'low', false, 'not_required', $5, $5
WHERE EXISTS (SELECT 1 FROM auctions WHERE tenant_id = $1 AND id = $2)
  AND EXISTS (SELECT 1 FROM users WHERE tenant_id = $1 AND id = $3 AND role = 'BIDDER')
ON CONFLICT (auction_id, user_id) DO NOTHING
RETURNING id, tenant_id, auction_id, user_id, status, risk_level, max_bid_limit_minor, deposit_required,
          deposit_amount_minor, deposit_status, internal_notes, decided_by, decided_at,
          approved_by_user_id, approved_at, created_at, updated_at`
	item, err := scanAuctionBidderApproval(r.DB.QueryRow(ctx, q, tenantID, auctionID, userID, uuid.New(), now))
	if err == nil {
		return item, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}
	current, err := r.GetAuctionApproval(ctx, tenantID, auctionID, userID)
	if err != nil {
		return nil, false, err
	}
	return current, false, nil
}

func (r *BidderRepo) GetAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID) (*domain.AuctionBidderApproval, error) {
	const q = `
SELECT id, tenant_id, auction_id, user_id, status, risk_level, max_bid_limit_minor, deposit_required,
       deposit_amount_minor, deposit_status, internal_notes, decided_by, decided_at,
       approved_by_user_id, approved_at, created_at, updated_at
FROM auction_bidder_approvals
WHERE tenant_id = $1 AND auction_id = $2 AND user_id = $3`
	item, err := scanAuctionBidderApproval(r.DB.QueryRow(ctx, q, tenantID, auctionID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("approval not found")
	}
	return item, err
}

func (r *BidderRepo) SetAuctionApproval(ctx context.Context, tenantID, auctionID, userID, decidedBy uuid.UUID, status domain.ApprovalStatus, now time.Time) (*domain.AuctionBidderApproval, error) {
	var approvedBy *uuid.UUID
	var approvedAt *time.Time
	if status == domain.ApprovalApproved {
		approvedBy = &decidedBy
		approvedAt = &now
	}
	const q = `
INSERT INTO auction_bidder_approvals (
  id, tenant_id, auction_id, user_id, status, risk_level, deposit_required, deposit_status,
  decided_by, decided_at, approved_by_user_id, approved_at, created_at, updated_at
)
SELECT $6, $1, $2, $3, $4, 'low', false, 'not_required', $5, $7, $8, $9, $7, $7
WHERE EXISTS (SELECT 1 FROM auctions WHERE tenant_id = $1 AND id = $2)
  AND EXISTS (SELECT 1 FROM users WHERE tenant_id = $1 AND id = $3 AND role = 'BIDDER')
ON CONFLICT (auction_id, user_id) DO UPDATE
SET status = EXCLUDED.status,
    decided_by = EXCLUDED.decided_by,
    decided_at = EXCLUDED.decided_at,
    approved_by_user_id = EXCLUDED.approved_by_user_id,
    approved_at = EXCLUDED.approved_at,
    updated_at = EXCLUDED.updated_at
RETURNING id, tenant_id, auction_id, user_id, status, risk_level, max_bid_limit_minor, deposit_required,
          deposit_amount_minor, deposit_status, internal_notes, decided_by, decided_at,
          approved_by_user_id, approved_at, created_at, updated_at`
	item, err := scanAuctionBidderApproval(r.DB.QueryRow(ctx, q, tenantID, auctionID, userID, status, decidedBy, uuid.New(), now, approvedBy, approvedAt))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("auction or bidder not found")
	}
	return item, err
}

func (r *BidderRepo) UpdateAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID, in port.UpdateAuctionBidderApprovalInput, now time.Time) (*domain.AuctionBidderApproval, error) {
	const q = `
INSERT INTO auction_bidder_approvals (
  id, tenant_id, auction_id, user_id, status, risk_level, max_bid_limit_minor,
  deposit_required, deposit_amount_minor, deposit_status, internal_notes, created_at, updated_at
)
SELECT $4, $1, $2, $3, 'PENDING', $5, $6, $7, $8, $9, $10, $11, $11
WHERE EXISTS (SELECT 1 FROM auctions WHERE tenant_id = $1 AND id = $2)
  AND EXISTS (SELECT 1 FROM users WHERE tenant_id = $1 AND id = $3 AND role = 'BIDDER')
ON CONFLICT (auction_id, user_id) DO UPDATE
SET risk_level = EXCLUDED.risk_level,
    max_bid_limit_minor = EXCLUDED.max_bid_limit_minor,
    deposit_required = EXCLUDED.deposit_required,
    deposit_amount_minor = EXCLUDED.deposit_amount_minor,
    deposit_status = EXCLUDED.deposit_status,
    internal_notes = EXCLUDED.internal_notes,
    updated_at = EXCLUDED.updated_at
RETURNING id, tenant_id, auction_id, user_id, status, risk_level, max_bid_limit_minor, deposit_required,
          deposit_amount_minor, deposit_status, internal_notes, decided_by, decided_at,
          approved_by_user_id, approved_at, created_at, updated_at`
	item, err := scanAuctionBidderApproval(r.DB.QueryRow(
		ctx,
		q,
		tenantID,
		auctionID,
		userID,
		uuid.New(),
		in.RiskLevel,
		in.MaxBidLimitMinor,
		in.DepositRequired,
		in.DepositAmountMinor,
		in.DepositStatus,
		in.InternalNotes,
		now,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("auction or bidder not found")
	}
	return item, err
}

func (r *BidderRepo) ListAuctionBidders(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.ListAuctionBiddersFilter) ([]domain.AuctionBidderAdminRow, error) {
	q := `
SELECT a.user_id, u.email, a.status,
       COALESCE(tb.status, 'ACTIVE') AS tenant_status,
       a.risk_level,
       a.max_bid_limit_minor,
       a.deposit_required,
       a.deposit_amount_minor,
       a.deposit_status,
       a.internal_notes,
       a.created_at, a.decided_at
FROM auction_bidder_approvals a
JOIN users u ON u.id = a.user_id AND u.tenant_id = a.tenant_id
LEFT JOIN tenant_bidders tb ON tb.tenant_id = a.tenant_id AND tb.user_id = a.user_id
WHERE a.tenant_id = $1 AND a.auction_id = $2`
	args := []any{tenantID, auctionID}
	nextArg := 3
	if filter.Status != nil {
		q += ` AND a.status = $` + strconv.Itoa(nextArg)
		args = append(args, *filter.Status)
		nextArg++
	}
	if filter.RiskLevel != nil {
		q += ` AND a.risk_level = $` + strconv.Itoa(nextArg)
		args = append(args, *filter.RiskLevel)
		nextArg++
	}
	if filter.DepositRequired != nil {
		q += ` AND a.deposit_required = $` + strconv.Itoa(nextArg)
		args = append(args, *filter.DepositRequired)
		nextArg++
	}
	if filter.DepositStatus != nil {
		q += ` AND a.deposit_status = $` + strconv.Itoa(nextArg)
		args = append(args, *filter.DepositStatus)
	}
	q += ` ORDER BY a.created_at DESC`
	rows, err := r.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.AuctionBidderAdminRow, 0)
	for rows.Next() {
		var row domain.AuctionBidderAdminRow
		if err := rows.Scan(
			&row.UserID,
			&row.Email,
			&row.Approval,
			&row.TenantStatus,
			&row.RiskLevel,
			&row.MaxBidLimitMinor,
			&row.DepositRequired,
			&row.DepositAmountMinor,
			&row.DepositStatus,
			&row.InternalNotes,
			&row.RequestedAt,
			&row.DecidedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func scanTenantBidder(row pgx.Row) (*domain.TenantBidderProfile, error) {
	var item domain.TenantBidderProfile
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.UserID,
		&item.Phone,
		&item.CompanyName,
		&item.AddressLine1,
		&item.AddressLine2,
		&item.City,
		&item.PostalCode,
		&item.Country,
		&item.Status,
		&item.RiskLevel,
		&item.InternalNotes,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanAuctionBidderApproval(row pgx.Row) (*domain.AuctionBidderApproval, error) {
	var item domain.AuctionBidderApproval
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.AuctionID,
		&item.UserID,
		&item.Status,
		&item.RiskLevel,
		&item.MaxBidLimitMinor,
		&item.DepositRequired,
		&item.DepositAmountMinor,
		&item.DepositStatus,
		&item.InternalNotes,
		&item.DecidedBy,
		&item.DecidedAt,
		&item.ApprovedByUserID,
		&item.ApprovedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}
