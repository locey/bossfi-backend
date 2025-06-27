#!/bin/bash

# BossFi 基础设施部署脚本
# 用于部署 nginx, postgresql, redis

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 检查必要的命令
check_prerequisites() {
    log_step "检查系统环境和依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    # 检查Docker是否运行
    if ! docker info &> /dev/null; then
        log_error "Docker 服务未运行，请启动 Docker 服务"
        exit 1
    fi
    
    log_info "✓ 系统环境检查通过"
}

# 创建必要的目录
create_directories() {
    log_step "创建必要的目录结构..."
    
    mkdir -p nginx/conf.d
    mkdir -p nginx/ssl
    mkdir -p logs/nginx
    mkdir -p logs/postgres
    mkdir -p logs/redis
    mkdir -p backups/postgres
    mkdir -p backups/redis
    
    log_info "✓ 目录结构创建完成"
}

# 设置权限
set_permissions() {
    log_step "设置文件权限..."
    
    # 设置nginx配置文件权限
    if [ -f "nginx/nginx.conf" ]; then
        chmod 644 nginx/nginx.conf
    fi
    
    if [ -f "nginx/conf.d/bossfi.conf" ]; then
        chmod 644 nginx/conf.d/bossfi.conf
    fi
    
    # 设置SSL目录权限（如果存在）
    if [ -d "nginx/ssl" ]; then
        chmod 700 nginx/ssl
    fi
    
    log_info "✓ 文件权限设置完成"
}

# 检查端口占用
check_ports() {
    log_step "检查端口占用情况..."
    
    ports=(80 443 5432 6379)
    for port in "${ports[@]}"; do
        if netstat -tuln 2>/dev/null | grep -q ":${port} "; then
            log_warn "端口 ${port} 已被占用，可能需要停止相关服务"
        else
            log_info "✓ 端口 ${port} 可用"
        fi
    done
}

# 停止已存在的服务
stop_existing_services() {
    log_step "停止已存在的基础设施服务..."
    
    if docker-compose -f docker-compose.infrastructure.yml ps -q 2>/dev/null | grep -q .; then
        log_info "发现运行中的基础设施服务，正在停止..."
        docker-compose -f docker-compose.infrastructure.yml down
        log_info "✓ 已停止现有服务"
    else
        log_info "✓ 没有发现运行中的基础设施服务"
    fi
}

# 拉取Docker镜像
pull_images() {
    log_step "拉取 Docker 镜像..."
    
    docker-compose -f docker-compose.infrastructure.yml pull
    
    log_info "✓ Docker 镜像拉取完成"
}

# 启动基础设施服务
start_infrastructure() {
    log_step "启动基础设施服务..."
    
    # 创建网络（如果不存在）
    if ! docker network ls | grep -q "bossfi-network"; then
        log_info "创建 Docker 网络..."
        docker network create bossfi-network --driver bridge --subnet=172.20.0.0/16
    fi
    
    # 启动服务
    docker-compose -f docker-compose.infrastructure.yml up -d
    
    log_info "✓ 基础设施服务启动完成"
}

# 等待服务就绪
wait_for_services() {
    log_step "等待服务启动完成..."
    
    # 等待PostgreSQL
    log_info "等待 PostgreSQL 启动..."
    max_attempts=30
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if docker-compose -f docker-compose.infrastructure.yml exec -T postgres pg_isready -U postgres -d bossfi &>/dev/null; then
            log_info "✓ PostgreSQL 已就绪"
            break
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    
    if [ $attempt -eq $max_attempts ]; then
        log_error "PostgreSQL 启动超时"
        exit 1
    fi
    
    # 等待Redis
    log_info "等待 Redis 启动..."
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if docker-compose -f docker-compose.infrastructure.yml exec -T redis redis-cli ping &>/dev/null; then
            log_info "✓ Redis 已就绪"
            break
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    
    if [ $attempt -eq $max_attempts ]; then
        log_error "Redis 启动超时"
        exit 1
    fi
    
    # 等待Nginx
    log_info "等待 Nginx 启动..."
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if curl -s http://localhost/health &>/dev/null; then
            log_info "✓ Nginx 已就绪"
            break
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    
    if [ $attempt -eq $max_attempts ]; then
        log_warn "Nginx 健康检查超时，但服务可能正常运行"
    fi
}

# 显示服务状态
show_status() {
    log_step "显示服务状态..."
    
    echo ""
    echo "=== 基础设施服务状态 ==="
    docker-compose -f docker-compose.infrastructure.yml ps
    
    echo ""
    echo "=== 服务访问信息 ==="
    log_info "PostgreSQL: localhost:5432 (用户: postgres, 密码: postgres123, 数据库: bossfi)"
    log_info "Redis: localhost:6379 (密码: redis123)"
    log_info "Nginx: http://localhost (健康检查: http://localhost/health)"
    
    echo ""
    echo "=== 常用管理命令 ==="
    log_info "查看日志: docker-compose -f docker-compose.infrastructure.yml logs -f [service]"
    log_info "重启服务: docker-compose -f docker-compose.infrastructure.yml restart [service]"
    log_info "停止服务: docker-compose -f docker-compose.infrastructure.yml down"
    log_info "进入容器: docker-compose -f docker-compose.infrastructure.yml exec [service] /bin/sh"
}

# 主函数
main() {
    echo ""
    log_info "========================================"
    log_info "       BossFi 基础设施部署脚本"
    log_info "========================================"
    echo ""
    
    check_prerequisites
    create_directories
    set_permissions
    check_ports
    stop_existing_services
    pull_images
    start_infrastructure
    wait_for_services
    show_status
    
    echo ""
    log_info "========================================"
    log_info "      基础设施部署完成！"
    log_info "========================================"
    echo ""
}

# 信号处理
trap 'log_error "脚本被中断"; exit 1' INT TERM

# 运行主函数
main "$@" 