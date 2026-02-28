package usecase

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type EmailOutboxService struct {
	Outbox   port.EmailOutboxRepository
	Mailer   port.EmailSender
	Provider string
	Limits   EmailRateLimits
	Logger   *log.Logger
}

type EmailRateLimits struct {
	Recipient10Min         int
	Recipient24H           int
	Tenant1H               int
	OutbidCooldown         time.Duration
	LotClosingCooldown     time.Duration
	CriticalRecipient10Min int
	CriticalRecipient24H   int
	CriticalTenant1H       int
}

type QueueEmailInput struct {
	TenantID   uuid.UUID
	UserID     *uuid.UUID
	ToEmail    string
	Subject    string
	HTMLBody   string
	TextBody   string
	Type       string
	EntityType string
	EntityID   *uuid.UUID
	Critical   bool
}

func (s *EmailOutboxService) Queue(ctx context.Context, in QueueEmailInput) (*domain.EmailOutbox, error) {
	if s.Outbox == nil || s.Mailer == nil {
		return nil, ErrValidation
	}
	toEmail := strings.TrimSpace(in.ToEmail)
	subject := strings.TrimSpace(in.Subject)
	htmlBody := strings.TrimSpace(in.HTMLBody)
	if in.TenantID == uuid.Nil || toEmail == "" || subject == "" || htmlBody == "" {
		return nil, ErrValidation
	}
	textBody := strings.TrimSpace(in.TextBody)
	if textBody == "" {
		textBody = plainTextFromHTML(htmlBody)
	}
	var provider *string
	if name := strings.TrimSpace(s.Provider); name != "" {
		provider = &name
	}
	allowed, err := s.rateLimitCheck(ctx, in, toEmail)
	if err != nil {
		return nil, err
	}
	if !allowed {
		s.printf("email outbox dropped by rate limit for %s type=%s", toEmail, strings.TrimSpace(in.Type))
		return nil, nil
	}
	kind := trimmedStringPtr(in.Type)
	entityType := trimmedStringPtr(in.EntityType)
	item := &domain.EmailOutbox{
		ID:           uuid.New(),
		TenantID:     in.TenantID,
		UserID:       in.UserID,
		ToEmail:      toEmail,
		Subject:      subject,
		HTMLBody:     htmlBody,
		TextBody:     textBody,
		Type:         kind,
		EntityType:   entityType,
		EntityID:     normalizeUUIDPtr(in.EntityID),
		Provider:     provider,
		Status:       domain.EmailOutboxPending,
		AttemptCount: 0,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.Outbox.Create(ctx, item); err != nil {
		return nil, err
	}
	go s.processAsync(item.ID)
	return item, nil
}

func (s *EmailOutboxService) ProcessOutboxEmail(ctx context.Context, id uuid.UUID) error {
	if s.Outbox == nil || s.Mailer == nil {
		return nil
	}
	item, err := s.Outbox.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if item.Status == domain.EmailOutboxSent {
		return nil
	}
	if item.AttemptCount >= 5 {
		return nil
	}
	if err := s.Mailer.Send(ctx, port.EmailMessage{
		To:       item.ToEmail,
		Subject:  item.Subject,
		HTMLBody: item.HTMLBody,
		TextBody: item.TextBody,
	}); err != nil {
		return s.markFailure(ctx, item, err.Error())
	}
	return s.Outbox.MarkSent(ctx, item.ID, time.Now().UTC())
}

func (s *EmailOutboxService) ProcessRetries(ctx context.Context, now time.Time, limit int) error {
	if s.Outbox == nil || s.Mailer == nil {
		return nil
	}
	items, err := s.Outbox.ListRetryable(ctx, now, limit)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := s.ProcessOutboxEmail(ctx, item.ID); err != nil {
			s.printf("email retry failed for %s: %v", item.ID, err)
		}
	}
	return nil
}

func (s *EmailOutboxService) markFailure(ctx context.Context, item *domain.EmailOutbox, message string) error {
	attempts := item.AttemptCount + 1
	nextAttemptAt := nextEmailRetry(time.Now().UTC(), attempts)
	return s.Outbox.MarkFailed(ctx, item.ID, attempts, nextAttemptAt, message)
}

