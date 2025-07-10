#!/bin/bash

# AI评分功能测试脚本
# 测试AiHubMix API对文章进行评分的功能

BASE_URL="http://localhost:8080/api/v1"
API_KEY="sk-VniUXYO3lAMincy8FcCb95AdCbE648De8dA4B0D96bF10380"

echo "=== AI评分功能测试 ==="

# 1. 测试获取未评分文章
echo "1. 获取未评分文章列表..."
curl -X GET "${BASE_URL}/ai-scoring/unscored?limit=5" \
  -H "Content-Type: application/json" \
  -w "\nHTTP状态码: %{http_code}\n\n"

# 2. 测试自动评分新文章
echo "2. 自动评分新文章..."
curl -X POST "${BASE_URL}/ai-scoring/auto-score?limit=3" \
  -H "Content-Type: application/json" \
  -w "\nHTTP状态码: %{http_code}\n\n"

# 3. 测试获取文章评分
echo "3. 获取文章评分 (假设文章ID为1)..."
curl -X GET "${BASE_URL}/ai-scoring/article/1" \
  -H "Content-Type: application/json" \
  -w "\nHTTP状态码: %{http_code}\n\n"

# 4. 测试手动评分单篇文章
echo "4. 手动评分单篇文章..."
curl -X POST "${BASE_URL}/ai-scoring/score" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "article_id": 1
  }' \
  -w "\nHTTP状态码: %{http_code}\n\n"

# 5. 测试批量评分文章
echo "5. 批量评分文章..."
curl -X POST "${BASE_URL}/ai-scoring/batch-score" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "article_ids": [1, 2, 3]
  }' \
  -w "\nHTTP状态码: %{http_code}\n\n"

echo "=== 测试完成 ==="
echo ""
echo "注意事项："
echo "1. 请确保服务器正在运行"
echo "2. 请确保数据库已执行AI评分迁移脚本"
echo "3. 请替换YOUR_TOKEN_HERE为有效的认证token"
echo "4. 请确保AiHubMix API密钥配置正确" 