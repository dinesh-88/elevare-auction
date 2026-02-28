package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type fakeAuthTokenService struct {
	claims *port.TokenClaims
	err    error
}

func (f *fakeAuthTokenService) Issue(userID, tenantID uuid.UUID, role domain.Role, locale domain.Locale) (string, error) {
	return "", errors.New("unused")
}

func (f *fakeAuthTokenService) Parse(token string) (*port.TokenClaims, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.claims, nil
}

type fakePermissionRepo struct {
	role domain.Role
	err  error
}

func (f *fakePermissionRepo) HasPermission(ctx context.Context, tenantID, userID uuid.UUID, permission domain.Permission) (bool, error) {
	return false, errors.New("unused")
}

func (f *fakePermissionRepo) GetEffectiveRole(ctx context.Context, tenantID, userID uuid.UUID) (domain.Role, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.role, nil
}

func TestRequireAuth_UnknownMembershipIsForbidden(t *testing.T) {
	tenantID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	req = req.WithContext(WithTenant(req.Context(), &domain.Tenant{ID: tenantID}))
	rr := httptest.NewRecorder()

	authMW := AuthMiddleware{
		Tokens: &fakeAuthTokenService{
			claims: &port.TokenClaims{
				UserID:   uuid.New(),
				TenantID: tenantID,
				Role:     domain.RoleOwner,
			},
		},
		Permissions: &fakePermissionRepo{err: errors.New("membership not found")},
	}

	authMW.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run")
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "forbidden" {
		t.Fatalf("expected forbidden code, got %#v", body["error"])
	}
	if body["message"] != "You do not have permission to perform this action." {
		t.Fatalf("expected standardized message, got %#v", body["message"])
	}
}

func TestRequireAuth_MissingSessionIsUnauthorized(t *testing.T) {
	tenantID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(WithTenant(req.Context(), &domain.Tenant{ID: tenantID}))
	rr := httptest.NewRecorder()

	authMW := AuthMiddleware{Tokens: &fakeAuthTokenService{}}

	authMW.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run")
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
