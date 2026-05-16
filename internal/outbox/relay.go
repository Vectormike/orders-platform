package outbox

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Relay struct {
	repository       *Repository
	publisher        Publisher
	logger           *log.Logger
	batchSize        int
	fallbackInterval time.Duration
	staleAfter       time.Duration
}

func NewRelay(repository *Repository, publisher Publisher, logger *log.Logger, batchSize int, fallbackInterval time.Duration, staleAfter time.Duration) *Relay {
	if batchSize <= 0 {
		batchSize = 50
	}
	if fallbackInterval <= 0 {
		fallbackInterval = 30 * time.Second
	}
	if staleAfter <= 0 {
		staleAfter = 5 * time.Minute
	}

	return &Relay{
		repository:       repository,
		publisher:        publisher,
		logger:           logger,
		batchSize:        batchSize,
		fallbackInterval: fallbackInterval,
		staleAfter:       staleAfter,
	}
}

func (r *Relay) Run(ctx context.Context, notifyCh <-chan struct{}) error {
	ticker := time.NewTicker(r.fallbackInterval)
	defer ticker.Stop()

	for {
		if err := r.repository.RequeueStuckProcessing(ctx, r.staleAfter); err != nil {
			r.logger.Printf("outbox relay requeue error: %v", err)
		}

		if err := r.processAllAvailable(ctx); err != nil {
			r.logger.Printf("outbox relay batch error: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-notifyCh:
		case <-ticker.C:
		}
	}
}

func (r *Relay) processAllAvailable(ctx context.Context) error {
	for {
		events, err := r.repository.ClaimPending(ctx, r.batchSize)
		if err != nil {
			return err
		}

		if len(events) == 0 {
			return nil
		}

		for _, event := range events {
			if err := r.publisher.Publish(ctx, event); err != nil {
				if markErr := r.repository.MarkPending(ctx, event.ID, err.Error()); markErr != nil {
					return fmt.Errorf("publish error %v and failed to mark pending: %w", err, markErr)
				}
				r.logger.Printf("publish failed for outbox %d: %v", event.ID, err)
				continue
			}

			if err := r.repository.MarkSent(ctx, event.ID); err != nil {
				return fmt.Errorf("mark sent for outbox %d: %w", event.ID, err)
			}
		}
	}
}
