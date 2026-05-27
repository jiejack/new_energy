package knowledge

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVector_Struct(t *testing.T) {
	v := &Vector{
		ID:       "v1",
		Data:     []float32{0.1, 0.2, 0.3},
		Metadata: map[string]interface{}{"source": "test"},
	}
	assert.Equal(t, "v1", v.ID)
	assert.Equal(t, 3, len(v.Data))
}

func TestSearchResult_Struct(t *testing.T) {
	sr := &SearchResult{
		Vector:   &Vector{ID: "v1"},
		Score:    0.95,
		Distance: 0.05,
	}
	assert.Equal(t, float32(0.95), sr.Score)
}

func TestIndexConfig_Struct(t *testing.T) {
	cfg := &IndexConfig{
		Name:      "test_index",
		Dimension: 128,
		Metric:    "cosine",
		ShardNum:  2,
		ReplicaNum: 1,
		AutoCreate: true,
	}
	assert.Equal(t, "test_index", cfg.Name)
	assert.Equal(t, 128, cfg.Dimension)
}

func TestSearchParams_Struct(t *testing.T) {
	params := &SearchParams{
		TopK:         10,
		RoundDecimal: 4,
		NProbe:       8,
	}
	assert.Equal(t, 10, params.TopK)
}

func TestDefaultRAGConfig(t *testing.T) {
	cfg := DefaultRAGConfig()
	assert.Equal(t, 20, cfg.TopK)
	assert.Equal(t, 5, cfg.FinalTopK)
	assert.Equal(t, float32(0.5), cfg.ScoreThreshold)
	assert.True(t, cfg.UseReranker)
	assert.True(t, cfg.UseHybridSearch)
	assert.Equal(t, float32(0.7), cfg.VectorWeight)
	assert.Equal(t, float32(0.3), cfg.KeywordWeight)
	assert.Equal(t, 3, cfg.ContextWindowSize)
	assert.Equal(t, 4096, cfg.MaxContextLength)
}

func TestRetrievalResult_Struct(t *testing.T) {
	rr := &RetrievalResult{
		Chunk:       &TextChunk{ID: "c1"},
		Score:       0.9,
		VectorScore: 0.85,
		KeywordScore: 0.95,
		Rank:        1,
		Source:      "hybrid",
	}
	assert.Equal(t, float32(0.9), rr.Score)
	assert.Equal(t, "hybrid", rr.Source)
}

func TestRAGContext_Struct(t *testing.T) {
	rc := &RAGContext{
		Query:       "test query",
		Results:     []*RetrievalResult{{Score: 0.9}},
		Context:     "context text",
		TotalTokens: 100,
		Duration:    50 * time.Millisecond,
	}
	assert.Equal(t, "test query", rc.Query)
	assert.Equal(t, 100, rc.TotalTokens)
}

func TestDefaultChunkConfig(t *testing.T) {
	cfg := DefaultChunkConfig()
	assert.Equal(t, 512, cfg.ChunkSize)
	assert.Equal(t, 50, cfg.ChunkOverlap)
	assert.Equal(t, 100, cfg.MinChunkSize)
	assert.Equal(t, 1024, cfg.MaxChunkSize)
	assert.True(t, cfg.SplitBySentence)
	assert.False(t, cfg.SplitByParagraph)
}

