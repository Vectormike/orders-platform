package fulfillment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"order-system/internal/model"
)

type failureKind string

const (
	failureKindTerminal  failureKind = "terminal"
	failureKindRetryable failureKind = "retryable"
)

type validationFailure struct {
	kind   failureKind
	reason string
}

func (f validationFailure) Error() string {
	return f.reason
}

func validateOrderForFulfillment(ctx context.Context, order model.Order, verifier ExternalVerifier) error {
	// Level 1: required format/presence checks.
	if strings.TrimSpace(order.ShippingAddressLine) == "" {
		return validationFailure{kind: failureKindTerminal, reason: "shipping address line is required"}
	}
	if strings.TrimSpace(order.ShippingCity) == "" {
		return validationFailure{kind: failureKindTerminal, reason: "shipping city is required"}
	}

	country := strings.ToUpper(strings.TrimSpace(order.ShippingCountryCode))
	if len(country) != 2 {
		return validationFailure{kind: failureKindTerminal, reason: "shipping country code must be ISO-2"}
	}

	token := strings.TrimSpace(order.PaymentToken)
	if token == "" {
		return validationFailure{kind: failureKindTerminal, reason: "payment token is required"}
	}
	if len(token) < 8 {
		return validationFailure{kind: failureKindTerminal, reason: "payment token format is invalid"}
	}

	// Level 2: business constraints.
	if !isSupportedCountry(country) {
		return validationFailure{kind: failureKindTerminal, reason: fmt.Sprintf("shipping country %q is not supported", country)}
	}
	if order.AmountCents <= 0 {
		return validationFailure{kind: failureKindTerminal, reason: "order amount must be greater than zero"}
	}

	// Level 3: external verification hook.
	// Real integrations (payment/auth/address provider) should return
	// retryable or terminal classifications from provider errors.
	if err := runExternalVerification(ctx, order, verifier); err != nil {
		return err
	}

	return nil
}

func isSupportedCountry(country string) bool {
	switch country {
	case "NG", "US", "GB":
		return true
	default:
		return false
	}
}

func runExternalVerification(ctx context.Context, order model.Order, verifier ExternalVerifier) error {
	if verifier == nil {
		return nil
	}

	if err := verifier.Verify(ctx, order); err != nil {
		var externalErr ExternalVerificationError
		if errors.As(err, &externalErr) {
			return validationFailure{kind: externalErr.kind, reason: externalErr.reason}
		}

		return validationFailure{
			kind:   failureKindRetryable,
			reason: "external verification failed",
		}
	}

	return nil
}
