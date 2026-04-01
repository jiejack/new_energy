package knowledge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrKnowledgeBaseNotFound  = errors.New("knowledge base not found")
	ErrDocumentNotFound       = errors.New("document not found")
	ErrKnowledgeBaseExists    = errors.New("knowledge base already exists")
	ErrInvalidKnowledgeBaseID = errors.New("invalid knowledge base ID")
	ErrInvalidDocumentID      = errors.New("invalid document ID")
)

// KnowledgeBaseStatus 知识库状态
type KnowledgeBaseStatus string

const (
	KnowledgeBaseStatusCreating   KnowledgeBaseStatus = "creating"
	KnowledgeBaseStatusActive     KnowledgeBaseStatus = "active"
	KnowledgeBaseStatusInactive   KnowledgeBaseStatus = "inactive"
	KnowledgeBaseStatusDeleting   KnowledgeBaseStatus = "deleting"
	KnowledgeBaseStatusError      KnowledgeBaseStatus = "error"
)

// KnowledgeBase 知识库
type KnowledgeBase struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	IndexName   string                 `json:"index_name"`
	Dimension   int                    `json:"dimension"`
	Status      KnowledgeBaseStatus    `json:"status"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DocumentCount int64                `json:"document_count"`
	ChunkCount    int64                `json:"chunk_count"`
}

// DocumentInfo 文档信息
type DocumentInfo struct {
	ID           string                 `json:"id"`
	KnowledgeBaseID string              `json:"knowledge_base_id"`
	Title        string                 `json:"title"`
	Source       string                 `json:"source"`
	ContentType  string                 `json:"content_type"`
	Size         int64                  `json:"size"`
	ChunkCount   int                    `json:"chunk_count"`
	Status       string                 `json:"status"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// CreateKnowledgeBaseRequest 创建知识库请求
type CreateKnowledgeBaseRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Dimension   int                    `json:"dimension"`
	Metric      string                 `json:"metric"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateKnowledgeBaseRequest 更新知识库请求
type UpdateKnowledgeBaseRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UploadDocumentRequest 上传文档请求
type UploadDocumentRequest struct {
	KnowledgeBaseID string                 `json:"knowledge_base_id"`
	Title           string                 `json:"title"`
	Content         string                 `json:"content"`
	Source          string                 `json:"source,omitempty"`
	ContentType     string                 `json:"content_type,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// KnowledgeBaseStats 知识库统计
type KnowledgeBaseStats struct {
	KnowledgeBaseID string    `json:"knowledge_base_id"`
	DocumentCount   int64     `json:"document_count"`
	ChunkCount      int64     `json:"chunk_count"`
	TotalSize       int64     `json:"total_size"`
	LastUpdated     time.Time `json:"last_updated"`
	VectorCount     int64     `json:"vector_count"`
	IndexSize       int64     `json:"index_size"`
}

// DocumentStore 文档存储接口
type DocumentStore interface {
	Save(ctx context.Context, doc *Document) error
	Get(ctx context.Context, id string) (*Document, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, knowledgeBaseID string, offset, limit int) ([]*Document, error)
	Count(ctx context.Context, knowledgeBaseID string) (int64, error)
}

// InMemoryDocumentStore 内存文档存储
type InMemoryDocumentStore struct {
	docs      map[string]*Document
	kbDocs    map[string][]string // knowledge base ID -> document IDs
	mu        sync.RWMutex
	logger    *zap.Logger
}

// NewInMemoryDocumentStore 创建内存文档存储
func NewInMemoryDocumentStore(logger *zap.Logger) *InMemoryDocumentStore {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &InMemoryDocumentStore{
		docs:   make(map[string]*Document),
		kbDocs: make(map[string][]string),
		logger: logger,
	}
}

