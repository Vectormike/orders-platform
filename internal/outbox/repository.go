package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Event struct {
	ID          int64
	AggregateID int64
	EventType   string
	Payload     []byte
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ClaimPending(ctx context.Context, limit int) ([]Event, error) {
	const query = `
WITH claimed AS (
	SELECT id
	FROM outbox
	WHERE status = 'pending'
	ORDER BY id
	FOR UPDATE SKIP LOCKED
	LIMIT $1
)
UPDATE outbox AS o
SET status = 'processing',
	processing_at = NOW()
FROM claimed
WHERE o.id = claimed.id
RETURNING o.id, o.aggregate_id, o.event_type, o.payload;
`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("claim outbox events: %w", err)
	}
	defer rows.Close()

	events := make([]Event, 0, limit)
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.AggregateID, &event.EventType, &event.Payload); err != nil {
			return nil, fmt.Errorf("scan claimed outbox event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claimed outbox events: %w", err)
	}

	return events, nil
}

func (r *Repository) MarkSent(ctx context.Context, id int64) error {
	const query = `
UPDATE outbox
SET status = 'sent',
	sent_at = NOW(),
	processing_at = NULL,
	last_error = NULL
WHERE id = $1;
`

	if _, err := r.pool.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("mark outbox sent: %w", err)
	}

	return nil
}

func (r *Repository) MarkPending(ctx context.Context, id int64, lastError string) error {
	const query = `
UPDATE outbox
SET status = 'pending',
	attempt_count = attempt_count + 1,
	processing_at = NULL,
	last_error = $2
WHERE id = $1;
`

	if _, err := r.pool.Exec(ctx, query, id, lastError); err != nil {
		return fmt.Errorf("mark outbox pending: %w", err)
	}

	return nil
}

func (r *Repository) RequeueStuckProcessing(ctx context.Context, staleAfter time.Duration) error {
	const query = `
UPDATE outbox
SET status = 'pending',
	processing_at = NULL
WHERE status = 'processing'
	AND processing_at < NOW() - ($1::interval);
`

	interval := fmt.Sprintf("%f seconds", staleAfter.Seconds())
	if _, err := r.pool.Exec(ctx, query, interval); err != nil {
		return fmt.Errorf("requeue stuck outbox rows: %w", err)
	}

	return nil
}
