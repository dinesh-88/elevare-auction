package middleware

import (
	"net/http"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
)

type AuthMiddleware struct {
	Tokens            port.TokenService
	AccessTokenCookie string
	Permissions       port.PermissionRepository
}

func (m AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""
		authz := strings.TrimSpace(r.Header.Get("Authorization"))
		parts := strings.SplitN(authz, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			token = strings.TrimSpace(parts[1])
		}
		if token == "" {
			cookieName := strings.TrimSpace(m.AccessTokenCookie)
			if cookieName == "" {
				cookieName = "access_token"
			}
			if c, err := r.Cookie(cookieName); err == nil {
				token = strings.TrimSpace(c.Value)
			}
		}
		if token == "" {
			writeError(w, "missing auth token", http.StatusUnauthorized)
			return
		}
		claims, err := m.Tokens.Parse(token)
		if err != nil {
			writeError(w, "invalid token", http.StatusUnauthorized)
			return
		}
		tenant, ok := TenantFromContext(r.Context())
		if !ok || claims.TenantID != tenant.ID {
			writeError(w, "cross-tenant token", http.StatusUnauthorized)
			return
		}
		if m.Permissions != nil {
			if role, err := m.Permissions.GetEffectiveRole(r.Context(), claims.TenantID, claims.UserID); err == nil && role != "" {
				claims.Role = role
			}
		}
		next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
	})
}

func RequireRole(role domain.Role) func(http.Handler) http.Handler {
	return RequireAnyRole(role)
}

func RequireAnyRole(roles ...domain.Role) func(http.Handler) http.Handler {
	allowed := make(map[domain.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				writeError(w, "forbidden", http.StatusForbidden)
				return
			}
			if _, exists := allowed[claims.Role]; !exists {
				writeError(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequirePermission(permission domain.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok || !domain.HasPermission(claims.Role, permission) {
				writeError(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
