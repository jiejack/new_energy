#!/bin/bash

# 单元测试脚本
# 用于运行所有单元测试

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
echo "  新能源监控系统 - 单元测试"
echo "======================================"
echo ""

# 切换到项目根目录
cd "$PROJECT_ROOT"

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试模块列表
declare -a MODULES=(
    "internal/domain/entity"
    "internal/application/service"
    "internal/infrastructure/persistence"
    "pkg/alarm/detector"
    "pkg/alarm/notifier"
    "pkg/alarm/rule"
    "pkg/collector"
    "pkg/compute/formula"
    "pkg/compute/rule"
    "pkg/storage/compression"
    "pkg/storage/index"
    "pkg/storage/partition"
    "pkg/storage/lifecycle"
    "pkg/protocol/iec104"
    "pkg/protocol/modbus"
    "pkg/protocol/iec61850"
    "pkg/processor"
    "pkg/monitoring/alerting"
)

# 运行每个模块的测试
for module in "${MODULES[@]}"; do
    echo -e "${BLUE}测试模块: $module${NC}"
    echo "--------------------------------------"
    
    if [ -d "$module" ]; then
        # 运行测试
        if go test -v -race -timeout 30s "./$module/..."; then
            echo -e "${GREEN}✓ $module 测试通过${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}✗ $module 测试失败${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        echo -e "${YELLOW}⚠ $module 目录不存在，跳过${NC}"
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
done

# 输出测试摘要
echo "======================================"
echo "  测试摘要"
echo "======================================"
echo -e "总模块数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"
echo "======================================"

# 返回适当的退出码
if [ $FAILED_TESTS -gt 0 ]; then
    exit 1
else
    exit 0
fi
