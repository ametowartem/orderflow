package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
	topic  string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	return &Producer{client: client, topic: topic}, nil
}

func (p *Producer) Publish(ctx context.Context, key string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	record := &kgo.Record{Topic: p.topic, Key: []byte(key), Value: payload}

	result := p.client.ProduceSync(ctx, record)
	return fmt.Errorf("publish error: %w", result.FirstErr())
}
