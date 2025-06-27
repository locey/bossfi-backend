#!/bin/bash

# BossFi Backend 更新脚本
# 使用方法: ./update.sh [branch]

set -e

# 获取脚本所在目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BRANCH=${1:-dev}
BACKUP_DIR="/opt/bossfi/backups"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# 检查Git仓库状态
check_git_status() {
    log "检查Git仓库状态..."
    
    cd $PROJECT_ROOT
    
    if [ ! -d ".git" ]; then
        error "当前目录不是Git仓库"
    fi
    
    # 检查是否有未提交的更改
    if ! git diff --quiet; then
        warn "检测到未提交的更改"
        git status --porcelain
        
        read -p "是否要备份并丢弃本地更改? (y/N): " confirm
        if [[ $confirm =~ ^[Yy]$ ]]; then
            # 备份本地更改
            local timestamp=$(date +%Y%m%d_%H%M%S)
            git stash push -m "Auto backup before update $timestamp"
            log "本地更改已暂存: stash@{0}"
        else
            error "请先处理本地更改或使用 git stash"
        fi
    fi
}

# 备份当前版本
backup_current_version() {
    log "备份当前版本..."
    
    mkdir -p $BACKUP_DIR
    
    local current_commit=$(git rev-parse HEAD)
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/version_backup_${timestamp}_${current_commit:0:8}.tar.gz"
    
    # 备份整个项目（排除.git目录）
    tar -czf $backup_file -C $PROJECT_ROOT --exclude='.git' .
    
    log "版本备份完成: $backup_file"
    
    # 备份数据库（如果服务正在运行）
    if docker ps | grep -q "bossfi-postgres"; then
        log "备份数据库..."
        docker exec bossfi-postgres pg_dump -U bossfi_user bossfi > $BACKUP_DIR/db_backup_${timestamp}.sql
        log "数据库备份完成"
    fi
}

# 拉取最新代码
pull_latest_code() {
    log "拉取最新代码 (分支: $BRANCH)..."
    
    cd $PROJECT_ROOT
    
    # 获取远程更新
    git fetch origin
    
    # 检查远程分支是否存在
    if ! git ls-remote --heads origin | grep -q "refs/heads/$BRANCH"; then
        error "远程分支 '$BRANCH' 不存在"
    fi
    
    # 获取当前和远程的commit hash
    local current_commit=$(git rev-parse HEAD)
    local remote_commit=$(git rev-parse origin/$BRANCH)
    
    if [ "$current_commit" = "$remote_commit" ]; then
        log "代码已是最新版本，无需更新"
        return 0
    fi
    
    log "发现新版本："
    log "  当前版本: ${current_commit:0:8}"
    log "  最新版本: ${remote_commit:0:8}"
    
    # 显示更新日志
    echo -e "\n${BLUE}更新内容:${NC}"
    git log --oneline --graph $current_commit..$remote_commit | head -10
    
    # 切换到目标分支并拉取
    git checkout $BRANCH
    git pull origin $BRANCH
    
    log "代码更新完成"
}

# 重新部署服务
redeploy_services() {
    log "重新部署服务..."
    
    cd $SCRIPT_DIR
    
    # 停止现有服务
    if [ -f "docker-compose.prod.yml" ]; then
        log "停止现有服务..."
        docker-compose -f docker-compose.prod.yml down
    fi
    
    # 重新构建并启动
    log "重新构建并启动服务..."
    docker-compose -f docker-compose.prod.yml build --no-cache
    docker-compose -f docker-compose.prod.yml up -d
    
    # 等待服务启动
    log "等待服务启动..."
    sleep 30
    
    # 健康检查
    check_health
}

# 健康检查
check_health() {
    log "执行健康检查..."
    
    local max_attempts=10
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s http://localhost/health >/dev/null 2>&1; then
            log "✅ 健康检查通过"
            return 0
        fi
        
        log "等待服务响应... ($attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    error "❌ 健康检查失败，服务可能未正常启动"
}

# 清理旧备份
cleanup_old_backups() {
    log "清理旧备份文件..."
    
    # 保留最近7天的备份
    find $BACKUP_DIR -name "version_backup_*.tar.gz" -mtime +7 -delete 2>/dev/null || true
    find $BACKUP_DIR -name "db_backup_*.sql" -mtime +7 -delete 2>/dev/null || true
    
    log "旧备份清理完成"
}

# 显示使用帮助
show_help() {
    echo "BossFi Backend 更新脚本"
    echo ""
    echo "使用方法:"
    echo "  sudo ./update.sh [分支名]"
    echo ""
    echo "参数:"
    echo "  分支名     要更新的Git分支 (默认: dev)"
    echo ""
    echo "示例:"
    echo "  sudo ./update.sh          # 更新dev分支"
    echo "  sudo ./update.sh main     # 更新main分支"
    echo "  sudo ./update.sh feature  # 更新feature分支"
    echo ""
    echo "注意:"
    echo "  - 此脚本需要root权限运行"
    echo "  - 更新前会自动备份当前版本和数据库"
    echo "  - 如有本地修改，会提示是否暂存"
}

# 主函数
main() {
    # 检查帮助参数
    if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
        show_help
        exit 0
    fi
    
    log "🚀 开始更新 BossFi Backend (分支: $BRANCH)"
    log "项目根目录: $PROJECT_ROOT"
    
    # 检查权限
    check_root
    
    # 检查Git状态
    check_git_status
    
    # 备份当前版本
    backup_current_version
    
    # 拉取最新代码
    pull_latest_code
    
    # 重新部署服务
    redeploy_services
    
    # 清理旧备份
    cleanup_old_backups
    
    log "✅ 更新完成！"
    log "🌐 服务地址: http://your-server-ip"
    log "📊 服务状态: sudo ./monitor.sh status"
    log "📋 查看日志: sudo ./monitor.sh logs"
    
    # 显示版本信息
    cd $PROJECT_ROOT
    local current_commit=$(git rev-parse HEAD)
    log "📝 当前版本: ${current_commit:0:8} ($(git log -1 --format=%s))"
}

# 捕获错误
trap 'error "更新过程中发生错误，请检查日志"' ERR

# 执行主函数
main "$@" 