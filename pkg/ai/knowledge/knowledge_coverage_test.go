package knowledge

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type coverageEmbeddingProvider struct {
	dimension int
	modelName string
}

func (p *coverageEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, p.dimension)
	for i := range vec {
		vec[i] = 0.1
	}
	return vec, nil
}

func (p *coverageEmbeddingProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range result {
		vec := make([]float32, p.dimension)
		for j := range vec {
			vec[j] = 0.1
		}
		result[i] = vec
	}
	return result, nil
}

func (p *coverageEmbeddingProvider) Dimension() int    { return p.dimension }
func (p *coverageEmbeddingProvider) ModelName() string { return p.modelName }

type coverageVectorDB struct {
	connected bool
	indexes   map[string]*IndexConfig
	vectors   map[string]map[string]*Vector
}

func newCoverageVectorDB() *coverageVectorDB {
	return &coverageVectorDB{
		indexes: make(map[string]*IndexConfig),
		vectors: make(map[string]map[string]*Vector),
	}
}

func (m *coverageVectorDB) Connect(ctx context.Context) error {
	m.connected = true
	return nil
}
func (m *coverageVectorDB) Disconnect(ctx context.Context) error {
	m.connected = false
	return nil
}
func (m *coverageVectorDB) IsConnected() bool { return m.connected }
func (m *coverageVectorDB) CreateIndex(ctx context.Context, config *IndexConfig) error {
	m.indexes[config.Name] = config
	m.vectors[config.Name] = make(map[string]*Vector)
	return nil
}
func (m *coverageVectorDB) DropIndex(ctx context.Context, indexName string) error {
	delete(m.indexes, indexName)
	delete(m.vectors, indexName)
	return nil
}
func (m *coverageVectorDB) HasIndex(ctx context.Context, indexName string) (bool, error) {
	_, ok := m.indexes[indexName]
	return ok, nil
}
func (m *coverageVectorDB) ListIndexes(ctx context.Context) ([]string, error) {
	var names []string
	for name := range m.indexes {
		names = append(names, name)
	}
	return names, nil
}
func (m *coverageVectorDB) DescribeIndex(ctx context.Context, indexName string) (*IndexConfig, error) {
	cfg, ok := m.indexes[indexName]
	if !ok {
		return nil, ErrIndexNotFound
	}
	return cfg, nil
}
func (m *coverageVectorDB) Insert(ctx context.Context, indexName string, vectors []*Vector) error {
	for _, v := range vectors {
		m.vectors[indexName][v.ID] = v
	}
	return nil
}
func (m *coverageVectorDB) Upsert(ctx context.Context, indexName string, vectors []*Vector) error {
	for _, v := range vectors {
		m.vectors[indexName][v.ID] = v
	}
	return nil
}
func (m *coverageVectorDB) Delete(ctx context.Context, indexName string, ids []string) error {
	for _, id := range ids {
		delete(m.vectors[indexName], id)
	}
	return nil
}
func (m *coverageVectorDB) Get(ctx context.Context, indexName string, ids []string) ([]*Vector, error) {
	var result []*Vector
	for _, id := range ids {
		if v, ok := m.vectors[indexName][id]; ok {
			result = append(result, v)
		}
	}
	return result, nil
}
func (m *coverageVectorDB) Search(ctx context.Context, indexName string, vector []float32, params *SearchParams) ([]*SearchResult, error) {
	return nil, nil
}
func (m *coverageVectorDB) SearchBatch(ctx context.Context, indexName string, vectors [][]float32, params *SearchParams) ([][]*SearchResult, error) {
	return nil, nil
}
func (m *coverageVectorDB) Count(ctx context.Context, indexName string) (int64, error) {
	if vs, ok := m.vectors[indexName]; ok {
		return int64(len(vs)), nil
	}
	return 0, nil
}
func (m *coverageVectorDB) HealthCheck(ctx context.Context) error {
	if !m.connected {
		return ErrVectorDBNotConnected
	}
	return nil
}

func TestTextSplitter_SplitBySize(t *testing.T) {
	config := &ChunkConfig{
		ChunkSize:    20,
		ChunkOverlap: 5,
		MinChunkSize: 5,
	}
	splitter := NewTextSplitter(config, nil)
	text := strings.Repeat("abcdefghij", 10)
	chunks, err := splitter.Split(text)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(chunks), 1)
	for _, chunk := range chunks {
		assert.NotEmpty(t, chunk.ID)
		assert.GreaterOrEqual(t, chunk.Position, 0)
	}
}

func TestTextSplitter_SplitBySentence(t *testing.T) {
	config := &ChunkConfig{
		ChunkSize:        100,
		ChunkOverlap:     10,
		MinChunkSize:     5,
		SplitBySentence:  true,
	}
	splitter := NewTextSplitter(config, nil)
	text := "This is sentence one. This is sentence two! Is this sentence three? Yes it is."
	chunks, err := splitter.Split(text)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(chunks), 1)
}

func TestTextSplitter_SplitByParagraph(t *testing.T) {
	config := &ChunkConfig{
		ChunkSize:         200,
		MinChunkSize:      5,
		MaxChunkSize:      500,
		SplitByParagraph:  true,
		SplitBySentence:   false,
	}
	splitter := NewTextSplitter(config, nil)
	text := "First paragraph here.\n\nSecond paragraph here with more content.\n\nThird paragraph."
	chunks, err := splitter.Split(text)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(chunks), 1)
}

func TestTextSplitter_EmptyText(t *testing.T) {
	splitter := NewTextSplitter(DefaultChunkConfig(), nil)
	_, err := splitter.Split("")
	assert.ErrorIs(t, err, ErrEmptyDocument)
}

func TestTextSplitter_InvalidChunkSize(t *testing.T) {
	config := &ChunkConfig{ChunkSize: 0}
	splitter := NewTextSplitter(config, nil)
	_, err := splitter.Split("some text")
	assert.ErrorIs(t, err, ErrInvalidChunkSize)
}

func TestTextSplitter_NilConfig(t *testing.T) {
	splitter := NewTextSplitter(nil, nil)
	require.NotNil(t, splitter)
	assert.Equal(t, 512, splitter.config.ChunkSize)
}

func TestTextSplitter_NilLogger(t *testing.T) {
	splitter := NewTextSplitter(DefaultChunkConfig(), nil)
	require.NotNil(t, splitter)
}

func TestGetLastNChars(t *testing.T) {
	assert.Equal(t, "hello", getLastNChars("hello", 10))
	assert.Equal(t, "llo", getLastNChars("hello", 3))
	assert.Equal(t, "世界", getLastNChars("你好世界", 2))
}

