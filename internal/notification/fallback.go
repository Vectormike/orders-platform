package notification

import "fmt"

func FallbackMessage(event Event) string {
	name := event.CustomerName
	if name == "" {
		name = "there"
	}

	switch event.EventType {
	case "order.fulfilled":
		return fmt.Sprintf(
			"Hi %s, great news. Your order #%d has been processed successfully and is now confirmed.",
			name, event.OrderID,
		)
	case "order.incomplete":
		return fmt.Sprintf(
			"Hi %s, we are still processing order #%d. There was a temporary issue: %s. We are retrying and will keep you updated.",
			name, event.OrderID, event.Reason,
		)
	case "order.failed":
		return fmt.Sprintf(
			"Hi %s, we could not complete order #%d. Reason: %s. Please update your payment/order details and try again.",
			name, event.OrderID, event.Reason,
		)
	default:
		return fmt.Sprintf(
			"Hi %s, there is an update for order #%d. %s",
			name, event.OrderID, event.Reason,
		)
	}
}
