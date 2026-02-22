package authz

import (
	"context"
	"errors"
	"net/http"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
)

var ErrForbidden = errors.New("forbidden")

func Can(role domain.Role, permission domain.Permission) bool {
	return domain.HasPermission(role, permission)
}

func Authorize(ctx context.Context, permission domain.Permission) error {
	uc, ok := middleware.UserContextFromContext(ctx)
	if !ok || !Can(uc.Role, permission) {
		return ErrForbidden
	}
	return nil
}

func Require(permission domain.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uc, ok := middleware.UserContextFromContext(r.Context())
			if !ok || !Can(uc.Role, permission) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
