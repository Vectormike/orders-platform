package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"

	"order-system/internal/config"
	"order-system/internal/database"
	"order-system/internal/outbox"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	brokers := parseBrokers(os.Getenv("KAFKA_BROKERS"))
	if len(brokers) == 0 {
		log.Fatal("KAFKA_BROKERS is required")
	}

	batchSize := getEnvInt("OUTBOX_BATCH_SIZE", 50)
	fallbackInterval := time.Duration(getEnvInt("OUTBOX_FALLBACK_SECONDS", 30)) * time.Second
	staleAfter := time.Duration(getEnvInt("OUTBOX_STALE_AFTER_SECONDS", 300)) * time.Second

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

	publisher := outbox.NewKafkaPublisher(brokers)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("close kafka publisher: %v", err)
		}
	}()

	repository := outbox.NewRepository(pool)
	relay := outbox.NewRelay(repository, publisher, log.Default(), batchSize, fallbackInterval, staleAfter)

	notifyCh := make(chan struct{}, 1)
	go listenForOutboxNotify(ctx, databaseURL, notifyCh)

	log.Println("outbox relay started")
	if err := relay.Run(ctx, notifyCh); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
	log.Println("outbox relay stopped")
}

func listenForOutboxNotify(ctx context.Context, databaseURL string, notifyCh chan<- struct{}) {
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Printf("listen connection failed, fallback polling only: %v", err)
		return
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "LISTEN outbox_new"); err != nil {
		log.Printf("LISTEN outbox_new failed, fallback polling only: %v", err)
		return
	}

	for {
		_, err := conn.WaitForNotification(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("wait for notification error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		select {
		case notifyCh <- struct{}{}:
		default:
		}
	}
}

func parseBrokers(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))
	for _, part := range parts {
		broker := strings.TrimSpace(part)
		if broker != "" {
			brokers = append(brokers, broker)
		}
	}

	return brokers
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
