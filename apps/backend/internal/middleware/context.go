package middleware

import (
	"context"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type contextKey string

const (
	TenantContextKey contextKey = "tenant"
	ClaimsContextKey contextKey = "claims"
	RequestMetaKey   contextKey = "request_meta"
)

type RequestMeta struct {
	RequestID string
	IP        string
	UserAgent string
}

type AuditContext struct {
	TenantID    *uuid.UUID
	ActorUserID *uuid.UUID
	ActorRole   string
	RequestID   string
	IP          string
	UserAgent   string
}

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

func WithRequestMeta(ctx context.Context, meta RequestMeta) context.Context {
	return context.WithValue(ctx, RequestMetaKey, meta)
}

func RequestMetaFromContext(ctx context.Context) (RequestMeta, bool) {
	meta, ok := ctx.Value(RequestMetaKey).(RequestMeta)
	return meta, ok
}

func AuditContextFromContext(ctx context.Context) AuditContext {
	out := AuditContext{ActorRole: "SYSTEM"}
	if tenant, ok := TenantFromContext(ctx); ok {
		tenantID := tenant.ID
		out.TenantID = &tenantID
	}
	if claims, ok := ClaimsFromContext(ctx); ok {
		userID := claims.UserID
		out.ActorUserID = &userID
		out.ActorRole = string(claims.Role)
	}
	if meta, ok := RequestMetaFromContext(ctx); ok {
		out.RequestID = meta.RequestID
		out.IP = meta.IP
		out.UserAgent = meta.UserAgent
	}
	return out
}

func UserContextFromContext(ctx context.Context) (*domain.UserContext, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return nil, false
	}
	return &domain.UserContext{
		TenantID: claims.TenantID,
		UserID:   claims.UserID,
		Role:     claims.Role,
	}, true
}
