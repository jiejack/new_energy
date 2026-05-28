package knowledge

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCov_MilvusClient_Connect(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	err := client.Connect(context.Background())
	assert.NoError(t, err)
	assert.True(t, client.IsConnected())
	err = client.Connect(context.Background())
	assert.NoError(t, err)
}

func TestCov_MilvusClient_Disconnect(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	err := client.Disconnect(context.Background())
	assert.NoError(t, err)
	client.Connect(context.Background())
	err = client.Disconnect(context.Background())
	assert.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestCov_MilvusClient_CreateIndex(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	err := client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 128, Metric: "cosine"})
	assert.NoError(t, err)
	err = client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 128, Metric: "cosine"})
	assert.Error(t, err)
	err = client.CreateIndex(context.Background(), &IndexConfig{Name: "bad_idx", Dimension: 0, Metric: "cosine"})
	assert.Error(t, err)
}

func TestCov_MilvusClient_CreateIndex_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	err := client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 128, Metric: "cosine"})
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_DropIndex(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 128, Metric: "cosine"})
	err := client.DropIndex(context.Background(), "test_idx")
	assert.NoError(t, err)
	err = client.DropIndex(context.Background(), "nonexistent")
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestCov_MilvusClient_DropIndex_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	err := client.DropIndex(context.Background(), "test_idx")
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_HasIndex(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 128, Metric: "cosine"})
	exists, err := client.HasIndex(context.Background(), "test_idx")
	assert.NoError(t, err)
	assert.True(t, exists)
	exists, err = client.HasIndex(context.Background(), "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestCov_MilvusClient_HasIndex_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	_, err := client.HasIndex(context.Background(), "test_idx")
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_ListIndexes(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "idx1", Dimension: 128, Metric: "cosine"})
	client.CreateIndex(context.Background(), &IndexConfig{Name: "idx2", Dimension: 64, Metric: "l2"})
	indexes, err := client.ListIndexes(context.Background())
	assert.NoError(t, err)
	assert.Len(t, indexes, 2)
}

func TestCov_MilvusClient_ListIndexes_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	_, err := client.ListIndexes(context.Background())
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_DescribeIndex(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 128, Metric: "cosine"})
	config, err := client.DescribeIndex(context.Background(), "test_idx")
	assert.NoError(t, err)
	assert.Equal(t, 128, config.Dimension)
	_, err = client.DescribeIndex(context.Background(), "nonexistent")
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestCov_MilvusClient_DescribeIndex_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	_, err := client.DescribeIndex(context.Background(), "test_idx")
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_Insert(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 3, Metric: "cosine"})
	err := client.Insert(context.Background(), "test_idx", []*Vector{{ID: "v1", Data: []float32{0.1, 0.2, 0.3}}})
	assert.NoError(t, err)
	err = client.Insert(context.Background(), "test_idx", []*Vector{{ID: "v2", Data: []float32{0.1, 0.2}}})
	assert.Error(t, err)
	err = client.Insert(context.Background(), "nonexistent", []*Vector{})
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestCov_MilvusClient_Insert_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	err := client.Insert(context.Background(), "test_idx", nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_Upsert(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 3, Metric: "cosine"})
	err := client.Upsert(context.Background(), "test_idx", []*Vector{{ID: "v1", Data: []float32{0.1, 0.2, 0.3}}})
	assert.NoError(t, err)
	err = client.Upsert(context.Background(), "test_idx", []*Vector{{ID: "v2", Data: []float32{0.1, 0.2}}})
	assert.Error(t, err)
}

