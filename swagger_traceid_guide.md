# 🚀 Swagger 文档和 TraceID 日志追踪功能指南

本指南介绍 BossFi Backend 项目中新增的 Swagger API 文档和 TraceID 日志追踪功能。

## 📊 Swagger API 文档

### 功能特性

1. **自动生成文档**: 项目启动时自动更新 Swagger 文档
2. **交互式 UI**: 提供友好的 Web 界面测试 API
3. **完整的 API 描述**: 包含请求/响应结构、参数说明、错误码等
4. **JWT 认证支持**: 集成 Bearer Token 认证测试

### 访问方式

启动项目后，访问以下地址查看 API 文档：

```
http://localhost:8080/swagger/index.html
```

### 生成和更新

```bash
# 手动生成 Swagger 文档
make swagger-generate

# 启动项目（自动生成文档）
make run

# 或者直接启动 Swagger 服务
make swagger-serve
```

### API 端点说明

#### 1. 健康检查
- **GET** `/health`
- 检查服务器运行状态
- 无需认证

#### 2. 认证相关
- **POST** `/api/v1/auth/nonce` - 获取签名消息和 nonce
- **POST** `/api/v1/auth/login` - 钱包签名登录
- **GET** `/api/v1/auth/profile` - 获取用户信息（需认证）
- **POST** `/api/v1/auth/logout` - 用户登出（需认证）

#### 3. 用户相关
- **GET** `/api/v1/users/me` - 获取当前用户信息（需认证）

#### 4. 区块链相关
- **GET** `/api/v1/blockchain/balance/{address}` - 获取地址余额（需认证）

#### 5. 管理员相关
- **GET** `/api/v1/admin/stats` - 获取系统统计（需认证）

## 🔍 TraceID 日志追踪

### 功能特性

1. **自动生成 TraceID**: 每个请求自动分配唯一追踪ID
2. **前端传递支持**: 支持前端通过请求头传递 TraceID
3. **全链路追踪**: 所有相关日志都包含相同的 TraceID
4. **结构化日志**: 使用 JSON 格式输出结构化日志

### TraceID 机制

#### 自动生成
如果前端没有提供 TraceID，系统会自动生成一个 UUID：

```
X-Trace-ID: 550e8400-e29b-41d4-a716-446655440000
```

#### 前端传递
前端可以通过请求头传递自定义 TraceID：

```javascript
fetch('/api/v1/auth/profile', {
  headers: {
    'Authorization': 'Bearer your-jwt-token',
    'X-Trace-ID': 'your-custom-trace-id'
  }
})
```

#### 响应头返回
服务器会在响应头中返回 TraceID：

```
X-Trace-ID: 550e8400-e29b-41d4-a716-446655440000
```

### 日志格式

所有日志都包含 TraceID，格式如下：

```json
{
  "timestamp": "2025-06-26T15:50:00+08:00",
  "level": "info",
  "message": "User logged in successfully",
  "caller": "controllers/auth_controller.go:120",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "wallet_address": "0x1234567890123456789012345678901234567890"
}
```

### 使用示例

#### 1. 前端请求示例

```javascript
// React/Vue 示例
const apiCall = async (url, options = {}) => {
  const traceId = generateTraceId(); // 可选：生成自定义 TraceID
  
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'X-Trace-ID': traceId,
      ...options.headers
    }
  });
  
  // 获取服务器返回的 TraceID
  const serverTraceId = response.headers.get('X-Trace-ID');
  console.log('Trace ID:', serverTraceId);
  
  return response;
};

// 登录示例
const login = async (walletAddress, signature, message) => {
  return apiCall('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({
      wallet_address: walletAddress,
      signature: signature,
      message: message
    })
  });
};
```

#### 2. 日志查询示例

```bash
# 查询特定 TraceID 的所有日志
grep "550e8400-e29b-41d4-a716-446655440000" app.log

# 使用 jq 处理 JSON 日志
cat app.log | jq 'select(.trace_id == "550e8400-e29b-41d4-a716-446655440000")'

# 按时间和 TraceID 排序
cat app.log | jq -s 'sort_by(.timestamp) | .[] | select(.trace_id == "your-trace-id")'
```

## 🛠️ 开发者使用指南

### 1. 启动开发环境

```bash
# 设置开发环境
make setup-dev

# 生成 Swagger 文档并启动服务
make run

# 或者单独启动 Swagger UI
make swagger-serve
```

### 2. 测试 API

1. 访问 Swagger UI: http://localhost:8080/swagger/index.html
2. 测试健康检查端点
3. 获取 nonce 进行钱包登录测试
4. 使用返回的 JWT token 测试受保护的端点

### 3. 查看日志

```bash
# 查看实时日志
tail -f logs/app.log

# 过滤特定级别的日志
grep '"level":"error"' logs/app.log

# 查看特定用户的操作日志
grep '"user_id":"your-user-id"' logs/app.log
```

### 4. 调试 TraceID

```bash
# 发送带有自定义 TraceID 的请求
curl -H "X-Trace-ID: debug-trace-001" \
     -H "Content-Type: application/json" \
     http://localhost:8080/health

# 查看该 TraceID 的所有日志
grep "debug-trace-001" logs/app.log
```

## 📋 CORS 配置

项目已配置支持 TraceID 的 CORS 设置：

```go
config.AllowHeaders = []string{
    "Origin", 
    "Content-Type", 
    "Accept", 
    "Authorization", 
    "X-Trace-ID"
}
config.ExposeHeaders = []string{"X-Trace-ID"}
```

## 🎯 最佳实践

### 1. TraceID 使用

- **前端**: 为每个用户会话生成一个基础 TraceID，每个请求可以添加序号
- **移动端**: 可以结合设备ID和时间戳生成 TraceID
- **调试**: 使用有意义的 TraceID 便于问题排查

### 2. 日志查询

- 使用 ELK Stack 或类似工具进行日志聚合和分析
- 为生产环境配置日志轮转和压缩
- 建立日志告警机制

### 3. API 文档维护

- 及时更新 Swagger 注释
- 为新的 API 端点添加完整的文档
- 定期检查文档的准确性

## 🔗 相关链接

- [Swagger UI](http://localhost:8080/swagger/index.html)
- [Gin Swagger 文档](https://github.com/swaggo/gin-swagger)
- [Swag 注释指南](https://github.com/swaggo/swag)
- [Logrus 文档](https://github.com/sirupsen/logrus)

## 📞 支持

如有问题，请：

1. 查看 Swagger 文档中的 API 说明
2. 检查日志中的 TraceID 追踪信息
3. 使用 `make help` 查看可用命令
4. 联系开发团队获取支持 