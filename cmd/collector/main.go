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
	fmt.Printf("New Energy Monitoring - Collector Service\n")
	fmt.Printf("Version: %s, Build Time: %s\n\n", Version, BuildTime)

	if err := initConfig(); err != nil {
		panic(fmt.Errorf("failed to init config: %w", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	collector := NewCollector()
	go collector.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down collector...")
	cancel()
	time.Sleep(2 * time.Second)
	fmt.Println("Collector exited")
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

type Collector struct {
	workers map[string]Worker
}

type Worker interface {
	Start(ctx context.Context) error
	Stop() error
	Name() string
}

func NewCollector() *Collector {
	return &Collector{
		workers: make(map[string]Worker),
	}
}

func (c *Collector) Start(ctx context.Context) {
	fmt.Println("Starting collector workers...")

	for name, worker := range c.workers {
		go func(n string, w Worker) {
			fmt.Printf("Starting worker: %s\n", n)
			if err := w.Start(ctx); err != nil {
				fmt.Printf("Worker %s stopped with error: %v\n", n, err)
			}
		}(name, worker)
	}

	<-ctx.Done()
	fmt.Println("Collector context cancelled, stopping workers...")

	for name, worker := range c.workers {
		fmt.Printf("Stopping worker: %s\n", name)
		if err := worker.Stop(); err != nil {
			fmt.Printf("Error stopping worker %s: %v\n", name, err)
		}
	}
}
