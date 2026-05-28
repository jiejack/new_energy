package knowledge

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTextSplitter_Split_Empty(t *testing.T) {
	splitter := NewTextSplitter(nil, nil)
	_, err := splitter.Split("")
	assert.Equal(t, ErrEmptyDocument, err)
}

func TestTextSplitter_Split_InvalidChunkSize(t *testing.T) {
	splitter := NewTextSplitter(&ChunkConfig{ChunkSize: 0}, nil)
	_, err := splitter.Split("hello world")
	assert.Equal(t, ErrInvalidChunkSize, err)
}

func TestTextSplitter_Split_BySentence(t *testing.T) {
	config := &ChunkConfig{
		ChunkSize:       100,
		ChunkOverlap:    10,
		MinChunkSize:    5,
		MaxChunkSize:    200,
		SplitBySentence: true,
	}
	splitter := NewTextSplitter(config, zap.NewNop())
	chunks, err := splitter.Split("This is sentence one. This is sentence two. This is sentence three.")
	require.NoError(t, err)
	assert.Greater(t, len(chunks), 0)
	for _, chunk := range chunks {
		assert.NotEmpty(t, chunk.ID)
	}
}

func TestTextSplitter_Split_ByParagraph(t *testing.T) {
	config := &ChunkConfig{
		ChunkSize:        100,
		ChunkOverlap:     10,
		MinChunkSize:     5,
		MaxChunkSize:     200,
		SplitByParagraph: true,
		SplitBySentence:  false,
	}
	splitter := NewTextSplitter(config, zap.NewNop())
	text := "First paragraph here.\n\nSecond paragraph here.\n\nThird paragraph here."
	chunks, err := splitter.Split(text)
	require.NoError(t, err)
	assert.Greater(t, len(chunks), 0)
}

func TestTextSplitter_Split_ByParagraph_LargeParagraph(t *testing.T) {
	config := &ChunkConfig{
		ChunkSize:        50,
		ChunkOverlap:     5,
		MinChunkSize:     5,
		MaxChunkSize:     50,
		SplitByParagraph: true,
		SplitBySentence:  true,
	}
	splitter := NewTextSplitter(config, zap.NewNop())
	text := "This is a very long paragraph that exceeds the maximum chunk size. It contains multiple sentences. Each sentence adds more content. The paragraph should be split further."
	chunks, err := splitter.Split(text)
	require.NoError(t, err)
	assert.Greater(t, len(chunks), 0)
}

func TestTextSplitter_Split_BySize(t *testing.T) {
	config := &ChunkConfig{
		ChunkSize:    20,
		ChunkOverlap: 5,
		MinChunkSize: 5,
		MaxChunkSize: 100,
	}
	splitter := NewTextSplitter(config, zap.NewNop())
	chunks, err := splitter.Split("This is a longer piece of text that should be split by size into multiple chunks.")
	require.NoError(t, err)
	assert.Greater(t, len(chunks), 0)
}

func TestTextSplitter_Split_DefaultConfig(t *testing.T) {
	config := &ChunkConfig{ChunkSize: 500, MinChunkSize: 5, MaxChunkSize: 1000, SplitBySentence: true}
	splitter := NewTextSplitter(config, nil)
	text := "This is a test sentence. This is another test sentence. And a third one for good measure."
	chunks, err := splitter.Split(text)
	require.NoError(t, err)
	assert.Greater(t, len(chunks), 0)
}

func TestTextSplitter_GenerateChunkID(t *testing.T) {
	splitter := NewTextSplitter(nil, nil)
	id1 := splitter.generateChunkID("content1", 0)
	id2 := splitter.generateChunkID("content1", 1)
	id3 := splitter.generateChunkID("content1", 0)
	assert.NotEqual(t, id1, id2)
	assert.Equal(t, id1, id3)
}

func TestGetLastNChars(t *testing.T) {
	assert.Equal(t, "llo", getLastNChars("hello", 3))
	assert.Equal(t, "hello", getLastNChars("hello", 10))
	assert.Equal(t, "", getLastNChars("", 3))
}

func TestEmbeddingCache_GetSet(t *testing.T) {
	cache := NewEmbeddingCache(100, nil)
	embedding := []float32{0.1, 0.2, 0.3}

	_, exists := cache.Get("test text")
	assert.False(t, exists)

	cache.Set("test text", embedding)
	got, exists := cache.Get("test text")
	assert.True(t, exists)
	assert.Equal(t, embedding, got)
}

func TestEmbeddingCache_Eviction(t *testing.T) {
	cache := NewEmbeddingCache(2, nil)
	cache.Set("text1", []float32{0.1})
	cache.Set("text2", []float32{0.2})
	cache.Set("text3", []float32{0.3})
	assert.LessOrEqual(t, cache.Size(), 2)
}