func TestEmbeddingCache_Operations(t *testing.T) {
	cache := NewEmbeddingCache(3, nil)
	assert.Equal(t, 0, cache.Size())

	vec := []float32{0.1, 0.2, 0.3}
	cache.Set("text1", vec)
	assert.Equal(t, 1, cache.Size())

	got, exists := cache.Get("text1")
	assert.True(t, exists)
	assert.Equal(t, vec, got)

	_, exists = cache.Get("nonexistent")
	assert.False(t, exists)

	cache.Set("text2", vec)
	cache.Set("text3", vec)
	assert.Equal(t, 3, cache.Size())

	cache.Set("text4", vec)
	assert.LessOrEqual(t, cache.Size(), 3)

	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}

func TestOpenAIEmbeddingProvider(t *testing.T) {
	config := &OpenAIEmbeddingConfig{
		Model: "text-embedding-ada-002",
	}
	provider := NewOpenAIEmbeddingProvider(config, nil)
	assert.Equal(t, 1536, provider.Dimension())
	assert.Equal(t, "text-embedding-ada-002", provider.ModelName())

	embeddings, err := provider.EmbedBatch(context.Background(), []string{"hello", "world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 2)
	assert.Len(t, embeddings[0], 1536)

	emb, err := provider.Embed(context.Background(), "hello")
	require.NoError(t, err)
	assert.Len(t, emb, 1536)
}

func TestOpenAIEmbeddingProvider_Models(t *testing.T) {
	tests := []struct {
		model      string
		dimension  int
	}{
		{"text-embedding-ada-002", 1536},
		{"text-embedding-3-small", 1536},
		{"text-embedding-3-large", 3072},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			p := NewOpenAIEmbeddingProvider(&OpenAIEmbeddingConfig{Model: tt.model}, nil)
			assert.Equal(t, tt.dimension, p.Dimension())
		})
	}
}

func TestOpenAIEmbeddingProvider_EmptyText(t *testing.T) {
	p := NewOpenAIEmbeddingProvider(&OpenAIEmbeddingConfig{Model: "text-embedding-ada-002"}, nil)
	_, err := p.Embed(context.Background(), "")
	assert.ErrorIs(t, err, ErrEmptyDocument)
}

func TestOpenAIEmbeddingProvider_EmptyBatch(t *testing.T) {
	p := NewOpenAIEmbeddingProvider(&OpenAIEmbeddingConfig{Model: "text-embedding-ada-002"}, nil)
	_, err := p.EmbedBatch(context.Background(), []string{})
	assert.ErrorIs(t, err, ErrEmptyDocument)
}

func TestLocalEmbeddingProvider(t *testing.T) {
	config := &LocalEmbeddingConfig{
		ModelPath: "/tmp/model",
		ModelType: "sentence-transformers",
		Dimension: 384,
	}
	provider := NewLocalEmbeddingProvider(config, nil)
	assert.Equal(t, 384, provider.Dimension())
	assert.Equal(t, "sentence-transformers", provider.ModelName())
	assert.False(t, provider.IsReady())

	err := provider.Initialize(context.Background())
	require.NoError(t, err)
	assert.True(t, provider.IsReady())

	emb, err := provider.Embed(context.Background(), "test text")
	require.NoError(t, err)
	assert.Len(t, emb, 384)

	batch, err := provider.EmbedBatch(context.Background(), []string{"a", "b"})
	require.NoError(t, err)
	assert.Len(t, batch, 2)
}

func TestLocalEmbeddingProvider_NotReady(t *testing.T) {
	provider := NewLocalEmbeddingProvider(&LocalEmbeddingConfig{Dimension: 128}, nil)
	_, err := provider.Embed(context.Background(), "test")
	assert.ErrorIs(t, err, ErrProviderNotReady)

	_, err = provider.EmbedBatch(context.Background(), []string{"test"})
	assert.ErrorIs(t, err, ErrProviderNotReady)
}

func TestLocalEmbeddingProvider_EmptyBatch(t *testing.T) {
	provider := NewLocalEmbeddingProvider(&LocalEmbeddingConfig{Dimension: 128}, nil)
	provider.Initialize(context.Background())
	_, err := provider.EmbedBatch(context.Background(), []string{})
	assert.ErrorIs(t, err, ErrEmptyDocument)
}

func TestDocumentEmbedder_EmbedDocument(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test-embed"}
	config := &DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}
	embedder := NewDocumentEmbedder(config, nil)

	doc := &Document{
		ID:      "doc-embed-1",
		Content: "This is a test document. It has multiple sentences. Each sentence provides information.",
		Title:   "Test Doc",
		Source:  "test",
		Metadata: map[string]interface{}{"knowledge_base_id": "kb-1"},
	}

	chunks, err := embedder.EmbedDocument(context.Background(), doc)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(chunks), 1)
	for _, chunk := range chunks {
		assert.Equal(t, "doc-embed-1", chunk.DocumentID)
		assert.NotEmpty(t, chunk.Embedding)
		assert.Equal(t, "Test Doc", chunk.Metadata["title"])
	}
}

func TestDocumentEmbedder_EmbedDocument_NilDoc(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	config := &DocumentEmbedderConfig{Provider: provider}
	embedder := NewDocumentEmbedder(config, nil)

	_, err := embedder.EmbedDocument(context.Background(), nil)
	assert.ErrorIs(t, err, ErrEmptyDocument)
}

func TestDocumentEmbedder_EmbedDocument_EmptyContent(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	config := &DocumentEmbedderConfig{Provider: provider}
	embedder := NewDocumentEmbedder(config, nil)

	_, err := embedder.EmbedDocument(context.Background(), &Document{ID: "empty"})
	assert.ErrorIs(t, err, ErrEmptyDocument)
}

func TestDocumentEmbedder_EmbedWithCache(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	config := &DocumentEmbedderConfig{Provider: provider, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)

	emb1, err := embedder.Embed(context.Background(), "test text")
	require.NoError(t, err)
	assert.Len(t, emb1, 3)

	emb2, err := embedder.Embed(context.Background(), "test text")
	require.NoError(t, err)
	assert.Equal(t, emb1, emb2)
}

