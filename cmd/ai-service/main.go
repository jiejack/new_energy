package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	fmt.Printf("New Energy Monitoring - AI Service\n")
	fmt.Printf("Version: %s, Build Time: %s\n\n", Version, BuildTime)

	if err := initConfig(); err != nil {
		panic(fmt.Errorf("failed to init config: %w", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	aiService := NewAIService()
	go aiService.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down AI service...")
	cancel()
	time.Sleep(2 * time.Second)
	fmt.Println("AI service exited")
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	return nil
}

type AIService struct {
	knowledgeBase   *KnowledgeBase
	smartQA         *SmartQA
	smartConfig     *SmartConfig
	smartOperation  *SmartOperation
}

func NewAIService() *AIService {
	return &AIService{
		knowledgeBase:   NewKnowledgeBase(),
		smartQA:         NewSmartQA(),
		smartConfig:     NewSmartConfig(),
		smartOperation:  NewSmartOperation(),
	}
}

func (s *AIService) Start(ctx context.Context) {
	fmt.Println("Starting AI service...")

	go s.knowledgeBase.Start(ctx)
	go s.smartQA.Start(ctx)
	go s.smartConfig.Start(ctx)
	go s.smartOperation.Start(ctx)

	<-ctx.Done()
	fmt.Println("AI service stopped")
}

type KnowledgeBase struct{}

func NewKnowledgeBase() *KnowledgeBase {
	return &KnowledgeBase{}
}

func (kb *KnowledgeBase) Start(ctx context.Context) {
	fmt.Println("Knowledge base service started")
	<-ctx.Done()
}

type SmartQA struct{}

func NewSmartQA() *SmartQA {
	return &SmartQA{}
}

func (qa *SmartQA) Start(ctx context.Context) {
	fmt.Println("Smart QA service started")
	<-ctx.Done()
}

type SmartConfig struct{}

func NewSmartConfig() *SmartConfig {
	return &SmartConfig{}
}

func (sc *SmartConfig) Start(ctx context.Context) {
	fmt.Println("Smart config service started")
	<-ctx.Done()
}

type SmartOperation struct{}

func NewSmartOperation() *SmartOperation {
	return &SmartOperation{}
}

func (so *SmartOperation) Start(ctx context.Context) {
	fmt.Println("Smart operation service started")
	<-ctx.Done()
}
