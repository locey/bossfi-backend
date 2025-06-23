#!/bin/bash

# BossFi Blockchain Backend 部署脚本

set -e

DOCKER_IMAGE="bossfi-blockchain-backend"
CONTAINER_NAME="bossfi-backend"
NETWORK_NAME="bossfi-network"

echo "🚀 Deploying BossFi Blockchain Backend..."

# 检查Docker和Docker Compose
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed."
    exit 1
fi

# 停止现有服务
echo "🛑 Stopping existing services..."
docker-compose down

# 构建新镜像
echo "🔨 Building Docker image..."
docker build -t $DOCKER_IMAGE .

# 清理旧镜像
echo "🧹 Cleaning up old images..."
docker image prune -f

# 启动服务
echo "🚀 Starting services..."
docker-compose up -d

# 等待服务启动
echo "⏳ Waiting for services to start..."
sleep 10

# 健康检查
echo "🔍 Performing health check..."
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Health check passed!"
else
    echo "❌ Health check failed!"
    echo "📋 Checking logs..."
    docker-compose logs app
    exit 1
fi

echo "🎉 Deployment completed successfully!"
echo "📊 Service status:"
docker-compose ps

echo "🌐 Service URLs:"
echo "  - API: http://localhost:8080"
echo "  - API Docs: http://localhost:8080/swagger/index.html"
echo "  - Health Check: http://localhost:8080/health" 