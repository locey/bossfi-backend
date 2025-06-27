#!/bin/bash

# BossFi Backend 部署脚本
# 使用方法: ./deploy.sh [dev|prod]

set -e

# 获取环境参数
ENVIRONMENT=${1:-prod}
PROJECT_NAME="bossfi"
DEPLOY_DIR="/opt/bossfi"
BACKUP_DIR="/opt/bossfi/backups"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# 检查是否为root用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "此脚本需要root权限运行"
    fi
}

# 检查依赖
check_dependencies() {
    log "检查系统依赖..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        error "Docker 未安装，请先安装Docker"
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose 未安装，请先安装Docker Compose"
    fi
    
    # 检查Git
    if ! command -v git &> /dev/null; then
        error "Git 未安装，请先安装Git"
    fi
    
    log "依赖检查完成"
}

# 创建目录结构
create_directories() {
    log "创建目录结构..."
    
    mkdir -p $DEPLOY_DIR
    mkdir -p $BACKUP_DIR
    mkdir -p $DEPLOY_DIR/logs
    mkdir -p $DEPLOY_DIR/ssl
    mkdir -p /var/www/frontend
    
    log "目录结构创建完成"
}

# 备份数据
backup_data() {
    if [ -d "$DEPLOY_DIR" ]; then
        log "备份现有数据..."
        
        BACKUP_FILE="$BACKUP_DIR/backup_$(date +%Y%m%d_%H%M%S).tar.gz"
        
        # 备份数据库
        if docker ps | grep -q "bossfi-postgres"; then
            log "备份数据库..."
            docker exec bossfi-postgres pg_dump -U bossfi_user bossfi > $BACKUP_DIR/db_backup_$(date +%Y%m%d_%H%M%S).sql
        fi
        
        # 备份配置文件
        tar -czf $BACKUP_FILE -C $DEPLOY_DIR . 2>/dev/null || true
        
        log "数据备份完成: $BACKUP_FILE"
    fi
}

# 停止现有服务
stop_services() {
    log "停止现有服务..."
    
    cd $DEPLOY_DIR
    if [ -f "docker-compose.prod.yml" ]; then
        docker-compose -f docker-compose.prod.yml down || true
    fi
    
    # 清理未使用的容器和镜像
    docker system prune -f || true
    
    log "服务停止完成"
}

# 部署新版本
deploy_new_version() {
    log "部署新版本..."
    
    # 复制部署文件
    cp -r ./deploy/* $DEPLOY_DIR/
    
    # 设置权限
    chmod +x $DEPLOY_DIR/*.sh
    chmod 600 $DEPLOY_DIR/env.prod
    
    cd $DEPLOY_DIR
    
    # 构建并启动服务
    log "构建Docker镜像..."
    docker-compose -f docker-compose.prod.yml build --no-cache
    
    log "启动服务..."
    docker-compose -f docker-compose.prod.yml up -d
    
    log "等待服务启动..."
    sleep 30
    
    # 健康检查
    check_health
    
    log "部署完成"
}

# 健康检查
check_health() {
    log "执行健康检查..."
    
    # 检查容器状态
    if ! docker ps | grep -q "bossfi-backend"; then
        error "后端服务未启动"
    fi
    
    if ! docker ps | grep -q "bossfi-postgres"; then
        error "数据库服务未启动"
    fi
    
    if ! docker ps | grep -q "bossfi-redis"; then
        error "Redis服务未启动"
    fi
    
    if ! docker ps | grep -q "bossfi-nginx"; then
        error "Nginx服务未启动"
    fi
    
    # 检查健康检查端点
    for i in {1..10}; do
        if curl -f http://localhost/health >/dev/null 2>&1; then
            log "健康检查通过"
            return 0
        fi
        log "等待服务响应... ($i/10)"
        sleep 5
    done
    
    error "健康检查失败"
}

# 监控资源使用
monitor_resources() {
    log "监控资源使用情况..."
    
    # 内存使用
    MEMORY_USAGE=$(free -m | awk 'NR==2{printf "%.1f%%", $3*100/$2}')
    log "内存使用率: $MEMORY_USAGE"
    
    # 磁盘使用
    DISK_USAGE=$(df -h / | awk 'NR==2{print $5}')
    log "磁盘使用率: $DISK_USAGE"
    
    # Docker容器状态
    log "Docker容器状态:"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    
    # 如果内存使用超过80%，发出警告
    if (( $(echo "$MEMORY_USAGE" | cut -d'%' -f1 | cut -d'.' -f1) > 80 )); then
        warn "内存使用率过高: $MEMORY_USAGE"
    fi
}

# 主函数
main() {
    log "开始部署 BossFi Backend ($ENVIRONMENT 环境)"
    
    check_root
    check_dependencies
    create_directories
    backup_data
    stop_services
    deploy_new_version
    monitor_resources
    
    log "部署完成！"
    log "访问地址:"
    log "  - 前端: http://your-server-ip"
    log "  - API: http://your-server-ip/api"
    log "  - 健康检查: http://your-server-ip/health"
    log "  - API文档: http://your-server-ip/swagger/index.html"
    
    log "监控命令:"
    log "  - 查看日志: docker-compose -f $DEPLOY_DIR/docker-compose.prod.yml logs -f"
    log "  - 查看状态: docker ps"
    log "  - 重启服务: docker-compose -f $DEPLOY_DIR/docker-compose.prod.yml restart"
}

# 如果脚本被直接执行
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 