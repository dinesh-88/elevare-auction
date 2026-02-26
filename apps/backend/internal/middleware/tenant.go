package middleware

import (
	"github.com/dinesh-88/elevare-auction/internal/port"
	"net"
	"net/http"
	"strings"
)

type TenantMiddleware struct {
	Tenants port.TenantRepository
}

func (m TenantMiddleware) Resolve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
		if host == "" {
			host = r.Host
		}
		if strings.Contains(host, ",") {
			host = strings.TrimSpace(strings.Split(host, ",")[0])
		}
		if strings.Contains(host, ":") {
			cleanHost, _, err := net.SplitHostPort(host)
			if err == nil {
				host = cleanHost
			}
		}
		tenant, err := m.Tenants.GetByHost(r.Context(), host)
		if err != nil || tenant == nil {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithTenant(r.Context(), tenant)))
	})
}
