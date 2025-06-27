#!/bin/bash

# BossFi 项目部署脚本
# 使用方法: sudo ./deploy-project.sh [init|update|restart|stop]

set -e

# 配置变量
PROJECT_NAME="bossfi-backend"
PROJECT_DIR="/opt/bossfi/${PROJECT_NAME}"
DEPLOY_DIR="${PROJECT_DIR}/deploy"
BACKUP_DIR="/opt/bossfi/backups"
LOG_DIR="/opt/bossfi/logs"
DATA_DIR="/opt/bossfi/data"

# Git 仓库配置
GIT_REPO="https://github.com/locey/bossfi-backend.git"
GIT_BRANCH="dev"

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

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# 检查是否为root用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "此脚本需要root权限运行，请使用 sudo ./deploy-project.sh"
    fi
}

# 检查基础环境
check_environment() {
    log "检查基础环境..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        error "Docker 未安装，请先运行 ./setup-environment.sh"
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose 未安装，请先运行 ./setup-environment.sh"
    fi
    
    # 检查项目目录
    if [ ! -d "/opt/bossfi" ]; then
        error "项目目录不存在，请先运行 ./setup-environment.sh"
    fi
    
    log "基础环境检查完成"
}

# 克隆或更新项目代码
clone_or_update_project() {
    log "获取项目代码..."
    
    if [ -d "$PROJECT_DIR" ]; then
        log "项目目录已存在，更新代码..."
        cd "$PROJECT_DIR"
        
        # 备份本地修改
        if [ -f ".env" ]; then
            cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
            log "已备份环境配置文件"
        fi
        
        # 更新代码
        git fetch origin
        git reset --hard origin/$GIT_BRANCH
        git pull origin $GIT_BRANCH
        
        # 恢复环境配置
        if [ -f ".env.backup.$(date +%Y%m%d_%H%M%S)" ]; then
            cp .env.backup.$(date +%Y%m%d_%H%M%S) .env
            log "已恢复环境配置文件"
        fi
    else
        log "克隆项目代码..."
        cd /opt/bossfi
        git clone -b $GIT_BRANCH $GIT_REPO $PROJECT_NAME
        cd "$PROJECT_DIR"
    fi
    
    log "项目代码获取完成"
}

# 配置环境变量
setup_environment_variables() {
    log "配置环境变量..."
    
    cd "$PROJECT_DIR"
    
    # 检查是否存在 .env 文件
    if [ ! -f ".env" ]; then
        if [ -f "env.example" ]; then
            cp env.example .env
            log "已从 env.example 创建 .env 文件"
        else
            error ".env 文件和 env.example 文件都不存在"
        fi
    fi
    
    # 设置文件权限
    chmod 600 .env
    
    # 检查关键环境变量
    source .env
    
    if [ -z "$DB_PASSWORD" ] || [ "$DB_PASSWORD" = "postgres" ]; then
        warn "请修改数据库密码: DB_PASSWORD"
    fi
    
    if [ -z "$JWT_SECRET" ] || [ "$JWT_SECRET" = "your-super-secret-key-here-change-in-production" ]; then
        warn "请修改JWT密钥: JWT_SECRET"
    fi
    
    # 自动修改为Docker环境配置
    sed -i 's/DB_HOST=localhost/DB_HOST=postgres/g' .env
    sed -i 's/REDIS_HOST=localhost/REDIS_HOST=redis/g' .env
    sed -i 's/GIN_MODE=debug/GIN_MODE=release/g' .env
    
    log "环境变量配置完成"
}

# 构建Docker镜像
build_docker_images() {
    log "构建Docker镜像..."
    
    cd "$DEPLOY_DIR"
    
    # 检查Docker Compose文件
    if [ ! -f "docker-compose.yml" ]; then
        error "Docker Compose 文件不存在: $DEPLOY_DIR/docker-compose.yml"
    fi
    
    # 加载环境变量
    if [ -f "$PROJECT_DIR/.env" ]; then
        export $(cat "$PROJECT_DIR/.env" | grep -v '^#' | xargs)
    fi
    
    # 构建镜像
    docker-compose build --no-cache bossfi-backend
    
    log "Docker镜像构建完成"
}

# 备份数据
backup_data() {
    log "备份数据..."
    
    # 创建备份目录
    BACKUP_FILE="$BACKUP_DIR/backup_$(date +%Y%m%d_%H%M%S).tar.gz"
    
    # 备份数据库
    if docker ps | grep -q "bossfi-postgres"; then
        log "备份数据库..."
        docker exec bossfi-postgres pg_dump -U bossfi_user bossfi > "$BACKUP_DIR/db_backup_$(date +%Y%m%d_%H%M%S).sql"
    fi
    
    # 备份Redis数据
    if docker ps | grep -q "bossfi-redis"; then
        log "备份Redis数据..."
        docker exec bossfi-redis redis-cli SAVE
        docker cp bossfi-redis:/data/dump.rdb "$BACKUP_DIR/redis_backup_$(date +%Y%m%d_%H%M%S).rdb"
    fi
    
    # 备份配置文件
    if [ -d "$PROJECT_DIR" ]; then
        tar -czf "$BACKUP_FILE" -C "$PROJECT_DIR" .env deploy/ 2>/dev/null || true
        log "配置文件备份完成: $BACKUP_FILE"
    fi
    
    # 清理旧备份（保留最近7天）
    find "$BACKUP_DIR" -name "backup_*.tar.gz" -mtime +7 -delete 2>/dev/null || true
    find "$BACKUP_DIR" -name "db_backup_*.sql" -mtime +7 -delete 2>/dev/null || true
    find "$BACKUP_DIR" -name "redis_backup_*.rdb" -mtime +7 -delete 2>/dev/null || true
    
    log "数据备份完成"
}

