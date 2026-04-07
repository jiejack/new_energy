# GitHub Wiki 同步脚本

本脚本用于将 docs/wiki/ 目录下的文档同步到 GitHub Wiki。

## 使用方法

### 方式一：手动同步（推荐）

1. 在 GitHub 仓库页面，点击 **Settings** → 徛 **Features** → **Wiki** → 启用 Wiki
2. 克隆 Wiki 仓库：
   ```bash
   git clone https://github.com/jiejack/new_energy.wiki.git
   ```
3. 复制文档：
   ```bash
   cp docs/wiki/*.md new_energy.wiki/
   cd new_energy.wiki
   git add .
   git commit -m "docs: sync wiki documentation"
   git push origin master
   ```

### 方式二：使用 GitHub CLI

```bash
# 安装 GitHub CLI
# Windows: winget install GitHub.cli
# Mac: brew install gh
# Linux: sudo apt install gh

# 登录
gh auth login

# 启用 Wiki
gh api repos/jiejack/new_energy --method PATCH -f has_wiki=true

# 创建 Wiki 页面
gh api repos/jiejack/new_energy/wiki/Home -X PUT -f content=@docs/wiki/Home.md
gh api repos/jiejack/new_energy/wiki/Installation-Guide -X PUT -f content=@docs/wiki/Installation-Guide.md
gh api repos/jiejack/new_energy/wiki/Quick-Start -X PUT -f content=@docs/wiki/Quick-Start.md
gh api repos/jiejack/new_energy/wiki/API-Documentation -X PUT -f content=@docs/wiki/API-Documentation.md
gh api repos/jiejack/new_energy/wiki/FAQ -X PUT -f content=@docs/wiki/FAQ.md
gh api repos/jiejack/new_energy/wiki/Project-Structure -X PUT -f content=@docs/wiki/Project-Structure.md
```

## Wiki 页面列表

| 页面 | 文件 | 描述 |
|-----|------|------|
| Home | Home.md | 项目首页 |
| Installation Guide | Installation-Guide.md | 安装指南 |
| Quick Start | Quick-Start.md | 快速开始 |
| API Documentation | API-Documentation.md | API文档 |
| FAQ | FAQ.md | 常见问题 |
| Project Structure | Project-Structure.md | 项目结构 |

## 注意事项

1. Wiki 页面名称不能包含空格，实际使用连字符
2. 文件名格式：`Page-Name.md`
3. 需要有仓库的写入权限
