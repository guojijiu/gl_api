# Cloud Platform API Makefile
# 功能说明：
# 1. 提供项目构建和部署命令
# 2. 支持开发、测试、生产环境
# 3. 包含代码质量检查和格式化
# 4. 提供数据库迁移和种子数据命令

.PHONY: help build dev test clean install deps lint format migrate seed docker-build docker-run docs

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := cloud-platform-api
BUILD_DIR := build
MAIN_FILE := main.go
DOCKER_IMAGE := cloud-platform-api
DOCKER_TAG := latest

# 帮助信息
help: ## 显示帮助信息
	@echo "Cloud Platform API 构建工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 安装依赖
install: ## 安装项目依赖
	@echo "安装Go依赖..."
	go mod download
	go mod tidy
	go mod verify

# 构建应用
build: ## 构建应用
	@echo "构建应用..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)"

# 开发模式运行
dev: ## 开发模式运行（热重载）
	@echo "启动开发模式..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air未安装，使用go run..."; \
		go run $(MAIN_FILE); \
	fi

# 生产模式运行
run: build ## 生产模式运行
	@echo "启动生产模式..."
	./$(BUILD_DIR)/$(APP_NAME)

# 测试
test: ## 运行测试
	@echo "运行测试..."
	go test -v ./...
	@echo "测试完成"

# 测试覆盖率
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试覆盖率..."
	@mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "覆盖率报告生成完成: coverage/coverage.html"

# 基准测试
bench: ## 运行基准测试
	@echo "运行基准测试..."
	go test -bench=. -benchmem ./...

# 代码检查
lint: ## 运行代码检查
	@echo "运行代码检查..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint未安装，跳过代码检查"; \
	fi

# 代码格式化
format: ## 格式化代码
	@echo "格式化代码..."
	go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "goimports未安装，跳过导入格式化"; \
	fi

# 代码质量检查
quality: lint format ## 运行代码质量检查
	@echo "代码质量检查完成"

# 数据库迁移
migrate: ## 运行数据库迁移
	@echo "运行数据库迁移..."
	go run scripts/migrate.go

# 数据库种子数据
seed: ## 填充数据库种子数据
	@echo "填充种子数据..."
	go run scripts/seed.go

# 数据库重置
db-reset: ## 重置数据库（删除所有表并重新迁移）
	@echo "重置数据库..."
	@read -p "确定要删除所有数据吗？(y/N): " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		go run scripts/migrate.go reset; \
		go run scripts/migrate.go; \
		echo "数据库重置完成"; \
	else \
		echo "操作已取消"; \
	fi

# 生成API文档
docs: ## 生成API文档
	@echo "生成API文档..."
	@if command -v swag > /dev/null; then \
		swag init -g $(MAIN_FILE) -o docs/swagger; \
		echo "API文档生成完成"; \
	else \
		echo "swag未安装，跳过文档生成"; \
	fi

# Docker构建
docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker镜像构建完成: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Docker运行
docker-run: docker-build ## 运行Docker容器
	@echo "运行Docker容器..."
	docker run -d --name $(APP_NAME) -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Docker容器启动完成"

# Docker停止
docker-stop: ## 停止Docker容器
	@echo "停止Docker容器..."
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true
	@echo "Docker容器已停止"

# Docker Compose
docker-compose-up: ## 使用Docker Compose启动服务
	@echo "启动Docker Compose服务..."
	docker-compose up -d
	@echo "Docker Compose服务启动完成"

# Docker Compose停止
docker-compose-down: ## 停止Docker Compose服务
	@echo "停止Docker Compose服务..."
	docker-compose down
	@echo "Docker Compose服务已停止"

# 清理构建文件
clean: ## 清理构建文件
	@echo "清理构建文件..."
	rm -rf $(BUILD_DIR)
	rm -rf coverage
	rm -rf tmp
	@echo "清理完成"

# 完整构建流程
all: clean install quality test build ## 完整构建流程
	@echo "完整构建流程完成"

# 开发环境设置
setup-dev: ## 设置开发环境
	@echo "设置开发环境..."
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "环境配置文件已创建: .env"; \
	fi
	@if [ ! -d storage ]; then \
		mkdir -p storage/app/public storage/app/private storage/logs storage/temp; \
		echo "存储目录已创建"; \
	fi
	@echo "开发环境设置完成"

# 安全检查
security: ## 运行安全检查
	@echo "运行安全检查..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec未安装，跳过安全检查"; \
	fi

# 性能分析
profile: ## 运行性能分析
	@echo "启动性能分析服务器..."
	go run $(MAIN_FILE) &
	@sleep 3
	@echo "性能分析服务器已启动: http://localhost:6060/debug/pprof/"
	@echo "按Ctrl+C停止"

