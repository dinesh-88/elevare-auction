package postgres

import (
	"context"
	"errors"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AssetRepo struct {
	DB *pgxpool.Pool
}

func (r *AssetRepo) Create(ctx context.Context, tenantID uuid.UUID, asset *domain.TenantAsset) error {
	if tenantID == uuid.Nil || asset == nil || asset.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	const q = `
INSERT INTO tenant_assets (id, tenant_id, kind, url, width, height, size_bytes, mime, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,now())`
	_, err := r.DB.Exec(ctx, q,
		asset.ID,
		asset.TenantID,
		asset.Kind,
		asset.URL,
		asset.Width,
		asset.Height,
		asset.SizeBytes,
		asset.MIME,
	)
	return mapDBError(err, "This asset could not be saved.")
}

func (r *AssetRepo) ListByKind(ctx context.Context, tenantID uuid.UUID, kind string, limit int) ([]domain.TenantAsset, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	query := `
SELECT id, tenant_id, kind, url, width, height, size_bytes, mime, created_at
FROM tenant_assets
WHERE tenant_id = $1`
	args := []any{tenantID}
	if kind != "" {
		query += ` AND kind = $2`
		args = append(args, kind)
		query += ` ORDER BY created_at DESC LIMIT $3`
		args = append(args, limit)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $2`
		args = append(args, limit)
	}
	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, mapDBError(err, "")
	}
	defer rows.Close()

	out := make([]domain.TenantAsset, 0)
	for rows.Next() {
		var item domain.TenantAsset
		if err := rows.Scan(&item.ID, &item.TenantID, &item.Kind, &item.URL, &item.Width, &item.Height, &item.SizeBytes, &item.MIME, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		return []domain.TenantAsset{}, nil
	}
	return out, nil
}