func TestEmbeddingCache_Clear(t *testing.T) {
	cache := NewEmbeddingCache(100, nil)
	cache.Set("text1", []float32{0.1})
	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}

func TestEmbeddingCache_Size(t *testing.T) {
	cache := NewEmbeddingCache(100, nil)
	assert.Equal(t, 0, cache.Size())
	cache.Set("text1", []float32{0.1})
	assert.Equal(t, 1, cache.Size())
}

func TestOpenAIEmbeddingProvider(t *testing.T) {
	config := &OpenAIEmbeddingConfig{APIKey: "test-key", Model: "text-embedding-ada-002"}
	provider := NewOpenAIEmbeddingProvider(config, nil)
	assert.Equal(t, 1536, provider.Dimension())
	assert.Equal(t, "text-embedding-ada-002", provider.ModelName())
}

func TestOpenAIEmbeddingProvider_Models(t *testing.T) {
	tests := []struct {
		model     string
		dimension int
	}{
		{"text-embedding-ada-002", 1536},
		{"text-embedding-3-small", 1536},
		{"text-embedding-3-large", 3072},
		{"unknown-model", 1536},
	}
	for _, tt := range tests {
		provider := NewOpenAIEmbeddingProvider(&OpenAIEmbeddingConfig{Model: tt.model}, nil)
		assert.Equal(t, tt.dimension, provider.Dimension(), "model: %s", tt.model)
	}
}

func TestOpenAIEmbeddingProvider_Embed(t *testing.T) {
	provider := NewOpenAIEmbeddingProvider(&OpenAIEmbeddingConfig{Model: "text-embedding-ada-002"}, nil)
	embedding, err := provider.Embed(context.Background(), "test text")
	require.NoError(t, err)
	assert.Equal(t, 1536, len(embedding))
}

func TestOpenAIEmbeddingProvider_Embed_Empty(t *testing.T) {
	provider := NewOpenAIEmbeddingProvider(&OpenAIEmbeddingConfig{Model: "text-embedding-ada-002"}, nil)
	_, err := provider.Embed(context.Background(), "")
	assert.Equal(t, ErrEmptyDocument, err)
}

func TestOpenAIEmbeddingProvider_EmbedBatch(t *testing.T) {
	provider := NewOpenAIEmbeddingProvider(&OpenAIEmbeddingConfig{Model: "text-embedding-ada-002"}, nil)
	embeddings, err := provider.EmbedBatch(context.Background(), []string{"text1", "text2"})
	require.NoError(t, err)
	assert.Equal(t, 2, len(embeddings))
}

func TestOpenAIEmbeddingProvider_EmbedBatch_Empty(t *testing.T) {
	provider := NewOpenAIEmbeddingProvider(&OpenAIEmbeddingConfig{Model: "text-embedding-ada-002"}, nil)
	_, err := provider.EmbedBatch(context.Background(), []string{})
	assert.Equal(t, ErrEmptyDocument, err)
}

func TestLocalEmbeddingProvider(t *testing.T) {
	config := &LocalEmbeddingConfig{ModelPath: "/tmp/model", ModelType: "sentence-transformers", Dimension: 384}
	provider := NewLocalEmbeddingProvider(config, nil)
	assert.Equal(t, 384, provider.Dimension())
	assert.Equal(t, "sentence-transformers", provider.ModelName())
	assert.False(t, provider.IsReady())
}

func TestLocalEmbeddingProvider_Initialize(t *testing.T) {
	provider := NewLocalEmbeddingProvider(&LocalEmbeddingConfig{Dimension: 384}, nil)
	err := provider.Initialize(context.Background())
	require.NoError(t, err)
	assert.True(t, provider.IsReady())
}

func TestLocalEmbeddingProvider_Embed_NotReady(t *testing.T) {
	provider := NewLocalEmbeddingProvider(&LocalEmbeddingConfig{Dimension: 384}, nil)
	_, err := provider.Embed(context.Background(), "test")
	assert.Equal(t, ErrProviderNotReady, err)
}

func TestLocalEmbeddingProvider_EmbedBatch_NotReady(t *testing.T) {
	provider := NewLocalEmbeddingProvider(&LocalEmbeddingConfig{Dimension: 384}, nil)
	_, err := provider.EmbedBatch(context.Background(), []string{"test"})
	assert.Equal(t, ErrProviderNotReady, err)
}

func TestLocalEmbeddingProvider_Embed_AfterInit(t *testing.T) {
	provider := NewLocalEmbeddingProvider(&LocalEmbeddingConfig{Dimension: 384}, nil)
	provider.Initialize(context.Background())
	embedding, err := provider.Embed(context.Background(), "test text")
	require.NoError(t, err)
	assert.Equal(t, 384, len(embedding))
}

