package fulfillment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"order-system/internal/model"
	"order-system/internal/ordercontract"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrInvalidTransition = errors.New("invalid order status transition")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) FulfillFromProcessingEvent(ctx context.Context, orderID int64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin fulfillment transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	order, err := getOrderForUpdate(ctx, tx, orderID)
	if err != nil {
		return err
	}

	currentStatus := ordercontract.OrderStatus(order.Status)
	if currentStatus == ordercontract.StatusFulfilled {
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit idempotent fulfillment transaction: %w", err)
		}
		return nil
	}

	if currentStatus != ordercontract.StatusProcessing {
		if !ordercontract.CanTransition(currentStatus, ordercontract.StatusProcessing) {
			return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, currentStatus, ordercontract.StatusProcessing)
		}
		if err := updateOrderStatus(ctx, tx, orderID, ordercontract.StatusProcessing); err != nil {
			return err
		}
	}

	if !ordercontract.CanTransition(ordercontract.StatusProcessing, ordercontract.StatusFulfilled) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, ordercontract.StatusProcessing, ordercontract.StatusFulfilled)
	}

	fulfilledOrder, err := updateOrderStatusAndReturn(ctx, tx, orderID, ordercontract.StatusFulfilled)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(fulfilledOrder)
	if err != nil {
		return fmt.Errorf("marshal fulfilled outbox payload: %w", err)
	}

	if err := insertOutboxEvent(ctx, tx, fulfilledOrder.ID, ordercontract.EventOrderFulfilled, payload); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "SELECT pg_notify('outbox_new', $1);", fmt.Sprintf("%d", fulfilledOrder.ID)); err != nil {
		return fmt.Errorf("notify outbox relay: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit fulfillment transaction: %w", err)
	}

	return nil
}

func getOrderForUpdate(ctx context.Context, tx pgx.Tx, orderID int64) (model.Order, error) {
	const query = `
SELECT id, customer_name, amount_cents, status, created_at
FROM orders
WHERE id = $1
FOR UPDATE;
`

	var order model.Order
	err := tx.QueryRow(ctx, query, orderID).
		Scan(&order.ID, &order.CustomerName, &order.AmountCents, &order.Status, &order.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, ErrOrderNotFound
		}
		return model.Order{}, fmt.Errorf("get order for fulfillment: %w", err)
	}

	return order, nil
}

func updateOrderStatus(ctx context.Context, tx pgx.Tx, orderID int64, status ordercontract.OrderStatus) error {
	const query = `
UPDATE orders
SET status = $2
WHERE id = $1;
`

	if _, err := tx.Exec(ctx, query, orderID, status); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

func updateOrderStatusAndReturn(ctx context.Context, tx pgx.Tx, orderID int64, status ordercontract.OrderStatus) (model.Order, error) {
	const query = `
UPDATE orders
SET status = $2
WHERE id = $1
RETURNING id, customer_name, amount_cents, status, created_at;
`

	var order model.Order
	if err := tx.QueryRow(ctx, query, orderID, status).
		Scan(&order.ID, &order.CustomerName, &order.AmountCents, &order.Status, &order.CreatedAt); err != nil {
		return model.Order{}, fmt.Errorf("update order status and return: %w", err)
	}

	return order, nil
}

func insertOutboxEvent(ctx context.Context, tx pgx.Tx, orderID int64, eventType ordercontract.EventType, payload []byte) error {
	const query = `
INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
VALUES ($1, $2, $3, $4);
`

	if _, err := tx.Exec(ctx, query, ordercontract.AggregateTypeOrder, orderID, eventType, payload); err != nil {
		return fmt.Errorf("insert outbox fulfillment event: %w", err)
	}

	return nil
}
