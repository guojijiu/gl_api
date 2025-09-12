# 云平台API - 完整命令文档

## 📋 目录

- [项目概述](#项目概述)
- [环境要求](#环境要求)
- [Makefile 命令](#makefile-命令)
- [Go 项目命令](#go-项目命令)
- [脚本命令](#脚本命令)
- [Docker 命令](#docker-命令)
- [Kubernetes 命令](#kubernetes-命令)
- [数据库命令](#数据库命令)
- [测试命令](#测试命令)
- [部署命令](#部署命令)
- [监控命令](#监控命令)
- [工具命令](#工具命令)
- [故障排除](#故障排除)

## 项目概述

云平台API是一个基于Go语言开发的现代化Web API框架，采用Gin + Laravel设计理念，提供完整的开发、测试、部署和监控解决方案。

**技术栈：**
- Go 1.21+
- Gin Web框架
- GORM ORM
- Redis缓存
- PostgreSQL/MySQL数据库
- Docker容器化
- Kubernetes编排
- Prometheus监控

## 环境要求

### 基础环境
- Go 1.21+
- Git
- Make (可选，用于Makefile命令)

### 开发工具
- golangci-lint (代码质量检查)
- air (热重载)
- hey (性能测试)
- pprof (性能分析)

### 容器化环境
- Docker 20.10+
- Docker Compose 2.0+

### 编排环境
- Kubernetes 1.20+
- kubectl
- Helm 3.0+ (可选)

## Makefile 命令

### 帮助和版本
```bash
# 显示所有可用命令
make help

# 显示版本信息
make version

# 检查开发环境
make check-env
```

### 开发命令
```bash
# 启动开发环境
make dev

# 启动开发环境（文件监控）
make dev-watch

# 快速启动（安装依赖、构建、测试）
make quick-start
```

### 构建命令
```bash
# 构建应用
make build

# 构建Linux版本
make build-linux

# 构建Windows版本
make build-windows

# 构建所有平台版本
make build-all

# 构建生产环境版本
make build-prod
```

### 测试命令
```bash
# 运行单元测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 运行集成测试
make test-integration

# 运行基准测试
make test-benchmark

# 运行所有测试
make test-all
```

### 代码质量命令
```bash
# 运行代码检查
make lint

# 格式化代码
make fmt

# 运行go vet
make vet

# 运行安全扫描
make security-scan
```

### 依赖管理
```bash
# 安装依赖
make deps

# 更新依赖
make deps-update

# 安装开发工具
make install-tools
```

### Docker命令
```bash
# 构建Docker镜像
make docker-build

# 运行Docker容器
make docker-run

# 停止Docker容器
make docker-stop

# 清理Docker资源
make clean-docker
```

### 部署命令
```bash
# 使用Docker部署
make deploy-docker

# 使用Kubernetes部署
make deploy-k8s
```

### 数据库命令
```bash
# 运行数据库迁移
make db-migrate

# 回滚数据库迁移
make db-rollback

# 填充测试数据
make db-seed
```

### 监控和日志
```bash
# 查看应用日志
make logs

# 启动监控
make monitor

# 检查应用健康状态
make health
```

### 性能测试
```bash
# 运行性能测试
make benchmark
```

### 文档生成
```bash
# 生成API文档
make docs
```

### 清理命令
```bash
# 清理构建文件
make clean

# 清理Docker资源
make clean-docker
```

### 统计信息
```bash
# 显示项目统计信息
make stats
```

## Go 项目命令

### 基础命令
```bash
# 运行应用
go run main.go

# 构建应用
go build -o bin/app main.go

# 安装依赖
go mod download
go mod tidy

# 运行测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行基准测试
go test -bench=. ./...

# 格式化代码
go fmt ./...

# 代码检查
go vet ./...

# 生成文档
go doc ./...
```

### 工具命令
```bash
# 生成JWT密钥
go run scripts/jwt-tools/generate-jwt-secret.go

# 数据库迁移
go run scripts/migrate.go -action migrate
go run scripts/migrate.go -action rollback -steps 1
go run scripts/migrate.go -action reset
go run scripts/migrate.go -action status

# 性能测试工具
go run scripts/performance-tools/performance_test.go
```

### 开发工具安装
```bash
# 安装热重载工具
go install github.com/cosmtrek/air@latest

# 安装代码质量检查工具
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装性能测试工具
go install github.com/rakyll/hey@latest

# 安装性能分析工具
go install github.com/google/pprof@latest

# 安装安全扫描工具
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# 安装代码覆盖率工具
go install github.com/axw/gocov/gocov@latest
go install github.com/AlekSi/gocov-xml@latest

# 安装测试报告工具
go install github.com/jstemmer/go-junit-report@latest
```

## 脚本命令

### 代码质量检查
```bash
# Linux/Mac
./scripts/code_quality.sh

# Windows
.\scripts\code_quality.ps1
```

**功能：**
- 代码格式化检查 (gofmt, goimports)
- 代码质量检查 (golangci-lint)
- 安全检查 (gosec)
- 依赖检查 (go mod tidy)
- 测试覆盖率检查
- 重复代码检查
- 生成质量报告

### 测试脚本
```bash
# 运行所有测试
./scripts/run_tests.sh

# 运行单元测试
./scripts/run_tests.sh unit

# 运行集成测试
./scripts/run_tests.sh integration

# 运行性能测试
./scripts/run_tests.sh performance

# 运行安全测试
./scripts/run_tests.sh security

# 清理测试环境
./scripts/run_tests.sh clean
```

### 性能测试脚本
```bash
# 运行所有性能测试
./scripts/run_benchmarks.sh

# 运行基准测试
./scripts/run_benchmarks.sh benchmark

# 运行负载测试
./scripts/run_benchmarks.sh load

# 运行压力测试
./scripts/run_benchmarks.sh stress

# 运行内存泄漏测试
./scripts/run_benchmarks.sh memory

# 清理测试环境
./scripts/run_benchmarks.sh clean
```

### 部署脚本
```bash
# Linux/Mac部署
./scripts/deploy.sh [local|docker|production]

# Windows部署
scripts\deploy.bat [local|docker|production]

# Docker部署
./scripts/deploy_docker.sh [build|deploy|start|stop|restart|logs|shell|status|cleanup|cleanup-full]

# Kubernetes部署
./scripts/deploy_k8s.sh [deploy|update|scale|status|logs|shell|cleanup]
```

### 文档维护脚本
```bash
# 检查文档完整性
./scripts/docs_maintenance.sh
```

## Docker 命令

### 基础Docker命令
```bash
# 构建镜像
docker build -t cloud-platform-api:latest .

# 运行容器
docker run -d --name cloud-platform-api -p 8080:8080 cloud-platform-api:latest

# 停止容器
docker stop cloud-platform-api

# 删除容器
docker rm cloud-platform-api

# 查看容器日志
docker logs -f cloud-platform-api

# 进入容器
docker exec -it cloud-platform-api /bin/sh

# 查看容器状态
docker ps -a
```

### Docker Compose命令
```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f

# 重启服务
docker-compose restart

# 构建并启动
docker-compose up --build -d

# 清理资源
docker-compose down -v --remove-orphans
```

### 生产环境Docker命令
```bash
# 使用生产环境配置
docker-compose -f docker-compose.prod.yml up -d

# 构建生产镜像
docker build -f Dockerfile.prod -t cloud-platform-api:prod .
```

## Kubernetes 命令

### 基础Kubernetes命令
```bash
# 查看集群信息
kubectl cluster-info

# 查看节点状态
kubectl get nodes

# 查看命名空间
kubectl get namespaces

# 创建命名空间
kubectl create namespace cloud-platform
```

### 应用部署
```bash
# 部署应用
kubectl apply -f k8s/deployment.yaml

# 创建服务
kubectl apply -f k8s/service.yaml

# 创建Ingress
kubectl apply -f k8s/ingress.yaml

# 创建HPA
kubectl apply -f k8s/hpa.yaml

# 查看部署状态
kubectl get pods -n cloud-platform

# 查看服务
kubectl get svc -n cloud-platform

# 查看Ingress
kubectl get ingress -n cloud-platform
```

### 应用管理
```bash
# 查看Pod详情
kubectl describe pod <pod-name> -n cloud-platform

# 查看Pod日志
kubectl logs -f deployment/cloud-platform-api -n cloud-platform

# 进入Pod
kubectl exec -it deployment/cloud-platform-api -n cloud-platform -- /bin/sh

# 扩缩容
kubectl scale deployment cloud-platform-api --replicas=5 -n cloud-platform

# 更新镜像
kubectl set image deployment/cloud-platform-api cloud-platform-api=your-registry/cloud-platform-api:v1.1.0 -n cloud-platform

# 滚动更新
kubectl rollout status deployment/cloud-platform-api -n cloud-platform

# 回滚
kubectl rollout undo deployment/cloud-platform-api -n cloud-platform
```

### 配置管理
```bash
# 创建ConfigMap
kubectl apply -f k8s/configmap.yaml

# 创建Secret
kubectl apply -f k8s/secret.yaml

# 查看配置
kubectl get configmap -n cloud-platform
kubectl get secret -n cloud-platform
```

### 监控和调试
```bash
# 查看HPA状态
kubectl get hpa -n cloud-platform

# 查看事件
kubectl get events -n cloud-platform

# 查看资源使用情况
kubectl top pods -n cloud-platform
kubectl top nodes

# 端口转发
kubectl port-forward service/cloud-platform-api-service 8080:80 -n cloud-platform
```

### 清理资源
```bash
# 删除部署
kubectl delete deployment cloud-platform-api -n cloud-platform

# 删除服务
kubectl delete service cloud-platform-api-service -n cloud-platform

# 删除命名空间（会删除所有相关资源）
kubectl delete namespace cloud-platform
```

## 数据库命令

### 数据库迁移
```bash
# 运行迁移
go run scripts/migrate.go -action migrate

# 回滚迁移
go run scripts/migrate.go -action rollback -steps 1

# 重置所有迁移
go run scripts/migrate.go -action reset

# 查看迁移状态
go run scripts/migrate.go -action status
```

### 数据库初始化
```bash
# 执行初始化SQL
psql -h localhost -U postgres -d cloud_platform -f scripts/init-db.sql

# 或使用MySQL
mysql -h localhost -u root -p cloud_platform < scripts/init-db.sql
```

### 数据库优化
```bash
# 执行优化SQL
psql -h localhost -U postgres -d cloud_platform -f scripts/optimize_database.sql
```

## 测试命令

### 单元测试
```bash
# 运行所有单元测试
go test ./...

# 运行特定包的测试
go test ./app/Services/...

# 运行测试并显示详细信息
go test -v ./...

# 运行测试并生成覆盖率报告
go test -cover ./...

# 运行测试并生成HTML覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 集成测试
```bash
# 运行集成测试
go test -tags=integration ./tests/Integration/...

# 运行负载测试
go test -tags=load ./tests/benchmark/...

# 运行安全测试
go test -tags=security ./tests/...
```

### 基准测试
```bash
# 运行基准测试
go test -bench=. ./tests/benchmark/...

# 运行基准测试并显示内存分配
go test -bench=. -benchmem ./tests/benchmark/...

# 运行基准测试并生成CPU profile
go test -bench=. -cpuprofile=cpu.prof ./tests/benchmark/...

# 分析CPU profile
go tool pprof cpu.prof
```

### 性能测试
```bash
# 使用hey进行HTTP负载测试
hey -n 1000 -c 10 http://localhost:8080/api/v1/health

# 使用hey进行压力测试
hey -n 10000 -c 100 -m GET http://localhost:8080/api/v1/health

# 使用pprof进行性能分析
go tool pprof http://localhost:8080/debug/pprof/profile
```

## 部署命令

### 本地部署
```bash
# 构建应用
go build -o bin/app main.go

# 设置环境变量
export DATABASE_URL="postgres://user:password@localhost:5432/cloud_platform?sslmode=disable"
export REDIS_URL="redis://localhost:6379/0"
export JWT_SECRET="your-secret-key"

# 运行应用
./bin/app
```

### Docker部署
```bash
# 构建镜像
docker build -t cloud-platform-api:latest .

# 运行容器
docker run -d \
  --name cloud-platform-api \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://user:password@host:5432/cloud_platform?sslmode=disable" \
  -e REDIS_URL="redis://host:6379/0" \
  -e JWT_SECRET="your-secret-key" \
  cloud-platform-api:latest
```

### Kubernetes部署
```bash
# 创建命名空间
kubectl create namespace cloud-platform

# 部署应用
kubectl apply -f k8s/ -n cloud-platform

# 检查部署状态
kubectl get pods -n cloud-platform
```

## 监控命令

### 应用监控
```bash
# 查看应用健康状态
curl http://localhost:8080/api/v1/health

# 查看详细健康检查
curl http://localhost:8080/api/v1/health/detailed

# 查看就绪状态
curl http://localhost:8080/api/v1/health/ready

# 查看存活状态
curl http://localhost:8080/api/v1/health/live
```

### Prometheus监控
```bash
# 查看指标
curl http://localhost:8080/api/v1/metrics

# 访问Prometheus UI
open http://localhost:9090
```

### Grafana监控
```bash
# 访问Grafana UI
open http://localhost:3000
# 默认用户名/密码: admin/admin
```

### 日志查看
```bash
# 查看应用日志
tail -f storage/logs/app.log

# 查看错误日志
tail -f storage/logs/error.log

# 查看访问日志
tail -f storage/logs/access.log
```

## 工具命令

### JWT工具
```bash
# 生成JWT密钥
go run scripts/jwt-tools/generate-jwt-secret.go
```

### 性能工具
```bash
# 运行性能测试
go run scripts/performance-tools/performance_test.go
```

### 代码质量工具
```bash
# 运行golangci-lint
golangci-lint run

# 运行gosec安全检查
gosec ./...

# 运行goimports
goimports -w .

# 运行gofmt
gofmt -w .
```

### 文档工具
```bash
# 生成Swagger文档
swag init -g main.go -o docs/swagger

# 生成Go文档
godoc -http=:6060
```

## 故障排除

### 常见问题

#### 1. 应用启动失败
```bash
# 检查端口是否被占用
netstat -tulpn | grep :8080

# 检查环境变量
env | grep -E "(DATABASE|REDIS|JWT)"

# 查看详细错误日志
tail -f storage/logs/error.log
```

#### 2. 数据库连接失败
```bash
# 检查数据库服务状态
systemctl status postgresql

# 测试数据库连接
psql -h localhost -U postgres -d cloud_platform

# 检查数据库配置
cat .env | grep DATABASE
```

#### 3. Redis连接失败
```bash
# 检查Redis服务状态
systemctl status redis

# 测试Redis连接
redis-cli ping

# 检查Redis配置
cat .env | grep REDIS
```

#### 4. Docker容器问题
```bash
# 查看容器日志
docker logs cloud-platform-api

# 检查容器状态
docker inspect cloud-platform-api

# 重启容器
docker restart cloud-platform-api
```

#### 5. Kubernetes部署问题
```bash
# 查看Pod状态
kubectl get pods -n cloud-platform

# 查看Pod详情
kubectl describe pod <pod-name> -n cloud-platform

# 查看Pod日志
kubectl logs <pod-name> -n cloud-platform

# 查看事件
kubectl get events -n cloud-platform
```

### 调试命令

#### 应用调试
```bash
# 启用调试模式
export LOG_LEVEL=debug
export GIN_MODE=debug

# 运行应用
go run main.go
```

#### 性能调试
```bash
# 生成CPU profile
go tool pprof http://localhost:8080/debug/pprof/profile

# 生成内存profile
go tool pprof http://localhost:8080/debug/pprof/heap

# 生成goroutine profile
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

#### 网络调试
```bash
# 检查端口监听
netstat -tulpn | grep :8080

# 测试HTTP连接
curl -v http://localhost:8080/api/v1/health

# 检查DNS解析
nslookup your-domain.com
```

### 日志分析

#### 应用日志
```bash
# 查看实时日志
tail -f storage/logs/app.log

# 搜索错误日志
grep -i error storage/logs/app.log

# 统计日志级别
grep -o '\[ERROR\]\|\[WARN\]\|\[INFO\]' storage/logs/app.log | sort | uniq -c
```

#### 系统日志
```bash
# 查看系统日志
journalctl -u cloud-platform-api -f

# 查看Docker日志
docker logs -f cloud-platform-api

# 查看Kubernetes日志
kubectl logs -f deployment/cloud-platform-api -n cloud-platform
```

## 总结

本命令文档涵盖了云平台API项目的所有可用命令，包括：

1. **Makefile命令** - 提供便捷的开发、构建、测试和部署命令
2. **Go项目命令** - 基础的Go开发命令和工具
3. **脚本命令** - 自动化脚本，包括代码质量、测试、部署等
4. **Docker命令** - 容器化部署和管理命令
5. **Kubernetes命令** - 容器编排和集群管理命令
6. **数据库命令** - 数据库迁移和管理命令
7. **测试命令** - 各种测试和性能测试命令
8. **部署命令** - 不同环境的部署命令
9. **监控命令** - 应用监控和日志查看命令
10. **工具命令** - 各种开发工具的使用命令
11. **故障排除** - 常见问题的诊断和解决命令

使用这些命令可以完成从开发到生产部署的完整流程，确保项目的质量和稳定性。

---

**注意：** 在使用任何命令前，请确保已正确配置环境变量和依赖项。建议在测试环境中先验证命令的正确性。
