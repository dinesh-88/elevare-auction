package middleware

import (
	"encoding/json"
	stderrors "errors"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	apperrors "github.com/dinesh-88/elevare-auction/internal/errors"
)

type errorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
	Debug     string `json:"debug,omitempty"`
}

func ErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC request_id=%s: %+v\n%s", requestIDFromWriter(w), rec, debug.Stack())
				WriteError(w, apperrors.New(apperrors.ErrInternal, "An unexpected error occurred.", http.StatusInternalServerError))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func WriteError(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if stderrors.As(err, &appErr) {
		writeAppError(w, appErr, err)
		return
	}

	log.Printf("UNHANDLED ERROR request_id=%s: %+v", requestIDFromWriter(w), err)
	writeAppError(w, apperrors.New(apperrors.ErrInternal, "An unexpected error occurred.", http.StatusInternalServerError), err)
}

func WriteMessage(w http.ResponseWriter, status int, message string) {
	writeAppError(w, appErrorFromStatus(status, message), nil)
}

func writeAppError(w http.ResponseWriter, appErr *apperrors.AppError, cause error) {
	if appErr == nil {
		appErr = apperrors.New(apperrors.ErrInternal, "An unexpected error occurred.", http.StatusInternalServerError)
	}
	if appErr.Status == 0 {
		appErr.Status = http.StatusInternalServerError
	}

	response := errorResponse{
		Error:     appErr.Code,
		Message:   sanitizeMessage(appErr.Status, appErr.Message),
		RequestID: requestIDFromWriter(w),
	}
	if isDevelopment() {
		debugMessage := strings.TrimSpace(appErr.Debug)
		if debugMessage == "" && cause != nil {
			debugMessage = strings.TrimSpace(cause.Error())
		}
		if debugMessage == "" && appErr.Message != response.Message {
			debugMessage = strings.TrimSpace(appErr.Message)
		}
		response.Debug = debugMessage
	}

	w.Header().Set("Content-Type", "application/json")
	if response.RequestID != "" {
		w.Header().Set("X-Request-ID", response.RequestID)
	}
	w.WriteHeader(appErr.Status)
	_ = json.NewEncoder(w).Encode(response)
}

func appErrorFromStatus(status int, message string) *apperrors.AppError {
	switch status {
	case http.StatusBadRequest:
		return apperrors.New(apperrors.ErrValidation, message, status)
	case http.StatusUnauthorized:
		return apperrors.New(apperrors.ErrUnauthorized, message, status)
	case http.StatusForbidden:
		return apperrors.New(apperrors.ErrForbidden, message, status)
	case http.StatusNotFound:
		return apperrors.New(apperrors.ErrNotFound, message, status)
	case http.StatusConflict:
		return apperrors.New(apperrors.ErrConflict, message, status)
	case http.StatusTooManyRequests:
		return apperrors.New(apperrors.ErrRateLimited, message, status)
	default:
		return apperrors.New(apperrors.ErrInternal, message, http.StatusInternalServerError)
	}
}

func sanitizeMessage(status int, message string) string {
	raw := strings.TrimSpace(message)
	if raw == "" {
		return defaultMessage(status)
	}
	if !isDevelopment() && (status >= 500 || looksSensitive(raw)) {
		return defaultMessage(status)
	}
	return raw
}

func defaultMessage(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "The request could not be processed."
	case http.StatusUnauthorized:
		return "Authentication is required."
	case http.StatusForbidden:
		return "You are not allowed to perform this action."
	case http.StatusNotFound:
		return "The requested resource was not found."
	case http.StatusConflict:
		return "The request conflicts with the current state."
	case http.StatusTooManyRequests:
		return "Too many requests. Please try again shortly."
	default:
		return "An unexpected error occurred."
	}
}

func looksSensitive(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	sensitiveMarkers := []string{
		"sqlstate",
		"syntax error",
		"pq:",
		"pgx",
		"constraint",
		"duplicate key",
		"violates",
		"panic",
		"stack trace",
		"driver:",
		"database:",
	}
	for _, marker := range sensitiveMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func isDevelopment() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "development")
}

func requestIDFromWriter(w http.ResponseWriter) string {
	return strings.TrimSpace(w.Header().Get("X-Request-ID"))
}
