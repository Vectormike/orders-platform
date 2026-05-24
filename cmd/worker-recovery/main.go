package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"

	"order-system/internal/config"
	"order-system/internal/database"
	"order-system/internal/ordercontract"
	"order-system/internal/recovery"
	"order-system/internal/runtime"
)

type incompleteEventPayload struct {
	ID     int64  `json:"id"`
	Reason string `json:"reason"`
}

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	brokers := runtime.ParseBrokers(os.Getenv("KAFKA_BROKERS"))
	if len(brokers) == 0 {
		log.Fatal("KAFKA_BROKERS is required")
	}

	groupID := os.Getenv("RECOVERY_GROUP_ID")
	if groupID == "" {
		groupID = "order-recovery-worker"
	}

	maxRetries := runtime.GetEnvInt("RECOVERY_MAX_RETRIES", 3)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := database.EnsureOrderSchema(ctx, pool); err != nil {
		log.Fatal(err)
	}

	repository := recovery.NewRepository(pool, maxRetries)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    ordercontract.TopicForEvent(ordercontract.EventOrderIncomplete),
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("close recovery reader: %v", err)
		}
	}()

	log.Printf("recovery worker started (group=%s topic=%s max_retries=%d)", groupID, ordercontract.EventOrderIncomplete, maxRetries)

	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			log.Printf("fetch recovery message: %v", err)
			continue
		}

		shouldCommit, processErr := processMessage(ctx, repository, message)
		if processErr != nil {
			log.Printf("process recovery message: %v", processErr)
		}

		if shouldCommit {
			if err := reader.CommitMessages(ctx, message); err != nil {
				log.Printf("commit recovery message offset: %v", err)
			}
		}
	}

	log.Println("recovery worker stopped")
}

func processMessage(ctx context.Context, repository *recovery.Repository, message kafka.Message) (bool, error) {
	var payload incompleteEventPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return true, errors.New("invalid incomplete event payload")
	}

	if payload.ID <= 0 {
		return true, errors.New("incomplete event missing valid order id")
	}

	err := repository.RecoverIncompleteOrder(ctx, payload.ID, payload.Reason)
	if err != nil {
		switch {
		case errors.Is(err, recovery.ErrOrderNotFound),
			errors.Is(err, recovery.ErrInvalidTransition),
			errors.Is(err, recovery.ErrOrderAlreadyFinal):
			return true, err
		default:
			return false, err
		}
	}

	return true, nil
}
