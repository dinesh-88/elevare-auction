package authz

import (
	"context"
	"errors"
	"net/http"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	apperrors "github.com/dinesh-88/elevare-auction/internal/errors"
	errmiddleware "github.com/dinesh-88/elevare-auction/internal/http/middleware"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

var ErrForbidden = errors.New("forbidden")
var permissionChecker port.PermissionRepository

func SetPermissionChecker(checker port.PermissionRepository) {
	permissionChecker = checker
}

func Can(role domain.Role, permission domain.Permission) bool {
	return domain.HasPermission(role, permission)
}

func HasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission domain.Permission) bool {
	if claims, ok := middleware.ClaimsFromContext(ctx); ok && claims.UserID == userID && claims.TenantID == tenantID {
		return Can(claims.Role, permission)
	}
	if permissionChecker != nil {
		ok, err := permissionChecker.HasPermission(ctx, tenantID, userID, permission)
		return err == nil && ok
	}
	return false
}

func ForbiddenError() error {
	return apperrors.New(apperrors.ErrForbidden, "You do not have permission to perform this action.", http.StatusForbidden)
}

func Authorize(ctx context.Context, permission domain.Permission) error {
	uc, ok := middleware.UserContextFromContext(ctx)
	if !ok || !HasPermission(ctx, uc.UserID, uc.TenantID, permission) {
		return ForbiddenError()
	}
	return nil
}

func Require(permission domain.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uc, ok := middleware.UserContextFromContext(r.Context())
			if !ok || !HasPermission(r.Context(), uc.UserID, uc.TenantID, permission) {
				errmiddleware.WriteError(w, ForbiddenError())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
