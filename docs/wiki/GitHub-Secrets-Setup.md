# GitHub Secrets 配置指南

本文档说明如何配置 GitHub Secrets 用于 CI/CD 工作流。

## 需要配置的 Secrets

| Secret 名称 | 用途 | 必需 |
|-------------|------|------|
| `GITHUB_TOKEN` | GitHub API 访问令牌 | 自动提供 |
| `DOCKER_USERNAME` | Docker Hub 用户名 | 可选（用于推送镜像） |
| `DOCKER_PASSWORD` | Docker Hub 密码/令牌 | 可选（用于推送镜像） |
| `CODECOV_TOKEN` | Codecov 上传令牌 | 可选（用于代码覆盖率） |
| `SLACK_WEBHOOK` | Slack 通知 Webhook | 可选（用于通知） |

## 配置方法

### 方式一：GitHub 网页界面

1. 打开仓库页面：https://github.com/jiejack/new_energy
2. 点击 **Settings** → **Secrets and variables** → **Actions**
3. 点击 **New repository secret**
4. 输入 Secret 名称和值
5. 点击 **Add secret**

### 方式二：GitHub CLI

```bash
# 登录 GitHub CLI
gh auth login

# 添加 Secrets
gh secret set DOCKER_USERNAME
gh secret set DOCKER_PASSWORD
gh secret set CODECOV_TOKEN

# 查看 Secrets 列表
gh secret list
```

### 方式三：GitHub API

```bash
# 使用 curl 和 Personal Access Token
curl -X PUT \
  -H "Authorization: token YOUR_PAT" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/jiejack/new_energy/actions/secrets/DOCKER_USERNAME \
  -d '{"encrypted_value":"BASE64_ENCODED_VALUE"}'
```

## 创建 Docker Hub 访问令牌

1. 登录 [Docker Hub](https://hub.docker.com/)
2. 点击 **Account Settings** → **Security**
3. 点击 **New Access Token**
4. 选择权限：Read, Write, Delete
5. 复制生成的令牌

## 创建 GitHub Personal Access Token

1. 打开 GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. 点击 **Generate new token**
3. 选择权限：
   - `repo` (完整仓库访问)
   - `workflow` (工作流访问)
   - `write:packages` (包写入)
   - `read:packages` (包读取)
   - `delete:packages` (包删除)
4. 点击 **Generate token**
5. 复制生成的令牌

## 验证 Secrets 配置

```bash
# 使用 GitHub CLI 验证
gh secret list

# 预期输出
DOCKER_USERNAME  Updated 2026-04-07
DOCKER_PASSWORD  Updated 2026-04-07
```

## CI/CD 工作流使用示例

```yaml
# .github/workflows/ci.yml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
```

## 安全注意事项

1. **不要在代码中硬编码 Secrets**
2. **定期轮换 Secrets**
3. **使用最小权限原则**
4. **不要在日志中暴露 Secrets**
5. **使用 GitHub 的加密存储**

## 故障排查

### Secret 未生效

- 检查 Secret 名称是否正确（区分大小写）
- 确认 Secret 已添加到正确的仓库
- 检查工作流中的引用语法 `${{ secrets.SECRET_NAME }}`

### Docker 登录失败

- 确认 Docker Hub 用户名正确
- 确认访问令牌未过期
- 确认令牌权限足够

---

**最后更新**: 2026-04-07
