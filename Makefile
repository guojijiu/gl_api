# 云平台API Makefile
# 提供常用的开发、测试、构建和部署命令

.PHONY: help build test clean run docker-build docker-run deploy

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := cloud-platform-api
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD)
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 颜色定义
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# 帮助信息
help: ## 显示帮助信息
	@echo "$(BLUE)云平台API - 可用命令:$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

# 开发相关命令
dev: ## 启动开发环境
	@echo "$(BLUE)启动开发环境...$(NC)"
	@go run main.go

dev-watch: ## 启动开发环境（文件监控）
	@echo "$(BLUE)启动开发环境（文件监控）...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)air未安装，使用go run...$(NC)"; \
		go run main.go; \
	fi

# 构建相关命令
build: ## 构建应用
	@echo "$(BLUE)构建应用...$(NC)"
	@go build $(LDFLAGS) -o bin/$(APP_NAME) main.go
	@echo "$(GREEN)构建完成: bin/$(APP_NAME)$(NC)"

build-linux: ## 构建Linux版本
	@echo "$(BLUE)构建Linux版本...$(NC)"
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux main.go
	@echo "$(GREEN)构建完成: bin/$(APP_NAME)-linux$(NC)"

build-windows: ## 构建Windows版本
	@echo "$(BLUE)构建Windows版本...$(NC)"
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows.exe main.go
	@echo "$(GREEN)构建完成: bin/$(APP_NAME)-windows.exe$(NC)"

build-all: build build-linux build-windows ## 构建所有平台版本

# 测试相关命令
test: ## 运行单元测试
	@echo "$(BLUE)运行单元测试...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "$(BLUE)运行测试并生成覆盖率报告...$(NC)"
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)覆盖率报告已生成: coverage.html$(NC)"

test-integration: ## 运行集成测试
	@echo "$(BLUE)运行集成测试...$(NC)"
	@go test -v -tags=integration ./tests/Integration/...

test-benchmark: ## 运行基准测试
	@echo "$(BLUE)运行基准测试...$(NC)"
	@go test -bench=. -benchmem ./tests/benchmark/...

test-all: test test-integration test-benchmark ## 运行所有测试

# 代码质量相关命令
lint: ## 运行代码检查
	@echo "$(BLUE)运行代码检查...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint未安装，跳过代码检查$(NC)"; \
	fi

fmt: ## 格式化代码
	@echo "$(BLUE)格式化代码...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)代码格式化完成$(NC)"

vet: ## 运行go vet
	@echo "$(BLUE)运行go vet...$(NC)"
	@go vet ./...

# 依赖管理
deps: ## 安装依赖
	@echo "$(BLUE)安装依赖...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)依赖安装完成$(NC)"

deps-update: ## 更新依赖
	@echo "$(BLUE)更新依赖...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)依赖更新完成$(NC)"

# Docker相关命令
docker-build: ## 构建Docker镜像
	@echo "$(BLUE)构建Docker镜像...$(NC)"
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "$(GREEN)Docker镜像构建完成: $(APP_NAME):$(VERSION)$(NC)"

docker-run: ## 运行Docker容器
	@echo "$(BLUE)运行Docker容器...$(NC)"
	@docker run -d --name $(APP_NAME) -p 8080:8080 $(APP_NAME):latest
	@echo "$(GREEN)Docker容器已启动$(NC)"

docker-stop: ## 停止Docker容器
	@echo "$(BLUE)停止Docker容器...$(NC)"
	@docker stop $(APP_NAME) || true
	@docker rm $(APP_NAME) || true
	@echo "$(GREEN)Docker容器已停止$(NC)"

# 部署相关命令
deploy-docker: ## 使用Docker部署
	@echo "$(BLUE)使用Docker部署...$(NC)"
	@chmod +x scripts/deploy_docker.sh
	@./scripts/deploy_docker.sh deploy

deploy-k8s: ## 使用Kubernetes部署
	@echo "$(BLUE)使用Kubernetes部署...$(NC)"
	@chmod +x scripts/deploy_k8s.sh
	@./scripts/deploy_k8s.sh deploy

# 监控和日志相关命令
logs: ## 查看应用日志
	@echo "$(BLUE)查看应用日志...$(NC)"
	@if docker ps -q -f name=$(APP_NAME) | grep -q .; then \
		docker logs -f $(APP_NAME); \
	else \
		echo "$(YELLOW)Docker容器未运行$(NC)"; \
	fi