func TestLocalEmbeddingProvider_EmbedBatch_AfterInit(t *testing.T) {
	provider := NewLocalEmbeddingProvider(&LocalEmbeddingConfig{Dimension: 384}, nil)
	provider.Initialize(context.Background())
	embeddings, err := provider.EmbedBatch(context.Background(), []string{"text1", "text2"})
	require.NoError(t, err)
	assert.Equal(t, 2, len(embeddings))
}

func TestLocalEmbeddingProvider_EmbedBatch_Empty(t *testing.T) {
	provider := NewLocalEmbeddingProvider(&LocalEmbeddingConfig{Dimension: 384}, nil)
	provider.Initialize(context.Background())
	_, err := provider.EmbedBatch(context.Background(), []string{})
	assert.Equal(t, ErrEmptyDocument, err)
}

func TestDocumentEmbedder_EmbedDocument(t *testing.T) {
	config := &DocumentEmbedderConfig{
		Provider:    &mockEmbeddingProvider{},
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, MaxChunkSize: 200, SplitBySentence: true},
		CacheSize:   100,
	}
	embedder := NewDocumentEmbedder(config, zap.NewNop())

	doc := &Document{
		ID:       "doc1",
		Content:  "This is a test document. It has multiple sentences. Each sentence adds content.",
		Title:    "Test Doc",
		Source:   "test",
		Metadata: map[string]interface{}{"author": "tester"},
	}

	chunks, err := embedder.EmbedDocument(context.Background(), doc)
	require.NoError(t, err)
	assert.Greater(t, len(chunks), 0)
	assert.Equal(t, "doc1", chunks[0].DocumentID)
	assert.Equal(t, "Test Doc", chunks[0].Metadata["title"])
}

func TestDocumentEmbedder_EmbedDocument_Empty(t *testing.T) {
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)

	_, err := embedder.EmbedDocument(context.Background(), &Document{ID: "doc1"})
	assert.Equal(t, ErrEmptyDocument, err)

	_, err = embedder.EmbedDocument(context.Background(), nil)
	assert.Equal(t, ErrEmptyDocument, err)
}

func TestDocumentEmbedder_Embed(t *testing.T) {
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)

	embedding, err := embedder.Embed(context.Background(), "test text")
	require.NoError(t, err)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, embedding)

	embedding2, err := embedder.Embed(context.Background(), "test text")
	require.NoError(t, err)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, embedding2)
}

func TestDocumentEmbedder_EmbedBatch(t *testing.T) {
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)

	embeddings, err := embedder.EmbedBatch(context.Background(), []string{"text1", "text2"})
	require.NoError(t, err)
	assert.Equal(t, 2, len(embeddings))
}

func TestDocumentEmbedder_GetCacheStats(t *testing.T) {
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	embedder.Embed(context.Background(), "test")

	stats := embedder.GetCacheStats()
	assert.Equal(t, 1, stats["cache_size"])
	assert.Equal(t, 3, stats["dimension"])
	assert.Equal(t, "mock", stats["model"])
}

func TestDocumentEmbedder_ClearCache(t *testing.T) {
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	embedder.Embed(context.Background(), "test")
	embedder.ClearCache()
	stats := embedder.GetCacheStats()
	assert.Equal(t, 0, stats["cache_size"])
}

func TestDocumentEmbedder_DefaultConfig(t *testing.T) {
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}}
	embedder := NewDocumentEmbedder(config, nil)
	assert.NotNil(t, embedder)
}

func TestEmbeddingService_ProcessDocument(t *testing.T) {
	config := &DocumentEmbedderConfig{
		Provider:    &mockEmbeddingProvider{},
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, MaxChunkSize: 200, SplitBySentence: true},
		CacheSize:   100,
	}
	embedder := NewDocumentEmbedder(config, zap.NewNop())
	service := NewEmbeddingService(embedder, nil)

	doc := &Document{ID: "doc1", Content: "This is a test document. It has multiple sentences."}
	chunks, err := service.ProcessDocument(context.Background(), doc)
	require.NoError(t, err)
	assert.Greater(t, len(chunks), 0)

	stats := service.GetStats()
	assert.Equal(t, int64(1), stats.TotalDocuments)
	assert.Greater(t, stats.TotalChunks, int64(0))
}

func TestEmbeddingService_BatchEmbedDocuments(t *testing.T) {
	config := &DocumentEmbedderConfig{
		Provider:    &mockEmbeddingProvider{},
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, MaxChunkSize: 200, SplitBySentence: true},
		CacheSize:   100,
	}
	embedder := NewDocumentEmbedder(config, zap.NewNop())
	service := NewEmbeddingService(embedder, nil)

	req := &BatchEmbeddingRequest{
		Documents: []*Document{
			{ID: "doc1", Content: "First document content here."},
			{ID: "doc2", Content: "Second document content here."},
		},
	}

	results := service.BatchEmbedDocuments(context.Background(), req)
	assert.Equal(t, 2, len(results))
	assert.Empty(t, results[0].Error)
	assert.Greater(t, len(results[0].Chunks), 0)
}

