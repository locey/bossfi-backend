# 用户评论功能文档

## 功能概述

用户评论功能允许登录用户查看自己的所有评论，包括评论的详细信息、关联的文章信息以及父评论信息（如果是回复评论）。

## 功能特性

### 1. 查看用户评论列表
- 支持分页查询
- 按评论创建时间倒序排列
- 包含完整的文章信息
- 包含父评论信息（如果是回复）

### 2. 评论信息展示
- 评论基本信息（ID、内容、点赞数、创建时间等）
- 关联的文章信息（标题、内容摘要、分类、作者等）
- 父评论信息（如果是回复评论）
- 用户信息（评论作者）

### 3. 数据优化
- 文章内容自动截取（前100字符）
- 父评论内容自动截取（前50字符）
- 预加载关联数据，减少数据库查询

## API 接口

### 获取用户评论列表

**接口地址：** `GET /api/v1/user/comments`

**请求头：**
```
Authorization: Bearer <token>
```

**查询参数：**
| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| page | int | 否 | 1 | 页码，最小值为1 |
| page_size | int | 否 | 10 | 每页数量，最小值为1，最大值为50 |

**请求示例：**
```bash
curl -X GET "http://localhost:8080/api/v1/user/comments?page=1&page_size=10" \
  -H "Authorization: Bearer your_token_here"
```

**响应格式：**
```json
{
  "comments": [
    {
      "id": 1,
      "user_id": 1,
      "article_id": 1,
      "parent_id": null,
      "content": "这是一条评论内容",
      "like_count": 5,
      "is_deleted": false,
      "created_at": "2025-01-01T00:00:00Z",
      "user": {
        "id": 1,
        "username": "testuser",
        "nickname": "测试用户",
        "avatar": "https://example.com/avatar.jpg"
      },
      "article": {
        "id": 1,
        "title": "文章标题",
        "content": "文章内容摘要（前100字符）...",
        "category_id": 1,
        "like_count": 10,
        "comment_count": 5,
        "view_count": 100,
        "created_at": "2025-01-01T00:00:00Z",
        "user": {
          "id": 2,
          "username": "author",
          "nickname": "文章作者",
          "avatar": "https://example.com/author.jpg"
        },
        "category": {
          "id": 1,
          "name": "技术",
          "description": "技术相关文章",
          "icon": "tech-icon",
          "color": "#FF5733"
        }
      },
      "parent": null
    },
    {
      "id": 2,
      "user_id": 1,
      "article_id": 1,
      "parent_id": 1,
      "content": "这是对第一条评论的回复",
      "like_count": 2,
      "is_deleted": false,
      "created_at": "2025-01-01T01:00:00Z",
      "user": {
        "id": 1,
        "username": "testuser",
        "nickname": "测试用户",
        "avatar": "https://example.com/avatar.jpg"
      },
      "article": {
        "id": 1,
        "title": "文章标题",
        "content": "文章内容摘要（前100字符）...",
        "category_id": 1,
        "like_count": 10,
        "comment_count": 5,
        "view_count": 100,
        "created_at": "2025-01-01T00:00:00Z",
        "user": {
          "id": 2,
          "username": "author",
          "nickname": "文章作者",
          "avatar": "https://example.com/author.jpg"
        },
        "category": {
          "id": 1,
          "name": "技术",
          "description": "技术相关文章",
          "icon": "tech-icon",
          "color": "#FF5733"
        }
      },
      "parent": {
        "id": 1,
        "content": "这是一条评论内容（前50字符）...",
        "like_count": 5,
        "created_at": "2025-01-01T00:00:00Z",
        "user": {
          "id": 1,
          "username": "testuser",
          "nickname": "测试用户",
          "avatar": "https://example.com/avatar.jpg"
        }
      }
    }
  ],
  "total": 2,
  "page": 1,
  "page_size": 10
}
```

**响应字段说明：**

#### UserCommentResponse
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 评论ID |
| user_id | uint | 用户ID |
| article_id | uint | 文章ID |
| parent_id | *uint | 父评论ID，null表示顶级评论 |
| content | string | 评论内容 |
| like_count | int | 点赞数 |
| is_deleted | bool | 是否已删除 |
| created_at | time.Time | 创建时间 |
| user | UserInfo | 评论作者信息 |
| article | ArticleInfo | 文章信息 |
| parent | *CommentInfo | 父评论信息，null表示无父评论 |

#### ArticleInfo
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 文章ID |
| title | string | 文章标题 |
| content | string | 文章内容摘要（前100字符） |
| category_id | *uint | 分类ID |
| like_count | int | 点赞数 |
| comment_count | int | 评论数 |
| view_count | int | 浏览数 |
| created_at | time.Time | 创建时间 |
| user | UserInfo | 文章作者信息 |
| category | *CategoryInfo | 分类信息 |

#### CommentInfo
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 评论ID |
| content | string | 评论内容摘要（前50字符） |
| like_count | int | 点赞数 |
| created_at | time.Time | 创建时间 |
| user | UserInfo | 评论作者信息 |

