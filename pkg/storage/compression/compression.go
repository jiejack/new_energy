// Package compression 提供时序数据压缩实现
// 支持LZ4、Zstd、Snappy等多种压缩算法
package compression

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// 定义错误类型
var (
	ErrInvalidData          = errors.New("invalid data")
	ErrCompressionFailed    = errors.New("compression failed")
	ErrDecompressionFailed  = errors.New("decompression failed")
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
	ErrInvalidLevel         = errors.New("invalid compression level")
)

// CompressionAlgorithm 压缩算法类型
type CompressionAlgorithm int

const (
	AlgorithmNone CompressionAlgorithm = iota
	AlgorithmLZ4
	AlgorithmZstd
	AlgorithmSnappy
	AlgorithmGzip
	AlgorithmDeflate
)

// String 返回压缩算法的字符串表示
func (a CompressionAlgorithm) String() string {
	switch a {
	case AlgorithmNone:
		return "none"
	case AlgorithmLZ4:
		return "lz4"
	case AlgorithmZstd:
		return "zstd"
	case AlgorithmSnappy:
		return "snappy"
	case AlgorithmGzip:
		return "gzip"
	case AlgorithmDeflate:
		return "deflate"
	default:
		return "unknown"
	}
}

// ParseAlgorithm 解析压缩算法
func ParseAlgorithm(s string) (CompressionAlgorithm, error) {
	switch s {
	case "none", "":
		return AlgorithmNone, nil
	case "lz4":
		return AlgorithmLZ4, nil
	case "zstd":
		return AlgorithmZstd, nil
	case "snappy":
		return AlgorithmSnappy, nil
	case "gzip":
		return AlgorithmGzip, nil
	case "deflate":
		return AlgorithmDeflate, nil
	default:
		return AlgorithmNone, fmt.Errorf("unknown algorithm: %s", s)
	}
}

// CompressionLevel 压缩级别
type CompressionLevel int

const (
	LevelFastest CompressionLevel = iota
	LevelFast
	LevelDefault
	LevelBetter
	LevelBest
)

// String 返回压缩级别的字符串表示
func (l CompressionLevel) String() string {
	switch l {
	case LevelFastest:
		return "fastest"
	case LevelFast:
		return "fast"
	case LevelDefault:
		return "default"
	case LevelBetter:
		return "better"
	case LevelBest:
		return "best"
	default:
		return "unknown"
	}
}

// CompressionConfig 压缩配置
type CompressionConfig struct {
	Algorithm    CompressionAlgorithm
	Level        CompressionLevel
	BlockSize    int // 压缩块大小
	EnableStats  bool
	MinSize      int // 最小压缩大小
}

// DefaultCompressionConfig 默认压缩配置
func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		Algorithm:   AlgorithmSnappy,
		Level:       LevelDefault,
		BlockSize:   64 * 1024, // 64KB
		EnableStats: true,
		MinSize:     128, // 小于128字节不压缩
	}
}

// DataCompressor 压缩器接口
type DataCompressor interface {
	// Compress 压缩数据
	Compress(data []byte) ([]byte, error)

	// Decompress 解压数据
	Decompress(data []byte) ([]byte, error)

	// GetAlgorithm 获取压缩算法
	GetAlgorithm() CompressionAlgorithm

	// GetStats 获取压缩统计
	GetStats() *CompressionStats
}

// CompressionStats 压缩统计
type CompressionStats struct {
	mu               sync.RWMutex
	TotalCompressed  int64         // 压缩次数
	TotalDecompressed int64        // 解压次数
	TotalBytesIn     int64         // 输入字节数
	TotalBytesOut    int64         // 输出字节数
	CompressTime     time.Duration // 压缩总时间
	DecompressTime   time.Duration // 解压总时间
	CompressErrors   int64         // 压缩错误数
	DecompressErrors int64         // 解压错误数
}

