#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Deploy the New Energy Monitoring System"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -m, --mode MODE         Deployment mode: docker or k8s (default: docker)"
    echo "  -f, --full              Deploy full stack (all microservices)"
    echo "  -b, --build-only        Only build images, don't deploy"
    echo "  -d, --deploy-only       Only deploy, skip building"
    echo "  -r, --registry REGISTRY Docker registry URL (default: localhost:5000)"
    echo "  -t, --tag TAG           Image tag (default: latest)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Deploy basic Docker stack"
    echo "  $0 --full                            # Deploy full Docker stack"
    echo "  $0 --mode k8s                        # Deploy to Kubernetes"
    echo "  $0 --registry my-registry.com --tag v1.0.0"
    exit 0
}

MODE="docker"
FULL_STACK=false
BUILD_ONLY=false
DEPLOY_ONLY=false
DOCKER_REGISTRY="localhost:5000"
IMAGE_TAG="latest"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        -m|--mode)
            MODE="$2"
            shift 2
            ;;
        -f|--full)
            FULL_STACK=true
            shift
            ;;
        -b|--build-only)
            BUILD_ONLY=true
            shift
            ;;
        -d|--deploy-only)
            DEPLOY_ONLY=true
            shift
            ;;
        -r|--registry)
            DOCKER_REGISTRY="$2"
            shift 2
            ;;
        -t|--tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

export DOCKER_REGISTRY="$DOCKER_REGISTRY"
export IMAGE_TAG="$IMAGE_TAG"

log_info "============================================"
log_info "New Energy Monitoring System Deployment"
log_info "============================================"
log_info "Mode: $MODE"
log_info "Full stack: $FULL_STACK"
log_info "Registry: $DOCKER_REGISTRY"
log_info "Tag: $IMAGE_TAG"
log_info ""

check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    if [ "$MODE" = "docker" ]; then
        if ! command -v docker-compose &> /dev/null; then
            log_error "docker-compose is not installed"
            exit 1
        fi
    fi
    
    if [ "$MODE" = "k8s" ]; then
        if ! command -v kubectl &> /dev/null; then
            log_error "kubectl is not installed"
            exit 1
        fi
    fi
    
    log_success "Dependencies check passed"
}

build_images() {
    log_info "Building Docker images..."
    
    if [ "$FULL_STACK" = true ]; then
        make docker-build-all
    else
        make docker-build
    fi
    
    log_success "Docker images built successfully"
}

push_images() {
    if [ "$DOCKER_REGISTRY" != "localhost:5000" ]; then
        log_info "Pushing Docker images..."
        make docker-push-all
        log_success "Docker images pushed successfully"
    fi
}

deploy_docker() {
    log_info "Deploying to Docker..."
    
    if [ "$FULL_STACK" = true ]; then
        make docker-full-up
    else
        make docker-up
    fi
    
    log_success "Docker deployment completed"
    
    log_info ""
    log_info "Access the services:"
    log_info "  Frontend: http://localhost:80"
    log_info "  API: http://localhost:8080"
    log_info "  Grafana: http://localhost:3000"
    log_info "  Prometheus: http://localhost:9090"
    log_info ""
    log_info "View logs with: make docker-logs or make docker-full-logs"
}

deploy_k8s() {
    log_info "Deploying to Kubernetes..."
    
    make k8s-deploy
    
    log_success "Kubernetes deployment completed"
    
    log_info ""
    log_info "Checking deployment status..."
    sleep 10
    make k8s-status
    
    log_info ""
    log_info "View logs with: make k8s-logs"
}

main() {
    check_dependencies
    
    if [ "$DEPLOY_ONLY" = false ]; then
        build_images
        push_images
    fi
    
    if [ "$BUILD_ONLY" = false ]; then
        if [ "$MODE" = "docker" ]; then
            deploy_docker
        elif [ "$MODE" = "k8s" ]; then
            deploy_k8s
        else
            log_error "Invalid mode: $MODE"
            exit 1
        fi
    fi
    
    log_info ""
    log_success "============================================"
    log_success "Deployment process completed!"
    log_success "============================================"
}

main
