#!/bin/bash

# 集成测试脚本
# 用于运行集成测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "======================================"
echo "  新能源监控系统 - 集成测试"
echo "======================================"
echo ""

# 切换到项目根目录
cd "$PROJECT_ROOT"

# 检查是否需要启动依赖服务
echo -e "${YELLOW}检查测试环境...${NC}"

# 检查 Docker 是否运行
if command -v docker &> /dev/null; then
    if docker info &> /dev/null; then
        echo -e "${GREEN}Docker 运行中${NC}"
        
        # 启动测试依赖服务
        echo -e "${YELLOW}启动测试依赖服务...${NC}"
        docker-compose -f deployments/docker/docker-compose.yml up -d postgres redis kafka 2>/dev/null || true
        
        # 等待服务启动
        echo -e "${YELLOW}等待服务启动...${NC}"
        sleep 10
    else
        echo -e "${YELLOW}Docker 未运行，跳过依赖服务启动${NC}"
    fi
else
    echo -e "${YELLOW}Docker 未安装，跳过依赖服务启动${NC}"
fi

# 运行集成测试
echo ""
echo -e "${BLUE}运行集成测试...${NC}"
echo "--------------------------------------"

# 设置测试环境变量
export TEST_ENV=integration
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=test_db
export DB_USER=test_user
export DB_PASSWORD=test_pass
export REDIS_HOST=localhost
export REDIS_PORT=6379
export KAFKA_BROKERS=localhost:9092

# 运行集成测试
if go test -v -race -tags=integration -timeout 5m ./tests/integration/...; then
    echo -e "${GREEN}✓ 集成测试通过${NC}"
    INTEGRATION_RESULT=0
else
    echo -e "${RED}✗ 集成测试失败${NC}"
    INTEGRATION_RESULT=1
fi

# 清理
echo ""
echo -e "${YELLOW}清理测试环境...${NC}"
if command -v docker &> /dev/null && docker info &> /dev/null; then
    docker-compose -f deployments/docker/docker-compose.yml down 2>/dev/null || true
fi

# 输出结果
echo ""
echo "======================================"
if [ $INTEGRATION_RESULT -eq 0 ]; then
    echo -e "${GREEN}集成测试完成：成功${NC}"
else
    echo -e "${RED}集成测试完成：失败${NC}"
fi
echo "======================================"

exit $INTEGRATION_RESULT
