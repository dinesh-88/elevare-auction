package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/port"
)

type TenantMiddleware struct {
	Tenants port.TenantRepository
}

func (m TenantMiddleware) Resolve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
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
