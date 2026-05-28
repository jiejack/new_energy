package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAssetDocumentHandler(docRepo *mockAssetDocumentRepo, assetRepo *mockAssetRepo) (*AssetDocumentHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewAssetDocumentService(docRepo, assetRepo)
	handler := NewAssetDocumentHandler(svc)
	r := gin.New()
	return handler, r
}

func TestAssetDocumentHandler_CreateDocument_Success(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDocumentHandler(docRepo, assetRepo)

	r.POST("/assets/documents", handler.CreateDocument)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	docRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AssetDocument")).Return(nil)

	body := map[string]interface{}{
		"asset_id": "asset-001", "document_type": "manual", "title": "Test Manual",
		"file_path": "/docs/manual.pdf", "upload_date": "2024-01-01",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/assets/documents", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAssetDocumentHandler_CreateDocument_BadRequest(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDocumentHandler(docRepo, assetRepo)

	r.POST("/assets/documents", handler.CreateDocument)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/assets/documents", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssetDocumentHandler_GetDocument_Success(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDocumentHandler(docRepo, assetRepo)

	r.GET("/assets/documents/:id", handler.GetDocument)

	docRepo.On("GetByID", mock.Anything, "doc-001").Return(&entity.AssetDocument{ID: "doc-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/documents/doc-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetDocumentHandler_GetDocument_NotFound(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDocumentHandler(docRepo, assetRepo)

	r.GET("/assets/documents/:id", handler.GetDocument)

	docRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/documents/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAssetDocumentHandler_UpdateDocument_Success(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDocumentHandler(docRepo, assetRepo)

	r.PUT("/assets/documents/:id", handler.UpdateDocument)

	existing := &entity.AssetDocument{ID: "doc-001", AssetID: "asset-001"}
	docRepo.On("GetByID", mock.Anything, "doc-001").Return(existing, nil)
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	docRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AssetDocument")).Return(nil)

	body := map[string]interface{}{
		"asset_id": "asset-001", "document_type": "manual", "title": "Updated Manual",
		"file_path": "/docs/manual_v2.pdf", "upload_date": "2024-01-01",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/assets/documents/doc-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetDocumentHandler_UpdateDocument_BadRequest(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDocumentHandler(docRepo, assetRepo)

	r.PUT("/assets/documents/:id", handler.UpdateDocument)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/assets/documents/doc-001", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssetDocumentHandler_DeleteDocument_Success(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDocumentHandler(docRepo, assetRepo)

	r.DELETE("/assets/documents/:id", handler.DeleteDocument)

	docRepo.On("GetByID", mock.Anything, "doc-001").Return(&entity.AssetDocument{ID: "doc-001"}, nil)
	docRepo.On("Delete", mock.Anything, "doc-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/assets/documents/doc-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAssetDocumentHandler_DeleteDocument_Error(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDocumentHandler(docRepo, assetRepo)

	r.DELETE("/assets/documents/:id", handler.DeleteDocument)

	docRepo.On("GetByID", mock.Anything, "doc-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/assets/documents/doc-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAssetDocumentHandler_ListDocuments_Success(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDocumentHandler(docRepo, assetRepo)

	r.GET("/assets/documents", handler.ListDocuments)

	docRepo.On("ListByAssetID", mock.Anything, "", (*string)(nil), 0, 10).Return([]*entity.AssetDocument{{ID: "doc-001"}}, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/documents", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetDocumentHandler_ListDocuments_Error(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDocumentHandler(docRepo, assetRepo)

	r.GET("/assets/documents", handler.ListDocuments)

	docRepo.On("ListByAssetID", mock.Anything, "", (*string)(nil), 0, 10).Return(([]*entity.AssetDocument)(nil), int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/documents", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewAssetDocumentHandler(t *testing.T) {
	docRepo := new(mockAssetDocumentRepo)
	assetRepo := new(mockAssetRepo)
	svc := service.NewAssetDocumentService(docRepo, assetRepo)
	handler := NewAssetDocumentHandler(svc)
	assert.NotNil(t, handler)
}

var _ time.Time = time.Time{}
