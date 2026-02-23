package usecase

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/dinesh-88/elevare-auction/internal/domain"
)

func RenderEmail(eventType domain.NotificationEventType, payload map[string]any) (string, string, error) {
	link, err := requiredString(payload, "link")
	if err != nil {
		return "", "", err
	}

	subject := ""
	body := ""
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
		subject = "You have been outbid"
		body = fmt.Sprintf("Your bid on %s has been surpassed. Current highest bid: %d.", lotTitle, amount)
	case domain.NotificationEventLotWon:
		lotTitle, err := requiredString(payload, "lot_title")
		if err != nil {
			return "", "", err
		}
		hammer, err := requiredInt64(payload, "hammer")
		if err != nil {
			return "", "", err
		}
		subject = "You won the lot"
		body = fmt.Sprintf("Congratulations. You won %s at %d.", lotTitle, hammer)
	case domain.NotificationEventInvoiceIssued:
		invoiceNumber, err := requiredString(payload, "invoice_number")
		if err != nil {
			return "", "", err
		}
		totalDue, err := requiredInt64(payload, "total_due")
		if err != nil {
			return "", "", err
		}
		subject = "Invoice issued"
		body = fmt.Sprintf("Invoice %s has been issued. Total due: %d.", invoiceNumber, totalDue)
	default:
		return "", "", fmt.Errorf("unsupported event type %q", eventType)
	}

	const tpl = `<!doctype html>
<html>
  <body style="font-family: Arial, sans-serif; color: #222; line-height: 1.5;">
    <p>{{ .Body }}</p>
    <p><a href="{{ .Link }}">View details</a></p>
  </body>
</html>`
	t, err := template.New("email").Parse(tpl)
	if err != nil {
		return "", "", err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, map[string]string{"Body": body, "Link": link}); err != nil {
		return "", "", err
	}
	return subject, b.String(), nil
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
