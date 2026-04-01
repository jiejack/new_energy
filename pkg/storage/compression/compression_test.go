package compression

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCompressor(t *testing.T) {
	config := Config{
		Algorithm: "gzip",
		Level:     6,
	}

	compressor, err := NewCompressor(config)
	assert.NoError(t, err)
	assert.NotNil(t, compressor)
}

func TestCompressor_Compress(t *testing.T) {
	config := Config{
		Algorithm: "gzip",
		Level:     6,
	}

	compressor, err := NewCompressor(config)
	assert.NoError(t, err)

	data := []byte("test data for compression testing, this should be compressible data that repeats itself")

	compressed, err := compressor.Compress(data)
	assert.NoError(t, err)
	assert.NotNil(t, compressed)
	// 压缩后应该更小（对于重复数据）
	assert.Less(t, len(compressed), len(data))
}

func TestCompressor_Decompress(t *testing.T) {
	config := Config{
		Algorithm: "gzip",
		Level:     6,
	}

	compressor, err := NewCompressor(config)
	assert.NoError(t, err)

	original := []byte("test data for compression testing")

	compressed, err := compressor.Compress(original)
	assert.NoError(t, err)

	decompressed, err := compressor.Decompress(compressed)
	assert.NoError(t, err)
	assert.Equal(t, original, decompressed)
}

func TestCompressor_CompressDecompress_LargeData(t *testing.T) {
	config := Config{
		Algorithm: "gzip",
		Level:     6,
	}

	compressor, err := NewCompressor(config)
	assert.NoError(t, err)

	// 创建大数据
	original := make([]byte, 10000)
	for i := range original {
		original[i] = byte(i % 256)
	}

	compressed, err := compressor.Compress(original)
	assert.NoError(t, err)

	decompressed, err := compressor.Decompress(compressed)
	assert.NoError(t, err)
	assert.Equal(t, original, decompressed)
}

func TestNewTimeSeriesCompressor(t *testing.T) {
	compressor := NewTimeSeriesCompressor()

	assert.NotNil(t, compressor)
}

func TestTimeSeriesCompressor_Compress(t *testing.T) {
	compressor := NewTimeSeriesCompressor()

	now := time.Now()
	points := []TimeSeriesPoint{
		{Timestamp: now, Value: 100.5},
		{Timestamp: now.Add(1 * time.Second), Value: 101.2},
		{Timestamp: now.Add(2 * time.Second), Value: 102.3},
		{Timestamp: now.Add(3 * time.Second), Value: 103.1},
	}

	compressed, err := compressor.Compress(points)
	assert.NoError(t, err)
	assert.NotNil(t, compressed)
}

func TestTimeSeriesCompressor_Decompress(t *testing.T) {
	compressor := NewTimeSeriesCompressor()

	now := time.Now()
	original := []TimeSeriesPoint{
		{Timestamp: now, Value: 100.5},
		{Timestamp: now.Add(1 * time.Second), Value: 101.2},
		{Timestamp: now.Add(2 * time.Second), Value: 102.3},
	}

	compressed, err := compressor.Compress(original)
	assert.NoError(t, err)

	decompressed, err := compressor.Decompress(compressed)
	assert.NoError(t, err)
	assert.Equal(t, len(original), len(decompressed))

	for i := range original {
		assert.Equal(t, original[i].Timestamp.Unix(), decompressed[i].Timestamp.Unix())
		assert.InDelta(t, original[i].Value, decompressed[i].Value, 0.01)
	}
}

func TestDeltaEncoding(t *testing.T) {
	now := time.Now()
	points := []TimeSeriesPoint{
		{Timestamp: now, Value: 100.0},
		{Timestamp: now.Add(1 * time.Second), Value: 101.0},
		{Timestamp: now.Add(2 * time.Second), Value: 102.0},
		{Timestamp: now.Add(3 * time.Second), Value: 103.0},
	}

	encoded := deltaEncode(points)
	assert.NotNil(t, encoded)

	decoded := deltaDecode(encoded, now)
	assert.Equal(t, len(points), len(decoded))

	for i := range points {
		assert.Equal(t, points[i].Timestamp.Unix(), decoded[i].Timestamp.Unix())
		assert.InDelta(t, points[i].Value, decoded[i].Value, 0.01)
	}
}

func TestXOREncoding(t *testing.T) {
	values := []float64{100.5, 100.6, 100.7, 100.8, 100.9}

	encoded := xorEncode(values)
	assert.NotNil(t, encoded)

	decoded := xorDecode(encoded, values[0])
	assert.Equal(t, len(values), len(decoded))

	for i := range values {
		assert.InDelta(t, values[i], decoded[i], 0.01)
	}
}

func TestCompressionRatio(t *testing.T) {
	config := Config{
		Algorithm: "gzip",
		Level:     9,
	}

	compressor, err := NewCompressor(config)
	assert.NoError(t, err)

	// 创建高度可压缩的数据
	original := bytes.Repeat([]byte("abcdefgh"), 1000)

	compressed, err := compressor.Compress(original)
	assert.NoError(t, err)

	ratio := float64(len(compressed)) / float64(len(original))
	assert.Less(t, ratio, 0.5) // 压缩率应该小于50%
}