func TestEmbeddingService_BatchEmbedDocuments_WithError(t *testing.T) {
	config := &DocumentEmbedderConfig{
		Provider:    &mockEmbeddingProvider{},
		ChunkConfig: &ChunkConfig{ChunkSize: 100, MinChunkSize: 5, MaxChunkSize: 200, SplitBySentence: true},
		CacheSize:   100,
	}
	embedder := NewDocumentEmbedder(config, zap.NewNop())
	service := NewEmbeddingService(embedder, nil)

	req := &BatchEmbeddingRequest{
		Documents: []*Document{
			{ID: "doc1", Content: "Valid content here."},
			{ID: "doc2"},
		},
	}

	results := service.BatchEmbedDocuments(context.Background(), req)
	assert.Equal(t, 2, len(results))
	assert.Empty(t, results[0].Error)
	assert.NotEmpty(t, results[1].Error)
}

func TestSimilarityCalculator_CosineSimilarity(t *testing.T) {
	calc := NewSimilarityCalculator()
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{0.0, 1.0, 0.0}
	sim, err := calc.CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.Less(t, sim, float32(0.01))

	c := []float32{1.0, 0.0, 0.0}
	d := []float32{1.0, 0.0, 0.0}
	sim2, err := calc.CosineSimilarity(c, d)
	require.NoError(t, err)
	assert.Greater(t, sim2, float32(0.99))
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
	assert.Equal(t, 2, len(indices))
	assert.Equal(t, 0, indices[0])
	assert.Greater(t, values[0], values[1])
}

func TestSimilarityCalculator_FindMostSimilar_Empty(t *testing.T) {
	calc := NewSimilarityCalculator()
	indices, values, err := calc.FindMostSimilar([]float32{1.0}, nil, 5)
	assert.NoError(t, err)
	assert.Nil(t, indices)
	assert.Nil(t, values)
}

func TestSimilarityCalculator_FindMostSimilar_TopKLarger(t *testing.T) {
	calc := NewSimilarityCalculator()
	query := []float32{1.0, 0.0, 0.0}
	vectors := [][]float32{{0.9, 0.1, 0.0}}

	indices, _, err := calc.FindMostSimilar(query, vectors, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, len(indices))
}

func TestSimpleKeywordSearcher_Search(t *testing.T) {
	searcher := NewSimpleKeywordSearcher(nil)
	ctx := context.Background()

	chunks := []*TextChunk{
		{ID: "c1", Content: "solar panel installation guide"},
		{ID: "c2", Content: "wind turbine maintenance manual"},
		{ID: "c3", Content: "solar energy storage system"},
	}
	searcher.Index(ctx, chunks)

	results, err := searcher.Search(ctx, "solar panel", 10)
	require.NoError(t, err)
	assert.Greater(t, len(results), 0)
}

