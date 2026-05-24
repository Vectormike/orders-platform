package recovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"order-system/internal/ordercontract"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrInvalidTransition = errors.New("invalid order status transition")
var ErrOrderAlreadyFinal = errors.New("order already in final status")

type Repository struct {
	pool       *pgxpool.Pool
	maxRetries int
}

type orderState struct {
	ID                int64
	Status            string
	RetryCount        int
	LastFailureReason string
}

func NewRepository(pool *pgxpool.Pool, maxRetries int) *Repository {
	if maxRetries <= 0 {
		maxRetries = 3
	}

	return &Repository{
		pool:       pool,
		maxRetries: maxRetries,
	}
}

func (r *Repository) RecoverIncompleteOrder(ctx context.Context, orderID int64, reason string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin recovery transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	order, err := getOrderForUpdate(ctx, tx, orderID)
	if err != nil {
		return err
	}

	status := ordercontract.OrderStatus(order.Status)
	if status == ordercontract.StatusFulfilled || status == ordercontract.StatusFailed {
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit final-state recovery transaction: %w", err)
		}
		return ErrOrderAlreadyFinal
	}

	if status != ordercontract.StatusIncomplete {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, status, ordercontract.StatusIncomplete)
	}

	nextRetryCount := order.RetryCount + 1
	if nextRetryCount >= r.maxRetries {
		if !ordercontract.CanTransition(ordercontract.StatusIncomplete, ordercontract.StatusFailed) {
			return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, ordercontract.StatusIncomplete, ordercontract.StatusFailed)
		}

		finalReason := fmt.Sprintf("max retries reached (%d): %s", r.maxRetries, reason)
		if err := updateOrderFailureState(ctx, tx, orderID, ordercontract.StatusFailed, nextRetryCount, finalReason); err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]any{
			"id":     orderID,
			"reason": finalReason,
		})
		if err != nil {
			return fmt.Errorf("marshal failed payload: %w", err)
		}
		if err := insertOutboxEvent(ctx, tx, orderID, ordercontract.EventOrderFailed, payload); err != nil {
			return err
		}
	} else {
		if !ordercontract.CanTransition(ordercontract.StatusIncomplete, ordercontract.StatusProcessing) {
			return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, ordercontract.StatusIncomplete, ordercontract.StatusProcessing)
		}

		if err := updateOrderFailureState(ctx, tx, orderID, ordercontract.StatusProcessing, nextRetryCount, reason); err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]any{
			"id": orderID,
		})
		if err != nil {
			return fmt.Errorf("marshal retry payload: %w", err)
		}
		if err := insertOutboxEvent(ctx, tx, orderID, ordercontract.EventOrderProcessing, payload); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, "SELECT pg_notify('outbox_new', $1);", fmt.Sprintf("%d", orderID)); err != nil {
		return fmt.Errorf("notify outbox relay: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit recovery transaction: %w", err)
	}

	return nil
}

func getOrderForUpdate(ctx context.Context, tx pgx.Tx, orderID int64) (orderState, error) {
	const query = `
SELECT id, status, retry_count, COALESCE(last_failure_reason, '')
FROM orders
WHERE id = $1
FOR UPDATE;
`

	var order orderState
	err := tx.QueryRow(ctx, query, orderID).
		Scan(&order.ID, &order.Status, &order.RetryCount, &order.LastFailureReason)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return orderState{}, ErrOrderNotFound
		}
		return orderState{}, fmt.Errorf("get order for recovery: %w", err)
	}

	return order, nil
}

func updateOrderFailureState(ctx context.Context, tx pgx.Tx, orderID int64, status ordercontract.OrderStatus, retryCount int, reason string) error {
	const query = `
UPDATE orders
SET status = $2,
	retry_count = $3,
	last_failure_reason = $4
WHERE id = $1;
`

	if _, err := tx.Exec(ctx, query, orderID, status, retryCount, reason); err != nil {
		return fmt.Errorf("update order recovery state: %w", err)
	}

	return nil
}

func insertOutboxEvent(ctx context.Context, tx pgx.Tx, orderID int64, eventType ordercontract.EventType, payload []byte) error {
	const query = `
INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
VALUES ($1, $2, $3, $4);
`

	if _, err := tx.Exec(ctx, query, ordercontract.AggregateTypeOrder, orderID, eventType, payload); err != nil {
		return fmt.Errorf("insert outbox recovery event: %w", err)
	}

	return nil
}
