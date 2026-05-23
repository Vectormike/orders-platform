package repository

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

type CreateOrderParams struct {
	CustomerName string
	AmountCents  int64
}

type OrderRepository interface {
	Create(ctx context.Context, params CreateOrderParams) (model.Order, error)
	GetByID(ctx context.Context, id int64) (model.Order, error)
}

type orderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) OrderRepository {
	return &orderRepository{pool: pool}
}

func (r *orderRepository) Create(ctx context.Context, params CreateOrderParams) (model.Order, error) {
	const insertOrderQuery = `
		INSERT INTO orders (customer_name, amount_cents, status)
		VALUES ($1, $2, $3)
		RETURNING id, customer_name, amount_cents, status, created_at;
	`

	const insertOutboxQuery = `
		INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4);
	`

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Order{}, fmt.Errorf("begin create order transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var order model.Order
	if err := tx.QueryRow(ctx, insertOrderQuery, params.CustomerName, params.AmountCents, ordercontract.StatusCreated).
		Scan(&order.ID, &order.CustomerName, &order.AmountCents, &order.Status, &order.CreatedAt); err != nil {
		return model.Order{}, fmt.Errorf("create order: %w", err)
	}

	payload, err := json.Marshal(order)
	if err != nil {
		return model.Order{}, fmt.Errorf("marshal outbox payload: %w", err)
	}

	if _, err := tx.Exec(ctx, insertOutboxQuery, ordercontract.AggregateTypeOrder, order.ID, ordercontract.EventOrderCreated, payload); err != nil {
		return model.Order{}, fmt.Errorf("insert outbox event: %w", err)
	}

	if _, err := tx.Exec(ctx, "SELECT pg_notify('outbox_new', $1);", fmt.Sprintf("%d", order.ID)); err != nil {
		return model.Order{}, fmt.Errorf("notify outbox relay: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Order{}, fmt.Errorf("commit create order transaction: %w", err)
	}

	return order, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (model.Order, error) {
	const query = `
SELECT id, customer_name, amount_cents, status, created_at
FROM orders
WHERE id = $1;
`

	var order model.Order
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&order.ID, &order.CustomerName, &order.AmountCents, &order.Status, &order.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, ErrOrderNotFound
		}
		return model.Order{}, fmt.Errorf("get order by id: %w", err)
	}

	return order, nil
}
