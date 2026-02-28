package middleware

import (
	"net/http"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	apperrors "github.com/dinesh-88/elevare-auction/internal/errors"
	errmiddleware "github.com/dinesh-88/elevare-auction/internal/http/middleware"
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
			errmiddleware.WriteError(w, apperrors.New(apperrors.ErrUnauthorized, "Authentication is required.", http.StatusUnauthorized))
			return
		}
		claims, err := m.Tokens.Parse(token)
		if err != nil {
			errmiddleware.WriteError(w, apperrors.New(apperrors.ErrUnauthorized, "Authentication is required.", http.StatusUnauthorized))
			return
		}
		tenant, ok := TenantFromContext(r.Context())
		if !ok || claims.TenantID != tenant.ID {
			errmiddleware.WriteError(w, apperrors.New(apperrors.ErrUnauthorized, "Authentication is required.", http.StatusUnauthorized))
			return
		}
		if m.Permissions != nil {
			role, err := m.Permissions.GetEffectiveRole(r.Context(), claims.TenantID, claims.UserID)
			if err != nil || role == "" {
				errmiddleware.WriteError(w, apperrors.New(apperrors.ErrForbidden, "You do not have permission to perform this action.", http.StatusForbidden))
				return
			}
			claims.Role = role
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
				errmiddleware.WriteError(w, apperrors.New(apperrors.ErrForbidden, "You do not have permission to perform this action.", http.StatusForbidden))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
