package helpers

import (
	"context"
	"sync"
	"time"
)

// MockKafka 模拟Kafka客户端
type MockKafka struct {
	mu       sync.RWMutex
	topics   map[string][]MockMessage
	consumer map[string][]chan MockMessage
}

// MockMessage 模拟Kafka消息
type MockMessage struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Timestamp time.Time
}

// NewMockKafka 创建模拟Kafka客户端
func NewMockKafka() *MockKafka {
	return &MockKafka{
		topics:   make(map[string][]MockMessage),
		consumer: make(map[string][]chan MockMessage),
	}
}

// Produce 生产消息
func (m *MockKafka) Produce(ctx context.Context, topic string, key, value []byte, headers map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.topics[topic] == nil {
		m.topics[topic] = make([]MockMessage, 0)
	}
	
	msg := MockMessage{
		Topic:     topic,
		Partition: 0,
		Offset:    int64(len(m.topics[topic])),
		Key:       key,
		Value:     value,
		Headers:   headers,
		Timestamp: time.Now(),
	}
	
	m.topics[topic] = append(m.topics[topic], msg)
	
	// 通知消费者
	if consumers, ok := m.consumer[topic]; ok {
		for _, ch := range consumers {
			select {
			case ch <- msg:
			default:
			}
		}
	}
	
	return nil
}

// Consume 消费消息
func (m *MockKafka) Consume(ctx context.Context, topic string, offset int64) ([]MockMessage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	messages, ok := m.topics[topic]
	if !ok {
		return []MockMessage{}, nil
	}
	
	if offset >= int64(len(messages)) {
		return []MockMessage{}, nil
	}
	
	return messages[offset:], nil
}

// Subscribe 订阅主题
func (m *MockKafka) Subscribe(topic string) <-chan MockMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	ch := make(chan MockMessage, 100)
	if m.consumer[topic] == nil {
		m.consumer[topic] = make([]chan MockMessage, 0)
	}
	m.consumer[topic] = append(m.consumer[topic], ch)
	
	return ch
}

// GetMessages 获取主题所有消息
func (m *MockKafka) GetMessages(topic string) []MockMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	messages, ok := m.topics[topic]
	if !ok {
		return []MockMessage{}
	}
	
	return messages
}

// GetMessageCount 获取消息数量
func (m *MockKafka) GetMessageCount(topic string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	messages, ok := m.topics[topic]
	if !ok {
		return 0
	}
	
	return len(messages)
}

// CreateTopic 创建主题
func (m *MockKafka) CreateTopic(topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.topics[topic] == nil {
		m.topics[topic] = make([]MockMessage, 0)
	}
}

// DeleteTopic 删除主题
func (m *MockKafka) DeleteTopic(topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.topics, topic)
	delete(m.consumer, topic)
}

// ListTopics 列出所有主题
func (m *MockKafka) ListTopics() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	topics := make([]string, 0, len(m.topics))
	for topic := range m.topics {
		topics = append(topics, topic)
	}
	return topics
}

// Clear 清空所有数据
func (m *MockKafka) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.topics = make(map[string][]MockMessage)
	m.consumer = make(map[string][]chan MockMessage)
}

// MockKafkaReader 模拟Kafka Reader
type MockKafkaReader struct {
	kafka  *MockKafka
	topic  string
	offset int64
}

// NewMockKafkaReader 创建模拟Kafka Reader
func NewMockKafkaReader(kafka *MockKafka, topic string) *MockKafkaReader {
	return &MockKafkaReader{
		kafka:  kafka,
		topic:  topic,
		offset: 0,
	}
}

// ReadMessage 读取消息
func (r *MockKafkaReader) ReadMessage(ctx context.Context) (MockMessage, error) {
	messages, err := r.kafka.Consume(ctx, r.topic, r.offset)
	if err != nil {
		return MockMessage{}, err
	}
	
	if len(messages) == 0 {
		// 等待新消息
		ch := r.kafka.Subscribe(r.topic)
		select {
		case msg := <-ch:
			r.offset++
			return msg, nil
		case <-ctx.Done():
			return MockMessage{}, ctx.Err()
		}
	}
	
	msg := messages[0]
	r.offset++
	return msg, nil
}

// Close 关闭Reader
func (r *MockKafkaReader) Close() error {
	return nil
}

// MockKafkaWriter 模拟Kafka Writer
type MockKafkaWriter struct {
	kafka *MockKafka
	topic string
}

// NewMockKafkaWriter 创建模拟Kafka Writer
func NewMockKafkaWriter(kafka *MockKafka, topic string) *MockKafkaWriter {
	return &MockKafkaWriter{
		kafka: kafka,
		topic: topic,
	}
}

// WriteMessages 写入消息
func (w *MockKafkaWriter) WriteMessages(ctx context.Context, messages ...MockMessage) error {
	for _, msg := range messages {
		err := w.kafka.Produce(ctx, w.topic, msg.Key, msg.Value, msg.Headers)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭Writer
func (w *MockKafkaWriter) Close() error {
	return nil
}
