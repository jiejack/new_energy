.PHONY: all build clean test test-unit test-integration test-coverage test-coverage-check lint docker fmt vet help wire wire-check

APP_NAME := new-energy-monitoring
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | awk '{print $$3}')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# 覆盖率阈值
COVERAGE_THRESHOLD := 80

all: clean build

build: wire
	@echo "Building $(APP_NAME)..."
	@go build $(LDFLAGS) -o bin/api-server ./cmd/api-server
	@go build $(LDFLAGS) -o bin/collector ./cmd/collector
	@go build $(LDFLAGS) -o bin/alarm ./cmd/alarm
	@go build $(LDFLAGS) -o bin/compute ./cmd/compute
	@go build $(LDFLAGS) -o bin/ai-service ./cmd/ai-service
	@go build $(LDFLAGS) -o bin/scheduler ./cmd/scheduler

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/api-server-linux ./cmd/api-server
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/collector-linux ./cmd/collector
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/alarm-linux ./cmd/alarm
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/compute-linux ./cmd/compute
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/ai-service-linux ./cmd/ai-service
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/scheduler-linux ./cmd/scheduler

test:
	@echo "Running all tests..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

test-unit:
	@echo "Running unit tests..."
	@bash tests/run_tests.sh

test-integration:
	@echo "Running integration tests..."
	@bash tests/integration_test.sh

test-coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | grep total

test-coverage-check: test
	@echo "Checking coverage threshold (>= $(COVERAGE_THRESHOLD)%)..."
	@bash tests/coverage.sh

lint:
	@echo "Running linters..."
	@golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

vet:
	@echo "Running go vet..."
	@go vet ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

docker-build:
	@echo "Building Docker images..."
	@docker-compose -f deployments/docker/docker-compose.yml build

docker-up:
	@echo "Starting Docker containers..."
	@docker-compose -f deployments/docker/docker-compose.yml up -d

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose -f deployments/docker/docker-compose.yml down

docker-logs:
	@docker-compose -f deployments/docker/docker-compose.yml logs -f

proto:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go-grpc_out=. api/proto/*.proto

swagger:
	@echo "Generating Swagger docs..."
	@swag init -g cmd/api-server/main.go -o docs --exclude pkg/protocol/iec61850

swagger-serve:
	@echo "Serving Swagger UI..."
	@docker run --rm -p 8081:8080 -e SWAGGER_JSON=/docs/swagger.json -v $(PWD)/docs:/docs swaggerapi/swagger-ui

swagger-validate:
	@echo "Validating Swagger spec..."
	@docker run --rm -v $(PWD)/docs:/docs swaggerapi/swagger-validator-online /docs/swagger.json

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

deps-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

wire:
	@echo "Generating wire code..."
	@cd cmd/api-server && wire

wire-check:
	@echo "Checking wire dependencies..."
	@cd cmd/api-server && wire check

wire-graph:
	@echo "Generating wire dependency graph..."
	@cd cmd/api-server && wire graph > wire_graph.dot

run-api: wire
	@go run ./cmd/api-server

run-collector:
	@go run ./cmd/collector

run-alarm:
	@go run ./cmd/alarm

run-compute:
	@go run ./cmd/compute

run-ai:
	@go run ./cmd/ai-service

run-scheduler:
	@go run ./cmd/scheduler

help:
	@echo "Available targets:"
	@echo "  make build              - Build all services"
	@echo "  make build-linux        - Build for Linux AMD64"
	@echo "  make test               - Run all tests"
	@echo "  make test-unit          - Run unit tests only"
	@echo "  make test-integration   - Run integration tests"
	@echo "  make test-coverage      - Run tests with coverage report"
	@echo "  make test-coverage-check - Run tests and check coverage threshold"
	@echo "  make lint               - Run linters"
	@echo "  make fmt                - Format code"
	@echo "  make vet                - Run go vet"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make docker-build       - Build Docker images"
	@echo "  make docker-up          - Start Docker containers"
	@echo "  make docker-down        - Stop Docker containers"
	@echo "  make docker-logs        - View Docker logs"
	@echo "  make proto              - Generate protobuf code"
	@echo "  make swagger            - Generate Swagger docs"
	@echo "  make swagger-serve      - Serve Swagger UI in Docker"
	@echo "  make swagger-validate   - Validate Swagger spec"
	@echo "  make deps               - Install dependencies"
	@echo "  make deps-update        - Update dependencies"
	@echo "  make wire               - Generate wire dependency injection code"
	@echo "  make wire-check         - Check wire dependencies"
	@echo "  make wire-graph         - Generate wire dependency graph"
	@echo "  make run-api            - Run API server"
	@echo "  make run-collector      - Run collector service"
	@echo "  make run-alarm          - Run alarm service"
	@echo "  make run-compute        - Run compute service"
	@echo "  make run-ai             - Run AI service"
	@echo "  make run-scheduler      - Run scheduler service"
