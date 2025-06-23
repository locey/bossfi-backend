# BossFi Blockchain Backend

基于DDD（领域驱动设计）架构的区块链后端服务，使用Gin + GORM MySQL + Redis技术栈构建。

## 📋 目录

- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [快速开始](#快速开始)
- [API文档](#api文档)
- [配置说明](#配置说明)
- [部署指南](#部署指南)
- [开发指南](#开发指南)
- [贡献指南](#贡献指南)

## 🚀 功能特性

- ✅ **用户管理**：注册、登录、资料管理、权限控制
- ✅ **钱包管理**：多链钱包创建、余额管理、状态控制
- ✅ **交易管理**：交易记录、状态跟踪、确认机制
- ✅ **安全机制**：JWT认证、密码加密、限流保护
- ✅ **管理后台**：用户管理、钱包管理、系统监控
- ✅ **多链支持**：Bitcoin、Ethereum、BSC、Polygon、TRON
- ✅ **监控日志**：结构化日志、性能监控、错误追踪
- ✅ **生产就绪**：Docker部署、优雅关闭、健康检查

## 🛠 技术栈

### 后端技术
- **框架**：Gin Web Framework
- **数据库**：MySQL 8.0 + GORM ORM
- **缓存**：Redis
- **认证**：JWT (JSON Web Token)
- **文档**：Swagger/OpenAPI
- **日志**：Zap + Lumberjack
- **配置**：Viper

### 区块链集成
- **Bitcoin**：btcd/btcutil
- **Ethereum**：go-ethereum
- **多链支持**：统一接口设计

### 部署运维
- **容器化**：Docker + Docker Compose
- **反向代理**：Nginx
- **监控**：结构化日志 + 健康检查

## 📁 项目结构

```
backend/
├── cmd/                    # 应用程序入口
│   └── server/
│       └── main.go        # 主服务器
├── internal/              # 内部包（不对外暴露）
│   ├── api/              # API层（控制器）
│   │   ├── routes.go     # 路由配置
│   │   ├── response.go   # 统一响应
│   │   ├── user_handler.go
│   │   └── wallet_handler.go
│   ├── service/          # 业务逻辑层
│   │   ├── user_service.go
│   │   └── wallet_service.go
│   ├── repository/       # 数据访问层
│   │   ├── user_repository.go
│   │   ├── wallet_repository.go
│   │   └── transaction_repository.go
│   └── domain/           # 领域模型层
│       ├── user/         # 用户领域
│       ├── wallet/       # 钱包领域
│       └── transaction/  # 交易领域
├── pkg/                  # 公共包
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── redis/           # Redis连接
│   ├── logger/          # 日志管理
│   └── middleware/      # 中间件
├── configs/             # 配置文件
│   └── config.yaml
├── migrations/          # 数据库迁移
│   └── 001_init.sql
├── docs/               # 文档目录
├── logs/               # 日志目录
├── go.mod              # Go模块文件
├── go.sum              # 依赖锁定文件
├── Dockerfile          # Docker构建文件
├── docker-compose.yml  # Docker编排文件
├── Makefile           # 构建脚本
└── README.md          # 项目说明
```

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 本地开发

1. **克隆项目**
```bash
git clone <repository-url>
cd backend
```

2. **安装依赖**
```bash
make deps
```

3. **配置环境**
```bash
# 复制配置文件
cp configs/config.yaml.example configs/config.yaml

# 编辑配置文件，修改数据库和Redis连接信息
vim configs/config.yaml
```

4. **初始化数据库**
```bash
# 创建数据库并执行迁移
make migrate
```

5. **启动服务**
```bash
# 开发模式（热重载）
make dev

# 或者直接运行
make run
```

6. **访问服务**
- API服务：http://localhost:8080
- API文档：http://localhost:8080/swagger/index.html
- 健康检查：http://localhost:8080/health

### Docker部署

1. **使用Docker Compose**
```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app
```

2. **单独构建**
```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run
```

## 📖 API文档

### 认证接口
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录

### 用户接口
- `GET /api/v1/user/profile` - 获取用户资料
- `PUT /api/v1/user/profile` - 更新用户资料
- `PUT /api/v1/user/password` - 修改密码

### 钱包接口
- `POST /api/v1/wallets` - 创建钱包
- `GET /api/v1/wallets/my` - 获取我的钱包
- `GET /api/v1/wallets/{id}` - 获取钱包详情
- `GET /api/v1/wallets/address/{address}` - 根据地址获取钱包

### 管理员接口
- `GET /api/v1/admin/users` - 用户列表
- `DELETE /api/v1/admin/users/{id}` - 删除用户
- `GET /api/v1/admin/wallets` - 钱包列表
- `PUT /api/v1/admin/wallets/{id}/freeze` - 冻结钱包

完整API文档请访问：http://localhost:8080/swagger/index.html

## ⚙️ 配置说明

主要配置项：

```yaml
# 服务器配置
server:
  port: 8080              # 服务端口
  mode: debug             # 运行模式：debug/release/test

# 数据库配置
database:
  host: localhost         # 数据库主机
  port: 3306             # 数据库端口
  database: bossfi_blockchain
  username: root
  password: root

# Redis配置
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

# JWT配置
jwt:
  secret: "your-secret-key"
  expire_time: 24h

# 安全配置
security:
  rate_limit: 100         # 每分钟请求限制
  cors_origins: ["*"]     # CORS允许的源
```

## 🚢 部署指南

### 生产环境部署

1. **环境准备**
```bash
# 服务器环境
- CentOS 7+ / Ubuntu 18+
- Docker 20+
- Docker Compose 1.25+
```

2. **配置文件**
```bash
# 生产环境配置
server:
  mode: release
  port: 8080

database:
  host: mysql
  port: 3306
  # 使用环境变量或安全的配置管理

security:
  cors_origins: ["https://yourdomain.com"]
  rate_limit: 1000
```

3. **SSL配置**
```bash
# 配置SSL证书
mkdir ssl
# 放置SSL证书文件
```

4. **启动服务**
```bash
# 生产环境启动
docker-compose -f docker-compose.prod.yml up -d
```

### 监控与维护

- **日志监控**：logs/app.log
- **性能监控**：/health端点
- **数据备份**：定期备份MySQL数据
- **安全更新**：定期更新依赖包

## 👨‍💻 开发指南

### 开发规范

1. **代码结构**
   - 遵循DDD架构模式
   - 保持层次清晰分离
   - 使用依赖注入

2. **命名规范**
   - 包名：小写，简短
   - 接口：以er结尾
   - 常量：大写，下划线分隔

3. **错误处理**
   - 使用自定义错误类型
   - 统一错误响应格式
   - 记录详细错误日志

### 测试

```bash
# 运行测试
make test

# 运行基准测试
go test -bench=. ./...

# 生成测试覆盖率
go test -cover ./...
```

### 新增功能

1. **添加新的领域模型**
   - 在`internal/domain`下创建新包
   - 定义实体、值对象、聚合根
   - 定义领域错误

2. **添加新的API**
   - 在`internal/api`下添加处理器
   - 在`routes.go`中注册路由
   - 添加Swagger注释

## 🤝 贡献指南

1. Fork项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开Pull Request

## 📄 许可证

本项目采用MIT许可证。详见[LICENSE](LICENSE)文件。

## 📞 联系方式

- 项目维护者：[Your Name]
- 邮箱：your.email@example.com
- 项目地址：[GitHub Repository]

---

⭐ 如果这个项目对你有帮助，请给个Star！ 