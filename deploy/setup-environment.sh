#!/bin/bash

# BossFi 基础环境安装脚本
# 适用于 Ubuntu/Debian 系统
# 使用方法: sudo ./setup-environment.sh

set -e

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
        error "此脚本需要root权限运行，请使用 sudo ./setup-environment.sh"
    fi
}

# 检测系统版本
detect_system() {
    log "检测系统版本..."
    
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
        log "检测到系统: $OS $VER"
    else
        error "无法检测系统版本"
    fi
    
    # 检查是否为支持的系统
    case $OS in
        "Ubuntu"|"Debian GNU/Linux")
            log "系统支持，继续安装..."
            ;;
        *)
            warn "未测试的系统，可能需要手动调整安装命令"
            ;;
    esac
}

# 更新系统包
update_system() {
    log "更新系统包..."
    apt-get update -y
    apt-get upgrade -y
    log "系统包更新完成"
}

# 安装基础工具
install_basic_tools() {
    log "安装基础工具..."
    apt-get install -y \
        curl \
        wget \
        git \
        unzip \
        vim \
        htop \
        tree \
        jq \
        ca-certificates \
        gnupg \
        lsb-release \
        software-properties-common \
        apt-transport-https
    log "基础工具安装完成"
}

# 安装 Docker
install_docker() {
    log "安装 Docker..."
    
    # 检查 Docker 是否已安装
    if command -v docker &> /dev/null; then
        log "Docker 已安装，版本: $(docker --version)"
        return 0
    fi
    
    # 添加 Docker 官方 GPG 密钥
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    
    # 添加 Docker 仓库
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # 更新包索引
    apt-get update -y
    
    # 安装 Docker
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    
    # 启动并启用 Docker 服务
    systemctl start docker
    systemctl enable docker
    
    # 验证安装
    docker --version
    log "Docker 安装完成"
}

# 安装 Docker Compose
install_docker_compose() {
    log "安装 Docker Compose..."
    
    # 检查 Docker Compose 是否已安装
    if command -v docker-compose &> /dev/null; then
        log "Docker Compose 已安装，版本: $(docker-compose --version)"
        return 0
    fi
    
    # 获取最新版本号
    DOCKER_COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | jq -r .tag_name)
    
    # 下载 Docker Compose
    curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    
    # 设置执行权限
    chmod +x /usr/local/bin/docker-compose
    
    # 创建软链接
    ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
    
    # 验证安装
    docker-compose --version
    log "Docker Compose 安装完成"
}

# 配置 Docker
configure_docker() {
    log "配置 Docker..."
    
    # 创建 docker 用户组（如果不存在）
    if ! getent group docker > /dev/null 2>&1; then
        groupadd docker
        log "创建 docker 用户组"
    fi
    
    # 配置 Docker 镜像加速器（使用阿里云镜像）
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json << EOF
{
    "registry-mirrors": [
        "https://docker.mirrors.ustc.edu.cn",
        "https://hub-mirror.c.163.com",
        "https://mirror.baidubce.com"
    ],
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "10m",
        "max-file": "3"
    },
    "storage-driver": "overlay2"
}
EOF
    
    # 重启 Docker 服务
    systemctl daemon-reload
    systemctl restart docker
    
    log "Docker 配置完成"
}

# 安装 Nginx
install_nginx() {
    log "安装 Nginx..."
    
    # 检查 Nginx 是否已安装
    if command -v nginx &> /dev/null; then
        log "Nginx 已安装，版本: $(nginx -v 2>&1)"
        return 0
    fi
    
    # 安装 Nginx
    apt-get install -y nginx
    
    # 启动并启用 Nginx 服务
    systemctl start nginx
    systemctl enable nginx
    
    # 验证安装
    nginx -v
    log "Nginx 安装完成"
}

# 配置防火墙
configure_firewall() {
    log "配置防火墙..."
    
    # 检查 ufw 是否安装
    if ! command -v ufw &> /dev/null; then
        apt-get install -y ufw
    fi
    
    # 配置防火墙规则
    ufw --force reset
    ufw default deny incoming
    ufw default allow outgoing
    
    # 允许 SSH
    ufw allow ssh
    ufw allow 22/tcp
    
    # 允许 HTTP 和 HTTPS
    ufw allow 80/tcp
    ufw allow 443/tcp
    
    # 允许应用端口（仅本地访问）
    ufw allow from 127.0.0.1 to any port 8080
    ufw allow from 127.0.0.1 to any port 5432
    ufw allow from 127.0.0.1 to any port 6379
    
    # 启用防火墙
    ufw --force enable
    
    log "防火墙配置完成"
}

# 创建项目目录结构
create_project_structure() {
    log "创建项目目录结构..."
    
    # 创建项目根目录
    mkdir -p /opt/bossfi
    mkdir -p /opt/bossfi/logs
    mkdir -p /opt/bossfi/backups
    mkdir -p /opt/bossfi/ssl
    mkdir -p /opt/bossfi/data/postgres
    mkdir -p /opt/bossfi/data/redis
    
    # 设置目录权限
    chown -R root:root /opt/bossfi
    chmod -R 755 /opt/bossfi
    
    log "项目目录结构创建完成"
}

# 安装监控工具
install_monitoring_tools() {
    log "安装监控工具..."
    
    # 安装系统监控工具
    apt-get install -y \
        htop \
        iotop \
        nethogs \
        ncdu \
        glances
    
    log "监控工具安装完成"
}

# 系统优化
optimize_system() {
    log "优化系统配置..."
    
    # 优化内核参数
    cat >> /etc/sysctl.conf << EOF

# BossFi 系统优化
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 5000
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_fin_timeout = 10
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_keepalive_probes = 3
net.ipv4.tcp_keepalive_intvl = 30
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
EOF
    
    # 应用内核参数
    sysctl -p
    
    # 优化文件描述符限制
    cat >> /etc/security/limits.conf << EOF

# BossFi 文件描述符限制
* soft nofile 65535
* hard nofile 65535
root soft nofile 65535
root hard nofile 65535
EOF
    
    log "系统优化完成"
}

# 主函数
main() {
    log "开始安装 BossFi 基础环境..."
    log "================================================"
    
    check_root
    detect_system
    update_system
    install_basic_tools
    install_docker
    install_docker_compose
    configure_docker
    install_nginx
    configure_firewall
    create_project_structure
    install_monitoring_tools
    optimize_system
    
    log "================================================"
    log "✅ BossFi 基础环境安装完成！"
    log ""
    log "📋 安装摘要:"
    log "  - Docker: $(docker --version 2>/dev/null || echo '未安装')"
    log "  - Docker Compose: $(docker-compose --version 2>/dev/null || echo '未安装')"
    log "  - Nginx: $(nginx -v 2>&1 || echo '未安装')"
    log "  - 项目目录: /opt/bossfi"
    log ""
    log "🔧 下一步:"
    log "  1. 将项目代码克隆到 /opt/bossfi/"
    log "  2. 运行项目部署脚本: ./deploy-project.sh"
    log ""
    log "💡 有用的命令:"
    log "  - 查看 Docker 状态: systemctl status docker"
    log "  - 查看防火墙状态: ufw status"
    log "  - 查看系统资源: htop"
    log "  - 查看项目日志: tail -f /opt/bossfi/logs/*.log"
}

# 如果脚本被直接执行
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 