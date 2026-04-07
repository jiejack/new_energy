#!/bin/bash
# 安全审计脚本

echo "Running security audit..."

# 1. 运行 go vet
echo "=== go vet ==="
go vet ./...

# 2. 运行 golangci-lint 安全检查
echo "=== golangci-lint security ==="
golangci-lint run --enable=gosec ./...

# 3. 检查依赖漏洞
echo "=== Dependency vulnerabilities ==="
go list -m -json all | nancy sleuth

# 4. 检查敏感信息泄露
echo "=== Secrets scan ==="
gitleaks detect --source . --no-git

echo "Security audit complete!"
