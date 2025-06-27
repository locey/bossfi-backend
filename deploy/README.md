# BossFi Backend 部署方案

## 🚀 部署架构

### 服务器配置
- **CPU**: 2核
- **内存**: 2GB (1975M)
- **磁盘**: 40GB
- **同时运行**: 后端 + 前端项目

### 资源分配
```
总资源：2GB内存，2核CPU
├── 系统预留：400MB内存
├── Nginx：50MB内存
├── PostgreSQL：400MB内存
├── Redis：100MB内存  
├── BossFi后端：300MB内存
└── 前端项目：剩余资源（~750MB）
```

## 📁 目录结构

```
/opt/bossfi/
├── docker-compose.prod.yml    # Docker Compose 配置
├── nginx.conf                 # Nginx 配置
├── env.prod                   # 环境变量
├── init.sql                   # 数据库初始化脚本
├── deploy.sh                  # 部署脚本
├── monitor.sh                 # 监控脚本
├── logs/                      # 日志目录
├── ssl/                       # SSL证书目录
└── backups/                   # 备份目录

/var/www/frontend/             # 前端静态文件目录
```

## 🛠️ 部署步骤

### 1. 服务器环境准备

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.21.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 重新登录以应用Docker组权限
newgrp docker
```

### 2. 上传项目文件

```bash
# 方法1: 使用Git
cd /opt
sudo git clone https://github.com/your-repo/bossfi-backend.git bossfi
sudo chown -R $USER:$USER /opt/bossfi

