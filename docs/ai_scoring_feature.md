# AI评分功能文档

## 概述

AI评分功能使用AiHubMix API对用户发布的文章进行自动评分，提供客观的内容质量评估。

## 功能特性

### 1. 自动评分
- 文章创建时自动触发AI评分
- 异步处理，不影响文章发布速度
- 评分范围：0-10分，保留两位小数

### 2. 评分标准
- **内容质量 (40%)**: 信息准确性、深度、原创性
- **表达清晰度 (30%)**: 语言流畅性、逻辑结构、可读性
- **价值贡献 (20%)**: 对读者的实用价值、启发性
- **创新性 (10%)**: 观点新颖性、独特见解

### 3. API接口

#### 公开接口（无需认证）

##### 获取文章评分
```
GET /api/v1/ai-scoring/article/{article_id}
```

##### 获取未评分文章
```
GET /api/v1/ai-scoring/unscored?limit=10
```

##### 自动评分新文章
```
POST /api/v1/ai-scoring/auto-score?limit=5
```

#### 需要认证的接口

##### 手动评分单篇文章
```
POST /api/v1/ai-scoring/score
Content-Type: application/json
Authorization: Bearer {token}

{
  "article_id": 1
}
```

##### 批量评分文章
```
POST /api/v1/ai-scoring/batch-score
Content-Type: application/json
Authorization: Bearer {token}

{
  "article_ids": [1, 2, 3]
}
```

## 数据库结构

### articles表新增字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| score | DECIMAL(3,2) | AI评分 (0-10分) |
| score_time | TIMESTAMP | AI评分时间 |
| score_reason | TEXT | AI评分理由 |
| score_status | INT | 评分状态: 0-待评分, 1-评分中, 2-评分成功, -1-评分失败 |

### 索引
- `idx_articles_score`: 评分索引
- `idx_articles_score_time`: 评分时间索引
- `idx_articles_score_null`: 未评分文章索引
- `idx_articles_score_status`: 评分状态索引
- `idx_articles_score_status_pending`: 待评分文章索引
- `idx_articles_score_status_failed`: 评分失败文章索引

## 配置

### 环境变量

```bash
# AiHubMix API配置
AIHUBMIX_API_KEY=sk-VniUXYO3lAMincy8FcCb95AdCbE648De8dA4B0D96bF10380
AIHUBMIX_BASE_URL=https://api.aihubmix.com

# 定时任务配置
AI_SCORING_RETRY_INTERVAL=0 */2 * * *  # 每2小时执行一次重试
```

### 默认配置

```go
AiHubMix: AiHubMixConfig{
    APIKey:  "sk-VniUXYO3lAMincy8FcCb95AdCbE648De8dA4B0D96bF10380",
    BaseURL: "https://api.aihubmix.com",
}
```

## 部署步骤

### 1. 数据库迁移

执行迁移脚本：
```sql
-- 文件: deploy/ai_scoring_migration.sql
-- 文件: deploy/ai_scoring_status_migration.sql
```

### 2. 环境配置

在`.env`文件中添加：
```bash
AIHUBMIX_API_KEY=sk-VniUXYO3lAMincy8FcCb95AdCbE648De8dA4B0D96bF10380
AIHUBMIX_BASE_URL=https://api.aihubmix.com
AI_SCORING_RETRY_INTERVAL=0 */2 * * *  # 每2小时执行一次重试
```

### 3. 重启服务

```bash
make run
```

## 测试

### 运行测试脚本

```bash
chmod +x scripts/test_ai_scoring.sh
./scripts/test_ai_scoring.sh
```

### 手动测试

1. 创建文章后，系统会自动触发AI评分
2. 使用API接口查询评分结果
3. 可以手动触发评分或批量评分

## 响应格式

### 评分响应
```json
{
  "article_id": 1,
  "score": 8.5,
  "reason": "文章内容质量较高，逻辑清晰，对读者有实用价值...",
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
    }
  ]
}
```

## 错误处理

### 常见错误

1. **API调用失败**: AiHubMix API不可用或配置错误
2. **文章不存在**: 指定的文章ID不存在
3. **评分解析失败**: AI返回的评分格式不正确
4. **网络超时**: API调用超时

### 错误响应格式
```json
{
  "error": "错误类型",
  "message": "详细错误信息"
}
```

## 监控和日志

### 日志记录
- AI评分成功/失败日志
- API调用耗时统计
- 评分结果统计

### 监控指标
- 评分成功率
- 平均评分
- API调用频率
- 错误率统计

## 注意事项

1. **API限制**: 注意AiHubMix API的调用频率限制
2. **成本控制**: AI API调用会产生费用，需要监控使用量
3. **数据隐私**: 确保文章内容在传输过程中的安全性
4. **评分准确性**: AI评分仅供参考，不应作为唯一的质量标准
5. **异步处理**: 评分是异步进行的，创建文章后评分结果可能稍后才会出现

## 扩展功能

### 未来可能的改进

1. **评分历史**: 记录评分历史，支持重新评分
2. **评分权重**: 根据文章类型调整评分权重
3. **用户反馈**: 结合用户反馈优化评分算法
4. **多模型支持**: 支持多种AI模型进行评分
5. **评分缓存**: 缓存评分结果，减少API调用 