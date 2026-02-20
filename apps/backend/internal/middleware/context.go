package middleware

import (
	"context"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
)

type contextKey string

const (
	TenantContextKey contextKey = "tenant"
	ClaimsContextKey contextKey = "claims"
)

func WithTenant(ctx context.Context, t *domain.Tenant) context.Context {
	return context.WithValue(ctx, TenantContextKey, t)
}

func TenantFromContext(ctx context.Context) (*domain.Tenant, bool) {
	t, ok := ctx.Value(TenantContextKey).(*domain.Tenant)
	return t, ok
}

func WithClaims(ctx context.Context, c *port.TokenClaims) context.Context {
	return context.WithValue(ctx, ClaimsContextKey, c)
}

func ClaimsFromContext(ctx context.Context) (*port.TokenClaims, bool) {
	c, ok := ctx.Value(ClaimsContextKey).(*port.TokenClaims)
	return c, ok
}