// Save 保存文档
func (s *InMemoryDocumentStore) Save(ctx context.Context, doc *Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.docs[doc.ID] = doc
	
	// 更新知识库文档列表
	if _, exists := s.kbDocs[doc.Metadata["knowledge_base_id"].(string)]; !exists {
		s.kbDocs[doc.Metadata["knowledge_base_id"].(string)] = []string{}
	}
	
	// 检查是否已存在
	found := false
	kbID := doc.Metadata["knowledge_base_id"].(string)
	for _, id := range s.kbDocs[kbID] {
		if id == doc.ID {
			found = true
			break
		}
	}
	
	if !found {
		s.kbDocs[kbID] = append(s.kbDocs[kbID], doc.ID)
	}

	return nil
}

// Get 获取文档
func (s *InMemoryDocumentStore) Get(ctx context.Context, id string) (*Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, exists := s.docs[id]
	if !exists {
		return nil, ErrDocumentNotFound
	}

	return doc, nil
}

// Delete 删除文档
func (s *InMemoryDocumentStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, exists := s.docs[id]
	if !exists {
		return ErrDocumentNotFound
	}

	// 从知识库文档列表中删除
	kbID := doc.Metadata["knowledge_base_id"].(string)
	newList := make([]string, 0)
	for _, docID := range s.kbDocs[kbID] {
		if docID != id {
			newList = append(newList, docID)
		}
	}
	s.kbDocs[kbID] = newList

	delete(s.docs, id)
	return nil
}

// List 列出文档
func (s *InMemoryDocumentStore) List(ctx context.Context, knowledgeBaseID string, offset, limit int) ([]*Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docIDs, exists := s.kbDocs[knowledgeBaseID]
	if !exists {
		return []*Document{}, nil
	}

	// 应用偏移和限制
	start := offset
	if start > len(docIDs) {
		start = len(docIDs)
	}

	end := start + limit
	if end > len(docIDs) {
		end = len(docIDs)
	}

	docs := make([]*Document, 0, end-start)
	for i := start; i < end; i++ {
		if doc, exists := s.docs[docIDs[i]]; exists {
			docs = append(docs, doc)
		}
	}

	return docs, nil
}

// Count 统计文档数量
func (s *InMemoryDocumentStore) Count(ctx context.Context, knowledgeBaseID string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docIDs, exists := s.kbDocs[knowledgeBaseID]
	if !exists {
		return 0, nil
	}

	return int64(len(docIDs)), nil
}

// KnowledgeManager 知识库管理器
type KnowledgeManager struct {
	vectorDB      VectorDB
	embedder      *DocumentEmbedder
	retriever     *RAGRetriever
	docStore      DocumentStore
	knowledgeBases map[string]*KnowledgeBase
	mu            sync.RWMutex
	logger        *zap.Logger
}

// NewKnowledgeManager 创建知识库管理器
func NewKnowledgeManager(
	vectorDB VectorDB,
	embedder *DocumentEmbedder,
	retriever *RAGRetriever,
	logger *zap.Logger,
) *KnowledgeManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &KnowledgeManager{
		vectorDB:      vectorDB,
		embedder:      embedder,
		retriever:     retriever,
		docStore:      NewInMemoryDocumentStore(logger),
		knowledgeBases: make(map[string]*KnowledgeBase),
		logger:        logger,
	}
}

