# Makefile for BossFi Backend

# 变量定义
APP_NAME := bossfi-backend
BINARY_NAME := $(APP_NAME)
DOCKER_IMAGE := $(APP_NAME):latest
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | awk '{print $$3}')

# Go 相关
GO_VERSION := 1.21
GOFLAGS := -ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# 构建标志
BUILD_FLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# Go 相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint

# 目录变量
BINARY_DIR := bin
COVERAGE_DIR := coverage

# 默认目标
.PHONY: help
help: ## 显示帮助信息
	@echo "BossFi Backend 构建工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发相关
.PHONY: install
install: ## 安装依赖
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: run
run: swagger-generate ## 运行应用（自动生成swagger文档）
	$(GOCMD) run ./api

.PHONY: build
build: ## 构建二进制文件
	@echo "构建 $(APP_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(APP_NAME) ./api
	@echo "构建完成: $(BINARY_DIR)/$(APP_NAME)"

# Swagger 相关
.PHONY: swagger-install
swagger-install: ## 安装 swag 工具
	$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest

.PHONY: swagger-generate
swagger-generate: ## 生成 Swagger 文档
	@echo "生成 Swagger 文档..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init --dir ./api --generalInfo main.go --output ./docs; \
	else \
		echo "swag 未安装，正在安装..."; \
		$(MAKE) swagger-install; \
		swag init --dir ./api --generalInfo main.go --output ./docs; \
	fi

.PHONY: swagger-serve
swagger-serve: swagger-generate ## 启动 Swagger UI 服务器
	@echo "启动 Swagger UI 服务器..."
	@echo "打开 http://localhost:8080/swagger/index.html 查看 API 文档"
	$(MAKE) run

.PHONY: test
test: test-unit test-integration ## 运行所有测试

.PHONY: test-unit
test-unit: ## 运行单元测试
	@echo "运行单元测试..."
	$(GOTEST) -v -race -short ./...

.PHONY: test-integration
test-integration: ## 运行集成测试
	@echo "运行集成测试..."
	$(GOTEST) -v -race ./test/...

.PHONY: test-cover
test-cover: ## 运行测试并生成覆盖率报告
	@echo "生成测试覆盖率报告..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "覆盖率报告已生成: $(COVERAGE_DIR)/coverage.html"

.PHONY: test-category
test-category: ## 测试分类功能
	@echo "测试分类功能..."
	@./scripts/test_category.sh

.PHONY: test-user-comments
test-user-comments: ## 测试用户评论功能
	@echo "测试用户评论功能..."
	@go test -v ./test/user_comments_test.go ./test/common_test.go

.PHONY: test-user-comments-e2e
test-user-comments-e2e: ## 测试用户评论功能端到端
	@echo "测试用户评论功能端到端..."
	@./scripts/test_user_comments.sh

.PHONY: test-watch
test-watch: ## 监视文件变化并自动运行测试
	@if command -v gotestsum >/dev/null 2>&1; then \
		echo "监视测试（需要安装 gotestsum）..."; \
		gotestsum --watch -- -race -short ./...; \
	else \
		echo "请安装 gotestsum: go install gotest.tools/gotestsum@latest"; \
	fi

.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GOFMT) -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "建议安装 goimports: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

.PHONY: vet
vet: ## 运行 go vet
	go vet ./...

.PHONY: lint
lint: ## 运行 golangci-lint
	@echo "运行代码检查..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run; \
	else \
		echo "请安装 golangci-lint: https://golangci-lint.run/usage/install/"; \
		echo "或使用: go vet ./..."; \
		go vet ./...; \
	fi

.PHONY: clean
clean: ## 清理构建文件
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -f api.exe

# Docker 相关
.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -t $(DOCKER_IMAGE) .
	docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE)

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	@echo "运行 Docker 容器..."
	docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

.PHONY: docker-stop
docker-stop: ## 停止 Docker 容器
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

.PHONY: docker-logs
docker-logs: ## 查看 Docker 容器日志
	docker logs -f $(APP_NAME)

# 数据库相关
.PHONY: db-init
db-init: ## 初始化数据库
	createdb bossfi || true
	psql -h localhost -U postgres -d bossfi -f deploy/init.sql
	psql -h localhost -U postgres -d bossfi -f deploy/category_data.sql

.PHONY: db-drop
db-drop: ## 删除数据库
	dropdb bossfi || true

.PHONY: db-reset
db-reset: db-drop db-init ## 重置数据库

.PHONY: db-migrate
db-migrate: ## 运行数据库迁移
	psql -h localhost -U postgres -d bossfi -f deploy/init.sql
	psql -h localhost -U postgres -d bossfi -f deploy/category_data.sql

# 部署相关
.PHONY: deploy-dev
deploy-dev: swagger-generate ## 部署到开发环境
	@echo "Deploying to development environment..."
	$(MAKE) docker-stop
	$(MAKE) docker-build
	$(MAKE) docker-run

.PHONY: deploy-prod
deploy-prod: swagger-generate ## 部署到生产环境
	@echo "Deploying to production environment..."
	@echo "Make sure to set production environment variables!"
	$(MAKE) build
	# 这里添加生产环境部署逻辑

# 工具相关
.PHONY: setup-dev
setup-dev: ## 设置开发环境
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "复制环境变量模板..."; \
		cp env.example .env; \
		echo "请编辑 .env 文件配置您的环境"; \
	fi
	$(MAKE) install
	$(MAKE) swagger-install

.PHONY: check
check: fmt vet test ## 运行所有检查

.PHONY: all
all: clean swagger-generate build test ## 构建和测试

# 版本相关
.PHONY: version
version: ## 显示版本信息
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"

# 生产部署相关
build-linux: ## 构建 Linux 版本
	@echo "构建 Linux 版本..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(APP_NAME)-linux ./api

build-windows: ## 构建 Windows 版本
	@echo "构建 Windows 版本..."
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DIR)/$(APP_NAME)-windows.exe ./api

build-all: build build-linux build-windows ## 构建所有平台版本

# 安全和质量检查
security-check: ## 运行安全检查
	@if command -v gosec >/dev/null 2>&1; then \
		echo "运行安全检查..."; \
		gosec ./...; \
	else \
		echo "请安装 gosec: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

benchmark: ## 运行性能测试
	@echo "运行性能测试..."
	$(GOTEST) -bench=. -benchmem ./...

# 显示项目信息
info: ## 显示项目信息
	@echo "项目信息:"
	@echo "  名称: $(APP_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  Go版本: $(GO_VERSION)"
	@echo "  构建时间: $(BUILD_TIME)" 