package fulfillment

import (
	"context"
	"strings"

	"order-system/internal/model"
)

type tokenExternalVerifier struct{}

func NewTokenExternalVerifier() ExternalVerifier {
	return tokenExternalVerifier{}
}

func (tokenExternalVerifier) Verify(_ context.Context, order model.Order) error {
	token := strings.TrimSpace(order.PaymentToken)

	// Lightweight deterministic mapper for local/dev behavior.
	// Real adapters should inspect provider response codes instead.
	if strings.HasPrefix(token, "retry_") {
		return RetryableExternalError("temporary payment verification failure", nil)
	}
	if strings.HasPrefix(token, "declined_") {
		return TerminalExternalError("payment was declined", nil)
	}

	return nil
}