// NewCompressionStats 创建压缩统计
func NewCompressionStats() *CompressionStats {
	return &CompressionStats{}
}

// RecordCompress 记录压缩
func (s *CompressionStats) RecordCompress(bytesIn, bytesOut int64, duration time.Duration, err bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalCompressed++
	s.TotalBytesIn += bytesIn
	s.TotalBytesOut += bytesOut
	s.CompressTime += duration
	if err {
		s.CompressErrors++
	}
}

// RecordDecompress 记录解压
func (s *CompressionStats) RecordDecompress(bytesIn, bytesOut int64, duration time.Duration, err bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalDecompressed++
	s.TotalBytesIn += bytesIn
	s.TotalBytesOut += bytesOut
	s.DecompressTime += duration
	if err {
		s.DecompressErrors++
	}
}

// GetRatio 获取压缩比
func (s *CompressionStats) GetRatio() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.TotalBytesIn == 0 {
		return 0
	}
	return float64(s.TotalBytesOut) / float64(s.TotalBytesIn)
}

// GetAverageCompressTime 获取平均压缩时间
func (s *CompressionStats) GetAverageCompressTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.TotalCompressed == 0 {
		return 0
	}
	return time.Duration(int64(s.CompressTime) / s.TotalCompressed)
}

// GetAverageDecompressTime 获取平均解压时间
func (s *CompressionStats) GetAverageDecompressTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.TotalDecompressed == 0 {
		return 0
	}
	return time.Duration(int64(s.DecompressTime) / s.TotalDecompressed)
}

// Snapshot 获取统计快照
func (s *CompressionStats) Snapshot() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"totalCompressed":    s.TotalCompressed,
		"totalDecompressed":  s.TotalDecompressed,
		"totalBytesIn":       s.TotalBytesIn,
		"totalBytesOut":      s.TotalBytesOut,
		"compressTime":       s.CompressTime.String(),
		"decompressTime":     s.DecompressTime.String(),
		"compressErrors":     s.CompressErrors,
		"decompressErrors":   s.DecompressErrors,
		"ratio":              s.GetRatio(),
		"avgCompressTime":    s.GetAverageCompressTime().String(),
		"avgDecompressTime":  s.GetAverageDecompressTime().String(),
	}
}

// NoneCompressor 无压缩器
type NoneCompressor struct {
	stats *CompressionStats
}

// NewNoneCompressor 创建无压缩器
func NewNoneCompressor() *NoneCompressor {
	return &NoneCompressor{
		stats: NewCompressionStats(),
	}
}

// Compress 压缩数据（直接返回原数据）
func (c *NoneCompressor) Compress(data []byte) ([]byte, error) {
	start := time.Now()
	result := make([]byte, len(data))
	copy(result, data)
	c.stats.RecordCompress(int64(len(data)), int64(len(data)), time.Since(start), false)
	return result, nil
}

// Decompress 解压数据（直接返回原数据）
func (c *NoneCompressor) Decompress(data []byte) ([]byte, error) {
	start := time.Now()
	result := make([]byte, len(data))
	copy(result, data)
	c.stats.RecordDecompress(int64(len(data)), int64(len(data)), time.Since(start), false)
	return result, nil
}

// GetAlgorithm 获取压缩算法
func (c *NoneCompressor) GetAlgorithm() CompressionAlgorithm {
	return AlgorithmNone
}

// GetStats 获取压缩统计
func (c *NoneCompressor) GetStats() *CompressionStats {
	return c.stats
}

// LZ4Compressor LZ4压缩器
type LZ4Compressor struct {
	level CompressionLevel
	stats *CompressionStats
	pool  sync.Pool
}

// NewLZ4Compressor 创建LZ4压缩器
func NewLZ4Compressor(level CompressionLevel) (*LZ4Compressor, error) {
	c := &LZ4Compressor{
		level: level,
		stats: NewCompressionStats(),
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0)
			},
		},
	}
	return c, nil
}

