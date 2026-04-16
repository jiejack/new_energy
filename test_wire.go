package main

import (
	"fmt"

	"github.com/new-energy-monitoring/internal/api/handler"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/infrastructure/persistence"
)

func main() {
	// 简单测试，确保所有的包都能正确导入
	fmt.Println("Testing imports...")
	fmt.Println("HandlerSet:", handler.HandlerSet)
	fmt.Println("ServiceSet:", service.ServiceSet)
	fmt.Println("RepositorySet:", persistence.RepositorySet)
	fmt.Println("Imports successful!")
}
