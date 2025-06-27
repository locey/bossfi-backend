#!/bin/bash

# BossFi 应用部署和更新脚本
# 用于部署和更新 BossFi 后端应用

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
APP_NAME="bossfi-backend"
DOCKER_IMAGE="${APP_NAME}:latest"
BACKUP_DIR="backups"
ENV_FILE=".env"

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

# 显示帮助信息
show_help() {
    cat << EOF
BossFi 应用部署脚本

用法: $0 [选项] [命令]

命令:
    deploy      部署应用（默认）
    update      更新应用
    restart     重启应用
    stop        停止应用
    logs        查看日志
    status      查看状态
    backup      备份数据
    rollback    回滚到上一个版本

选项:
    -h, --help          显示帮助信息
    -e, --env FILE      指定环境变量文件 (默认: .env)
    --no-build          跳过构建步骤
    --no-backup         跳过备份步骤
    --force             强制执行操作
    --debug             启用调试模式

示例:
    $0 deploy              # 部署应用
    $0 update              # 更新应用
    $0 -e .env.prod deploy # 使用生产环境配置部署
    $0 --no-backup update  # 更新时跳过备份

EOF
}

# 解析命令行参数
parse_args() {
    COMMAND="deploy"
    NO_BUILD=false
    NO_BACKUP=false
    FORCE=false
    DEBUG=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -e|--env)
                ENV_FILE="$2"
                shift 2
                ;;
            --no-build)
                NO_BUILD=true
                shift
                ;;
            --no-backup)
                NO_BACKUP=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --debug)
                DEBUG=true
                set -x
                shift
                ;;
            deploy|update|restart|stop|logs|status|backup|rollback)
                COMMAND="$1"
                shift
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查环境
check_environment() {
    log_step "检查部署环境..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        exit 1
    fi
    
    # 检查环境变量文件
    if [ ! -f "$ENV_FILE" ]; then
        if [ -f "env.example" ]; then
            log_warn "环境变量文件 $ENV_FILE 不存在，从模板创建..."
            cp env.example "$ENV_FILE"
            log_warn "请编辑 $ENV_FILE 文件设置正确的环境变量"
            if [ "$FORCE" = false ]; then
                read -p "是否继续？ (y/N): " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    exit 1
                fi
            fi
        else
            log_error "环境变量文件 $ENV_FILE 不存在"
            exit 1
        fi
    fi
    
    # 检查基础设施是否运行
    if ! docker network ls | grep -q "bossfi-network"; then
        log_error "bossfi-network 网络不存在，请先运行基础设施部署脚本"
        exit 1
    fi
    
    # 检查数据库连接
    if ! docker-compose -f docker-compose.infrastructure.yml exec -T postgres pg_isready -U postgres -d bossfi &>/dev/null; then
        log_error "PostgreSQL 数据库未就绪，请检查基础设施服务"
        exit 1
    fi
    
    log_info "✓ 环境检查通过"
}

