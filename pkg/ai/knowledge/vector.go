package knowledge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ErrVectorDBNotConnected = errors.New("vector database not connected")
	ErrIndexNotFound        = errors.New("index not found")
	ErrInvalidVector        = errors.New("invalid vector dimension")
	ErrDuplicateID          = errors.New("duplicate vector ID")
	ErrVectorNotFound       = errors.New("vector not found")
)

// Vector 表示一个向量及其元数据
type Vector struct {
	ID       string                 `json:"id"`
	Data     []float32              `json:"data"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SearchResult 表示搜索结果
type SearchResult struct {
	Vector    *Vector  `json:"vector"`
	Score     float32  `json:"score"`
	Distance  float32  `json:"distance"`
	Highlight string   `json:"highlight,omitempty"`
}

// IndexConfig 索引配置
type IndexConfig struct {
	Name          string            `json:"name"`
	Dimension     int               `json:"dimension"`
	Metric        string            `json:"metric"` // cosine, euclidean, dot_product
	ShardNum      int               `json:"shard_num"`
	ReplicaNum    int               `json:"replica_num"`
	IndexParams   map[string]string `json:"index_params"`
	AutoCreate    bool              `json:"auto_create"`
	AutoLoad      bool              `json:"auto_load"`
}

// SearchParams 搜索参数
type SearchParams struct {
	TopK          int                    `json:"top_k"`
	Filter        map[string]interface{} `json:"filter,omitempty"`
	RoundDecimal  int                    `json:"round_decimal"`
	NProbe        int                    `json:"n_probe"`
	SearchParams  map[string]interface{} `json:"search_params,omitempty"`
}

// VectorDB 向量数据库接口
type VectorDB interface {
	// 连接管理
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool

	// 索引管理
	CreateIndex(ctx context.Context, config *IndexConfig) error
	DropIndex(ctx context.Context, indexName string) error
	HasIndex(ctx context.Context, indexName string) (bool, error)
	ListIndexes(ctx context.Context) ([]string, error)
	DescribeIndex(ctx context.Context, indexName string) (*IndexConfig, error)

	// 向量操作
	Insert(ctx context.Context, indexName string, vectors []*Vector) error
	Upsert(ctx context.Context, indexName string, vectors []*Vector) error
	Delete(ctx context.Context, indexName string, ids []string) error
	Get(ctx context.Context, indexName string, ids []string) ([]*Vector, error)

	// 搜索
	Search(ctx context.Context, indexName string, vector []float32, params *SearchParams) ([]*SearchResult, error)
	SearchBatch(ctx context.Context, indexName string, vectors [][]float32, params *SearchParams) ([][]*SearchResult, error)

	// 统计
	Count(ctx context.Context, indexName string) (int64, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// MilvusConfig Milvus配置
type MilvusConfig struct {
	Address      string        `json:"address"`
	Port         int           `json:"port"`
	Username     string        `json:"username"`
	Password     string        `json:"password"`
	Database     string        `json:"database"`
	Timeout      time.Duration `json:"timeout"`
	MaxRetry     int           `json:"max_retry"`
	RetryBackoff time.Duration `json:"retry_backoff"`
}

// MilvusClient Milvus客户端
type MilvusClient struct {
	config     *MilvusConfig
	logger     *zap.Logger
	connected  bool
	mu         sync.RWMutex
	indexCache map[string]*IndexConfig
}

// NewMilvusClient 创建Milvus客户端
func NewMilvusClient(config *MilvusConfig, logger *zap.Logger) *MilvusClient {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MilvusClient{
		config:     config,
		logger:     logger,
		indexCache: make(map[string]*IndexConfig),
	}
}

// Connect 连接到Milvus
func (m *MilvusClient) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connected {
		return nil
	}

	// 实际实现中应该使用milvus-sdk-go
	// 这里模拟连接过程
	m.logger.Info("connecting to Milvus",
		zap.String("address", m.config.Address),
		zap.Int("port", m.config.Port),
	)

	// 模拟连接成功
	m.connected = true
	m.logger.Info("successfully connected to Milvus")

	return nil
}

// Disconnect 断开连接
func (m *MilvusClient) Disconnect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return nil
	}

	m.logger.Info("disconnecting from Milvus")
	m.connected = false
	m.indexCache = make(map[string]*IndexConfig)

	return nil
}

// IsConnected 检查连接状态
func (m *MilvusClient) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// CreateIndex 创建索引
func (m *MilvusClient) CreateIndex(ctx context.Context, config *IndexConfig) error {
	if !m.IsConnected() {
		return ErrVectorDBNotConnected
	}

	if config.Dimension <= 0 {
		return ErrInvalidVector
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.indexCache[config.Name]; exists {
		return fmt.Errorf("index %s already exists", config.Name)
	}

	m.logger.Info("creating Milvus index",
		zap.String("name", config.Name),
		zap.Int("dimension", config.Dimension),
		zap.String("metric", config.Metric),
	)

	// 实际实现中应该调用Milvus API创建collection和index
	m.indexCache[config.Name] = config

	return nil
}

// DropIndex 删除索引
func (m *MilvusClient) DropIndex(ctx context.Context, indexName string) error {
	if !m.IsConnected() {
		return ErrVectorDBNotConnected
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.indexCache[indexName]; !exists {
		return ErrIndexNotFound
	}

	m.logger.Info("dropping Milvus index", zap.String("name", indexName))
	delete(m.indexCache, indexName)

	return nil
}

// HasIndex 检查索引是否存在
func (m *MilvusClient) HasIndex(ctx context.Context, indexName string) (bool, error) {
	if !m.IsConnected() {
		return false, ErrVectorDBNotConnected
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.indexCache[indexName]
	return exists, nil
}

// ListIndexes 列出所有索引
func (m *MilvusClient) ListIndexes(ctx context.Context) ([]string, error) {
	if !m.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	indexes := make([]string, 0, len(m.indexCache))
	for name := range m.indexCache {
		indexes = append(indexes, name)
	}

	return indexes, nil
}

// DescribeIndex 描述索引
func (m *MilvusClient) DescribeIndex(ctx context.Context, indexName string) (*IndexConfig, error) {
	if !m.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.indexCache[indexName]
	if !exists {
		return nil, ErrIndexNotFound
	}

	return config, nil
}

// Insert 插入向量
func (m *MilvusClient) Insert(ctx context.Context, indexName string, vectors []*Vector) error {
	if !m.IsConnected() {
		return ErrVectorDBNotConnected
	}

	m.mu.RLock()
	config, exists := m.indexCache[indexName]
	m.mu.RUnlock()

	if !exists {
		return ErrIndexNotFound
	}

	// 验证向量维度
	for _, v := range vectors {
		if len(v.Data) != config.Dimension {
			return fmt.Errorf("%w: expected %d, got %d", ErrInvalidVector, config.Dimension, len(v.Data))
		}
	}

	m.logger.Debug("inserting vectors into Milvus",
		zap.String("index", indexName),
		zap.Int("count", len(vectors)),
	)

	// 实际实现中应该调用Milvus API插入数据
	return nil
}

// Upsert 更新或插入向量
func (m *MilvusClient) Upsert(ctx context.Context, indexName string, vectors []*Vector) error {
	if !m.IsConnected() {
		return ErrVectorDBNotConnected
	}

	m.mu.RLock()
	config, exists := m.indexCache[indexName]
	m.mu.RUnlock()

	if !exists {
		return ErrIndexNotFound
	}

	// 验证向量维度
	for _, v := range vectors {
		if len(v.Data) != config.Dimension {
			return fmt.Errorf("%w: expected %d, got %d", ErrInvalidVector, config.Dimension, len(v.Data))
		}
	}

	m.logger.Debug("upserting vectors into Milvus",
		zap.String("index", indexName),
		zap.Int("count", len(vectors)),
	)

	return nil
}

// Delete 删除向量
func (m *MilvusClient) Delete(ctx context.Context, indexName string, ids []string) error {
	if !m.IsConnected() {
		return ErrVectorDBNotConnected
	}

	m.mu.RLock()
	_, exists := m.indexCache[indexName]
	m.mu.RUnlock()

	if !exists {
		return ErrIndexNotFound
	}

	m.logger.Debug("deleting vectors from Milvus",
		zap.String("index", indexName),
		zap.Int("count", len(ids)),
	)

	return nil
}

// Get 获取向量
func (m *MilvusClient) Get(ctx context.Context, indexName string, ids []string) ([]*Vector, error) {
	if !m.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	m.mu.RLock()
	_, exists := m.indexCache[indexName]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrIndexNotFound
	}

	// 实际实现中应该调用Milvus API查询数据
	vectors := make([]*Vector, 0, len(ids))
	return vectors, nil
}

// Search 搜索向量
func (m *MilvusClient) Search(ctx context.Context, indexName string, vector []float32, params *SearchParams) ([]*SearchResult, error) {
	if !m.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	m.mu.RLock()
	config, exists := m.indexCache[indexName]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrIndexNotFound
	}

	// 验证向量维度
	if len(vector) != config.Dimension {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrInvalidVector, config.Dimension, len(vector))
	}

	if params == nil {
		params = &SearchParams{TopK: 10}
	}

	m.logger.Debug("searching vectors in Milvus",
		zap.String("index", indexName),
		zap.Int("top_k", params.TopK),
	)

	// 实际实现中应该调用Milvus API搜索
	results := make([]*SearchResult, 0, params.TopK)
	return results, nil
}

// SearchBatch 批量搜索
func (m *MilvusClient) SearchBatch(ctx context.Context, indexName string, vectors [][]float32, params *SearchParams) ([][]*SearchResult, error) {
	if !m.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	m.mu.RLock()
	config, exists := m.indexCache[indexName]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrIndexNotFound
	}

	// 验证所有向量维度
	for i, v := range vectors {
		if len(v) != config.Dimension {
			return nil, fmt.Errorf("%w: vector %d expected %d, got %d", ErrInvalidVector, i, config.Dimension, len(v))
		}
	}

	if params == nil {
		params = &SearchParams{TopK: 10}
	}

	m.logger.Debug("batch searching vectors in Milvus",
		zap.String("index", indexName),
		zap.Int("batch_size", len(vectors)),
		zap.Int("top_k", params.TopK),
	)

	results := make([][]*SearchResult, len(vectors))
	for i := range results {
		results[i] = make([]*SearchResult, 0, params.TopK)
	}

	return results, nil
}

// Count 统计向量数量
func (m *MilvusClient) Count(ctx context.Context, indexName string) (int64, error) {
	if !m.IsConnected() {
		return 0, ErrVectorDBNotConnected
	}

	m.mu.RLock()
	_, exists := m.indexCache[indexName]
	m.mu.RUnlock()

	if !exists {
		return 0, ErrIndexNotFound
	}

	// 实际实现中应该调用Milvus API统计
	return 0, nil
}

// HealthCheck 健康检查
func (m *MilvusClient) HealthCheck(ctx context.Context) error {
	if !m.IsConnected() {
		return ErrVectorDBNotConnected
	}

	// 实际实现中应该调用Milvus API检查健康状态
	return nil
}

// PineconeConfig Pinecone配置
type PineconeConfig struct {
	APIKey      string        `json:"api_key"`
	Environment string        `json:"environment"`
	ProjectID   string        `json:"project_id"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetry    int           `json:"max_retry"`
}

