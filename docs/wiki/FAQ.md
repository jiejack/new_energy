# 常见问题解答 (FAQ)

本文档收集了新能源监控系统使用过程中的常见问题及解决方案。

## 安装部署

### Q1: Docker启动失败，提示端口被占用？

**问题**: 执行 `docker-compose up -d` 时报端口占用错误。

**解决方案**:
1. 检查端口占用情况：
   ```bash
   # Linux/Mac
   lsof -i :8080
   lsof -i :5432
   lsof -i :6379
   
   # Windows
   netstat -ano | findstr :8080
   ```

2. 停止占用端口的服务或修改 `docker-compose.yml` 中的端口映射。

### Q2: 数据库连接失败？

**问题**: 日志显示 `connection refused` 或 `authentication failed`。

**解决方案**:
1. 确认PostgreSQL服务已启动：
   ```bash
   docker-compose ps postgres
   ```

2. 检查连接参数：
   ```bash
   # 测试连接
   docker exec -it nem-postgres psql -U nem -d nem_system
   ```

3. 检查 `.env` 文件中的数据库配置是否正确。

### Q3: 前端页面无法访问后端API？

**问题**: 前端控制台报跨域错误 (CORS)。

**解决方案**:
1. 检查后端CORS配置：
   ```yaml
   # configs/config.yaml
   server:
     cors:
       allowed_origins:
         - "http://localhost"
         - "http://localhost:80"
   ```

2. 确保前端请求的API地址正确。

### Q4: Kubernetes部署时镜像拉取失败？

**问题**: Pod状态为 `ImagePullBackOff`。

**解决方案**:
1. 检查镜像是否存在：
   ```bash
   docker images | grep nem
   ```

2. 如果使用私有仓库，配置imagePullSecrets：
   ```bash
   kubectl create secret docker-registry regcred \
     --docker-server=<registry> \
     --docker-username=<user> \
     --docker-password=<password>
   ```

---

## 功能使用

### Q5: 如何重置管理员密码？

**解决方案**:
```bash
# 进入后端容器
docker exec -it nem-backend /bin/sh

# 执行密码重置
./nem reset-password --username admin --password NewPassword@123
```

### Q6: 告警通知未收到？

**问题**: 告警触发但未收到邮件/短信通知。

**排查步骤**:
1. 检查通知渠道配置是否正确：
   - 进入 **告警管理** → **通知配置**
   - 测试连接是否成功

2. 检查告警规则的通知渠道是否配置：
   - 进入 **告警管理** → **告警规则**
   - 确认规则已关联通知渠道

3. 检查用户是否配置了联系方式：
   - 进入 **系统设置** → **用户管理**
   - 确认用户邮箱/手机号已填写

### Q7: 数据采集延迟？

**问题**: 实时数据更新慢，有延迟。

**排查步骤**:
1. 检查采集服务状态：
   ```bash
   docker-compose logs nem-collector
   ```

2. 检查设备通信状态：
   - 进入 **设备管理** → **设备管理**
   - 查看设备连接状态

3. 检查网络延迟：
   ```bash
   ping <device_ip>
   ```

### Q8: 如何批量导入设备？

**解决方案**:
1. 准备CSV文件：
   ```csv
   name,type,station_id,protocol,address,port
   逆变器#1,inverter,uuid,modbus_tcp,192.168.1.100,502
   逆变器#2,inverter,uuid,modbus_tcp,192.168.1.101,502
   ```

2. 进入 **设备管理** → **设备管理** → **导入**

3. 上传CSV文件并确认导入。

### Q9: 历史数据查询慢？

**问题**: 查询长时间范围数据响应慢。

**解决方案**:
1. 使用聚合查询减少数据量：
   - 选择合适的采样间隔（5分钟/1小时/1天）

2. 优化查询时间范围：
   - 避免一次性查询超过30天的原始数据

3. 检查数据库索引：
   ```sql
   EXPLAIN ANALYZE SELECT * FROM timeseries_data WHERE device_id = 'xxx' AND time BETWEEN '...' AND '...';
   ```

