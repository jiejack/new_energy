package compression

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressionAlgorithm_String(t *testing.T) {
	tests := []struct {
		alg CompressionAlgorithm
		exp string
	}{
		{AlgorithmNone, "none"},
		{AlgorithmLZ4, "lz4"},
		{AlgorithmZstd, "zstd"},
		{AlgorithmSnappy, "snappy"},
		{AlgorithmGzip, "gzip"},
		{AlgorithmDeflate, "deflate"},
		{CompressionAlgorithm(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.exp, tt.alg.String())
	}
}

func TestParseAlgorithm(t *testing.T) {
	tests := []struct {
		input   string
		exp     CompressionAlgorithm
		hasErr  bool
	}{
		{"none", AlgorithmNone, false},
		{"", AlgorithmNone, false},
		{"lz4", AlgorithmLZ4, false},
		{"zstd", AlgorithmZstd, false},
		{"snappy", AlgorithmSnappy, false},
		{"gzip", AlgorithmGzip, false},
		{"deflate", AlgorithmDeflate, false},
		{"unknown", AlgorithmNone, true},
	}
	for _, tt := range tests {
		alg, err := ParseAlgorithm(tt.input)
		if tt.hasErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, tt.exp, alg)
	}
}

func TestCompressionLevel_String(t *testing.T) {
	tests := []struct {
		lvl CompressionLevel
		exp string
	}{
		{LevelFastest, "fastest"},
		{LevelFast, "fast"},
		{LevelDefault, "default"},
		{LevelBetter, "better"},
		{LevelBest, "best"},
		{CompressionLevel(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.exp, tt.lvl.String())
	}
}

func TestDefaultCompressionConfig(t *testing.T) {
	cfg := DefaultCompressionConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, AlgorithmSnappy, cfg.Algorithm)
	assert.Equal(t, LevelDefault, cfg.Level)
	assert.Equal(t, 64*1024, cfg.BlockSize)
	assert.True(t, cfg.EnableStats)
	assert.Equal(t, 128, cfg.MinSize)
}

func TestNewCompressionStats(t *testing.T) {
	stats := NewCompressionStats()
	require.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalCompressed)
	assert.Equal(t, int64(0), stats.TotalDecompressed)
}

func TestCompressionStats_RecordCompress(t *testing.T) {
	stats := NewCompressionStats()
	stats.RecordCompress(1000, 500, 0, false)
	assert.Equal(t, int64(1), stats.TotalCompressed)
	assert.Equal(t, int64(1000), stats.TotalBytesIn)
	assert.Equal(t, int64(500), stats.TotalBytesOut)
	assert.Equal(t, int64(0), stats.CompressErrors)

	stats.RecordCompress(500, 0, 0, true)
	assert.Equal(t, int64(2), stats.TotalCompressed)
	assert.Equal(t, int64(1), stats.CompressErrors)
}

func TestCompressionStats_RecordDecompress(t *testing.T) {
	stats := NewCompressionStats()
	stats.RecordDecompress(500, 1000, 0, false)
	assert.Equal(t, int64(1), stats.TotalDecompressed)
	assert.Equal(t, int64(0), stats.DecompressErrors)

	stats.RecordDecompress(500, 0, 0, true)
	assert.Equal(t, int64(1), stats.DecompressErrors)
}

func TestCompressionStats_GetRatio(t *testing.T) {
	stats := NewCompressionStats()
	assert.Equal(t, float64(0), stats.GetRatio())

	stats.RecordCompress(1000, 500, 0, false)
	assert.InDelta(t, 0.5, stats.GetRatio(), 0.01)
}

func TestCompressionStats_GetAverageCompressTime(t *testing.T) {
	stats := NewCompressionStats()
	assert.Equal(t, int64(0), int64(stats.GetAverageCompressTime()))

	stats.RecordCompress(100, 50, 1000, false)
	assert.Equal(t, int64(1000), int64(stats.GetAverageCompressTime()))
}