func nextEmailRetry(now time.Time, attempts int) *time.Time {
	var delay time.Duration
	switch attempts {
	case 1:
		delay = 5 * time.Minute
	case 2:
		delay = 15 * time.Minute
	case 3:
		delay = time.Hour
	case 4:
		delay = 6 * time.Hour
	default:
		return nil
	}
	next := now.Add(delay)
	return &next
}

func (s *EmailOutboxService) processAsync(id uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.ProcessOutboxEmail(ctx, id); err != nil {
		s.printf("email outbox async send failed for %s: %v", id, err)
	}
}

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func plainTextFromHTML(html string) string {
	text := htmlTagPattern.ReplaceAllString(html, " ")
	return strings.Join(strings.Fields(text), " ")
}

func (s *EmailOutboxService) printf(format string, args ...any) {
	if s.Logger != nil {
		s.Logger.Printf(format, args...)
		return
	}
	log.Printf(format, args...)
}

func (s *EmailOutboxService) rateLimitCheck(ctx context.Context, in QueueEmailInput, toEmail string) (bool, error) {
	if s.Outbox == nil {
		return true, nil
	}
	now := time.Now().UTC()
	limits := s.normalizedLimits(in.Critical)

	if limits.Recipient10Min > 0 {
		count, err := s.Outbox.CountByRecipientSince(ctx, in.TenantID, toEmail, now.Add(-10*time.Minute))
		if err != nil {
			return false, err
		}
		if count >= int64(limits.Recipient10Min) {
			return false, nil
		}
	}

	if limits.Recipient24H > 0 {
		count, err := s.Outbox.CountByRecipientSince(ctx, in.TenantID, toEmail, now.Add(-24*time.Hour))
		if err != nil {
			return false, err
		}
		if count >= int64(limits.Recipient24H) {
			return false, nil
		}
	}

	if limits.Tenant1H > 0 {
		count, err := s.Outbox.CountByTenantSince(ctx, in.TenantID, now.Add(-time.Hour))
		if err != nil {
			return false, err
		}
		if count >= int64(limits.Tenant1H) {
			return false, nil
		}
	}

	if in.Critical {
		return true, nil
	}

	cooldown := s.cooldownForType(in.Type)
	if cooldown <= 0 {
		return true, nil
	}
	exists, err := s.Outbox.ExistsRecentByScope(
		ctx,
		in.TenantID,
		normalizeUUIDPtr(in.UserID),
		toEmail,
		strings.TrimSpace(in.Type),
		strings.TrimSpace(in.EntityType),
		normalizeUUIDPtr(in.EntityID),
		now.Add(-cooldown),
	)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func (s *EmailOutboxService) normalizedLimits(critical bool) EmailRateLimits {
	limits := s.Limits
	if limits.Recipient10Min <= 0 {
		limits.Recipient10Min = 3
	}
	if limits.Recipient24H <= 0 {
		limits.Recipient24H = 20
	}
	if limits.Tenant1H <= 0 {
		limits.Tenant1H = 200
	}
	if limits.OutbidCooldown <= 0 {
		limits.OutbidCooldown = 60 * time.Second
	}
	if limits.LotClosingCooldown <= 0 {
		limits.LotClosingCooldown = 10 * time.Minute
	}
	if limits.CriticalRecipient10Min <= 0 {
		limits.CriticalRecipient10Min = 10
	}
	if limits.CriticalRecipient24H <= 0 {
		limits.CriticalRecipient24H = 100
	}
	if limits.CriticalTenant1H <= 0 {
		limits.CriticalTenant1H = 1000
	}
	if critical {
		limits.Recipient10Min = limits.CriticalRecipient10Min
		limits.Recipient24H = limits.CriticalRecipient24H
		limits.Tenant1H = limits.CriticalTenant1H
	}
	return limits
}

func (s *EmailOutboxService) cooldownForType(kind string) time.Duration {
	switch strings.ToUpper(strings.TrimSpace(kind)) {
	case string(domain.NotificationEventOutbidded):
		return s.normalizedLimits(false).OutbidCooldown
	case "LOT_CLOSING":
		return s.normalizedLimits(false).LotClosingCooldown
	default:
		return 0
	}
}

func trimmedStringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeUUIDPtr(id *uuid.UUID) *uuid.UUID {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return id
}