func TestDocumentEmbedder_EmbedBatchWithCache(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	config := &DocumentEmbedderConfig{Provider: provider, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)

	embedder.Embed(context.Background(), "cached text")

	results, err := embedder.EmbedBatch(context.Background(), []string{"cached text", "new text"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestDocumentEmbedder_GetCacheStats(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test-model"}
	config := &DocumentEmbedderConfig{Provider: provider, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)

	stats := embedder.GetCacheStats()
	assert.Equal(t, 0, stats["cache_size"])
	assert.Equal(t, 3, stats["dimension"])
	assert.Equal(t, "test-model", stats["model"])
}

func TestDocumentEmbedder_ClearCache(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	config := &DocumentEmbedderConfig{Provider: provider, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)

	embedder.Embed(context.Background(), "text to cache")
	stats := embedder.GetCacheStats()
	assert.Equal(t, 1, stats["cache_size"])

	embedder.ClearCache()
	stats = embedder.GetCacheStats()
	assert.Equal(t, 0, stats["cache_size"])
}

func TestDocumentEmbedder_NilConfig(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	config := &DocumentEmbedderConfig{Provider: provider}
	embedder := NewDocumentEmbedder(config, nil)
	require.NotNil(t, embedder)
}

func TestEmbeddingService_ProcessDocument(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	config := &DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}
	embedder := NewDocumentEmbedder(config, nil)
	service := NewEmbeddingService(embedder, nil)

	doc := &Document{
		ID:      "svc-doc-1",
		Content: "Service test document with enough content for processing.",
		Metadata: map[string]interface{}{"knowledge_base_id": "kb-1"},
	}

	chunks, err := service.ProcessDocument(context.Background(), doc)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(chunks), 1)

	stats := service.GetStats()
	assert.Equal(t, int64(1), stats.TotalDocuments)
	assert.GreaterOrEqual(t, stats.TotalChunks, int64(1))
}

func TestEmbeddingService_BatchEmbedDocuments(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	config := &DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}
	embedder := NewDocumentEmbedder(config, nil)
	service := NewEmbeddingService(embedder, nil)

	req := &BatchEmbeddingRequest{
		Documents: []*Document{
			{ID: "batch-1", Content: "First document content for batch processing.", Metadata: map[string]interface{}{"knowledge_base_id": "kb-1"}},
			{ID: "batch-2", Content: "Second document content for batch processing.", Metadata: map[string]interface{}{"knowledge_base_id": "kb-1"}},
		},
		BatchSize: 10,
	}

	results := service.BatchEmbedDocuments(context.Background(), req)
	assert.Len(t, results, 2)
	assert.Empty(t, results[0].Error)
	assert.GreaterOrEqual(t, len(results[0].Chunks), 1)
}

func TestEmbeddingService_BatchEmbedDocuments_WithErrors(t *testing.T) {
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	config := &DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}
	embedder := NewDocumentEmbedder(config, nil)
	service := NewEmbeddingService(embedder, nil)

	req := &BatchEmbeddingRequest{
		Documents: []*Document{
			{ID: "ok-doc", Content: "Valid document.", Metadata: map[string]interface{}{"knowledge_base_id": "kb-1"}},
			{ID: "empty-doc", Content: "", Metadata: map[string]interface{}{"knowledge_base_id": "kb-1"}},
		},
	}

	results := service.BatchEmbedDocuments(context.Background(), req)
	assert.Len(t, results, 2)
	assert.Empty(t, results[0].Error)
	assert.NotEmpty(t, results[1].Error)
}

func TestSimilarityCalculator_CosineSimilarity(t *testing.T) {
	calc := NewSimilarityCalculator()
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{0.0, 1.0, 0.0}
	sim, err := calc.CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.LessOrEqual(t, sim, float32(0.01))

	c := []float32{1.0, 0.0, 0.0}
	sim, err = calc.CosineSimilarity(a, c)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, sim, float32(0.99))
}

func TestSimilarityCalculator_FindMostSimilar(t *testing.T) {
	calc := NewSimilarityCalculator()
	query := []float32{1.0, 0.0, 0.0}
	vectors := [][]float32{
		{0.9, 0.1, 0.0},
		{0.0, 1.0, 0.0},
		{0.8, 0.2, 0.0},
	}
	indices, values, err := calc.FindMostSimilar(query, vectors, 2)
	require.NoError(t, err)
	assert.Len(t, indices, 2)
	assert.Len(t, values, 2)
	assert.GreaterOrEqual(t, values[0], values[1])
}

func TestSimilarityCalculator_FindMostSimilar_Empty(t *testing.T) {
	calc := NewSimilarityCalculator()
	indices, values, err := calc.FindMostSimilar([]float32{1.0}, [][]float32{}, 5)
	require.NoError(t, err)
	assert.Nil(t, indices)
	assert.Nil(t, values)
}

func TestInMemoryDocumentStore_AllOperations(t *testing.T) {
	store := NewInMemoryDocumentStore(nil)
	ctx := context.Background()

	doc1 := &Document{
		ID:      "ds-doc-1",
		Content: "Document one content",
		Title:   "Doc One",
		Source:  "upload",
		Metadata: map[string]interface{}{"knowledge_base_id": "kb-store-test"},
	}

	err := store.Save(ctx, doc1)
	require.NoError(t, err)

	got, err := store.Get(ctx, "ds-doc-1")
	require.NoError(t, err)
	assert.Equal(t, "Doc One", got.Title)

	_, err = store.Get(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrDocumentNotFound)

	docs, err := store.List(ctx, "kb-store-test", 0, 10)
	require.NoError(t, err)
	assert.Len(t, docs, 1)

	count, err := store.Count(ctx, "kb-store-test")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = store.Count(ctx, "nonexistent-kb")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	docs, err = store.List(ctx, "nonexistent-kb", 0, 10)
	require.NoError(t, err)
	assert.Len(t, docs, 0)
}

func TestInMemoryDocumentStore_SaveDuplicate(t *testing.T) {
	store := NewInMemoryDocumentStore(nil)
	ctx := context.Background()

	doc := &Document{
		ID:       "dup-doc",
		Content:  "Original",
		Metadata: map[string]interface{}{"knowledge_base_id": "kb-dup"},
	}
	store.Save(ctx, doc)

	doc.Content = "Updated"
	store.Save(ctx, doc)

	docs, _ := store.List(ctx, "kb-dup", 0, 10)
	assert.Len(t, docs, 1)
}

func TestInMemoryDocumentStore_Delete(t *testing.T) {
	store := NewInMemoryDocumentStore(nil)
	ctx := context.Background()

	doc := &Document{
		ID:       "del-doc",
		Content:  "To delete",
		Metadata: map[string]interface{}{"knowledge_base_id": "kb-del"},
	}
	store.Save(ctx, doc)

	err := store.Delete(ctx, "del-doc")
	require.NoError(t, err)

	_, err = store.Get(ctx, "del-doc")
	assert.ErrorIs(t, err, ErrDocumentNotFound)

	err = store.Delete(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrDocumentNotFound)
}

func TestInMemoryDocumentStore_ListPagination(t *testing.T) {
	store := NewInMemoryDocumentStore(nil)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		store.Save(ctx, &Document{
			ID:       "page-doc-" + string(rune('A'+i)),
			Content:  "Content " + string(rune('A'+i)),
			Metadata: map[string]interface{}{"knowledge_base_id": "kb-page"},
		})
	}

	docs, err := store.List(ctx, "kb-page", 0, 3)
	require.NoError(t, err)
	assert.Len(t, docs, 3)

	docs, err = store.List(ctx, "kb-page", 3, 3)
	require.NoError(t, err)
	assert.Len(t, docs, 2)

	docs, err = store.List(ctx, "kb-page", 100, 10)
	require.NoError(t, err)
	assert.Len(t, docs, 0)
}

