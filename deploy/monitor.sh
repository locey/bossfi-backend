#!/bin/bash

# BossFi Backend 监控脚本
# 使用方法: ./monitor.sh [status|logs|restart|backup]

set -e

# 配置变量
DEPLOY_DIR="/opt/bossfi"
COMPOSE_FILE="$DEPLOY_DIR/docker-compose.prod.yml"
LOG_DIR="$DEPLOY_DIR/logs"
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
}

# 检查服务状态
check_status() {
    log "=== BossFi 服务状态 ==="
    
    if ! command -v docker &> /dev/null; then
        error "Docker 未安装"
        return 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose 未安装"
        return 1
    fi
    
    # 检查容器状态
    echo -e "\n${BLUE}Docker 容器状态:${NC}"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep bossfi || echo "没有运行的BossFi容器"
    
    # 检查服务健康状态
    echo -e "\n${BLUE}服务健康检查:${NC}"
    
    # 后端健康检查
    if curl -f -s http://localhost/health >/dev/null 2>&1; then
        echo -e "✅ 后端服务: ${GREEN}正常${NC}"
    else
        echo -e "❌ 后端服务: ${RED}异常${NC}"
    fi
    
    # 数据库连接检查
    if docker exec bossfi-postgres pg_isready -U bossfi_user -d bossfi >/dev/null 2>&1; then
        echo -e "✅ 数据库: ${GREEN}正常${NC}"
    else
        echo -e "❌ 数据库: ${RED}异常${NC}"
    fi
    
    # Redis连接检查
    if docker exec bossfi-redis redis-cli ping | grep -q "PONG"; then
        echo -e "✅ Redis: ${GREEN}正常${NC}"
    else
        echo -e "❌ Redis: ${RED}异常${NC}"
    fi
    
    # Nginx状态检查
    if docker exec bossfi-nginx nginx -t >/dev/null 2>&1; then
        echo -e "✅ Nginx: ${GREEN}正常${NC}"
    else
        echo -e "❌ Nginx: ${RED}配置异常${NC}"
    fi
}

# 检查资源使用
check_resources() {
    log "=== 系统资源使用情况 ==="
    
    # 系统资源
    echo -e "\n${BLUE}系统资源:${NC}"
    
    # 内存使用
    MEMORY_INFO=$(free -h | awk 'NR==2{printf "使用: %s / %s (%.1f%%)", $3, $2, $3*100/$2}')
    echo "🧠 内存: $MEMORY_INFO"
    
    # CPU负载
    LOAD_AVERAGE=$(uptime | awk -F'load average:' '{print $2}')
    echo "🔄 CPU负载:$LOAD_AVERAGE"
    
    # 磁盘使用
    DISK_INFO=$(df -h / | awk 'NR==2{printf "使用: %s / %s (%s)", $3, $2, $5}')
    echo "💾 磁盘: $DISK_INFO"
    
    # Docker资源使用
    echo -e "\n${BLUE}Docker容器资源使用:${NC}"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" | grep bossfi || echo "没有运行的BossFi容器"
    
    # 检查磁盘空间警告
    DISK_USAGE=$(df / | awk 'NR==2{print $5}' | cut -d'%' -f1)
    if [ $DISK_USAGE -gt 80 ]; then
        warn "磁盘使用率过高: ${DISK_USAGE}%"
    fi
    
    # 检查内存使用警告
    MEMORY_USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    if [ $MEMORY_USAGE -gt 80 ]; then
        warn "内存使用率过高: ${MEMORY_USAGE}%"
    fi
}

