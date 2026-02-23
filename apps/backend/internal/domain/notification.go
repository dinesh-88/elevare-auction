package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationEventType string

const (
	NotificationEventOutbidded     NotificationEventType = "OUTBIDDED"
	NotificationEventLotWon        NotificationEventType = "LOT_WON"
	NotificationEventInvoiceIssued NotificationEventType = "INVOICE_ISSUED"
)

type NotificationChannel string

const (
	NotificationChannelEmail NotificationChannel = "EMAIL"
	NotificationChannelInApp NotificationChannel = "IN_APP"
)

type NotificationJobStatus string

const (
	NotificationJobStatusPending NotificationJobStatus = "PENDING"
	NotificationJobStatusSending NotificationJobStatus = "SENDING"
	NotificationJobStatusSent    NotificationJobStatus = "SENT"
	NotificationJobStatusFailed  NotificationJobStatus = "FAILED"
)

type NotificationEvent struct {
	TenantID         uuid.UUID
	EventType        NotificationEventType
	RecipientUserIDs []uuid.UUID
	Payload          map[string]any
}

type Notification struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	UserID    uuid.UUID
	Type      NotificationEventType
	Title     string
	Body      string
	Link      *string
	Metadata  json.RawMessage
	ReadAt    *time.Time
	CreatedAt time.Time
}

type NotificationJob struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	UserID    uuid.UUID
	Channel   NotificationChannel
	EventType NotificationEventType
	DedupeKey string
	Payload   json.RawMessage
	Status    NotificationJobStatus
	Attempts  int
	NextRunAt time.Time
	LastError *string
	CreatedAt time.Time
	SentAt    *time.Time
}

type NotificationPreference struct {
	TenantID     uuid.UUID
	UserID       uuid.UUID
	EventType    NotificationEventType
	EmailEnabled bool
	InAppEnabled bool
}
