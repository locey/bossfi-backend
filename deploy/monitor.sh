#!/bin/bash

# BossFi 服务监控脚本
# 使用方法: ./monitor.sh [status|logs|restart|health]

set -e

# 配置变量
PROJECT_DIR="/opt/bossfi/bossfi-backend"
DEPLOY_DIR="${PROJECT_DIR}/deploy"
LOG_DIR="/opt/bossfi/logs"

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
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# 检查服务状态
check_service_status() {
    log "检查服务状态..."
    
    local services=("bossfi-postgres" "bossfi-redis" "bossfi-backend" "bossfi-nginx")
    local all_running=true
    
    echo ""
    printf "%-20s %-15s %-10s %-30s\n" "服务名称" "状态" "健康状态" "端口"
    echo "--------------------------------------------------------------------"
    
    for service in "${services[@]}"; do
        if docker ps --format "table {{.Names}}" | grep -q "$service"; then
            local status="运行中"
            local health=$(docker inspect --format='{{.State.Health.Status}}' "$service" 2>/dev/null || echo "unknown")
            local ports=$(docker port "$service" 2>/dev/null | tr '\n' ' ' || echo "N/A")
            
            if [ "$health" = "healthy" ] || [ "$health" = "unknown" ]; then
                printf "%-20s ${GREEN}%-15s${NC} %-10s %-30s\n" "$service" "$status" "$health" "$ports"
            else
                printf "%-20s ${GREEN}%-15s${NC} ${RED}%-10s${NC} %-30s\n" "$service" "$status" "$health" "$ports"
                all_running=false
            fi
        else
            printf "%-20s ${RED}%-15s${NC} %-10s %-30s\n" "$service" "未运行" "N/A" "N/A"
            all_running=false
        fi
    done
    
    echo ""
    if [ "$all_running" = true ]; then
        log "✅ 所有服务运行正常"
    else
        warn "⚠️  部分服务存在问题"
    fi
}

# 检查资源使用情况
check_resource_usage() {
    log "检查资源使用情况..."
    
    echo ""
    echo "系统资源使用情况:"
    echo "--------------------------------------------------------------------"
    
    # 内存使用
    local memory_info=$(free -h)
    echo "内存使用情况:"
    echo "$memory_info"
    
    echo ""
    
    # 磁盘使用
    echo "磁盘使用情况:"
    df -h /opt/bossfi
    
    echo ""
    
    # Docker容器资源使用
    if docker ps | grep -q "bossfi-"; then
        echo "容器资源使用情况:"
        docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}" | grep bossfi
    else
        echo "没有运行的BossFi容器"
    fi
}

# 检查网络连接
check_network_connectivity() {
    log "检查网络连接..."
    
    echo ""
    echo "网络连接检查:"
    echo "--------------------------------------------------------------------"
    
    # 检查本地端口
    local ports=("80:HTTP" "443:HTTPS" "8080:Backend" "5432:PostgreSQL" "6379:Redis")
    
    for port_info in "${ports[@]}"; do
        local port=$(echo "$port_info" | cut -d':' -f1)
        local service=$(echo "$port_info" | cut -d':' -f2)
        
        if netstat -tln | grep -q ":$port "; then
            printf "%-15s ${GREEN}%-10s${NC} %s\n" "$service" "监听中" "端口 $port"
        else
            printf "%-15s ${RED}%-10s${NC} %s\n" "$service" "未监听" "端口 $port"
        fi
    done
    
    echo ""
    
    # 检查API健康状态
    if curl -f http://localhost/health >/dev/null 2>&1; then
        log "✅ API健康检查通过"
    else
        warn "⚠️  API健康检查失败"
    fi
}

# 检查日志文件
check_logs() {
    log "检查日志文件..."
    
    echo ""
    echo "日志文件状态:"
    echo "--------------------------------------------------------------------"
    
    local log_files=("$LOG_DIR/access.log" "$LOG_DIR/error.log")
    
    for log_file in "${log_files[@]}"; do
        if [ -f "$log_file" ]; then
            local size=$(du -h "$log_file" | cut -f1)
            local lines=$(wc -l < "$log_file")
            printf "%-30s ${GREEN}存在${NC} %s (%s 行)\n" "$(basename "$log_file")" "$size" "$lines"
        else
            printf "%-30s ${RED}不存在${NC}\n" "$(basename "$log_file")"
        fi
    done
    
    echo ""
    
    # 显示最近的错误日志
    if [ -f "$LOG_DIR/error.log" ]; then
        local error_count=$(grep -c "ERROR" "$LOG_DIR/error.log" 2>/dev/null || echo "0")
        if [ "$error_count" -gt 0 ]; then
            warn "发现 $error_count 个错误，最近的错误:"
            tail -5 "$LOG_DIR/error.log" | grep "ERROR" || echo "没有最近的错误"
        else
            log "✅ 没有发现错误日志"
        fi
    fi
}

