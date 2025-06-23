#!/bin/bash

# BossFi Blockchain Backend 启动脚本

set -e

echo "🚀 Starting BossFi Blockchain Backend..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# 检查Go版本
GO_VERSION=$(go version | cut -d' ' -f3 | cut -d'.' -f1,2)
REQUIRED_VERSION="go1.21"

if [[ "$GO_VERSION" < "$REQUIRED_VERSION" ]]; then
    echo "❌ Go version $REQUIRED_VERSION or later is required. Current version: $GO_VERSION"
    exit 1
fi

# 检查配置文件
if [ ! -f "configs/config.toml" ]; then
    echo "❌ Configuration file not found. Please create configs/config.toml"
    echo "📋 You can copy from configs/config.toml.example if available"
    exit 1
fi

# 创建必要的目录
mkdir -p logs
mkdir -p tmp

echo "📦 Installing dependencies..."
go mod download
go mod tidy

echo "🔧 Building application..."
go build -o bossfi-blockchain-backend ./cmd/server

echo "✅ Build completed successfully!"

# 检查数据库连接（可选）
echo "🔍 Checking database connection..."
# 这里可以添加数据库连接检查逻辑

echo "🎯 Starting server..."
./bossfi-blockchain-backend

echo "🎉 Server started successfully!" 