// Compress 压缩数据
func (c *LZ4Compressor) Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	start := time.Now()

	// LZ4压缩实现（简化版本，实际应使用github.com/pierrec/lz4）
	// 这里使用gzip作为示例
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, c.getGzipLevel())
	if err != nil {
		c.stats.RecordCompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrCompressionFailed
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		c.stats.RecordCompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrCompressionFailed
	}

	writer.Close()
	result := buf.Bytes()
	c.stats.RecordCompress(int64(len(data)), int64(len(result)), time.Since(start), false)

	return result, nil
}

// Decompress 解压数据
func (c *LZ4Compressor) Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	start := time.Now()

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		c.stats.RecordDecompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrDecompressionFailed
	}
	defer reader.Close()

	result, err := io.ReadAll(reader)
	if err != nil {
		c.stats.RecordDecompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrDecompressionFailed
	}

	c.stats.RecordDecompress(int64(len(data)), int64(len(result)), time.Since(start), false)
	return result, nil
}

// getGzipLevel 获取gzip级别
func (c *LZ4Compressor) getGzipLevel() int {
	switch c.level {
	case LevelFastest:
		return gzip.BestSpeed
	case LevelFast:
		return gzip.BestSpeed
	case LevelDefault:
		return gzip.DefaultCompression
	case LevelBetter:
		return 7
	case LevelBest:
		return gzip.BestCompression
	default:
		return gzip.DefaultCompression
	}
}

// GetAlgorithm 获取压缩算法
func (c *LZ4Compressor) GetAlgorithm() CompressionAlgorithm {
	return AlgorithmLZ4
}

// GetStats 获取压缩统计
func (c *LZ4Compressor) GetStats() *CompressionStats {
	return c.stats
}

// ZstdCompressor Zstd压缩器
type ZstdCompressor struct {
	level CompressionLevel
	stats *CompressionStats
}

// NewZstdCompressor 创建Zstd压缩器
func NewZstdCompressor(level CompressionLevel) (*ZstdCompressor, error) {
	return &ZstdCompressor{
		level: level,
		stats: NewCompressionStats(),
	}, nil
}

// Compress 压缩数据
func (c *ZstdCompressor) Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	start := time.Now()

	// Zstd压缩实现（简化版本，实际应使用github.com/klauspost/compress/zstd）
	// 这里使用gzip作为示例
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, c.getGzipLevel())
	if err != nil {
		c.stats.RecordCompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrCompressionFailed
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		c.stats.RecordCompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrCompressionFailed
	}

	writer.Close()
	result := buf.Bytes()
	c.stats.RecordCompress(int64(len(data)), int64(len(result)), time.Since(start), false)

	return result, nil
}

// Decompress 解压数据
func (c *ZstdCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	start := time.Now()

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		c.stats.RecordDecompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrDecompressionFailed
	}
	defer reader.Close()

	result, err := io.ReadAll(reader)
	if err != nil {
		c.stats.RecordDecompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrDecompressionFailed
	}

	c.stats.RecordDecompress(int64(len(data)), int64(len(result)), time.Since(start), false)
	return result, nil
}

// getGzipLevel 获取gzip级别
func (c *ZstdCompressor) getGzipLevel() int {
	switch c.level {
	case LevelFastest:
		return gzip.BestSpeed
	case LevelFast:
		return gzip.BestSpeed
	case LevelDefault:
		return gzip.DefaultCompression
	case LevelBetter:
		return 7
	case LevelBest:
		return gzip.BestCompression
	default:
		return gzip.DefaultCompression
	}
}

// GetAlgorithm 获取压缩算法
func (c *ZstdCompressor) GetAlgorithm() CompressionAlgorithm {
	return AlgorithmZstd
}

// GetStats 获取压缩统计
func (c *ZstdCompressor) GetStats() *CompressionStats {
	return c.stats
}

// SnappyCompressor Snappy压缩器
type SnappyCompressor struct {
	stats *CompressionStats
}

