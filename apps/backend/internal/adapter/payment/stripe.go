package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type StripeGateway struct {
	SecretKey     string
	WebhookSecret string
	HTTPClient    *http.Client
	Now           func() time.Time
}

func (g *StripeGateway) CreateCheckoutSession(ctx context.Context, in port.CreateCheckoutSessionInput) (port.CheckoutSession, error) {
	if strings.TrimSpace(g.SecretKey) == "" {
		return port.CheckoutSession{}, errors.New("stripe is not configured")
	}
	if in.AmountMinor <= 0 || strings.TrimSpace(in.Currency) == "" {
		return port.CheckoutSession{}, errors.New("invalid checkout session payload")
	}
	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("success_url", in.SuccessURL)
	form.Set("cancel_url", in.CancelURL)
	form.Set("customer_email", strings.TrimSpace(in.CustomerEmail))
	form.Set("line_items[0][price_data][currency]", strings.ToLower(strings.TrimSpace(in.Currency)))
	form.Set("line_items[0][price_data][product_data][name]", strings.TrimSpace(in.Description))
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(in.AmountMinor, 10))
	form.Set("line_items[0][quantity]", "1")
	form.Set("metadata[invoice_id]", in.InvoiceID.String())
	form.Set("metadata[tenant_id]", in.TenantID.String())
	form.Set("metadata[invoice_number]", strings.TrimSpace(in.InvoiceNumber))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.stripe.com/v1/checkout/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return port.CheckoutSession{}, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(g.SecretKey))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := g.client().Do(req)
	if err != nil {
		return port.CheckoutSession{}, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= http.StatusBadRequest {
		return port.CheckoutSession{}, fmt.Errorf("stripe checkout session failed")
	}

	var payload struct {
		ID            string `json:"id"`
		URL           string `json:"url"`
		PaymentIntent string `json:"payment_intent"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return port.CheckoutSession{}, err
	}
	if strings.TrimSpace(payload.ID) == "" || strings.TrimSpace(payload.URL) == "" {
		return port.CheckoutSession{}, errors.New("stripe returned an invalid checkout session")
	}
	return port.CheckoutSession{
		ID:              payload.ID,
		URL:             payload.URL,
		PaymentIntentID: payload.PaymentIntent,
	}, nil
}

func (g *StripeGateway) ParseWebhook(payload []byte, signatureHeader string) (*port.PaymentWebhookEvent, error) {
	secret := strings.TrimSpace(g.WebhookSecret)
	if secret == "" {
		return nil, errors.New("stripe webhook is not configured")
	}
	timestamp, signatures := parseSignatureHeader(signatureHeader)
	if timestamp == "" || len(signatures) == 0 {
		return nil, errors.New("invalid stripe signature")
	}
	if !g.verifySignature(secret, timestamp, payload, signatures) {
		return nil, errors.New("invalid stripe signature")
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, errors.New("invalid stripe signature")
	}
	now := time.Now().UTC()
	if g.Now != nil {
		now = g.Now().UTC()
	}
	eventTime := time.Unix(ts, 0).UTC()
	if now.Sub(eventTime) > 5*time.Minute || eventTime.Sub(now) > 5*time.Minute {
		return nil, errors.New("stale stripe webhook signature")
	}

	var event struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID            string `json:"id"`
				PaymentIntent string `json:"payment_intent"`
				AmountTotal   int64  `json:"amount_total"`
				Currency      string `json:"currency"`
				Metadata      struct {
					InvoiceID string `json:"invoice_id"`
					TenantID  string `json:"tenant_id"`
				} `json:"metadata"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, errors.New("invalid stripe payload")
	}
	if event.Type != "checkout.session.completed" {
		return nil, nil
	}
	invoiceID, err := uuid.Parse(strings.TrimSpace(event.Data.Object.Metadata.InvoiceID))
	if err != nil {
		return nil, errors.New("invalid stripe payload")
	}
	tenantID, err := uuid.Parse(strings.TrimSpace(event.Data.Object.Metadata.TenantID))
	if err != nil {
		return nil, errors.New("invalid stripe payload")
	}
	return &port.PaymentWebhookEvent{
		EventID:           strings.TrimSpace(event.ID),
		EventType:         strings.TrimSpace(event.Type),
		InvoiceID:         invoiceID,
		TenantID:          tenantID,
		CheckoutSessionID: strings.TrimSpace(event.Data.Object.ID),
		PaymentIntentID:   strings.TrimSpace(event.Data.Object.PaymentIntent),
		AmountMinor:       event.Data.Object.AmountTotal,
		Currency:          strings.ToLower(strings.TrimSpace(event.Data.Object.Currency)),
		Payload:           append([]byte(nil), payload...),
	}, nil
}

func (g *StripeGateway) verifySignature(secret, timestamp string, payload []byte, signatures []string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	for _, sig := range signatures {
		if hmac.Equal([]byte(expected), []byte(sig)) {
			return true
		}
	}
	return false
}

func parseSignatureHeader(raw string) (string, []string) {
	parts := strings.Split(raw, ",")
	timestamp := ""
	signatures := make([]string, 0, 1)
	for _, part := range parts {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			timestamp = value
		case "v1":
			signatures = append(signatures, value)
		}
	}
	return timestamp, signatures
}

func (g *StripeGateway) client() *http.Client {
	if g.HTTPClient != nil {
		return g.HTTPClient
	}
	return &http.Client{Timeout: 15 * time.Second}
}