func TestKnowledgeManager_CreateKnowledgeBase(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())

	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedderConfig := &DocumentEmbedderConfig{Provider: provider, CacheSize: 100}
	embedder := NewDocumentEmbedder(embedderConfig, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)

	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, err := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{
		Name:        "Test KB",
		Description: "A test knowledge base",
		Dimension:   3,
		Metric:      "cosine",
	})
	require.NoError(t, err)
	assert.Equal(t, "Test KB", kb.Name)
	assert.Equal(t, KnowledgeBaseStatusActive, kb.Status)
	assert.NotEmpty(t, kb.ID)
	assert.NotEmpty(t, kb.IndexName)
}

func TestKnowledgeManager_CreateKnowledgeBase_Validation(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	_, err := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "", Dimension: 3})
	assert.Error(t, err)

	_, err = mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 0})
	assert.Error(t, err)
}

func TestKnowledgeManager_GetKnowledgeBase(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "GetTest", Dimension: 3})

	got, err := mgr.GetKnowledgeBase(context.Background(), kb.ID)
	require.NoError(t, err)
	assert.Equal(t, kb.ID, got.ID)

	_, err = mgr.GetKnowledgeBase(context.Background(), "")
	assert.ErrorIs(t, err, ErrInvalidKnowledgeBaseID)

	_, err = mgr.GetKnowledgeBase(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, ErrKnowledgeBaseNotFound)
}

func TestKnowledgeManager_UpdateKnowledgeBase(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "UpdateTest", Dimension: 3})

	updated, err := mgr.UpdateKnowledgeBase(context.Background(), kb.ID, &UpdateKnowledgeBaseRequest{
		Name:        "Updated Name",
		Description: "Updated description",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated description", updated.Description)

	_, err = mgr.UpdateKnowledgeBase(context.Background(), "", nil)
	assert.ErrorIs(t, err, ErrInvalidKnowledgeBaseID)

	_, err = mgr.UpdateKnowledgeBase(context.Background(), "nonexistent", nil)
	assert.ErrorIs(t, err, ErrKnowledgeBaseNotFound)
}

func TestKnowledgeManager_DeleteKnowledgeBase(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "DeleteTest", Dimension: 3})

	err := mgr.DeleteKnowledgeBase(context.Background(), kb.ID)
	require.NoError(t, err)

	_, err = mgr.GetKnowledgeBase(context.Background(), kb.ID)
	assert.ErrorIs(t, err, ErrKnowledgeBaseNotFound)

	err = mgr.DeleteKnowledgeBase(context.Background(), "")
	assert.ErrorIs(t, err, ErrInvalidKnowledgeBaseID)
}

func TestKnowledgeManager_ListKnowledgeBases(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "KB1", Dimension: 3})
	mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "KB2", Dimension: 3})

	kbs, err := mgr.ListKnowledgeBases(context.Background(), 0, 10)
	require.NoError(t, err)
	assert.Len(t, kbs, 2)
}

func TestKnowledgeManager_UploadDocument(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "UploadTest", Dimension: 3})

	docInfo, err := mgr.UploadDocument(context.Background(), &UploadDocumentRequest{
		KnowledgeBaseID: kb.ID,
		Title:           "Test Document",
		Content:         "This is a test document with enough content for embedding and indexing.",
		Source:          "test",
		ContentType:     "text/plain",
	})
	require.NoError(t, err)
	assert.Equal(t, kb.ID, docInfo.KnowledgeBaseID)
	assert.Equal(t, "Test Document", docInfo.Title)
	assert.Equal(t, "indexed", docInfo.Status)
	assert.GreaterOrEqual(t, docInfo.ChunkCount, 1)
}

func TestKnowledgeManager_UploadDocument_Validation(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	_, err := mgr.UploadDocument(context.Background(), &UploadDocumentRequest{KnowledgeBaseID: "", Content: "test"})
	assert.ErrorIs(t, err, ErrInvalidKnowledgeBaseID)

	_, err = mgr.UploadDocument(context.Background(), &UploadDocumentRequest{KnowledgeBaseID: "kb-x", Content: ""})
	assert.ErrorIs(t, err, ErrEmptyDocument)

	_, err = mgr.UploadDocument(context.Background(), &UploadDocumentRequest{KnowledgeBaseID: "nonexistent", Content: "test"})
	assert.ErrorIs(t, err, ErrKnowledgeBaseNotFound)
}

func TestKnowledgeManager_DeleteDocument(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "DelDocTest", Dimension: 3})

	docInfo, _ := mgr.UploadDocument(context.Background(), &UploadDocumentRequest{
		KnowledgeBaseID: kb.ID,
		Title:           "Doc to Delete",
		Content:         "Content for deletion test with enough text for processing.",
	})

	err := mgr.DeleteDocument(context.Background(), kb.ID, docInfo.ID)
	require.NoError(t, err)

	_, err = mgr.GetDocument(context.Background(), docInfo.ID)
	assert.ErrorIs(t, err, ErrDocumentNotFound)
}

func TestKnowledgeManager_DeleteDocument_Validation(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	err := mgr.DeleteDocument(context.Background(), "", "doc-1")
	assert.ErrorIs(t, err, ErrInvalidKnowledgeBaseID)

	err = mgr.DeleteDocument(context.Background(), "kb-1", "")
	assert.ErrorIs(t, err, ErrInvalidDocumentID)
}

func TestKnowledgeManager_GetDocument(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	_, err := mgr.GetDocument(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, ErrDocumentNotFound)
}

func TestKnowledgeManager_ListDocuments(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "ListDocTest", Dimension: 3})

	mgr.UploadDocument(context.Background(), &UploadDocumentRequest{
		KnowledgeBaseID: kb.ID,
		Title:           "Doc1",
		Content:         "First document content for listing test.",
	})
	mgr.UploadDocument(context.Background(), &UploadDocumentRequest{
		KnowledgeBaseID: kb.ID,
		Title:           "Doc2",
		Content:         "Second document content for listing test.",
	})

	docInfos, err := mgr.ListDocuments(context.Background(), kb.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, docInfos, 2)
}

func TestKnowledgeManager_GetKnowledgeBaseStats(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "StatsTest", Dimension: 3})

	stats, err := mgr.GetKnowledgeBaseStats(context.Background(), kb.ID)
	require.NoError(t, err)
	assert.Equal(t, kb.ID, stats.KnowledgeBaseID)

	_, err = mgr.GetKnowledgeBaseStats(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, ErrKnowledgeBaseNotFound)
}

func TestKnowledgeManager_CheckHealth(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "HealthTest", Dimension: 3})

	health, err := mgr.CheckHealth(context.Background(), kb.ID)
	require.NoError(t, err)
	assert.Equal(t, "healthy", health.VectorDBStatus)
	assert.Equal(t, string(KnowledgeBaseStatusActive), health.Status)

	_, err = mgr.CheckHealth(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, ErrKnowledgeBaseNotFound)
}