func TestSimpleKeywordSearcher_Search_EmptyQuery(t *testing.T) {
	searcher := NewSimpleKeywordSearcher(nil)
	_, err := searcher.Search(context.Background(), "", 10)
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestSimpleKeywordSearcher_Search_NoResults(t *testing.T) {
	searcher := NewSimpleKeywordSearcher(nil)
	searcher.Index(context.Background(), []*TextChunk{{ID: "c1", Content: "hello world"}})
	results, err := searcher.Search(context.Background(), "nonexistent query terms", 10)
	require.NoError(t, err)
	assert.Equal(t, 0, len(results))
}

func TestSimpleKeywordSearcher_Delete(t *testing.T) {
	searcher := NewSimpleKeywordSearcher(nil)
	ctx := context.Background()

	chunks := []*TextChunk{{ID: "c1", Content: "solar panel guide"}}
	searcher.Index(ctx, chunks)

	err := searcher.Delete(ctx, []string{"c1"})
	require.NoError(t, err)

	results, _ := searcher.Search(ctx, "solar panel", 10)
	assert.Equal(t, 0, len(results))
}

func TestSimpleKeywordSearcher_Delete_NonExistent(t *testing.T) {
	searcher := NewSimpleKeywordSearcher(nil)
	err := searcher.Delete(context.Background(), []string{"nonexistent"})
	assert.NoError(t, err)
}

func TestTokenize(t *testing.T) {
	tokens := tokenize("Hello, World! This is a test.")
	assert.Contains(t, tokens, "hello")
	assert.Contains(t, tokens, "world")
}

func TestTokenize_Chinese(t *testing.T) {
	tokens := tokenize("你好，世界！这是一个测试。")
	assert.Greater(t, len(tokens), 0)
}

func TestSimpleReranker_Rerank(t *testing.T) {
	reranker := NewSimpleReranker(nil)
	ctx := context.Background()

	results := []*RetrievalResult{
		{Chunk: &TextChunk{Content: "solar panel installation"}, Score: 0.5},
		{Chunk: &TextChunk{Content: "wind turbine guide"}, Score: 0.8},
	}

	reranked, err := reranker.Rerank(ctx, "solar panel", results)
	require.NoError(t, err)
	assert.Equal(t, 2, len(reranked))
	assert.Equal(t, 1, reranked[0].Rank)
}

func TestSimpleReranker_Rerank_Empty(t *testing.T) {
	reranker := NewSimpleReranker(nil)
	results, err := reranker.Rerank(context.Background(), "query", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestContextWindowManager_BuildContext(t *testing.T) {
	mgr := NewContextWindowManager(1000, nil)
	results := []*RetrievalResult{
		{Chunk: &TextChunk{Content: "First chunk content"}},
		{Chunk: &TextChunk{Content: "Second chunk content"}},
	}
	ctx := mgr.BuildContext(results)
	assert.Contains(t, ctx, "First chunk content")
	assert.Contains(t, ctx, "Second chunk content")
}

func TestContextWindowManager_BuildContext_Truncation(t *testing.T) {
	mgr := NewContextWindowManager(200, nil)
	results := []*RetrievalResult{
		{Chunk: &TextChunk{Content: "This is a very long chunk that should be truncated because it exceeds the maximum context length and needs to be cut off"}},
		{Chunk: &TextChunk{Content: "Short content"}},
	}
	ctx := mgr.BuildContext(results)
	assert.NotEmpty(t, ctx)
}

func TestContextWindowManager_BuildContext_SmallRemaining(t *testing.T) {
	mgr := NewContextWindowManager(30, nil)
	results := []*RetrievalResult{
		{Chunk: &TextChunk{Content: "First chunk with some content"}},
		{Chunk: &TextChunk{Content: "Second chunk content here"}},
	}
	ctx := mgr.BuildContext(results)
	assert.NotEmpty(t, ctx)
}

func TestCitationTracker_ExtractCitations(t *testing.T) {
	tracker := NewCitationTracker(nil)
	results := []*RetrievalResult{
		{
			Chunk: &TextChunk{ID: "c1", DocumentID: "doc1", Content: "content1", Metadata: map[string]interface{}{"title": "Title1", "source": "upload"}},
			Score: 0.9,
		},
		{
			Chunk: &TextChunk{ID: "c2", DocumentID: "doc2", Content: "content2"},
			Score: 0.7,
		},
	}

	citations := tracker.ExtractCitations(results)
	assert.Equal(t, 2, len(citations))
	assert.Equal(t, "Title1", citations[0].Title)
	assert.Equal(t, "upload", citations[0].Source)
}

func TestRAGRetriever_Retrieve_EmptyQuery(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)

	_, err := retriever.Retrieve(context.Background(), "", "test")
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestRAGRetriever_Retrieve_NotConnected(t *testing.T) {
	db := newMockVectorDB()
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)

	_, err := retriever.Retrieve(context.Background(), "test query", "test")
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestRAGRetriever_IndexChunks(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	db.CreateIndex(context.Background(), &IndexConfig{Name: "test", Dimension: 3})

	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)

	chunks := []*TextChunk{
		{ID: "c1", DocumentID: "doc1", Content: "test content", Embedding: []float32{0.1, 0.2, 0.3}},
	}

	err := retriever.IndexChunks(context.Background(), "test", chunks)
	require.NoError(t, err)
}

func TestRAGRetriever_DeleteChunks(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	db.CreateIndex(context.Background(), &IndexConfig{Name: "test", Dimension: 3})

	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)

	err := retriever.DeleteChunks(context.Background(), "test", []string{"c1"})
	require.NoError(t, err)
}

func TestRAGRetriever_GetCitations(t *testing.T) {
	db := newMockVectorDB()
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)

	results := []*RetrievalResult{
		{Chunk: &TextChunk{ID: "c1", DocumentID: "doc1", Content: "content"}, Score: 0.9},
	}

	citations := retriever.GetCitations(results)
	assert.Equal(t, 1, len(citations))
}

func TestRAGService_GetStats(t *testing.T) {
	db := newMockVectorDB()
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	service := NewRAGService(retriever, nil)

	stats := service.GetStats()
	assert.Equal(t, int64(0), stats.TotalQueries)
}

