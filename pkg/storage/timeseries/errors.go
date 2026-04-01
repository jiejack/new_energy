package timeseries

import (
	"errors"
	"fmt"
)

// 时序数据库错误定义
var (
	// ErrInvalidConfig 无效配置
	ErrInvalidConfig = errors.New("invalid configuration")
	// ErrUnsupportedDBType 不支持的数据库类型
	ErrUnsupportedDBType = errors.New("unsupported database type")
	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errors.New("connection failed")
	// ErrNotConnected 未连接
	ErrNotConnected = errors.New("not connected")
	// ErrDatabaseNotFound 数据库不存在
	ErrDatabaseNotFound = errors.New("database not found")
	// ErrTableNotFound 表不存在
	ErrTableNotFound = errors.New("table not found")
	// ErrQueryFailed 查询失败
	ErrQueryFailed = errors.New("query failed")
	// ErrWriteFailed 写入失败
	ErrWriteFailed = errors.New("write failed")
	// ErrBatchTooLarge 批次太大
	ErrBatchTooLarge = errors.New("batch too large")
	// ErrTimeout 超时
	ErrTimeout = errors.New("operation timeout")
	// ErrInvalidQuery 无效查询
	ErrInvalidQuery = errors.New("invalid query")
	// ErrInvalidDataPoint 无效数据点
	ErrInvalidDataPoint = errors.New("invalid data point")
	// ErrBufferFull 缓冲区满
	ErrBufferFull = errors.New("buffer full")
	// ErrClosed 已关闭
	ErrClosed = errors.New("client closed")
)

// QueryError 查询错误
type QueryError struct {
	Query   string
	Message string
	Cause   error
}

func (e *QueryError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("query error: %s (query: %s, cause: %v)", e.Message, e.Query, e.Cause)
	}
	return fmt.Sprintf("query error: %s (query: %s)", e.Message, e.Query)
}

func (e *QueryError) Unwrap() error {
	return e.Cause
}

// NewQueryError 创建查询错误
func NewQueryError(query, message string, cause error) *QueryError {
	return &QueryError{
		Query:   query,
		Message: message,
		Cause:   cause,
	}
}

// WriteError 写入错误
type WriteError struct {
	Points  int
	Message string
	Cause   error
}

func (e *WriteError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("write error: %s (points: %d, cause: %v)", e.Message, e.Points, e.Cause)
	}
	return fmt.Sprintf("write error: %s (points: %d)", e.Message, e.Points)
}

func (e *WriteError) Unwrap() error {
	return e.Cause
}

// NewWriteError 创建写入错误
func NewWriteError(points int, message string, cause error) *WriteError {
	return &WriteError{
		Points:  points,
		Message: message,
		Cause:   cause,
	}
}

// ConnectionError 连接错误
type ConnectionError struct {
	Host    string
	Message string
	Cause   error
}

func (e *ConnectionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("connection error: %s (host: %s, cause: %v)", e.Message, e.Host, e.Cause)
	}
	return fmt.Sprintf("connection error: %s (host: %s)", e.Message, e.Host)
}

func (e *ConnectionError) Unwrap() error {
	return e.Cause
}

// NewConnectionError 创建连接错误
func NewConnectionError(host, message string, cause error) *ConnectionError {
	return &ConnectionError{
		Host:    host,
		Message: message,
		Cause:   cause,
	}
}

// IsRetryableError 判断是否可重试错误
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 连接错误可重试
	var connErr *ConnectionError
	if errors.As(err, &connErr) {
		return true
	}

	// 超时错误可重试
	if errors.Is(err, ErrTimeout) {
		return true
	}

	return false
}

// IsNotFoundError 判断是否为未找到错误
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrDatabaseNotFound) || errors.Is(err, ErrTableNotFound)
}
