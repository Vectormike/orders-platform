package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WhatsAppTwilioSender struct {
	accountSID string
	authToken  string
	from       string
	client     *http.Client
}

func NewWhatsAppTwilioSender(accountSID, authToken, from string) *WhatsAppTwilioSender {
	return &WhatsAppTwilioSender{
		accountSID: strings.TrimSpace(accountSID),
		authToken:  strings.TrimSpace(authToken),
		from:       normalizeTwilioWhatsAppAddress(from),
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *WhatsAppTwilioSender) Send(ctx context.Context, message Message) error {
	if strings.TrimSpace(message.Recipient) == "" {
		return fmt.Errorf("whatsapp recipient is required")
	}
	if strings.TrimSpace(s.from) == "" {
		return fmt.Errorf("twilio from number is required")
	}

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.accountSID)

	form := url.Values{}
	form.Set("To", normalizeTwilioWhatsAppAddress(message.Recipient))
	form.Set("From", s.from)
	form.Set("Body", message.Body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create twilio request: %w", err)
	}
	req.SetBasicAuth(s.accountSID, s.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send twilio request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		apiMessage := strings.TrimSpace(string(responseBody))
		if apiMessage == "" {
			return fmt.Errorf("twilio returned status %d", resp.StatusCode)
		}
		return fmt.Errorf("twilio returned status %d: %s", resp.StatusCode, apiMessage)
	}

	var parsed struct {
		SID      string `json:"sid"`
		ErrorMsg string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("decode twilio response: %w", err)
	}
	if strings.TrimSpace(parsed.SID) == "" {
		return fmt.Errorf("twilio response missing sid")
	}

	return nil
}

func normalizeTwilioWhatsAppAddress(value string) string {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "whatsapp:") {
		return trimmed
	}
	return "whatsapp:" + trimmed
}
