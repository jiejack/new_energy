package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/stretchr/testify/assert"
)

func TestExportHandler_Export_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewExportService(nil, nil, nil)
	handler := NewExportHandler(svc)

	r := gin.New()
	r.POST("/export", handler.Export)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/export", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportHandler_Export_InvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewExportService(nil, nil, nil)
	handler := NewExportHandler(svc)

	r := gin.New()
	r.POST("/export", handler.Export)

	body := service.ExportRequest{
		Type:   "invalid",
		Format: service.ExportFormatExcel,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/export", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportHandler_Export_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewExportService(nil, nil, nil)
	handler := NewExportHandler(svc)

	r := gin.New()
	r.POST("/export", handler.Export)

	body := service.ExportRequest{
		Type:   service.ExportTypeAlarm,
		Format: "invalid",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/export", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportHandler_NewHandler(t *testing.T) {
	svc := service.NewExportService(nil, nil, nil)
	handler := NewExportHandler(svc)
	assert.NotNil(t, handler)
}
