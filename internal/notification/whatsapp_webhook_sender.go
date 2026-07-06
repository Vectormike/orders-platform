package notification

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

type WhatsAppMetaSender struct {
	accessToken   string
	phoneNumberID string
	apiVersion    string
	client        *http.Client
}

func NewWhatsAppMetaSender(accessToken, phoneNumberID, apiVersion string) *WhatsAppMetaSender {
	if apiVersion == "" {
		apiVersion = "v21.0"
	}

	return &WhatsAppMetaSender{
		accessToken:   accessToken,
		phoneNumberID: phoneNumberID,
		apiVersion:    apiVersion,
		client:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *WhatsAppMetaSender) Send(ctx context.Context, message Message) error {
	if message.Recipient == "" {
		return fmt.Errorf("whatsapp recipient is required")
	}

	requestBody := map[string]any{
		"messaging_product": "whatsapp",
		"to":                message.Recipient,
		"type":              "text",
		"text": map[string]any{
			"preview_url": false,
			"body":        message.Body,
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("marshal whatsapp meta payload: %w", err)
	}

	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", s.apiVersion, s.phoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create whatsapp meta request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send whatsapp meta request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		apiMessage := strings.TrimSpace(string(responseBody))
		if apiMessage == "" {
			return fmt.Errorf("whatsapp meta returned status %d", resp.StatusCode)
		}
		return fmt.Errorf("whatsapp meta returned status %d: %s", resp.StatusCode, apiMessage)
	}

	return nil
}
