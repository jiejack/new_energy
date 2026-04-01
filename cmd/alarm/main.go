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
	fmt.Printf("New Energy Monitoring - Alarm Service\n")
	fmt.Printf("Version: %s, Build Time: %s\n\n", Version, BuildTime)

	if err := initConfig(); err != nil {
		panic(fmt.Errorf("failed to init config: %w", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	alarmService := NewAlarmService()
	go alarmService.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down alarm service...")
	cancel()
	time.Sleep(2 * time.Second)
	fmt.Println("Alarm service exited")
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

type AlarmService struct {
	ruleEngine *RuleEngine
	notifier   *Notifier
}

func NewAlarmService() *AlarmService {
	return &AlarmService{
		ruleEngine: NewRuleEngine(),
		notifier:   NewNotifier(),
	}
}

func (s *AlarmService) Start(ctx context.Context) {
	fmt.Println("Starting alarm service...")

	go s.ruleEngine.Start(ctx)
	go s.notifier.Start(ctx)

	<-ctx.Done()
	fmt.Println("Alarm service stopped")
}

type RuleEngine struct{}

func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

func (e *RuleEngine) Start(ctx context.Context) {
	fmt.Println("Rule engine started")
	<-ctx.Done()
}

type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) Start(ctx context.Context) {
	fmt.Println("Notifier started")
	<-ctx.Done()
}