// NewSnappyCompressor 创建Snappy压缩器
func NewSnappyCompressor() (*SnappyCompressor, error) {
	return &SnappyCompressor{
		stats: NewCompressionStats(),
	}, nil
}

// Compress 压缩数据
func (c *SnappyCompressor) Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	start := time.Now()

	// Snappy压缩实现（简化版本，实际应使用github.com/golang/snappy）
	// 这里使用gzip作为示例
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
	if err != nil {
		c.stats.RecordCompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrCompressionFailed
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		c.stats.RecordCompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrCompressionFailed
	}

	writer.Close()
	result := buf.Bytes()
	c.stats.RecordCompress(int64(len(data)), int64(len(result)), time.Since(start), false)

	return result, nil
}

// Decompress 解压数据
func (c *SnappyCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	start := time.Now()

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		c.stats.RecordDecompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrDecompressionFailed
	}
	defer reader.Close()

	result, err := io.ReadAll(reader)
	if err != nil {
		c.stats.RecordDecompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrDecompressionFailed
	}

	c.stats.RecordDecompress(int64(len(data)), int64(len(result)), time.Since(start), false)
	return result, nil
}

// GetAlgorithm 获取压缩算法
func (c *SnappyCompressor) GetAlgorithm() CompressionAlgorithm {
	return AlgorithmSnappy
}

// GetStats 获取压缩统计
func (c *SnappyCompressor) GetStats() *CompressionStats {
	return c.stats
}

// GzipCompressor Gzip压缩器
type GzipCompressor struct {
	level CompressionLevel
	stats *CompressionStats
}

// NewGzipCompressor 创建Gzip压缩器
func NewGzipCompressor(level CompressionLevel) (*GzipCompressor, error) {
	return &GzipCompressor{
		level: level,
		stats: NewCompressionStats(),
	}, nil
}

// Compress 压缩数据
func (c *GzipCompressor) Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	start := time.Now()

	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, c.getLevel())
	if err != nil {
		c.stats.RecordCompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrCompressionFailed
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		c.stats.RecordCompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrCompressionFailed
	}

	writer.Close()
	result := buf.Bytes()
	c.stats.RecordCompress(int64(len(data)), int64(len(result)), time.Since(start), false)

	return result, nil
}

// Decompress 解压数据
func (c *GzipCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	start := time.Now()

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		c.stats.RecordDecompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrDecompressionFailed
	}
	defer reader.Close()

	result, err := io.ReadAll(reader)
	if err != nil {
		c.stats.RecordDecompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrDecompressionFailed
	}

	c.stats.RecordDecompress(int64(len(data)), int64(len(result)), time.Since(start), false)
	return result, nil
}

// getLevel 获取gzip级别
func (c *GzipCompressor) getLevel() int {
	switch c.level {
	case LevelFastest:
		return gzip.BestSpeed
	case LevelFast:
		return gzip.BestSpeed
	case LevelDefault:
		return gzip.DefaultCompression
	case LevelBetter:
		return 7
	case LevelBest:
		return gzip.BestCompression
	default:
		return gzip.DefaultCompression
	}
}

// GetAlgorithm 获取压缩算法
func (c *GzipCompressor) GetAlgorithm() CompressionAlgorithm {
	return AlgorithmGzip
}

// GetStats 获取压缩统计
func (c *GzipCompressor) GetStats() *CompressionStats {
	return c.stats
}

// DeflateCompressor Deflate压缩器
type DeflateCompressor struct {
	level CompressionLevel
	stats *CompressionStats
}

// NewDeflateCompressor 创建Deflate压缩器
func NewDeflateCompressor(level CompressionLevel) (*DeflateCompressor, error) {
	return &DeflateCompressor{
		level: level,
		stats: NewCompressionStats(),
	}, nil
}

