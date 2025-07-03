# 文章分类功能说明

## 功能概述

BossFi Backend 新增了文章分类功能，支持对文章进行分类管理，并提供分类筛选和关键字搜索功能。

## 主要特性

### 1. 分类管理
- 创建、编辑、删除文章分类
- 分类支持名称、描述、图标、颜色等属性
- 分类排序和活跃状态管理
- 分类文章数量统计

### 2. 文章分类
- 文章可以关联到特定分类
- 支持无分类文章（category_id 为 null）
- 分类信息在文章详情中显示

### 3. 搜索和筛选
- 按分类筛选文章
- 关键字搜索（标题和内容）
- 支持组合筛选（分类 + 关键字 + 用户等）

## 数据库设计

### 文章分类表 (article_categories)
```sql
CREATE TABLE article_categories (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description VARCHAR(200),
    icon VARCHAR(100),
    color VARCHAR(7),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 文章表更新
```sql
ALTER TABLE articles ADD COLUMN category_id BIGINT;
ALTER TABLE articles ADD CONSTRAINT fk_articles_category_id 
    FOREIGN KEY (category_id) REFERENCES article_categories(id) ON DELETE SET NULL;
```

## API 接口

### 分类管理接口

#### 1. 创建分类
```
POST /api/v1/categories
Content-Type: application/json
Authorization: Bearer <token>

{
    "name": "技术",
    "description": "技术相关文章",
    "icon": "tech-icon",
    "color": "#FF5733",
    "sort_order": 1
}
```

#### 2. 获取分类列表
```
GET /api/v1/categories?page=1&page_size=10&is_active=true
```

#### 3. 获取活跃分类
```
GET /api/v1/categories/active
```

#### 4. 获取分类详情
```
GET /api/v1/categories/{id}
```

#### 5. 更新分类
```
PUT /api/v1/categories/{id}
Content-Type: application/json
Authorization: Bearer <token>

{
    "name": "更新后的分类名",
    "description": "更新后的描述",
    "icon": "updated-icon",
    "color": "#33FF57",
    "sort_order": 2,
    "is_active": true
}
```

#### 6. 删除分类
```
DELETE /api/v1/categories/{id}
Authorization: Bearer <token>
```

### 文章接口更新

#### 1. 创建文章（支持分类）
```
POST /api/v1/articles
Content-Type: application/json
Authorization: Bearer <token>

{
    "title": "文章标题",
    "content": "文章内容",
    "category_id": 1,
    "images": ["https://example.com/image1.jpg"]
}
```

#### 2. 获取文章列表（支持分类筛选和搜索）
```
GET /api/v1/articles?page=1&page_size=10&category_id=1&keyword=区块链&sort_by=created_at&sort_order=desc
```

查询参数说明：
- `category_id`: 分类ID筛选
- `keyword`: 关键字搜索（标题和内容）
- `user_id`: 用户ID筛选
- `sort_by`: 排序字段（created_at, like_count, view_count）
- `sort_order`: 排序方向（asc, desc）

#### 3. 更新文章（支持分类）
```
PUT /api/v1/articles/{id}
Content-Type: application/json
Authorization: Bearer <token>

{
    "title": "更新后的标题",
    "content": "更新后的内容",
    "category_id": 2,
    "images": ["https://example.com/image2.jpg"]
}
```

## 响应格式

### 分类响应
```json
{
    "id": 1,
    "name": "技术",
    "description": "技术相关文章",
    "icon": "tech-icon",
    "color": "#FF5733",
    "sort_order": 1,
    "is_active": true,
    "article_count": 10,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
}
```

### 文章响应（包含分类信息）
```json
{
    "id": 1,
    "user_id": 1,
    "category_id": 1,
    "title": "文章标题",
    "content": "文章内容",
    "images": ["https://example.com/image.jpg"],
    "like_count": 10,
    "comment_count": 5,
    "view_count": 100,
    "is_deleted": false,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z",
    "user": {
        "id": 1,
        "username": "用户名",
        "avatar": "https://example.com/avatar.jpg",
        "wallet_address": "0x1234..."
    },
    "category": {
        "id": 1,
        "name": "技术",
        "description": "技术相关文章",
        "icon": "tech-icon",
        "color": "#FF5733"
    }
}
```

## 默认分类数据

系统初始化时会创建以下默认分类：

1. **技术** - 技术相关文章，包括编程、开发、架构等
2. **区块链** - 区块链技术、加密货币、DeFi等相关内容
3. **投资** - 投资理财、市场分析、投资策略等
4. **生活** - 日常生活、个人感悟、生活技巧等
5. **新闻** - 行业新闻、热点事件、重要公告等
6. **教程** - 学习教程、操作指南、最佳实践等
7. **观点** - 个人观点、行业分析、深度思考等
8. **其他** - 其他类型的内容

## 使用示例

### 前端实现示例

#### 1. 获取分类列表
```javascript
// 获取所有活跃分类
const getCategories = async () => {
    const response = await fetch('/api/v1/categories/active');
    const data = await response.json();
    return data.categories;
};
```

#### 2. 分类筛选文章
```javascript
// 按分类筛选文章
const getArticlesByCategory = async (categoryId, page = 1) => {
    const params = new URLSearchParams({
        page: page,
        page_size: 10,
        category_id: categoryId,
        sort_by: 'created_at',
        sort_order: 'desc'
    });
    
    const response = await fetch(`/api/v1/articles?${params}`);
    const data = await response.json();
    return data;
};
```

#### 3. 关键字搜索
```javascript
// 关键字搜索文章
const searchArticles = async (keyword, page = 1) => {
    const params = new URLSearchParams({
        page: page,
        page_size: 10,
        keyword: keyword,
        sort_by: 'created_at',
        sort_order: 'desc'
    });
    
    const response = await fetch(`/api/v1/articles?${params}`);
    const data = await response.json();
    return data;
};
```

#### 4. 组合筛选
```javascript
// 组合筛选：分类 + 关键字
const filterArticles = async (categoryId, keyword, page = 1) => {
    const params = new URLSearchParams({
        page: page,
        page_size: 10,
        sort_by: 'created_at',
        sort_order: 'desc'
    });
    
    if (categoryId) {
        params.append('category_id', categoryId);
    }
    
    if (keyword) {
        params.append('keyword', keyword);
    }
    
    const response = await fetch(`/api/v1/articles?${params}`);
    const data = await response.json();
    return data;
};
```

## 注意事项

1. **分类删除限制**: 只能删除没有文章的分类，有文章的分类无法删除
2. **分类名称唯一性**: 分类名称必须唯一，不能重复
3. **文章分类可选**: 文章可以不设置分类（category_id 为 null）
4. **搜索性能**: 关键字搜索使用 ILIKE 进行模糊匹配，建议在标题和内容字段上建立索引
5. **权限控制**: 分类的创建、编辑、删除需要用户认证，可以根据需要添加管理员权限检查

## 测试

运行分类功能测试：
```bash
# 运行分类测试
go test -v ./test/category_test.go

# 运行所有测试
make test
```

## 部署

1. 更新数据库结构：
```bash
# 运行数据库迁移
psql -h localhost -U postgres -d bossfi -f deploy/init.sql
psql -h localhost -U postgres -d bossfi -f deploy/category_data.sql
```

2. 重启应用：
```bash
make run
```

3. 验证功能：
- 访问 Swagger 文档：http://localhost:8080/swagger/index.html
- 测试分类相关接口
- 测试文章的分类筛选和搜索功能 