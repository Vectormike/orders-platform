ppppackage outbox

import (
	"context"
	"fmt"
	"strconv"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, event Event) error {
	message := kafka.Message{
		Topic: event.EventType,
		Key:   []byte(strconv.FormatInt(event.AggregateID, 10)),
		Value: event.Payload,
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("publish event %d to kafka: %w", event.ID, err)
	}

	return nil
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