// CreateKnowledgeBase 创建知识库
func (km *KnowledgeManager) CreateKnowledgeBase(ctx context.Context, req *CreateKnowledgeBaseRequest) (*KnowledgeBase, error) {
	if req.Name == "" {
		return nil, errors.New("knowledge base name is required")
	}

	if req.Dimension <= 0 {
		return nil, errors.New("dimension must be positive")
	}

	// 生成ID
	kbID := uuid.New().String()
	indexName := fmt.Sprintf("kb_%s", kbID)

	// 创建知识库对象
	kb := &KnowledgeBase{
		ID:          kbID,
		Name:        req.Name,
		Description: req.Description,
		IndexName:   indexName,
		Dimension:   req.Dimension,
		Status:      KnowledgeBaseStatusCreating,
		Config:      req.Config,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 创建向量索引
	indexConfig := &IndexConfig{
		Name:      indexName,
		Dimension: req.Dimension,
		Metric:    req.Metric,
		AutoCreate: true,
		AutoLoad:  true,
	}

	if err := km.vectorDB.CreateIndex(ctx, indexConfig); err != nil {
		km.logger.Error("failed to create index",
			zap.String("kb_id", kbID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create vector index: %w", err)
	}

	// 保存知识库
	km.mu.Lock()
	km.knowledgeBases[kbID] = kb
	km.mu.Unlock()

	// 更新状态
	kb.Status = KnowledgeBaseStatusActive

	km.logger.Info("knowledge base created",
		zap.String("kb_id", kbID),
		zap.String("name", req.Name),
		zap.Int("dimension", req.Dimension),
	)

	return kb, nil
}

// GetKnowledgeBase 获取知识库
func (km *KnowledgeManager) GetKnowledgeBase(ctx context.Context, id string) (*KnowledgeBase, error) {
	if id == "" {
		return nil, ErrInvalidKnowledgeBaseID
	}

	km.mu.RLock()
	defer km.mu.RUnlock()

	kb, exists := km.knowledgeBases[id]
	if !exists {
		return nil, ErrKnowledgeBaseNotFound
	}

	return kb, nil
}

// UpdateKnowledgeBase 更新知识库
func (km *KnowledgeManager) UpdateKnowledgeBase(ctx context.Context, id string, req *UpdateKnowledgeBaseRequest) (*KnowledgeBase, error) {
	if id == "" {
		return nil, ErrInvalidKnowledgeBaseID
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	kb, exists := km.knowledgeBases[id]
	if !exists {
		return nil, ErrKnowledgeBaseNotFound
	}

	// 更新字段
	if req.Name != "" {
		kb.Name = req.Name
	}
	if req.Description != "" {
		kb.Description = req.Description
	}
	if req.Config != nil {
		kb.Config = req.Config
	}
	if req.Metadata != nil {
		kb.Metadata = req.Metadata
	}

	kb.UpdatedAt = time.Now()

	km.logger.Info("knowledge base updated",
		zap.String("kb_id", id),
	)

	return kb, nil
}

// DeleteKnowledgeBase 删除知识库
func (km *KnowledgeManager) DeleteKnowledgeBase(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidKnowledgeBaseID
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	kb, exists := km.knowledgeBases[id]
	if !exists {
		return ErrKnowledgeBaseNotFound
	}

	// 更新状态
	kb.Status = KnowledgeBaseStatusDeleting

	// 删除向量索引
	if err := km.vectorDB.DropIndex(ctx, kb.IndexName); err != nil {
		km.logger.Warn("failed to drop index",
			zap.String("kb_id", id),
			zap.Error(err),
		)
	}

	// 删除文档（实际实现中应该批量删除）
	// 这里简化处理

	// 删除知识库
	delete(km.knowledgeBases, id)

	km.logger.Info("knowledge base deleted",
		zap.String("kb_id", id),
	)

	return nil
}

// ListKnowledgeBases 列出知识库
func (km *KnowledgeManager) ListKnowledgeBases(ctx context.Context, offset, limit int) ([]*KnowledgeBase, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	// 转换为切片
	kbs := make([]*KnowledgeBase, 0, len(km.knowledgeBases))
	for _, kb := range km.knowledgeBases {
		kbs = append(kbs, kb)
	}

	// 应用偏移和限制
	start := offset
	if start > len(kbs) {
		start = len(kbs)
	}

	end := start + limit
	if end > len(kbs) {
		end = len(kbs)
	}

	return kbs[start:end], nil
}

// UploadDocument 上传文档
func (km *KnowledgeManager) UploadDocument(ctx context.Context, req *UploadDocumentRequest) (*DocumentInfo, error) {
	if req.KnowledgeBaseID == "" {
		return nil, ErrInvalidKnowledgeBaseID
	}

	if req.Content == "" {
		return nil, ErrEmptyDocument
	}

	// 检查知识库是否存在
	km.mu.RLock()
	kb, exists := km.knowledgeBases[req.KnowledgeBaseID]
	km.mu.RUnlock()

	if !exists {
		return nil, ErrKnowledgeBaseNotFound
	}

	// 创建文档
	docID := uuid.New().String()
	now := time.Now()
	
	doc := &Document{
		ID:        docID,
		Content:   req.Content,
		Title:     req.Title,
		Source:    req.Source,
		Metadata: map[string]interface{}{
			"knowledge_base_id": req.KnowledgeBaseID,
			"content_type":      req.ContentType,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 合并元数据
	for k, v := range req.Metadata {
		doc.Metadata[k] = v
	}

	// 嵌入文档
	chunks, err := km.embedder.EmbedDocument(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to embed document: %w", err)
	}

	// 索引到向量数据库
	if err := km.retriever.IndexChunks(ctx, kb.IndexName, chunks); err != nil {
		return nil, fmt.Errorf("failed to index document: %w", err)
	}

	// 保存文档
	if err := km.docStore.Save(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	// 更新知识库统计
	km.mu.Lock()
	kb.DocumentCount++
	kb.ChunkCount += int64(len(chunks))
	kb.UpdatedAt = now
	km.mu.Unlock()

	// 返回文档信息
	docInfo := &DocumentInfo{
		ID:              docID,
		KnowledgeBaseID: req.KnowledgeBaseID,
		Title:           req.Title,
		Source:          req.Source,
		ContentType:     req.ContentType,
		Size:           int64(len(req.Content)),
		ChunkCount:     len(chunks),
		Status:         "indexed",
		Metadata:       req.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	km.logger.Info("document uploaded",
		zap.String("doc_id", docID),
		zap.String("kb_id", req.KnowledgeBaseID),
		zap.Int("chunks", len(chunks)),
	)

	return docInfo, nil
}

// DeleteDocument 删除文档
func (km *KnowledgeManager) DeleteDocument(ctx context.Context, knowledgeBaseID, documentID string) error {
	if knowledgeBaseID == "" {
		return ErrInvalidKnowledgeBaseID
	}

	if documentID == "" {
		return ErrInvalidDocumentID
	}

	// 检查知识库是否存在
	km.mu.RLock()
	kb, exists := km.knowledgeBases[knowledgeBaseID]
	km.mu.RUnlock()

	if !exists {
		return ErrKnowledgeBaseNotFound
	}

	// 获取文档
	doc, err := km.docStore.Get(ctx, documentID)
	if err != nil {
		return err
	}

	// 嵌入文档以获取块ID
	chunks, err := km.embedder.EmbedDocument(ctx, doc)
	if err != nil {
		km.logger.Warn("failed to embed document for deletion",
			zap.String("doc_id", documentID),
			zap.Error(err),
		)
	}

	// 从向量数据库删除
	if len(chunks) > 0 {
		chunkIDs := make([]string, len(chunks))
		for i, chunk := range chunks {
			chunkIDs[i] = chunk.ID
		}

		if err := km.retriever.DeleteChunks(ctx, kb.IndexName, chunkIDs); err != nil {
			km.logger.Warn("failed to delete chunks from vector DB",
				zap.String("doc_id", documentID),
				zap.Error(err),
			)
		}
	}

	// 删除文档
	if err := km.docStore.Delete(ctx, documentID); err != nil {
		return err
	}

	// 更新知识库统计
	km.mu.Lock()
	kb.DocumentCount--
	kb.ChunkCount -= int64(len(chunks))
	kb.UpdatedAt = time.Now()
	km.mu.Unlock()

	km.logger.Info("document deleted",
		zap.String("doc_id", documentID),
		zap.String("kb_id", knowledgeBaseID),
	)

	return nil
}

// GetDocument 获取文档
func (km *KnowledgeManager) GetDocument(ctx context.Context, documentID string) (*Document, error) {
	return km.docStore.Get(ctx, documentID)
}

// ListDocuments 列出文档
func (km *KnowledgeManager) ListDocuments(ctx context.Context, knowledgeBaseID string, offset, limit int) ([]*DocumentInfo, error) {
	docs, err := km.docStore.List(ctx, knowledgeBaseID, offset, limit)
	if err != nil {
		return nil, err
	}

	docInfos := make([]*DocumentInfo, len(docs))
	for i, doc := range docs {
		docInfos[i] = &DocumentInfo{
			ID:              doc.ID,
			KnowledgeBaseID: knowledgeBaseID,
			Title:           doc.Title,
			Source:          doc.Source,
			Size:           int64(len(doc.Content)),
			CreatedAt:      doc.CreatedAt,
			UpdatedAt:      doc.UpdatedAt,
		}
	}

	return docInfos, nil
}

// GetKnowledgeBaseStats 获取知识库统计
func (km *KnowledgeManager) GetKnowledgeBaseStats(ctx context.Context, knowledgeBaseID string) (*KnowledgeBaseStats, error) {
	km.mu.RLock()
	kb, exists := km.knowledgeBases[knowledgeBaseID]
	km.mu.RUnlock()

	if !exists {
		return nil, ErrKnowledgeBaseNotFound
	}

	// 获取向量数量
	vectorCount, err := km.vectorDB.Count(ctx, kb.IndexName)
	if err != nil {
		km.logger.Warn("failed to get vector count",
			zap.String("kb_id", knowledgeBaseID),
			zap.Error(err),
		)
	}

	// 获取文档数量
	docCount, err := km.docStore.Count(ctx, knowledgeBaseID)
	if err != nil {
		km.logger.Warn("failed to get document count",
			zap.String("kb_id", knowledgeBaseID),
			zap.Error(err),
		)
	}

	return &KnowledgeBaseStats{
		KnowledgeBaseID: knowledgeBaseID,
		DocumentCount:   docCount,
		ChunkCount:      kb.ChunkCount,
		LastUpdated:     kb.UpdatedAt,
		VectorCount:     vectorCount,
	}, nil
}

// SearchKnowledgeBase 搜索知识库
func (km *KnowledgeManager) SearchKnowledgeBase(ctx context.Context, knowledgeBaseID, query string, topK int) (*RAGContext, error) {
	km.mu.RLock()
	kb, exists := km.knowledgeBases[knowledgeBaseID]
	km.mu.RUnlock()

	if !exists {
		return nil, ErrKnowledgeBaseNotFound
	}

	// 使用RAG检索器搜索
	ragContext, err := km.retriever.Retrieve(ctx, query, kb.IndexName)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return ragContext, nil
}

// BatchUploadDocuments 批量上传文档
func (km *KnowledgeManager) BatchUploadDocuments(ctx context.Context, reqs []*UploadDocumentRequest) ([]*DocumentInfo, []error) {
	results := make([]*DocumentInfo, len(reqs))
	errors := make([]error, len(reqs))

	var wg sync.WaitGroup

	for i, req := range reqs {
		wg.Add(1)
		go func(idx int, r *UploadDocumentRequest) {
			defer wg.Done()
			docInfo, err := km.UploadDocument(ctx, r)
			results[idx] = docInfo
			errors[idx] = err
		}(i, req)
	}

	wg.Wait()

	return results, errors
}

// KnowledgeBaseHealth 知识库健康状态
type KnowledgeBaseHealth struct {
	KnowledgeBaseID string    `json:"knowledge_base_id"`
	Status          string    `json:"status"`
	VectorDBStatus  string    `json:"vector_db_status"`
	DocumentCount   int64     `json:"document_count"`
	VectorCount     int64     `json:"vector_count"`
	LastChecked     time.Time `json:"last_checked"`
	Error           string    `json:"error,omitempty"`
}

// CheckHealth 检查知识库健康状态
func (km *KnowledgeManager) CheckHealth(ctx context.Context, knowledgeBaseID string) (*KnowledgeBaseHealth, error) {
	km.mu.RLock()
	kb, exists := km.knowledgeBases[knowledgeBaseID]
	km.mu.RUnlock()

	if !exists {
		return nil, ErrKnowledgeBaseNotFound
	}

	health := &KnowledgeBaseHealth{
		KnowledgeBaseID: knowledgeBaseID,
		Status:         string(kb.Status),
		LastChecked:    time.Now(),
	}

	// 检查向量数据库
	if err := km.vectorDB.HealthCheck(ctx); err != nil {
		health.VectorDBStatus = "unhealthy"
		health.Error = err.Error()
	} else {
		health.VectorDBStatus = "healthy"
	}

	// 获取统计
	health.DocumentCount = kb.DocumentCount
	if vectorCount, err := km.vectorDB.Count(ctx, kb.IndexName); err == nil {
		health.VectorCount = vectorCount
	}

	return health, nil
}

// KnowledgeManagerStats 管理器统计
type KnowledgeManagerStats struct {
	TotalKnowledgeBases int64 `json:"total_knowledge_bases"`
	TotalDocuments      int64 `json:"total_documents"`
	TotalChunks         int64 `json:"total_chunks"`
	ActiveKnowledgeBases int64 `json:"active_knowledge_bases"`
}

// GetStats 获取管理器统计
func (km *KnowledgeManager) GetStats() *KnowledgeManagerStats {
	km.mu.RLock()
	defer km.mu.RUnlock()

	stats := &KnowledgeManagerStats{
		TotalKnowledgeBases: int64(len(km.knowledgeBases)),
	}

	for _, kb := range km.knowledgeBases {
		stats.TotalDocuments += kb.DocumentCount
		stats.TotalChunks += kb.ChunkCount
		if kb.Status == KnowledgeBaseStatusActive {
			stats.ActiveKnowledgeBases++
		}
	}

	return stats
}

// ExportKnowledgeBase 导出知识库
func (km *KnowledgeManager) ExportKnowledgeBase(ctx context.Context, knowledgeBaseID string) (map[string]interface{}, error) {
	km.mu.RLock()
	kb, exists := km.knowledgeBases[knowledgeBaseID]
	km.mu.RUnlock()

	if !exists {
		return nil, ErrKnowledgeBaseNotFound
	}

	// 获取所有文档
	docs, err := km.docStore.List(ctx, knowledgeBaseID, 0, 10000)
	if err != nil {
		return nil, err
	}

	export := map[string]interface{}{
		"knowledge_base": kb,
		"documents":      docs,
		"exported_at":    time.Now(),
	}

	return export, nil
}

// ImportKnowledgeBase 导入知识库
func (km *KnowledgeManager) ImportKnowledgeBase(ctx context.Context, data map[string]interface{}) (*KnowledgeBase, error) {
	// 解析知识库数据
	kbData, ok := data["knowledge_base"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid knowledge base data")
	}

	// 创建知识库
	req := &CreateKnowledgeBaseRequest{
		Name:        kbData["name"].(string),
		Description: kbData["description"].(string),
		Dimension:   int(kbData["dimension"].(float64)),
	}

	kb, err := km.CreateKnowledgeBase(ctx, req)
	if err != nil {
		return nil, err
	}

	// 导入文档
	docsData, ok := data["documents"].([]interface{})
	if ok {
		for _, docData := range docsData {
			docMap, ok := docData.(map[string]interface{})
			if !ok {
				continue
			}

			uploadReq := &UploadDocumentRequest{
				KnowledgeBaseID: kb.ID,
				Title:          docMap["title"].(string),
				Content:        docMap["content"].(string),
				Source:         docMap["source"].(string),
			}

			if _, err := km.UploadDocument(ctx, uploadReq); err != nil {
				km.logger.Warn("failed to import document",
					zap.Error(err),
				)
			}
		}
	}

	km.logger.Info("knowledge base imported",
		zap.String("kb_id", kb.ID),
	)

	return kb, nil
}
