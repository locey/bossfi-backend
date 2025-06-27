# BossFi Backend 部署指南

## 项目概述

BossFi Backend 是一个基于 Go 语言开发的区块链 DeFi 应用后端服务，提供用户管理、钱包登录、区块链数据同步等功能。

### 技术栈
- **后端**: Go 1.21 + Gin 框架
- **数据库**: PostgreSQL 15
- **缓存**: Redis 7
- **容器化**: Docker + Docker Compose
- **反向代理**: Nginx
- **区块链**: 以太坊 (go-ethereum)

## 部署架构

```
Internet
    ↓
  Nginx (Port 80/443)
    ↓
BossFi Backend (Port 8080)
    ↓
PostgreSQL (Port 5432) + Redis (Port 6379)
```

## 快速部署

### 1. 环境要求

- Docker >= 20.10
- Docker Compose >= 2.0
- Linux/Unix 系统
- 至少 2GB RAM
- 至少 10GB 磁盘空间

### 2. 基础设施部署

首先部署基础设施服务（Nginx、PostgreSQL、Redis）：

```bash
# 给脚本执行权限
chmod +x deploy-infrastructure.sh

# 部署基础设施
./deploy-infrastructure.sh
```

### 3. 应用部署

配置环境变量并部署应用：

```bash
# 复制环境变量模板
cp env.example .env

# 编辑环境变量（重要：修改数据库密码、JWT密钥等）
nano .env

# 给脚本执行权限
chmod +x deploy-app.sh

# 部署应用
./deploy-app.sh deploy
```

## 详细配置

### 环境变量配置

重要的环境变量需要在 `.env` 文件中配置：

```bash
# 数据库配置
DB_PASSWORD=your-secure-password

# JWT 配置 - 必须修改为强密码
JWT_SECRET=your-super-secure-jwt-secret-key

# 区块链配置 - 填入实际值
BLOCKCHAIN_RPC_URL=https://mainnet.infura.io/v3/your-project-id
CONTRACT_ADDRESS=0xYourContractAddress
PRIVATE_KEY=your-private-key
```

### SSL证书配置（生产环境）

1. 将SSL证书文件放入 `nginx/ssl/` 目录：
   ```
   nginx/ssl/
   ├── bossfi.com.crt
   └── bossfi.com.key
   ```

2. 修改 `nginx/conf.d/bossfi.conf`，启用HTTPS配置块

3. 重启Nginx：
   ```bash
   docker-compose -f docker-compose.infrastructure.yml restart nginx
   ```

## 运维管理

### 应用管理命令

```bash
# 查看应用状态
./deploy-app.sh status

# 查看日志
./deploy-app.sh logs

# 重启应用
./deploy-app.sh restart

# 更新应用
./deploy-app.sh update

# 停止应用
./deploy-app.sh stop

# 备份数据
./deploy-app.sh backup

# 回滚应用
./deploy-app.sh rollback
```

### 基础设施管理

```bash
# 查看基础设施状态
docker-compose -f docker-compose.infrastructure.yml ps

# 查看特定服务日志
docker-compose -f docker-compose.infrastructure.yml logs -f postgres
docker-compose -f docker-compose.infrastructure.yml logs -f redis
docker-compose -f docker-compose.infrastructure.yml logs -f nginx

# 重启特定服务
docker-compose -f docker-compose.infrastructure.yml restart postgres

# 进入容器
docker-compose -f docker-compose.infrastructure.yml exec postgres psql -U postgres -d bossfi
docker-compose -f docker-compose.infrastructure.yml exec redis redis-cli
```

### 数据库管理

```bash
# 连接数据库
docker-compose -f docker-compose.infrastructure.yml exec postgres psql -U postgres -d bossfi

# 备份数据库
docker-compose -f docker-compose.infrastructure.yml exec postgres pg_dump -U postgres -d bossfi > backup.sql

# 恢复数据库
docker-compose -f docker-compose.infrastructure.yml exec -T postgres psql -U postgres -d bossfi < backup.sql
```

### 监控和日志

1. **应用监控**：
   - 健康检查端点：`http://localhost:8080/api/v1/health`
   - API文档：`http://localhost:8080/swagger/index.html`

2. **日志查看**：
   ```bash
   # 应用日志
   docker-compose -f docker-compose.app.yml logs -f

   # Nginx访问日志
   docker-compose -f docker-compose.infrastructure.yml exec nginx tail -f /var/log/nginx/access.log

   # PostgreSQL日志
   docker-compose -f docker-compose.infrastructure.yml logs -f postgres
   ```

## 性能优化

### 数据库优化

在 `docker-compose.infrastructure.yml` 中调整 PostgreSQL 配置：

```yaml
postgres:
  environment:
    - POSTGRES_SHARED_PRELOAD_LIBRARIES=pg_stat_statements
  command: >
    postgres
    -c shared_preload_libraries=pg_stat_statements
    -c max_connections=200
    -c shared_buffers=256MB
    -c effective_cache_size=1GB
```

### Redis优化

```yaml
redis:
  command: >
    redis-server
    --appendonly yes
    --requirepass redis123
    --maxmemory 512mb
    --maxmemory-policy allkeys-lru
```

## 安全配置

### 1. 修改默认密码

```bash
# 数据库密码
DB_PASSWORD=your-secure-database-password

# Redis密码
REDIS_PASSWORD=your-secure-redis-password

# JWT密钥
JWT_SECRET=your-super-secure-jwt-secret-minimum-32-characters
```

### 2. 网络安全

- 仅暴露必要端口（80, 443）
- 使用防火墙限制访问
- 启用SSL/TLS加密

### 3. 容器安全

- 使用非root用户运行容器
- 定期更新基础镜像
- 限制容器资源使用

## 故障排除

### 常见问题

1. **应用启动失败**：
   ```bash
   # 查看详细日志
   ./deploy-app.sh logs
   
   # 检查环境变量
   docker-compose -f docker-compose.app.yml config
   ```

2. **数据库连接失败**：
   ```bash
   # 检查数据库状态
   docker-compose -f docker-compose.infrastructure.yml exec postgres pg_isready -U postgres
   
   # 检查网络连接
   docker network inspect bossfi-network
   ```

3. **Redis连接失败**：
   ```bash
   # 测试Redis连接
   docker-compose -f docker-compose.infrastructure.yml exec redis redis-cli ping
   ```

### 日志分析

- 应用日志位置：容器内 `/app/logs/`
- Nginx日志：容器内 `/var/log/nginx/`
- PostgreSQL日志：通过 `docker logs` 查看

## 备份策略

### 自动备份

建议设置定时备份：

```bash
# 创建备份脚本
cat > /etc/cron.daily/bossfi-backup << 'EOF'
#!/bin/bash
cd /path/to/bossfi-backend
./deploy-app.sh backup
# 清理30天前的备份
find backups/ -type d -mtime +30 -exec rm -rf {} \;
EOF

chmod +x /etc/cron.daily/bossfi-backup
```

### 备份恢复

```bash
# 列出可用备份
ls -la backups/

# 恢复到特定备份
./deploy-app.sh rollback
```

## 更新升级

### 应用更新

```bash
# 拉取最新代码
git pull

# 更新应用
./deploy-app.sh update
```

### 系统升级

1. 备份数据
2. 更新Docker镜像
3. 重新部署
4. 验证功能

## 联系支持

如有问题，请查看：
- 项目文档
- GitHub Issues
- 联系开发团队 