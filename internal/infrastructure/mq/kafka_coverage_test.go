package mq

import (
	"context"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestNewKafkaProducer(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	producer := NewKafkaProducer(cfg, "test-topic")
	assert.NotNil(t, producer)
	assert.NotNil(t, producer.writer)
}

func TestKafkaProducer_Close(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	producer := NewKafkaProducer(cfg, "test-topic")
	assert.NotNil(t, producer)

	err := producer.Close()
	assert.NoError(t, err)
}

func TestKafkaProducer_SendMessage_MarshalKeyError(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	producer := NewKafkaProducer(cfg, "test-topic")
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := producer.SendMessage(ctx, make(chan int), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal key")
}

func TestKafkaProducer_SendMessage_MarshalValueError(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	producer := NewKafkaProducer(cfg, "test-topic")
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := producer.SendMessage(ctx, "key1", make(chan int))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal value")
}

func TestKafkaProducer_SendMessage_ContextCancelled(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	producer := NewKafkaProducer(cfg, "test-topic")
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := producer.SendMessage(ctx, "key1", "value1")
	assert.Error(t, err)
}

func TestKafkaProducer_SendBytes_ContextCancelled(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	producer := NewKafkaProducer(cfg, "test-topic")
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := producer.SendBytes(ctx, []byte("key"), []byte("value"))
	assert.Error(t, err)
}

func TestNewKafkaConsumer(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	consumer := NewKafkaConsumer(cfg, "test-topic", "test-group")
	assert.NotNil(t, consumer)
	assert.NotNil(t, consumer.reader)
}

func TestKafkaConsumer_Close(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	consumer := NewKafkaConsumer(cfg, "test-topic", "test-group")
	assert.NotNil(t, consumer)

	err := consumer.Close()
	assert.NoError(t, err)
}

func TestKafkaConsumer_ReadMessage_ContextCancelled(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	consumer := NewKafkaConsumer(cfg, "test-topic", "test-group")
	defer consumer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := consumer.ReadMessage(ctx)
	assert.Error(t, err)
}

func TestKafkaConsumer_FetchMessage_ContextCancelled(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	consumer := NewKafkaConsumer(cfg, "test-topic", "test-group")
	defer consumer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := consumer.FetchMessage(ctx)
	assert.Error(t, err)
}

func TestKafkaConsumer_Consume_ContextCancelled(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	consumer := NewKafkaConsumer(cfg, "test-topic", "test-group")
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := consumer.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
		return nil
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestKafkaConfig_Defaults(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"broker1:9092", "broker2:9092"},
		TopicPrefix: "nem.",
	}
	assert.Len(t, cfg.Brokers, 2)
	assert.Equal(t, "nem.", cfg.TopicPrefix)
}

func TestKafkaProducer_TopicPrefix(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "prefix.",
	}
	producer := NewKafkaProducer(cfg, "my-topic")
	defer producer.Close()

	assert.Equal(t, "prefix.my-topic", producer.writer.Topic)
}

func TestKafkaConsumer_TopicPrefix(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "prefix.",
	}
	consumer := NewKafkaConsumer(cfg, "my-topic", "my-group")
	defer consumer.Close()
}

func TestKafkaConsumer_CommitMessages_ContextCancelled(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	consumer := NewKafkaConsumer(cfg, "test-topic", "test-group")
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := consumer.CommitMessages(ctx)
	assert.Error(t, err)
}

func TestKafkaConsumer_Consume_ReadError(t *testing.T) {
	cfg := KafkaConfig{
		Brokers:     []string{"localhost:9092"},
		TopicPrefix: "test.",
	}
	consumer := NewKafkaConsumer(cfg, "test-topic", "test-group")
	defer consumer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := consumer.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
		return nil
	})
	assert.Error(t, err)
}
