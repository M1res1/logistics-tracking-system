package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       10e3,
			MaxBytes:       10e6,
			MaxWait:        1 * time.Second,
			CommitInterval: time.Second,
		}),
	}
}

type Handler func(key, value []byte) error

func (c *Consumer) Consume(ctx context.Context, handler Handler) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		if err := handler(msg.Key, msg.Value); err != nil {
			// log error but continue consuming
			continue
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
