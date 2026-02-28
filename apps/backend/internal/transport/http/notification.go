package httptransport

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) ListMyNotifications(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeErrorText(w, "claims missing", http.StatusUnauthorized)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	cursorCreatedAt, cursorID, err := parseNotificationCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeErrorText(w, "invalid cursor", http.StatusBadRequest)
		return
	}

	rows, err := h.Notification.ListForUser(r.Context(), tenant.ID, claims.UserID, limit, cursorCreatedAt, cursorID)
	if err != nil {
		writeErrorText(w, err.Error(), http.StatusInternalServerError)
		return
	}
	nextCursor := ""
	if len(rows) == limit {
		last := rows[len(rows)-1]
		nextCursor = buildNotificationCursor(last.CreatedAt, last.ID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": rows, "next_cursor": nextCursor})
}

func (h *Handler) MarkMyNotificationRead(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeErrorText(w, "claims missing", http.StatusUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorText(w, "invalid notification id", http.StatusBadRequest)
		return
	}
	if err := h.Notification.MarkRead(r.Context(), tenant.ID, claims.UserID, id); err != nil {
		writeErrorText(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) MarkAllMyNotificationsRead(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeErrorText(w, "claims missing", http.StatusUnauthorized)
		return
	}
	if err := h.Notification.MarkAllRead(r.Context(), tenant.ID, claims.UserID); err != nil {
		writeErrorText(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetMyNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeErrorText(w, "claims missing", http.StatusUnauthorized)
		return
	}
	prefs, err := h.Notification.GetPreferences(r.Context(), tenant.ID, claims.UserID)
	if err != nil {
		writeErrorText(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": prefs})
}

func (h *Handler) UpdateMyNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeErrorText(w, "claims missing", http.StatusUnauthorized)
		return
	}
	var req struct {
		Items []struct {
			EventType    domain.NotificationEventType `json:"event_type"`
			EmailEnabled bool                         `json:"email_enabled"`
			InAppEnabled bool                         `json:"in_app_enabled"`
		} `json:"items"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErrorText(w, "invalid payload", http.StatusBadRequest)
		return
	}
	in := make([]usecase.NotificationPreferenceInput, 0, len(req.Items))
	for _, item := range req.Items {
		in = append(in, usecase.NotificationPreferenceInput{
			EventType:    item.EventType,
			EmailEnabled: item.EmailEnabled,
			InAppEnabled: item.InAppEnabled,
		})
	}
	if err := h.Notification.UpdatePreferences(r.Context(), tenant.ID, claims.UserID, in); err != nil {
		status := http.StatusInternalServerError
		if err == usecase.ErrValidation || err == usecase.ErrForbidden {
			status = http.StatusBadRequest
		}
		writeErrorText(w, err.Error(), status)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func buildNotificationCursor(createdAt time.Time, id uuid.UUID) string {
	raw := fmt.Sprintf("%d|%s", createdAt.UnixNano(), id.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func parseNotificationCursor(raw string) (*time.Time, *uuid.UUID, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, nil, err
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid cursor")
	}
	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, nil, err
	}
	t := time.Unix(0, nanos).UTC()
	return &t, &id, nil
}
