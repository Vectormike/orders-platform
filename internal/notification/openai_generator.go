package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultOpenAIEndpoint = "https://api.openai.com/v1/responses"

type OpenAIGenerator struct {
	apiKey   string
	model    string
	endpoint string
	client   *http.Client
}

func NewOpenAIGenerator(apiKey, model, endpoint string) *OpenAIGenerator {
	if model == "" {
		model = "gpt-4.1-mini"
	}
	if endpoint == "" {
		endpoint = defaultOpenAIEndpoint
	}

	return &OpenAIGenerator{
		apiKey:   apiKey,
		model:    model,
		endpoint: endpoint,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *OpenAIGenerator) Generate(ctx context.Context, event Event, fallback string) (string, error) {
	systemPrompt := "You write concise customer support updates for order events. Keep under 300 characters. Be clear, calm, and actionable. Do not include technical internals."
	userPrompt := fmt.Sprintf(
		"Event: %s\nOrder ID: %d\nCustomer: %s\nReason: %s\nFallback: %s\nWrite one final customer message.",
		event.EventType, event.OrderID, event.CustomerName, event.Reason, fallback,
	)

	requestBody := map[string]any{
		"model": g.model,
		"input": []map[string]any{
			{
				"role": "system",
				"content": []map[string]string{
					{"type": "input_text", "text": systemPrompt},
				},
			},
			{
				"role": "user",
				"content": []map[string]string{
					{"type": "input_text", "text": userPrompt},
				},
			},
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create openai request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send openai request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai returned status %d", resp.StatusCode)
	}

	var parsed struct {
		OutputText string `json:"output_text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode openai response: %w", err)
	}

	text := strings.TrimSpace(parsed.OutputText)
	if text == "" {
		return "", fmt.Errorf("openai response had empty output_text")
	}

	return text, nil
}
