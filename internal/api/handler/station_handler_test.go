package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/stretchr/testify/assert"
)

func setupStationHandler() (*StationHandler, *RegionHandler, *PointHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	stationSvc := service.NewStationService(nil, nil, nil)
	regionSvc := service.NewRegionService(nil, nil)
	pointSvc := service.NewPointService(nil)
	stationHandler := NewStationHandler(stationSvc)
	regionHandler := NewRegionHandler(regionSvc)
	pointHandler := NewPointHandler(pointSvc)
	r := gin.New()
	return stationHandler, regionHandler, pointHandler, r
}

func TestStationHandler_CreateStation(t *testing.T) {
	stationH, _, _, r := setupStationHandler()
	r.POST("/stations", stationH.CreateStation)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/stations", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestStationHandler_GetStation(t *testing.T) {
	stationH, _, _, r := setupStationHandler()
	r.GET("/stations/:id", stationH.GetStation)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stations/station-001", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStationHandler_ListStations(t *testing.T) {
	stationH, _, _, r := setupStationHandler()
	r.GET("/stations", stationH.ListStations)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stations", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStationHandler_UpdateStation(t *testing.T) {
	stationH, _, _, r := setupStationHandler()
	r.PUT("/stations/:id", stationH.UpdateStation)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/stations/station-001", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStationHandler_DeleteStation(t *testing.T) {
	stationH, _, _, r := setupStationHandler()
	r.DELETE("/stations/:id", stationH.DeleteStation)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/stations/station-001", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRegionHandler_CreateRegion(t *testing.T) {
	_, regionH, _, r := setupStationHandler()
	r.POST("/regions", regionH.CreateRegion)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/regions", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRegionHandler_GetRegion(t *testing.T) {
	_, regionH, _, r := setupStationHandler()
	r.GET("/regions/:id", regionH.GetRegion)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/regions/region-001", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegionHandler_ListRegions(t *testing.T) {
	_, regionH, _, r := setupStationHandler()
	r.GET("/regions", regionH.ListRegions)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/regions", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegionHandler_UpdateRegion(t *testing.T) {
	_, regionH, _, r := setupStationHandler()
	r.PUT("/regions/:id", regionH.UpdateRegion)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/regions/region-001", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegionHandler_DeleteRegion(t *testing.T) {
	_, regionH, _, r := setupStationHandler()
	r.DELETE("/regions/:id", regionH.DeleteRegion)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/regions/region-001", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestPointHandler_CreatePoint(t *testing.T) {
	_, _, pointH, r := setupStationHandler()
	r.POST("/points", pointH.CreatePoint)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/points", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestPointHandler_GetPoint(t *testing.T) {
	_, _, pointH, r := setupStationHandler()
	r.GET("/points/:id", pointH.GetPoint)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/points/point-001", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPointHandler_ListPoints(t *testing.T) {
	_, _, pointH, r := setupStationHandler()
	r.GET("/points", pointH.ListPoints)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/points", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPointHandler_UpdatePoint(t *testing.T) {
	_, _, pointH, r := setupStationHandler()
	r.PUT("/points/:id", pointH.UpdatePoint)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/points/point-001", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPointHandler_DeletePoint(t *testing.T) {
	_, _, pointH, r := setupStationHandler()
	r.DELETE("/points/:id", pointH.DeletePoint)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/points/point-001", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestNewStationHandler(t *testing.T) {
	stationSvc := service.NewStationService(nil, nil, nil)
	handler := NewStationHandler(stationSvc)
	assert.NotNil(t, handler)
}

func TestNewRegionHandler(t *testing.T) {
	regionSvc := service.NewRegionService(nil, nil)
	handler := NewRegionHandler(regionSvc)
	assert.NotNil(t, handler)
}

func TestNewPointHandler(t *testing.T) {
	pointSvc := service.NewPointService(nil)
	handler := NewPointHandler(pointSvc)
	assert.NotNil(t, handler)
}
