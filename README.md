# BossFi Backend

基于区块链的去中心化求职平台后端服务

## 📋 项目简介

BossFi Backend 是一个使用 Go 语言开发的现代化后端服务，为去中心化求职平台提供 API 支持。项目采用清洁架构设计，支持钱包登录、帖子管理、质押奖励等核心功能。

## 🚀 技术栈

- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL
- **ORM**: GORM
- **认证**: JWT + 钱包签名
- **日志**: Zap
- **文档**: Swagger
- **配置**: TOML

## 🏗️ 项目结构

```
bossfi-backend/
├── cmd/                    # 应用程序入口
│   └── server/            # 服务器启动文件
├── internal/              # 内部应用代码
│   ├── api/              # API 路由和处理器
│   │   ├── v1/           # v1 版本 API handlers
│   │   ├── routes.go     # 路由配置
│   │   ├── response.go   # 响应处理
│   │   └── v1.go         # v1 路由注册
│   ├── domain/           # 业务实体
│   │   ├── user/         # 用户实体
│   │   ├── post/         # 帖子实体
│   │   └── stake/        # 质押实体
│   ├── repository/       # 数据访问层
│   └── service/          # 业务逻辑层
├── pkg/                  # 公共包
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── logger/          # 日志管理
│   ├── middleware/      # 中间件
│   └── mreturn/         # 统一响应格式
├── configs/             # 配置文件
├── migrations/          # 数据库迁移文件
└── scripts/             # 部署和工具脚本
```

## 📦 安装和运行

### 环境要求

- Go 1.21 或更高版本
- PostgreSQL 12 或更高版本
- Git

### 1. 克隆项目

```bash
git clone https://github.com/your-username/bossfi-backend.git
cd bossfi-backend
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库

#### 创建数据库和用户

```sql
-- 连接到 PostgreSQL
psql -U postgres

-- 创建数据库
CREATE DATABASE bossfi;

-- 创建用户
CREATE USER bossfier WITH PASSWORD 'your_password';

-- 授权
GRANT ALL PRIVILEGES ON DATABASE bossfi TO bossfier;
GRANT ALL ON SCHEMA public TO bossfier;
GRANT CREATE ON SCHEMA public TO bossfier;
```

#### 运行迁移

```bash
# 使用提供的脚本
cd scripts
./create_database.bat  # Windows
# 或
bash create_database.sh  # Linux/Mac
```

### 4. 配置文件

复制并编辑配置文件：

```bash
cp configs/config.toml.example configs/config.toml
```

编辑 `configs/config.toml`：

```toml
[server]
port = 8080
mode = "debug"
read_timeout = "60s"
write_timeout = "60s"

[database]
driver = "postgres"
host = "localhost"
port = 5432
database = "bossfi"
username = "bossfier"
password = "your_db_password"
sslmode = "disable"
timezone = "UTC"

[jwt]
secret = "your-super-secret-jwt-key-change-this-in-production"
expire_time = "24h"

[logger]
level = "info"
filename = "./logs/app.log"
max_size = 100
max_age = 30
max_backups = 5
compress = true
```

**⚠️ 重要安全提示：**
- 请务必修改 JWT 密钥为你自己的强密钥
- 生产环境中使用强密码
- 不要将包含敏感信息的配置文件提交到版本控制

### 5. 运行服务

#### 开发模式

```bash
go run ./cmd/server
```

#### 编译运行

```bash
# 编译
go build -o main ./cmd/server

# 运行
./main      # Linux/Mac
main.exe    # Windows (如果在Windows上编译)
```

服务默认运行在 `http://localhost:8080`

## 📚 API 文档

### 接口概览

#### 认证相关
- `POST /api/v1/auth/nonce` - 生成登录随机数
- `POST /api/v1/auth/login` - 钱包签名登录