func TestKnowledgeManager_GetStats(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Stats1", Dimension: 3})
	mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Stats2", Dimension: 3})

	stats := mgr.GetStats()
	assert.Equal(t, int64(2), stats.TotalKnowledgeBases)
	assert.Equal(t, int64(2), stats.ActiveKnowledgeBases)
}

func TestKnowledgeManager_ExportKnowledgeBase(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "ExportTest", Dimension: 3})

	export, err := mgr.ExportKnowledgeBase(context.Background(), kb.ID)
	require.NoError(t, err)
	assert.NotNil(t, export["knowledge_base"])
	assert.NotNil(t, export["documents"])

	_, err = mgr.ExportKnowledgeBase(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, ErrKnowledgeBaseNotFound)
}

func TestKnowledgeManager_BatchUploadDocuments(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "BatchTest", Dimension: 3})

	results, errors := mgr.BatchUploadDocuments(context.Background(), []*UploadDocumentRequest{
		{KnowledgeBaseID: kb.ID, Title: "Batch1", Content: "First batch document content."},
		{KnowledgeBaseID: kb.ID, Title: "Batch2", Content: "Second batch document content."},
	})
	assert.Len(t, results, 2)
	assert.Len(t, errors, 2)
	assert.Nil(t, errors[0])
	assert.Nil(t, errors[1])
}

func TestSimpleKeywordSearcher_Operations(t *testing.T) {
	searcher := NewSimpleKeywordSearcher(nil)
	ctx := context.Background()

	chunks := []*TextChunk{
		{ID: "kw-1", Content: "solar panel installation guide"},
		{ID: "kw-2", Content: "wind turbine maintenance manual"},
		{ID: "kw-3", Content: "solar energy storage system"},
	}
	err := searcher.Index(ctx, chunks)
	require.NoError(t, err)

	results, err := searcher.Search(ctx, "solar panel", 5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)

	_, err = searcher.Search(ctx, "", 5)
	assert.ErrorIs(t, err, ErrInvalidQuery)

	err = searcher.Delete(ctx, []string{"kw-1"})
	require.NoError(t, err)
}

func TestSimpleReranker_Rerank(t *testing.T) {
	reranker := NewSimpleReranker(nil)
	ctx := context.Background()

	results := []*RetrievalResult{
		{Chunk: &TextChunk{Content: "solar panel efficiency"}, Score: 0.8},
		{Chunk: &TextChunk{Content: "wind turbine output"}, Score: 0.6},
	}

	reranked, err := reranker.Rerank(ctx, "solar energy", results)
	require.NoError(t, err)
	assert.Len(t, reranked, 2)
	assert.Equal(t, 1, reranked[0].Rank)
}

func TestSimpleReranker_EmptyResults(t *testing.T) {
	reranker := NewSimpleReranker(nil)
	results, err := reranker.Rerank(context.Background(), "query", []*RetrievalResult{})
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestContextWindowManager_BuildContext(t *testing.T) {
	mgr := NewContextWindowManager(500, nil)

	results := []*RetrievalResult{
		{Chunk: &TextChunk{Content: "First chunk of text for context building."}, Score: 0.9},
		{Chunk: &TextChunk{Content: "Second chunk of text for context building."}, Score: 0.8},
		{Chunk: &TextChunk{Content: "Third chunk of text for context building."}, Score: 0.7},
	}

	context := mgr.BuildContext(results)
	assert.Contains(t, context, "文档1")
	assert.Contains(t, context, "First chunk")
}

func TestContextWindowManager_Truncation(t *testing.T) {
	mgr := NewContextWindowManager(50, nil)

	results := []*RetrievalResult{
		{Chunk: &TextChunk{Content: strings.Repeat("Long content here ", 20)}, Score: 0.9},
		{Chunk: &TextChunk{Content: "Short"}, Score: 0.8},
	}

	context := mgr.BuildContext(results)
	assert.LessOrEqual(t, len(context), 200)
}

func TestCitationTracker_ExtractCitations(t *testing.T) {
	tracker := NewCitationTracker(nil)

	results := []*RetrievalResult{
		{
			Chunk: &TextChunk{ID: "c-1", DocumentID: "d-1", Content: "Content one"},
			Score: 0.9,
		},
		{
			Chunk: &TextChunk{
				ID:         "c-2",
				DocumentID: "d-2",
				Content:    "Content two",
				Metadata:   map[string]interface{}{"title": "Doc Two", "source": "upload"},
			},
			Score: 0.8,
		},
	}

	citations := tracker.ExtractCitations(results)
	assert.Len(t, citations, 2)
	assert.Equal(t, "c-1", citations[0].ChunkID)
	assert.Equal(t, "Doc Two", citations[1].Title)
	assert.Equal(t, "upload", citations[1].Source)
}

func TestRAGRetriever_IndexAndDeleteChunks(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)

	vdb.CreateIndex(context.Background(), &IndexConfig{Name: "test-idx", Dimension: 3})

	chunks := []*TextChunk{
		{ID: "idx-1", DocumentID: "doc-1", Content: "Test chunk one", Embedding: []float32{0.1, 0.2, 0.3}},
		{ID: "idx-2", DocumentID: "doc-1", Content: "Test chunk two", Embedding: []float32{0.4, 0.5, 0.6}},
	}

	err := retriever.IndexChunks(context.Background(), "test-idx", chunks)
	require.NoError(t, err)

	err = retriever.DeleteChunks(context.Background(), "test-idx", []string{"idx-1"})
	require.NoError(t, err)
}

func TestRAGRetriever_GetCitations(t *testing.T) {
	vdb := newCoverageVectorDB()
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)

	results := []*RetrievalResult{
		{Chunk: &TextChunk{ID: "cit-1", DocumentID: "doc-1", Content: "Citation content"}, Score: 0.9},
	}

	citations := retriever.GetCitations(results)
	assert.Len(t, citations, 1)
	assert.Equal(t, "cit-1", citations[0].ChunkID)
}

func TestRAGService_Query(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, &RAGConfig{
		UseHybridSearch: false,
		UseReranker:     false,
		ScoreThreshold:  0.0,
		TopK:            5,
		FinalTopK:       3,
	}, nil)
	service := NewRAGService(retriever, nil)

	stats := service.GetStats()
	assert.Equal(t, int64(0), stats.TotalQueries)
}

