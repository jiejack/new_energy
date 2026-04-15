package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/new-energy-monitoring/docs" // swagger docs
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSClient struct {
	conn *websocket.Conn
	send chan []byte
	mu   sync.Mutex
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

var wsClients = make(map[*WSClient]bool)
var wsClientsMu sync.RWMutex

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	client := &WSClient{
		conn: conn,
		send: make(chan []byte, 256),
	}

	wsClientsMu.Lock()
	wsClients[client] = true
	wsClientsMu.Unlock()

	go client.writePump()
	go client.readPump()
}

func (c *WSClient) readPump() {
	defer func() {
		wsClientsMu.Lock()
		delete(wsClients, c)
		wsClientsMu.Unlock()
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket read error: %v\n", err)
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "subscribe-power":
			go c.sendRealtimeData()
		case "subscribe-alarm":
			go c.sendAlarmData()
		}
	}
}

func (c *WSClient) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		c.mu.Lock()
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		c.mu.Unlock()
		if err != nil {
			break
		}
	}
}

func (c *WSClient) sendRealtimeData() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		wsClientsMu.RLock()
		_, ok := wsClients[c]
		wsClientsMu.RUnlock()

		if !ok {
			return
		}

		data := map[string]interface{}{
			"type": "realtime-power",
			"payload": map[string]float64{
				"station1": 1250.5 + float64(time.Now().Unix()%100),
				"station2": 890.3 + float64(time.Now().Unix()%100),
				"station3": 2100.8 + float64(time.Now().Unix()%100),
			},
			"timestamp": time.Now().Unix(),
		}

		msg, _ := json.Marshal(data)
		select {
		case c.send <- msg:
		default:
			return
		}
	}
}

func (c *WSClient) sendAlarmData() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		wsClientsMu.RLock()
		_, ok := wsClients[c]
		wsClientsMu.RUnlock()

		if !ok {
			return
		}

		select {
		case <-ticker.C:
			data := map[string]interface{}{
				"type": "alarm",
				"payload": map[string]interface{}{
					"id":      time.Now().UnixNano(),
					"level":   "warning",
					"title":   "测试告警",
					"message": "这是一条测试告警信息",
					"time":    time.Now().Format("2006-01-02 15:04:05"),
				},
			}

			msg, _ := json.Marshal(data)
			select {
			case c.send <- msg:
			default:
			}
		}
	}
}

// @title 新能源监控系统 API
// @version 1.0
// @description 新能源监控系统RESTful API接口文档，提供区域管理、厂站管理、设备管理、采集点管理、告警管理、数据查询等功能。
// @termsOfService http://swagger.io/terms/

// @contact.name API支持团队
// @contact.url http://www.new-energy-monitoring.com/support
// @contact.email support@new-energy-monitoring.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT认证令牌，格式: Bearer {token}

// @tag.name 区域管理
// @tag.description 区域的增删改查操作

// @tag.name 厂站管理
// @tag.description 厂站的增删改查操作

// @tag.name 设备管理
// @tag.description 设备的增删改查操作

// @tag.name 采集点管理
// @tag.description 采集点的增删改查操作

// @tag.name 告警管理
// @tag.description 告警查询和处理操作

// @tag.name 数据查询
// @tag.description 实时数据和历史数据查询

// @tag.name 控制操作
// @tag.description 遥控和参数设置操作

// @tag.name AI服务
// @tag.description 智能问答和配置建议

// @tag.name 用户管理
// @tag.description 用户认证和管理操作
func main() {
	fmt.Printf("New Energy Monitoring - API Server\n")
	fmt.Printf("Version: %s, Build Time: %s\n\n", Version, BuildTime)

	if err := initConfig(); err != nil {
		panic(fmt.Errorf("failed to init config: %w", err))
	}

	router := setupRouter()
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", viper.GetInt("server.port")),
		Handler:           router,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second, // 防止Slowloris攻击
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("failed to start server: %w", err))
		}
	}()

	fmt.Printf("Server started on port %d\n", viper.GetInt("server.port"))
	fmt.Printf("Swagger UI: http://localhost:%d/swagger/index.html\n", viper.GetInt("server.port"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		panic(fmt.Errorf("server forced to shutdown: %w", err))
	}

	fmt.Println("Server exited")
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")

	return nil
}

func setupRouter() *gin.Engine {
	mode := viper.GetString("server.mode")
	gin.SetMode(mode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check endpoints
	router.GET("/health", healthCheck)
	router.GET("/ready", readyCheck)

	// WebSocket endpoint
	router.GET("/ws", wsHandler)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	{
		// Region management
		api.GET("/regions", listRegions)
		api.POST("/regions", createRegion)
		api.GET("/regions/:id", getRegion)
		api.PUT("/regions/:id", updateRegion)
		api.DELETE("/regions/:id", deleteRegion)

		// Station management
		api.GET("/stations", listStations)
		api.POST("/stations", createStation)
		api.GET("/stations/:id", getStation)
		api.PUT("/stations/:id", updateStation)
		api.DELETE("/stations/:id", deleteStation)
		api.GET("/stations/:id/statistics", getStationStatistics)

		// Device management
		api.GET("/devices", listDevices)
		api.POST("/devices", createDevice)
		api.GET("/devices/:id", getDevice)
		api.PUT("/devices/:id", updateDevice)
		api.DELETE("/devices/:id", deleteDevice)

		// Point management
		api.GET("/points", listPoints)
		api.POST("/points", createPoint)
		api.GET("/points/:id", getPoint)
		api.PUT("/points/:id", updatePoint)
		api.DELETE("/points/:id", deletePoint)

		// Alarm management
		api.GET("/alarms", listAlarms)
		api.GET("/alarms/:id", getAlarm)
		api.PUT("/alarms/:id/ack", ackAlarm)
		api.PUT("/alarms/:id/clear", clearAlarm)
		api.GET("/alarms/statistics", getAlarmStatistics)

		// Alarm rules management
		api.GET("/alarm-rules", listAlarmRules)
		api.POST("/alarm-rules", createAlarmRule)
		api.GET("/alarm-rules/:id", getAlarmRule)
		api.PUT("/alarm-rules/:id", updateAlarmRule)
		api.DELETE("/alarm-rules/:id", deleteAlarmRule)
		api.POST("/alarm-rules/:id/enable", enableAlarmRule)
		api.POST("/alarm-rules/:id/disable", disableAlarmRule)

		// Notification config management
		api.GET("/notification-configs", listNotificationConfigs)
		api.GET("/notification-configs/:type", getNotificationConfig)
		api.PUT("/notification-configs/:type", updateNotificationConfig)
		api.POST("/notification-configs/:type/enable", enableNotificationConfig)
		api.POST("/notification-configs/:type/disable", disableNotificationConfig)
		api.POST("/notification-configs/:type/test", testNotificationConfig)

		// Report management
		api.GET("/reports", generateReport)
		api.GET("/reports/export", exportReport)

		// Operation logs
		api.GET("/operation-logs", listOperationLogs)

		// Data query
		api.GET("/data/realtime", getRealtimeData)
		api.GET("/data/history", getHistoryData)
		api.GET("/data/statistics", getStatistics)

		// Control operations
		api.POST("/control/operate", controlOperate)
		api.POST("/control/setpoint", setPoint)

		// AI services
		api.POST("/ai/qa", aiQA)
		api.POST("/ai/config/suggest", aiConfigSuggest)

		// User management
		api.POST("/auth/login", login)
		api.POST("/auth/logout", logout)
		api.GET("/users", listUsers)
		api.POST("/users", createUser)
		api.GET("/users/:id", getUser)
		api.PUT("/users/:id", updateUser)
		api.DELETE("/users/:id", deleteUser)
		api.PUT("/users/:id/password", changePassword)

		// Profile endpoints
		api.GET("/profile", getProfile)
		api.PUT("/profile", updateProfile)
		api.GET("/profile/preferences", getPreferences)
		api.PUT("/profile/preferences", updatePreferences)
		api.POST("/profile/avatar", uploadAvatar)
	}

	return router
}

// healthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务是否健康运行
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "healthy",
		"version":    Version,
		"build_time": BuildTime,
		"timestamp":  time.Now().Unix(),
	})
}

