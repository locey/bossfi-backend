# BossFi Backend 部署指南

这是 BossFi 后端项目的 Linux Docker 部署方案，分为基础环境部署和项目部署两个阶段。

## 📋 系统要求

- **操作系统**: Ubuntu 18.04+ / Debian 10+
- **内存**: 至少 2GB RAM
- **磁盘**: 至少 20GB 可用空间
- **网络**: 需要访问 GitHub 和 Docker Hub

## 🚀 快速开始

### 1. 基础环境安装

首先在服务器上安装基础环境（Docker、Docker Compose、Nginx等）：

```bash
# 下载部署脚本
wget https://raw.githubusercontent.com/locey/bossfi-backend/dev/deploy/setup-environment.sh

# 或者从项目中复制
# 上传 setup-environment.sh 到服务器

# 设置执行权限
chmod +x setup-environment.sh

# 运行基础环境安装（需要root权限）
sudo ./setup-environment.sh
```

### 2. 项目部署

基础环境安装完成后，部署 BossFi 项目：

```bash
# 进入项目目录
cd /opt/bossfi

# 如果已经有项目代码，进入部署目录
cd bossfi-backend/deploy

# 设置执行权限
chmod +x *.sh

# 初始化部署项目
sudo ./deploy-project.sh init
```

## 📁 文件结构

```
/opt/bossfi/
├── bossfi-backend/          # 项目代码目录
│   ├── deploy/              # 部署配置目录
│   │   ├── setup-environment.sh    # 基础环境安装脚本
│   │   ├── deploy-project.sh       # 项目部署脚本
│   │   ├── monitor.sh              # 监控脚本
│   │   ├── docker-compose.yml      # Docker编排配置
│   │   ├── nginx.conf              # Nginx配置
│   │   └── init.sql                # 数据库初始化脚本
│   └── .env                 # 环境变量配置
├── logs/                    # 日志文件目录
├── backups/                 # 备份文件目录
├── ssl/                     # SSL证书目录
└── data/                    # 数据持久化目录
    ├── postgres/            # PostgreSQL数据
    └── redis/               # Redis数据
```

## 🔧 配置说明

### 环境变量配置

项目使用 `.env` 文件管理环境变量，主要配置项：

```bash
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_USER=bossfi_user
DB_PASSWORD=your_secure_password
DB_NAME=bossfi

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379

# JWT配置
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRE_HOURS=24

# 应用配置
PORT=8080
GIN_MODE=release
LOG_LEVEL=info

# 区块链配置
BLOCKCHAIN_RPC_URL=https://mainnet.infura.io/v3/your-project-id
CONTRACT_ADDRESS=0x1234567890123456789012345678901234567890
```

### Docker 服务配置

- **PostgreSQL**: 512MB内存限制，数据持久化
- **Redis**: 128MB内存限制，AOF持久化
- **BossFi Backend**: 512MB内存限制，健康检查
- **Nginx**: 64MB内存限制，反向代理和负载均衡

## 📋 部署命令

### 基础环境脚本

```bash
# 安装基础环境
sudo ./setup-environment.sh
```

### 项目部署脚本

```bash
# 初始化部署
sudo ./deploy-project.sh init

# 更新项目
sudo ./deploy-project.sh update

# 重启服务
sudo ./deploy-project.sh restart

# 停止服务
sudo ./deploy-project.sh stop

# 查看状态
sudo ./deploy-project.sh status

# 查看日志
sudo ./deploy-project.sh logs [backend|postgres|redis|nginx]

# 手动备份
sudo ./deploy-project.sh backup
```

### 监控脚本

```bash
# 查看服务状态
./monitor.sh status

# 完整健康检查
./monitor.sh health

# 查看实时日志
./monitor.sh logs [all|backend|postgres|redis|nginx]

# 重启特定服务
./monitor.sh restart [all|backend|postgres|redis|nginx]

# 查看资源使用
./monitor.sh resource

# 检查网络连接
./monitor.sh network

# 显示性能统计
./monitor.sh perf
```