func TestRAGService_HybridSearch(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, &RAGConfig{
		UseHybridSearch: true,
		UseReranker:     false,
		ScoreThreshold:  0.0,
		TopK:            5,
		FinalTopK:       3,
	}, nil)
	service := NewRAGService(retriever, nil)

	vdb.CreateIndex(context.Background(), &IndexConfig{Name: "hybrid-idx", Dimension: 3})
	chunks := []*TextChunk{
		{ID: "hy-1", DocumentID: "doc-1", Content: "solar panel data", Embedding: []float32{0.1, 0.2, 0.3}},
	}
	retriever.IndexChunks(context.Background(), "hybrid-idx", chunks)

	result, err := service.HybridSearch(context.Background(), "solar panel", "hybrid-idx")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMultiQueryRetriever_EmptyQueries(t *testing.T) {
	vdb := newCoverageVectorDB()
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mqr := NewMultiQueryRetriever(retriever, nil)

	_, err := mqr.Retrieve(context.Background(), []string{}, "idx")
	assert.ErrorIs(t, err, ErrInvalidQuery)
}

func TestMilvusClient_Operations(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{
		Address: "localhost",
		Port:    19530,
	}, nil)

	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, client.IsConnected())

	err = client.Connect(ctx)
	require.NoError(t, err)

	err = client.CreateIndex(ctx, &IndexConfig{Name: "milvus-test", Dimension: 3, Metric: "cosine"})
	require.NoError(t, err)

	has, err := client.HasIndex(ctx, "milvus-test")
	require.NoError(t, err)
	assert.True(t, has)

	indexes, err := client.ListIndexes(ctx)
	require.NoError(t, err)
	assert.Contains(t, indexes, "milvus-test")

	cfg, err := client.DescribeIndex(ctx, "milvus-test")
	require.NoError(t, err)
	assert.Equal(t, 3, cfg.Dimension)

	_, err = client.DescribeIndex(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrIndexNotFound)

	err = client.CreateIndex(ctx, &IndexConfig{Name: "milvus-test", Dimension: 3})
	assert.Error(t, err)

	err = client.Insert(ctx, "milvus-test", []*Vector{
		{ID: "v1", Data: []float32{0.1, 0.2, 0.3}},
	})
	require.NoError(t, err)

	err = client.Upsert(ctx, "milvus-test", []*Vector{
		{ID: "v2", Data: []float32{0.4, 0.5, 0.6}},
	})
	require.NoError(t, err)

	err = client.Insert(ctx, "milvus-test", []*Vector{
		{ID: "v3", Data: []float32{0.1, 0.2}},
	})
	assert.Error(t, err)

	_, err = client.Get(ctx, "milvus-test", []string{"v1"})
	require.NoError(t, err)

	err = client.Delete(ctx, "milvus-test", []string{"v1"})
	require.NoError(t, err)

	_, err = client.Search(ctx, "milvus-test", []float32{0.1, 0.2, 0.3}, &SearchParams{TopK: 5})
	require.NoError(t, err)

	batchResults, err := client.SearchBatch(ctx, "milvus-test", [][]float32{{0.1, 0.2, 0.3}}, &SearchParams{TopK: 5})
	require.NoError(t, err)
	assert.Len(t, batchResults, 1)

	_, err = client.Count(ctx, "milvus-test")
	require.NoError(t, err)

	err = client.HealthCheck(ctx)
	require.NoError(t, err)

	err = client.DropIndex(ctx, "milvus-test")
	require.NoError(t, err)

	err = client.DropIndex(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrIndexNotFound)

	err = client.Disconnect(ctx)
	require.NoError(t, err)
	assert.False(t, client.IsConnected())

	err = client.Disconnect(ctx)
	require.NoError(t, err)
}

func TestMilvusClient_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	ctx := context.Background()

	assert.False(t, client.IsConnected())

	err := client.CreateIndex(ctx, &IndexConfig{Name: "x", Dimension: 3})
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	err = client.DropIndex(ctx, "x")
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	_, err = client.HasIndex(ctx, "x")
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	_, err = client.ListIndexes(ctx)
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	_, err = client.DescribeIndex(ctx, "x")
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	err = client.Insert(ctx, "x", nil)
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	err = client.Upsert(ctx, "x", nil)
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	err = client.Delete(ctx, "x", nil)
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	_, err = client.Get(ctx, "x", nil)
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	_, err = client.Search(ctx, "x", nil, nil)
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	_, err = client.SearchBatch(ctx, "x", nil, nil)
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	_, err = client.Count(ctx, "x")
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)

	err = client.HealthCheck(ctx)
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)
}

func TestMilvusClient_InvalidDimension(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	ctx := context.Background()
	client.Connect(ctx)

	err := client.CreateIndex(ctx, &IndexConfig{Name: "inv-dim", Dimension: 0})
	assert.ErrorIs(t, err, ErrInvalidVector)

	err = client.CreateIndex(ctx, &IndexConfig{Name: "inv-dim", Dimension: 3})
	require.NoError(t, err)

	_, err = client.Search(ctx, "inv-dim", []float32{0.1, 0.2}, nil)
	assert.ErrorIs(t, err, ErrInvalidVector)

	_, err = client.SearchBatch(ctx, "inv-dim", [][]float32{{0.1, 0.2}}, nil)
	assert.ErrorIs(t, err, ErrInvalidVector)
}

func TestMilvusClient_InsertIndexNotFound(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	ctx := context.Background()
	client.Connect(ctx)

	err := client.Insert(ctx, "nonexistent", []*Vector{{ID: "v1", Data: []float32{0.1}}})
	assert.ErrorIs(t, err, ErrIndexNotFound)

	err = client.Upsert(ctx, "nonexistent", []*Vector{{ID: "v1", Data: []float32{0.1}}})
	assert.ErrorIs(t, err, ErrIndexNotFound)

	err = client.Delete(ctx, "nonexistent", []string{"v1"})
	assert.ErrorIs(t, err, ErrIndexNotFound)

	_, err = client.Get(ctx, "nonexistent", []string{"v1"})
	assert.ErrorIs(t, err, ErrIndexNotFound)

	_, err = client.Count(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrIndexNotFound)
}

func TestPineconeClient_Operations(t *testing.T) {
	client := NewPineconeClient(&PineconeConfig{
		APIKey:      "test-key",
		Environment: "us-west1",
		ProjectID:   "proj-1",
	}, nil)
	ctx := context.Background()

	client.Connect(ctx)
	assert.True(t, client.IsConnected())

	client.CreateIndex(ctx, &IndexConfig{Name: "pinecone-test", Dimension: 3, Metric: "cosine"})

	has, _ := client.HasIndex(ctx, "pinecone-test")
	assert.True(t, has)

	client.Insert(ctx, "pinecone-test", []*Vector{{ID: "pv1", Data: []float32{0.1, 0.2, 0.3}}})
	client.Upsert(ctx, "pinecone-test", []*Vector{{ID: "pv2", Data: []float32{0.4, 0.5, 0.6}}})
	client.Delete(ctx, "pinecone-test", []string{"pv1"})
	client.Get(ctx, "pinecone-test", []string{"pv2"})
	client.Search(ctx, "pinecone-test", []float32{0.1, 0.2, 0.3}, nil)
	client.SearchBatch(ctx, "pinecone-test", [][]float32{{0.1, 0.2, 0.3}}, nil)
	client.Count(ctx, "pinecone-test")
	client.HealthCheck(ctx)
	client.DropIndex(ctx, "pinecone-test")
	client.Disconnect(ctx)
}