func TestCompressionStats_GetAverageDecompressTime(t *testing.T) {
	stats := NewCompressionStats()
	assert.Equal(t, int64(0), int64(stats.GetAverageDecompressTime()))

	stats.RecordDecompress(50, 100, 2000, false)
	assert.Equal(t, int64(2000), int64(stats.GetAverageDecompressTime()))
}

func TestCompressionStats_Snapshot(t *testing.T) {
	stats := NewCompressionStats()
	stats.RecordCompress(1000, 500, 100, false)
	snap := stats.Snapshot()
	assert.NotNil(t, snap)
	assert.Equal(t, int64(1), snap["totalCompressed"])
}

func TestNoneCompressor(t *testing.T) {
	c := NewNoneCompressor()
	require.NotNil(t, c)
	assert.Equal(t, AlgorithmNone, c.GetAlgorithm())

	data := []byte("hello world")
	compressed, err := c.Compress(data)
	require.NoError(t, err)
	assert.Equal(t, data, compressed)

	decompressed, err := c.Decompress(compressed)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)

	stats := c.GetStats()
	require.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.TotalCompressed)
	assert.Equal(t, int64(1), stats.TotalDecompressed)
}

func TestLZ4Compressor(t *testing.T) {
	c, err := NewLZ4Compressor(LevelDefault)
	require.NoError(t, err)
	assert.Equal(t, AlgorithmLZ4, c.GetAlgorithm())

	data := []byte("hello world, this is a test for lz4 compression")
	compressed, err := c.Compress(data)
	require.NoError(t, err)

	decompressed, err := c.Decompress(compressed)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)

	_, err = c.Compress([]byte{})
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidData, err)

	_, err = c.Decompress([]byte{})
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidData, err)
}

