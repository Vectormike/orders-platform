package fulfillment

import (
	"context"
	"fmt"

	"order-system/internal/model"
)

type ExternalVerifier interface {
	Verify(ctx context.Context, order model.Order) error
}

type ExternalVerificationError struct {
	kind   failureKind
	reason string
	err    error
}

func (e ExternalVerificationError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.reason, e.err)
	}
	return e.reason
}

func (e ExternalVerificationError) Unwrap() error {
	return e.err
}

func RetryableExternalError(reason string, cause error) error {
	return ExternalVerificationError{
		kind:   failureKindRetryable,
		reason: reason,
		err:    cause,
	}
}

func TerminalExternalError(reason string, cause error) error {
	return ExternalVerificationError{
		kind:   failureKindTerminal,
		reason: reason,
		err:    cause,
	}
}

type noopExternalVerifier struct{}

func (noopExternalVerifier) Verify(context.Context, model.Order) error {
	return nil
}

func NewNoopExternalVerifier() ExternalVerifier {
	return noopExternalVerifier{}
}