# 方法2: 使用SCP上传
scp -r ./bossfi-backend/deploy/* user@your-server:/opt/bossfi/
```

### 3. 配置环境变量

```bash
cd /opt/bossfi
cp env.prod .env

# 编辑环境变量
sudo nano .env
```

**重要配置项：**
```bash
# 修改数据库密码
DB_PASSWORD=your_secure_password_here

# 修改JWT密钥（至少32字符）
JWT_SECRET=your_super_secret_jwt_key_32_chars_min

# 配置区块链RPC
BLOCKCHAIN_RPC_URL=https://mainnet.infura.io/v3/your-project-id
CONTRACT_ADDRESS=0x你的合约地址

# 私钥（如果需要）
PRIVATE_KEY=你的私钥
```

### 4. 执行部署

```bash
# 赋予执行权限
chmod +x deploy.sh monitor.sh

# 执行部署
sudo ./deploy.sh prod
```

### 5. 前端项目部署

```bash
# 创建前端目录
sudo mkdir -p /var/www/frontend

# 上传前端构建文件到 /var/www/frontend
# 例如：
scp -r ./frontend/dist/* user@your-server:/var/www/frontend/

# 设置权限
sudo chown -R www-data:www-data /var/www/frontend
sudo chmod -R 755 /var/www/frontend
```

## 🔧 管理命令

### 监控服务状态
```bash
cd /opt/bossfi

# 检查服务状态
./monitor.sh status

# 检查资源使用
./monitor.sh resources

# 查看日志
./monitor.sh logs

# 查看特定服务日志
./monitor.sh logs bossfi-backend
```

### 常用Docker命令
```bash
cd /opt/bossfi

# 查看容器状态
docker-compose -f docker-compose.prod.yml ps

# 重启所有服务
docker-compose -f docker-compose.prod.yml restart

# 重启特定服务
docker-compose -f docker-compose.prod.yml restart bossfi-backend

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f bossfi-backend

# 进入容器
docker exec -it bossfi-backend /bin/sh
docker exec -it bossfi-postgres psql -U bossfi_user -d bossfi
```

### 数据库管理
```bash
# 连接数据库
docker exec -it bossfi-postgres psql -U bossfi_user -d bossfi

# 备份数据库
./monitor.sh backup

# 手动备份
docker exec bossfi-postgres pg_dump -U bossfi_user bossfi > backup.sql

# 恢复数据库
docker exec -i bossfi-postgres psql -U bossfi_user -d bossfi < backup.sql
```

## 🔍 访问地址

部署完成后，可以通过以下地址访问：

- **前端**: `http://your-server-ip`
- **API**: `http://your-server-ip/api`
- **健康检查**: `http://your-server-ip/health`
- **API文档**: `http://your-server-ip/swagger/index.html`

## 📊 性能优化

### 1. 数据库优化
```sql
-- 在PostgreSQL中执行
-- 优化配置（根据2GB内存）
ALTER SYSTEM SET shared_buffers = '128MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;

SELECT pg_reload_conf();
```

### 2. Redis优化
Redis配置已在docker-compose.yml中优化：
- 最大内存：80MB
- 内存策略：allkeys-lru
- 持久化：AOF

### 3. Nginx优化
- 启用Gzip压缩
- 静态文件缓存
- 连接池
- 限流保护

## 🔒 安全配置

### 1. 防火墙设置
```bash
# 安装UFW
sudo apt install ufw

# 配置防火墙
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. SSL证书（可选）
```bash
# 安装Certbot
sudo apt install certbot python3-certbot-nginx

# 获取SSL证书
sudo certbot --nginx -d your-domain.com

# 自动续期
sudo crontab -e
# 添加：0 12 * * * /usr/bin/certbot renew --quiet
```

### 3. 定期更新
```bash
# 创建更新脚本
cat > /opt/bossfi/update.sh << 'EOF'
#!/bin/bash
cd /opt/bossfi
./monitor.sh backup
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
docker system prune -f
EOF

chmod +x /opt/bossfi/update.sh

# 添加定时任务（每周日凌晨3点更新）
sudo crontab -e
# 添加：0 3 * * 0 /opt/bossfi/update.sh
```

## 🚨 故障排除

### 常见问题

1. **容器启动失败**
   ```bash
   # 查看详细错误
   docker-compose -f docker-compose.prod.yml logs
   
   # 检查配置文件
   docker-compose -f docker-compose.prod.yml config
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker exec bossfi-postgres pg_isready -U bossfi_user -d bossfi
   
   # 查看数据库日志
   docker-compose -f docker-compose.prod.yml logs postgres
   ```

3. **内存不足**
   ```bash
   # 查看内存使用
   free -h
   docker stats
   
   # 清理Docker缓存
   docker system prune -a
   ```

4. **磁盘空间不足**
   ```bash
   # 查看磁盘使用
   df -h
   du -sh /opt/bossfi/* | sort -h
   
   # 清理日志
   docker-compose -f docker-compose.prod.yml logs --tail=0
   journalctl --vacuum-time=7d
   ```

### 紧急恢复

```bash
# 快速重启所有服务
cd /opt/bossfi
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d

# 恢复数据库备份
docker exec -i bossfi-postgres psql -U bossfi_user -d bossfi < /opt/bossfi/backups/latest_backup.sql
```

## 📈 监控告警

### 设置监控脚本
```bash
# 创建监控检查脚本
cat > /opt/bossfi/health_check.sh << 'EOF'
#!/bin/bash
cd /opt/bossfi

# 检查服务健康状态
if ! curl -f -s http://localhost/health >/dev/null; then
    echo "$(date): BossFi Backend服务异常" >> /var/log/bossfi-alert.log
    # 可以在这里添加邮件或短信通知
fi

# 检查内存使用
MEMORY_USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
if [ $MEMORY_USAGE -gt 85 ]; then
    echo "$(date): 内存使用率过高: ${MEMORY_USAGE}%" >> /var/log/bossfi-alert.log
fi

# 检查磁盘使用
DISK_USAGE=$(df / | awk 'NR==2{print $5}' | cut -d'%' -f1)
if [ $DISK_USAGE -gt 85 ]; then
    echo "$(date): 磁盘使用率过高: ${DISK_USAGE}%" >> /var/log/bossfi-alert.log
fi
EOF

chmod +x /opt/bossfi/health_check.sh

# 添加定时检查（每5分钟）
sudo crontab -e
# 添加：*/5 * * * * /opt/bossfi/health_check.sh
```

## 📝 维护计划

### 定期维护任务

1. **每天**: 自动备份数据库
2. **每周**: 检查日志，清理临时文件
3. **每月**: 更新系统包，检查安全补丁
4. **每季度**: 性能调优，容量规划

### 升级步骤

1. 备份数据
2. 测试新版本
3. 滚动升级
4. 验证功能
5. 监控运行状态

这个部署方案专门为您的2GB内存服务器优化，确保后端和前端项目都能稳定运行。 