func TestZstdCompressor(t *testing.T) {
	c, err := NewZstdCompressor(LevelDefault)
	require.NoError(t, err)
	assert.Equal(t, AlgorithmZstd, c.GetAlgorithm())

	data := []byte("hello world, this is a test for zstd compression")
	compressed, err := c.Compress(data)
	require.NoError(t, err)

	decompressed, err := c.Decompress(compressed)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestSnappyCompressor(t *testing.T) {
	c, err := NewSnappyCompressor()
	require.NoError(t, err)
	assert.Equal(t, AlgorithmSnappy, c.GetAlgorithm())

	data := []byte("hello world, this is a test for snappy compression")
	compressed, err := c.Compress(data)
	require.NoError(t, err)

	decompressed, err := c.Decompress(compressed)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestGzipCompressor(t *testing.T) {
	c, err := NewGzipCompressor(LevelDefault)
	require.NoError(t, err)
	assert.Equal(t, AlgorithmGzip, c.GetAlgorithm())

	data := []byte("hello world, this is a test for gzip compression")
	compressed, err := c.Compress(data)
	require.NoError(t, err)

	decompressed, err := c.Decompress(compressed)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)

	_, err = c.Compress([]byte{})
	assert.Equal(t, ErrInvalidData, err)
}

func TestDeflateCompressor(t *testing.T) {
	c, err := NewDeflateCompressor(LevelDefault)
	require.NoError(t, err)
	assert.Equal(t, AlgorithmDeflate, c.GetAlgorithm())

	data := []byte("hello world, this is a test for deflate compression")
	compressed, err := c.Compress(data)
	require.NoError(t, err)

	decompressed, err := c.Decompress(compressed)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestCompressorFactory(t *testing.T) {
	f := NewCompressorFactory()
	require.NotNil(t, f)

	c, err := f.CreateCompressor(&CompressionConfig{Algorithm: AlgorithmNone})
	require.NoError(t, err)
	assert.Equal(t, AlgorithmNone, c.GetAlgorithm())

	got, ok := f.GetCompressor(AlgorithmNone)
	assert.True(t, ok)
	assert.Equal(t, c, got)

	_, ok = f.GetCompressor(AlgorithmGzip)
	assert.False(t, ok)

	c, err = f.CreateCompressor(&CompressionConfig{Algorithm: AlgorithmGzip, Level: LevelDefault})
	require.NoError(t, err)
	assert.Equal(t, AlgorithmGzip, c.GetAlgorithm())

	c, err = f.CreateCompressor(&CompressionConfig{Algorithm: AlgorithmLZ4, Level: LevelFastest})
	require.NoError(t, err)
	assert.Equal(t, AlgorithmLZ4, c.GetAlgorithm())

	c, err = f.CreateCompressor(&CompressionConfig{Algorithm: AlgorithmZstd, Level: LevelBest})
	require.NoError(t, err)
	assert.Equal(t, AlgorithmZstd, c.GetAlgorithm())

	c, err = f.CreateCompressor(&CompressionConfig{Algorithm: AlgorithmSnappy})
	require.NoError(t, err)
	assert.Equal(t, AlgorithmSnappy, c.GetAlgorithm())

	c, err = f.CreateCompressor(&CompressionConfig{Algorithm: AlgorithmDeflate, Level: LevelDefault})
	require.NoError(t, err)
	assert.Equal(t, AlgorithmDeflate, c.GetAlgorithm())

	_, err = f.CreateCompressor(&CompressionConfig{Algorithm: CompressionAlgorithm(99)})
	assert.Equal(t, ErrUnsupportedAlgorithm, err)
}

func TestCompressionManager(t *testing.T) {
	mgr := NewCompressionManager(&CompressionConfig{
		Algorithm: AlgorithmNone,
		MinSize:   0,
	})
	require.NotNil(t, mgr)

	data := []byte("hello world, testing compression manager")
	compressed, err := mgr.Compress(data)
	require.NoError(t, err)

	decompressed, err := mgr.Decompress(compressed, AlgorithmNone)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestCompressionManager_WithAlgorithm(t *testing.T) {
	mgr := NewCompressionManager(&CompressionConfig{
		Algorithm: AlgorithmGzip,
		Level:     LevelDefault,
		MinSize:   0,
	})
	require.NotNil(t, mgr)

	data := []byte("testing with specific algorithm")
	compressed, err := mgr.CompressWithAlgorithm(data, AlgorithmGzip)
	require.NoError(t, err)

	decompressed, err := mgr.Decompress(compressed, AlgorithmGzip)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestCompressionManager_SetDefaultAlgorithm(t *testing.T) {
	mgr := NewCompressionManager(nil)
	mgr.SetDefaultAlgorithm(AlgorithmGzip)
	stats := mgr.GetStats()
	assert.Equal(t, "gzip", stats["defaultAlgorithm"])
}

func TestCompressionManager_SetDefaultLevel(t *testing.T) {
	mgr := NewCompressionManager(nil)
	mgr.SetDefaultLevel(LevelBest)
	stats := mgr.GetStats()
	assert.Equal(t, "best", stats["defaultLevel"])
}

func TestCompressedBlock_EncodeDecode(t *testing.T) {
	block := &CompressedBlock{
		Algorithm:      AlgorithmGzip,
		OriginalSize:   1000,
		CompressedSize: 500,
		Data:           []byte("compressed data"),
		Checksum:       12345,
		Timestamp:      time.Now(),
	}

	encoded, err := block.Encode()
	require.NoError(t, err)
	assert.True(t, len(encoded) > 32)

	decoded, err := DecodeCompressedBlock(encoded)
	require.NoError(t, err)
	assert.Equal(t, block.Algorithm, decoded.Algorithm)
	assert.Equal(t, block.OriginalSize, decoded.OriginalSize)
	assert.Equal(t, block.CompressedSize, decoded.CompressedSize)
	assert.Equal(t, block.Checksum, decoded.Checksum)
	assert.Equal(t, block.Data, decoded.Data)
}

func TestDecodeCompressedBlock_InvalidData(t *testing.T) {
	_, err := DecodeCompressedBlock([]byte("short"))
	assert.Equal(t, ErrInvalidData, err)
}

func TestBlockCompressor(t *testing.T) {
	compressor := NewNoneCompressor()
	bc := NewBlockCompressor(compressor, 64)
	require.NotNil(t, bc)

	data := bytes.Repeat([]byte("abcdefgh"), 20)
	blocks, err := bc.CompressBlocks(data)
	require.NoError(t, err)
	assert.True(t, len(blocks) > 1)

	recovered, err := bc.DecompressBlocks(blocks)
	require.NoError(t, err)
	assert.Equal(t, data, recovered)
}

func TestCompressionRatio(t *testing.T) {
	r := &CompressionRatio{}
	r.Add(1000, 500)
	r.Add(2000, 800)

	assert.InDelta(t, 0.433, r.GetRatio(), 0.01)
	assert.InDelta(t, 0.567, r.GetSavings(), 0.01)

	stats := r.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(2), stats["samples"])

	r.Reset()
	assert.Equal(t, float64(0), r.GetRatio())
}

func TestGzipCompressor_Levels(t *testing.T) {
	data := bytes.Repeat([]byte("test data for compression level testing"), 10)

	levels := []CompressionLevel{LevelFastest, LevelFast, LevelDefault, LevelBetter, LevelBest}
	for _, level := range levels {
		c, err := NewGzipCompressor(level)
		require.NoError(t, err)
		compressed, err := c.Compress(data)
		require.NoError(t, err)
		decompressed, err := c.Decompress(compressed)
		require.NoError(t, err)
		assert.Equal(t, data, decompressed, "level %s roundtrip failed", level)
	}
}

func TestDeflateCompressor_Levels(t *testing.T) {
	data := bytes.Repeat([]byte("test data for compression level testing"), 10)

	levels := []CompressionLevel{LevelFastest, LevelFast, LevelDefault, LevelBetter, LevelBest}
	for _, level := range levels {
		c, err := NewDeflateCompressor(level)
		require.NoError(t, err)
		compressed, err := c.Compress(data)
		require.NoError(t, err)
		decompressed, err := c.Decompress(compressed)
		require.NoError(t, err)
		assert.Equal(t, data, decompressed, "level %s roundtrip failed", level)
	}
}

func TestLZ4Compressor_Levels(t *testing.T) {
	data := bytes.Repeat([]byte("test data for compression level testing"), 10)

	levels := []CompressionLevel{LevelFastest, LevelFast, LevelDefault, LevelBetter, LevelBest}
	for _, level := range levels {
		c, err := NewLZ4Compressor(level)
		require.NoError(t, err)
		compressed, err := c.Compress(data)
		require.NoError(t, err)
		decompressed, err := c.Decompress(compressed)
		require.NoError(t, err)
		assert.Equal(t, data, decompressed, "level %s roundtrip failed", level)
	}
}

func TestZstdCompressor_Levels(t *testing.T) {
	data := bytes.Repeat([]byte("test data for compression level testing"), 10)

	levels := []CompressionLevel{LevelFastest, LevelFast, LevelDefault, LevelBetter, LevelBest}
	for _, level := range levels {
		c, err := NewZstdCompressor(level)
		require.NoError(t, err)
		compressed, err := c.Compress(data)
		require.NoError(t, err)
		decompressed, err := c.Decompress(compressed)
		require.NoError(t, err)
		assert.Equal(t, data, decompressed, "level %s roundtrip failed", level)
	}
}

func TestGzipDecompressionFailed(t *testing.T) {
	c, err := NewGzipCompressor(LevelDefault)
	require.NoError(t, err)
	_, err = c.Decompress([]byte("not valid gzip data"))
	assert.Equal(t, ErrDecompressionFailed, err)
}

func TestDeflateDecompressionFailed(t *testing.T) {
	c, err := NewDeflateCompressor(LevelDefault)
	require.NoError(t, err)
	_, err = c.Decompress([]byte("not valid deflate data"))
	assert.Equal(t, ErrDecompressionFailed, err)
}
