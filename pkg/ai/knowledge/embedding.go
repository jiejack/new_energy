package knowledge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ErrEmptyDocument      = errors.New("document is empty")
	ErrInvalidChunkSize   = errors.New("invalid chunk size")
	ErrEmbeddingFailed    = errors.New("embedding generation failed")
	ErrProviderNotReady   = errors.New("embedding provider not ready")
	ErrUnsupportedModel   = errors.New("unsupported embedding model")
)

// EmbeddingProvider 嵌入提供者接口
type EmbeddingProvider interface {
	// Embed 生成单个文本的嵌入向量
	Embed(ctx context.Context, text string) ([]float32, error)
	// EmbedBatch 批量生成嵌入向量
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
	// Dimension 返回嵌入向量维度
	Dimension() int
	// ModelName 返回模型名称
	ModelName() string
}

// Document 表示一个文档
type Document struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Title       string                 `json:"title,omitempty"`
	Source      string                 `json:"source,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TextChunk 表示文本块
type TextChunk struct {
	ID         string                 `json:"id"`
	DocumentID string                 `json:"document_id"`
	Content    string                 `json:"content"`
	Position   int                    `json:"position"`
	StartPos   int                    `json:"start_pos"`
	EndPos     int                    `json:"end_pos"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Embedding  []float32              `json:"embedding,omitempty"`
}

// ChunkConfig 分块配置
type ChunkConfig struct {
	ChunkSize    int `json:"chunk_size"`     // 块大小（字符数）
	ChunkOverlap int `json:"chunk_overlap"`  // 重叠大小
	MinChunkSize int `json:"min_chunk_size"` // 最小块大小
	MaxChunkSize int `json:"max_chunk_size"` // 最大块大小
	SplitBySentence bool `json:"split_by_sentence"` // 按句子分割
	SplitByParagraph bool `json:"split_by_paragraph"` // 按段落分割
	SplitByToken bool `json:"split_by_token"` // 按token分割
}

// DefaultChunkConfig 默认分块配置
func DefaultChunkConfig() *ChunkConfig {
	return &ChunkConfig{
		ChunkSize:        512,
		ChunkOverlap:     50,
		MinChunkSize:     100,
		MaxChunkSize:     1024,
		SplitBySentence:  true,
		SplitByParagraph: false,
		SplitByToken:     false,
	}
}

// TextSplitter 文本分块器
type TextSplitter struct {
	config *ChunkConfig
	logger *zap.Logger
}

// NewTextSplitter 创建文本分块器
func NewTextSplitter(config *ChunkConfig, logger *zap.Logger) *TextSplitter {
	if config == nil {
		config = DefaultChunkConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &TextSplitter{
		config: config,
		logger: logger,
	}
}

// Split 分割文本为块
func (ts *TextSplitter) Split(text string) ([]*TextChunk, error) {
	if text == "" {
		return nil, ErrEmptyDocument
	}

	if ts.config.ChunkSize <= 0 {
		return nil, ErrInvalidChunkSize
	}

	var chunks []*TextChunk

	if ts.config.SplitByParagraph {
		chunks = ts.splitByParagraph(text)
	} else if ts.config.SplitBySentence {
		chunks = ts.splitBySentence(text)
	} else {
		chunks = ts.splitBySize(text)
	}

	// 设置块ID和位置
	for i, chunk := range chunks {
		chunk.ID = ts.generateChunkID(chunk.Content, i)
		chunk.Position = i
	}

	ts.logger.Debug("text split completed",
		zap.Int("total_chunks", len(chunks)),
		zap.Int("chunk_size", ts.config.ChunkSize),
	)

	return chunks, nil
}

// splitByParagraph 按段落分割
func (ts *TextSplitter) splitByParagraph(text string) []*TextChunk {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []*TextChunk
	currentPos := 0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// 如果段落太大，进一步分割
		if len(para) > ts.config.MaxChunkSize {
			subChunks := ts.splitBySentence(para)
			for _, sc := range subChunks {
				sc.StartPos = currentPos
				sc.EndPos = currentPos + len(sc.Content)
				chunks = append(chunks, sc)
				currentPos += len(sc.Content) + 2 // +2 for "\n\n"
			}
		} else {
			startPos := strings.Index(text[currentPos:], para) + currentPos
			chunks = append(chunks, &TextChunk{
				Content:  para,
				StartPos: startPos,
				EndPos:   startPos + len(para),
			})
			currentPos = startPos + len(para) + 2
		}
	}

	return chunks
}