func TestCov_MilvusClient_Upsert_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	err := client.Upsert(context.Background(), "test_idx", nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_Delete(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 3, Metric: "cosine"})
	err := client.Delete(context.Background(), "test_idx", []string{"v1"})
	assert.NoError(t, err)
	err = client.Delete(context.Background(), "nonexistent", []string{"v1"})
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestCov_MilvusClient_Delete_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	err := client.Delete(context.Background(), "test_idx", nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_Get(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 3, Metric: "cosine"})
	vectors, err := client.Get(context.Background(), "test_idx", []string{"v1"})
	assert.NoError(t, err)
	assert.Empty(t, vectors)
	_, err = client.Get(context.Background(), "nonexistent", []string{"v1"})
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestCov_MilvusClient_Get_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	_, err := client.Get(context.Background(), "test_idx", nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_Search(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 3, Metric: "cosine"})
	results, err := client.Search(context.Background(), "test_idx", []float32{0.1, 0.2, 0.3}, &SearchParams{TopK: 5})
	assert.NoError(t, err)
	assert.Empty(t, results)
	results, err = client.Search(context.Background(), "test_idx", []float32{0.1, 0.2, 0.3}, nil)
	assert.NoError(t, err)
	_, err = client.Search(context.Background(), "test_idx", []float32{0.1, 0.2}, &SearchParams{TopK: 5})
	assert.Error(t, err)
	_, err = client.Search(context.Background(), "nonexistent", []float32{0.1, 0.2, 0.3}, nil)
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestCov_MilvusClient_Search_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	_, err := client.Search(context.Background(), "test_idx", nil, nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_SearchBatch(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 3, Metric: "cosine"})
	results, err := client.SearchBatch(context.Background(), "test_idx", [][]float32{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}}, &SearchParams{TopK: 5})
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	results, err = client.SearchBatch(context.Background(), "test_idx", [][]float32{{0.1, 0.2, 0.3}}, nil)
	assert.NoError(t, err)
	_, err = client.SearchBatch(context.Background(), "test_idx", [][]float32{{0.1, 0.2}}, nil)
	assert.Error(t, err)
}

func TestCov_MilvusClient_SearchBatch_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	_, err := client.SearchBatch(context.Background(), "test_idx", nil, nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_Count(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	client.Connect(context.Background())
	client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 3, Metric: "cosine"})
	count, err := client.Count(context.Background(), "test_idx")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
	_, err = client.Count(context.Background(), "nonexistent")
	assert.Equal(t, ErrIndexNotFound, err)
}

func TestCov_MilvusClient_Count_NotConnected(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	_, err := client.Count(context.Background(), "test_idx")
	assert.Equal(t, ErrVectorDBNotConnected, err)
}

func TestCov_MilvusClient_HealthCheck(t *testing.T) {
	client := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	err := client.HealthCheck(context.Background())
	assert.Equal(t, ErrVectorDBNotConnected, err)
	client.Connect(context.Background())
	err = client.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestCov_PineconeClient_Connect(t *testing.T) {
	client := NewPineconeClient(&PineconeConfig{APIKey: "test", Environment: "us-west1"}, nil)
	err := client.Connect(context.Background())
	assert.NoError(t, err)
	assert.True(t, client.IsConnected())
	err = client.Connect(context.Background())
	assert.NoError(t, err)
}

func TestCov_PineconeClient_Disconnect(t *testing.T) {
	client := NewPineconeClient(&PineconeConfig{APIKey: "test", Environment: "us-west1"}, nil)
	client.Connect(context.Background())
	err := client.Disconnect(context.Background())
	assert.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestCov_PineconeClient_FullCRUD(t *testing.T) {
	client := NewPineconeClient(&PineconeConfig{APIKey: "test", Environment: "us-west1"}, nil)
	client.Connect(context.Background())
	err := client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 3, Metric: "cosine"})
	assert.NoError(t, err)
	exists, err := client.HasIndex(context.Background(), "test_idx")
	assert.NoError(t, err)
	assert.True(t, exists)
	indexes, err := client.ListIndexes(context.Background())
	assert.NoError(t, err)
	assert.Len(t, indexes, 1)
	config, err := client.DescribeIndex(context.Background(), "test_idx")
	assert.NoError(t, err)
	assert.Equal(t, 3, config.Dimension)
	err = client.Insert(context.Background(), "test_idx", []*Vector{{ID: "v1", Data: []float32{0.1, 0.2, 0.3}}})
	assert.NoError(t, err)
	err = client.Upsert(context.Background(), "test_idx", []*Vector{{ID: "v1", Data: []float32{0.1, 0.2, 0.3}}})
	assert.NoError(t, err)
	err = client.Delete(context.Background(), "test_idx", []string{"v1"})
	assert.NoError(t, err)
	vectors, err := client.Get(context.Background(), "test_idx", []string{"v1"})
	assert.NoError(t, err)
	assert.Empty(t, vectors)
	results, err := client.Search(context.Background(), "test_idx", []float32{0.1, 0.2, 0.3}, &SearchParams{TopK: 5})
	assert.NoError(t, err)
	assert.Empty(t, results)
	batchResults, err := client.SearchBatch(context.Background(), "test_idx", [][]float32{{0.1, 0.2, 0.3}}, nil)
	assert.NoError(t, err)
	assert.Len(t, batchResults, 1)
	count, err := client.Count(context.Background(), "test_idx")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
	err = client.HealthCheck(context.Background())
	assert.NoError(t, err)
	err = client.DropIndex(context.Background(), "test_idx")
	assert.NoError(t, err)
}

