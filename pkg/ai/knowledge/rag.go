package knowledge

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ErrNoResults          = errors.New("no results found")
	ErrInvalidQuery       = errors.New("invalid query")
	ErrRetrieverNotReady  = errors.New("retriever not ready")
	ErrRerankerNotReady   = errors.New("reranker not ready")
)

// RAGConfig RAG配置
type RAGConfig struct {
	TopK              int     `json:"top_k"`               // 初始检索数量
	FinalTopK         int     `json:"final_top_k"`         // 最终返回数量
	ScoreThreshold    float32 `json:"score_threshold"`     // 分数阈值
	UseReranker       bool    `json:"use_reranker"`        // 是否使用重排序
	UseHybridSearch   bool    `json:"use_hybrid_search"`   // 是否使用混合检索
	VectorWeight      float32 `json:"vector_weight"`       // 向量检索权重
	KeywordWeight     float32 `json:"keyword_weight"`      // 关键词检索权重
	ContextWindowSize int     `json:"context_window_size"` // 上下文窗口大小
	MaxContextLength  int     `json:"max_context_length"`  // 最大上下文长度
}

// DefaultRAGConfig 默认RAG配置
func DefaultRAGConfig() *RAGConfig {
	return &RAGConfig{
		TopK:              20,
		FinalTopK:         5,
		ScoreThreshold:    0.5,
		UseReranker:       true,
		UseHybridSearch:   true,
		VectorWeight:      0.7,
		KeywordWeight:     0.3,
		ContextWindowSize: 3,
		MaxContextLength:  4096,
	}
}

// RetrievalResult 检索结果
type RetrievalResult struct {
	Chunk       *TextChunk `json:"chunk"`
	Score       float32    `json:"score"`
	VectorScore float32    `json:"vector_score,omitempty"`
	KeywordScore float32   `json:"keyword_score,omitempty"`
	Rank        int        `json:"rank"`
	Source      string     `json:"source"` // vector, keyword, hybrid
}