---

## 性能优化

### Q10: 系统响应慢？

**排查步骤**:
1. 检查资源使用：
   ```bash
   docker stats
   kubectl top pods
   ```

2. 检查数据库连接池：
   ```sql
   SELECT count(*) FROM pg_stat_activity;
   ```

3. 检查Redis内存使用：
   ```bash
   redis-cli info memory
   ```

4. 查看慢查询日志：
   ```bash
   docker exec nem-postgres cat /var/log/postgresql/slow.log
   ```

### Q11: 内存占用过高？

**解决方案**:
1. 调整JVM/Go内存限制：
   ```yaml
   # docker-compose.yml
   environment:
     - GOMEMLIMIT=4GiB
   ```

2. 配置资源限制：
   ```yaml
   deploy:
     resources:
       limits:
         memory: 512M
   ```

3. 检查内存泄漏：
   ```bash
   curl http://localhost:6060/debug/pprof/heap > heap.out
   go tool pprof heap.out
   ```

---

## 安全相关

### Q12: 如何修改JWT密钥？

**解决方案**:
1. 生成新密钥：
   ```bash
   openssl rand -base64 32
   ```

2. 更新配置：
   ```env
   JWT_SECRET=your_new_secret_key
   ```

3. 重启服务使配置生效。

### Q13: 如何配置HTTPS？

**解决方案**:
1. 准备SSL证书：
   ```bash
   # 使用Let's Encrypt
   certbot certonly --standalone -d your-domain.com
   ```

2. 配置Nginx：
   ```nginx
   server {
       listen 443 ssl;
       ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
       ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
   }
   ```

### Q14: 如何限制登录失败次数？

**解决方案**:
系统默认配置：
- 连续失败5次锁定账户15分钟

修改配置：
```yaml
# configs/config.yaml
auth:
  max_login_attempts: 5
  lockout_duration: 15m
```

---

## 数据备份

### Q15: 如何备份数据库？

**解决方案**:
```bash
# 手动备份
docker exec nem-postgres pg_dump -U nem nem_system > backup_$(date +%Y%m%d).sql

# 自动备份脚本
./scripts/backup.sh
```

### Q16: 如何恢复数据？

**解决方案**:
```bash
# 恢复数据库
cat backup_20260407.sql | docker exec -i nem-postgres psql -U nem nem_system
```

---

## 开发相关

### Q17: 如何本地调试？

**解决方案**:
1. 启动依赖服务：
   ```bash
   docker-compose up -d postgres redis
   ```

2. 配置环境变量：
   ```bash
   export DB_HOST=localhost
   export REDIS_HOST=localhost
   ```

3. 启动后端：
   ```bash
   go run cmd/api-server/main.go
   ```

4. 启动前端：
   ```bash
   cd web && npm run dev
   ```

### Q18: 如何运行测试？

**解决方案**:
```bash
# 后端测试
go test ./... -v -cover

# 前端测试
cd web && npm run test

# E2E测试
cd web && npm run test:e2e
```

---

## 其他问题

### Q19: 如何查看系统日志？

**解决方案**:
```bash
# Docker日志
docker-compose logs -f nem-backend

# Kubernetes日志
kubectl logs -f deployment/nem-backend -n nem-system

# 应用日志文件
tail -f /var/log/nem/app.log
```

### Q20: 如何获取技术支持？

**渠道**:
1. GitHub Issues: [提交问题](https://github.com/jiejack/new_energy/issues)
2. 文档中心: 查阅相关文档
3. 邮件支持: support@example.com

---

## 问题未解决？

如果您的问题不在上述列表中，请：

1. 查阅 [故障排查指南](./Troubleshooting)
2. 搜索 [GitHub Issues](https://github.com/jiejack/new_energy/issues)
3. 提交新的 Issue，包含以下信息：
   - 系统版本
   - 操作系统
   - 错误日志
   - 复现步骤

---

**最后更新**: 2026-04-07
