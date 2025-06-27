#!/bin/bash

# 检查脚本版本
echo "=== BossFi 脚本版本检查 ==="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "当前脚本目录: $SCRIPT_DIR"
echo ""

# 检查deploy.sh中是否还有旧的cp命令
echo "=== 检查 deploy.sh ==="
if grep -q "cp -r \./deploy/\*" "$SCRIPT_DIR/deploy.sh" 2>/dev/null; then
    echo "❌ deploy.sh 包含旧的cp命令，需要更新"
    echo "问题行："
    grep -n "cp -r \./deploy/\*" "$SCRIPT_DIR/deploy.sh"
else
    echo "✅ deploy.sh 没有发现旧的cp命令"
fi

# 检查是否有其他cp命令
echo ""
echo "=== 搜索所有cp命令 ==="
if grep -n "cp " "$SCRIPT_DIR/deploy.sh" 2>/dev/null; then
    echo "发现cp命令，请检查是否正确"
else
    echo "✅ 没有发现cp命令"
fi

# 检查关键路径变量
echo ""
echo "=== 检查路径变量定义 ==="
if grep -q "SCRIPT_DIR=" "$SCRIPT_DIR/deploy.sh"; then
    echo "✅ 找到SCRIPT_DIR定义"
    grep -n "SCRIPT_DIR=" "$SCRIPT_DIR/deploy.sh"
else
    echo "❌ 未找到SCRIPT_DIR定义"
fi

if grep -q "PROJECT_ROOT=" "$SCRIPT_DIR/deploy.sh"; then
    echo "✅ 找到PROJECT_ROOT定义"
    grep -n "PROJECT_ROOT=" "$SCRIPT_DIR/deploy.sh"
else
    echo "❌ 未找到PROJECT_ROOT定义"
fi

# 检查脚本修改时间
echo ""
echo "=== 脚本文件信息 ==="
ls -la "$SCRIPT_DIR/deploy.sh"
ls -la "$SCRIPT_DIR/monitor.sh"
ls -la "$SCRIPT_DIR/update.sh" 2>/dev/null || echo "update.sh 不存在"

echo ""
echo "=== 检查完成 ==="
echo "如果发现问题，请重新从GitHub拉取最新代码或手动更新脚本" 