# 完整的健康检查
full_health_check() {
    log "执行完整健康检查..."
    
    check_service_status
    check_resource_usage
    check_network_connectivity
    check_logs
    
    echo ""
    log "健康检查完成"
}

# 显示实时日志
show_live_logs() {
    local service=${1:-all}
    
    if [ ! -d "$DEPLOY_DIR" ]; then
        error "部署目录不存在: $DEPLOY_DIR"
    fi
    
    cd "$DEPLOY_DIR"
    
    case $service in
        "all")
            log "显示所有服务的实时日志..."
            docker-compose logs --tail=100 -f
            ;;
        "backend")
            log "显示后端服务的实时日志..."
            docker-compose logs --tail=100 -f bossfi-backend
            ;;
        "postgres")
            log "显示数据库服务的实时日志..."
            docker-compose logs --tail=100 -f postgres
            ;;
        "redis")
            log "显示Redis服务的实时日志..."
            docker-compose logs --tail=100 -f redis
            ;;
        "nginx")
            log "显示Nginx服务的实时日志..."
            docker-compose logs --tail=100 -f nginx
            ;;
        *)
            error "未知的服务: $service"
            ;;
    esac
}

# 重启服务
restart_service() {
    local service=${1:-all}
    
    if [ ! -d "$DEPLOY_DIR" ]; then
        error "部署目录不存在: $DEPLOY_DIR"
    fi
    
    cd "$DEPLOY_DIR"
    
    case $service in
        "all")
            log "重启所有服务..."
            docker-compose restart
            ;;
        "backend")
            log "重启后端服务..."
            docker-compose restart bossfi-backend
            ;;
        "postgres")
            log "重启数据库服务..."
            docker-compose restart postgres
            ;;
        "redis")
            log "重启Redis服务..."
            docker-compose restart redis
            ;;
        "nginx")
            log "重启Nginx服务..."
            docker-compose restart nginx
            ;;
        *)
            error "未知的服务: $service"
            ;;
    esac
    
    log "等待服务重启..."
    sleep 10
    
    check_service_status
}

# 显示性能统计
show_performance_stats() {
    log "显示性能统计..."
    
    echo ""
    echo "系统负载:"
    uptime
    
    echo ""
    echo "CPU使用情况:"
    top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print "CPU使用率: " 100 - $1 "%"}'
    
    echo ""
    echo "内存使用情况:"
    free -m | awk 'NR==2{printf "内存使用率: %.1f%% (%d/%d MB)\n", $3*100/$2, $3, $2}'
    
    echo ""
    echo "磁盘I/O:"
    iostat -x 1 1 | tail -n +4
    
    echo ""
    echo "网络连接数:"
    netstat -an | grep :80 | wc -l | awk '{print "HTTP连接数: " $1}'
    netstat -an | grep :443 | wc -l | awk '{print "HTTPS连接数: " $1}'
}

# 显示帮助信息
show_help() {
    echo "BossFi 服务监控脚本"
    echo ""
    echo "使用方法:"
    echo "  ./monitor.sh <command> [options]"
    echo ""
    echo "命令:"
    echo "  status     - 显示服务状态"
    echo "  health     - 执行完整健康检查"
    echo "  logs       - 显示实时日志 [all|backend|postgres|redis|nginx]"
    echo "  restart    - 重启服务 [all|backend|postgres|redis|nginx]"
    echo "  resource   - 显示资源使用情况"
    echo "  network    - 检查网络连接"
    echo "  perf       - 显示性能统计"
    echo "  help       - 显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  ./monitor.sh status"
    echo "  ./monitor.sh logs backend"
    echo "  ./monitor.sh restart nginx"
    echo ""
}

# 主函数
main() {
    local command=${1:-status}
    
    case $command in
        "status")
            check_service_status
            ;;
        "health")
            full_health_check
            ;;
        "logs")
            show_live_logs $2
            ;;
        "restart")
            restart_service $2
            ;;
        "resource")
            check_resource_usage
            ;;
        "network")
            check_network_connectivity
            ;;
        "perf")
            show_performance_stats
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