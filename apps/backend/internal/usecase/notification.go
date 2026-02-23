package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type NotificationService struct {
	Notifications port.NotificationRepository
	Jobs          port.NotificationJobRepository
	Preferences   port.NotificationPreferenceRepository
	Audit         port.AuditRepository
	Broadcaster   port.UserNotificationBroadcaster
}

func (s *NotificationService) Emit(ctx context.Context, event domain.NotificationEvent) error {
	if event.TenantID == uuid.Nil || event.EventType == "" || len(event.RecipientUserIDs) == 0 {
		return ErrValidation
	}
	seenRecipients := make(map[uuid.UUID]struct{}, len(event.RecipientUserIDs))
	for _, userID := range event.RecipientUserIDs {
		if userID == uuid.Nil {
			continue
		}
		if _, seen := seenRecipients[userID]; seen {
			continue
		}
		seenRecipients[userID] = struct{}{}

		pref, err := s.Preferences.Get(ctx, event.TenantID, userID, event.EventType)
		if err != nil {
			return err
		}

		metadata, err := json.Marshal(event.Payload)
		if err != nil {
			return err
		}
		link, _ := event.Payload["link"].(string)
		var linkPtr *string
		if strings.TrimSpace(link) != "" {
			linkPtr = &link
		}

		if pref.InAppEnabled {
			title, body, err := renderInApp(event.EventType, event.Payload)
			if err != nil {
				return err
			}
			notificationID, err := s.Notifications.Insert(ctx, event.TenantID, userID, event.EventType, title, body, linkPtr, metadata)
			if err != nil {
				return err
			}
			if s.Broadcaster != nil {
				n := domain.Notification{
					ID:        notificationID,
					TenantID:  event.TenantID,
					UserID:    userID,
					Type:      event.EventType,
					Title:     title,
					Body:      body,
					Link:      linkPtr,
					Metadata:  metadata,
					CreatedAt: time.Now().UTC(),
				}
				s.Broadcaster.BroadcastUserNotification(ctx, event.TenantID, userID, n)
			}
		}

		if pref.EmailEnabled {
			job := domain.NotificationJob{
				ID:        uuid.New(),
				TenantID:  event.TenantID,
				UserID:    userID,
				Channel:   domain.NotificationChannelEmail,
				EventType: event.EventType,
				DedupeKey: buildDedupeKey(event, userID),
				Payload:   metadata,
				Status:    domain.NotificationJobStatusPending,
				Attempts:  0,
				NextRunAt: time.Now().UTC(),
			}
			inserted, err := s.Jobs.Enqueue(ctx, job)
			if err != nil {
				return err
			}
			if inserted && s.Audit != nil {
				auditCtx := middleware.AuditContextFromContext(ctx)
				jobID := job.ID
				payload, _ := json.Marshal(map[string]any{
					"job_id":     jobID,
					"event_type": event.EventType,
					"channel":    job.Channel,
					"recipient":  userID,
				})
				_ = s.Audit.Append(ctx, port.AuditEntry{
					TenantID:    &event.TenantID,
					ActorUserID: auditCtx.ActorUserID,
					ActorRole:   auditCtx.ActorRole,
					Action:      "NOTIFICATION_JOB_ENQUEUED",
					EntityType:  "NOTIFICATION_JOB",
					EntityID:    &jobID,
					RequestID:   auditCtx.RequestID,
					IP:          auditCtx.IP,
					UserAgent:   auditCtx.UserAgent,
					Metadata:    payload,
				})
			}
		}
	}
	return nil
}

func buildDedupeKey(event domain.NotificationEvent, userID uuid.UUID) string {
	seed := dedupeSeed(event, userID)
	sum := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(sum[:])
}

func dedupeSeed(event domain.NotificationEvent, userID uuid.UUID) string {
	ref := ""
	if v, ok := event.Payload["dedupe_ref"].(string); ok {
		ref = strings.TrimSpace(v)
	}
	if ref == "" {
		ref = fallbackDedupeRef(event.EventType, event.Payload)
	}
	return fmt.Sprintf("%s|%s|%s|%s|%s", event.TenantID, userID, domain.NotificationChannelEmail, event.EventType, ref)
}

func fallbackDedupeRef(eventType domain.NotificationEventType, payload map[string]any) string {
	switch eventType {
	case domain.NotificationEventOutbidded:
		if bidID, ok := payload["bid_id"].(string); ok {
			return "bid:" + bidID
		}
	case domain.NotificationEventLotWon:
		if lotID, ok := payload["lot_id"].(string); ok {
			return "lot:" + lotID
		}
	case domain.NotificationEventInvoiceIssued:
		if invoiceID, ok := payload["invoice_id"].(string); ok {
			return "invoice:" + invoiceID
		}
	}
	return "generic"
}

type NotificationUseCase struct {
	Notifications port.NotificationRepository
	Preferences   port.NotificationPreferenceRepository
}

type NotificationPreferenceInput struct {
	EventType    domain.NotificationEventType
	EmailEnabled bool
	InAppEnabled bool
}

func (uc *NotificationUseCase) ListForUser(ctx context.Context, tenantID, userID uuid.UUID, limit int, cursorCreatedAt *time.Time, cursorID *uuid.UUID) ([]domain.Notification, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	return uc.Notifications.ListForUser(ctx, tenantID, userID, limit, cursorCreatedAt, cursorID)
}

func (uc *NotificationUseCase) MarkRead(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	return uc.Notifications.MarkRead(ctx, tenantID, userID, notificationID)
}

func (uc *NotificationUseCase) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	return uc.Notifications.MarkAllRead(ctx, tenantID, userID)
}

func (uc *NotificationUseCase) GetPreferences(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.NotificationPreference, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	stored, err := uc.Preferences.List(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	byType := map[domain.NotificationEventType]domain.NotificationPreference{}
	for _, item := range stored {
		byType[item.EventType] = item
	}
	allowed := []domain.NotificationEventType{
		domain.NotificationEventOutbidded,
		domain.NotificationEventLotWon,
		domain.NotificationEventInvoiceIssued,
	}
	out := make([]domain.NotificationPreference, 0, len(allowed))
	for _, eventType := range allowed {
		if pref, ok := byType[eventType]; ok {
			out = append(out, pref)
			continue
		}
		out = append(out, domain.NotificationPreference{
			TenantID:     tenantID,
			UserID:       userID,
			EventType:    eventType,
			EmailEnabled: true,
			InAppEnabled: true,
		})
	}
	return out, nil
}

func (uc *NotificationUseCase) UpdatePreferences(ctx context.Context, tenantID, userID uuid.UUID, in []NotificationPreferenceInput) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	allowed := map[domain.NotificationEventType]struct{}{
		domain.NotificationEventOutbidded:     {},
		domain.NotificationEventLotWon:        {},
		domain.NotificationEventInvoiceIssued: {},
	}
	for _, item := range in {
		if _, ok := allowed[item.EventType]; !ok {
			return ErrValidation
		}
		if err := uc.Preferences.Upsert(ctx, tenantID, userID, item.EventType, item.EmailEnabled, item.InAppEnabled); err != nil {
			return err
		}
	}
	return nil
}
