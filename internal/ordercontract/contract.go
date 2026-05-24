package ordercontract

import "fmt"

type OrderStatus string

const (
	StatusCreated    OrderStatus = "created"
	StatusProcessing OrderStatus = "processing"
	StatusFulfilled  OrderStatus = "fulfilled"
	StatusIncomplete OrderStatus = "incomplete"
	StatusFailed     OrderStatus = "failed"
)

type EventType string

const (
	EventOrderCreated    EventType = "order.created"
	EventOrderProcessing EventType = "order.processing"
	EventOrderFulfilled  EventType = "order.fulfilled"
	EventOrderIncomplete EventType = "order.incomplete"
	EventOrderFailed     EventType = "order.failed"
)

const AggregateTypeOrder = "order"

func CanTransition(from, to OrderStatus) bool {
	switch from {
	case StatusCreated:
		return to == StatusProcessing || to == StatusIncomplete
	case StatusProcessing:
		return to == StatusFulfilled || to == StatusIncomplete
	case StatusIncomplete:
		return to == StatusProcessing || to == StatusFailed
	default:
		return false
	}
}

func TopicForEvent(eventType EventType) string {
	// Keep one-topic-per-event for now to preserve strict event semantics.
	return string(eventType)
}

func RelayPublishEventType(eventType EventType) (EventType, error) {
	switch eventType {
	case EventOrderCreated:
		return EventOrderProcessing, nil
	case EventOrderProcessing, EventOrderFulfilled, EventOrderIncomplete, EventOrderFailed:
		return eventType, nil
	default:
		return "", fmt.Errorf("unsupported outbox event type: %s", eventType)
	}
}
