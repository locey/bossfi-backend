#!/bin/bash

# BossFi 分类功能测试脚本

set -e

echo "🚀 开始测试 BossFi 分类功能..."

# 检查服务是否运行
echo "📡 检查服务状态..."
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ 服务未运行，请先启动服务: make run"
    exit 1
fi

echo "✅ 服务运行正常"

# 获取测试token
echo "🔑 获取测试token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/test-token \
    -H "Content-Type: application/json" \
    -d '{"wallet_address": "0x1234567890123456789012345678901234567890"}')

TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ 获取token失败"
    echo "响应: $TOKEN_RESPONSE"
    exit 1
fi

echo "✅ Token获取成功"

# 测试创建分类
echo "📝 测试创建分类..."
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/categories \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "name": "测试分类",
        "description": "这是一个测试分类",
        "icon": "test-icon",
        "color": "#FF5733",
        "sort_order": 1
    }')

CATEGORY_ID=$(echo $CREATE_RESPONSE | jq -r '.id')

if [ "$CATEGORY_ID" = "null" ] || [ -z "$CATEGORY_ID" ]; then
    echo "❌ 创建分类失败"
    echo "响应: $CREATE_RESPONSE"
    exit 1
fi

echo "✅ 分类创建成功，ID: $CATEGORY_ID"

# 测试获取分类列表
echo "📋 测试获取分类列表..."
CATEGORIES_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/categories?page=1&page_size=10")

if echo $CATEGORIES_RESPONSE | jq -e '.categories' > /dev/null; then
    echo "✅ 获取分类列表成功"
else
    echo "❌ 获取分类列表失败"
    echo "响应: $CATEGORIES_RESPONSE"
    exit 1
fi

# 测试获取活跃分类
echo "🔥 测试获取活跃分类..."
ACTIVE_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/categories/active")

if echo $ACTIVE_RESPONSE | jq -e '.categories' > /dev/null; then
    echo "✅ 获取活跃分类成功"
else
    echo "❌ 获取活跃分类失败"
    echo "响应: $ACTIVE_RESPONSE"
    exit 1
fi

# 测试创建带分类的文章
echo "📄 测试创建带分类的文章..."
ARTICLE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/articles \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"title\": \"测试文章 - 分类功能\",
        \"content\": \"这是一篇测试文章，用于验证分类功能。包含区块链、技术等关键字。\",
        \"category_id\": $CATEGORY_ID,
        \"images\": [\"https://example.com/test.jpg\"]
    }")

ARTICLE_ID=$(echo $ARTICLE_RESPONSE | jq -r '.id')

if [ "$ARTICLE_ID" = "null" ] || [ -z "$ARTICLE_ID" ]; then
    echo "❌ 创建文章失败"
    echo "响应: $ARTICLE_RESPONSE"
    exit 1
fi

echo "✅ 文章创建成功，ID: $ARTICLE_ID"

# 测试按分类筛选文章
echo "🔍 测试按分类筛选文章..."
FILTER_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/articles?category_id=$CATEGORY_ID&page=1&page_size=10")

if echo $FILTER_RESPONSE | jq -e '.articles' > /dev/null; then
    echo "✅ 按分类筛选文章成功"
else
    echo "❌ 按分类筛选文章失败"
    echo "响应: $FILTER_RESPONSE"
    exit 1
fi

# 测试关键字搜索
echo "🔎 测试关键字搜索..."
SEARCH_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/articles?keyword=区块链&page=1&page_size=10")

if echo $SEARCH_RESPONSE | jq -e '.articles' > /dev/null; then
    echo "✅ 关键字搜索成功"
else
    echo "❌ 关键字搜索失败"
    echo "响应: $SEARCH_RESPONSE"
    exit 1
fi

# 测试组合筛选
echo "🎯 测试组合筛选..."
COMBINE_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/articles?category_id=$CATEGORY_ID&keyword=技术&page=1&page_size=10")

if echo $COMBINE_RESPONSE | jq -e '.articles' > /dev/null; then
    echo "✅ 组合筛选成功"
else
    echo "❌ 组合筛选失败"
    echo "响应: $COMBINE_RESPONSE"
    exit 1
fi

# 测试更新分类
echo "✏️ 测试更新分类..."
UPDATE_RESPONSE=$(curl -s -X PUT http://localhost:8080/api/v1/categories/$CATEGORY_ID \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "name": "更新后的测试分类",
        "description": "这是更新后的描述",
        "icon": "updated-icon",
        "color": "#33FF57",
        "sort_order": 2,
        "is_active": true
    }')

if echo $UPDATE_RESPONSE | jq -e '.id' > /dev/null; then
    echo "✅ 更新分类成功"
else
    echo "❌ 更新分类失败"
    echo "响应: $UPDATE_RESPONSE"
    exit 1
fi

# 清理测试数据
echo "🧹 清理测试数据..."
curl -s -X DELETE http://localhost:8080/api/v1/articles/$ARTICLE_ID \
    -H "Authorization: Bearer $TOKEN" > /dev/null

curl -s -X DELETE http://localhost:8080/api/v1/categories/$CATEGORY_ID \
    -H "Authorization: Bearer $TOKEN" > /dev/null

echo "✅ 测试数据清理完成"

echo ""
echo "🎉 所有测试通过！分类功能正常工作"
echo ""
echo "📊 测试总结："
echo "  ✅ 分类创建"
echo "  ✅ 分类列表获取"
echo "  ✅ 活跃分类获取"
echo "  ✅ 文章分类关联"
echo "  ✅ 分类筛选"
echo "  ✅ 关键字搜索"
echo "  ✅ 组合筛选"
echo "  ✅ 分类更新"
echo "  ✅ 数据清理"
echo ""
echo "🚀 分类功能已成功集成到 BossFi Backend！" 