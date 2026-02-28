package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	apperrors "github.com/dinesh-88/elevare-auction/internal/errors"
)

func TestWriteError_AppErrorReturnsStatusAndCode(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	rr := httptest.NewRecorder()
	rr.Header().Set("X-Request-ID", "req-123")

	WriteError(rr, apperrors.New(apperrors.ErrForbidden, "forbidden", http.StatusForbidden))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != apperrors.ErrForbidden {
		t.Fatalf("expected code %q, got %#v", apperrors.ErrForbidden, body["error"])
	}
	if body["message"] != "forbidden" {
		t.Fatalf("expected message to be preserved, got %#v", body["message"])
	}
	if body["request_id"] != "req-123" {
		t.Fatalf("expected request id, got %#v", body["request_id"])
	}
}

func TestWriteError_ProductionHidesDebugAndSensitiveDetails(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	rr := httptest.NewRecorder()

	WriteError(rr, errors.New(`pq: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "SQLSTATE") || strings.Contains(body, "duplicate key") || strings.Contains(body, "pq:") {
		t.Fatalf("sensitive database details leaked: %s", body)
	}
	if strings.Contains(body, `"debug"`) {
		t.Fatalf("debug details should not be present in production: %s", body)
	}
}

func TestWriteError_DevelopmentShowsDebug(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	rr := httptest.NewRecorder()

	WriteError(rr, errors.New("sqlstate test"))

	body := rr.Body.String()
	if !strings.Contains(body, `"debug":"sqlstate test"`) {
		t.Fatalf("expected debug details in development, got %s", body)
	}
}

func TestErrorMiddleware_RecoversFromPanic(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	handler := ErrorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	rr.Header().Set("X-Request-ID", "req-panic")

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	if strings.Contains(rr.Body.String(), "boom") {
		t.Fatalf("panic details leaked: %s", rr.Body.String())
	}
}

func TestWriteMessage_ProductionSanitizesSensitivePayload(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	rr := httptest.NewRecorder()

	WriteMessage(rr, http.StatusBadRequest, "ERROR: syntax error at or near SELECT (SQLSTATE 42601)")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "SQLSTATE") || strings.Contains(body, "syntax error") {
		t.Fatalf("sensitive validation detail leaked: %s", body)
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	_ = os.Unsetenv("APP_ENV")
	os.Exit(code)
}
