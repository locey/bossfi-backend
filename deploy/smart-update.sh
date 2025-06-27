#!/bin/bash

# BossFi 智能更新脚本
# 自动备份环境配置，更新代码，恢复配置

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BRANCH=${1:-dev}

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%H:%M:%S')] $1${NC}"
    exit 1
}

main() {
    log "🚀 开始智能更新 BossFi Backend (分支: $BRANCH)"
    
    cd $PROJECT_ROOT
    
    # 1. 备份环境配置
    log "📦 备份环境配置..."
    ENV_BACKUP_DIR="/tmp/bossfi_env_backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p $ENV_BACKUP_DIR
    
    # 备份存在的环境文件
    ENV_FILES_FOUND=false
    for file in .env; do
        if [ -f "$file" ]; then
            cp "$file" "$ENV_BACKUP_DIR/"
            log "✅ 已备份: $file"
            ENV_FILES_FOUND=true
        fi
    done
    
    if [ "$ENV_FILES_FOUND" = false ]; then
        warn "⚠️ 未找到环境配置文件，将跳过备份"
    fi
    
    # 2. 暂存本地修改
    if ! git diff --quiet; then
        log "💾 暂存本地修改..."
        git stash push -m "自动备份 $(date +%Y%m%d_%H%M%S)"
    fi
    
    # 3. 更新代码
    log "⬇️ 拉取最新代码..."
    git fetch origin $BRANCH
    git checkout $BRANCH
    git reset --hard origin/$BRANCH
    
    # 4. 恢复环境配置
    log "🔄 恢复环境配置..."
    ENV_FILES_RESTORED=false
    for file in .env; do
        backup_file="$ENV_BACKUP_DIR/$(basename $file)"
        if [ -f "$backup_file" ]; then
            cp "$backup_file" "$file"
            chmod 600 "$file"
            log "✅ 已恢复: $file"
            ENV_FILES_RESTORED=true
        fi
    done
    
    if [ "$ENV_FILES_RESTORED" = false ]; then
        warn "⚠️ 未找到备份的环境配置，请手动配置"
        if [ -f "env.example" ]; then
            log "💡 提示: 可以复制 env.example 为 .env 并修改配置"
        fi
    fi
    
    # 5. 设置权限
    log "🔒 设置脚本权限..."
    chmod +x deploy/*.sh
    
    # 6. 清理备份
    rm -rf $ENV_BACKUP_DIR
    
    # 7. 显示更新信息
    local current_commit=$(git rev-parse HEAD)
    log "✅ 更新完成！"
    log "📝 当前版本: ${current_commit:0:8}"
    log "🔧 环境配置已保留"
    
    echo ""
    echo "接下来可以执行："
    echo "  sudo ./deploy/deploy.sh prod    # 重新部署"
    echo "  ./deploy/monitor.sh status      # 检查状态"
}

# 检查是否在正确目录
if [ ! -f "go.mod" ]; then
    error "请在项目根目录运行此脚本"
fi

# 检查是否为root
if [[ $EUID -ne 0 ]]; then
    error "此脚本需要root权限运行"
fi

main "$@" 