# 停止服务
stop_services() {
    log "停止现有服务..."
    
    cd "$DEPLOY_DIR"
    
    if [ -f "docker-compose.yml" ]; then
        docker-compose down || true
    fi
    
    # 清理未使用的容器和镜像
    docker system prune -f || true
    
    log "服务停止完成"
}

# 启动服务
start_services() {
    log "启动服务..."
    
    cd "$DEPLOY_DIR"
    
    # 加载环境变量
    if [ -f "$PROJECT_DIR/.env" ]; then
        export $(cat "$PROJECT_DIR/.env" | grep -v '^#' | xargs)
    fi
    
    # 启动服务
    docker-compose up -d
    
    log "等待服务启动..."
    sleep 60
    
    # 健康检查
    check_health
    
    log "服务启动完成"
}

# 健康检查
check_health() {
    log "执行健康检查..."
    
    # 检查容器状态
    local failed=0
    
    if ! docker ps | grep -q "bossfi-postgres"; then
        error "PostgreSQL 服务未启动"
        failed=1
    fi
    
    if ! docker ps | grep -q "bossfi-redis"; then
        error "Redis 服务未启动"
        failed=1
    fi
    
    if ! docker ps | grep -q "bossfi-backend"; then
        error "后端服务未启动"
        failed=1
    fi
    
    if ! docker ps | grep -q "bossfi-nginx"; then
        error "Nginx 服务未启动"
        failed=1
    fi
    
    if [ $failed -eq 1 ]; then
        error "健康检查失败"
    fi
    
    # 检查健康检查端点
    for i in {1..12}; do
        if curl -f http://localhost/health >/dev/null 2>&1; then
            log "健康检查通过"
            return 0
        fi
        log "等待服务响应... ($i/12)"
        sleep 10
    done
    
    warn "健康检查超时，请检查服务状态"
}

# 显示状态
show_status() {
    log "服务状态:"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep bossfi || echo "没有运行的BossFi服务"
    
    echo ""
    log "资源使用情况:"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep bossfi || echo "没有运行的BossFi服务"
    
    echo ""
    log "磁盘使用情况:"
    df -h /opt/bossfi
    
    echo ""
    log "最近的日志:"
    if [ -f "$LOG_DIR/access.log" ]; then
        tail -5 "$LOG_DIR/access.log" 2>/dev/null || echo "访问日志不存在"
    fi
}

# 查看日志
show_logs() {
    local service=${1:-all}
    
    cd "$DEPLOY_DIR"
    
    case $service in
        "all")
            docker-compose logs --tail=50 -f
            ;;
        "backend")
            docker-compose logs --tail=50 -f bossfi-backend
            ;;
        "postgres")
            docker-compose logs --tail=50 -f postgres
            ;;
        "redis")
            docker-compose logs --tail=50 -f redis
            ;;
        "nginx")
            docker-compose logs --tail=50 -f nginx
            ;;
        *)
            error "未知的服务: $service"
            ;;
    esac
}

# 初始化部署
init_deploy() {
    log "开始初始化部署..."
    
    check_root
    check_environment
    clone_or_update_project
    setup_environment_variables
    build_docker_images
    start_services
    show_status
    
    log "✅ 初始化部署完成！"
}

# 更新项目
update_project() {
    log "开始更新项目..."
    
    check_root
    check_environment
    backup_data
    clone_or_update_project
    setup_environment_variables
    build_docker_images
    stop_services
    start_services
    show_status
    
    log "✅ 项目更新完成！"
}

# 重启服务
restart_services() {
    log "重启服务..."
    
    check_root
    stop_services
    start_services
    show_status
    
    log "✅ 服务重启完成！"
}

# 显示帮助信息
show_help() {
    echo "BossFi 项目部署脚本"
    echo ""
    echo "使用方法:"
    echo "  sudo ./deploy-project.sh <command> [options]"
    echo ""
    echo "命令:"
    echo "  init     - 初始化部署项目"
    echo "  update   - 更新项目代码并重新部署"
    echo "  restart  - 重启所有服务"
    echo "  stop     - 停止所有服务"
    echo "  status   - 显示服务状态"
    echo "  logs     - 查看服务日志"
    echo "  backup   - 手动备份数据"
    echo "  help     - 显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  sudo ./deploy-project.sh init"
    echo "  sudo ./deploy-project.sh update"
    echo "  sudo ./deploy-project.sh logs backend"
    echo ""
}

# 主函数
main() {
    local command=${1:-help}
    
    case $command in
        "init")
            init_deploy
            ;;
        "update")
            update_project
            ;;
        "restart")
            restart_services
            ;;
        "stop")
            check_root
            stop_services
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs $2
            ;;
        "backup")
            check_root
            backup_data
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            error "未知命令: $command"
            show_help
            ;;
    esac
}

# 如果脚本被直接执行
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 