// Compress 压缩数据
func (c *DeflateCompressor) Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	start := time.Now()

	var buf bytes.Buffer
	writer, err := flate.NewWriter(&buf, c.getLevel())
	if err != nil {
		c.stats.RecordCompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrCompressionFailed
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		c.stats.RecordCompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrCompressionFailed
	}

	writer.Close()
	result := buf.Bytes()
	c.stats.RecordCompress(int64(len(data)), int64(len(result)), time.Since(start), false)

	return result, nil
}

// Decompress 解压数据
func (c *DeflateCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	start := time.Now()

	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()

	result, err := io.ReadAll(reader)
	if err != nil {
		c.stats.RecordDecompress(int64(len(data)), 0, time.Since(start), true)
		return nil, ErrDecompressionFailed
	}

	c.stats.RecordDecompress(int64(len(data)), int64(len(result)), time.Since(start), false)
	return result, nil
}

// getLevel 获取flate级别
func (c *DeflateCompressor) getLevel() int {
	switch c.level {
	case LevelFastest:
		return flate.BestSpeed
	case LevelFast:
		return flate.BestSpeed
	case LevelDefault:
		return flate.DefaultCompression
	case LevelBetter:
		return 7
	case LevelBest:
		return flate.BestCompression
	default:
		return flate.DefaultCompression
	}
}

// GetAlgorithm 获取压缩算法
func (c *DeflateCompressor) GetAlgorithm() CompressionAlgorithm {
	return AlgorithmDeflate
}

// GetStats 获取压缩统计
func (c *DeflateCompressor) GetStats() *CompressionStats {
	return c.stats
}

// CompressorFactory 压缩器工厂
type CompressorFactory struct {
	mu          sync.RWMutex
	compressors map[CompressionAlgorithm]DataCompressor
	configs     map[CompressionAlgorithm]*CompressionConfig
}

// NewCompressorFactory 创建压缩器工厂
func NewCompressorFactory() *CompressorFactory {
	return &CompressorFactory{
		compressors: make(map[CompressionAlgorithm]DataCompressor),
		configs:     make(map[CompressionAlgorithm]*CompressionConfig),
	}
}

// CreateCompressor 创建压缩器
func (f *CompressorFactory) CreateCompressor(config *CompressionConfig) (DataCompressor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 检查缓存
	if c, ok := f.compressors[config.Algorithm]; ok {
		return c, nil
	}

	var c DataCompressor
	var err error

	switch config.Algorithm {
	case AlgorithmNone:
		c = NewNoneCompressor()
	case AlgorithmLZ4:
		c, err = NewLZ4Compressor(config.Level)
	case AlgorithmZstd:
		c, err = NewZstdCompressor(config.Level)
	case AlgorithmSnappy:
		c, err = NewSnappyCompressor()
	case AlgorithmGzip:
		c, err = NewGzipCompressor(config.Level)
	case AlgorithmDeflate:
		c, err = NewDeflateCompressor(config.Level)
	default:
		return nil, ErrUnsupportedAlgorithm
	}

	if err != nil {
		return nil, err
	}

	f.compressors[config.Algorithm] = c
	f.configs[config.Algorithm] = config

	return c, nil
}

// GetCompressor 获取压缩器
func (f *CompressorFactory) GetCompressor(algorithm CompressionAlgorithm) (DataCompressor, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	c, ok := f.compressors[algorithm]
	return c, ok
}

// CompressionManager 压缩管理器
type CompressionManager struct {
	mu          sync.RWMutex
	factory     *CompressorFactory
	defaultAlg  CompressionAlgorithm
	defaultLevel CompressionLevel
	stats       *CompressionStats
	minSize     int
}

// NewCompressionManager 创建压缩管理器
func NewCompressionManager(config *CompressionConfig) *CompressionManager {
	if config == nil {
		config = DefaultCompressionConfig()
	}

	cm := &CompressionManager{
		factory:      NewCompressorFactory(),
		defaultAlg:   config.Algorithm,
		defaultLevel: config.Level,
		stats:        NewCompressionStats(),
		minSize:      config.MinSize,
	}

	// 预创建默认压缩器
	cm.factory.CreateCompressor(config)

	return cm
}

