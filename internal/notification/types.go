package notification

import "context"

type Event struct {
	EventType    string
	OrderID      int64
	CustomerName string
	Reason       string
}

type Message struct {
	Channel      string
	Recipient    string
	OrderID      int64
	EventType    string
	CustomerName string
	Body         string
}

type Sender interface {
	Send(ctx context.Context, message Message) error
}

type Generator interface {
	Generate(ctx context.Context, event Event, fallback string) (string, error)
}
