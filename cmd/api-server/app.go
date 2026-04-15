package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/handler"
	"github.com/new-energy-monitoring/internal/infrastructure/cache"
	"github.com/new-energy-monitoring/internal/infrastructure/config"
	"github.com/new-energy-monitoring/internal/infrastructure/logger"
	"github.com/new-energy-monitoring/internal/infrastructure/mq"
	"github.com/new-energy-monitoring/internal/infrastructure/persistence"
	"github.com/new-energy-monitoring/pkg/auth"
	"go.uber.org/zap"
)

// App 应用程序结构体
type App struct {
	config     *config.Config
	logger     *zap.Logger
	database   *persistence.Database
	redis      *cache.RedisClient
	kafka      *mq.KafkaProducer
	httpServer *http.Server
}

// NewApp 创建应用程序实例
func NewApp(
	cfg *config.Config,
	log *zap.Logger,
	db *persistence.Database,
	redis *cache.RedisClient,
	kafka *mq.KafkaProducer,
	httpServer *http.Server,
) *App {
	return &App{
		config:     cfg,
		logger:     log,
		database:   db,
		redis:      redis,
		kafka:      kafka,
		httpServer: httpServer,
	}
}

// Run 启动应用程序
func (a *App) Run() error {
	// 启动 HTTP 服务器
	go func() {
		a.logger.Info("Starting HTTP server", zap.Int("port", a.config.Server.Port))
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	a.logger.Info("Application started successfully")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down application...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
	}

	// 关闭数据库连接
	if err := a.database.Close(); err != nil {
		a.logger.Error("Failed to close database", zap.Error(err))
	}

	// 关闭 Redis 连接
	if err := a.redis.Close(); err != nil {
		a.logger.Error("Failed to close Redis", zap.Error(err))
	}

	// 关闭 Kafka 连接
	if err := a.kafka.Close(); err != nil {
		a.logger.Error("Failed to close Kafka", zap.Error(err))
	}

	a.logger.Info("Application stopped")
	return nil
}

// NewConfig 创建配置实例
func NewConfig() (*config.Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./configs/config.yaml"
	}

	return config.Load(configPath)
}

// NewLogger 创建日志实例
func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	logCfg := &logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
		Output: cfg.Logging.Output,
	}

	if err := logger.Init(logCfg); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	return logger.Log, nil
}