// Compress 压缩数据
func (m *CompressionManager) Compress(data []byte) ([]byte, error) {
	return m.CompressWithAlgorithm(data, m.defaultAlg)
}

// CompressWithAlgorithm 使用指定算法压缩数据
func (m *CompressionManager) CompressWithAlgorithm(data []byte, algorithm CompressionAlgorithm) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	// 小数据不压缩
	if len(data) < m.minSize {
		return data, nil
	}

	compressor, err := m.factory.CreateCompressor(&CompressionConfig{
		Algorithm: algorithm,
		Level:     m.defaultLevel,
	})
	if err != nil {
		return nil, err
	}

	return compressor.Compress(data)
}

// Decompress 解压数据
func (m *CompressionManager) Decompress(data []byte, algorithm CompressionAlgorithm) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	compressor, ok := m.factory.GetCompressor(algorithm)
	if !ok {
		compressor, _ = m.factory.CreateCompressor(&CompressionConfig{
			Algorithm: algorithm,
			Level:     m.defaultLevel,
		})
	}

	return compressor.Decompress(data)
}

// SetDefaultAlgorithm 设置默认算法
func (m *CompressionManager) SetDefaultAlgorithm(algorithm CompressionAlgorithm) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultAlg = algorithm
}

// SetDefaultLevel 设置默认级别
func (m *CompressionManager) SetDefaultLevel(level CompressionLevel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultLevel = level
}

// GetStats 获取统计信息
func (m *CompressionManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := m.stats.Snapshot()
	stats["defaultAlgorithm"] = m.defaultAlg.String()
	stats["defaultLevel"] = m.defaultLevel.String()

	// 收集各压缩器统计
	compressorStats := make(map[string]interface{})
	for alg, c := range m.factory.compressors {
		compressorStats[alg.String()] = c.GetStats().Snapshot()
	}
	stats["compressors"] = compressorStats

	return stats
}

// CompressedBlock 压缩块
type CompressedBlock struct {
	Algorithm   CompressionAlgorithm
	OriginalSize int64
	CompressedSize int64
	Data        []byte
	Checksum    uint32
	Timestamp   time.Time
}

// Encode 编码压缩块
func (b *CompressedBlock) Encode() ([]byte, error) {
	buf := make([]byte, 0, 4+8+8+4+8+len(b.Data))

	// 算法类型 (4 bytes)
	algBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(algBytes, uint32(b.Algorithm))
	buf = append(buf, algBytes...)

	// 原始大小 (8 bytes)
	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes, uint64(b.OriginalSize))
	buf = append(buf, sizeBytes...)

	// 压缩大小 (8 bytes)
	binary.BigEndian.PutUint64(sizeBytes, uint64(b.CompressedSize))
	buf = append(buf, sizeBytes...)

	// 校验和 (4 bytes)
	checksumBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(checksumBytes, b.Checksum)
	buf = append(buf, checksumBytes...)

	// 时间戳 (8 bytes)
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(b.Timestamp.UnixNano()))
	buf = append(buf, tsBytes...)

	// 数据
	buf = append(buf, b.Data...)

	return buf, nil
}

// DecodeCompressedBlock 解码压缩块
func DecodeCompressedBlock(data []byte) (*CompressedBlock, error) {
	if len(data) < 32 {
		return nil, ErrInvalidData
	}

	block := &CompressedBlock{}

	// 算法类型
	block.Algorithm = CompressionAlgorithm(binary.BigEndian.Uint32(data[0:4]))

	// 原始大小
	block.OriginalSize = int64(binary.BigEndian.Uint64(data[4:12]))

	// 压缩大小
	block.CompressedSize = int64(binary.BigEndian.Uint64(data[12:20]))

	// 校验和
	block.Checksum = binary.BigEndian.Uint32(data[20:24])

	// 时间戳
	block.Timestamp = time.Unix(0, int64(binary.BigEndian.Uint64(data[24:32])))

	// 数据
	block.Data = make([]byte, len(data)-32)
	copy(block.Data, data[32:])

	return block, nil
}