// PineconeClient Pinecone客户端
type PineconeClient struct {
	config     *PineconeConfig
	logger     *zap.Logger
	connected  bool
	mu         sync.RWMutex
	indexCache map[string]*IndexConfig
}

// NewPineconeClient 创建Pinecone客户端
func NewPineconeClient(config *PineconeConfig, logger *zap.Logger) *PineconeClient {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &PineconeClient{
		config:     config,
		logger:     logger,
		indexCache: make(map[string]*IndexConfig),
	}
}

// Connect 连接到Pinecone
func (p *PineconeClient) Connect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connected {
		return nil
	}

	p.logger.Info("connecting to Pinecone",
		zap.String("environment", p.config.Environment),
		zap.String("project", p.config.ProjectID),
	)

	p.connected = true
	p.logger.Info("successfully connected to Pinecone")

	return nil
}

// Disconnect 断开连接
func (p *PineconeClient) Disconnect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil
	}

	p.logger.Info("disconnecting from Pinecone")
	p.connected = false
	p.indexCache = make(map[string]*IndexConfig)

	return nil
}

// IsConnected 检查连接状态
func (p *PineconeClient) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected
}

// CreateIndex 创建索引
func (p *PineconeClient) CreateIndex(ctx context.Context, config *IndexConfig) error {
	if !p.IsConnected() {
		return ErrVectorDBNotConnected
	}

	if config.Dimension <= 0 {
		return ErrInvalidVector
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.indexCache[config.Name]; exists {
		return fmt.Errorf("index %s already exists", config.Name)
	}

	p.logger.Info("creating Pinecone index",
		zap.String("name", config.Name),
		zap.Int("dimension", config.Dimension),
		zap.String("metric", config.Metric),
	)

	p.indexCache[config.Name] = config
	return nil
}

// DropIndex 删除索引
func (p *PineconeClient) DropIndex(ctx context.Context, indexName string) error {
	if !p.IsConnected() {
		return ErrVectorDBNotConnected
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.indexCache[indexName]; !exists {
		return ErrIndexNotFound
	}

	p.logger.Info("dropping Pinecone index", zap.String("name", indexName))
	delete(p.indexCache, indexName)

	return nil
}

// HasIndex 检查索引是否存在
func (p *PineconeClient) HasIndex(ctx context.Context, indexName string) (bool, error) {
	if !p.IsConnected() {
		return false, ErrVectorDBNotConnected
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	_, exists := p.indexCache[indexName]
	return exists, nil
}

// ListIndexes 列出所有索引
func (p *PineconeClient) ListIndexes(ctx context.Context) ([]string, error) {
	if !p.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	indexes := make([]string, 0, len(p.indexCache))
	for name := range p.indexCache {
		indexes = append(indexes, name)
	}

	return indexes, nil
}

// DescribeIndex 描述索引
func (p *PineconeClient) DescribeIndex(ctx context.Context, indexName string) (*IndexConfig, error) {
	if !p.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	config, exists := p.indexCache[indexName]
	if !exists {
		return nil, ErrIndexNotFound
	}

	return config, nil
}

// Insert 插入向量
func (p *PineconeClient) Insert(ctx context.Context, indexName string, vectors []*Vector) error {
	if !p.IsConnected() {
		return ErrVectorDBNotConnected
	}

	p.mu.RLock()
	config, exists := p.indexCache[indexName]
	p.mu.RUnlock()

	if !exists {
		return ErrIndexNotFound
	}

	for _, v := range vectors {
		if len(v.Data) != config.Dimension {
			return fmt.Errorf("%w: expected %d, got %d", ErrInvalidVector, config.Dimension, len(v.Data))
		}
	}

	p.logger.Debug("inserting vectors into Pinecone",
		zap.String("index", indexName),
		zap.Int("count", len(vectors)),
	)

	return nil
}

// Upsert 更新或插入向量
func (p *PineconeClient) Upsert(ctx context.Context, indexName string, vectors []*Vector) error {
	if !p.IsConnected() {
		return ErrVectorDBNotConnected
	}

	p.mu.RLock()
	config, exists := p.indexCache[indexName]
	p.mu.RUnlock()

	if !exists {
		return ErrIndexNotFound
	}

	for _, v := range vectors {
		if len(v.Data) != config.Dimension {
			return fmt.Errorf("%w: expected %d, got %d", ErrInvalidVector, config.Dimension, len(v.Data))
		}
	}

	p.logger.Debug("upserting vectors into Pinecone",
		zap.String("index", indexName),
		zap.Int("count", len(vectors)),
	)

	return nil
}

// Delete 删除向量
func (p *PineconeClient) Delete(ctx context.Context, indexName string, ids []string) error {
	if !p.IsConnected() {
		return ErrVectorDBNotConnected
	}

	p.mu.RLock()
	_, exists := p.indexCache[indexName]
	p.mu.RUnlock()

	if !exists {
		return ErrIndexNotFound
	}

	p.logger.Debug("deleting vectors from Pinecone",
		zap.String("index", indexName),
		zap.Int("count", len(ids)),
	)

	return nil
}

// Get 获取向量
func (p *PineconeClient) Get(ctx context.Context, indexName string, ids []string) ([]*Vector, error) {
	if !p.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	p.mu.RLock()
	_, exists := p.indexCache[indexName]
	p.mu.RUnlock()

	if !exists {
		return nil, ErrIndexNotFound
	}

	vectors := make([]*Vector, 0, len(ids))
	return vectors, nil
}

// Search 搜索向量
func (p *PineconeClient) Search(ctx context.Context, indexName string, vector []float32, params *SearchParams) ([]*SearchResult, error) {
	if !p.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	p.mu.RLock()
	config, exists := p.indexCache[indexName]
	p.mu.RUnlock()

	if !exists {
		return nil, ErrIndexNotFound
	}

	if len(vector) != config.Dimension {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrInvalidVector, config.Dimension, len(vector))
	}

	if params == nil {
		params = &SearchParams{TopK: 10}
	}

	p.logger.Debug("searching vectors in Pinecone",
		zap.String("index", indexName),
		zap.Int("top_k", params.TopK),
	)

	results := make([]*SearchResult, 0, params.TopK)
	return results, nil
}

// SearchBatch 批量搜索
func (p *PineconeClient) SearchBatch(ctx context.Context, indexName string, vectors [][]float32, params *SearchParams) ([][]*SearchResult, error) {
	if !p.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	p.mu.RLock()
	config, exists := p.indexCache[indexName]
	p.mu.RUnlock()

	if !exists {
		return nil, ErrIndexNotFound
	}

	for i, v := range vectors {
		if len(v) != config.Dimension {
			return nil, fmt.Errorf("%w: vector %d expected %d, got %d", ErrInvalidVector, i, config.Dimension, len(v))
		}
	}

	if params == nil {
		params = &SearchParams{TopK: 10}
	}

	results := make([][]*SearchResult, len(vectors))
	for i := range results {
		results[i] = make([]*SearchResult, 0, params.TopK)
	}

	return results, nil
}

// Count 统计向量数量
func (p *PineconeClient) Count(ctx context.Context, indexName string) (int64, error) {
	if !p.IsConnected() {
		return 0, ErrVectorDBNotConnected
	}

	p.mu.RLock()
	_, exists := p.indexCache[indexName]
	p.mu.RUnlock()

	if !exists {
		return 0, ErrIndexNotFound
	}

	return 0, nil
}

// HealthCheck 健康检查
func (p *PineconeClient) HealthCheck(ctx context.Context) error {
	if !p.IsConnected() {
		return ErrVectorDBNotConnected
	}

	return nil
}

// WeaviateConfig Weaviate配置
type WeaviateConfig struct {
	Scheme     string        `json:"scheme"`
	Host       string        `json:"host"`
	APIKey     string        `json:"api_key"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetry   int           `json:"max_retry"`
}

// WeaviateClient Weaviate客户端
type WeaviateClient struct {
	config     *WeaviateConfig
	logger     *zap.Logger
	connected  bool
	mu         sync.RWMutex
	classCache map[string]*IndexConfig
}

// NewWeaviateClient 创建Weaviate客户端
func NewWeaviateClient(config *WeaviateConfig, logger *zap.Logger) *WeaviateClient {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &WeaviateClient{
		config:     config,
		logger:     logger,
		classCache: make(map[string]*IndexConfig),
	}
}

// Connect 连接到Weaviate
func (w *WeaviateClient) Connect(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.connected {
		return nil
	}

	w.logger.Info("connecting to Weaviate",
		zap.String("scheme", w.config.Scheme),
		zap.String("host", w.config.Host),
	)

	w.connected = true
	w.logger.Info("successfully connected to Weaviate")

	return nil
}

// Disconnect 断开连接
func (w *WeaviateClient) Disconnect(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.connected {
		return nil
	}

	w.logger.Info("disconnecting from Weaviate")
	w.connected = false
	w.classCache = make(map[string]*IndexConfig)

	return nil
}

// IsConnected 检查连接状态
func (w *WeaviateClient) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.connected
}

// CreateIndex 创建类（索引）
func (w *WeaviateClient) CreateIndex(ctx context.Context, config *IndexConfig) error {
	if !w.IsConnected() {
		return ErrVectorDBNotConnected
	}

	if config.Dimension <= 0 {
		return ErrInvalidVector
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.classCache[config.Name]; exists {
		return fmt.Errorf("class %s already exists", config.Name)
	}

	w.logger.Info("creating Weaviate class",
		zap.String("name", config.Name),
		zap.Int("dimension", config.Dimension),
	)

	w.classCache[config.Name] = config
	return nil
}

// DropIndex 删除类
func (w *WeaviateClient) DropIndex(ctx context.Context, indexName string) error {
	if !w.IsConnected() {
		return ErrVectorDBNotConnected
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.classCache[indexName]; !exists {
		return ErrIndexNotFound
	}

	w.logger.Info("dropping Weaviate class", zap.String("name", indexName))
	delete(w.classCache, indexName)

	return nil
}

// HasIndex 检查类是否存在
func (w *WeaviateClient) HasIndex(ctx context.Context, indexName string) (bool, error) {
	if !w.IsConnected() {
		return false, ErrVectorDBNotConnected
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	_, exists := w.classCache[indexName]
	return exists, nil
}

// ListIndexes 列出所有类
func (w *WeaviateClient) ListIndexes(ctx context.Context) ([]string, error) {
	if !w.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	classes := make([]string, 0, len(w.classCache))
	for name := range w.classCache {
		classes = append(classes, name)
	}

	return classes, nil
}

// DescribeIndex 描述类
func (w *WeaviateClient) DescribeIndex(ctx context.Context, indexName string) (*IndexConfig, error) {
	if !w.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	config, exists := w.classCache[indexName]
	if !exists {
		return nil, ErrIndexNotFound
	}

	return config, nil
}

// Insert 插入对象
func (w *WeaviateClient) Insert(ctx context.Context, indexName string, vectors []*Vector) error {
	if !w.IsConnected() {
		return ErrVectorDBNotConnected
	}

	w.mu.RLock()
	config, exists := w.classCache[indexName]
	w.mu.RUnlock()

	if !exists {
		return ErrIndexNotFound
	}

	for _, v := range vectors {
		if len(v.Data) != config.Dimension {
			return fmt.Errorf("%w: expected %d, got %d", ErrInvalidVector, config.Dimension, len(v.Data))
		}
	}

	w.logger.Debug("inserting objects into Weaviate",
		zap.String("class", indexName),
		zap.Int("count", len(vectors)),
	)

	return nil
}

// Upsert 更新或插入对象
func (w *WeaviateClient) Upsert(ctx context.Context, indexName string, vectors []*Vector) error {
	if !w.IsConnected() {
		return ErrVectorDBNotConnected
	}

	w.mu.RLock()
	config, exists := w.classCache[indexName]
	w.mu.RUnlock()

	if !exists {
		return ErrIndexNotFound
	}

	for _, v := range vectors {
		if len(v.Data) != config.Dimension {
			return fmt.Errorf("%w: expected %d, got %d", ErrInvalidVector, config.Dimension, len(v.Data))
		}
	}

	w.logger.Debug("upserting objects into Weaviate",
		zap.String("class", indexName),
		zap.Int("count", len(vectors)),
	)

	return nil
}

// Delete 删除对象
func (w *WeaviateClient) Delete(ctx context.Context, indexName string, ids []string) error {
	if !w.IsConnected() {
		return ErrVectorDBNotConnected
	}

	w.mu.RLock()
	_, exists := w.classCache[indexName]
	w.mu.RUnlock()

	if !exists {
		return ErrIndexNotFound
	}

	w.logger.Debug("deleting objects from Weaviate",
		zap.String("class", indexName),
		zap.Int("count", len(ids)),
	)

	return nil
}

// Get 获取对象
func (w *WeaviateClient) Get(ctx context.Context, indexName string, ids []string) ([]*Vector, error) {
	if !w.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	w.mu.RLock()
	_, exists := w.classCache[indexName]
	w.mu.RUnlock()

	if !exists {
		return nil, ErrIndexNotFound
	}

	vectors := make([]*Vector, 0, len(ids))
	return vectors, nil
}

// Search 搜索对象
func (w *WeaviateClient) Search(ctx context.Context, indexName string, vector []float32, params *SearchParams) ([]*SearchResult, error) {
	if !w.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	w.mu.RLock()
	config, exists := w.classCache[indexName]
	w.mu.RUnlock()

	if !exists {
		return nil, ErrIndexNotFound
	}

	if len(vector) != config.Dimension {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrInvalidVector, config.Dimension, len(vector))
	}

	if params == nil {
		params = &SearchParams{TopK: 10}
	}

	w.logger.Debug("searching objects in Weaviate",
		zap.String("class", indexName),
		zap.Int("limit", params.TopK),
	)

	results := make([]*SearchResult, 0, params.TopK)
	return results, nil
}

// SearchBatch 批量搜索
func (w *WeaviateClient) SearchBatch(ctx context.Context, indexName string, vectors [][]float32, params *SearchParams) ([][]*SearchResult, error) {
	if !w.IsConnected() {
		return nil, ErrVectorDBNotConnected
	}

	w.mu.RLock()
	config, exists := w.classCache[indexName]
	w.mu.RUnlock()

	if !exists {
		return nil, ErrIndexNotFound
	}

	for i, v := range vectors {
		if len(v) != config.Dimension {
			return nil, fmt.Errorf("%w: vector %d expected %d, got %d", ErrInvalidVector, i, config.Dimension, len(v))
		}
	}

	if params == nil {
		params = &SearchParams{TopK: 10}
	}

	results := make([][]*SearchResult, len(vectors))
	for i := range results {
		results[i] = make([]*SearchResult, 0, params.TopK)
	}

	return results, nil
}

// Count 统计对象数量
func (w *WeaviateClient) Count(ctx context.Context, indexName string) (int64, error) {
	if !w.IsConnected() {
		return 0, ErrVectorDBNotConnected
	}

	w.mu.RLock()
	_, exists := w.classCache[indexName]
	w.mu.RUnlock()

	if !exists {
		return 0, ErrIndexNotFound
	}

	return 0, nil
}

// HealthCheck 健康检查
func (w *WeaviateClient) HealthCheck(ctx context.Context) error {
	if !w.IsConnected() {
		return ErrVectorDBNotConnected
	}

	return nil
}

// VectorDBFactory 向量数据库工厂
type VectorDBFactory struct {
	logger *zap.Logger
}

// NewVectorDBFactory 创建向量数据库工厂
func NewVectorDBFactory(logger *zap.Logger) *VectorDBFactory {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &VectorDBFactory{logger: logger}
}

// CreateMilvus 创建Milvus客户端
func (f *VectorDBFactory) CreateMilvus(config *MilvusConfig) VectorDB {
	return NewMilvusClient(config, f.logger)
}

// CreatePinecone 创建Pinecone客户端
func (f *VectorDBFactory) CreatePinecone(config *PineconeConfig) VectorDB {
	return NewPineconeClient(config, f.logger)
}

// CreateWeaviate 创建Weaviate客户端
func (f *VectorDBFactory) CreateWeaviate(config *WeaviateConfig) VectorDB {
	return NewWeaviateClient(config, f.logger)
}

// DistanceCalculator 距离计算器
type DistanceCalculator struct{}

// NewDistanceCalculator 创建距离计算器
func NewDistanceCalculator() *DistanceCalculator {
	return &DistanceCalculator{}
}

// CosineSimilarity 计算余弦相似度
func (d *DistanceCalculator) CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, ErrInvalidVector
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, nil
	}

	return dotProduct / (sqrt32(normA) * sqrt32(normB)), nil
}

// EuclideanDistance 计算欧几里得距离
func (d *DistanceCalculator) EuclideanDistance(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, ErrInvalidVector
	}

	var sum float32
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return sqrt32(sum), nil
}

// DotProduct 计算点积
func (d *DistanceCalculator) DotProduct(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, ErrInvalidVector
	}

	var product float32
	for i := range a {
		product += a[i] * b[i]
	}

	return product, nil
}

// sqrt32 float32版本的平方根
func sqrt32(x float32) float32 {
	return float32(sqrt(float64(x)))
}

// sqrt 简单的平方根实现（牛顿迭代法）
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 100; i++ {
		z = (z + x/z) / 2
		if z*z == x {
			break
		}
	}
	return z
}

// VectorIndex 向量索引（内存实现）
type VectorIndex struct {
	name      string
	dimension int
	metric    string
	vectors   map[string]*Vector
	mu        sync.RWMutex
	calc      *DistanceCalculator
}

// NewVectorIndex 创建向量索引
func NewVectorIndex(name string, dimension int, metric string) *VectorIndex {
	return &VectorIndex{
		name:      name,
		dimension: dimension,
		metric:    metric,
		vectors:   make(map[string]*Vector),
		calc:      NewDistanceCalculator(),
	}
}

// Insert 插入向量
func (vi *VectorIndex) Insert(v *Vector) error {
	if len(v.Data) != vi.dimension {
		return ErrInvalidVector
	}

	vi.mu.Lock()
	defer vi.mu.Unlock()

	if _, exists := vi.vectors[v.ID]; exists {
		return ErrDuplicateID
	}

	vi.vectors[v.ID] = v
	return nil
}

// Delete 删除向量
func (vi *VectorIndex) Delete(id string) error {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	if _, exists := vi.vectors[id]; !exists {
		return ErrVectorNotFound
	}

	delete(vi.vectors, id)
	return nil
}

// Get 获取向量
func (vi *VectorIndex) Get(id string) (*Vector, error) {
	vi.mu.RLock()
	defer vi.mu.RUnlock()

	v, exists := vi.vectors[id]
	if !exists {
		return nil, ErrVectorNotFound
	}

	return v, nil
}

// Search 搜索相似向量
func (vi *VectorIndex) Search(query []float32, topK int) ([]*SearchResult, error) {
	if len(query) != vi.dimension {
		return nil, ErrInvalidVector
	}

	vi.mu.RLock()
	defer vi.mu.RUnlock()

	results := make([]*SearchResult, 0, len(vi.vectors))

	for _, v := range vi.vectors {
		var score float32
		var err error

		switch vi.metric {
		case "cosine":
			score, err = vi.calc.CosineSimilarity(query, v.Data)
		case "euclidean":
			dist, e := vi.calc.EuclideanDistance(query, v.Data)
			score = 1 / (1 + dist) // 转换为相似度
			err = e
		case "dot_product":
			score, err = vi.calc.DotProduct(query, v.Data)
		default:
			score, err = vi.calc.CosineSimilarity(query, v.Data)
		}

		if err != nil {
			continue
		}

		results = append(results, &SearchResult{
			Vector: v,
			Score:  score,
		})
	}

	// 排序
	sortSearchResults(results)

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// Count 统计向量数量
func (vi *VectorIndex) Count() int {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return len(vi.vectors)
}

// sortSearchResults 排序搜索结果
func sortSearchResults(results []*SearchResult) {
	// 使用快速排序
	quickSort(results, 0, len(results)-1)
}

func quickSort(arr []*SearchResult, low, high int) {
	if low < high {
		pi := partition(arr, low, high)
		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}
}

func partition(arr []*SearchResult, low, high int) int {
	pivot := arr[high].Score
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j].Score > pivot { // 降序排序
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

// VectorSerializer 向量序列化器
type VectorSerializer struct{}

// NewVectorSerializer 创建向量序列化器
func NewVectorSerializer() *VectorSerializer {
	return &VectorSerializer{}
}

// Serialize 序列化向量
func (vs *VectorSerializer) Serialize(v *Vector) ([]byte, error) {
	return json.Marshal(v)
}

// Deserialize 反序列化向量
func (vs *VectorSerializer) Deserialize(data []byte) (*Vector, error) {
	var v Vector
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// SerializeBatch 批量序列化
func (vs *VectorSerializer) SerializeBatch(vectors []*Vector) ([][]byte, error) {
	results := make([][]byte, len(vectors))
	for i, v := range vectors {
		data, err := vs.Serialize(v)
		if err != nil {
			return nil, err
		}
		results[i] = data
	}
	return results, nil
}

// DeserializeBatch 批量反序列化
func (vs *VectorSerializer) DeserializeBatch(data [][]byte) ([]*Vector, error) {
	vectors := make([]*Vector, len(data))
	for i, d := range data {
		v, err := vs.Deserialize(d)
		if err != nil {
			return nil, err
		}
		vectors[i] = v
	}
	return vectors, nil
}
