# bossfi-backend 项目 README

## 项目概述


## 根据官网命名规范建议

1. 文件名：全小写，单词间用下划线 如：config_loader.go
2. 函数/变量：驼峰式命名（CamelCase） 如：getUserInfo()、GetUserInfo()
3. 包名：推荐使用简洁、小写的单个单词命名，确实需要多个单词，可以如：package user_utils

参考：https://go.dev/doc/effective_go#file_names

## 目录结构说明

### src/
- **app/**: 应用核心代码
    - **model/**: 数据模型定义
    - **router/**: 路由定义
    - **service/**: 业务逻辑实现
    - **controller/**: 控制器层
- **core/**: 核心组件
    - **db/**: 数据库相关操作
        - `init.go`: 数据库初始化
        - `pgsql.go`: PostgreSQL 数据库操作
        - `redis.go`: Redis 操作
    - **gin/**: Gin 框架相关
        - **router/**: 路由封装
        - **middleware/**: 中间件
            - `recory.go`: 恢复中间件
            - `http_log.go`: HTTP 日志中间件
            - `language.go`: 语言中间件
    - **log/**: 日志处理
    - **app.go**: 应用启动入口
    - **config/**: 配置管理
    - **result/**: 统一响应格式
- **common/**: 公共组件
    - `context.go`: 上下文相关工具

### config/
- `config.toml`: 主配置文件
- `config.toml.example`: 配置示例文件

### go.mod & go.sum
- Go 模块依赖管理文件

## 核心功能

1. **多语言支持**:
    - 通过 `middleware/language.go` 实现语言中间件
    - 支持从请求头或 URL 参数获取语言标识

2. **统一响应格式**:
    - 定义在 `src/core/result/result.go` 中
    - 包含状态码、消息和数据字段

3. **错误处理**:
    - 预定义了多种错误码和对应的多语言消息

4. **数据库访问**:
    - 支持 PostgreSQL 和 Redis

## 快速开始

1. 克隆项目
2. 复制 `config.toml.example` 为 `config.toml` 并修改配置
3. 运行 `go mod tidy` 安装依赖
4. 运行 `go run main.go` 启动服务

## API 文档(后续增加swagger)
GET /api/v1/test
GET /api/v1/demo/:id
POST /api/v1/create
PUT /api/v1/update
DELETE /api/v1/delete/:id
GET /api/v1/list
GET /api/v1/page?page=1&page_size=10