# 监控
monitor: ## 启动监控
	@echo "启动监控..."
	@if [ -f monitoring/prometheus.yml ]; then \
		prometheus --config.file=monitoring/prometheus.yml; \
	else \
		echo "Prometheus配置文件不存在"; \
	fi

# 部署到开发环境
deploy-dev: build ## 部署到开发环境
	@echo "部署到开发环境..."
	@if [ -f scripts/deploy-dev.sh ]; then \
		./scripts/deploy-dev.sh; \
	else \
		echo "开发环境部署脚本不存在"; \
	fi

# 部署到生产环境
deploy-prod: build ## 部署到生产环境
	@echo "部署到生产环境..."
	@if [ -f scripts/deploy-prod.sh ]; then \
		./scripts/deploy-prod.sh; \
	else \
		echo "生产环境部署脚本不存在"; \
	fi

# 备份数据库
backup: ## 备份数据库
	@echo "备份数据库..."
	@if [ -f scripts/backup.sh ]; then \
		./scripts/backup.sh; \
	else \
		echo "备份脚本不存在"; \
	fi

# 恢复数据库
restore: ## 恢复数据库
	@echo "恢复数据库..."
	@if [ -f scripts/restore.sh ]; then \
		./scripts/restore.sh; \
	else \
		echo "恢复脚本不存在"; \
	fi

# 健康检查
health: ## 健康检查
	@echo "执行健康检查..."
	@curl -f http://localhost:8080/api/v1/health || echo "服务未运行"

# 版本信息
version: ## 显示版本信息
	@echo "应用版本: $(shell git describe --tags --always --dirty)"
	@echo "Go版本: $(shell go version)"
	@echo "构建时间: $(shell date)"

# 依赖更新
update-deps: ## 更新依赖
	@echo "更新依赖..."
	go get -u ./...
	go mod tidy
	go mod verify

# 生成发布包
release: clean build ## 生成发布包
	@echo "生成发布包..."
	@mkdir -p release
	@cp $(BUILD_DIR)/$(APP_NAME) release/
	@cp -r storage release/
	@cp env.example release/
	@cp README.md release/
	@cp docs/DEPLOYMENT.md release/
	@tar -czf release/$(APP_NAME)-$(shell git describe --tags --always).tar.gz -C release .
	@echo "发布包生成完成: release/$(APP_NAME)-$(shell git describe --tags --always).tar.gz"

# 安装开发工具
install-tools: ## 安装开发工具
	@echo "安装开发工具..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/stretchr/testify@latest
	@echo "开发工具安装完成"

# 项目统计
stats: ## 显示项目统计信息
	@echo "项目统计信息:"
	@echo "Go文件数量: $(shell find . -name "*.go" | wc -l)"
	@echo "测试文件数量: $(shell find . -name "*_test.go" | wc -l)"
	@echo "代码行数: $(shell find . -name "*.go" -exec wc -l {} + | tail -1)"
	@echo "提交次数: $(shell git rev-list --count HEAD)"
	@echo "最后提交: $(shell git log -1 --format=%cd)"

# 代码复杂度分析
complexity: ## 分析代码复杂度
	@echo "分析代码复杂度..."
	@if command -v gocyclo > /dev/null; then \
		gocyclo -over 10 .; \
	else \
		echo "gocyclo未安装，跳过复杂度分析"; \
	fi

# 依赖分析
deps-analysis: ## 分析依赖
	@echo "分析依赖..."
	go mod graph | dot -Tpng -o deps.png
	@echo "依赖图已生成: deps.png"

# 内存分析
mem-profile: ## 生成内存分析报告
	@echo "生成内存分析报告..."
	@mkdir -p profiles
	go test -memprofile=profiles/mem.prof ./...
	go tool pprof -png profiles/mem.prof > profiles/mem.png
	@echo "内存分析报告已生成: profiles/mem.png"

# CPU分析
cpu-profile: ## 生成CPU分析报告
	@echo "生成CPU分析报告..."
	@mkdir -p profiles
	go test -cpuprofile=profiles/cpu.prof ./...
	go tool pprof -png profiles/cpu.prof > profiles/cpu.png
	@echo "CPU分析报告已生成: profiles/cpu.png"

# 代码生成
generate: ## 生成代码
	@echo "生成代码..."
	go generate ./...
	@echo "代码生成完成"

# 验证构建
verify: build test lint ## 验证构建
	@echo "构建验证完成"

# 快速开发
quick: format lint test ## 快速开发流程
	@echo "快速开发流程完成"

# 完整检查
check: install quality test-coverage security ## 完整检查
	@echo "完整检查完成"

