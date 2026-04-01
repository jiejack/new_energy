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
	fmt.Printf("New Energy Monitoring - Scheduler Service\n")
	fmt.Printf("Version: %s, Build Time: %s\n\n", Version, BuildTime)

	if err := initConfig(); err != nil {
		panic(fmt.Errorf("failed to init config: %w", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := NewScheduler()
	go scheduler.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down scheduler...")
	cancel()
	time.Sleep(2 * time.Second)
	fmt.Println("Scheduler exited")
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

type Scheduler struct {
	taskQueue chan Task
	workers   []*Worker
}

type Task struct {
	ID      string
	Type    string
	Payload interface{}
}

type Worker struct {
	ID int
}

func NewScheduler() *Scheduler {
	workerCount := viper.GetInt("scheduler.workers")
	if workerCount == 0 {
		workerCount = 10
	}

	s := &Scheduler{
		taskQueue: make(chan Task, 10000),
		workers:   make([]*Worker, workerCount),
	}

	for i := 0; i < workerCount; i++ {
		s.workers[i] = &Worker{ID: i}
	}

	return s
}

func (s *Scheduler) Start(ctx context.Context) {
	fmt.Printf("Starting scheduler with %d workers...\n", len(s.workers))

	for _, worker := range s.workers {
		go s.runWorker(ctx, worker)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Scheduler stopped")
			return
		case <-ticker.C:
			s.checkScheduledTasks()
		}
	}
}

func (s *Scheduler) runWorker(ctx context.Context, worker *Worker) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-s.taskQueue:
			s.processTask(worker, task)
		}
	}
}

func (s *Scheduler) processTask(worker *Worker, task Task) {
	fmt.Printf("Worker %d processing task: %s\n", worker.ID, task.ID)
}

func (s *Scheduler) checkScheduledTasks() {
}