func TestRAGService_HybridSearch(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	db.CreateIndex(context.Background(), &IndexConfig{Name: "test", Dimension: 3})

	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	service := NewRAGService(retriever, nil)

	result, err := service.HybridSearch(context.Background(), "test query", "test")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMultiQueryRetriever_EmptyQueries(t *testing.T) {
	db := newMockVectorDB()
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mqr := NewMultiQueryRetriever(retriever, nil)

	_, err := mqr.Retrieve(context.Background(), nil, "test")
	assert.Equal(t, ErrInvalidQuery, err)
}

func TestDistanceCalculator_CosineSimilarity(t *testing.T) {
	calc := NewDistanceCalculator()
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{1.0, 0.0, 0.0}
	sim, err := calc.CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.Greater(t, sim, float32(0.99))
}

func TestDistanceCalculator_CosineSimilarity_DifferentLengths(t *testing.T) {
	calc := NewDistanceCalculator()
	_, err := calc.CosineSimilarity([]float32{1.0, 0.0}, []float32{1.0, 0.0, 0.0})
	assert.Error(t, err)
}

func TestDistanceCalculator_CosineSimilarity_ZeroVector(t *testing.T) {
	calc := NewDistanceCalculator()
	sim, err := calc.CosineSimilarity([]float32{0.0, 0.0, 0.0}, []float32{1.0, 0.0, 0.0})
	require.NoError(t, err)
	assert.Equal(t, float32(0.0), sim)
}

func TestDistanceCalculator_EuclideanDistance(t *testing.T) {
	calc := NewDistanceCalculator()
	dist, err := calc.EuclideanDistance([]float32{0.0, 0.0}, []float32{3.0, 4.0})
	require.NoError(t, err)
	assert.Equal(t, float32(5.0), dist)
}

func TestDistanceCalculator_EuclideanDistance_DifferentLengths(t *testing.T) {
	calc := NewDistanceCalculator()
	_, err := calc.EuclideanDistance([]float32{1.0}, []float32{1.0, 2.0})
	assert.Error(t, err)
}

func TestDistanceCalculator_DotProduct(t *testing.T) {
	calc := NewDistanceCalculator()
	dp, err := calc.DotProduct([]float32{1.0, 2.0, 3.0}, []float32{4.0, 5.0, 6.0})
	require.NoError(t, err)
	assert.Equal(t, float32(32.0), dp)
}

func TestDistanceCalculator_DotProduct_DifferentLengths(t *testing.T) {
	calc := NewDistanceCalculator()
	_, err := calc.DotProduct([]float32{1.0}, []float32{1.0, 2.0})
	assert.Error(t, err)
}

func TestVectorIndex(t *testing.T) {
	vi := NewVectorIndex("test", 3, "cosine")

	v1 := &Vector{ID: "v1", Data: []float32{1.0, 0.0, 0.0}}
	v2 := &Vector{ID: "v2", Data: []float32{0.0, 1.0, 0.0}}
	v3 := &Vector{ID: "v3", Data: []float32{0.9, 0.1, 0.0}}

	err := vi.Insert(v1)
	require.NoError(t, err)
	err = vi.Insert(v2)
	require.NoError(t, err)
	err = vi.Insert(v3)
	require.NoError(t, err)

	assert.Equal(t, 3, vi.Count())

	got, err := vi.Get("v1")
	require.NoError(t, err)
	assert.Equal(t, "v1", got.ID)

	results, err := vi.Search([]float32{1.0, 0.0, 0.0}, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "v1", results[0].Vector.ID)

	err = vi.Delete("v1")
	require.NoError(t, err)
	assert.Equal(t, 2, vi.Count())
}

func TestVectorIndex_DuplicateInsert(t *testing.T) {
	vi := NewVectorIndex("test", 3, "cosine")
	vi.Insert(&Vector{ID: "v1", Data: []float32{1.0, 0.0, 0.0}})
	err := vi.Insert(&Vector{ID: "v1", Data: []float32{0.5, 0.5, 0.0}})
	assert.Equal(t, ErrDuplicateID, err)
}

func TestVectorIndex_Get_NotFound(t *testing.T) {
	vi := NewVectorIndex("test", 3, "cosine")
	_, err := vi.Get("nonexistent")
	assert.Equal(t, ErrVectorNotFound, err)
}

func TestVectorIndex_Delete_NotFound(t *testing.T) {
	vi := NewVectorIndex("test", 3, "cosine")
	err := vi.Delete("nonexistent")
	assert.Equal(t, ErrVectorNotFound, err)
}

func TestVectorIndex_Search_Empty(t *testing.T) {
	vi := NewVectorIndex("test", 3, "cosine")
	results, err := vi.Search([]float32{1.0, 0.0, 0.0}, 5)
	require.NoError(t, err)
	assert.Equal(t, 0, len(results))
}

func TestVectorSerializer(t *testing.T) {
	vs := NewVectorSerializer()

	v := &Vector{ID: "v1", Data: []float32{0.1, 0.2, 0.3}, Metadata: map[string]interface{}{"key": "value"}}
	data, err := vs.Serialize(v)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	got, err := vs.Deserialize(data)
	require.NoError(t, err)
	assert.Equal(t, "v1", got.ID)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, got.Data)
}