## 🔍 常用操作

### 查看服务状态

```bash
# 查看所有容器状态
docker ps

# 查看服务状态
./monitor.sh status

# 查看资源使用
docker stats
```

### 查看日志

```bash
# 查看所有服务日志
./monitor.sh logs all

# 查看后端服务日志
./monitor.sh logs backend

# 查看Nginx访问日志
tail -f /opt/bossfi/logs/access.log

# 查看错误日志
tail -f /opt/bossfi/logs/error.log
```

### 备份和恢复

```bash
# 手动备份
sudo ./deploy-project.sh backup

# 查看备份文件
ls -la /opt/bossfi/backups/

# 恢复数据库（示例）
docker exec -i bossfi-postgres psql -U bossfi_user -d bossfi < /opt/bossfi/backups/db_backup_20241201_120000.sql
```

## 🌐 访问地址

部署完成后，可以通过以下地址访问：

- **API接口**: `http://your-server-ip/api/`
- **健康检查**: `http://your-server-ip/health`
- **API文档**: `http://your-server-ip/swagger/index.html`

### API接口示例

```bash
# 健康检查
curl http://your-server-ip/health

# 获取签名消息
curl -X POST http://your-server-ip/api/auth/nonce \
  -H "Content-Type: application/json" \
  -d '{"wallet_address": "0x1234567890123456789012345678901234567890"}'

# 钱包登录
curl -X POST http://your-server-ip/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0x1234567890123456789012345678901234567890",
    "signature": "0x...",
    "message": "..."
  }'
```

## 🛠️ 故障排查

### 常见问题

1. **容器启动失败**
   ```bash
   # 查看容器日志
   docker logs bossfi-backend
   
   # 检查配置
   docker-compose config
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker exec bossfi-postgres pg_isready -U bossfi_user
   
   # 查看数据库日志
   ./monitor.sh logs postgres
   ```

3. **内存不足**
   ```bash
   # 查看内存使用
   free -h
   docker stats
   
   # 调整容器内存限制
   vim docker-compose.yml
   ```

4. **端口冲突**
   ```bash
   # 检查端口占用
   netstat -tlnp | grep :80
   netstat -tlnp | grep :8080
   ```

### 日志位置

- **应用日志**: `/opt/bossfi/logs/`
- **Nginx日志**: `/opt/bossfi/logs/access.log`, `/opt/bossfi/logs/error.log`
- **Docker日志**: `docker logs <container_name>`

## 🔐 安全配置

### 防火墙配置

```bash
# 查看防火墙状态
sudo ufw status

# 允许特定IP访问
sudo ufw allow from YOUR_IP to any port 22
sudo ufw allow from YOUR_IP to any port 80
sudo ufw allow from YOUR_IP to any port 443
```

### SSL证书配置

```bash
# 安装Certbot
sudo apt install certbot python3-certbot-nginx

# 获取SSL证书
sudo certbot --nginx -d your-domain.com

# 自动续期
sudo crontab -e
# 添加: 0 12 * * * /usr/bin/certbot renew --quiet
```

## 📊 监控和维护

### 定期维护任务

```bash
# 清理Docker资源
docker system prune -f

# 清理旧日志
find /opt/bossfi/logs -name "*.log" -mtime +30 -delete

# 清理旧备份
find /opt/bossfi/backups -mtime +7 -delete

# 更新系统
sudo apt update && sudo apt upgrade -y
```

### 性能监控

```bash
# 系统资源监控
./monitor.sh perf

# 实时监控
htop
iotop
nethogs
```

## 🆘 支持

如果遇到问题，请：

1. 查看日志文件排查问题
2. 运行健康检查：`./monitor.sh health`
3. 检查服务状态：`./monitor.sh status`
4. 查看项目文档和GitHub Issues

## 📝 更新日志

- **v1.0.0**: 初始版本，包含基础部署功能
- 支持 Docker 容器化部署
- 包含完整的监控和备份功能
- 支持一键部署和更新 