#### 用户管理
- `GET /api/v1/users/profile` - 获取用户资料
- `PUT /api/v1/users/profile` - 更新用户资料
- `GET /api/v1/users/stats` - 获取用户统计
- `GET /api/v1/users/search` - 搜索用户

#### 帖子管理
- `GET /api/v1/posts` - 获取帖子列表
- `POST /api/v1/posts` - 创建帖子
- `GET /api/v1/posts/{id}` - 获取帖子详情
- `PUT /api/v1/posts/{id}` - 更新帖子
- `DELETE /api/v1/posts/{id}` - 删除帖子
- `POST /api/v1/posts/{id}/like` - 点赞帖子

#### 质押功能
- `POST /api/v1/stakes` - 创建质押
- `GET /api/v1/stakes/{id}` - 获取质押详情
- `POST /api/v1/stakes/{id}/unstake` - 请求解质押
- `POST /api/v1/stakes/rewards/claim` - 领取奖励

### Swagger 文档

启动服务后访问：`http://localhost:8080/swagger/index.html`

### 认证方式

#### 1. 获取 Nonce

```bash
curl -X POST http://localhost:8080/api/v1/auth/nonce \
  -H "Content-Type: application/json" \
  -d '{"wallet_address": "0x..."}'
```

#### 2. 钱包签名登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0x...",
    "signature": "0x...",
    "message": "nonce_message"
  }'
```

#### 3. 使用 JWT Token

```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer your_jwt_token"
```

## 🔧 开发指南

### 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 使用有意义的变量和函数名
- 添加必要的注释

### 项目特性

#### 1. 清洁架构
- **Domain**: 业务实体和规则
- **Repository**: 数据访问抽象
- **Service**: 业务逻辑实现
- **Handler**: HTTP 请求处理

#### 2. 中间件支持
- **认证中间件**: JWT token 验证
- **日志中间件**: 请求日志记录
- **CORS中间件**: 跨域支持
- **限流中间件**: API 访问限制
- **追踪中间件**: 请求链路追踪

#### 3. 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

#### 4. 错误处理
- 统一的错误响应格式
- 详细的错误日志记录
- 用户友好的错误信息

### 添加新功能

1. **添加实体**: 在 `internal/domain/` 下创建新的实体
2. **添加仓库**: 在 `internal/repository/` 下实现数据访问
3. **添加服务**: 在 `internal/service/` 下实现业务逻辑
4. **添加处理器**: 在 `internal/api/v1/` 下添加 HTTP 处理器
5. **注册路由**: 在 `internal/api/v1.go` 中注册新路由

## 🐳 Docker 部署

### 使用 Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 单独构建

```bash
# 构建镜像
docker build -t bossfi-backend .

# 运行容器
docker run -p 8080:8080 bossfi-backend
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/service/...

# 运行测试并显示覆盖率
go test -cover ./...
```

### 健康检查

```bash
curl http://localhost:8080/health
```

预期响应：
```json
{
  "status": "ok",
  "service": "bossfi-backend",
  "version": "1.0.0"
}
```

## 📝 数据库迁移

### 从 MySQL 迁移到 PostgreSQL

项目已完成从 MySQL 到 PostgreSQL 的迁移，详细迁移步骤：

1. 更新依赖包
2. 修改配置文件
3. 调整数据模型
4. 运行迁移脚本

具体迁移文档请参考项目中的迁移指南。

## 🔍 日志和监控

### 日志级别
- `debug`: 调试信息
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息

### 日志格式
支持 JSON 和文本两种格式，推荐生产环境使用 JSON 格式。

### 请求追踪
每个请求都会生成唯一的 TraceID，方便问题排查。

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系方式

- 项目链接: [https://github.com/locey/bossfi-backend](https://github.com/locey/bossfi-backend)
- 问题反馈: [GitHub Issues](https://github.com/locey/bossfi-backend/issues)

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

---

**BossFi Backend** - 让去中心化求职变得简单 🚀 