func TestWeaviateClient_Operations(t *testing.T) {
	client := NewWeaviateClient(&WeaviateConfig{
		Scheme: "http",
		Host:   "localhost:8080",
	}, nil)
	ctx := context.Background()

	client.Connect(ctx)
	assert.True(t, client.IsConnected())

	client.CreateIndex(ctx, &IndexConfig{Name: "weaviate-test", Dimension: 3, Metric: "cosine"})

	has, _ := client.HasIndex(ctx, "weaviate-test")
	assert.True(t, has)

	client.Insert(ctx, "weaviate-test", []*Vector{{ID: "wv1", Data: []float32{0.1, 0.2, 0.3}}})
	client.Upsert(ctx, "weaviate-test", []*Vector{{ID: "wv2", Data: []float32{0.4, 0.5, 0.6}}})
	client.Delete(ctx, "weaviate-test", []string{"wv1"})
	client.Get(ctx, "weaviate-test", []string{"wv2"})
	client.Search(ctx, "weaviate-test", []float32{0.1, 0.2, 0.3}, nil)
	client.SearchBatch(ctx, "weaviate-test", [][]float32{{0.1, 0.2, 0.3}}, nil)
	client.Count(ctx, "weaviate-test")
	client.HealthCheck(ctx)
	client.DropIndex(ctx, "weaviate-test")
	client.Disconnect(ctx)
}

func TestVectorDBFactory(t *testing.T) {
	factory := NewVectorDBFactory(nil)

	milvus := factory.CreateMilvus(&MilvusConfig{Address: "localhost", Port: 19530})
	assert.NotNil(t, milvus)

	pinecone := factory.CreatePinecone(&PineconeConfig{APIKey: "key"})
	assert.NotNil(t, pinecone)

	weaviate := factory.CreateWeaviate(&WeaviateConfig{Host: "localhost:8080"})
	assert.NotNil(t, weaviate)
}

func TestDistanceCalculator(t *testing.T) {
	calc := NewDistanceCalculator()

	a := []float32{1.0, 0.0}
	b := []float32{0.0, 1.0}

	sim, err := calc.CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.LessOrEqual(t, sim, float32(0.01))

	dist, err := calc.EuclideanDistance(a, b)
	require.NoError(t, err)
	assert.Greater(t, dist, float32(1.0))

	dot, err := calc.DotProduct(a, b)
	require.NoError(t, err)
	assert.LessOrEqual(t, dot, float32(0.01))

	_, err = calc.CosineSimilarity([]float32{1.0}, []float32{1.0, 2.0})
	assert.ErrorIs(t, err, ErrInvalidVector)

	_, err = calc.EuclideanDistance([]float32{1.0}, []float32{1.0, 2.0})
	assert.ErrorIs(t, err, ErrInvalidVector)

	_, err = calc.DotProduct([]float32{1.0}, []float32{1.0, 2.0})
	assert.ErrorIs(t, err, ErrInvalidVector)
}

func TestDistanceCalculator_ZeroVectors(t *testing.T) {
	calc := NewDistanceCalculator()
	sim, err := calc.CosineSimilarity([]float32{0.0, 0.0}, []float32{1.0, 1.0})
	require.NoError(t, err)
	assert.Equal(t, float32(0), sim)
}

func TestVectorIndex_Operations(t *testing.T) {
	idx := NewVectorIndex("test-idx", 3, "cosine")

	err := idx.Insert(&Vector{ID: "vi-1", Data: []float32{1.0, 0.0, 0.0}})
	require.NoError(t, err)

	err = idx.Insert(&Vector{ID: "vi-2", Data: []float32{0.0, 1.0, 0.0}})
	require.NoError(t, err)

	assert.Equal(t, 2, idx.Count())

	err = idx.Insert(&Vector{ID: "vi-1", Data: []float32{0.5, 0.5, 0.0}})
	assert.ErrorIs(t, err, ErrDuplicateID)

	err = idx.Insert(&Vector{ID: "vi-3", Data: []float32{0.1, 0.2}})
	assert.ErrorIs(t, err, ErrInvalidVector)

	got, err := idx.Get("vi-1")
	require.NoError(t, err)
	assert.Equal(t, "vi-1", got.ID)

	_, err = idx.Get("nonexistent")
	assert.ErrorIs(t, err, ErrVectorNotFound)

	results, err := idx.Search([]float32{1.0, 0.0, 0.0}, 5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)

	_, err = idx.Search([]float32{0.1, 0.2}, 5)
	assert.ErrorIs(t, err, ErrInvalidVector)

	err = idx.Delete("vi-1")
	require.NoError(t, err)
	assert.Equal(t, 1, idx.Count())

	err = idx.Delete("nonexistent")
	assert.ErrorIs(t, err, ErrVectorNotFound)
}