// splitBySentence 按句子分割
func (ts *TextSplitter) splitBySentence(text string) []*TextChunk {
	// 中文和英文句子分隔符
	sentenceRegex := regexp.MustCompile(`[。！？.!?]+[\s]*`)
	
	var chunks []*TextChunk
	var currentChunk strings.Builder
	currentSize := 0
	startPos := 0
	lastEnd := 0

	// 找到所有句子边界
	indices := sentenceRegex.FindAllStringIndex(text, -1)
	indices = append(indices, []int{len(text), len(text)})

	for _, idx := range indices {
		sentence := text[lastEnd:idx[1]]
		sentenceSize := len(sentence)

		// 如果当前块+新句子不超过块大小，添加到当前块
		if currentSize+sentenceSize <= ts.config.ChunkSize {
			currentChunk.WriteString(sentence)
			currentSize += sentenceSize
		} else {
			// 保存当前块
			content := currentChunk.String()
			if currentSize >= ts.config.MinChunkSize {
				chunks = append(chunks, &TextChunk{
					Content:  content,
					StartPos: startPos,
					EndPos:   startPos + len(content),
				})
			}

			// 开始新块，考虑重叠
			if ts.config.ChunkOverlap > 0 && currentSize > ts.config.ChunkOverlap {
				overlapText := getLastNChars(currentChunk.String(), ts.config.ChunkOverlap)
				currentChunk.Reset()
				currentChunk.WriteString(overlapText)
				currentSize = len(overlapText)
				startPos = startPos + len(content) - ts.config.ChunkOverlap
			} else {
				currentChunk.Reset()
				currentSize = 0
				startPos = lastEnd
			}

			currentChunk.WriteString(sentence)
			currentSize += sentenceSize
		}

		lastEnd = idx[1]
	}

	// 添加最后一个块
	if currentSize >= ts.config.MinChunkSize {
		content := currentChunk.String()
		chunks = append(chunks, &TextChunk{
			Content:  content,
			StartPos: startPos,
			EndPos:   startPos + len(content),
		})
	}

	return chunks
}

// splitBySize 按大小分割
func (ts *TextSplitter) splitBySize(text string) []*TextChunk {
	var chunks []*TextChunk
	textLen := len(text)

	for i := 0; i < textLen; i += ts.config.ChunkSize - ts.config.ChunkOverlap {
		end := i + ts.config.ChunkSize
		if end > textLen {
			end = textLen
		}

		chunk := text[i:end]
		if len(chunk) >= ts.config.MinChunkSize {
			chunks = append(chunks, &TextChunk{
				Content:  chunk,
				StartPos: i,
				EndPos:   end,
			})
		}

		// 如果已经到末尾，退出
		if end >= textLen {
			break
		}
	}

	return chunks
}

// generateChunkID 生成块ID
func (ts *TextSplitter) generateChunkID(content string, position int) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", content, position)))
	return hex.EncodeToString(hash[:16])
}

// getLastNChars 获取最后N个字符
func getLastNChars(text string, n int) string {
	runes := []rune(text)
	if len(runes) <= n {
		return text
	}
	return string(runes[len(runes)-n:])
}

// EmbeddingCache 嵌入缓存
type EmbeddingCache struct {
	cache   map[string][]float32
	maxSize int
	mu      sync.RWMutex
	logger  *zap.Logger
}

// NewEmbeddingCache 创建嵌入缓存
func NewEmbeddingCache(maxSize int, logger *zap.Logger) *EmbeddingCache {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &EmbeddingCache{
		cache:   make(map[string][]float32),
		maxSize: maxSize,
		logger:  logger,
	}
}

// Get 获取缓存的嵌入向量
func (ec *EmbeddingCache) Get(text string) ([]float32, bool) {
	key := ec.hashText(text)
	
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	embedding, exists := ec.cache[key]
	return embedding, exists
}

// Set 设置缓存的嵌入向量
func (ec *EmbeddingCache) Set(text string, embedding []float32) {
	key := ec.hashText(text)
	
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	// 如果缓存已满，删除旧的条目
	if len(ec.cache) >= ec.maxSize {
		// 简单策略：删除第一个条目
		for k := range ec.cache {
			delete(ec.cache, k)
			break
		}
	}
	
	ec.cache[key] = embedding
}

// Clear 清空缓存
func (ec *EmbeddingCache) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.cache = make(map[string][]float32)
}

