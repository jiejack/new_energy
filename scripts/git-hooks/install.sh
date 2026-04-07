#!/bin/bash
# 安装 Git Hooks
# 位置: scripts/git-hooks/install.sh

set -e

echo "========================================="
echo "安装 Git Hooks"
echo "========================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
GIT_HOOKS_DIR="$PROJECT_ROOT/.git/hooks"

# 检查 .git 目录是否存在
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo -e "${RED}错误: 未找到 .git 目录，请确保在 Git 仓库中运行此脚本${NC}"
    exit 1
fi

# 创建 hooks 目录（如果不存在）
mkdir -p "$GIT_HOOKS_DIR"

# 安装 hooks
HOOKS=("pre-commit" "pre-push" "commit-msg")

for hook in "${HOOKS[@]}"; do
    SOURCE="$SCRIPT_DIR/$hook"
    TARGET="$GIT_HOOKS_DIR/$hook"

    if [ -f "$SOURCE" ]; then
        # 备份现有的 hook
        if [ -f "$TARGET" ]; then
            echo -e "${YELLOW}备份现有的 $hook...${NC}"
            mv "$TARGET" "$TARGET.backup.$(date +%Y%m%d%H%M%S)"
        fi

        # 复制新的 hook
        cp "$SOURCE" "$TARGET"

        # 设置可执行权限
        chmod +x "$TARGET"

        echo -e "${GREEN}✓ 已安装 $hook${NC}"
    else
        echo -e "${RED}错误: 未找到 $hook 文件${NC}"
    fi
done

echo ""
echo "========================================="
echo -e "${GREEN}Git Hooks 安装完成！${NC}"
echo ""
echo "已安装的 Hooks:"
echo "  - pre-commit:  代码提交前检查（格式、Lint、快速测试）"
echo "  - pre-push:    代码推送前检查（完整测试、覆盖率）"
echo "  - commit-msg:  提交信息格式检查"
echo ""
echo "提示:"
echo "  - 可以通过 'git commit --no-verify' 跳过 pre-commit 检查"
echo "  - 可以通过 'git push --no-verify' 跳过 pre-push 检查"
echo "  - 建议不要跳过检查，以保持代码质量"
echo "========================================="