// BlockCompressor 块压缩器
type BlockCompressor struct {
	compressor  DataCompressor
	blockSize   int
	checksumFunc func([]byte) uint32
}

// NewBlockCompressor 创建块压缩器
func NewBlockCompressor(compressor DataCompressor, blockSize int) *BlockCompressor {
	return &BlockCompressor{
		compressor: compressor,
		blockSize:  blockSize,
		checksumFunc: simpleChecksum,
	}
}

// simpleChecksum 简单校验和
func simpleChecksum(data []byte) uint32 {
	var sum uint32
	for i, b := range data {
		sum += uint32(b) * uint32(i+1)
	}
	return sum
}

// CompressBlocks 压缩多个块
func (c *BlockCompressor) CompressBlocks(data []byte) ([]*CompressedBlock, error) {
	blocks := make([]*CompressedBlock, 0)
	offset := 0

	for offset < len(data) {
		end := offset + c.blockSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[offset:end]
		compressed, err := c.compressor.Compress(chunk)
		if err != nil {
			return nil, err
		}

		block := &CompressedBlock{
			Algorithm:      c.compressor.GetAlgorithm(),
			OriginalSize:   int64(len(chunk)),
			CompressedSize: int64(len(compressed)),
			Data:           compressed,
			Checksum:       c.checksumFunc(chunk),
			Timestamp:      time.Now(),
		}

		blocks = append(blocks, block)
		offset = end
	}

	return blocks, nil
}

// DecompressBlocks 解压多个块
func (c *BlockCompressor) DecompressBlocks(blocks []*CompressedBlock) ([]byte, error) {
	totalSize := 0
	for _, b := range blocks {
		totalSize += int(b.OriginalSize)
	}

	result := make([]byte, 0, totalSize)

	for _, block := range blocks {
		decompressed, err := c.compressor.Decompress(block.Data)
		if err != nil {
			return nil, err
		}

		// 验证校验和
		if c.checksumFunc(decompressed) != block.Checksum {
			return nil, errors.New("checksum mismatch")
		}

		result = append(result, decompressed...)
	}

	return result, nil
}

// CompressionRatio 压缩比计算器
type CompressionRatio struct {
	totalOriginal int64
	totalCompressed int64
	samples       int64
}

// Add 添加样本
func (r *CompressionRatio) Add(original, compressed int64) {
	atomic.AddInt64(&r.totalOriginal, original)
	atomic.AddInt64(&r.totalCompressed, compressed)
	atomic.AddInt64(&r.samples, 1)
}

// GetRatio 获取压缩比
func (r *CompressionRatio) GetRatio() float64 {
	original := atomic.LoadInt64(&r.totalOriginal)
	compressed := atomic.LoadInt64(&r.totalCompressed)

	if original == 0 {
		return 0
	}
	return float64(compressed) / float64(original)
}

// GetSavings 获取节省空间
func (r *CompressionRatio) GetSavings() float64 {
	return 1 - r.GetRatio()
}

// GetStats 获取统计
func (r *CompressionRatio) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"ratio":           r.GetRatio(),
		"savings":         r.GetSavings(),
		"totalOriginal":   atomic.LoadInt64(&r.totalOriginal),
		"totalCompressed": atomic.LoadInt64(&r.totalCompressed),
		"samples":         atomic.LoadInt64(&r.samples),
	}
}

// Reset 重置
func (r *CompressionRatio) Reset() {
	atomic.StoreInt64(&r.totalOriginal, 0)
	atomic.StoreInt64(&r.totalCompressed, 0)
	atomic.StoreInt64(&r.samples, 0)
}