// RAGContext RAG上下文
type RAGContext struct {
	Query       string             `json:"query"`
	Results     []*RetrievalResult `json:"results"`
	Context     string             `json:"context"`
	TotalTokens int                `json:"total_tokens"`
	Duration    time.Duration      `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// KeywordSearcher 关键词搜索器接口
type KeywordSearcher interface {
	Search(ctx context.Context, query string, topK int) ([]*TextChunk, error)
	Index(ctx context.Context, chunks []*TextChunk) error
	Delete(ctx context.Context, ids []string) error
}

// Reranker 重排序器接口
type Reranker interface {
	Rerank(ctx context.Context, query string, results []*RetrievalResult) ([]*RetrievalResult, error)
}

// SimpleKeywordSearcher 简单关键词搜索器
type SimpleKeywordSearcher struct {
	chunks map[string]*TextChunk
	index  map[string][]string // keyword -> chunk IDs
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewSimpleKeywordSearcher 创建简单关键词搜索器
func NewSimpleKeywordSearcher(logger *zap.Logger) *SimpleKeywordSearcher {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SimpleKeywordSearcher{
		chunks: make(map[string]*TextChunk),
		index:  make(map[string][]string),
		logger: logger,
	}
}

// Search 关键词搜索
func (s *SimpleKeywordSearcher) Search(ctx context.Context, query string, topK int) ([]*TextChunk, error) {
	if query == "" {
		return nil, ErrInvalidQuery
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 分词
	keywords := tokenize(query)
	
	// 统计每个chunk的匹配次数
	matchCount := make(map[string]int)
	for _, keyword := range keywords {
		if chunkIDs, exists := s.index[keyword]; exists {
			for _, id := range chunkIDs {
				matchCount[id]++
			}
		}
	}

	// 按匹配次数排序
	type scored struct {
		id    string
		count int
	}
	var scoredChunks []scored
	for id, count := range matchCount {
		scoredChunks = append(scoredChunks, scored{id: id, count: count})
	}

	sort.Slice(scoredChunks, func(i, j int) bool {
		return scoredChunks[i].count > scoredChunks[j].count
	})

	// 返回topK结果
	results := make([]*TextChunk, 0, topK)
	for i := 0; i < len(scoredChunks) && i < topK; i++ {
		if chunk, exists := s.chunks[scoredChunks[i].id]; exists {
			results = append(results, chunk)
		}
	}

	return results, nil
}

// Index 索引文本块
func (s *SimpleKeywordSearcher) Index(ctx context.Context, chunks []*TextChunk) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, chunk := range chunks {
		// 存储chunk
		s.chunks[chunk.ID] = chunk

		// 建立倒排索引
		keywords := tokenize(chunk.Content)
		for _, keyword := range keywords {
			s.index[keyword] = append(s.index[keyword], chunk.ID)
		}
	}

	s.logger.Debug("indexed chunks", zap.Int("count", len(chunks)))
	return nil
}

// Delete 删除文本块
func (s *SimpleKeywordSearcher) Delete(ctx context.Context, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range ids {
		if chunk, exists := s.chunks[id]; exists {
			// 从倒排索引中删除
			keywords := tokenize(chunk.Content)
			for _, keyword := range keywords {
				ids := s.index[keyword]
				newIDs := make([]string, 0)
				for _, chunkID := range ids {
					if chunkID != id {
						newIDs = append(newIDs, chunkID)
					}
				}
				if len(newIDs) > 0 {
					s.index[keyword] = newIDs
				} else {
					delete(s.index, keyword)
				}
			}
			delete(s.chunks, id)
		}
	}

	return nil
}

// tokenize 分词
func tokenize(text string) []string {
	// 简单分词：按空格和标点符号分割
	text = strings.ToLower(text)
	
	// 替换标点符号为空格
	replacer := strings.NewReplacer(
		",", " ",
		".", " ",
		"!", " ",
		"?", " ",
		";", " ",
		":", " ",
		"\"", " ",
		"'", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"，", " ",
		"。", " ",
		"！", " ",
		"？", " ",
		"；", " ",
		"：", " ",
		"（", " ",
		"）", " ",
		"【", " ",
		"】", " ",
	)
	text = replacer.Replace(text)

	// 分割
	words := strings.Fields(text)
	
	// 去重
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, word := range words {
		if len(word) >= 2 && !seen[word] { // 忽略单字符
			seen[word] = true
			result = append(result, word)
		}
	}

	return result
}

// SimpleReranker 简单重排序器
type SimpleReranker struct {
	logger *zap.Logger
}

// NewSimpleReranker 创建简单重排序器
func NewSimpleReranker(logger *zap.Logger) *SimpleReranker {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SimpleReranker{logger: logger}
}

// Rerank 重排序
func (r *SimpleReranker) Rerank(ctx context.Context, query string, results []*RetrievalResult) ([]*RetrievalResult, error) {
	if len(results) == 0 {
		return results, nil
	}

	// 计算查询词在chunk中的出现频率
	queryTerms := tokenize(query)
	
	for _, result := range results {
		chunkTerms := tokenize(result.Chunk.Content)
		
		// 计算BM25风格的分数
		termFreq := 0
		for _, qt := range queryTerms {
			for _, ct := range chunkTerms {
				if qt == ct {
					termFreq++
				}
			}
		}
		
		// 结合原始分数和词频分数
		bm25Score := float32(termFreq) / float32(len(chunkTerms)+1)
		result.Score = result.Score*0.7 + bm25Score*0.3
	}

	// 重新排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 更新排名
	for i := range results {
		results[i].Rank = i + 1
	}

	return results, nil
}

// ContextWindowManager 上下文窗口管理器
type ContextWindowManager struct {
	maxLength int
	logger    *zap.Logger
}

// NewContextWindowManager 创建上下文窗口管理器
func NewContextWindowManager(maxLength int, logger *zap.Logger) *ContextWindowManager {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ContextWindowManager{
		maxLength: maxLength,
		logger:    logger,
	}
}

// BuildContext 构建上下文
func (cwm *ContextWindowManager) BuildContext(results []*RetrievalResult) string {
	var context strings.Builder
	currentLength := 0

	for i, result := range results {
		chunkText := result.Chunk.Content
		
		// 检查是否超过最大长度
		if currentLength + len(chunkText) > cwm.maxLength {
			// 尝试截断
			remaining := cwm.maxLength - currentLength
			if remaining > 100 { // 至少保留100个字符
				chunkText = chunkText[:remaining] + "..."
			} else {
				break
			}
		}

		// 添加分隔符
		if i > 0 {
			context.WriteString("\n\n---\n\n")
			currentLength += 7
		}

		// 添加引用标记
		context.WriteString(fmt.Sprintf("[文档%d]\n", i+1))
		currentLength += len(fmt.Sprintf("[文档%d]\n", i+1))

		context.WriteString(chunkText)
		currentLength += len(chunkText)
	}

	return context.String()
}

// Citation 引用
type Citation struct {
	ChunkID    string `json:"chunk_id"`
	DocumentID string `json:"document_id"`
	Title      string `json:"title,omitempty"`
	Source     string `json:"source,omitempty"`
	Position   int    `json:"position"`
	Content    string `json:"content"`
	Score      float32 `json:"score"`
}

// CitationTracker 引用追踪器
type CitationTracker struct {
	logger *zap.Logger
}

// NewCitationTracker 创建引用追踪器
func NewCitationTracker(logger *zap.Logger) *CitationTracker {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CitationTracker{logger: logger}
}

// ExtractCitations 提取引用
func (ct *CitationTracker) ExtractCitations(results []*RetrievalResult) []*Citation {
	citations := make([]*Citation, len(results))
	
	for i, result := range results {
		citation := &Citation{
			ChunkID:    result.Chunk.ID,
			DocumentID: result.Chunk.DocumentID,
			Position:   i + 1,
			Content:    result.Chunk.Content,
			Score:      result.Score,
		}

		// 从元数据中提取标题和来源
		if result.Chunk.Metadata != nil {
			if title, ok := result.Chunk.Metadata["title"].(string); ok {
				citation.Title = title
			}
			if source, ok := result.Chunk.Metadata["source"].(string); ok {
				citation.Source = source
			}
		}

		citations[i] = citation
	}

	return citations
}

// RAGRetriever RAG检索器
type RAGRetriever struct {
	vectorDB         VectorDB
	embedder         *DocumentEmbedder
	keywordSearcher  KeywordSearcher
	reranker         Reranker
	config           *RAGConfig
	contextManager   *ContextWindowManager
	citationTracker  *CitationTracker
	logger           *zap.Logger
	mu               sync.RWMutex
}

// NewRAGRetriever 创建RAG检索器
func NewRAGRetriever(
	vectorDB VectorDB,
	embedder *DocumentEmbedder,
	config *RAGConfig,
	logger *zap.Logger,
) *RAGRetriever {
	if config == nil {
		config = DefaultRAGConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &RAGRetriever{
		vectorDB:        vectorDB,
		embedder:        embedder,
		keywordSearcher: NewSimpleKeywordSearcher(logger),
		reranker:        NewSimpleReranker(logger),
		config:          config,
		contextManager:  NewContextWindowManager(config.MaxContextLength, logger),
		citationTracker: NewCitationTracker(logger),
		logger:          logger,
	}
}

// Retrieve 检索
func (r *RAGRetriever) Retrieve(ctx context.Context, query string, indexName string) (*RAGContext, error) {
	startTime := time.Now()

	if query == "" {
		return nil, ErrInvalidQuery
	}

	if !r.vectorDB.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	// 生成查询向量
	queryVector, err := r.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	var results []*RetrievalResult

	// 混合检索
	if r.config.UseHybridSearch {
		results, err = r.hybridSearch(ctx, query, queryVector, indexName)
	} else {
		results, err = r.vectorSearch(ctx, queryVector, indexName)
	}

	if err != nil {
		return nil, err
	}

	// 过滤低分结果
	results = r.filterResults(results)

	if len(results) == 0 {
		return nil, ErrNoResults
	}

	// 重排序
	if r.config.UseReranker && r.reranker != nil {
		results, err = r.reranker.Rerank(ctx, query, results)
		if err != nil {
			r.logger.Warn("reranking failed", zap.Error(err))
		}
	}

	// 限制最终结果数量
	if len(results) > r.config.FinalTopK {
		results = results[:r.config.FinalTopK]
	}

	// 构建上下文
	context := r.contextManager.BuildContext(results)

	// 构建RAG上下文
	ragContext := &RAGContext{
		Query:       query,
		Results:     results,
		Context:     context,
		TotalTokens: len(context), // 简化的token计数
		Duration:    time.Since(startTime),
		Metadata: map[string]interface{}{
			"index_name":      indexName,
			"use_hybrid":      r.config.UseHybridSearch,
			"use_reranker":    r.config.UseReranker,
			"initial_top_k":   r.config.TopK,
			"final_top_k":     r.config.FinalTopK,
		},
	}

	r.logger.Info("retrieval completed",
		zap.String("query", query),
		zap.Int("results", len(results)),
		zap.Duration("duration", ragContext.Duration),
	)

	return ragContext, nil
}

// vectorSearch 向量检索
func (r *RAGRetriever) vectorSearch(ctx context.Context, queryVector []float32, indexName string) ([]*RetrievalResult, error) {
	searchParams := &SearchParams{
		TopK: r.config.TopK,
	}

	searchResults, err := r.vectorDB.Search(ctx, indexName, queryVector, searchParams)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	results := make([]*RetrievalResult, len(searchResults))
	for i, sr := range searchResults {
		results[i] = &RetrievalResult{
			Chunk:       sr.Vector.Metadata["chunk"].(*TextChunk),
			Score:       sr.Score,
			VectorScore: sr.Score,
			Rank:        i + 1,
			Source:      "vector",
		}
	}

	return results, nil
}

// hybridSearch 混合检索
func (r *RAGRetriever) hybridSearch(ctx context.Context, query string, queryVector []float32, indexName string) ([]*RetrievalResult, error) {
	// 并行执行向量检索和关键词检索
	var vectorResults []*RetrievalResult
	var keywordResults []*TextChunk
	var vectorErr, keywordErr error
	var wg sync.WaitGroup

	wg.Add(2)

	// 向量检索
	go func() {
		defer wg.Done()
		vectorResults, vectorErr = r.vectorSearch(ctx, queryVector, indexName)
	}()

	// 关键词检索
	go func() {
		defer wg.Done()
		keywordResults, keywordErr = r.keywordSearcher.Search(ctx, query, r.config.TopK)
	}()

	wg.Wait()

	// 处理错误
	if vectorErr != nil {
		r.logger.Warn("vector search failed", zap.Error(vectorErr))
	}
	if keywordErr != nil {
		r.logger.Warn("keyword search failed", zap.Error(keywordErr))
	}

	// 合并结果
	return r.mergeResults(vectorResults, keywordResults), nil
}

// mergeResults 合并检索结果
func (r *RAGRetriever) mergeResults(vectorResults []*RetrievalResult, keywordResults []*TextChunk) []*RetrievalResult {
	// 使用map去重
	resultMap := make(map[string]*RetrievalResult)

	// 添加向量检索结果
	for _, result := range vectorResults {
		if result.Chunk != nil {
			resultMap[result.Chunk.ID] = &RetrievalResult{
				Chunk:       result.Chunk,
				VectorScore: result.Score,
				Score:       result.Score * r.config.VectorWeight,
				Source:      "vector",
			}
		}
	}

	// 添加关键词检索结果
	for i, chunk := range keywordResults {
		if existing, exists := resultMap[chunk.ID]; exists {
			// 合并分数
			keywordScore := float32(r.config.TopK-i) / float32(r.config.TopK)
			existing.KeywordScore = keywordScore
			existing.Score = existing.VectorScore*r.config.VectorWeight + keywordScore*r.config.KeywordWeight
			existing.Source = "hybrid"
		} else {
			// 新结果
			keywordScore := float32(r.config.TopK-i) / float32(r.config.TopK)
			resultMap[chunk.ID] = &RetrievalResult{
				Chunk:        chunk,
				KeywordScore: keywordScore,
				Score:        keywordScore * r.config.KeywordWeight,
				Source:       "keyword",
			}
		}
	}

	// 转换为切片并排序
	results := make([]*RetrievalResult, 0, len(resultMap))
	for _, result := range resultMap {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 设置排名
	for i := range results {
		results[i].Rank = i + 1
	}

	return results
}

// filterResults 过滤结果
func (r *RAGRetriever) filterResults(results []*RetrievalResult) []*RetrievalResult {
	filtered := make([]*RetrievalResult, 0)
	
	for _, result := range results {
		if result.Score >= r.config.ScoreThreshold {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// IndexChunks 索引文本块
func (r *RAGRetriever) IndexChunks(ctx context.Context, indexName string, chunks []*TextChunk) error {
	// 准备向量
	vectors := make([]*Vector, len(chunks))
	for i, chunk := range chunks {
		vectors[i] = &Vector{
			ID:   chunk.ID,
			Data: chunk.Embedding,
			Metadata: map[string]interface{}{
				"chunk":      chunk,
				"document_id": chunk.DocumentID,
			},
		}
	}

	// 插入向量数据库
	if err := r.vectorDB.Upsert(ctx, indexName, vectors); err != nil {
		return fmt.Errorf("failed to index vectors: %w", err)
	}

	// 索引到关键词搜索器
	if err := r.keywordSearcher.Index(ctx, chunks); err != nil {
		r.logger.Warn("failed to index keywords", zap.Error(err))
	}

	r.logger.Info("chunks indexed",
		zap.String("index", indexName),
		zap.Int("count", len(chunks)),
	)

	return nil
}

// DeleteChunks 删除文本块
func (r *RAGRetriever) DeleteChunks(ctx context.Context, indexName string, chunkIDs []string) error {
	// 从向量数据库删除
	if err := r.vectorDB.Delete(ctx, indexName, chunkIDs); err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	// 从关键词搜索器删除
	if err := r.keywordSearcher.Delete(ctx, chunkIDs); err != nil {
		r.logger.Warn("failed to delete from keyword searcher", zap.Error(err))
	}

	return nil
}

// GetCitations 获取引用
func (r *RAGRetriever) GetCitations(results []*RetrievalResult) []*Citation {
	return r.citationTracker.ExtractCitations(results)
}

// RAGStats RAG统计
type RAGStats struct {
	TotalQueries      int64         `json:"total_queries"`
	TotalResults      int64         `json:"total_results"`
	AverageDuration   time.Duration `json:"average_duration"`
	AverageResults    float64       `json:"average_results"`
	CacheHits         int64         `json:"cache_hits"`
	CacheMisses       int64         `json:"cache_misses"`
	VectorSearchTime  time.Duration `json:"vector_search_time"`
	KeywordSearchTime time.Duration `json:"keyword_search_time"`
	RerankTime        time.Duration `json:"rerank_time"`
}

// RAGService RAG服务
type RAGService struct {
	retriever *RAGRetriever
	stats     *RAGStats
	mu        sync.RWMutex
	logger    *zap.Logger
}

// NewRAGService 创建RAG服务
func NewRAGService(retriever *RAGRetriever, logger *zap.Logger) *RAGService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RAGService{
		retriever: retriever,
		stats:     &RAGStats{},
		logger:    logger,
	}
}

// Query 查询
func (rs *RAGService) Query(ctx context.Context, query string, indexName string) (*RAGContext, error) {
	ragContext, err := rs.retriever.Retrieve(ctx, query, indexName)
	if err != nil {
		return nil, err
	}

	// 更新统计
	rs.mu.Lock()
	rs.stats.TotalQueries++
	rs.stats.TotalResults += int64(len(ragContext.Results))
	rs.stats.AverageResults = float64(rs.stats.TotalResults) / float64(rs.stats.TotalQueries)
	rs.mu.Unlock()

	return ragContext, nil
}

// GetStats 获取统计
func (rs *RAGService) GetStats() *RAGStats {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	
	stats := *rs.stats
	return &stats
}

// HybridSearchResult 混合搜索结果
type HybridSearchResult struct {
	VectorResults   []*RetrievalResult `json:"vector_results"`
	KeywordResults  []*TextChunk       `json:"keyword_results"`
	MergedResults   []*RetrievalResult `json:"merged_results"`
	VectorDuration  time.Duration      `json:"vector_duration"`
	KeywordDuration time.Duration      `json:"keyword_duration"`
}

// HybridSearch 混合搜索
func (rs *RAGService) HybridSearch(ctx context.Context, query string, indexName string) (*HybridSearchResult, error) {
	startTime := time.Now()
	
	// 生成查询向量
	queryVector, err := rs.retriever.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	result := &HybridSearchResult{}

	// 向量检索
	vectorStart := time.Now()
	result.VectorResults, _ = rs.retriever.vectorSearch(ctx, queryVector, indexName)
	result.VectorDuration = time.Since(vectorStart)

	// 关键词检索
	keywordStart := time.Now()
	result.KeywordResults, _ = rs.retriever.keywordSearcher.Search(ctx, query, rs.retriever.config.TopK)
	result.KeywordDuration = time.Since(keywordStart)

	// 合并结果
	result.MergedResults = rs.retriever.mergeResults(result.VectorResults, result.KeywordResults)

	rs.logger.Info("hybrid search completed",
		zap.Duration("total_duration", time.Since(startTime)),
		zap.Int("vector_results", len(result.VectorResults)),
		zap.Int("keyword_results", len(result.KeywordResults)),
		zap.Int("merged_results", len(result.MergedResults)),
	)

	return result, nil
}

// MultiQueryRetriever 多查询检索器
type MultiQueryRetriever struct {
	retriever *RAGRetriever
	logger    *zap.Logger
}

// NewMultiQueryRetriever 创建多查询检索器
func NewMultiQueryRetriever(retriever *RAGRetriever, logger *zap.Logger) *MultiQueryRetriever {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MultiQueryRetriever{
		retriever: retriever,
		logger:    logger,
	}
}

// Retrieve 多查询检索
func (mqr *MultiQueryRetriever) Retrieve(ctx context.Context, queries []string, indexName string) (*RAGContext, error) {
	if len(queries) == 0 {
		return nil, ErrInvalidQuery
	}

	// 并行检索
	var wg sync.WaitGroup
	resultsChan := make(chan []*RetrievalResult, len(queries))
	errorsChan := make(chan error, len(queries))

	for _, query := range queries {
		wg.Add(1)
		go func(q string) {
			defer wg.Done()
			ragContext, err := mqr.retriever.Retrieve(ctx, q, indexName)
			if err != nil {
				errorsChan <- err
				return
			}
			resultsChan <- ragContext.Results
		}(query)
	}

	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	// 检查错误
	for err := range errorsChan {
		if err != nil {
			mqr.logger.Warn("query failed", zap.Error(err))
		}
	}

	// 合并所有结果
	allResults := make(map[string]*RetrievalResult)
	for results := range resultsChan {
		for _, result := range results {
			if existing, exists := allResults[result.Chunk.ID]; exists {
				// 取最高分
				if result.Score > existing.Score {
					allResults[result.Chunk.ID] = result
				}
			} else {
				allResults[result.Chunk.ID] = result
			}
		}
	}

	// 转换为切片并排序
	mergedResults := make([]*RetrievalResult, 0, len(allResults))
	for _, result := range allResults {
		mergedResults = append(mergedResults, result)
	}

	sort.Slice(mergedResults, func(i, j int) bool {
		return mergedResults[i].Score > mergedResults[j].Score
	})

	// 限制结果数量
	if len(mergedResults) > mqr.retriever.config.FinalTopK {
		mergedResults = mergedResults[:mqr.retriever.config.FinalTopK]
	}

	// 构建上下文
	context := mqr.retriever.contextManager.BuildContext(mergedResults)

	return &RAGContext{
		Query:    queries[0], // 使用第一个查询作为主查询
		Results:  mergedResults,
		Context:  context,
		Metadata: map[string]interface{}{
			"query_count": len(queries),
		},
	}, nil
}
