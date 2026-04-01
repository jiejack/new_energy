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
	fmt.Printf("New Energy Monitoring - Compute Service\n")
	fmt.Printf("Version: %s, Build Time: %s\n\n", Version, BuildTime)

	if err := initConfig(); err != nil {
		panic(fmt.Errorf("failed to init config: %w", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	computeService := NewComputeService()
	go computeService.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down compute service...")
	cancel()
	time.Sleep(2 * time.Second)
	fmt.Println("Compute service exited")
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

type ComputeService struct {
	formulaEngine *FormulaEngine
	ruleEngine    *ComputeRuleEngine
}

func NewComputeService() *ComputeService {
	return &ComputeService{
		formulaEngine: NewFormulaEngine(),
		ruleEngine:    NewComputeRuleEngine(),
	}
}

func (s *ComputeService) Start(ctx context.Context) {
	fmt.Println("Starting compute service...")

	go s.formulaEngine.Start(ctx)
	go s.ruleEngine.Start(ctx)

	<-ctx.Done()
	fmt.Println("Compute service stopped")
}

type FormulaEngine struct{}

func NewFormulaEngine() *FormulaEngine {
	return &FormulaEngine{}
}

func (e *FormulaEngine) Start(ctx context.Context) {
	fmt.Println("Formula engine started")
	<-ctx.Done()
}

type ComputeRuleEngine struct{}

func NewComputeRuleEngine() *ComputeRuleEngine {
	return &ComputeRuleEngine{}
}

func (e *ComputeRuleEngine) Start(ctx context.Context) {
	fmt.Println("Compute rule engine started")
	<-ctx.Done()
}
