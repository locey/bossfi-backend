#!/bin/bash

# 路径测试脚本
echo "=== BossFi 路径测试 ==="

# 获取脚本所在目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "脚本目录: $SCRIPT_DIR"
echo "项目根目录: $PROJECT_ROOT"

echo ""
echo "=== 检查关键文件是否存在 ==="

# 检查部署相关文件
files=(
    "$SCRIPT_DIR/docker-compose.prod.yml"
    "$PROJECT_ROOT/.env"
    "$SCRIPT_DIR/nginx.conf"
    "$SCRIPT_DIR/deploy.sh"
    "$SCRIPT_DIR/monitor.sh"
    "$PROJECT_ROOT/Dockerfile"
    "$PROJECT_ROOT/go.mod"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file"
    else
        echo "❌ $file"
    fi
done

echo ""
echo "=== 检查目录结构 ==="

# 检查目录
dirs=(
    "$PROJECT_ROOT/api"
    "$PROJECT_ROOT/config"
    "$PROJECT_ROOT/db"
    "$SCRIPT_DIR"
)

for dir in "${dirs[@]}"; do
    if [ -d "$dir" ]; then
        echo "✅ $dir/"
    else
        echo "❌ $dir/"
    fi
done

echo ""
echo "=== Docker Compose 构建上下文测试 ==="
cd "$SCRIPT_DIR"
if [ -f "docker-compose.prod.yml" ]; then
    echo "当前目录: $(pwd)"
    echo "构建上下文 (..) 指向: $(cd .. && pwd)"
    if [ -f "../Dockerfile" ]; then
        echo "✅ Dockerfile 在构建上下文中找到"
    else
        echo "❌ Dockerfile 未找到"
    fi
else
    echo "❌ docker-compose.prod.yml 不存在"
fi

echo ""
echo "=== 测试完成 ===" 