package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type ErrorMonitor struct {
	client    *redis.Client
	serviceName string
	errorChan chan *ErrorEvent
	buffer    []*ErrorEvent
	mu        sync.Mutex
	flushInterval time.Duration
	maxBufferSize int
}

type ErrorEvent struct {
	ID          string                 `json:"id"`
	ServiceName string                 `json:"service_name"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Stack       string                 `json:"stack,omitempty"`
	File        string                 `json:"file,omitempty"`
	Line        int                    `json:"line,omitempty"`
	Function    string                 `json:"function,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
}

type ErrorMonitorConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	ServiceName   string
	BufferSize    int
	FlushInterval time.Duration
}

func NewErrorMonitor(cfg *ErrorMonitorConfig) (*ErrorMonitor, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	bufferSize := cfg.BufferSize
	if bufferSize <= 0 {
		bufferSize = 100
	}

	flushInterval := cfg.FlushInterval
	if flushInterval <= 0 {
		flushInterval = 5 * time.Second
	}

	em := &ErrorMonitor{
		client:        client,
		serviceName:   cfg.ServiceName,
		errorChan:     make(chan *ErrorEvent, bufferSize),
		buffer:        make([]*ErrorEvent, 0, bufferSize),
		flushInterval: flushInterval,
		maxBufferSize: bufferSize,
	}

	go em.processErrors()
	go em.flushWorker()

	return em, nil
}

func (em *ErrorMonitor) Close() error {
	close(em.errorChan)
	em.flush()
	return em.client.Close()
}

func (em *ErrorMonitor) CaptureError(ctx context.Context, err error, contextData map[string]interface{}) {
	event := em.createErrorEvent("error", err.Error(), contextData)
	em.capture(event)
}

func (em *ErrorMonitor) CaptureWarning(ctx context.Context, message string, contextData map[string]interface{}) {
	event := em.createErrorEvent("warning", message, contextData)
	em.capture(event)
}

func (em *ErrorMonitor) CaptureInfo(ctx context.Context, message string, contextData map[string]interface{}) {
	event := em.createErrorEvent("info", message, contextData)
	em.capture(event)
}

func (em *ErrorMonitor) CapturePanic(ctx context.Context, recovered interface{}, contextData map[string]interface{}) {
	message := fmt.Sprintf("panic: %v", recovered)
	event := em.createErrorEvent("fatal", message, contextData)
	em.capture(event)
}

func (em *ErrorMonitor) createErrorEvent(level, message string, contextData map[string]interface{}) *ErrorEvent {
	_, file, line, ok := runtime.Caller(2)
	function := ""
	if ok {
		function = runtime.FuncForPC(uintptr(0)).Name()
	}

	event := &ErrorEvent{
		ID:          generateID(),
		ServiceName: em.serviceName,
		Timestamp:   time.Now(),
		Level:       level,
		Message:     message,
		File:        file,
		Line:        line,
		Function:    function,
		Context:     contextData,
	}

	if contextData != nil {
		if requestID, ok := contextData["request_id"].(string); ok {
			event.RequestID = requestID
		}
		if userID, ok := contextData["user_id"].(string); ok {
			event.UserID = userID
		}
	}

	return event
}

func (em *ErrorMonitor) capture(event *ErrorEvent) {
	select {
	case em.errorChan <- event:
	default:
		em.mu.Lock()
		if len(em.buffer) < em.maxBufferSize {
			em.buffer = append(em.buffer, event)
		}
		em.mu.Unlock()
	}
}

func (em *ErrorMonitor) processErrors() {
	for event := range em.errorChan {
		em.mu.Lock()
		em.buffer = append(em.buffer, event)
		if len(em.buffer) >= em.maxBufferSize {
			go em.flush()
		}
		em.mu.Unlock()
	}
}

func (em *ErrorMonitor) flushWorker() {
	ticker := time.NewTicker(em.flushInterval)
	defer ticker.Stop()

	for range ticker.C {
		em.flush()
	}
}

func (em *ErrorMonitor) flush() {
	em.mu.Lock()
	if len(em.buffer) == 0 {
		em.mu.Unlock()
		return
	}

	events := make([]*ErrorEvent, len(em.buffer))
	copy(events, em.buffer)
	em.buffer = em.buffer[:0]
	em.mu.Unlock()

	ctx := context.Background()
	pipe := em.client.Pipeline()

	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		key := fmt.Sprintf("errors:%s:%s", em.serviceName, event.ID)
		pipe.Set(ctx, key, data, 24*time.Hour)

		listKey := fmt.Sprintf("errors:%s:list", em.serviceName)
		pipe.LPush(ctx, listKey, event.ID)
		pipe.LTrim(ctx, listKey, 0, 999)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		fmt.Printf("failed to flush errors: %v\n", err)
	}
}

func (em *ErrorMonitor) GetRecentErrors(ctx context.Context, limit int) ([]*ErrorEvent, error) {
	listKey := fmt.Sprintf("errors:%s:list", em.serviceName)
	ids, err := em.client.LRange(ctx, listKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	events := make([]*ErrorEvent, 0, len(ids))
	for _, id := range ids {
		key := fmt.Sprintf("errors:%s:%s", em.serviceName, id)
		data, err := em.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var event ErrorEvent
		if err := json.Unmarshal(data, &event); err != nil {
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}

func (em *ErrorMonitor) GetErrorStats(ctx context.Context, duration time.Duration) (map[string]int64, error) {
	stats := make(map[string]int64)

	events, err := em.GetRecentErrors(ctx, 1000)
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().Add(-duration)
	for _, event := range events {
		if event.Timestamp.Before(cutoff) {
			continue
		}

		stats[event.Level]++
	}

	return stats, nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