func TestVectorIndex_Metrics(t *testing.T) {
	idx := NewVectorIndex("metric-idx", 2, "cosine")
	idx.Insert(&Vector{ID: "m1", Data: []float32{1.0, 0.0}})
	idx.Insert(&Vector{ID: "m2", Data: []float32{0.0, 1.0}})

	results, err := idx.Search([]float32{1.0, 0.0}, 5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)

	idxEuclidean := NewVectorIndex("euc-idx", 2, "euclidean")
	idxEuclidean.Insert(&Vector{ID: "e1", Data: []float32{1.0, 0.0}})
	idxEuclidean.Insert(&Vector{ID: "e2", Data: []float32{0.0, 1.0}})
	results, err = idxEuclidean.Search([]float32{1.0, 0.0}, 5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)

	idxDot := NewVectorIndex("dot-idx", 2, "dot_product")
	idxDot.Insert(&Vector{ID: "d1", Data: []float32{1.0, 0.0}})
	idxDot.Insert(&Vector{ID: "d2", Data: []float32{0.0, 1.0}})
	results, err = idxDot.Search([]float32{1.0, 0.0}, 5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestVectorSerializer(t *testing.T) {
	vs := NewVectorSerializer()

	v := &Vector{ID: "ser-1", Data: []float32{0.1, 0.2, 0.3}, Metadata: map[string]interface{}{"key": "val"}}

	data, err := vs.Serialize(v)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	got, err := vs.Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, "ser-1", got.ID)

	vectors := []*Vector{v, {ID: "ser-2", Data: []float32{0.4, 0.5}}}
	batchData, err := vs.SerializeBatch(vectors)
	require.NoError(t, err)
	assert.Len(t, batchData, 2)

	batchGot, err := vs.DeserializeBatch(batchData)
	require.NoError(t, err)
	assert.Len(t, batchGot, 2)
}

func TestEmbeddingStats_Struct(t *testing.T) {
	stats := &EmbeddingStats{
		TotalDocuments:   100,
		TotalChunks:      500,
		TotalEmbeddings:  500,
		CacheHits:        200,
		CacheMisses:      300,
		AverageChunkSize: 256,
	}
	assert.Equal(t, int64(100), stats.TotalDocuments)
	assert.Equal(t, int64(500), stats.TotalChunks)
}

func TestBatchEmbeddingRequest_Struct(t *testing.T) {
	req := &BatchEmbeddingRequest{
		Documents: []*Document{{ID: "d1", Content: "test"}},
		BatchSize: 10,
	}
	assert.Equal(t, 10, req.BatchSize)
}

func TestBatchEmbeddingResult_Struct(t *testing.T) {
	result := &BatchEmbeddingResult{
		DocumentID: "d1",
		Chunks:     []*TextChunk{{ID: "c1", Content: "chunk"}},
		Error:      "",
	}
	assert.Equal(t, "d1", result.DocumentID)
	assert.Empty(t, result.Error)
}

func TestKnowledgeBaseHealth_Struct(t *testing.T) {
	health := &KnowledgeBaseHealth{
		KnowledgeBaseID: "kb-1",
		Status:         "active",
		VectorDBStatus:  "healthy",
		DocumentCount:   10,
		VectorCount:     50,
		LastChecked:     time.Now(),
	}
	assert.Equal(t, "healthy", health.VectorDBStatus)
}

func TestKnowledgeManagerStats_Struct(t *testing.T) {
	stats := &KnowledgeManagerStats{
		TotalKnowledgeBases: 5,
		TotalDocuments:      100,
		TotalChunks:         500,
		ActiveKnowledgeBases: 3,
	}
	assert.Equal(t, int64(5), stats.TotalKnowledgeBases)
}

func TestRAGStats_Struct(t *testing.T) {
	stats := &RAGStats{
		TotalQueries:    100,
		TotalResults:    450,
		AverageDuration: 50 * time.Millisecond,
		AverageResults:  4.5,
		CacheHits:       30,
		CacheMisses:     70,
	}
	assert.Equal(t, int64(100), stats.TotalQueries)
}

func TestHybridSearchResult_Struct(t *testing.T) {
	result := &HybridSearchResult{
		VectorResults:   []*RetrievalResult{{Score: 0.9}},
		KeywordResults:  []*TextChunk{{ID: "kr-1"}},
		MergedResults:   []*RetrievalResult{{Score: 0.85}},
		VectorDuration:  10 * time.Millisecond,
		KeywordDuration: 5 * time.Millisecond,
	}
	assert.Len(t, result.VectorResults, 1)
}

func TestCitation_Struct(t *testing.T) {
	cit := &Citation{
		ChunkID:    "c-1",
		DocumentID: "d-1",
		Title:      "Test Doc",
		Source:     "upload",
		Position:   1,
		Content:    "Cited content",
		Score:      0.95,
	}
	assert.Equal(t, "c-1", cit.ChunkID)
	assert.Equal(t, float32(0.95), cit.Score)
}

func TestOpenAIEmbeddingConfig_Struct(t *testing.T) {
	cfg := &OpenAIEmbeddingConfig{
		APIKey:    "sk-xxx",
		Model:     "text-embedding-3-small",
		BaseURL:   "https://api.openai.com/v1",
		Timeout:   30 * time.Second,
		MaxRetry:  3,
		BatchSize: 100,
	}
	assert.Equal(t, "sk-xxx", cfg.APIKey)
	assert.Equal(t, 100, cfg.BatchSize)
}

func TestLocalEmbeddingConfig_Struct(t *testing.T) {
	cfg := &LocalEmbeddingConfig{
		ModelPath: "/models/embedding",
		ModelType: "sentence-transformers",
		Dimension: 768,
		MaxSeqLen: 512,
		BatchSize: 32,
	}
	assert.Equal(t, 768, cfg.Dimension)
}

func TestDocumentEmbedderConfig_Struct(t *testing.T) {
	cfg := &DocumentEmbedderConfig{
		Provider:    &coverageEmbeddingProvider{dimension: 3},
		ChunkConfig: DefaultChunkConfig(),
		CacheSize:   5000,
	}
	assert.Equal(t, 5000, cfg.CacheSize)
}

func TestKnowledgeBaseStatus_AllConstants(t *testing.T) {
	assert.Equal(t, KnowledgeBaseStatus("creating"), KnowledgeBaseStatusCreating)
	assert.Equal(t, KnowledgeBaseStatus("active"), KnowledgeBaseStatusActive)
	assert.Equal(t, KnowledgeBaseStatus("inactive"), KnowledgeBaseStatusInactive)
	assert.Equal(t, KnowledgeBaseStatus("deleting"), KnowledgeBaseStatusDeleting)
	assert.Equal(t, KnowledgeBaseStatus("error"), KnowledgeBaseStatusError)
}

func TestKnowledgeManager_ImportKnowledgeBase(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{
		Provider:    provider,
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, SplitBySentence: true},
		CacheSize:  100,
	}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	data := map[string]interface{}{
		"knowledge_base": map[string]interface{}{
			"name":        "Imported KB",
			"description": "Imported description",
			"dimension":   float64(3),
		},
		"documents": []interface{}{
			map[string]interface{}{
				"title":   "Imported Doc",
				"content": "Content of the imported document for testing.",
				"source":  "import",
			},
		},
	}

	kb, err := mgr.ImportKnowledgeBase(context.Background(), data)
	require.NoError(t, err)
	assert.Equal(t, "Imported KB", kb.Name)
}

func TestKnowledgeManager_ImportKnowledgeBase_InvalidData(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)
	mgr := NewKnowledgeManager(vdb, embedder, retriever, nil)

	_, err := mgr.ImportKnowledgeBase(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
}

func TestTokenize(t *testing.T) {
	tokens := tokenize("Hello, world! This is a test.")
	assert.Contains(t, tokens, "hello")
	assert.Contains(t, tokens, "world")
	assert.Contains(t, tokens, "this")
}

func TestRAGRetriever_Retrieve_EmptyQuery(t *testing.T) {
	vdb := newCoverageVectorDB()
	vdb.Connect(context.Background())
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)

	_, err := retriever.Retrieve(context.Background(), "", "idx")
	assert.ErrorIs(t, err, ErrInvalidQuery)
}

func TestRAGRetriever_Retrieve_NotConnected(t *testing.T) {
	vdb := newCoverageVectorDB()
	provider := &coverageEmbeddingProvider{dimension: 3, modelName: "test"}
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{Provider: provider}, nil)
	retriever := NewRAGRetriever(vdb, embedder, nil, nil)

	_, err := retriever.Retrieve(context.Background(), "test query", "idx")
	assert.ErrorIs(t, err, ErrVectorDBNotConnected)
}
