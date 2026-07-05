package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Processor struct {
	channel   string
	recipient string
	sender    Sender
	generator Generator
}

func NewProcessor(channel, recipient string, sender Sender, generator Generator) *Processor {
	return &Processor{
		channel:   channel,
		recipient: recipient,
		sender:    sender,
		generator: generator,
	}
}

func (p *Processor) Process(ctx context.Context, topic string, payload []byte) error {
	event, err := decodeEvent(topic, payload)
	if err != nil {
		return err
	}

	fallback := FallbackMessage(event)
	messageText := fallback
	if p.generator != nil {
		generated, err := p.generator.Generate(ctx, event, fallback)
		if err == nil && strings.TrimSpace(generated) != "" {
			messageText = strings.TrimSpace(generated)
		}
	}

	message := Message{
		Channel:      p.channel,
		Recipient:    p.recipient,
		OrderID:      event.OrderID,
		EventType:    event.EventType,
		CustomerName: event.CustomerName,
		Body:         messageText,
	}

	if err := p.sender.Send(ctx, message); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}

	return nil
}

func decodeEvent(topic string, payload []byte) (Event, error) {
	event := Event{
		EventType: topic,
		Reason:    "no reason provided",
	}

	var parsed struct {
		ID           int64  `json:"id"`
		CustomerName string `json:"customer_name"`
		Reason       string `json:"reason"`
		Order        struct {
			ID           int64  `json:"id"`
			CustomerName string `json:"customer_name"`
		} `json:"order"`
	}

	if err := json.Unmarshal(payload, &parsed); err != nil {
		return Event{}, fmt.Errorf("decode notification event payload: %w", err)
	}

	if parsed.Order.ID > 0 {
		event.OrderID = parsed.Order.ID
		event.CustomerName = parsed.Order.CustomerName
	} else {
		event.OrderID = parsed.ID
		event.CustomerName = parsed.CustomerName
	}

	if event.OrderID <= 0 {
		return Event{}, fmt.Errorf("notification payload missing valid order id")
	}

	if strings.TrimSpace(parsed.Reason) != "" {
		event.Reason = strings.TrimSpace(parsed.Reason)
	}

	return event, nil
}