# 查看日志
show_logs() {
    log "=== 服务日志 ==="
    
    local service=${1:-""}
    
    if [ ! -f "$COMPOSE_FILE" ]; then
        error "Docker Compose 文件不存在: $COMPOSE_FILE"
        return 1
    fi
    
    cd $DEPLOY_DIR
    
    if [ -z "$service" ]; then
        echo "选择要查看的服务日志:"
        echo "1) 所有服务"
        echo "2) 后端服务"
        echo "3) 数据库"
        echo "4) Redis"
        echo "5) Nginx"
        read -p "请选择 (1-5): " choice
        
        case $choice in
            1) docker-compose -f docker-compose.prod.yml logs -f --tail=100 ;;
            2) docker-compose -f docker-compose.prod.yml logs -f --tail=100 bossfi-backend ;;
            3) docker-compose -f docker-compose.prod.yml logs -f --tail=100 postgres ;;
            4) docker-compose -f docker-compose.prod.yml logs -f --tail=100 redis ;;
            5) docker-compose -f docker-compose.prod.yml logs -f --tail=100 nginx ;;
            *) error "无效选择" ;;
        esac
    else
        docker-compose -f docker-compose.prod.yml logs -f --tail=100 $service
    fi
}

# 重启服务
restart_services() {
    log "=== 重启服务 ==="
    
    local service=${1:-""}
    
    if [ ! -f "$COMPOSE_FILE" ]; then
        error "Docker Compose 文件不存在: $COMPOSE_FILE"
        return 1
    fi
    
    cd $DEPLOY_DIR
    
    if [ -z "$service" ]; then
        log "重启所有服务..."
        docker-compose -f docker-compose.prod.yml restart
    else
        log "重启服务: $service"
        docker-compose -f docker-compose.prod.yml restart $service
    fi
    
    log "等待服务启动..."
    sleep 10
    
    check_status
}

# 备份数据
backup_data() {
    log "=== 数据备份 ==="
    
    mkdir -p $BACKUP_DIR
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    
    # 备份数据库
    log "备份数据库..."
    if docker exec bossfi-postgres pg_dump -U bossfi_user bossfi > $BACKUP_DIR/db_backup_$timestamp.sql; then
        log "数据库备份完成: $BACKUP_DIR/db_backup_$timestamp.sql"
    else
        error "数据库备份失败"
    fi
    
    # 备份Redis数据
    log "备份Redis数据..."
    if docker exec bossfi-redis redis-cli SAVE >/dev/null 2>&1; then
        docker cp bossfi-redis:/data/dump.rdb $BACKUP_DIR/redis_backup_$timestamp.rdb
        log "Redis备份完成: $BACKUP_DIR/redis_backup_$timestamp.rdb"
    else
        warn "Redis备份失败"
    fi
    
    # 清理旧备份（保留最近7天）
    log "清理旧备份文件..."
    find $BACKUP_DIR -name "*.sql" -mtime +7 -delete 2>/dev/null || true
    find $BACKUP_DIR -name "*.rdb" -mtime +7 -delete 2>/dev/null || true
    
    log "备份完成"
}

# 显示帮助信息
show_help() {
    echo -e "${BLUE}BossFi Backend 监控脚本${NC}"
    echo ""
    echo "使用方法:"
    echo "  $0 status              - 检查服务状态"
    echo "  $0 resources           - 检查资源使用情况" 
    echo "  $0 logs [service]      - 查看服务日志"
    echo "  $0 restart [service]   - 重启服务"
    echo "  $0 backup              - 备份数据"
    echo "  $0 help                - 显示帮助信息"
    echo ""
    echo "服务名称: bossfi-backend, postgres, redis, nginx"
    echo ""
    echo "示例:"
    echo "  $0 status              # 检查所有服务状态"
    echo "  $0 logs bossfi-backend # 查看后端日志"
    echo "  $0 restart postgres    # 重启数据库"
}

# 主函数
main() {
    case "${1:-status}" in
        "status")
            check_status
            check_resources
            ;;
        "resources")
            check_resources
            ;;
        "logs")
            show_logs $2
            ;;
        "restart")
            restart_services $2
            ;;
        "backup")
            backup_data
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 如果脚本被直接执行
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 