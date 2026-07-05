package notification

import (
	"context"
	"fmt"
)

type StdoutSender struct{}

func NewStdoutSender() *StdoutSender {
	return &StdoutSender{}
}

func (s *StdoutSender) Send(_ context.Context, message Message) error {
	fmt.Printf(
		"[notification channel=%s recipient=%s order_id=%d event=%s] %s\n",
		message.Channel, message.Recipient, message.OrderID, message.EventType, message.Body,
	)
	return nil
}
