package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/segmentio/kafka-go"

	"order-system/internal/config"
	"order-system/internal/notification"
	"order-system/internal/ordercontract"
	"order-system/internal/runtime"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}

	brokers := runtime.ParseBrokers(os.Getenv("KAFKA_BROKERS"))
	if len(brokers) == 0 {
		log.Fatal("KAFKA_BROKERS is required")
	}

	groupID := os.Getenv("NOTIFICATIONS_GROUP_ID")
	if groupID == "" {
		groupID = "order-notifications-worker"
	}

	channel := strings.TrimSpace(os.Getenv("NOTIFICATION_CHANNEL"))
	if channel == "" {
		channel = "stdout"
	}
	recipient := strings.TrimSpace(os.Getenv("NOTIFICATION_RECIPIENT"))
	if recipient == "" {
		recipient = "customer"
	}

	var sender notification.Sender
	switch channel {
	case "whatsapp":
		accessToken := strings.TrimSpace(os.Getenv("WHATSAPP_ACCESS_TOKEN"))
		if accessToken == "" {
			log.Fatal("WHATSAPP_ACCESS_TOKEN is required when NOTIFICATION_CHANNEL=whatsapp")
		}
		phoneNumberID := strings.TrimSpace(os.Getenv("WHATSAPP_PHONE_NUMBER_ID"))
		if phoneNumberID == "" {
			log.Fatal("WHATSAPP_PHONE_NUMBER_ID is required when NOTIFICATION_CHANNEL=whatsapp")
		}
		apiVersion := strings.TrimSpace(os.Getenv("WHATSAPP_API_VERSION"))
		sender = notification.NewWhatsAppMetaSender(accessToken, phoneNumberID, apiVersion)
	default:
		channel = "stdout"
		sender = notification.NewStdoutSender()
	}

	var generator notification.Generator
	openAIKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if openAIKey != "" {
		model := strings.TrimSpace(os.Getenv("OPENAI_MODEL"))
		endpoint := strings.TrimSpace(os.Getenv("OPENAI_RESPONSES_ENDPOINT"))
		generator = notification.NewOpenAIGenerator(openAIKey, model, endpoint)
	}

	processor := notification.NewProcessor(channel, recipient, sender, generator)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		GroupTopics: []string{
			ordercontract.TopicForEvent(ordercontract.EventOrderFulfilled),
			ordercontract.TopicForEvent(ordercontract.EventOrderIncomplete),
			ordercontract.TopicForEvent(ordercontract.EventOrderFailed),
		},
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("close notifications reader: %v", err)
		}
	}()

	log.Printf("notifications worker started (group=%s channel=%s)", groupID, channel)

	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			log.Printf("fetch notifications message: %v", err)
			continue
		}

		if err := processor.Process(ctx, message.Topic, message.Value); err != nil {
			log.Printf("process notifications message: %v", err)
			continue
		}

		if err := reader.CommitMessages(ctx, message); err != nil {
			log.Printf("commit notifications message offset: %v", err)
		}
	}

	log.Println("notifications worker stopped")
}
