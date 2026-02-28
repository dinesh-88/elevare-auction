package middleware

import (
	"net/http"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	apperrors "github.com/dinesh-88/elevare-auction/internal/errors"
	errmiddleware "github.com/dinesh-88/elevare-auction/internal/http/middleware"
)

func RequireOnboardingComplete() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := TenantFromContext(r.Context())
			if !ok || tenant == nil {
				errmiddleware.WriteError(w, apperrors.New(apperrors.ErrForbidden, "Complete onboarding before accessing this feature.", http.StatusForbidden))
				return
			}
			if tenant.OnboardingStatus != domain.OnboardingCompleted {
				errmiddleware.WriteError(w, apperrors.New("onboarding_required", "Complete onboarding before accessing this feature.", http.StatusForbidden))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