func TestCov_PineconeClient_NotConnected(t *testing.T) {
	client := NewPineconeClient(&PineconeConfig{APIKey: "test", Environment: "us-west1"}, nil)
	assert.Equal(t, ErrVectorDBNotConnected, client.CreateIndex(context.Background(), nil))
	assert.Equal(t, ErrVectorDBNotConnected, client.DropIndex(context.Background(), "x"))
	_, err := client.HasIndex(context.Background(), "x")
	assert.Equal(t, ErrVectorDBNotConnected, err)
	_, err = client.ListIndexes(context.Background())
	assert.Equal(t, ErrVectorDBNotConnected, err)
	_, err = client.DescribeIndex(context.Background(), "x")
	assert.Equal(t, ErrVectorDBNotConnected, err)
	assert.Equal(t, ErrVectorDBNotConnected, client.Insert(context.Background(), "x", nil))
	assert.Equal(t, ErrVectorDBNotConnected, client.Upsert(context.Background(), "x", nil))
	assert.Equal(t, ErrVectorDBNotConnected, client.Delete(context.Background(), "x", nil))
	_, err = client.Get(context.Background(), "x", nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
	_, err = client.Search(context.Background(), "x", nil, nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
	_, err = client.SearchBatch(context.Background(), "x", nil, nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
	_, err = client.Count(context.Background(), "x")
	assert.Equal(t, ErrVectorDBNotConnected, err)
	assert.Equal(t, ErrVectorDBNotConnected, client.HealthCheck(context.Background()))
}

func TestCov_WeaviateClient_Connect(t *testing.T) {
	client := NewWeaviateClient(&WeaviateConfig{Scheme: "http", Host: "localhost:8080"}, nil)
	err := client.Connect(context.Background())
	assert.NoError(t, err)
	assert.True(t, client.IsConnected())
	err = client.Connect(context.Background())
	assert.NoError(t, err)
}

func TestCov_WeaviateClient_Disconnect(t *testing.T) {
	client := NewWeaviateClient(&WeaviateConfig{Scheme: "http", Host: "localhost:8080"}, nil)
	client.Connect(context.Background())
	err := client.Disconnect(context.Background())
	assert.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestCov_WeaviateClient_FullCRUD(t *testing.T) {
	client := NewWeaviateClient(&WeaviateConfig{Scheme: "http", Host: "localhost:8080"}, nil)
	client.Connect(context.Background())
	err := client.CreateIndex(context.Background(), &IndexConfig{Name: "test_idx", Dimension: 3, Metric: "cosine"})
	assert.NoError(t, err)
	exists, err := client.HasIndex(context.Background(), "test_idx")
	assert.NoError(t, err)
	assert.True(t, exists)
	indexes, err := client.ListIndexes(context.Background())
	assert.NoError(t, err)
	assert.Len(t, indexes, 1)
	config, err := client.DescribeIndex(context.Background(), "test_idx")
	assert.NoError(t, err)
	assert.Equal(t, 3, config.Dimension)
	err = client.Insert(context.Background(), "test_idx", []*Vector{{ID: "v1", Data: []float32{0.1, 0.2, 0.3}}})
	assert.NoError(t, err)
	err = client.Upsert(context.Background(), "test_idx", []*Vector{{ID: "v1", Data: []float32{0.1, 0.2, 0.3}}})
	assert.NoError(t, err)
	err = client.Delete(context.Background(), "test_idx", []string{"v1"})
	assert.NoError(t, err)
	vectors, err := client.Get(context.Background(), "test_idx", []string{"v1"})
	assert.NoError(t, err)
	assert.Empty(t, vectors)
	results, err := client.Search(context.Background(), "test_idx", []float32{0.1, 0.2, 0.3}, &SearchParams{TopK: 5})
	assert.NoError(t, err)
	assert.Empty(t, results)
	batchResults, err := client.SearchBatch(context.Background(), "test_idx", [][]float32{{0.1, 0.2, 0.3}}, nil)
	assert.NoError(t, err)
	assert.Len(t, batchResults, 1)
	count, err := client.Count(context.Background(), "test_idx")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
	err = client.HealthCheck(context.Background())
	assert.NoError(t, err)
	err = client.DropIndex(context.Background(), "test_idx")
	assert.NoError(t, err)
}

func TestCov_WeaviateClient_NotConnected(t *testing.T) {
	client := NewWeaviateClient(&WeaviateConfig{Scheme: "http", Host: "localhost:8080"}, nil)
	assert.Equal(t, ErrVectorDBNotConnected, client.CreateIndex(context.Background(), nil))
	assert.Equal(t, ErrVectorDBNotConnected, client.DropIndex(context.Background(), "x"))
	_, err := client.HasIndex(context.Background(), "x")
	assert.Equal(t, ErrVectorDBNotConnected, err)
	_, err = client.ListIndexes(context.Background())
	assert.Equal(t, ErrVectorDBNotConnected, err)
	_, err = client.DescribeIndex(context.Background(), "x")
	assert.Equal(t, ErrVectorDBNotConnected, err)
	assert.Equal(t, ErrVectorDBNotConnected, client.Insert(context.Background(), "x", nil))
	assert.Equal(t, ErrVectorDBNotConnected, client.Upsert(context.Background(), "x", nil))
	assert.Equal(t, ErrVectorDBNotConnected, client.Delete(context.Background(), "x", nil))
	_, err = client.Get(context.Background(), "x", nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
	_, err = client.Search(context.Background(), "x", nil, nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
	_, err = client.SearchBatch(context.Background(), "x", nil, nil)
	assert.Equal(t, ErrVectorDBNotConnected, err)
	_, err = client.Count(context.Background(), "x")
	assert.Equal(t, ErrVectorDBNotConnected, err)
	assert.Equal(t, ErrVectorDBNotConnected, client.HealthCheck(context.Background()))
}

type covMockEmbeddingProvider struct{}

func (m *covMockEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, 128)
	for i := range vec {
		vec[i] = float32(i) * 0.01
	}
	return vec, nil
}

func (m *covMockEmbeddingProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		vec, _ := m.Embed(ctx, texts[i])
		result[i] = vec
	}
	return result, nil
}

func (m *covMockEmbeddingProvider) Dimension() int { return 128 }

func (m *covMockEmbeddingProvider) ModelName() string { return "mock" }

func newTestKM() *KnowledgeManager {
	vectorDB := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	vectorDB.Connect(context.Background())
	embedder := NewDocumentEmbedder(&DocumentEmbedderConfig{
		Provider:    &covMockEmbeddingProvider{},
		ChunkConfig: &ChunkConfig{ChunkSize: 100, ChunkOverlap: 10},
	}, nil)
	retriever := NewRAGRetriever(vectorDB, embedder, DefaultRAGConfig(), nil)
	return NewKnowledgeManager(vectorDB, embedder, retriever, nil)
}

func TestCov_KnowledgeManager_GetDocument(t *testing.T) {
	km := newTestKM()
	doc := &Document{ID: "doc1", Title: "Test", Content: "content", Source: "test", Metadata: map[string]interface{}{"knowledge_base_id": "kb1"}}
	km.docStore.Save(context.Background(), doc)
	result, err := km.GetDocument(context.Background(), "doc1")
	require.NoError(t, err)
	assert.Equal(t, "doc1", result.ID)
}

func TestCov_KnowledgeManager_ExportKnowledgeBase(t *testing.T) {
	km := newTestKM()
	km.CreateKnowledgeBase(context.Background(), &CreateKnowledgeBaseRequest{
		Name: "test_kb", Description: "test", Dimension: 128,
	})
	var kbID string
	for id := range km.knowledgeBases {
		kbID = id
		break
	}
	export, err := km.ExportKnowledgeBase(context.Background(), kbID)
	require.NoError(t, err)
	assert.NotNil(t, export["knowledge_base"])
	assert.NotNil(t, export["documents"])
	_, err = km.ExportKnowledgeBase(context.Background(), "nonexistent")
	assert.Equal(t, ErrKnowledgeBaseNotFound, err)
}

func TestCov_KnowledgeManager_ImportKnowledgeBase(t *testing.T) {
	km := newTestKM()
	data := map[string]interface{}{
		"knowledge_base": map[string]interface{}{
			"name":        "imported_kb",
			"description": "imported",
			"dimension":   float64(128),
		},
		"documents": []interface{}{
			map[string]interface{}{
				"title":   "doc1",
				"content": "content1",
				"source":  "test",
			},
		},
	}
	kb, err := km.ImportKnowledgeBase(context.Background(), data)
	require.NoError(t, err)
	assert.NotNil(t, kb)
	assert.Equal(t, "imported_kb", kb.Name)
	_, err = km.ImportKnowledgeBase(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
}

func TestCov_KnowledgeManager_DeleteDocument_KBNotFound(t *testing.T) {
	km := newTestKM()
	err := km.DeleteDocument(context.Background(), "nonexistent_kb", "doc1")
	assert.Equal(t, ErrKnowledgeBaseNotFound, err)
}

func TestCov_RAGService_Query(t *testing.T) {
	vectorDB := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	vectorDB.Connect(context.Background())
	retriever := NewRAGRetriever(vectorDB, nil, &RAGConfig{TopK: 5, VectorWeight: 0.7, KeywordWeight: 0.3, ScoreThreshold: 0.5}, nil)
	rs := NewRAGService(retriever, nil)
	assert.NotNil(t, rs)
	stats := rs.GetStats()
	assert.NotNil(t, stats)
}

func TestCov_MultiQueryRetriever_Retrieve(t *testing.T) {
	vectorDB := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	vectorDB.Connect(context.Background())
	retriever := NewRAGRetriever(vectorDB, nil, &RAGConfig{TopK: 5, FinalTopK: 10, VectorWeight: 0.7, KeywordWeight: 0.3, ScoreThreshold: 0.5}, nil)
	mqr := NewMultiQueryRetriever(retriever, nil)
	assert.NotNil(t, mqr)
}

func TestCov_VectorDBFactory(t *testing.T) {
	factory := NewVectorDBFactory(nil)
	milvus := factory.CreateMilvus(&MilvusConfig{Address: "localhost", Port: 19530})
	assert.NotNil(t, milvus)
	pinecone := factory.CreatePinecone(&PineconeConfig{APIKey: "test", Environment: "us-west1"})
	assert.NotNil(t, pinecone)
	weaviate := factory.CreateWeaviate(&WeaviateConfig{Scheme: "http", Host: "localhost:8080"})
	assert.NotNil(t, weaviate)
}

func TestCov_VectorIndex_Search_TopK(t *testing.T) {
	idx := NewVectorIndex("test", 3, "cosine")
	idx.Insert(&Vector{ID: "v1", Data: []float32{1.0, 0.0, 0.0}})
	idx.Insert(&Vector{ID: "v2", Data: []float32{0.9, 0.1, 0.0}})
	idx.Insert(&Vector{ID: "v3", Data: []float32{0.0, 1.0, 0.0}})
	results, err := idx.Search([]float32{1.0, 0.0, 0.0}, 2)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestCov_VectorSerializer_Deserialize_Invalid(t *testing.T) {
	vs := &VectorSerializer{}
	_, err := vs.Deserialize([]byte("invalid"))
	assert.Error(t, err)
}

func TestCov_RAGRetriever_mergeResults(t *testing.T) {
	vectorDB := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	vectorDB.Connect(context.Background())
	r := NewRAGRetriever(vectorDB, nil, &RAGConfig{TopK: 5, VectorWeight: 0.7, KeywordWeight: 0.3, ScoreThreshold: 0.5}, nil)
	vectorResults := []*RetrievalResult{
		{Chunk: &TextChunk{ID: "c1", Content: "hello"}, Score: 0.9, VectorScore: 0.9},
		{Chunk: &TextChunk{ID: "c2", Content: "world"}, Score: 0.8, VectorScore: 0.8},
	}
	keywordResults := []*TextChunk{
		{ID: "c1", Content: "hello"},
		{ID: "c3", Content: "foo"},
	}
	results := r.mergeResults(vectorResults, keywordResults)
	assert.Len(t, results, 3)
	for _, r := range results {
		if r.Chunk.ID == "c1" {
			assert.Equal(t, "hybrid", r.Source)
		}
	}
}

func TestCov_RAGRetriever_filterResults(t *testing.T) {
	vectorDB := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	vectorDB.Connect(context.Background())
	r := NewRAGRetriever(vectorDB, nil, &RAGConfig{TopK: 5, VectorWeight: 0.7, KeywordWeight: 0.3, ScoreThreshold: 0.5}, nil)
	results := []*RetrievalResult{
		{Chunk: &TextChunk{ID: "c1"}, Score: 0.9},
		{Chunk: &TextChunk{ID: "c2"}, Score: 0.3},
		{Chunk: &TextChunk{ID: "c3"}, Score: 0.6},
	}
	filtered := r.filterResults(results)
	assert.Len(t, filtered, 2)
}

func TestCov_RAGRetriever_DeleteChunks_Error(t *testing.T) {
	vectorDB := NewMilvusClient(&MilvusConfig{Address: "localhost", Port: 19530}, nil)
	r := NewRAGRetriever(vectorDB, nil, &RAGConfig{TopK: 5, VectorWeight: 0.7, KeywordWeight: 0.3, ScoreThreshold: 0.5}, nil)
	err := r.DeleteChunks(context.Background(), "nonexistent", []string{"c1"})
	assert.Error(t, err)
}