func TestDocument_Struct(t *testing.T) {
	doc := &Document{
		ID:        "doc1",
		Content:   "Test content",
		Title:     "Test Document",
		Source:    "upload",
		Metadata:  map[string]interface{}{"author": "test"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	assert.Equal(t, "doc1", doc.ID)
	assert.Equal(t, "Test content", doc.Content)
}

func TestTextChunk_Struct(t *testing.T) {
	chunk := &TextChunk{
		ID:         "chunk1",
		DocumentID: "doc1",
		Content:    "chunk content",
		Position:   0,
		StartPos:   0,
		EndPos:     100,
	}
	assert.Equal(t, "chunk1", chunk.ID)
	assert.Equal(t, 0, chunk.Position)
}

func TestKnowledgeBaseStatus_Constants(t *testing.T) {
	assert.Equal(t, KnowledgeBaseStatus("creating"), KnowledgeBaseStatusCreating)
	assert.Equal(t, KnowledgeBaseStatus("active"), KnowledgeBaseStatusActive)
	assert.Equal(t, KnowledgeBaseStatus("inactive"), KnowledgeBaseStatusInactive)
	assert.Equal(t, KnowledgeBaseStatus("deleting"), KnowledgeBaseStatusDeleting)
	assert.Equal(t, KnowledgeBaseStatus("error"), KnowledgeBaseStatusError)
}

func TestKnowledgeBase_Struct(t *testing.T) {
	kb := &KnowledgeBase{
		ID:            "kb1",
		Name:          "Test KB",
		Description:   "A test knowledge base",
		IndexName:     "test_index",
		Dimension:     128,
		Status:        KnowledgeBaseStatusActive,
		DocumentCount: 10,
		ChunkCount:    50,
	}
	assert.Equal(t, "kb1", kb.ID)
	assert.Equal(t, KnowledgeBaseStatusActive, kb.Status)
}

func TestDocumentInfo_Struct(t *testing.T) {
	di := &DocumentInfo{
		ID:              "di1",
		KnowledgeBaseID: "kb1",
		Title:           "Test Doc",
		Source:          "upload",
		ContentType:     "text/plain",
		Size:            int64(1024),
		ChunkCount:      5,
		Status:          "indexed",
	}
	assert.Equal(t, "di1", di.ID)
	assert.Equal(t, 5, di.ChunkCount)
}

func TestCreateKnowledgeBaseRequest_Struct(t *testing.T) {
	req := &CreateKnowledgeBaseRequest{
		Name:        "New KB",
		Description: "A new knowledge base",
		Dimension:   256,
		Metric:      "cosine",
	}
	assert.Equal(t, "New KB", req.Name)
	assert.Equal(t, 256, req.Dimension)
}

func TestUpdateKnowledgeBaseRequest_Struct(t *testing.T) {
	req := &UpdateKnowledgeBaseRequest{
		Name:        "Updated KB",
		Description: "Updated description",
	}
	assert.Equal(t, "Updated KB", req.Name)
}

func TestErrors(t *testing.T) {
	assert.Error(t, ErrVectorDBNotConnected)
	assert.Error(t, ErrIndexNotFound)
	assert.Error(t, ErrInvalidVector)
	assert.Error(t, ErrDuplicateID)
	assert.Error(t, ErrVectorNotFound)
	assert.Error(t, ErrNoResults)
	assert.Error(t, ErrInvalidQuery)
	assert.Error(t, ErrRetrieverNotReady)
	assert.Error(t, ErrRerankerNotReady)
	assert.Error(t, ErrEmptyDocument)
	assert.Error(t, ErrInvalidChunkSize)
	assert.Error(t, ErrEmbeddingFailed)
	assert.Error(t, ErrProviderNotReady)
	assert.Error(t, ErrUnsupportedModel)
	assert.Error(t, ErrKnowledgeBaseNotFound)
	assert.Error(t, ErrDocumentNotFound)
	assert.Error(t, ErrKnowledgeBaseExists)
	assert.Error(t, ErrInvalidKnowledgeBaseID)
	assert.Error(t, ErrInvalidDocumentID)
}

func TestRAGConfig_Struct(t *testing.T) {
	cfg := &RAGConfig{
		TopK:              10,
		FinalTopK:         3,
		ScoreThreshold:    0.7,
		UseReranker:       false,
		UseHybridSearch:   true,
		VectorWeight:      0.8,
		KeywordWeight:     0.2,
		ContextWindowSize: 5,
		MaxContextLength:  2048,
	}
	assert.Equal(t, 10, cfg.TopK)
	assert.Equal(t, float32(0.8), cfg.VectorWeight)
}

func TestChunkConfig_Struct(t *testing.T) {
	cfg := &ChunkConfig{
		ChunkSize:        256,
		ChunkOverlap:     25,
		MinChunkSize:     50,
		MaxChunkSize:     512,
		SplitBySentence:  true,
		SplitByParagraph: true,
	}
	assert.Equal(t, 256, cfg.ChunkSize)
	assert.True(t, cfg.SplitByParagraph)
}

func TestEmbeddingProvider_Interface(t *testing.T) {
	var _ EmbeddingProvider = &mockEmbeddingProvider{}
}

type mockEmbeddingProvider struct{}

func (m *mockEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *mockEmbeddingProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range result {
		result[i] = []float32{0.1, 0.2, 0.3}
	}
	return result, nil
}

func (m *mockEmbeddingProvider) Dimension() int  { return 3 }
func (m *mockEmbeddingProvider) ModelName() string { return "mock" }

func TestVectorDB_Interface(t *testing.T) {
	var _ VectorDB = &mockVectorDB{}
}

type mockVectorDB struct {
	connected bool
	indexes   map[string]*IndexConfig
	vectors   map[string]map[string]*Vector
}

func newMockVectorDB() *mockVectorDB {
	return &mockVectorDB{
		indexes: make(map[string]*IndexConfig),
		vectors: make(map[string]map[string]*Vector),
	}
}

func (m *mockVectorDB) Connect(ctx context.Context) error {
	m.connected = true
	return nil
}

func (m *mockVectorDB) Disconnect(ctx context.Context) error {
	m.connected = false
	return nil
}

func (m *mockVectorDB) IsConnected() bool { return m.connected }

func (m *mockVectorDB) CreateIndex(ctx context.Context, config *IndexConfig) error {
	m.indexes[config.Name] = config
	m.vectors[config.Name] = make(map[string]*Vector)
	return nil
}

func (m *mockVectorDB) DropIndex(ctx context.Context, indexName string) error {
	delete(m.indexes, indexName)
	delete(m.vectors, indexName)
	return nil
}

func (m *mockVectorDB) HasIndex(ctx context.Context, indexName string) (bool, error) {
	_, ok := m.indexes[indexName]
	return ok, nil
}

func (m *mockVectorDB) ListIndexes(ctx context.Context) ([]string, error) {
	var names []string
	for name := range m.indexes {
		names = append(names, name)
	}
	return names, nil
}

func (m *mockVectorDB) DescribeIndex(ctx context.Context, indexName string) (*IndexConfig, error) {
	cfg, ok := m.indexes[indexName]
	if !ok {
		return nil, ErrIndexNotFound
	}
	return cfg, nil
}

func (m *mockVectorDB) Insert(ctx context.Context, indexName string, vectors []*Vector) error {
	if _, ok := m.vectors[indexName]; !ok {
		return ErrIndexNotFound
	}
	for _, v := range vectors {
		m.vectors[indexName][v.ID] = v
	}
	return nil
}

func (m *mockVectorDB) Upsert(ctx context.Context, indexName string, vectors []*Vector) error {
	return m.Insert(ctx, indexName, vectors)
}

func (m *mockVectorDB) Delete(ctx context.Context, indexName string, ids []string) error {
	if _, ok := m.vectors[indexName]; !ok {
		return ErrIndexNotFound
	}
	for _, id := range ids {
		delete(m.vectors[indexName], id)
	}
	return nil
}

func (m *mockVectorDB) Get(ctx context.Context, indexName string, ids []string) ([]*Vector, error) {
	if _, ok := m.vectors[indexName]; !ok {
		return nil, ErrIndexNotFound
	}
	var result []*Vector
	for _, id := range ids {
		if v, ok := m.vectors[indexName][id]; ok {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockVectorDB) Search(ctx context.Context, indexName string, vector []float32, params *SearchParams) ([]*SearchResult, error) {
	return nil, nil
}

func (m *mockVectorDB) SearchBatch(ctx context.Context, indexName string, vectors [][]float32, params *SearchParams) ([][]*SearchResult, error) {
	return nil, nil
}

func (m *mockVectorDB) Count(ctx context.Context, indexName string) (int64, error) {
	if _, ok := m.vectors[indexName]; !ok {
		return 0, ErrIndexNotFound
	}
	return int64(len(m.vectors[indexName])), nil
}

func (m *mockVectorDB) HealthCheck(ctx context.Context) error {
	if !m.connected {
		return ErrVectorDBNotConnected
	}
	return nil
}

func TestMockVectorDB(t *testing.T) {
	db := newMockVectorDB()
	ctx := context.Background()

	err := db.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, db.IsConnected())

	err = db.CreateIndex(ctx, &IndexConfig{Name: "test", Dimension: 3, Metric: "cosine"})
	require.NoError(t, err)

	has, err := db.HasIndex(ctx, "test")
	require.NoError(t, err)
	assert.True(t, has)

	indexes, err := db.ListIndexes(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(indexes))

	cfg, err := db.DescribeIndex(ctx, "test")
	require.NoError(t, err)
	assert.Equal(t, 3, cfg.Dimension)

	vectors := []*Vector{
		{ID: "v1", Data: []float32{0.1, 0.2, 0.3}},
		{ID: "v2", Data: []float32{0.4, 0.5, 0.6}},
	}
	err = db.Insert(ctx, "test", vectors)
	require.NoError(t, err)

	got, err := db.Get(ctx, "test", []string{"v1"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(got))

	err = db.Delete(ctx, "test", []string{"v1"})
	require.NoError(t, err)

	err = db.DropIndex(ctx, "test")
	require.NoError(t, err)

	err = db.Disconnect(ctx)
	require.NoError(t, err)
	assert.False(t, db.IsConnected())
}
