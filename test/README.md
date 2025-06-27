# 测试文档

本文档介绍 BossFi Backend 项目的测试结构和运行方法。

## 📁 测试文件结构

```
bossfi-backend/
├── test/                           # 集成测试
│   ├── setup_test.go              # 测试环境设置
│   ├── integration_test.go        # 集成测试
│   └── README.md                  # 本文档
├── utils/
│   └── crypto_test.go             # 加密工具单元测试
├── api/
│   ├── controllers/
│   │   └── auth_controller_test.go # 控制器测试
│   └── services/
│       └── user_service_test.go    # 服务层测试
├── middleware/
│   └── auth_test.go               # 中间件测试
└── config/
    └── config_test.go             # 配置测试
```

## 🧪 测试类型

### 1. 单元测试
测试单个函数或方法的功能，位于各个包的 `*_test.go` 文件中。

- **utils/crypto_test.go**: 测试加密、JWT、签名验证等功能
- **config/config_test.go**: 测试配置加载和验证
- **middleware/auth_test.go**: 测试认证中间件
- **api/services/user_service_test.go**: 测试用户服务业务逻辑
- **api/controllers/auth_controller_test.go**: 测试控制器HTTP处理

### 2. 集成测试
测试多个组件之间的交互，位于 `test/` 目录中。

- **integration_test.go**: 测试完整的API流程和组件集成

## 🚀 运行测试

### 环境准备

1. **安装依赖**
   ```bash
   make deps
   ```

2. **设置测试环境**
   ```bash
   # 复制环境变量文件
   cp env.example .env
   
   # 编辑 .env 文件，配置测试数据库和Redis
   # 建议使用独立的测试数据库避免影响开发数据
   ```

3. **启动依赖服务**
   ```bash
   # 使用 Docker Compose 启动 PostgreSQL 和 Redis
   docker-compose up -d postgres redis
   ```

### 运行测试命令

```bash
# 运行所有测试
make test

# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 生成覆盖率报告
make test-coverage

# 监视模式（自动运行测试）
make test-watch
```

### 详细的 Go 命令

```bash
# 运行所有测试
go test -v ./...

# 运行特定包的测试
go test -v ./utils/
go test -v ./api/services/

# 运行特定测试函数
go test -v -run TestGenerateNonce ./utils/

# 运行测试并显示详细输出
go test -v -race ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 📊 测试覆盖率

生成的覆盖率报告会保存在 `coverage/` 目录中：

- `coverage.out`: 覆盖率数据文件
- `coverage.html`: HTML格式的覆盖率报告

## 🔧 测试配置

### 环境变量

测试使用以下环境变量（会覆盖默认配置）：

```env
# 测试数据库
TEST_DB_NAME=bossfi_test
TEST_REDIS_DB=1
TEST_JWT_SECRET=test-secret-key

# 其他测试配置
GIN_MODE=test
LOG_LEVEL=error
CRON_ENABLED=false
```

### 测试数据库

建议为测试使用独立的数据库：

1. 创建测试数据库：
   ```sql
   CREATE DATABASE bossfi_test;
   ```

2. 测试会自动：
   - 在每个测试套件开始前初始化数据库
   - 在每个测试用例前清理数据
   - 在测试套件结束后清理环境

## 📝 编写测试

### 单元测试示例

```go
func TestMyFunction(t *testing.T) {
    // 准备测试数据
    input := "test-input"
    expected := "expected-output"
    
    // 执行被测试的函数
    result := MyFunction(input)
    
    // 断言结果
    assert.Equal(t, expected, result)
}
```

### 使用测试套件

```go
type MyTestSuite struct {
    suite.Suite
    // 测试用的字段
}

func (suite *MyTestSuite) SetupSuite() {
    // 在整个测试套件开始前执行
}

func (suite *MyTestSuite) SetupTest() {
    // 在每个测试用例前执行
}

func (suite *MyTestSuite) TestSomething() {
    // 测试用例
}

func TestMyTestSuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

## 🐛 调试测试

### 查看测试输出

```bash
# 详细输出
go test -v ./...

# 显示测试运行时间
go test -v -timeout 30s ./...

# 只运行失败的测试
go test -v -count=1 ./...
```

### 测试特定场景

```bash
# 测试特定函数
go test -run TestGenerateNonce

# 测试特定包
go test ./utils/

# 跳过集成测试
go test -short ./...
```

## 📈 持续集成

在CI/CD流水线中运行测试：

```yaml
# GitHub Actions 示例
- name: Run tests
  run: |
    make test-unit
    make test-integration
    make test-coverage
```

## 🔍 测试最佳实践

1. **测试命名**: 使用描述性的测试名称
2. **测试隔离**: 每个测试应该独立运行
3. **数据清理**: 测试前后清理测试数据
4. **模拟依赖**: 使用mock对象隔离外部依赖
5. **边界测试**: 测试边界条件和错误情况
6. **覆盖率目标**: 保持80%以上的代码覆盖率

## 🚨 注意事项

1. **数据库**: 测试会清空测试数据库，请勿使用生产数据库
2. **Redis**: 测试会清空指定的Redis数据库
3. **并发**: 某些测试可能需要串行运行以避免竞态条件
4. **外部依赖**: 区块链相关测试可能依赖外部服务

## 📚 相关资源

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Framework](https://github.com/stretchr/testify)
- [Gin Testing Guide](https://gin-gonic.com/docs/testing/)
- [GORM Testing](https://gorm.io/docs/testing.html) 