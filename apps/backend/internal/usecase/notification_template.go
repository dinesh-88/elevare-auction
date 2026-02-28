package usecase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	emailtpl "github.com/dinesh-88/elevare-auction/internal/email"
)

var defaultTemplateRenderer = emailtpl.NewTemplateRenderer()

func RenderEmail(eventType domain.NotificationEventType, payload map[string]any) (string, string, error) {
	rendered, err := RenderEmailFull(eventType, domain.LocaleEnglish, emailtpl.BrandContext{TenantName: "Elevare"}, payload)
	if err != nil {
		return "", "", err
	}
	return rendered.Subject, rendered.HTML, nil
}

func RenderEmailFull(eventType domain.NotificationEventType, locale domain.Locale, brand emailtpl.BrandContext, payload map[string]any) (emailtpl.RenderedEmail, error) {
	if err := validateEmailPayload(eventType, payload); err != nil {
		return emailtpl.RenderedEmail{}, err
	}
	return defaultTemplateRenderer.Render(emailTemplateKey(eventType), locale, brand, payload)
}

func emailTemplateKey(eventType domain.NotificationEventType) string {
	switch eventType {
	case domain.NotificationEventOutbidded:
		return "outbid"
	case domain.NotificationEventLotWon:
		return "lot_won"
	case domain.NotificationEventInvoiceIssued:
		return "invoice_ready"
	default:
		return strings.ToLower(string(eventType))
	}
}

func tenantLogoURL(tenant *domain.Tenant) string {
	for _, blob := range []json.RawMessage{tenant.BrandingPublished, tenant.Branding, tenant.BrandingDraft} {
		if v := jsonString(blob, "logo_url"); v != "" {
			return v
		}
	}
	return ""
}

func tenantSupportEmail(tenant *domain.Tenant) string {
	for _, blob := range []json.RawMessage{tenant.ContentPublished, tenant.ContentDraft} {
		if v := jsonNestedString(blob, "contact", "email"); v != "" {
			return v
		}
		if v := jsonNestedString(blob, "footer", "contact_email"); v != "" {
			return v
		}
	}
	return ""
}

func tenantPrimaryURL(tenant *domain.Tenant) string {
	if tenant == nil {
		return ""
	}
	if tenant.CustomDomain != nil && strings.TrimSpace(*tenant.CustomDomain) != "" {
		return "https://" + strings.TrimSpace(*tenant.CustomDomain)
	}
	if strings.TrimSpace(tenant.Subdomain) != "" {
		return "https://" + strings.TrimSpace(tenant.Subdomain) + ".elevareauctions.com"
	}
	return ""
}

func jsonString(raw json.RawMessage, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	if v, ok := payload[key].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func jsonNestedString(raw json.RawMessage, keys ...string) string {
	if len(raw) == 0 || len(keys) == 0 {
		return ""
	}
	var current any
	if err := json.Unmarshal(raw, &current); err != nil {
		return ""
	}
	for _, key := range keys {
		obj, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = obj[key]
	}
	if value, ok := current.(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func renderInApp(eventType domain.NotificationEventType, payload map[string]any) (title, body string, err error) {
	switch eventType {
	case domain.NotificationEventOutbidded:
		lotTitle, err := requiredString(payload, "lot_title")
		if err != nil {
			return "", "", err
		}
		amount, err := requiredInt64(payload, "amount")
		if err != nil {
			return "", "", err
		}
		return "Outbid Notice", fmt.Sprintf("A higher bid was placed on %s. Current bid: %d.", lotTitle, amount), nil
	case domain.NotificationEventLotWon:
		lotTitle, err := requiredString(payload, "lot_title")
		if err != nil {
			return "", "", err
		}
		hammer, err := requiredInt64(payload, "hammer")
		if err != nil {
			return "", "", err
		}
		return "Lot Won", fmt.Sprintf("You won %s at %d.", lotTitle, hammer), nil
	case domain.NotificationEventInvoiceIssued:
		invoiceNumber, err := requiredString(payload, "invoice_number")
		if err != nil {
			return "", "", err
		}
		totalDue, err := requiredInt64(payload, "total_due")
		if err != nil {
			return "", "", err
		}
		return "Invoice Issued", fmt.Sprintf("Invoice %s is ready. Total due: %d.", invoiceNumber, totalDue), nil
	default:
		return "", "", fmt.Errorf("unsupported event type %q", eventType)
	}
}

func validateEmailPayload(eventType domain.NotificationEventType, payload map[string]any) error {
	if _, err := requiredString(payload, "link"); err != nil {
		return err
	}
	switch eventType {
	case domain.NotificationEventOutbidded:
		if _, err := requiredString(payload, "lot_title"); err != nil {
			return err
		}
		if _, err := requiredInt64(payload, "amount"); err != nil {
			return err
		}
	case domain.NotificationEventLotWon:
		if _, err := requiredString(payload, "lot_title"); err != nil {
			return err
		}
		if _, err := requiredInt64(payload, "hammer"); err != nil {
			return err
		}
	case domain.NotificationEventInvoiceIssued:
		if _, err := requiredString(payload, "invoice_number"); err != nil {
			return err
		}
		if _, err := requiredInt64(payload, "total_due"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported event type %q", eventType)
	}
	return nil
}
