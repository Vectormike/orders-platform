package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func EnsureOrderSchema(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `
CREATE TABLE IF NOT EXISTS orders (
	id BIGSERIAL PRIMARY KEY,
	customer_name TEXT NOT NULL,
	amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
	shipping_address_line TEXT NOT NULL DEFAULT '',
	shipping_city TEXT NOT NULL DEFAULT '',
	shipping_country_code TEXT NOT NULL DEFAULT '',
	payment_token TEXT NOT NULL DEFAULT '',
	retry_count INTEGER NOT NULL DEFAULT 0,
	last_failure_reason TEXT NULL,
	status TEXT NOT NULL DEFAULT 'created',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS outbox (
	id BIGSERIAL PRIMARY KEY,
	aggregate_type TEXT NOT NULL,
	aggregate_id BIGINT NOT NULL,
	event_type TEXT NOT NULL,
	payload JSONB NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	attempt_count INTEGER NOT NULL DEFAULT 0,
	last_error TEXT NULL,
	processing_at TIMESTAMPTZ NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	sent_at TIMESTAMPTZ NULL
);

ALTER TABLE outbox ADD COLUMN IF NOT EXISTS attempt_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE outbox ADD COLUMN IF NOT EXISTS last_error TEXT NULL;
ALTER TABLE outbox ADD COLUMN IF NOT EXISTS processing_at TIMESTAMPTZ NULL;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_address_line TEXT NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_city TEXT NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_country_code TEXT NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_token TEXT NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS last_failure_reason TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_outbox_status_created_at
ON outbox (status, created_at, id);
`

	if _, err := pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("ensure orders schema: %w", err)
	}

	return nil
}
