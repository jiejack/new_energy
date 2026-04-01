package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConfig struct {
	Brokers     []string
	TopicPrefix string
}

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(cfg KafkaConfig, topic string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.TopicPrefix + topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        true,
	}
	
	return &KafkaProducer{writer: writer}
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

func (p *KafkaProducer) SendMessage(ctx context.Context, key, value interface{}) error {
	keyBytes, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal key: %w", err)
	}
	
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	msg := kafka.Message{
		Key:   keyBytes,
		Value: valueBytes,
		Time:  time.Now(),
	}
	
	return p.writer.WriteMessages(ctx, msg)
}

func (p *KafkaProducer) SendBytes(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}
	
	return p.writer.WriteMessages(ctx, msg)
}

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(cfg KafkaConfig, topic, groupID string) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.TopicPrefix + topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	
	return &KafkaConsumer{reader: reader}
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

func (c *KafkaConsumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

func (c *KafkaConsumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

func (c *KafkaConsumer) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return c.reader.CommitMessages(ctx, msgs...)
}

type MessageHandler func(ctx context.Context, msg kafka.Message) error

func (c *KafkaConsumer) Consume(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.ReadMessage(ctx)
			if err != nil {
				return fmt.Errorf("failed to read message: %w", err)
			}
			
			if err := handler(ctx, msg); err != nil {
				return fmt.Errorf("failed to handle message: %w", err)
			}
		}
	}
}

const (
	TopicDataCollect   = "data.collect"
	TopicAlarmEvent    = "alarm.event"
	TopicAlarmNotify   = "alarm.notify"
	TopicDeviceStatus  = "device.status"
	TopicComputeResult = "compute.result"
)

type CollectDataMessage struct {
	PointID   string    `json:"point_id"`
	Value     float64   `json:"value"`
	Quality   byte      `json:"quality"`
	Timestamp int64     `json:"timestamp"`
	Source    string    `json:"source"`
}

type AlarmEventMessage struct {
	AlarmID   string    `json:"alarm_id"`
	PointID   string    `json:"point_id"`
	DeviceID  string    `json:"device_id"`
	StationID string    `json:"station_id"`
	Type      string    `json:"type"`
	Level     int       `json:"level"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp int64     `json:"timestamp"`
}

type DeviceStatusMessage struct {
	DeviceID  string `json:"device_id"`
	StationID string `json:"station_id"`
	Status    int    `json:"status"`
	Timestamp int64  `json:"timestamp"`
}