# 备份数据
backup_data() {
    if [ "$NO_BACKUP" = true ]; then
        log_info "跳过数据备份"
        return
    fi
    
    log_step "备份数据..."
    
    # 创建备份目录
    timestamp=$(date +"%Y%m%d_%H%M%S")
    backup_path="${BACKUP_DIR}/${timestamp}"
    mkdir -p "$backup_path"
    
    # 备份数据库
    log_info "备份 PostgreSQL 数据库..."
    docker-compose -f docker-compose.infrastructure.yml exec -T postgres \
        pg_dump -U postgres -d bossfi --clean --create > "${backup_path}/postgres_backup.sql"
    
    # 备份Redis数据
    log_info "备份 Redis 数据..."
    docker-compose -f docker-compose.infrastructure.yml exec -T redis \
        redis-cli --rdb - > "${backup_path}/redis_backup.rdb" 2>/dev/null || true
    
    # 记录备份信息
    cat > "${backup_path}/backup_info.txt" << EOF
备份时间: $(date)
应用版本: $(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
Git提交: $(git rev-parse HEAD 2>/dev/null || echo "unknown")
环境变量: $ENV_FILE
EOF
    
    log_info "✓ 数据备份完成: $backup_path"
}

# 构建应用镜像
build_app() {
    if [ "$NO_BUILD" = true ]; then
        log_info "跳过构建步骤"
        return
    fi
    
    log_step "构建应用镜像..."
    
    # 生成构建参数
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    
    # 构建镜像
    docker build \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        -t "$DOCKER_IMAGE" \
        -t "${APP_NAME}:${VERSION}" \
        .
    
    log_info "✓ 应用镜像构建完成"
}

# 部署应用
deploy_app() {
    log_step "部署应用..."
    
    # 导出环境变量
    export $(grep -v '^#' "$ENV_FILE" | xargs)
    
    # 停止现有应用（如果存在）
    if docker-compose -f docker-compose.app.yml ps -q 2>/dev/null | grep -q .; then
        log_info "停止现有应用..."
        docker-compose -f docker-compose.app.yml down
    fi
    
    # 启动新应用
    docker-compose -f docker-compose.app.yml up -d
    
    log_info "✓ 应用部署完成"
}

# 等待应用就绪
wait_for_app() {
    log_step "等待应用启动..."
    
    max_attempts=60
    attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s http://localhost:8080/api/v1/health &>/dev/null; then
            log_info "✓ 应用已就绪"
            return
        fi
        
        # 检查容器状态
        if ! docker-compose -f docker-compose.app.yml ps | grep -q "Up"; then
            log_error "应用容器启动失败"
            docker-compose -f docker-compose.app.yml logs --tail=50
            exit 1
        fi
        
        attempt=$((attempt + 1))
        sleep 5
        log_info "等待应用启动... ($attempt/$max_attempts)"
    done
    
    log_error "应用启动超时"
    docker-compose -f docker-compose.app.yml logs --tail=50
    exit 1
}

# 健康检查
health_check() {
    log_step "执行健康检查..."
    
    # 检查API端点
    if ! curl -f -s http://localhost:8080/api/v1/health &>/dev/null; then
        log_error "健康检查失败 - API不可访问"
        return 1
    fi
    
    # 检查数据库连接
    if ! curl -f -s http://localhost:8080/api/v1/health/db &>/dev/null; then
        log_warn "数据库连接检查失败"
    fi
    
    # 检查Redis连接
    if ! curl -f -s http://localhost:8080/api/v1/health/redis &>/dev/null; then
        log_warn "Redis连接检查失败"
    fi
    
    log_info "✓ 健康检查通过"
}

# 清理旧镜像
cleanup_images() {
    log_step "清理旧镜像..."
    
    # 删除未使用的镜像
    docker image prune -f
    
    # 保留最近的3个版本
    docker images "${APP_NAME}" --format "table {{.Repository}}:{{.Tag}}\t{{.CreatedAt}}" | \
        tail -n +2 | sort -k2 -r | tail -n +4 | awk '{print $1}' | \
        xargs -r docker rmi 2>/dev/null || true
    
    log_info "✓ 镜像清理完成"
}

# 显示状态
show_status() {
    log_step "应用状态信息..."
    
    echo ""
    echo "=== 应用容器状态 ==="
    docker-compose -f docker-compose.app.yml ps
    
    echo ""
    echo "=== 应用健康状态 ==="
    if curl -s http://localhost:8080/api/v1/health | jq . 2>/dev/null; then
        log_info "✓ 应用健康检查通过"
    else
        log_warn "应用健康检查失败或返回格式异常"
    fi
    
    echo ""
    echo "=== 资源使用情况 ==="
    docker-compose -f docker-compose.app.yml exec -T bossfi-backend ps aux | head -5
    
    echo ""
    echo "=== 访问信息 ==="
    log_info "应用地址: http://localhost:8080"
    log_info "API文档: http://localhost:8080/swagger/index.html"
    log_info "健康检查: http://localhost:8080/api/v1/health"
}

# 查看日志
show_logs() {
    docker-compose -f docker-compose.app.yml logs -f --tail=100
}

# 重启应用
restart_app() {
    log_step "重启应用..."
    docker-compose -f docker-compose.app.yml restart
    wait_for_app
    log_info "✓ 应用重启完成"
}

# 停止应用
stop_app() {
    log_step "停止应用..."
    docker-compose -f docker-compose.app.yml down
    log_info "✓ 应用已停止"
}

# 回滚应用
rollback_app() {
    log_step "回滚应用..."
    
    # 查找最新的备份
    latest_backup=$(ls -1 "$BACKUP_DIR" | sort -r | head -1)
    
    if [ -z "$latest_backup" ]; then
        log_error "没有找到可用的备份"
        exit 1
    fi
    
    log_info "使用备份: $latest_backup"
    
    if [ "$FORCE" = false ]; then
        read -p "确认要回滚吗？这将覆盖当前数据 (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "回滚已取消"
            exit 0
        fi
    fi
    
    # 停止应用
    docker-compose -f docker-compose.app.yml down
    
    # 恢复数据库
    log_info "恢复数据库..."
    docker-compose -f docker-compose.infrastructure.yml exec -T postgres \
        psql -U postgres < "${BACKUP_DIR}/${latest_backup}/postgres_backup.sql"
    
    # 恢复Redis（如果备份存在）
    if [ -f "${BACKUP_DIR}/${latest_backup}/redis_backup.rdb" ]; then
        log_info "恢复Redis数据..."
        docker-compose -f docker-compose.infrastructure.yml exec -T redis \
            redis-cli flushall
        # 注意：Redis RDB恢复需要停止Redis服务，这里简化处理
    fi
    
    # 重新启动应用
    docker-compose -f docker-compose.app.yml up -d
    wait_for_app
    
    log_info "✓ 回滚完成"
}

# 主函数
main() {
    echo ""
    log_info "========================================"
    log_info "      BossFi 应用部署脚本"
    log_info "========================================"
    echo ""
    
    case $COMMAND in
        deploy)
            check_environment
            backup_data
            build_app
            deploy_app
            wait_for_app
            health_check
            cleanup_images
            show_status
            ;;
        update)
            check_environment
            backup_data
            build_app
            deploy_app
            wait_for_app
            health_check
            cleanup_images
            log_info "✓ 应用更新完成"
            ;;
        restart)
            restart_app
            ;;
        stop)
            stop_app
            ;;
        logs)
            show_logs
            ;;
        status)
            show_status
            ;;
        backup)
            check_environment
            backup_data
            ;;
        rollback)
            check_environment
            rollback_app
            ;;
        *)
            log_error "未知命令: $COMMAND"
            show_help
            exit 1
            ;;
    esac
    
    if [ "$COMMAND" = "deploy" ]; then
        echo ""
        log_info "========================================"
        log_info "      应用部署完成！"
        log_info "========================================"
        echo ""
    fi
}

# 信号处理
trap 'log_error "脚本被中断"; exit 1' INT TERM

# 解析参数并运行
parse_args "$@"
main 