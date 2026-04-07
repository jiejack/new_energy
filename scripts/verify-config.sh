#!/usr/bin/env bash
# 验证脚本 - 检查所有配置是否正确

set -e

echo "=========================================="
echo "新能源监控系统 - 配置验证脚本"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 计数器
PASS=0
FAIL=0

# 检查函数
check() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}[PASS]${NC} $1"
        ((PASS++))
    else
        echo -e "${RED}[FAIL]${NC} $1"
        ((FAIL++))
    fi
}

# 1. 检查必要文件
echo "1. 检查必要文件..."
[ -f "Dockerfile.backend" ]
check "Dockerfile.backend 存在"

[ -f "web/Dockerfile" ]
check "web/Dockerfile 存在"

[ -f "docker-compose.yml" ]
check "docker-compose.yml 存在"

[ -f ".dockerignore" ]
check ".dockerignore 存在"

echo ""

# 2. 检查 Kubernetes 配置
echo "2. 检查 Kubernetes 配置..."
[ -f "k8s/namespace.yaml" ]
check "k8s/namespace.yaml 存在"

[ -f "k8s/configmap.yaml" ]
check "k8s/configmap.yaml 存在"

[ -f "k8s/secrets.yaml" ]
check "k8s/secrets.yaml 存在"

[ -f "k8s/deployment-backend.yaml" ]
check "k8s/deployment-backend.yaml 存在"

[ -f "k8s/deployment-frontend.yaml" ]
check "k8s/deployment-frontend.yaml 存在"

[ -f "k8s/service.yaml" ]
check "k8s/service.yaml 存在"

[ -f "k8s/ingress.yaml" ]
check "k8s/ingress.yaml 存在"

echo ""

# 3. 检查 GitHub Actions
echo "3. 检查 GitHub Actions 工作流..."
[ -f ".github/workflows/ci.yml" ]
check ".github/workflows/ci.yml 存在"

[ -f ".github/workflows/cd.yml" ]
check ".github/workflows/cd.yml 存在"

[ -f ".github/workflows/test-coverage.yml" ]
check ".github/workflows/test-coverage.yml 存在"

echo ""

# 4. 检查文档
echo "4. 检查文档..."
[ -f "docs/deployment-guide.md" ]
check "docs/deployment-guide.md 存在"

[ -f "docs/operations-guide.md" ]
check "docs/operations-guide.md 存在"

[ -f "docs/troubleshooting.md" ]
check "docs/troubleshooting.md 存在"

echo ""

# 5. 检查工具是否安装
echo "5. 检查工具是否安装..."
command -v docker &> /dev/null
check "Docker 已安装"

command -v docker-compose &> /dev/null || command -v docker &> /dev/null
check "Docker Compose 已安装"

command -v kubectl &> /dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}[PASS]${NC} kubectl 已安装"
    ((PASS++))
else
    echo -e "${YELLOW}[WARN]${NC} kubectl 未安装 (可选)"
fi

command -v go &> /dev/null
check "Go 已安装"

command -v node &> /dev/null
check "Node.js 已安装"

echo ""

# 6. YAML 语法检查 (如果有 yamllint)
echo "6. YAML 语法检查..."
if command -v yamllint &> /dev/null; then
    yamllint -d relaxed k8s/*.yaml 2>/dev/null
    check "Kubernetes YAML 语法正确"
    
    yamllint -d relaxed .github/workflows/*.yml 2>/dev/null
    check "GitHub Actions YAML 语法正确"
else
    echo -e "${YELLOW}[SKIP]${NC} yamllint 未安装，跳过 YAML 语法检查"
fi

echo ""

# 7. Docker 配置验证
echo "7. Docker 配置验证..."
if command -v docker &> /dev/null; then
    docker-compose config &> /dev/null
    check "docker-compose.yml 语法正确"
else
    echo -e "${YELLOW}[SKIP]${NC} Docker 未运行，跳过 Docker 配置验证"
fi

echo ""

# 总结
echo "=========================================="
echo "验证结果汇总"
echo "=========================================="
echo -e "${GREEN}通过: $PASS${NC}"
echo -e "${RED}失败: $FAIL${NC}"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}所有检查通过！配置验证成功。${NC}"
    exit 0
else
    echo -e "${RED}有 $FAIL 项检查失败，请检查相关配置。${NC}"
    exit 1
fi
