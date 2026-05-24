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
	"order-system/internal/fulfillment"
	"order-system/internal/ordercontract"
	"order-system/internal/runtime"
)

type processingEventPayload struct {
	ID int64 `json:"id"`
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

	groupID := os.Getenv("FULFILLMENT_GROUP_ID")
	if groupID == "" {
		groupID = "order-fulfillment-worker"
	}

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

	verifier := fulfillment.NewTokenExternalVerifier()
	repository := fulfillment.NewRepository(pool, verifier)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    ordercontract.TopicForEvent(ordercontract.EventOrderProcessing),
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("close fulfillment reader: %v", err)
		}
	}()

	log.Printf("fulfillment worker started (group=%s topic=%s)", groupID, ordercontract.EventOrderProcessing)

	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			log.Printf("fetch kafka message: %v", err)
			continue
		}

		shouldCommit, processErr := processMessage(ctx, repository, message)
		if processErr != nil {
			log.Printf("process fulfillment message: %v", processErr)
		}

		if shouldCommit {
			if err := reader.CommitMessages(ctx, message); err != nil {
				log.Printf("commit fulfillment message offset: %v", err)
			}
		}
	}

	log.Println("fulfillment worker stopped")
}

func processMessage(ctx context.Context, repository *fulfillment.Repository, message kafka.Message) (bool, error) {
	var payload processingEventPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return true, errors.New("invalid processing event payload")
	}

	if payload.ID <= 0 {
		return true, errors.New("processing event missing valid order id")
	}

	err := repository.FulfillFromProcessingEvent(ctx, payload.ID)
	if err != nil {
		switch {
		case errors.Is(err, fulfillment.ErrOrderNotFound),
			errors.Is(err, fulfillment.ErrInvalidTransition),
			errors.Is(err, fulfillment.ErrTerminalFailureHandled),
			errors.Is(err, fulfillment.ErrRetryableFailureHandled):
			// Drop poison/non-actionable events to keep consumption moving.
			return true, err
		default:
			// Retriable/system failures should not commit so Kafka redelivers.
			return false, err
		}
	}

	return true, nil
}