// readyCheck 就绪检查
// @Summary 就绪检查
// @Description 检查服务是否就绪
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response
// @Router /ready [get]
func readyCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ready": true,
	})
}

// listRegions 获取区域列表
// @Summary 获取区域列表
// @Description 获取所有区域的列表，支持按父区域ID过滤
// @Tags 区域管理
// @Accept json
// @Produce json
// @Param parent_id query string false "父区域ID"
// @Success 200 {object} dto.Response{data=[]dto.RegionResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /regions [get]
func listRegions(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      []dto.RegionResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// createRegion 创建区域
// @Summary 创建区域
// @Description 创建新的区域
// @Tags 区域管理
// @Accept json
// @Produce json
// @Param region body dto.CreateRegionRequest true "区域信息"
// @Success 201 {object} dto.Response{data=dto.RegionResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /regions [post]
func createRegion(c *gin.Context) {
	c.JSON(http.StatusCreated, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.RegionResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// getRegion 获取区域详情
// @Summary 获取区域详情
// @Description 根据ID获取区域详细信息
// @Tags 区域管理
// @Accept json
// @Produce json
// @Param id path string true "区域ID"
// @Success 200 {object} dto.Response{data=dto.RegionResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Router /regions/{id} [get]
func getRegion(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.RegionResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// updateRegion 更新区域
// @Summary 更新区域
// @Description 更新区域信息
// @Tags 区域管理
// @Accept json
// @Produce json
// @Param id path string true "区域ID"
// @Param region body dto.UpdateRegionRequest true "区域信息"
// @Success 200 {object} dto.Response{data=dto.RegionResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /regions/{id} [put]
func updateRegion(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.RegionResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// deleteRegion 删除区域
// @Summary 删除区域
// @Description 删除指定区域
// @Tags 区域管理
// @Accept json
// @Produce json
// @Param id path string true "区域ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /regions/{id} [delete]
func deleteRegion(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// listStations 获取厂站列表
// @Summary 获取厂站列表
// @Description 获取所有厂站的列表，支持分页和过滤
// @Tags 厂站管理
// @Accept json
// @Produce json
// @Param sub_region_id query string false "子区域ID"
// @Param type query string false "厂站类型"
// @Param status query int false "状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} dto.PagedResponse{data=[]dto.StationResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /stations [get]
func listStations(c *gin.Context) {
	now := time.Now()
	stations := []dto.StationResponse{
		{
			ID:             "station_001",
			Code:           "BJ-CY-001",
			Name:           "北京朝阳光伏电站",
			Type:           "solar",
			SubRegionID:    "region_001",
			Capacity:       5000,
			VoltageLevel:   "35kV",
			Longitude:      116.4074,
			Latitude:       39.9042,
			Address:        "北京市朝阳区XXX路XXX号",
			Status:         1,
			CommissionDate: &now,
			Description:    "北京市朝阳区大型光伏电站",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "station_002",
			Code:           "SH-PD-001",
			Name:           "上海浦东风电场",
			Type:           "wind",
			SubRegionID:    "region_002",
			Capacity:       10000,
			VoltageLevel:   "110kV",
			Longitude:      121.5441,
			Latitude:       31.2304,
			Address:        "上海市浦东新区XXX路XXX号",
			Status:         1,
			CommissionDate: &now,
			Description:    "上海浦东大型风电场",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "station_003",
			Code:           "GZ-PY-001",
			Name:           "广州番禺储能站",
			Type:           "storage",
			SubRegionID:    "region_003",
			Capacity:       2000,
			VoltageLevel:   "10kV",
			Longitude:      113.3647,
			Latitude:       22.9375,
			Address:        "广州市番禺区XXX路XXX号",
			Status:         1,
			CommissionDate: &now,
			Description:    "广州番禺储能示范项目",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "station_004",
			Code:           "SZ-NS-001",
			Name:           "深圳南山光伏电站",
			Type:           "solar",
			SubRegionID:    "region_003",
			Capacity:       8000,
			VoltageLevel:   "35kV",
			Longitude:      113.9308,
			Latitude:       22.5332,
			Address:        "深圳市南山区XXX路XXX号",
			Status:         1,
			CommissionDate: &now,
			Description:    "深圳南山大型光伏电站",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "station_005",
			Code:           "HZ-XH-001",
			Name:           "杭州西湖光伏电站",
			Type:           "solar",
			SubRegionID:    "region_002",
			Capacity:       3500,
			VoltageLevel:   "35kV",
			Longitude:      120.1551,
			Latitude:       30.2741,
			Address:        "杭州市西湖区XXX路XXX号",
			Status:         1,
			CommissionDate: &now,
			Description:    "杭州西湖分布式光伏项目",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "station_006",
			Code:           "NJ-PJ-001",
			Name:           "南京浦口风电场",
			Type:           "wind",
			SubRegionID:    "region_002",
			Capacity:       6000,
			VoltageLevel:   "110kV",
			Longitude:      118.7969,
			Latitude:       32.0603,
			Address:        "南京市浦口区XXX路XXX号",
			Status:         0,
			CommissionDate: &now,
			Description:    "南京浦口风电场",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	page := 1
	pageSize := 100
	c.JSON(http.StatusOK, dto.PagedResponse{
		Code:      0,
		Message:   "success",
		Data:      stations,
		Timestamp: time.Now().Unix(),
		Page:      page,
		PageSize:  pageSize,
		Total:     int64(len(stations)),
	})
}

// createStation 创建厂站
// @Summary 创建厂站
// @Description 创建新的厂站
// @Tags 厂站管理
// @Accept json
// @Produce json
// @Param station body dto.CreateStationRequest true "厂站信息"
// @Success 201 {object} dto.Response{data=dto.StationResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /stations [post]
func createStation(c *gin.Context) {
	c.JSON(http.StatusCreated, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.StationResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// getStation 获取厂站详情
// @Summary 获取厂站详情
// @Description 根据ID获取厂站详细信息
// @Tags 厂站管理
// @Accept json
// @Produce json
// @Param id path string true "厂站ID"
// @Success 200 {object} dto.Response{data=dto.StationResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Router /stations/{id} [get]
func getStation(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.StationResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// updateStation 更新厂站
// @Summary 更新厂站
// @Description 更新厂站信息
// @Tags 厂站管理
// @Accept json
// @Produce json
// @Param id path string true "厂站ID"
// @Param station body dto.UpdateStationRequest true "厂站信息"
// @Success 200 {object} dto.Response{data=dto.StationResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /stations/{id} [put]
func updateStation(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.StationResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// deleteStation 删除厂站
// @Summary 删除厂站
// @Description 删除指定厂站
// @Tags 厂站管理
// @Accept json
// @Produce json
// @Param id path string true "厂站ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /stations/{id} [delete]
func deleteStation(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// getStationStatistics 获取电站统计信息
// @Summary 获取电站统计信息
// @Description 获取指定电站的设备统计、功率、发电量等信息
// @Tags 厂站管理
// @Accept json
// @Produce json
// @Param id path string true "电站ID"
// @Success 200 {object} dto.Response{data=dto.StationStatisticsResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Router /stations/{id}/statistics [get]
func getStationStatistics(c *gin.Context) {
	stationID := c.Param("id")
	stationStats := map[string]interface{}{
		"station_001": map[string]interface{}{
			"deviceCount":        25,
			"onlineDeviceCount":  24,
			"offlineDeviceCount": 1,
			"alarmCount":         5,
			"power":              4500,
			"energy":             28500,
		},
		"station_002": map[string]interface{}{
			"deviceCount":        20,
			"onlineDeviceCount":  20,
			"offlineDeviceCount": 0,
			"alarmCount":         3,
			"power":              8500,
			"energy":             51000,
		},
		"station_003": map[string]interface{}{
			"deviceCount":        10,
			"onlineDeviceCount":  10,
			"offlineDeviceCount": 0,
			"alarmCount":         2,
			"power":              1800,
			"energy":             10800,
		},
		"station_004": map[string]interface{}{
			"deviceCount":        30,
			"onlineDeviceCount":  28,
			"offlineDeviceCount": 2,
			"alarmCount":         4,
			"power":              7200,
			"energy":             43200,
		},
		"station_005": map[string]interface{}{
			"deviceCount":        15,
			"onlineDeviceCount":  15,
			"offlineDeviceCount": 0,
			"alarmCount":         1,
			"power":              3200,
			"energy":             19200,
		},
		"station_006": map[string]interface{}{
			"deviceCount":        12,
			"onlineDeviceCount":  0,
			"offlineDeviceCount": 12,
			"alarmCount":         8,
			"power":              0,
			"energy":             0,
		},
	}

	if stats, ok := stationStats[stationID]; ok {
		c.JSON(http.StatusOK, dto.Response{
			Code:      0,
			Message:   "success",
			Data:      stats,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"deviceCount":        0,
			"onlineDeviceCount":  0,
			"offlineDeviceCount": 0,
			"alarmCount":         0,
			"power":              0,
			"energy":             0,
		},
		Timestamp: time.Now().Unix(),
	})
}

// listDevices 获取设备列表
// @Summary 获取设备列表
// @Description 获取所有设备的列表，支持分页和过滤
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param station_id query string false "厂站ID"
// @Param type query string false "设备类型"
// @Param status query int false "状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} dto.PagedResponse{data=[]dto.DeviceResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /devices [get]
func listDevices(c *gin.Context) {
	now := time.Now()
	devices := []dto.DeviceResponse{
		{
			ID:           "device_001",
			Code:         "INV-BJ-001",
			Name:         "逆变器 #01",
			Type:         "inverter",
			StationID:    "station_001",
			Manufacturer: "华为",
			Model:        "SUN2000-100KTL",
			SerialNumber: "SN2101012345",
			RatedPower:   100,
			RatedVoltage: 380,
			RatedCurrent: 150,
			Protocol:     "modbus",
			IPAddress:    "192.168.1.101",
			Port:         502,
			SlaveID:      1,
			Status:       1,
			LastOnline:   &now,
			InstallDate:  &now,
			WarrantyDate: &now,
			Description:  "北京朝阳1号光伏逆变器",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "device_002",
			Code:         "INV-BJ-002",
			Name:         "逆变器 #02",
			Type:         "inverter",
			StationID:    "station_001",
			Manufacturer: "华为",
			Model:        "SUN2000-100KTL",
			SerialNumber: "SN2101012346",
			RatedPower:   100,
			RatedVoltage: 380,
			RatedCurrent: 150,
			Protocol:     "modbus",
			IPAddress:    "192.168.1.102",
			Port:         502,
			SlaveID:      2,
			Status:       1,
			LastOnline:   &now,
			InstallDate:  &now,
			WarrantyDate: &now,
			Description:  "北京朝阳2号光伏逆变器",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "device_003",
			Code:         "WT-SH-001",
			Name:         "风机 #01",
			Type:         "wind_turbine",
			StationID:    "station_002",
			Manufacturer: "金风科技",
			Model:        "GW140-2500",
			SerialNumber: "JF2102034567",
			RatedPower:   2500,
			RatedVoltage: 690,
			RatedCurrent: 2100,
			Protocol:     "IEC104",
			IPAddress:    "192.168.2.101",
			Port:         2404,
			SlaveID:      1,
			Status:       1,
			LastOnline:   &now,
			InstallDate:  &now,
			WarrantyDate: &now,
			Description:  "上海浦东1号风机",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "device_004",
			Code:         "WT-SH-002",
			Name:         "风机 #02",
			Type:         "wind_turbine",
			StationID:    "station_002",
			Manufacturer: "金风科技",
			Model:        "GW140-2500",
			SerialNumber: "JF2102034568",
			RatedPower:   2500,
			RatedVoltage: 690,
			RatedCurrent: 2100,
			Protocol:     "IEC104",
			IPAddress:    "192.168.2.102",
			Port:         2404,
			SlaveID:      2,
			Status:       1,
			LastOnline:   &now,
			InstallDate:  &now,
			WarrantyDate: &now,
			Description:  "上海浦东2号风机",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "device_005",
			Code:         "BMS-GZ-001",
			Name:         "储能BMS #01",
			Type:         "bms",
			StationID:    "station_003",
			Manufacturer: "宁德时代",
			Model:        "CATL-BMS-500",
			SerialNumber: "CATL2103045678",
			RatedPower:   500,
			RatedVoltage: 768,
			RatedCurrent: 650,
			Protocol:     "CAN",
			IPAddress:    "192.168.3.101",
			Port:         0,
			SlaveID:      0,
			Status:       1,
			LastOnline:   &now,
			InstallDate:  &now,
			WarrantyDate: &now,
			Description:  "广州番禺储能BMS系统",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "device_006",
			Code:         "INV-SZ-001",
			Name:         "逆变器 #01",
			Type:         "inverter",
			StationID:    "station_004",
			Manufacturer: "阳光电源",
			Model:        "SG125HX",
			SerialNumber: "SG2104067890",
			RatedPower:   125,
			RatedVoltage: 540,
			RatedCurrent: 138,
			Protocol:     "modbus",
			IPAddress:    "192.168.4.101",
			Port:         502,
			SlaveID:      1,
			Status:       1,
			LastOnline:   &now,
			InstallDate:  &now,
			WarrantyDate: &now,
			Description:  "深圳南山1号逆变器",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "device_007",
			Code:         "INV-HZ-001",
			Name:         "逆变器 #01",
			Type:         "inverter",
			StationID:    "station_005",
			Manufacturer: "华为",
			Model:        "SUN2000-50KTL",
			SerialNumber: "SN2105078901",
			RatedPower:   50,
			RatedVoltage: 380,
			RatedCurrent: 76,
			Protocol:     "modbus",
			IPAddress:    "192.168.5.101",
			Port:         502,
			SlaveID:      1,
			Status:       1,
			LastOnline:   &now,
			InstallDate:  &now,
			WarrantyDate: &now,
			Description:  "杭州西湖1号逆变器",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "device_008",
			Code:         "WT-NJ-001",
			Name:         "风机 #01",
			Type:         "wind_turbine",
			StationID:    "station_006",
			Manufacturer: "明阳智能",
			Model:        "MY-3.0MW",
			SerialNumber: "MY2106089012",
			RatedPower:   3000,
			RatedVoltage: 690,
			RatedCurrent: 2500,
			Protocol:     "IEC104",
			IPAddress:    "192.168.6.101",
			Port:         2404,
			SlaveID:      1,
			Status:       0,
			LastOnline:   &now,
			InstallDate:  &now,
			WarrantyDate: &now,
			Description:  "南京浦口1号风机",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	c.JSON(http.StatusOK, dto.PagedResponse{
		Code:      0,
		Message:   "success",
		Data:      devices,
		Timestamp: time.Now().Unix(),
		Total:     int64(len(devices)),
		Page:      1,
		PageSize:  20,
	})
}

// createDevice 创建设备
// @Summary 创建设备
// @Description 创建新的设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param device body dto.CreateDeviceRequest true "设备信息"
// @Success 201 {object} dto.Response{data=dto.DeviceResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /devices [post]
func createDevice(c *gin.Context) {
	c.JSON(http.StatusCreated, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.DeviceResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// getDevice 获取设备详情
// @Summary 获取设备详情
// @Description 根据ID获取设备详细信息
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Success 200 {object} dto.Response{data=dto.DeviceResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Router /devices/{id} [get]
func getDevice(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.DeviceResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// updateDevice 更新设备
// @Summary 更新设备
// @Description 更新设备信息
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Param device body dto.UpdateDeviceRequest true "设备信息"
// @Success 200 {object} dto.Response{data=dto.DeviceResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /devices/{id} [put]
func updateDevice(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.DeviceResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// deleteDevice 删除设备
// @Summary 删除设备
// @Description 删除指定设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /devices/{id} [delete]
func deleteDevice(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// listPoints 获取采集点列表
// @Summary 获取采集点列表
// @Description 获取所有采集点的列表，支持分页和过滤
// @Tags 采集点管理
// @Accept json
// @Produce json
// @Param device_id query string false "设备ID"
// @Param type query string false "点类型"
// @Param status query int false "状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} dto.PagedResponse{data=[]dto.PointResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /points [get]
func listPoints(c *gin.Context) {
	c.JSON(http.StatusOK, dto.PagedResponse{
		Code:      0,
		Message:   "success",
		Data:      []dto.PointResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// createPoint 创建采集点
// @Summary 创建采集点
// @Description 创建新的采集点
// @Tags 采集点管理
// @Accept json
// @Produce json
// @Param point body dto.CreatePointRequest true "采集点信息"
// @Success 201 {object} dto.Response{data=dto.PointResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /points [post]
func createPoint(c *gin.Context) {
	c.JSON(http.StatusCreated, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.PointResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// getPoint 获取采集点详情
// @Summary 获取采集点详情
// @Description 根据ID获取采集点详细信息
// @Tags 采集点管理
// @Accept json
// @Produce json
// @Param id path string true "采集点ID"
// @Success 200 {object} dto.Response{data=dto.PointResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Router /points/{id} [get]
func getPoint(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.PointResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// updatePoint 更新采集点
// @Summary 更新采集点
// @Description 更新采集点信息
// @Tags 采集点管理
// @Accept json
// @Produce json
// @Param id path string true "采集点ID"
// @Param point body dto.UpdatePointRequest true "采集点信息"
// @Success 200 {object} dto.Response{data=dto.PointResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /points/{id} [put]
func updatePoint(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.PointResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// deletePoint 删除采集点
// @Summary 删除采集点
// @Description 删除指定采集点
// @Tags 采集点管理
// @Accept json
// @Produce json
// @Param id path string true "采集点ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /points/{id} [delete]
func deletePoint(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// listAlarms 获取告警列表
// @Summary 获取告警列表
// @Description 获取所有告警的列表，支持分页和过滤
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param station_id query string false "厂站ID"
// @Param device_id query string false "设备ID"
// @Param level query int false "告警级别"
// @Param status query int false "状态"
// @Param type query string false "告警类型"
// @Param start_time query int false "开始时间戳"
// @Param end_time query int false "结束时间戳"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} dto.PagedResponse{data=[]dto.AlarmResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /alarms [get]
func listAlarms(c *gin.Context) {
	now := time.Now()
	alarms := []dto.AlarmResponse{
		{
			ID:          "alarm_001",
			PointID:     "point_001",
			DeviceID:    "device_001",
			StationID:   "station_001",
			Type:        "limit",
			Level:       3,
			Title:       "逆变器温度过高",
			Message:     "逆变器#01温度为87.5°C，已超过85°C阈值",
			Value:       87.5,
			Threshold:   85.0,
			Status:      1,
			TriggeredAt: now.Add(-30 * time.Minute),
		},
		{
			ID:          "alarm_002",
			PointID:     "point_002",
			DeviceID:    "device_002",
			StationID:   "station_002",
			Type:        "limit",
			Level:       2,
			Title:       "功率下降告警",
			Message:     "风电场当前功率为45%，低于正常阈值50%",
			Value:       45.0,
			Threshold:   50.0,
			Status:      1,
			TriggeredAt: now.Add(-1 * time.Hour),
		},
		{
			ID:          "alarm_003",
			PointID:     "point_003",
			DeviceID:    "device_003",
			StationID:   "station_003",
			Type:        "communication",
			Level:       2,
			Title:       "通讯中断",
			Message:     "储能站BMS通讯中断超过5分钟",
			Value:       0,
			Threshold:   0,
			Status:      1,
			TriggeredAt: now.Add(-2 * time.Hour),
		},
		{
			ID:          "alarm_004",
			PointID:     "point_004",
			DeviceID:    "device_004",
			StationID:   "station_004",
			Type:        "limit",
			Level:       1,
			Title:       "辐照度异常",
			Message:     "光伏电站瞬时辐照度低于100W/m²，可能存在遮挡",
			Value:       85.0,
			Threshold:   100.0,
			Status:      0,
			TriggeredAt: now.Add(-3 * time.Hour),
		},
		{
			ID:          "alarm_005",
			PointID:     "point_005",
			DeviceID:    "device_005",
			StationID:   "station_005",
			Type:        "limit",
			Level:       2,
			Title:       "电池SOC过低",
			Message:     "储能电池SOC为15%，低于正常阈值20%",
			Value:       15.0,
			Threshold:   20.0,
			Status:      1,
			TriggeredAt: now.Add(-45 * time.Minute),
		},
		{
			ID:          "alarm_006",
			PointID:     "point_001",
			DeviceID:    "device_001",
			StationID:   "station_001",
			Type:        "quality",
			Level:       1,
			Title:       "电能质量告警",
			Message:     "电站功率因数低于0.9",
			Value:       0.85,
			Threshold:   0.9,
			Status:      0,
			TriggeredAt: now.Add(-4 * time.Hour),
		},
		{
			ID:          "alarm_007",
			PointID:     "point_006",
			DeviceID:    "device_006",
			StationID:   "station_006",
			Type:        "limit",
			Level:       3,
			Title:       "风机转速过高",
			Message:     "风机#03转速为25rpm，超过保护阈值20rpm",
			Value:       25.0,
			Threshold:   20.0,
			Status:      1,
			TriggeredAt: now.Add(-15 * time.Minute),
		},
		{
			ID:          "alarm_008",
			PointID:     "point_007",
			DeviceID:    "device_007",
			StationID:   "station_002",
			Type:        "communication",
			Level:       2,
			Title:       "风机通讯故障",
			Message:     "风电场风机#05通讯中断",
			Value:       0,
			Threshold:   0,
			Status:      1,
			TriggeredAt: now.Add(-50 * time.Minute),
		},
	}

	c.JSON(http.StatusOK, dto.PagedResponse{
		Code:      0,
		Message:   "success",
		Data:      alarms,
		Timestamp: time.Now().Unix(),
		Page:      1,
		PageSize:  20,
		Total:     int64(len(alarms)),
	})
}

// getAlarm 获取告警详情
// @Summary 获取告警详情
// @Description 根据ID获取告警详细信息
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param id path string true "告警ID"
// @Success 200 {object} dto.Response{data=dto.AlarmResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Router /alarms/{id} [get]
func getAlarm(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.AlarmResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// ackAlarm 确认告警
// @Summary 确认告警
// @Description 确认指定告警
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param id path string true "告警ID"
// @Param request body dto.AckAlarmRequest true "确认信息"
// @Success 200 {object} dto.Response{data=dto.AlarmResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /alarms/{id}/ack [put]
func ackAlarm(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.AlarmResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// clearAlarm 清除告警
// @Summary 清除告警
// @Description 清除指定告警
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param id path string true "告警ID"
// @Param request body dto.ClearAlarmRequest true "清除信息"
// @Success 200 {object} dto.Response{data=dto.AlarmResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /alarms/{id}/clear [put]
func clearAlarm(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.AlarmResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// getAlarmStatistics 获取告警统计
// @Summary 获取告警统计
// @Description 获取告警统计数据
// @Tags 告警管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.AlarmStatisticsResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /alarms/statistics [get]
func getAlarmStatistics(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: dto.AlarmStatisticsResponse{
			Total:        156,
			Active:       23,
			Acknowledged: 45,
			Cleared:      88,
			ByLevel: map[int]int64{
				1: 45,
				2: 78,
				3: 33,
			},
			ByType: map[string]int64{
				"limit":         89,
				"communication": 42,
				"quality":       25,
			},
		},
		Timestamp: time.Now().Unix(),
	})
}

// getRealtimeData 获取实时数据
// @Summary 获取实时数据
// @Description 获取指定采集点的实时数据
// @Tags 数据查询
// @Accept json
// @Produce json
// @Param point_ids query string true "采集点ID列表，逗号分隔"
// @Success 200 {object} dto.Response{data=[]dto.RealtimeDataResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /data/realtime [get]
func getRealtimeData(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      []dto.RealtimeDataResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// getHistoryData 获取历史数据
// @Summary 获取历史数据
// @Description 获取指定采集点的历史数据
// @Tags 数据查询
// @Accept json
// @Produce json
// @Param point_id query string true "采集点ID"
// @Param start_time query int true "开始时间戳"
// @Param end_time query int true "结束时间戳"
// @Param interval query int false "采样间隔(秒)"
// @Success 200 {object} dto.Response{data=[]dto.HistoryDataResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /data/history [get]
func getHistoryData(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      []dto.HistoryDataResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// getStatistics 获取统计数据
// @Summary 获取统计数据
// @Description 获取统计数据
// @Tags 数据查询
// @Accept json
// @Produce json
// @Param station_id query string false "厂站ID"
// @Param type query string false "统计类型"
// @Param start_time query int false "开始时间戳"
// @Param end_time query int false "结束时间戳"
// @Success 200 {object} dto.Response{data=dto.StatisticsResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /data/statistics [get]
func getStatistics(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: []dto.StatisticsResponse{
			{
				StationID:     "station_001",
				TotalPower:    4500,
				DailyEnergy:   28500,
				MonthlyEnergy: 856000,
				PR:            0.85,
				Availability:  0.98,
				Date:          "2024-03-01",
			},
			{
				StationID:     "station_002",
				TotalPower:    8500,
				DailyEnergy:   51000,
				MonthlyEnergy: 1530000,
				PR:            0.82,
				Availability:  0.96,
				Date:          "2024-03-01",
			},
			{
				StationID:     "station_003",
				TotalPower:    1800,
				DailyEnergy:   10800,
				MonthlyEnergy: 324000,
				PR:            0.90,
				Availability:  0.99,
				Date:          "2024-03-01",
			},
			{
				StationID:     "station_004",
				TotalPower:    7200,
				DailyEnergy:   43200,
				MonthlyEnergy: 1296000,
				PR:            0.84,
				Availability:  0.97,
				Date:          "2024-03-01",
			},
			{
				StationID:     "station_005",
				TotalPower:    3200,
				DailyEnergy:   19200,
				MonthlyEnergy: 576000,
				PR:            0.86,
				Availability:  0.98,
				Date:          "2024-03-01",
			},
			{
				StationID:     "station_006",
				TotalPower:    0,
				DailyEnergy:   0,
				MonthlyEnergy: 0,
				PR:            0,
				Availability:  0,
				Date:          "2024-03-01",
			},
		},
		Timestamp: time.Now().Unix(),
	})
}

// controlOperate 遥控操作
// @Summary 遥控操作
// @Description 执行遥控操作
// @Tags 控制操作
// @Accept json
// @Produce json
// @Param request body dto.ControlOperateRequest true "遥控操作信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /control/operate [post]
func controlOperate(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: time.Now().Unix(),
	})
}

// setPoint 参数设置
// @Summary 参数设置
// @Description 执行参数设置操作
// @Tags 控制操作
// @Accept json
// @Produce json
// @Param request body dto.SetPointRequest true "参数设置信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /control/setpoint [post]
func setPoint(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: time.Now().Unix(),
	})
}

// aiQA AI问答
// @Summary AI问答
// @Description 智能问答服务
// @Tags AI服务
// @Accept json
// @Produce json
// @Param request body dto.AIQARequest true "问答请求"
// @Success 200 {object} dto.Response{data=dto.AIQAResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /ai/qa [post]
func aiQA(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.AIQAResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// aiConfigSuggest AI配置建议
// @Summary AI配置建议
// @Description 智能配置建议服务
// @Tags AI服务
// @Accept json
// @Produce json
// @Param request body dto.AIConfigSuggestRequest true "配置建议请求"
// @Success 200 {object} dto.Response{data=dto.AIConfigSuggestResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /ai/config/suggest [post]
func aiConfigSuggest(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.AIConfigSuggestResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// login 用户登录
// @Summary 用户登录
// @Description 用户登录认证
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} dto.Response{data=dto.LoginResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func login(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.LoginResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// logout 用户登出
// @Summary 用户登出
// @Description 用户登出
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response
// @Router /auth/logout [post]
func logout(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: time.Now().Unix(),
	})
}

// listUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取所有用户的列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=[]dto.UserResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [get]
func listUsers(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      []dto.UserResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// createUser 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "用户信息"
// @Success 201 {object} dto.Response{data=dto.UserResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [post]
func createUser(c *gin.Context) {
	c.JSON(http.StatusCreated, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.UserResponse{},
		Timestamp: time.Now().Unix(),
	})
}

// getUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} dto.Response{data=dto.UserResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/{id} [get]
func getUser(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.UserResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// updateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param user body dto.UpdateUserRequest true "用户信息"
// @Success 200 {object} dto.Response{data=dto.UserResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/{id} [put]
func updateUser(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      dto.UserResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// deleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/{id} [delete]
func deleteUser(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// changePassword 修改密码
// @Summary 修改密码
// @Description 修改用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param request body dto.ChangePasswordRequest true "密码信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/{id}/password [put]
func changePassword(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: time.Now().Unix(),
	})
}

// getProfile 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.ProfileResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Router /profile [get]
func getProfile(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: dto.ProfileResponse{
			ID:         "1",
			Username:   "admin",
			Nickname:   "系统管理员",
			Email:      "admin@example.com",
			Phone:      "13800138000",
			Avatar:     "https://api.dicebear.com/7.x/avataaars/svg?seed=admin",
			Role:       "admin",
			Status:     1,
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		},
		Timestamp: time.Now().Unix(),
	})
}

// updateProfile 更新当前用户信息
// @Summary 更新当前用户信息
// @Description 更新当前登录用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param profile body dto.UpdateProfileRequest true "用户信息"
// @Success 200 {object} dto.Response{data=dto.ProfileResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Router /profile [put]
func updateProfile(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: dto.ProfileResponse{
			ID:         "1",
			Username:   "admin",
			Nickname:   "系统管理员",
			Email:      "admin@example.com",
			Phone:      "13800138000",
			Avatar:     "https://api.dicebear.com/7.x/avataaars/svg?seed=admin",
			Role:       "admin",
			Status:     1,
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		},
		Timestamp: time.Now().Unix(),
	})
}

// getPreferences 获取偏好设置
// @Summary 获取偏好设置
// @Description 获取当前用户的偏好设置
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=dto.PreferencesResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Router /profile/preferences [get]
func getPreferences(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: dto.PreferencesResponse{
			Theme:           "light",
			Language:        "zh-CN",
			Timezone:        "Asia/Shanghai",
			NotifyEnabled:   true,
			NotifyTypes:     []string{"alarm", "system"},
			DashboardLayout: "default",
		},
		Timestamp: time.Now().Unix(),
	})
}

// updatePreferences 更新偏好设置
// @Summary 更新偏好设置
// @Description 更新当前用户的偏好设置
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param preferences body dto.UpdatePreferencesRequest true "偏好设置"
// @Success 200 {object} dto.Response{data=dto.PreferencesResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Router /profile/preferences [put]
func updatePreferences(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: dto.PreferencesResponse{
			Theme:           "light",
			Language:        "zh-CN",
			Timezone:        "Asia/Shanghai",
			NotifyEnabled:   true,
			NotifyTypes:     []string{"alarm", "system"},
			DashboardLayout: "default",
		},
		Timestamp: time.Now().Unix(),
	})
}

// uploadAvatar 上传头像
// @Summary 上传头像
// @Description 上传用户头像
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param avatar body dto.UploadAvatarRequest true "头像信息"
// @Success 200 {object} dto.Response{data=dto.UploadAvatarResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Router /profile/avatar [post]
func uploadAvatar(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: dto.UploadAvatarResponse{
			Avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=admin",
		},
		Timestamp: time.Now().Unix(),
	})
}

// AlarmRuleResponse 告警规则响应
type AlarmRuleResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Type           string   `json:"type"`
	Level          int      `json:"level"`
	Condition      string   `json:"condition"`
	Threshold      float64  `json:"threshold"`
	Duration       int      `json:"duration"`
	PointID        *string  `json:"point_id,omitempty"`
	DeviceID       *string  `json:"device_id,omitempty"`
	StationID      *string  `json:"station_id,omitempty"`
	NotifyChannels []string `json:"notify_channels"`
	NotifyUsers    []string `json:"notify_users"`
	Status         int      `json:"status"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

// listAlarmRules 获取告警规则列表
// @Summary 获取告警规则列表
// @Description 获取所有告警规则的列表，支持分页和过滤
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param type query string false "规则类型"
// @Param level query int false "告警级别"
// @Param status query int false "状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} dto.PagedResponse{data=[]AlarmRuleResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /alarm-rules [get]
func listAlarmRules(c *gin.Context) {
	rules := []AlarmRuleResponse{
		{
			ID:             "rule_001",
			Name:           "温度超限告警",
			Description:    "设备温度超过设定阈值时触发",
			Type:           "limit",
			Level:          3,
			Condition:      "temperature > threshold",
			Threshold:      80,
			Duration:       60,
			NotifyChannels: []string{"email", "sms"},
			NotifyUsers:    []string{"admin", "operator"},
			Status:         1,
			CreatedAt:      time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05"),
			UpdatedAt:      time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			ID:             "rule_002",
			Name:           "功率异常告警",
			Description:    "设备功率低于正常范围时触发",
			Type:           "limit",
			Level:          2,
			Condition:      "power < threshold",
			Threshold:      0.8,
			Duration:       120,
			NotifyChannels: []string{"email"},
			NotifyUsers:    []string{"operator"},
			Status:         1,
			CreatedAt:      time.Now().Add(-12 * time.Hour).Format("2006-01-02 15:04:05"),
			UpdatedAt:      time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			ID:             "rule_003",
			Name:           "通信中断告警",
			Description:    "设备通信中断超过设定时间时触发",
			Type:           "trend",
			Level:          4,
			Condition:      "offline_duration > threshold",
			Threshold:      300,
			Duration:       0,
			NotifyChannels: []string{"email", "sms", "webhook"},
			NotifyUsers:    []string{"admin", "operator", "maintenance"},
			Status:         1,
			CreatedAt:      time.Now().Add(-6 * time.Hour).Format("2006-01-02 15:04:05"),
			UpdatedAt:      time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	c.JSON(http.StatusOK, dto.PagedResponse{
		Code:      0,
		Message:   "success",
		Data:      rules,
		Timestamp: time.Now().Unix(),
		Page:      1,
		PageSize:  10,
		Total:     int64(len(rules)),
	})
}

// createAlarmRule 创建告警规则
// @Summary 创建告警规则
// @Description 创建新的告警规则
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param rule body object true "告警规则信息"
// @Success 201 {object} dto.Response{data=AlarmRuleResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /alarm-rules [post]
func createAlarmRule(c *gin.Context) {
	c.JSON(http.StatusCreated, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      AlarmRuleResponse{ID: "rule_new"},
		Timestamp: time.Now().Unix(),
	})
}

// getAlarmRule 获取告警规则详情
// @Summary 获取告警规则详情
// @Description 根据ID获取告警规则详细信息
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param id path string true "告警规则ID"
// @Success 200 {object} dto.Response{data=AlarmRuleResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Router /alarm-rules/{id} [get]
func getAlarmRule(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      AlarmRuleResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// updateAlarmRule 更新告警规则
// @Summary 更新告警规则
// @Description 更新告警规则信息
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param id path string true "告警规则ID"
// @Param rule body object true "告警规则信息"
// @Success 200 {object} dto.Response{data=AlarmRuleResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /alarm-rules/{id} [put]
func updateAlarmRule(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      AlarmRuleResponse{ID: c.Param("id")},
		Timestamp: time.Now().Unix(),
	})
}

// deleteAlarmRule 删除告警规则
// @Summary 删除告警规则
// @Description 删除指定告警规则
// @Tags 告警规则管理
// @Accept json
// @Produce json
// @Param id path string true "告警规则ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.ErrorResponse
// @Router /alarm-rules/{id} [delete]
func deleteAlarmRule(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: time.Now().Unix(),
	})
}

// NotificationConfigResponse 通知配置响应
type NotificationConfigResponse struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Name    string                 `json:"name"`
	Config  map[string]interface{} `json:"config"`
	Enabled bool                   `json:"enabled"`
}

// listNotificationConfigs 获取通知配置列表
// @Summary 获取通知配置列表
// @Description 获取所有通知配置
// @Tags 通知配置管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=[]NotificationConfigResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /notification-configs [get]
func listNotificationConfigs(c *gin.Context) {
	configs := []NotificationConfigResponse{
		{
			ID:   "notif_001",
			Type: "email",
			Name: "邮件通知",
			Config: map[string]interface{}{
				"smtp_host": "smtp.example.com",
				"smtp_port": 465,
				"username":  "alert@example.com",
				"from":      "alert@example.com",
				"use_tls":   true,
			},
			Enabled: false,
		},
		{
			ID:   "notif_002",
			Type: "sms",
			Name: "短信通知",
			Config: map[string]interface{}{
				"access_key": "",
				"secret_key": "",
				"sign_name":  "新能源监控",
			},
			Enabled: false,
		},
		{
			ID:   "notif_003",
			Type: "webhook",
			Name: "Webhook通知",
			Config: map[string]interface{}{
				"url":    "",
				"method": "POST",
			},
			Enabled: false,
		},
		{
			ID:   "notif_004",
			Type: "wechat",
			Name: "微信通知",
			Config: map[string]interface{}{
				"corp_id":  "",
				"agent_id": "",
			},
			Enabled: false,
		},
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      configs,
		Timestamp: time.Now().Unix(),
	})
}

// getNotificationConfig 获取通知配置详情
// @Summary 获取通知配置详情
// @Description 根据类型获取通知配置
// @Tags 通知配置管理
// @Accept json
// @Produce json
// @Param type path string true "通知类型"
// @Success 200 {object} dto.Response{data=NotificationConfigResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Router /notification-configs/{type} [get]
func getNotificationConfig(c *gin.Context) {
	notifType := c.Param("type")
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: NotificationConfigResponse{
			ID:   "notif_" + notifType,
			Type: notifType,
			Name: notifType + "通知",
		},
		Timestamp: time.Now().Unix(),
	})
}

// updateNotificationConfig 更新通知配置
// @Summary 更新通知配置
// @Description 更新指定类型的通知配置
// @Tags 通知配置管理
// @Accept json
// @Produce json
// @Param type path string true "通知类型"
// @Param config body object true "配置信息"
// @Success 200 {object} dto.Response{data=NotificationConfigResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /notification-configs/{type} [put]
func updateNotificationConfig(c *gin.Context) {
	notifType := c.Param("type")
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: NotificationConfigResponse{
			ID:   "notif_" + notifType,
			Type: notifType,
		},
		Timestamp: time.Now().Unix(),
	})
}

// enableNotificationConfig 启用通知配置
// @Summary 启用通知配置
// @Description 启用指定类型的通知配置
// @Tags 通知配置管理
// @Accept json
// @Produce json
// @Param type path string true "通知类型"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.ErrorResponse
// @Router /notification-configs/{type}/enable [post]
func enableNotificationConfig(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: time.Now().Unix(),
	})
}

// disableNotificationConfig 禁用通知配置
// @Summary 禁用通知配置
// @Description 禁用指定类型的通知配置
// @Tags 通知配置管理
// @Accept json
// @Produce json
// @Param type path string true "通知类型"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.ErrorResponse
// @Router /notification-configs/{type}/disable [post]
func disableNotificationConfig(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: time.Now().Unix(),
	})
}

// testNotificationConfig 测试通知配置
// @Summary 测试通知配置
// @Description 测试指定类型的通知配置是否可用
// @Tags 通知配置管理
// @Accept json
// @Produce json
// @Param type path string true "通知类型"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /notification-configs/{type}/test [post]
func testNotificationConfig(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "Test successful",
		Timestamp: time.Now().Unix(),
	})
}

// generateReport 生成统计报表
// @Summary 生成统计报表
// @Description 根据条件生成电站统计报表
// @Tags 报表管理
// @Accept json
// @Produce json
// @Param type query string false "报表类型" Enums(daily, weekly, monthly)
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param station_id query string false "电站ID"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Router /reports [get]
func generateReport(c *gin.Context) {
	reportType := c.DefaultQuery("type", "daily")
	startTime := c.DefaultQuery("start_time", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endTime := c.DefaultQuery("end_time", time.Now().Format("2006-01-02"))

	report := map[string]interface{}{
		"type":       reportType,
		"start_time": startTime,
		"end_time":   endTime,
		"stations": []map[string]interface{}{
			{
				"station_id":   "station_001",
				"station_name": "光伏电站A",
				"total_power":  125000,
				"yoy_change":   12.5,
				"mom_change":   5.2,
				"alarm_count":  15,
				"online_rate":  99.5,
			},
			{
				"station_id":   "station_002",
				"station_name": "风电场B",
				"total_power":  89000,
				"yoy_change":   8.3,
				"mom_change":   -2.1,
				"alarm_count":  8,
				"online_rate":  98.2,
			},
		},
		"summary": map[string]interface{}{
			"total_power":     214000,
			"total_alarms":    23,
			"avg_online_rate": 98.85,
		},
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      report,
		Timestamp: time.Now().Unix(),
	})
}

// exportReport 导出报表
// @Summary 导出报表
// @Description 导出报表为Excel或CSV格式
// @Tags 报表管理
// @Accept json
// @Produce octet-stream
// @Param type query string false "报表类型"
// @Param format query string false "导出格式" Enums(excel, csv)
// @Success 200 {file} file
// @Failure 400 {object} dto.ErrorResponse
// @Router /reports/export [get]
func exportReport(c *gin.Context) {
	format := c.DefaultQuery("format", "excel")
	filename := fmt.Sprintf("report_%s.%s", time.Now().Format("20060102150405"), format)

	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.String(http.StatusOK, "电站名称,发电量(kWh),同比,环比,告警数,在线率\n")
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "Export successful",
	})
}

// OperationLogResponse 操作日志响应
type OperationLogResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Action     string `json:"action"`
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
	RequestIP  string `json:"request_ip"`
	Status     int    `json:"status"`
	Duration   int64  `json:"duration"`
	CreatedAt  string `json:"created_at"`
}

// listOperationLogs 获取操作日志列表
// @Summary 获取操作日志列表
// @Description 获取系统操作日志列表，支持分页和过滤
// @Tags 操作日志
// @Accept json
// @Produce json
// @Param user_id query string false "用户ID"
// @Param action query string false "操作类型"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} dto.PagedResponse{data=[]OperationLogResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /operation-logs [get]
func listOperationLogs(c *gin.Context) {
	logs := []OperationLogResponse{
		{
			ID:         "log_001",
			UserID:     "user_001",
			Username:   "admin",
			Method:     "POST",
			Path:       "/api/v1/stations",
			Action:     "create",
			Resource:   "station",
			ResourceID: "station_new",
			RequestIP:  "192.168.1.100",
			Status:     200,
			Duration:   45,
			CreatedAt:  time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
		},
		{
			ID:         "log_002",
			UserID:     "user_001",
			Username:   "admin",
			Method:     "PUT",
			Path:       "/api/v1/devices/device_001",
			Action:     "update",
			Resource:   "device",
			ResourceID: "device_001",
			RequestIP:  "192.168.1.100",
			Status:     200,
			Duration:   32,
			CreatedAt:  time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04:05"),
		},
		{
			ID:        "log_003",
			UserID:    "user_002",
			Username:  "operator",
			Method:    "POST",
			Path:      "/api/v1/auth/login",
			Action:    "login",
			Resource:  "auth",
			RequestIP: "192.168.1.101",
			Status:    200,
			Duration:  128,
			CreatedAt: time.Now().Add(-3 * time.Hour).Format("2006-01-02 15:04:05"),
		},
	}

	c.JSON(http.StatusOK, dto.PagedResponse{
		Code:      0,
		Message:   "success",
		Data:      logs,
		Timestamp: time.Now().Unix(),
		Page:      1,
		PageSize:  20,
		Total:     int64(len(logs)),
	})
}
