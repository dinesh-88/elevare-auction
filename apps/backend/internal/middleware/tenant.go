package middleware

import (
	"github.com/dinesh-88/elevare-auction/internal/port"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type TenantMiddleware struct {
	Tenants port.TenantRepository
}

func (m TenantMiddleware) Resolve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
		if host == "" {
			host = strings.TrimSpace(r.Header.Get("X-Tenant-Host"))
		}
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
		host = strings.ToLower(strings.TrimSpace(host))

		// Dedicated API domains (e.g. api.elevareauctions.com) do not identify the tenant.
		// For browser requests, use Origin host as a tenant hint.
		if strings.HasPrefix(host, "api.") {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin != "" {
				if originURL, err := url.Parse(origin); err == nil {
					originHost := strings.ToLower(strings.TrimSpace(originURL.Hostname()))
					if originHost != "" && !strings.HasPrefix(originHost, "api.") {
						host = originHost
					}
				}
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