monitor: ## 启动监控
	@echo "$(BLUE)启动监控...$(NC)"
	@if command -v htop > /dev/null; then \
		htop; \
	else \
		echo "$(YELLOW)htop未安装，使用top...$(NC)"; \
		top; \
	fi

# 数据库相关命令
db-migrate: ## 运行数据库迁移
	@echo "$(BLUE)运行数据库迁移...$(NC)"
	@go run scripts/migrate.go up

db-rollback: ## 回滚数据库迁移
	@echo "$(BLUE)回滚数据库迁移...$(NC)"
	@go run scripts/migrate.go down

db-seed: ## 填充测试数据
	@echo "$(BLUE)填充测试数据...$(NC)"
	@go run scripts/seed.go

# 清理相关命令
clean: ## 清理构建文件
	@echo "$(BLUE)清理构建文件...$(NC)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -rf test-results/
	@rm -rf benchmark-results/
	@echo "$(GREEN)清理完成$(NC)"

clean-docker: ## 清理Docker资源
	@echo "$(BLUE)清理Docker资源...$(NC)"
	@docker stop $(APP_NAME) || true
	@docker rm $(APP_NAME) || true
	@docker rmi $(APP_NAME):$(VERSION) || true
	@docker rmi $(APP_NAME):latest || true
	@echo "$(GREEN)Docker资源清理完成$(NC)"

# 工具安装
install-tools: ## 安装开发工具
	@echo "$(BLUE)安装开发工具...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/rakyll/hey@latest
	@go install github.com/google/pprof@latest
	@echo "$(GREEN)开发工具安装完成$(NC)"

# 性能测试
benchmark: ## 运行性能测试
	@echo "$(BLUE)运行性能测试...$(NC)"
	@chmod +x scripts/run_benchmarks.sh
	@./scripts/run_benchmarks.sh all

# 安全扫描
security-scan: ## 运行安全扫描
	@echo "$(BLUE)运行安全扫描...$(NC)"
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)gosec未安装，跳过安全扫描$(NC)"; \
	fi

# 文档生成
docs: ## 生成API文档
	@echo "$(BLUE)生成API文档...$(NC)"
	@if command -v swag > /dev/null; then \
		swag init -g main.go -o docs/swagger; \
	else \
		echo "$(YELLOW)swag未安装，跳过文档生成$(NC)"; \
	fi

# 版本信息
version: ## 显示版本信息
	@echo "$(BLUE)版本信息:$(NC)"
	@echo "  应用名称: $(APP_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git提交: $(GIT_COMMIT)"
	@echo "  Go版本: $(shell go version)"

# 环境检查
check-env: ## 检查开发环境
	@echo "$(BLUE)检查开发环境...$(NC)"
	@echo "Go版本: $(shell go version)"
	@echo "Docker版本: $(shell docker --version 2>/dev/null || echo '未安装')"
	@echo "Kubectl版本: $(shell kubectl version --client 2>/dev/null || echo '未安装')"
	@echo "Git版本: $(shell git --version)"

# 快速启动
quick-start: deps build test ## 快速启动（安装依赖、构建、测试）
	@echo "$(GREEN)快速启动完成！$(NC)"
	@echo "运行 'make run' 启动应用"

# 生产环境构建
build-prod: ## 构建生产环境版本
	@echo "$(BLUE)构建生产环境版本...$(NC)"
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo $(LDFLAGS) -o bin/$(APP_NAME)-prod main.go
	@echo "$(GREEN)生产环境版本构建完成: bin/$(APP_NAME)-prod$(NC)"

# 健康检查
health: ## 检查应用健康状态
	@echo "$(BLUE)检查应用健康状态...$(NC)"
	@curl -s http://localhost:8080/api/v1/health | jq . || echo "$(YELLOW)应用未运行或jq未安装$(NC)"

# 统计信息
stats: ## 显示项目统计信息
	@echo "$(BLUE)项目统计信息:$(NC)"
	@echo "  代码行数: $(shell find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1)"
	@echo "  Go文件数: $(shell find . -name '*.go' -not -path './vendor/*' | wc -l)"
	@echo "  测试文件数: $(shell find . -name '*_test.go' -not -path './vendor/*' | wc -l)"
	@echo "  文档文件数: $(shell find . -name '*.md' | wc -l)"