func TestVectorSerializer_Batch(t *testing.T) {
	vs := NewVectorSerializer()

	vectors := []*Vector{
		{ID: "v1", Data: []float32{0.1, 0.2}},
		{ID: "v2", Data: []float32{0.3, 0.4}},
	}

	data, err := vs.SerializeBatch(vectors)
	require.NoError(t, err)
	assert.Equal(t, 2, len(data))

	got, err := vs.DeserializeBatch(data)
	require.NoError(t, err)
	assert.Equal(t, 2, len(got))
	assert.Equal(t, "v1", got[0].ID)
	assert.Equal(t, "v2", got[1].ID)
}

func TestVectorDBFactory(t *testing.T) {
	factory := NewVectorDBFactory(zap.NewNop())

	milvus := factory.CreateMilvus(&MilvusConfig{Address: "localhost:19530"})
	assert.NotNil(t, milvus)

	pinecone := factory.CreatePinecone(&PineconeConfig{APIKey: "test", Environment: "test"})
	assert.NotNil(t, pinecone)

	weaviate := factory.CreateWeaviate(&WeaviateConfig{Host: "localhost", Scheme: "http"})
	assert.NotNil(t, weaviate)
}

func TestMilvusClient(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost:19530"}, nil)
	assert.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

func TestPineconeClient(t *testing.T) {
	client := NewPineconeClient(&PineconeConfig{APIKey: "test", Environment: "test"}, nil)
	assert.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

func TestWeaviateClient(t *testing.T) {
	client := NewWeaviateClient(&WeaviateConfig{Host: "localhost", Scheme: "http"}, nil)
	assert.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

func TestInMemoryDocumentStore(t *testing.T) {
	store := NewInMemoryDocumentStore(zap.NewNop())
	ctx := context.Background()

	doc := &Document{
		ID:       "doc1",
		Content:  "Test content",
		Title:    "Test Doc",
		Source:   "upload",
		Metadata: map[string]interface{}{"knowledge_base_id": "kb1"},
	}

	err := store.Save(ctx, doc)
	require.NoError(t, err)

	got, err := store.Get(ctx, "doc1")
	require.NoError(t, err)
	assert.Equal(t, "doc1", got.ID)

	docs, err := store.List(ctx, "kb1", 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, len(docs))

	count, err := store.Count(ctx, "kb1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	err = store.Delete(ctx, "doc1")
	require.NoError(t, err)

	_, err = store.Get(ctx, "doc1")
	assert.Error(t, err)
}

func TestInMemoryDocumentStore_Get_NotFound(t *testing.T) {
	store := NewInMemoryDocumentStore(nil)
	_, err := store.Get(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestInMemoryDocumentStore_Delete_NotFound(t *testing.T) {
	store := NewInMemoryDocumentStore(nil)
	err := store.Delete(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestKnowledgeManager_CreateKnowledgeBase(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, zap.NewNop())

	req := &CreateKnowledgeBaseRequest{
		Name:        "Test KB",
		Description: "A test knowledge base",
		Dimension:   128,
		Metric:      "cosine",
	}

	kb, err := mgr.CreateKnowledgeBase(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, kb.ID)
	assert.Equal(t, "Test KB", kb.Name)
	assert.Equal(t, KnowledgeBaseStatusActive, kb.Status)
}

func TestKnowledgeManager_GetKnowledgeBase(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 128, Metric: "cosine"})
	got, err := mgr.GetKnowledgeBase(context.Background(), kb.ID)
	require.NoError(t, err)
	assert.Equal(t, kb.ID, got.ID)
}

func TestKnowledgeManager_GetKnowledgeBase_NotFound(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	_, err := mgr.GetKnowledgeBase(context.Background(), "nonexistent")
	assert.Equal(t, ErrKnowledgeBaseNotFound, err)
}

func TestKnowledgeManager_ListKnowledgeBases(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "KB1", Dimension: 128, Metric: "cosine"})
	mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "KB2", Dimension: 128, Metric: "cosine"})

	kbs, err := mgr.ListKnowledgeBases(context.Background(), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, len(kbs))
}

func TestKnowledgeManager_UpdateKnowledgeBase(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Original", Dimension: 128, Metric: "cosine"})
	updated, err := mgr.UpdateKnowledgeBase(context.Background(), kb.ID, &UpdateKnowledgeBaseRequest{Name: "Updated", Description: "New desc"})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, "New desc", updated.Description)
}

func TestKnowledgeManager_UpdateKnowledgeBase_NotFound(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	_, err := mgr.UpdateKnowledgeBase(context.Background(), "nonexistent", &UpdateKnowledgeBaseRequest{Name: "Test"})
	assert.Equal(t, ErrKnowledgeBaseNotFound, err)
}

func TestKnowledgeManager_DeleteKnowledgeBase(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 128, Metric: "cosine"})
	err := mgr.DeleteKnowledgeBase(context.Background(), kb.ID)
	require.NoError(t, err)

	_, err = mgr.GetKnowledgeBase(context.Background(), kb.ID)
	assert.Equal(t, ErrKnowledgeBaseNotFound, err)
}

