package postgres

import (
	"context"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DemoRequestRepo struct {
	DB *pgxpool.Pool
}

func (r *DemoRequestRepo) Create(ctx context.Context, item *domain.DemoRequest) error {
	const q = `
INSERT INTO demo_requests (id, name, institution, email, message, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.DB.Exec(ctx, q, item.ID, item.Name, item.Institution, item.Email, item.Message, item.CreatedAt)
	return mapDBError(err, "Your request could not be submitted.")
}