## 错误响应

### 401 Unauthorized
```json
{
  "error": "unauthorized"
}
```

### 400 Bad Request
```json
{
  "error": "invalid pagination parameters"
}
```

### 500 Internal Server Error
```json
{
  "error": "database error"
}
```

## 使用示例

### JavaScript/TypeScript 示例

```typescript
// 获取用户评论列表
async function getUserComments(page = 1, pageSize = 10) {
  try {
    const response = await fetch(`/api/v1/user/comments?page=${page}&page_size=${pageSize}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('获取用户评论失败:', error);
    throw error;
  }
}

// 使用示例
getUserComments(1, 10)
  .then(data => {
    console.log('用户评论列表:', data);
    data.comments.forEach(comment => {
      console.log(`评论: ${comment.content}`);
      console.log(`文章: ${comment.article.title}`);
      if (comment.parent) {
        console.log(`回复: ${comment.parent.content}`);
      }
    });
  })
  .catch(error => {
    console.error('错误:', error);
  });
```

### Python 示例

```python
import requests

def get_user_comments(token, page=1, page_size=10):
    """获取用户评论列表"""
    url = f"http://localhost:8080/api/v1/user/comments"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    params = {
        "page": page,
        "page_size": page_size
    }
    
    try:
        response = requests.get(url, headers=headers, params=params)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"获取用户评论失败: {e}")
        raise

# 使用示例
try:
    data = get_user_comments("your_token_here", 1, 10)
    print(f"总评论数: {data['total']}")
    
    for comment in data['comments']:
        print(f"评论: {comment['content']}")
        print(f"文章: {comment['article']['title']}")
        if comment['parent']:
            print(f"回复: {comment['parent']['content']}")
        print("---")
        
except Exception as e:
    print(f"错误: {e}")
```

## 数据库设计

### 相关表结构

#### article_comments 表
```sql
CREATE TABLE article_comments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    article_id INTEGER NOT NULL,
    parent_id INTEGER,
    content TEXT NOT NULL,
    like_count INTEGER DEFAULT 0,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (article_id) REFERENCES articles(id),
    FOREIGN KEY (parent_id) REFERENCES article_comments(id)
);
```

#### articles 表
```sql
CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    category_id INTEGER,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (category_id) REFERENCES article_categories(id)
);
```

#### article_categories 表
```sql
CREATE TABLE article_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(200),
    icon VARCHAR(100),
    color VARCHAR(7),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 性能优化

### 1. 数据库查询优化
- 使用预加载（Preload）减少N+1查询问题
- 只查询未删除的评论（is_deleted = false）
- 按创建时间倒序排列，便于分页

### 2. 数据截取优化
- 文章内容自动截取前100字符，减少传输量
- 父评论内容自动截取前50字符，保持简洁

### 3. 缓存策略
- 可以考虑对热门文章的评论进行缓存
- 用户评论列表可以设置短期缓存

## 安全考虑

### 1. 认证授权
- 所有用户评论接口都需要JWT认证
- 只能查看自己的评论，不能查看其他用户的评论

### 2. 数据验证
- 分页参数验证（page >= 1, page_size >= 1 && <= 50）
- 用户ID验证，确保只能访问自己的数据

### 3. SQL注入防护
- 使用参数化查询
- 使用GORM的预加载功能

## 测试用例

### 单元测试
- 测试分页功能
- 测试评论排序
- 测试文章信息加载
- 测试父评论信息加载
- 测试未认证访问
- 测试无效参数

### 集成测试
- 测试完整的用户评论流程
- 测试多用户评论场景
- 测试回复评论场景

## 部署说明

### 1. 数据库迁移
确保数据库表结构已正确创建：
```bash
make migrate
```

### 2. 启动服务
```bash
make run
```

### 3. 运行测试
```bash
make test
```

## 监控和日志

### 1. 性能监控
- 监控API响应时间
- 监控数据库查询性能
- 监控内存使用情况

### 2. 错误日志
- 记录认证失败
- 记录数据库错误
- 记录参数验证错误

### 3. 访问日志
- 记录API调用频率
- 记录用户行为分析

## 扩展功能

### 1. 评论搜索
- 支持按关键词搜索评论内容
- 支持按文章标题搜索

### 2. 评论过滤
- 支持按时间范围过滤
- 支持按文章分类过滤
- 支持按评论类型过滤（顶级评论/回复）

### 3. 评论统计
- 用户评论总数统计
- 用户评论趋势分析
- 热门评论文章统计

## 常见问题

### Q1: 如何获取特定用户的评论？
A: 目前只支持获取当前登录用户的评论，这是出于隐私和安全考虑。

### Q2: 评论内容被截取了怎么办？
A: 如果需要完整内容，可以通过评论ID调用评论详情接口获取。

### Q3: 如何获取评论的回复列表？
A: 可以通过评论列表接口的parent_id参数获取特定评论的回复。

### Q4: 评论排序可以自定义吗？
A: 目前只支持按创建时间倒序排列，后续可以扩展支持其他排序方式。 