func TestKnowledgeManager_DeleteKnowledgeBase_NotFound(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	err := mgr.DeleteKnowledgeBase(context.Background(), "nonexistent")
	assert.Equal(t, ErrKnowledgeBaseNotFound, err)
}

func TestKnowledgeManager_UploadDocument(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 3, Metric: "cosine"})

	req := &UploadDocumentRequest{
		KnowledgeBaseID: kb.ID,
		Title:           "Test Doc",
		Content:         "This is a test document. It has multiple sentences for testing.",
		Source:          "upload",
	}

	docInfo, err := mgr.UploadDocument(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, docInfo.ID)
	assert.Equal(t, "Test Doc", docInfo.Title)
}

func TestKnowledgeManager_UploadDocument_KBNotFound(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	_, err := mgr.UploadDocument(context.Background(), &UploadDocumentRequest{KnowledgeBaseID: "nonexistent", Content: "test"})
	assert.Equal(t, ErrKnowledgeBaseNotFound, err)
}

func TestKnowledgeManager_DeleteDocument(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 3, Metric: "cosine"})
	doc, _ := mgr.UploadDocument(context.Background(), &UploadDocumentRequest{KnowledgeBaseID: kb.ID, Content: "Test content for document.", Title: "Test"})

	err := mgr.DeleteDocument(context.Background(), kb.ID, doc.ID)
	require.NoError(t, err)
}

func TestKnowledgeManager_DeleteDocument_KBNotFound(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	err := mgr.DeleteDocument(context.Background(), "nonexistent", "doc1")
	assert.Equal(t, ErrKnowledgeBaseNotFound, err)
}

func TestKnowledgeManager_SearchKnowledgeBase(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 3, Metric: "cosine"})

	result, err := mgr.SearchKnowledgeBase(context.Background(), kb.ID, "test query", 5)
	if err != nil {
		assert.Contains(t, err.Error(), "no results")
	} else {
		assert.NotNil(t, result)
	}
}

func TestKnowledgeManager_SearchKnowledgeBase_KBNotFound(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	_, err := mgr.SearchKnowledgeBase(context.Background(), "nonexistent", "test", 5)
	assert.Equal(t, ErrKnowledgeBaseNotFound, err)
}

func TestKnowledgeManager_GetKnowledgeBaseStats(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 3, Metric: "cosine"})
	stats, err := mgr.GetKnowledgeBaseStats(context.Background(), kb.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.DocumentCount)
}

func TestKnowledgeManager_GetKnowledgeBaseStats_NotFound(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	_, err := mgr.GetKnowledgeBaseStats(context.Background(), "nonexistent")
	assert.Equal(t, ErrKnowledgeBaseNotFound, err)
}

func TestKnowledgeManager_GetStats(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 3, Metric: "cosine"})
	stats := mgr.GetStats()
	assert.Equal(t, int64(1), stats.TotalKnowledgeBases)
}

func TestKnowledgeManager_CheckHealth(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 3, Metric: "cosine"})
	health, err := mgr.CheckHealth(context.Background(), kb.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", health.Status)
}

func TestKnowledgeManager_BatchUploadDocuments(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 3, Metric: "cosine"})

	reqs := []*UploadDocumentRequest{
		{KnowledgeBaseID: kb.ID, Title: "Doc1", Content: "First document content.", Source: "upload"},
		{KnowledgeBaseID: kb.ID, Title: "Doc2", Content: "Second document content.", Source: "upload"},
	}

	docs, errs := mgr.BatchUploadDocuments(context.Background(), reqs)
	assert.Equal(t, 2, len(docs))
	assert.Equal(t, 2, len(errs))
	assert.Nil(t, errs[0])
	assert.Nil(t, errs[1])
}

func TestKnowledgeManager_ListDocuments(t *testing.T) {
	db := newMockVectorDB()
	db.Connect(context.Background())
	config := &DocumentEmbedderConfig{Provider: &mockEmbeddingProvider{}, CacheSize: 100}
	embedder := NewDocumentEmbedder(config, nil)
	retriever := NewRAGRetriever(db, embedder, nil, nil)
	mgr := NewKnowledgeManager(db, embedder, retriever, nil)

	kb, _ := mgr.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{Name: "Test", Dimension: 3, Metric: "cosine"})
	mgr.UploadDocument(context.Background(), &UploadDocumentRequest{KnowledgeBaseID: kb.ID, Content: "Test content.", Title: "Doc1"})
	mgr.UploadDocument(context.Background(), &UploadDocumentRequest{KnowledgeBaseID: kb.ID, Content: "More content.", Title: "Doc2"})

	docs, err := mgr.ListDocuments(context.Background(), kb.ID, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, len(docs))
}
