# AI评分功能使用示例

## 快速开始

### 1. 环境准备

确保已配置AiHubMix API密钥：
```bash
export AIHUBMIX_API_KEY="sk-VniUXYO3lAMincy8FcCb95AdCbE648De8dA4B0D96bF10380"
```

### 2. 数据库迁移

执行迁移脚本：
```bash
psql -d bossfi -f deploy/ai_scoring_migration.sql
```

### 3. 启动服务

```bash
make run
```

## 使用示例

### 示例1: 创建文章并自动评分

```bash
# 1. 创建文章
curl -X POST "http://localhost:8080/api/v1/articles" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "区块链技术发展趋势",
    "content": "区块链技术正在快速发展，从比特币的诞生到现在，已经经历了多个发展阶段...",
    "images": [],
    "category_id": 1
  }'

# 2. 等待几秒钟后查询评分结果
curl -X GET "http://localhost:8080/api/v1/ai-scoring/article/1"
```

### 示例2: 手动评分文章

```bash
# 对指定文章进行评分
curl -X POST "http://localhost:8080/api/v1/ai-scoring/score" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "article_id": 1
  }'
```

### 示例3: 批量评分

```bash
# 批量评分多篇文章
curl -X POST "http://localhost:8080/api/v1/ai-scoring/batch-score" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "article_ids": [1, 2, 3, 4, 5]
  }'
```

### 示例4: 获取未评分文章

```bash
# 获取未评分的文章列表
curl -X GET "http://localhost:8080/api/v1/ai-scoring/unscored?limit=10"
```

### 示例5: 自动评分新文章

```bash
# 自动对未评分的文章进行评分
curl -X POST "http://localhost:8080/api/v1/ai-scoring/auto-score?limit=5"
```

## 预期响应

### 成功评分响应
```json
{
  "article_id": 1,
  "score": 8.5,
  "reason": "文章内容质量较高，逻辑清晰，对读者有实用价值。作者对区块链技术有深入的理解，文章结构合理，信息准确可靠。",
  "scored_at": "2024-01-01T12:00:00Z",
  "success": true
}
```

### 批量评分响应
```json
{
  "total_articles": 3,
  "scored_count": 2,
  "failed_count": 1,
  "results": [
    {
      "article_id": 1,
      "score": 8.5,
      "reason": "评分理由...",
      "scored_at": "2024-01-01T12:00:00Z",
      "success": true
    },
    {
      "article_id": 2,
      "score": 7.2,
      "reason": "评分理由...",
      "scored_at": "2024-01-01T12:01:00Z",
      "success": true
    },
    {
      "article_id": 3,
      "scored_at": "2024-01-01T12:02:00Z",
      "success": false,
      "error_message": "文章不存在"
    }
  ]
}
```

## 错误处理示例

### API调用失败
```json
{
  "error": "API request failed",
  "message": "API request failed with status: 401"
}
```

### 文章不存在
```json
{
  "error": "Article not found",
  "message": "article not found: record not found"
}
```

### 参数错误
```json
{
  "error": "Invalid request body",
  "message": "Key: 'ScoreArticleRequest.ArticleID' Error:Field validation for 'ArticleID' failed on the 'required' tag"
}
```

## 测试脚本

运行完整的测试脚本：
```bash
./scripts/test_ai_scoring.sh
```

## 注意事项

1. **API密钥**: 确保AiHubMix API密钥配置正确
2. **网络连接**: 确保服务器能够访问AiHubMix API
3. **异步处理**: 评分是异步进行的，可能需要等待几秒钟
4. **错误处理**: 注意处理API调用失败的情况
5. **成本控制**: 监控API调用次数，避免超出限制 