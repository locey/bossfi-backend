#!/bin/bash

# AI评分重试功能测试脚本
# 测试定时任务和手动重试功能

BASE_URL="http://localhost:8080/api/v1"

echo "=== AI评分重试功能测试 ==="

# 1. 测试获取未评分文章
echo "1. 获取未评分文章列表..."
curl -X GET "${BASE_URL}/ai-scoring/unscored?limit=5" \
  -H "Content-Type: application/json" \
  -w "\nHTTP状态码: %{http_code}\n\n"

# 2. 测试手动重试失败的评分
echo "2. 手动重试失败的评分..."
curl -X POST "${BASE_URL}/ai-scoring/retry-failed?limit=5" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -w "\nHTTP状态码: %{http_code}\n\n"

# 3. 测试手动重试待评分的文章
echo "3. 手动重试待评分的文章..."
curl -X POST "${BASE_URL}/ai-scoring/retry-pending?limit=10" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -w "\nHTTP状态码: %{http_code}\n\n"

# 4. 测试获取文章评分（包含状态信息）
echo "4. 获取文章评分（包含状态信息）..."
curl -X GET "${BASE_URL}/ai-scoring/article/1" \
  -H "Content-Type: application/json" \
  -w "\nHTTP状态码: %{http_code}\n\n"

# 5. 测试自动评分新文章
echo "5. 自动评分新文章..."
curl -X POST "${BASE_URL}/ai-scoring/auto-score?limit=3" \
  -H "Content-Type: application/json" \
  -w "\nHTTP状态码: %{http_code}\n\n"

echo "=== 测试完成 ==="
echo ""
echo "注意事项："
echo "1. 请确保服务器正在运行"
echo "2. 请确保数据库已执行评分状态迁移脚本"
echo "3. 请替换YOUR_TOKEN_HERE为有效的认证token"
echo "4. 定时任务默认每2小时执行一次，可通过环境变量AI_SCORING_RETRY_INTERVAL调整"
echo "5. 定时任务会自动重试失败的评分和待评分的文章" 