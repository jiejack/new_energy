#!/bin/bash

# 测试覆盖率脚本
# 用于生成和验证测试覆盖率报告

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
COVERAGE_THRESHOLD=80
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"
COVERAGE_JSON="coverage.json"

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "======================================"
echo "  新能源监控系统 - 测试覆盖率报告"
echo "======================================"
echo ""

# 切换到项目根目录
cd "$PROJECT_ROOT"

# 清理旧的覆盖率文件
echo -e "${YELLOW}清理旧的覆盖率文件...${NC}"
rm -f "$COVERAGE_FILE" "$COVERAGE_HTML" "$COVERAGE_JSON"

# 运行测试并生成覆盖率数据
echo -e "${YELLOW}运行测试并收集覆盖率数据...${NC}"
go test -v -race -coverprofile="$COVERAGE_FILE" -covermode=atomic ./...

# 检查测试是否成功
if [ $? -ne 0 ]; then
    echo -e "${RED}测试失败！${NC}"
    exit 1
fi

# 生成HTML覆盖率报告
echo -e "${YELLOW}生成HTML覆盖率报告...${NC}"
go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML"

# 计算总覆盖率
echo -e "${YELLOW}计算覆盖率统计...${NC}"
TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')

# 输出覆盖率信息
echo ""
echo "======================================"
echo "  覆盖率统计"
echo "======================================"

# 按模块显示覆盖率
echo ""
echo "模块覆盖率："
echo "--------------------------------------"
go tool cover -func="$COVERAGE_FILE" | grep -E "^github.com/new-energy-monitoring/(internal|pkg)/" | \
    awk '{
        split($1, parts, "/")
        if (parts[3] == "internal") {
            module = parts[4] "/" parts[5]
        } else if (parts[3] == "pkg") {
            module = parts[4] "/" parts[5]
        } else {
            module = parts[3] "/" parts[4]
        }
        coverage[module] += $3
        count[module]++
    }
    END {
        for (m in coverage) {
            if (count[m] > 0) {
                avg = coverage[m] / count[m]
                printf "  %-40s %6.2f%%\n", m, avg
            }
        }
    }' | sort

echo ""
echo "--------------------------------------"
echo -e "总覆盖率: ${GREEN}${TOTAL_COVERAGE}%${NC}"
echo "--------------------------------------"

# 检查是否达到阈值
if (( $(echo "$TOTAL_COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then
    echo -e "${GREEN}✓ 覆盖率达标 (>= ${COVERAGE_THRESHOLD}%)${NC}"
    echo ""
    echo "报告文件："
    echo "  - 文本报告: $COVERAGE_FILE"
    echo "  - HTML报告: $COVERAGE_HTML"
    echo ""
    exit 0
else
    echo -e "${RED}✗ 覆盖率未达标 (< ${COVERAGE_THRESHOLD}%)${NC}"
    echo ""
    echo "建议："
    echo "  1. 为未覆盖的代码添加单元测试"
    echo "  2. 查看 HTML 报告了解具体未覆盖的代码行"
    echo "  3. 重点关注核心业务逻辑的测试覆盖"
    echo ""
    echo "报告文件："
    echo "  - 文本报告: $COVERAGE_FILE"
    echo "  - HTML报告: $COVERAGE_HTML"
    echo ""
    exit 1
fi