// Size 返回缓存大小
func (ec *EmbeddingCache) Size() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return len(ec.cache)
}

// hashText 计算文本哈希
func (ec *EmbeddingCache) hashText(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// OpenAIEmbeddingConfig OpenAI嵌入配置
type OpenAIEmbeddingConfig struct {
	APIKey      string        `json:"api_key"`
	Model       string        `json:"model"`       // text-embedding-ada-002, text-embedding-3-small, text-embedding-3-large
	BaseURL     string        `json:"base_url"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetry    int           `json:"max_retry"`
	BatchSize   int           `json:"batch_size"`
}

// OpenAIEmbeddingProvider OpenAI嵌入提供者
type OpenAIEmbeddingProvider struct {
	config     *OpenAIEmbeddingConfig
	logger     *zap.Logger
	dimension  int
	modelName  string
}

// NewOpenAIEmbeddingProvider 创建OpenAI嵌入提供者
func NewOpenAIEmbeddingProvider(config *OpenAIEmbeddingConfig, logger *zap.Logger) *OpenAIEmbeddingProvider {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	// 根据模型设置维度
	dimension := 1536 // 默认维度
	switch config.Model {
	case "text-embedding-ada-002":
		dimension = 1536
	case "text-embedding-3-small":
		dimension = 1536
	case "text-embedding-3-large":
		dimension = 3072
	}
	
	return &OpenAIEmbeddingProvider{
		config:    config,
		logger:    logger,
		dimension: dimension,
		modelName: config.Model,
	}
}

// Embed 生成单个文本的嵌入向量
func (p *OpenAIEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, ErrEmptyDocument
	}

	embeddings, err := p.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, ErrEmbeddingFailed
	}

	return embeddings[0], nil
}

// EmbedBatch 批量生成嵌入向量
func (p *OpenAIEmbeddingProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, ErrEmptyDocument
	}

	p.logger.Debug("generating embeddings",
		zap.String("model", p.modelName),
		zap.Int("batch_size", len(texts)),
	)

	// 实际实现中应该调用OpenAI API
	// 这里返回模拟的嵌入向量
	embeddings := make([][]float32, len(texts))
	for i := range embeddings {
		embeddings[i] = make([]float32, p.dimension)
		// 模拟：生成随机向量（实际应该调用API）
		for j := range embeddings[i] {
			embeddings[i][j] = 0.1 // 简化示例
		}
	}

	return embeddings, nil
}

// Dimension 返回嵌入向量维度
func (p *OpenAIEmbeddingProvider) Dimension() int {
	return p.dimension
}

// ModelName 返回模型名称
func (p *OpenAIEmbeddingProvider) ModelName() string {
	return p.modelName
}

// LocalEmbeddingConfig 本地嵌入配置
type LocalEmbeddingConfig struct {
	ModelPath   string `json:"model_path"`
	ModelType   string `json:"model_type"` // sentence-transformers, etc.
	Dimension   int    `json:"dimension"`
	MaxSeqLen   int    `json:"max_seq_len"`
	BatchSize   int    `json:"batch_size"`
}

// LocalEmbeddingProvider 本地嵌入提供者
type LocalEmbeddingProvider struct {
	config    *LocalEmbeddingConfig
	logger    *zap.Logger
	ready     bool
	mu        sync.RWMutex
}

// NewLocalEmbeddingProvider 创建本地嵌入提供者
func NewLocalEmbeddingProvider(config *LocalEmbeddingConfig, logger *zap.Logger) *LocalEmbeddingProvider {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LocalEmbeddingProvider{
		config: config,
		logger: logger,
		ready:  false,
	}
}

// Initialize 初始化模型
func (p *LocalEmbeddingProvider) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logger.Info("initializing local embedding model",
		zap.String("model_path", p.config.ModelPath),
		zap.String("model_type", p.config.ModelType),
	)

	// 实际实现中应该加载模型
	p.ready = true

	return nil
}

// Embed 生成单个文本的嵌入向量
func (p *LocalEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if !p.IsReady() {
		return nil, ErrProviderNotReady
	}

	embeddings, err := p.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, ErrEmbeddingFailed
	}

	return embeddings[0], nil
}

// EmbedBatch 批量生成嵌入向量
func (p *LocalEmbeddingProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if !p.IsReady() {
		return nil, ErrProviderNotReady
	}

	if len(texts) == 0 {
		return nil, ErrEmptyDocument
	}

	p.logger.Debug("generating local embeddings",
		zap.Int("batch_size", len(texts)),
	)

	// 实际实现中应该调用本地模型
	embeddings := make([][]float32, len(texts))
	for i := range embeddings {
		embeddings[i] = make([]float32, p.config.Dimension)
		for j := range embeddings[i] {
			embeddings[i][j] = 0.1
		}
	}

	return embeddings, nil
}

// Dimension 返回嵌入向量维度
func (p *LocalEmbeddingProvider) Dimension() int {
	return p.config.Dimension
}

// ModelName 返回模型名称
func (p *LocalEmbeddingProvider) ModelName() string {
	return p.config.ModelType
}

// IsReady 检查是否就绪
func (p *LocalEmbeddingProvider) IsReady() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ready
}

// DocumentEmbedder 文档嵌入器
type DocumentEmbedder struct {
	provider EmbeddingProvider
	splitter *TextSplitter
	cache    *EmbeddingCache
	logger   *zap.Logger
}

// DocumentEmbedderConfig 文档嵌入器配置
type DocumentEmbedderConfig struct {
	Provider      EmbeddingProvider
	ChunkConfig   *ChunkConfig
	CacheSize     int
}

// NewDocumentEmbedder 创建文档嵌入器
func NewDocumentEmbedder(config *DocumentEmbedderConfig, logger *zap.Logger) *DocumentEmbedder {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config.ChunkConfig == nil {
		config.ChunkConfig = DefaultChunkConfig()
	}

	if config.CacheSize <= 0 {
		config.CacheSize = 10000
	}

	return &DocumentEmbedder{
		provider: config.Provider,
		splitter: NewTextSplitter(config.ChunkConfig, logger),
		cache:    NewEmbeddingCache(config.CacheSize, logger),
		logger:   logger,
	}
}

// EmbedDocument 嵌入文档
func (de *DocumentEmbedder) EmbedDocument(ctx context.Context, doc *Document) ([]*TextChunk, error) {
	if doc == nil || doc.Content == "" {
		return nil, ErrEmptyDocument
	}

	// 分割文本
	chunks, err := de.splitter.Split(doc.Content)
	if err != nil {
		return nil, err
	}

	// 设置文档ID
	for _, chunk := range chunks {
		chunk.DocumentID = doc.ID
		if chunk.Metadata == nil {
			chunk.Metadata = make(map[string]interface{})
		}
		// 添加文档元数据
		for k, v := range doc.Metadata {
			chunk.Metadata[k] = v
		}
		chunk.Metadata["title"] = doc.Title
		chunk.Metadata["source"] = doc.Source
	}

	// 批量生成嵌入向量
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	embeddings, err := de.EmbedBatch(ctx, texts)
	if err != nil {
		return nil, err
	}

	// 设置嵌入向量
	for i, chunk := range chunks {
		if i < len(embeddings) {
			chunk.Embedding = embeddings[i]
		}
	}

	de.logger.Info("document embedded",
		zap.String("doc_id", doc.ID),
		zap.Int("chunks", len(chunks)),
	)

	return chunks, nil
}

// Embed 生成单个文本的嵌入向量
func (de *DocumentEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// 检查缓存
	if embedding, exists := de.cache.Get(text); exists {
		return embedding, nil
	}

	// 生成嵌入向量
	embedding, err := de.provider.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	de.cache.Set(text, embedding)

	return embedding, nil
}

// EmbedBatch 批量生成嵌入向量
func (de *DocumentEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	// 分离已缓存和未缓存的文本
	cachedEmbeddings := make([][]float32, len(texts))
	uncachedIndices := make([]int, 0)
	uncachedTexts := make([]string, 0)

	for i, text := range texts {
		if embedding, exists := de.cache.Get(text); exists {
			cachedEmbeddings[i] = embedding
		} else {
			uncachedIndices = append(uncachedIndices, i)
			uncachedTexts = append(uncachedTexts, text)
		}
	}

	// 生成未缓存的嵌入向量
	if len(uncachedTexts) > 0 {
		embeddings, err := de.provider.EmbedBatch(ctx, uncachedTexts)
		if err != nil {
			return nil, err
		}

		// 设置结果并缓存
		for i, idx := range uncachedIndices {
			if i < len(embeddings) {
				cachedEmbeddings[idx] = embeddings[i]
				de.cache.Set(uncachedTexts[i], embeddings[i])
			}
		}
	}

	return cachedEmbeddings, nil
}

// GetCacheStats 获取缓存统计
func (de *DocumentEmbedder) GetCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"cache_size": de.cache.Size(),
		"dimension":  de.provider.Dimension(),
		"model":      de.provider.ModelName(),
	}
}

// ClearCache 清空缓存
func (de *DocumentEmbedder) ClearCache() {
	de.cache.Clear()
	de.logger.Info("embedding cache cleared")
}

// EmbeddingStats 嵌入统计
type EmbeddingStats struct {
	TotalDocuments   int64 `json:"total_documents"`
	TotalChunks      int64 `json:"total_chunks"`
	TotalEmbeddings  int64 `json:"total_embeddings"`
	CacheHits        int64 `json:"cache_hits"`
	CacheMisses      int64 `json:"cache_misses"`
	AverageChunkSize int64 `json:"average_chunk_size"`
}

// EmbeddingService 嵌入服务
type EmbeddingService struct {
	embedder *DocumentEmbedder
	stats    *EmbeddingStats
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewEmbeddingService 创建嵌入服务
func NewEmbeddingService(embedder *DocumentEmbedder, logger *zap.Logger) *EmbeddingService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &EmbeddingService{
		embedder: embedder,
		stats:    &EmbeddingStats{},
		logger:   logger,
	}
}

// ProcessDocument 处理文档
func (es *EmbeddingService) ProcessDocument(ctx context.Context, doc *Document) ([]*TextChunk, error) {
	chunks, err := es.embedder.EmbedDocument(ctx, doc)
	if err != nil {
		return nil, err
	}

	es.mu.Lock()
	es.stats.TotalDocuments++
	es.stats.TotalChunks += int64(len(chunks))
	es.stats.TotalEmbeddings += int64(len(chunks))
	es.mu.Unlock()

	return chunks, nil
}

// GetStats 获取统计信息
func (es *EmbeddingService) GetStats() *EmbeddingStats {
	es.mu.RLock()
	defer es.mu.RUnlock()
	
	stats := *es.stats
	return &stats
}

// BatchEmbeddingRequest 批量嵌入请求
type BatchEmbeddingRequest struct {
	Documents []*Document `json:"documents"`
	BatchSize int         `json:"batch_size"`
}

// BatchEmbeddingResult 批量嵌入结果
type BatchEmbeddingResult struct {
	DocumentID string       `json:"document_id"`
	Chunks     []*TextChunk `json:"chunks"`
	Error      string       `json:"error,omitempty"`
}

// BatchEmbedDocuments 批量嵌入文档
func (es *EmbeddingService) BatchEmbedDocuments(ctx context.Context, req *BatchEmbeddingRequest) []*BatchEmbeddingResult {
	results := make([]*BatchEmbeddingResult, len(req.Documents))

	for i, doc := range req.Documents {
		chunks, err := es.ProcessDocument(ctx, doc)
		result := &BatchEmbeddingResult{
			DocumentID: doc.ID,
		}

		if err != nil {
			result.Error = err.Error()
		} else {
			result.Chunks = chunks
		}

		results[i] = result
	}

	return results
}

// SimilarityCalculator 相似度计算器
type SimilarityCalculator struct {
	calc *DistanceCalculator
}

// NewSimilarityCalculator 创建相似度计算器
func NewSimilarityCalculator() *SimilarityCalculator {
	return &SimilarityCalculator{
		calc: NewDistanceCalculator(),
	}
}

// CosineSimilarity 计算余弦相似度
func (sc *SimilarityCalculator) CosineSimilarity(a, b []float32) (float32, error) {
	return sc.calc.CosineSimilarity(a, b)
}

// FindMostSimilar 找到最相似的向量
func (sc *SimilarityCalculator) FindMostSimilar(query []float32, vectors [][]float32, topK int) ([]int, []float32, error) {
	if len(vectors) == 0 {
		return nil, nil, nil
	}

	type score struct {
		index int
		value float32
	}

	scores := make([]score, len(vectors))
	for i, v := range vectors {
		sim, err := sc.CosineSimilarity(query, v)
		if err != nil {
			continue
		}
		scores[i] = score{index: i, value: sim}
	}

	// 排序（降序）
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].value > scores[i].value {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// 返回topK
	if topK > len(scores) {
		topK = len(scores)
	}

	indices := make([]int, topK)
	values := make([]float32, topK)
	for i := 0; i < topK; i++ {
		indices[i] = scores[i].index
		values[i] = scores[i].value
	}

	return indices, values, nil
}