// NewDatabase 创建数据库实例
func NewDatabase(cfg *config.Config, log *zap.Logger) (*persistence.Database, error) {
	dbCfg := persistence.DatabaseConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	db, err := persistence.NewDatabase(dbCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	log.Info("Database connected successfully")
	return db, nil
}

// NewRedis 创建 Redis 实例
func NewRedis(cfg *config.Config, log *zap.Logger) (*cache.RedisClient, error) {
	redisCfg := cache.RedisConfig{
		Addrs:    cfg.Redis.Addrs,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	}

	redis, err := cache.NewRedisClient(redisCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	log.Info("Redis connected successfully")
	return redis, nil
}

// NewKafka 创建 Kafka 实例
func NewKafka(cfg *config.Config, log *zap.Logger) (*mq.KafkaProducer, error) {
	kafkaCfg := mq.KafkaConfig{
		Brokers:     cfg.Kafka.Brokers,
		TopicPrefix: cfg.Kafka.TopicPrefix,
	}

	kafka := mq.NewKafkaProducer(kafkaCfg, mq.TopicDataCollect)

	log.Info("Kafka producer created successfully")
	return kafka, nil
}

// NewJWTManager 创建 JWT 管理器实例
func NewJWTManager(cfg *config.Config) *auth.JWTManager {
	jwtConfig := &auth.JWTConfig{
		Secret:        cfg.Auth.JWT.Secret,
		AccessExpire:  cfg.Auth.JWT.AccessExpire,
		RefreshExpire: cfg.Auth.JWT.RefreshExpire,
	}
	return auth.NewJWTManager(jwtConfig)
}

// NewPasswordManager 创建密码管理器实例
func NewPasswordManager(cfg *config.Config) *auth.PasswordManager {
	passwordConfig := &auth.PasswordConfig{
		MinLength:        cfg.Auth.Password.MinLength,
		RequireUppercase: cfg.Auth.Password.RequireUppercase,
		RequireLowercase: cfg.Auth.Password.RequireLowercase,
		RequireDigit:     cfg.Auth.Password.RequireDigit,
	}
	return auth.NewPasswordManager(passwordConfig)
}

// NewHTTPServer 创建 HTTP 服务器实例
func NewHTTPServer(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	deviceHandler *handler.DeviceHandler,
	alarmHandler *handler.AlarmHandler,
	alarmRuleHandler *handler.AlarmRuleHandler,
	stationHandler *handler.StationHandler,
	regionHandler *handler.RegionHandler,
	pointHandler *handler.PointHandler,
	qaHandler *handler.QAHandler,
	configHandler *handler.ConfigHandler,
	notificationConfigHandler *handler.NotificationConfigHandler,
	exportHandler *handler.ExportHandler,
	reportHandler *handler.ReportHandler,
	operationLogHandler *handler.OperationLogHandler,
	energyEfficiencyHandler *handler.EnergyEfficiencyHandler,
	// carbonEmissionHandler *handler.CarbonEmissionHandler,
) *http.Server {
	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ready": true,
		})
	})

	// API 路由组
	api := router.Group("/api/v1")
	{
		// 认证路由
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 用户路由
		users := api.Group("/users")
		{
			users.GET("", userHandler.ListUsers)
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.PUT("/:id/password", userHandler.ChangePassword)
		}

		// 设备路由
		devices := api.Group("/devices")
		{
			devices.GET("", deviceHandler.ListDevices)
			devices.POST("", deviceHandler.CreateDevice)
			devices.GET("/:id", deviceHandler.GetDevice)
			devices.PUT("/:id", deviceHandler.UpdateDevice)
			devices.DELETE("/:id", deviceHandler.DeleteDevice)
		}

		// 告警路由
		alarms := api.Group("/alarms")
		{
			alarms.GET("", alarmHandler.ListAlarms)
			alarms.GET("/:id", alarmHandler.GetAlarm)
			alarms.PUT("/:id/ack", alarmHandler.AcknowledgeAlarm)
			alarms.PUT("/:id/clear", alarmHandler.ClearAlarm)
		}

		// 厂站路由
		stations := api.Group("/stations")
		{
			stations.GET("", stationHandler.ListStations)
			stations.POST("", stationHandler.CreateStation)
			stations.GET("/:id", stationHandler.GetStation)
			stations.PUT("/:id", stationHandler.UpdateStation)
			stations.DELETE("/:id", stationHandler.DeleteStation)
		}

		// 区域路由
		regions := api.Group("/regions")
		{
			regions.GET("", regionHandler.ListRegions)
			regions.POST("", regionHandler.CreateRegion)
			regions.GET("/:id", regionHandler.GetRegion)
			regions.PUT("/:id", regionHandler.UpdateRegion)
			regions.DELETE("/:id", regionHandler.DeleteRegion)
		}

		// 采集点路由
		points := api.Group("/points")
		{
			points.GET("", pointHandler.ListPoints)
			points.POST("", pointHandler.CreatePoint)
			points.GET("/:id", pointHandler.GetPoint)
			points.PUT("/:id", pointHandler.UpdatePoint)
			points.DELETE("/:id", pointHandler.DeletePoint)
		}

		// 操作日志路由
		operationLogs := api.Group("/operation-logs")
		{
			operationLogs.GET("", operationLogHandler.ListLogs)
			operationLogs.GET("/:id", operationLogHandler.GetLog)
			operationLogs.POST("", operationLogHandler.CreateLog)
			operationLogs.DELETE("/cleanup", operationLogHandler.DeleteOldLogs)
		}

		// 告警规则路由
		alarmRules := api.Group("/alarm-rules")
		{
			alarmRules.GET("", alarmRuleHandler.ListAlarmRules)
			alarmRules.POST("", alarmRuleHandler.CreateAlarmRule)
			alarmRules.GET("/:id", alarmRuleHandler.GetAlarmRule)
			alarmRules.PUT("/:id", alarmRuleHandler.UpdateAlarmRule)
			alarmRules.DELETE("/:id", alarmRuleHandler.DeleteAlarmRule)
			alarmRules.POST("/:id/enable", alarmRuleHandler.EnableAlarmRule)
			alarmRules.POST("/:id/disable", alarmRuleHandler.DisableAlarmRule)
		}

		// 系统配置路由
		configs := api.Group("/configs")
		{
			configs.GET("", configHandler.GetAllConfigs)
			configs.GET("/list", configHandler.ListConfigs)
			configs.POST("", configHandler.CreateConfig)
			configs.POST("/batch", configHandler.BatchUpdateConfigs)
			configs.GET("/:category", configHandler.GetConfigsByCategory)
			configs.GET("/:category/:key", configHandler.GetConfig)
			configs.PUT("/:category/:key", configHandler.UpdateConfig)
			configs.DELETE("/:category/:key", configHandler.DeleteConfig)
		}

		// 通知配置路由
		notificationConfigs := api.Group("/notification-configs")
		{
			notificationConfigs.GET("", notificationConfigHandler.GetAllConfigs)
			notificationConfigs.GET("/:type", notificationConfigHandler.GetConfigByType)
			notificationConfigs.PUT("/:type", notificationConfigHandler.UpdateConfig)
			notificationConfigs.POST("/:type/enable", notificationConfigHandler.EnableConfig)
			notificationConfigs.POST("/:type/disable", notificationConfigHandler.DisableConfig)
			notificationConfigs.POST("/:type/test", notificationConfigHandler.TestConfig)
		}

		// QA路由
		qa := api.Group("/qa")
		{
			qa.POST("/sessions", qaHandler.CreateSession)
			qa.GET("/sessions", qaHandler.ListSessions)
			qa.GET("/sessions/:id", qaHandler.GetSession)
			qa.DELETE("/sessions/:id", qaHandler.DeleteSession)
			qa.POST("/sessions/:id/archive", qaHandler.ArchiveSession)
			qa.GET("/sessions/:id/history", qaHandler.GetHistory)
			qa.POST("/ask", qaHandler.Ask)
		}

		// 报表路由
		reports := api.Group("/reports")
		{
			reports.GET("", reportHandler.GenerateReport)
			reports.GET("/export", reportHandler.ExportReport)
		}
		
		// 能效分析路由
		energyEfficiency := api.Group("/energy-efficiency")
		{
			energyEfficiency.POST("/records", energyEfficiencyHandler.CreateEnergyEfficiencyRecord)
			energyEfficiency.POST("/records/batch", energyEfficiencyHandler.BatchCreateEnergyEfficiencyRecords)
			energyEfficiency.GET("/records", energyEfficiencyHandler.ListEnergyEfficiencyRecords)
			energyEfficiency.GET("/records/:id", energyEfficiencyHandler.GetEnergyEfficiencyRecord)
			energyEfficiency.GET("/trend", energyEfficiencyHandler.GetEnergyEfficiencyTrend)
			energyEfficiency.GET("/statistics", energyEfficiencyHandler.GetEnergyEfficiencyStatistics)
			energyEfficiency.GET("/comparison", energyEfficiencyHandler.GetEnergyEfficiencyComparison)
			energyEfficiency.POST("/analyses", energyEfficiencyHandler.CreateEnergyEfficiencyAnalysis)
			energyEfficiency.GET("/analyses", energyEfficiencyHandler.ListEnergyEfficiencyAnalyses)
			energyEfficiency.GET("/analyses/:id", energyEfficiencyHandler.GetEnergyEfficiencyAnalysis)
			energyEfficiency.GET("/analyses/latest", energyEfficiencyHandler.GetLatestEnergyEfficiencyAnalysis)
		}
		
		// 碳排放监测路由 (暂时注释，待后续完善)
		// carbonEmission := api.Group("/carbon-emission")
		// {
		// 	carbonEmission.POST("/records", carbonEmissionHandler.CreateCarbonEmissionRecord)
		// 	carbonEmission.POST("/records/batch", carbonEmissionHandler.BatchCreateCarbonEmissionRecords)
		// 	carbonEmission.GET("/records", carbonEmissionHandler.ListCarbonEmissionRecords)
		// 	carbonEmission.GET("/records/:id", carbonEmissionHandler.GetCarbonEmissionRecord)
		// 	carbonEmission.GET("/trend", carbonEmissionHandler.GetCarbonEmissionTrend)
		// 	carbonEmission.GET("/statistics", carbonEmissionHandler.GetCarbonEmissionStatistics)
		// 	carbonEmission.GET("/comparison", carbonEmissionHandler.GetCarbonEmissionComparison)
		// 	carbonEmission.POST("/analyses", carbonEmissionHandler.CreateCarbonEmissionAnalysis)
		// 	carbonEmission.GET("/analyses", carbonEmissionHandler.ListCarbonEmissionAnalyses)
		// 	carbonEmission.GET("/analyses/:id", carbonEmissionHandler.GetCarbonEmissionAnalysis)
		// 	carbonEmission.GET("/analyses/latest", carbonEmissionHandler.GetLatestCarbonEmissionAnalysis)
		// }
	}

	// WebSocket 路由
	router.GET("/ws", func(c *gin.Context) {
		// WebSocket 连接处理
		c.JSON(http.StatusOK, gin.H{
			"message": "WebSocket endpoint - upgrade required",
		})
	})

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return srv
}
