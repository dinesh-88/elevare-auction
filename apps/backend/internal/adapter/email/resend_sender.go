package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ResendSender struct {
	APIKey  string
	From    string
	Client  *http.Client
	BaseURL string
}

func (s *ResendSender) Send(ctx context.Context, to, subject, html string) error {
	apiKey := strings.TrimSpace(s.APIKey)
	from := strings.TrimSpace(s.From)
	if apiKey == "" || from == "" {
		return fmt.Errorf("resend configuration missing")
	}
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("recipient email is required")
	}
	payload, err := json.Marshal(map[string]any{
		"from":    from,
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	})
	if err != nil {
		return err
	}
	endpoint := strings.TrimRight(strings.TrimSpace(s.BaseURL), "/")
	if endpoint == "" {
		endpoint = "https://api.resend.com"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/emails", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("resend send failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
}
