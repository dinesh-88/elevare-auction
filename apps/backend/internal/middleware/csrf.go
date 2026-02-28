package middleware

import (
	"net/http"
	"strings"
)

// RequireCSRFCookieHeader enforces double-submit CSRF for cookie-authenticated requests.
func RequireCSRFCookieHeader(csrfCookieName string) func(http.Handler) http.Handler {
	cookieName := strings.TrimSpace(csrfCookieName)
	if cookieName == "" {
		cookieName = "csrf_token"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
				c, err := r.Cookie(cookieName)
				if err != nil || strings.TrimSpace(c.Value) == "" {
					writeError(w, "missing csrf cookie", http.StatusForbidden)
					return
				}
				headerToken := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
				if headerToken == "" || headerToken != strings.TrimSpace(c.Value) {
					writeError(w, "invalid csrf token", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
