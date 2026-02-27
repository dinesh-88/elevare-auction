package httptransport

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetSessionCookies_UsesSharedParentDomain(t *testing.T) {
	h := &Handler{
		JWTTTL:              30 * time.Minute,
		TenantAppBaseDomain: "elevareauctions.com",
	}
	rr := httptest.NewRecorder()

	h.setSessionCookies(rr, "token")

	cookies := rr.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}
	for _, cookie := range cookies {
		if cookie.Domain != "elevareauctions.com" {
			t.Fatalf("expected shared cookie domain, got %q", cookie.Domain)
		}
	}
}

func TestSetSessionCookies_OmitsDomainForLocalhost(t *testing.T) {
	h := &Handler{
		JWTTTL:              30 * time.Minute,
		TenantAppBaseDomain: "localhost",
	}
	rr := httptest.NewRecorder()

	h.setSessionCookies(rr, "token")

	for _, cookie := range rr.Result().Cookies() {
		if cookie.Domain != "" {
			t.Fatalf("expected host-only localhost cookie, got %q", cookie.Domain)
		}
	}
}
