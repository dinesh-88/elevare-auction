package middleware

import (
	"net"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func CaptureRequestMeta(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		}
		meta := RequestMeta{
			RequestID: chimiddleware.GetReqID(r.Context()),
			IP:        ip,
			UserAgent: r.UserAgent(),
		}
		next.ServeHTTP(w, r.WithContext(WithRequestMeta(r.Context(), meta)))
	})
}
