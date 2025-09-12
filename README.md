# Cloud Platform API

基于Gin + Laravel设计理念的现代化Web开发框架，提供完整的云平台API解决方案。

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-1.9+-00D4AA?style=flat-square)](https://gin-gonic.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Supported-2496ED?style=flat-square&logo=docker)](docker-compose.yml)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Ready-326CE5?style=flat-square&logo=kubernetes)](k8s/)

## 📋 目录

- [核心特性](#-核心特性)
- [系统架构](#-系统架构)
- [快速开始](#-快速开始)
- [配置说明](#-配置说明)
- [API文档](#-api文档)
- [部署指南](#-部署指南)
- [监控和日志](#-监控和日志)
- [安全特性](#-安全特性)
- [性能优化](#-性能优化)
- [开发指南](#-开发指南)
- [故障排除](#-故障排除)
- [贡献指南](#-贡献指南)
- [文档索引](#-文档索引)

## 📚 文档索引

> **快速导航**：查看 [完整文档索引](docs/INDEX.md) 获取所有文档的快速访问链接

### 🚀 快速开始
- [快速开始指南](docs/QUICK_START.md) - 5分钟快速体验
- [开发环境设置](docs/DEVELOPMENT.md) - 详细的开发环境配置
- [测试指南](docs/TESTING.md) - 完整的测试框架说明

### 📖 系统文档
- [API接口文档](docs/API.md) - 完整的API接口说明
- [配置系统](app/Config/README.md) - 配置管理详解
- [存储系统](storage/README.md) - 文件存储和管理
- [日志系统](docs/LOGGING_SYSTEM.md) - 日志管理和监控

### 🔧 高级功能
- [安全系统](docs/SECURITY_SYSTEM.md) - 安全防护和认证
- [性能监控](docs/PERFORMANCE_MONITORING_SYSTEM.md) - 系统性能监控
- [查询优化](docs/QUERY_OPTIMIZATION_SYSTEM.md) - 数据库性能优化
- [WebSocket系统](docs/WEBSOCKET_SYSTEM.md) - 实时通信功能
- [监控告警](docs/MONITORING_SYSTEM.md) - 监控和告警系统

### 🚀 部署相关
- [部署指南](docs/DEPLOYMENT.md) - 各种环境部署方法
- [Kubernetes部署](k8s/README.md) - K8s集群部署
- [脚本工具](scripts/README.md) - 自动化脚本使用
- [故障排除](scripts/TROUBLESHOOTING.md) - 常见问题解决

### 📝 文档维护
- [文档贡献指南](docs/CONTRIBUTING.md) - 如何贡献文档
- [文档更新日志](docs/CHANGELOG.md) - 文档变更记录
- [文档维护脚本](scripts/docs_maintenance.sh) - 文档检查工具

## 🚀 核心特性

### 🔐 安全认证系统
- **JWT Token认证** - 安全的无状态认证机制，支持刷新令牌
- **基于角色的权限控制** - 细粒度的权限管理，支持多级权限
- **Token黑名单机制** - 支持token撤销和登出，增强安全性
- **密码强度验证** - 自动检测密码安全性，支持自定义规则
- **邮箱验证系统** - 完整的邮箱验证流程，支持验证码
- **密码重置功能** - 安全的密码重置机制，支持邮箱和短信
- **多因素认证** - 支持TOTP、短信验证码等MFA方式
- **API密钥管理** - 支持API密钥生成、管理和撤销

### 🛡️ 安全防护
- **XSS攻击防护** - 自动检测和阻止XSS攻击，支持CSP策略
- **SQL注入检测** - 实时SQL注入攻击检测，支持参数化查询
- **CSRF保护** - 跨站请求伪造防护，支持双重提交Cookie
- **请求速率限制** - 防止暴力攻击和DDoS，支持IP和用户级别限制
- **文件上传安全检查** - 安全的文件上传验证，支持病毒扫描
- **输入数据验证** - 全面的输入数据清理和验证，支持自定义规则
- **安全头设置** - 自动设置安全相关的HTTP头
- **IP白名单/黑名单** - 支持IP访问控制

### 📊 监控和日志
- **实时健康检查** - 系统状态监控，支持多维度检查
- **性能指标收集** - 详细的性能监控，包括响应时间、吞吐量等
- **Prometheus集成** - 标准化的监控指标，支持Grafana可视化
- **结构化日志记录** - 完整的操作日志，支持JSON格式
- **错误追踪** - 详细的错误信息和堆栈跟踪，支持错误聚合
- **审计日志** - 用户操作审计记录，支持合规要求
- **实时告警** - 支持邮件、短信、Webhook等多种告警方式
- **日志分析** - 支持日志搜索、过滤和分析

### 🗄️ 数据管理
- **多数据库支持** - MySQL、PostgreSQL、SQLite，支持读写分离
- **数据库迁移系统** - 版本化的数据库结构管理，支持回滚
- **连接池监控** - 数据库连接状态监控，支持动态调整
- **自动备份系统** - 数据备份和恢复，支持增量备份
- **缓存策略** - Redis缓存和内存缓存降级，支持多级缓存
- **数据加密** - 支持敏感数据加密存储
- **数据同步** - 支持主从数据库同步
- **查询优化** - 自动查询优化建议和慢查询监控

### 🔧 开发工具
- **自动化测试** - 单元测试、集成测试、性能测试，支持覆盖率报告
- **API文档生成** - Swagger自动文档，支持在线测试
- **代码质量检查** - 代码格式化和静态分析，支持CI/CD集成
- **热重载开发** - 开发环境热重载，提高开发效率
- **配置热重载** - 支持配置文件变更时自动重载，无需重启服务
- **熔断器模式** - 防止级联故障，提高系统稳定性
- **Docker支持** - 容器化部署，支持多环境部署
- **Kubernetes支持** - 支持K8s部署，包括HPA、PDB等
- **性能分析** - 支持pprof性能分析工具
- **调试工具** - 支持远程调试和日志追踪

## 🏗️ 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        Cloud Platform API                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │   Web UI    │  │  Mobile App │  │  API Client │  │  Admin  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                        Load Balancer                            │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │   Nginx     │  │   Nginx     │  │   Nginx     │  │  Nginx  │ │
│  │  (API-1)    │  │  (API-2)    │  │  (API-3)    │  │ (API-N) │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                    Cloud Platform API Layer                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │  Controllers│  │  Middleware │  │   Services   │  │  Models │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                      Data Layer                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │   MySQL     │  │   Redis     │  │   Storage   │  │  Queue  │ │
│  │ (Primary)   │  │  (Cache)    │  │  (Files)    │  │ (Jobs)  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### 核心组件

#### 1. API层 (Controllers)
- **认证控制器** - 处理用户登录、注册、密码重置等
- **用户控制器** - 管理用户信息和权限
- **内容控制器** - 处理文章、分类、标签等内容管理
- **存储控制器** - 处理文件上传、下载、管理
- **监控控制器** - 提供健康检查和系统状态

#### 2. 中间件层 (Middleware)
- **认证中间件** - JWT验证和权限检查
- **安全中间件** - XSS、CSRF、SQL注入防护
- **日志中间件** - 请求日志记录和审计
- **限流中间件** - 请求频率限制和DDoS防护
- **监控中间件** - 性能指标收集

#### 3. 服务层 (Services)
- **认证服务** - 用户认证和授权逻辑
- **用户服务** - 用户管理业务逻辑
- **存储服务** - 文件存储和管理
- **监控服务** - 系统监控和告警
- **缓存服务** - 数据缓存和优化

#### 4. 数据层 (Models)
- **用户模型** - 用户数据结构和关系
- **内容模型** - 文章、分类、标签等数据结构
- **审计模型** - 操作日志和审计记录
- **监控模型** - 性能指标和系统状态

### 技术栈

#### 后端技术
- **Go 1.21+** - 主要编程语言
- **Gin** - Web框架
- **GORM** - ORM框架
- **JWT-Go** - JWT认证
- **Viper** - 配置管理
- **Zap** - 日志框架
- **Prometheus** - 监控指标

#### 数据库
- **MySQL 8.0+** - 主数据库
- **PostgreSQL 12+** - 可选数据库
- **SQLite 3** - 开发/测试数据库
- **Redis 6.0+** - 缓存和会话存储

#### 部署和运维
- **Docker** - 容器化
- **Kubernetes** - 容器编排
- **Nginx** - 反向代理和负载均衡
- **Prometheus** - 监控系统
- **Grafana** - 监控可视化

## 📋 系统要求

### 最低要求
- **Go**: 1.21+
- **内存**: 512MB
- **CPU**: 1核心
- **存储**: 1GB可用空间

### 推荐配置
- **Go**: 1.21+
- **内存**: 2GB+
- **CPU**: 2核心+
- **存储**: 10GB+ SSD

### 数据库要求
- **MySQL**: 8.0+ (推荐)
- **PostgreSQL**: 12+ (可选)
- **SQLite**: 3.x (开发环境)
- **Redis**: 6.0+ (缓存)

## 🛠️ 快速开始

### 方法一：使用Docker Compose（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/your-username/cloud-platform-api.git
cd cloud-platform-api

# 2. 复制环境配置文件
cp env.example .env

# 3. 启动所有服务（包括数据库、Redis等）
docker-compose up -d

# 4. 查看服务状态
docker-compose ps

# 5. 查看日志
docker-compose logs -f cloud-platform-api
```

### 方法二：本地开发环境

#### 1. 环境准备
```bash
# 安装Go 1.21+
# 下载地址：https://golang.org/dl/

# 验证Go安装
go version

# 安装MySQL/PostgreSQL（可选，开发环境可使用SQLite）
# MySQL: https://dev.mysql.com/downloads/
# PostgreSQL: https://www.postgresql.org/download/

# 安装Redis（可选）
# Redis: https://redis.io/download
```

#### 2. 克隆和配置项目
```bash
# 克隆项目
git clone https://github.com/your-username/cloud-platform-api.git
cd cloud-platform-api

# 安装Go依赖
go mod download
go mod tidy

# 复制环境配置文件
cp env.example .env

# 编辑配置文件（根据你的环境调整）
# Windows: notepad .env
# Linux/Mac: nano .env 或 vim .env
```

#### 3. 配置数据库
```bash
# 开发环境使用SQLite（默认配置）
# 无需额外配置，直接运行即可

# 使用MySQL
# 1. 创建数据库
mysql -u root -p
CREATE DATABASE cloud_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 2. 修改.env文件
# DB_DRIVER=mysql
# DB_HOST=localhost
# DB_PORT=3306
# DB_USERNAME=root
# DB_PASSWORD=your_password
# DB_DATABASE=cloud_platform

# 使用PostgreSQL
# 1. 创建数据库
psql -U postgres
CREATE DATABASE cloud_platform;

# 2. 修改.env文件
# DB_DRIVER=postgres
# DB_HOST=localhost
# DB_PORT=5432
# DB_USERNAME=postgres
# DB_PASSWORD=your_password
# DB_DATABASE=cloud_platform
```

#### 4. 运行数据库迁移
```bash
# 运行数据库迁移
go run scripts/migrate.go

# 或者使用Make命令
make migrate
```

#### 5. 启动应用
```bash
# 开发模式（热重载）
make dev

# 或者直接运行
go run main.go

# 生产模式
make build
./build/cloud-platform-api

# 或者使用go run
go run main.go
```

### 方法三：使用脚本快速启动

```bash
# Windows用户
.\scripts\quick_start.ps1 -Environment development

# Linux/Mac用户
./scripts/quick_start.sh development
```

### 验证安装

#### 1. 检查服务状态
```bash
# 检查API服务
curl http://localhost:8080/api/v1/health

# 预期响应
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

#### 2. 查看API文档
```bash
# 访问Swagger文档
# 浏览器打开：http://localhost:8080/swagger/index.html
```

#### 3. 测试API接口
```bash
# 注册用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# 用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 常见问题

#### 1. 端口被占用
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac
lsof -i :8080
kill -9 <PID>
```

#### 2. 数据库连接失败
- 检查数据库服务是否启动
- 验证.env文件中的数据库配置
- 确认数据库用户权限

#### 3. 依赖安装失败
```bash
# 清理模块缓存
go clean -modcache
go mod download
go mod tidy
```

#### 4. 权限问题（Linux/Mac）
```bash
# 给脚本执行权限
chmod +x scripts/*.sh
chmod +x scripts/*.ps1
```

## 🔧 配置说明

### 配置文件结构

项目支持多种配置方式，按优先级从高到低：
1. 环境变量
2. `.env` 文件
3. 默认配置

### 环境变量配置

#### 服务器配置
```env
# 服务器基本配置
SERVER_PORT=8080                    # 服务器端口
SERVER_MODE=debug                   # 运行模式: debug/production
SERVER_BASE_URL=http://localhost:8080  # 基础URL
SERVER_READ_TIMEOUT=30s             # 读取超时
SERVER_WRITE_TIMEOUT=30s            # 写入超时
SERVER_IDLE_TIMEOUT=120s            # 空闲超时
SERVER_MAX_HEADER_BYTES=1048576     # 最大请求头大小
```

#### 数据库配置
```env
# 数据库配置
DB_DRIVER=mysql                     # 数据库驱动: mysql/postgres/sqlite
DB_HOST=localhost                   # 数据库主机
DB_PORT=3306                        # 数据库端口
DB_USERNAME=root                    # 数据库用户名
DB_PASSWORD=your-password           # 数据库密码
DB_DATABASE=cloud_platform          # 数据库名称
DB_CHARSET=utf8mb4                  # 字符集
DB_MAX_OPEN_CONNS=100               # 最大打开连接数
DB_MAX_IDLE_CONNS=10                # 最大空闲连接数
DB_CONN_MAX_LIFETIME=3600s          # 连接最大生存时间
```

#### JWT认证配置
```env
# JWT配置
JWT_SECRET=your-super-secret-jwt-key-change-in-production-must-be-at-least-32-characters-long
JWT_EXPIRE_TIME=24                  # Token过期时间（小时）
JWT_REFRESH_EXPIRE_TIME=168         # 刷新Token过期时间（小时）
JWT_ISSUER=cloud-platform-api       # JWT签发者
JWT_AUDIENCE=cloud-platform-users   # JWT受众
```

#### Redis缓存配置
```env
# Redis配置
REDIS_HOST=localhost                # Redis主机
REDIS_PORT=6379                     # Redis端口
REDIS_PASSWORD=                     # Redis密码（可选）
REDIS_DATABASE=0                    # Redis数据库编号
REDIS_POOL_SIZE=10                  # 连接池大小
REDIS_MIN_IDLE_CONNS=5              # 最小空闲连接数
REDIS_MAX_RETRIES=3                 # 最大重试次数
REDIS_DIAL_TIMEOUT=5s               # 连接超时
REDIS_READ_TIMEOUT=3s               # 读取超时
REDIS_WRITE_TIMEOUT=3s              # 写入超时
```

#### 存储配置
```env
# 文件存储配置
STORAGE_UPLOAD_PATH=./storage/app/public    # 上传路径
STORAGE_MAX_FILE_SIZE=10                    # 最大文件大小（MB）
STORAGE_ALLOWED_TYPES=jpg,jpeg,png,gif,pdf,doc,docx  # 允许的文件类型
STORAGE_PRIVATE_PATH=./storage/app/private  # 私有文件路径
STORAGE_PUBLIC_PATH=./storage/app/public    # 公共文件路径
STORAGE_TEMP_PATH=./storage/temp            # 临时文件路径
STORAGE_LOG_PATH=./storage/logs             # 日志文件路径
STORAGE_CACHE_PATH=./storage/framework/cache # 缓存文件路径
```

#### 安全配置
```env
# 安全防护配置
SECURITY_ENABLE_XSS_PROTECTION=true         # 启用XSS防护
SECURITY_ENABLE_SQL_INJECTION_CHECK=true    # 启用SQL注入检测
SECURITY_ENABLE_CSRF_PROTECTION=true        # 启用CSRF防护
SECURITY_ENABLE_RATE_LIMIT=true             # 启用速率限制
SECURITY_MAX_LOGIN_ATTEMPTS=5               # 最大登录尝试次数
SECURITY_LOCKOUT_DURATION=15m               # 账户锁定时间
SECURITY_PASSWORD_MIN_LENGTH=8              # 密码最小长度
SECURITY_PASSWORD_REQUIRE_UPPERCASE=true    # 密码需要大写字母
SECURITY_PASSWORD_REQUIRE_LOWERCASE=true    # 密码需要小写字母
SECURITY_PASSWORD_REQUIRE_NUMBER=true       # 密码需要数字
SECURITY_PASSWORD_REQUIRE_SYMBOL=true       # 密码需要特殊字符
```

#### 监控配置
```env
# 监控配置
MONITORING_ENABLE_METRICS=true              # 启用指标收集
MONITORING_ENABLE_HEALTH_CHECK=true         # 启用健康检查
MONITORING_ENABLE_PROMETHEUS=true           # 启用Prometheus集成
MONITORING_METRICS_PATH=/metrics            # 指标端点路径
MONITORING_HEALTH_PATH=/health              # 健康检查端点路径
MONITORING_LOG_LEVEL=info                   # 日志级别
MONITORING_LOG_FORMAT=json                  # 日志格式: json/text
```

#### 邮件配置
```env
# 邮件服务配置
EMAIL_HOST=smtp.gmail.com                   # SMTP服务器
EMAIL_PORT=587                              # SMTP端口
EMAIL_USERNAME=your-email@gmail.com         # 邮箱用户名
EMAIL_PASSWORD=your-app-password            # 邮箱密码或应用密码
EMAIL_FROM_NAME=Cloud Platform API          # 发件人名称
EMAIL_FROM_ADDRESS=noreply@example.com      # 发件人邮箱
EMAIL_USE_TLS=true                          # 使用TLS
EMAIL_USE_SSL=false                         # 使用SSL
```

#### 日志配置
```env
# 日志配置
LOG_LEVEL=info                              # 日志级别: debug/info/warn/error
LOG_FORMAT=json                             # 日志格式: json/text
LOG_OUTPUT=stdout                           # 日志输出: stdout/file/both
LOG_FILE_PATH=./storage/logs/app.log        # 日志文件路径
LOG_MAX_SIZE=100                            # 日志文件最大大小（MB）
LOG_MAX_AGE=30                              # 日志文件最大保存天数
LOG_MAX_BACKUPS=10                          # 日志文件最大备份数
LOG_COMPRESS=true                           # 是否压缩日志文件
```

### 环境特定配置

#### 开发环境 (.env.development)
```env
SERVER_MODE=debug
LOG_LEVEL=debug
LOG_FORMAT=text
DB_DRIVER=sqlite
REDIS_HOST=
MONITORING_ENABLE_METRICS=false
SECURITY_ENABLE_RATE_LIMIT=false
```

#### 测试环境 (.env.testing)
```env
SERVER_MODE=debug
LOG_LEVEL=info
LOG_FORMAT=json
DB_DRIVER=sqlite
REDIS_HOST=localhost
MONITORING_ENABLE_METRICS=true
SECURITY_ENABLE_RATE_LIMIT=true
```

#### 生产环境 (.env.production)
```env
SERVER_MODE=production
LOG_LEVEL=warn
LOG_FORMAT=json
DB_DRIVER=mysql
REDIS_HOST=redis-server
MONITORING_ENABLE_METRICS=true
SECURITY_ENABLE_RATE_LIMIT=true
# 生产环境必须修改所有默认密码和密钥
```

### 配置验证

应用启动时会自动验证配置的有效性：

```bash
# 验证配置
go run main.go --validate-config

# 查看当前配置
go run main.go --show-config

# 测试数据库连接
go run main.go --test-db

# 测试Redis连接
go run main.go --test-redis
```

### 配置最佳实践

1. **环境隔离**: 为不同环境使用不同的配置文件
2. **敏感信息**: 使用环境变量存储敏感信息，不要提交到版本控制
3. **默认值**: 为所有配置项提供合理的默认值
4. **验证**: 在应用启动时验证配置的有效性
5. **文档**: 保持配置文档的更新和完整
6. **安全**: 生产环境使用强密码和密钥
7. **监控**: 监控配置变更和配置错误

## 📚 API文档

### 接口概览

所有API接口都遵循RESTful设计原则，使用JSON格式进行数据交换。

**基础URL**: `http://localhost:8080/api/v1`

**认证方式**: Bearer Token (JWT)

**响应格式**:
```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 认证接口

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "string",
  "email": "string",
  "password": "string",
  "password_confirmation": "string"
}
```

**响应示例**:
```json
{
  "code": 201,
  "message": "用户注册成功",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "created_at": "2024-01-01T00:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "string",
  "password": "string"
}
```

#### 用户登出
```http
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

#### 获取用户资料
```http
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

#### 更新用户资料
```http
PUT /api/v1/auth/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "string",
  "email": "string"
}
```

#### 刷新Token
```http
POST /api/v1/auth/refresh
Authorization: Bearer <token>
```

#### 请求密码重置
```http
POST /api/v1/auth/password/reset-request
Content-Type: application/json

{
  "email": "string"
}
```

#### 重置密码
```http
POST /api/v1/auth/password/reset
Content-Type: application/json

{
  "token": "string",
  "password": "string",
  "password_confirmation": "string"
}
```

#### 请求邮箱验证
```http
POST /api/v1/auth/email/verify-request
Authorization: Bearer <token>
```

#### 验证邮箱
```http
POST /api/v1/auth/email/verify
Content-Type: application/json

{
  "token": "string"
}
```

### 用户管理接口

#### 获取用户列表
```http
GET /api/v1/users?page=1&limit=10&search=keyword
Authorization: Bearer <token>
```

**查询参数**:
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10，最大100）
- `search`: 搜索关键词
- `sort`: 排序字段
- `order`: 排序方向（asc/desc）

#### 获取用户详情
```http
GET /api/v1/users/:id
Authorization: Bearer <token>
```

#### 更新用户信息
```http
PUT /api/v1/users/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "string",
  "email": "string",
  "role": "string"
}
```

#### 删除用户
```http
DELETE /api/v1/users/:id
Authorization: Bearer <token>
```

#### 获取用户的文章列表
```http
GET /api/v1/users/:id/posts?page=1&limit=10
Authorization: Bearer <token>
```

### 内容管理接口

#### 获取文章列表
```http
GET /api/v1/posts?page=1&limit=10&category_id=1&tag_id=1&status=published
Authorization: Bearer <token>
```

**查询参数**:
- `page`: 页码
- `limit`: 每页数量
- `category_id`: 分类ID
- `tag_id`: 标签ID
- `status`: 状态（draft/published/archived）
- `search`: 搜索关键词

#### 获取文章详情
```http
GET /api/v1/posts/:id
Authorization: Bearer <token>
```

#### 创建文章
```http
POST /api/v1/posts
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "string",
  "content": "string",
  "excerpt": "string",
  "category_id": 1,
  "tag_ids": [1, 2, 3],
  "status": "draft",
  "featured_image": "string"
}
```

#### 更新文章
```http
PUT /api/v1/posts/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "string",
  "content": "string",
  "excerpt": "string",
  "category_id": 1,
  "tag_ids": [1, 2, 3],
  "status": "published"
}
```

#### 删除文章
```http
DELETE /api/v1/posts/:id
Authorization: Bearer <token>
```

### 分类管理接口

#### 获取分类列表
```http
GET /api/v1/categories?page=1&limit=10
Authorization: Bearer <token>
```

#### 创建分类
```http
POST /api/v1/categories
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",
  "description": "string",
  "parent_id": 0
}
```

#### 更新分类
```http
PUT /api/v1/categories/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",
  "description": "string",
  "parent_id": 0
}
```

#### 删除分类
```http
DELETE /api/v1/categories/:id
Authorization: Bearer <token>
```

### 标签管理接口

#### 获取标签列表
```http
GET /api/v1/tags?page=1&limit=10
Authorization: Bearer <token>
```

#### 创建标签
```http
POST /api/v1/tags
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",
  "color": "string"
}
```

### 存储管理接口

#### 文件上传
```http
POST /api/v1/storage/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <file>
path: "uploads"
type: "public"
```

**响应示例**:
```json
{
  "code": 200,
  "message": "文件上传成功",
  "data": {
    "filename": "example.jpg",
    "path": "uploads/2024/01/01/example.jpg",
    "url": "http://localhost:8080/api/v1/storage/download/uploads/2024/01/01/example.jpg",
    "size": 1024,
    "mime_type": "image/jpeg"
  }
}
```

#### 文件下载
```http
GET /api/v1/storage/download/*path?type=public
Authorization: Bearer <token>
```

#### 删除文件
```http
DELETE /api/v1/storage/delete/*path?type=public
Authorization: Bearer <token>
```

#### 获取文件列表
```http
GET /api/v1/storage/list?path=uploads&type=public&page=1&limit=10
Authorization: Bearer <token>
```

#### 获取文件信息
```http
GET /api/v1/storage/info/*path?type=public
Authorization: Bearer <token>
```

### 监控接口

#### 健康检查
```http
GET /api/v1/health
```

**响应示例**:
```json
{
  "code": 200,
  "message": "服务正常",
  "data": {
    "status": "ok",
    "timestamp": "2024-01-01T00:00:00Z",
    "version": "1.0.0",
    "uptime": "1h30m45s"
  }
}
```

#### 详细健康检查
```http
GET /api/v1/health/detailed
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "服务正常",
  "data": {
    "status": "ok",
    "timestamp": "2024-01-01T00:00:00Z",
    "version": "1.0.0",
    "uptime": "1h30m45s",
    "database": {
      "status": "ok",
      "response_time": "5ms"
    },
    "redis": {
      "status": "ok",
      "response_time": "2ms"
    },
    "storage": {
      "status": "ok",
      "free_space": "50GB"
    }
  }
}
```

#### 获取系统指标
```http
GET /api/v1/metrics
Authorization: Bearer <token>
```

#### 获取系统状态
```http
GET /api/v1/status
Authorization: Bearer <token>
```

### 错误码说明

| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| 200 | 200 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未授权 |
| 403 | 403 | 禁止访问 |
| 404 | 404 | 资源不存在 |
| 409 | 409 | 资源冲突 |
| 422 | 422 | 数据验证失败 |
| 429 | 429 | 请求过于频繁 |
| 500 | 500 | 服务器内部错误 |

### 分页响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "pages": 10,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

### 在线API文档

访问以下地址查看完整的交互式API文档：

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **ReDoc**: `http://localhost:8080/redoc`
- **API Schema**: `http://localhost:8080/swagger/doc.json`

### API使用示例

#### 使用curl测试API

```bash
# 1. 注册用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "password_confirmation": "password123"
  }'

# 2. 用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# 3. 获取用户资料（需要Token）
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 4. 创建文章
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试文章",
    "content": "这是测试文章的内容",
    "excerpt": "文章摘要",
    "category_id": 1,
    "tag_ids": [1, 2],
    "status": "published"
  }'

# 5. 上传文件
curl -X POST http://localhost:8080/api/v1/storage/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/your/file.jpg" \
  -F "path=uploads" \
  -F "type=public"
```

#### 使用JavaScript测试API

```javascript
// 基础API客户端
class CloudPlatformAPI {
  constructor(baseURL, token = null) {
    this.baseURL = baseURL;
    this.token = token;
  }

  setToken(token) {
    this.token = token;
  }

  async request(method, endpoint, data = null) {
    const url = `${this.baseURL}${endpoint}`;
    const options = {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
    };

    if (this.token) {
      options.headers['Authorization'] = `Bearer ${this.token}`;
    }

    if (data) {
      options.body = JSON.stringify(data);
    }

    const response = await fetch(url, options);
    return await response.json();
  }

  // 认证相关
  async register(userData) {
    return this.request('POST', '/api/v1/auth/register', userData);
  }

  async login(credentials) {
    return this.request('POST', '/api/v1/auth/login', credentials);
  }

  async getProfile() {
    return this.request('GET', '/api/v1/auth/profile');
  }

  // 文章相关
  async getPosts(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    return this.request('GET', `/api/v1/posts?${queryString}`);
  }

  async createPost(postData) {
    return this.request('POST', '/api/v1/posts', postData);
  }

  // 文件上传
  async uploadFile(file, path = 'uploads', type = 'public') {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);
    formData.append('type', type);

    const response = await fetch(`${this.baseURL}/api/v1/storage/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
      },
      body: formData,
    });

    return await response.json();
  }
}

// 使用示例
const api = new CloudPlatformAPI('http://localhost:8080');

// 注册和登录
const registerResult = await api.register({
  username: 'testuser',
  email: 'test@example.com',
  password: 'password123',
  password_confirmation: 'password123'
});

const loginResult = await api.login({
  email: 'test@example.com',
  password: 'password123'
});

api.setToken(loginResult.data.token);

// 获取用户资料
const profile = await api.getProfile();
console.log(profile);

// 创建文章
const post = await api.createPost({
  title: '测试文章',
  content: '这是测试文章的内容',
  status: 'published'
});

// 上传文件
const fileInput = document.getElementById('fileInput');
const file = fileInput.files[0];
const uploadResult = await api.uploadFile(file);
console.log(uploadResult);
```

## 🔒 安全特性

### 认证和授权
- JWT Token认证
- 基于角色的权限控制
- Token黑名单机制
- 密码强度验证

### 安全防护
- XSS攻击防护
- SQL注入检测
- CSRF保护
- 请求速率限制
- 文件上传安全检查

### 数据保护
- 密码安全哈希
- 敏感信息加密
- 输入数据验证
- 输出数据清理

## 📊 监控和日志

### 健康检查
- 数据库连接状态监控
- Redis连接状态监控
- 系统资源使用监控
- 存储系统状态监控

### 性能监控
- 请求响应时间统计
- 数据库查询性能监控
- 内存使用情况监控
- 连接池状态监控

### 日志记录
- 请求日志记录
- SQL查询日志
- 错误日志记录
- 安全事件日志

## 🚀 部署指南

### 部署方式概览

| 部署方式 | 适用场景 | 复杂度 | 扩展性 | 推荐度 |
|----------|----------|--------|--------|--------|
| Docker Compose | 开发/测试/小规模生产 | 低 | 中 | ⭐⭐⭐⭐⭐ |
| Kubernetes | 大规模生产环境 | 高 | 高 | ⭐⭐⭐⭐⭐ |
| 传统部署 | 简单环境 | 中 | 低 | ⭐⭐⭐ |
| 云服务 | 快速部署 | 低 | 高 | ⭐⭐⭐⭐ |

### 方法一：Docker Compose部署（推荐）

#### 1. 准备环境
```bash
# 安装Docker和Docker Compose
# Docker: https://docs.docker.com/get-docker/
# Docker Compose: https://docs.docker.com/compose/install/

# 验证安装
docker --version
docker-compose --version
```

#### 2. 配置环境
```bash
# 克隆项目
git clone https://github.com/your-username/cloud-platform-api.git
cd cloud-platform-api

# 复制环境配置文件
cp env.example .env

# 编辑配置文件
nano .env
```

#### 3. 启动服务
```bash
# 启动所有服务（包括数据库、Redis等）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f cloud-platform-api
```

#### 4. 验证部署
```bash
# 检查API服务
curl http://localhost:8080/api/v1/health

# 检查数据库连接
docker-compose exec cloud-platform-api go run main.go --test-db

# 检查Redis连接
docker-compose exec cloud-platform-api go run main.go --test-redis
```

#### 5. 管理服务
```bash
# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 更新服务
docker-compose pull
docker-compose up -d

# 查看资源使用情况
docker-compose top

# 清理资源
docker-compose down -v
```

### 方法二：Kubernetes部署

#### 1. 环境准备
```bash
# 安装kubectl
# 下载地址：https://kubernetes.io/docs/tasks/tools/

# 安装Helm（可选）
# 下载地址：https://helm.sh/docs/intro/install/

# 验证安装
kubectl version --client
helm version
```

#### 2. 配置Kubernetes集群
```bash
# 创建命名空间
kubectl apply -f k8s/namespace.yaml

# 创建配置映射
kubectl apply -f k8s/configmap.yaml

# 创建密钥
kubectl apply -f k8s/secret.yaml

# 创建RBAC
kubectl apply -f k8s/rbac.yaml
```

#### 3. 部署应用
```bash
# 部署应用
kubectl apply -f k8s/deployment.yaml

# 创建服务
kubectl apply -f k8s/service.yaml

# 创建入口
kubectl apply -f k8s/ingress.yaml

# 配置网络策略
kubectl apply -f k8s/networkpolicy.yaml
```

#### 4. 配置自动扩缩容
```bash
# 创建HPA
kubectl apply -f k8s/hpa.yaml

# 创建PDB
kubectl apply -f k8s/pdb.yaml
```

#### 5. 验证部署
```bash
# 查看Pod状态
kubectl get pods -n cloud-platform

# 查看服务状态
kubectl get svc -n cloud-platform

# 查看入口状态
kubectl get ingress -n cloud-platform

# 查看日志
kubectl logs -f deployment/cloud-platform-api -n cloud-platform
```

#### 6. 管理部署
```bash
# 更新部署
kubectl set image deployment/cloud-platform-api cloud-platform-api=your-registry/cloud-platform-api:v1.1.0 -n cloud-platform

# 滚动更新
kubectl rollout status deployment/cloud-platform-api -n cloud-platform

# 回滚
kubectl rollout undo deployment/cloud-platform-api -n cloud-platform

# 扩缩容
kubectl scale deployment cloud-platform-api --replicas=3 -n cloud-platform
```

### 方法三：传统部署

#### 1. 环境准备
```bash
# 安装Go 1.21+
# 下载地址：https://golang.org/dl/

# 安装MySQL/PostgreSQL
# MySQL: https://dev.mysql.com/downloads/
# PostgreSQL: https://www.postgresql.org/download/

# 安装Redis
# Redis: https://redis.io/download

# 安装Nginx
# Nginx: http://nginx.org/en/download.html
```

#### 2. 构建应用
```bash
# 克隆项目
git clone https://github.com/your-username/cloud-platform-api.git
cd cloud-platform-api

# 安装依赖
go mod download
go mod tidy

# 构建应用
make build

# 或者手动构建
go build -o cloud-platform-api main.go
```

#### 3. 配置数据库
```bash
# 创建数据库
mysql -u root -p
CREATE DATABASE cloud_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 运行迁移
./cloud-platform-api migrate
```

#### 4. 配置Nginx
```nginx
# /etc/nginx/sites-available/cloud-platform-api
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/v1/storage/download/ {
        alias /path/to/storage/app/public/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

#### 5. 配置SSL证书
```bash
# 使用Let's Encrypt
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com

# 或者使用自签名证书
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/ssl/private/cloud-platform-api.key \
    -out /etc/ssl/certs/cloud-platform-api.crt
```

#### 6. 配置系统服务
```ini
# /etc/systemd/system/cloud-platform-api.service
[Unit]
Description=Cloud Platform API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/cloud-platform-api
ExecStart=/opt/cloud-platform-api/cloud-platform-api
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

#### 7. 启动服务
```bash
# 启动应用服务
sudo systemctl enable cloud-platform-api
sudo systemctl start cloud-platform-api

# 启动Nginx
sudo systemctl enable nginx
sudo systemctl start nginx

# 检查状态
sudo systemctl status cloud-platform-api
sudo systemctl status nginx
```

### 方法四：云服务部署

#### AWS部署
```bash
# 使用AWS ECS
aws ecs create-cluster --cluster-name cloud-platform-api
aws ecs register-task-definition --cli-input-json file://task-definition.json
aws ecs create-service --cluster cloud-platform-api --service-name cloud-platform-api --task-definition cloud-platform-api

# 使用AWS EKS
eksctl create cluster --name cloud-platform-api --region us-west-2
kubectl apply -f k8s/
```

#### 阿里云部署
```bash
# 使用阿里云容器服务
aliyun ecs CreateInstance --ImageId ubuntu_20_04_x64_20G_alibase_20210318.vhd
aliyun ecs StartInstance --InstanceId i-xxx

# 使用阿里云ACK
aliyun cs CreateCluster --name cloud-platform-api --region cn-hangzhou
```

#### 腾讯云部署
```bash
# 使用腾讯云TKE
tencentcloud tke CreateCluster --ClusterName cloud-platform-api --Region ap-beijing
tencentcloud tke CreateNodePool --ClusterId cls-xxx --NodePoolName cloud-platform-api
```

### 生产环境最佳实践

#### 1. 安全配置
```bash
# 修改默认密码和密钥
# 启用HTTPS
# 配置防火墙
# 设置访问控制
# 启用审计日志
```

#### 2. 性能优化
```bash
# 配置连接池
# 启用缓存
# 优化数据库查询
# 配置CDN
# 启用压缩
```

#### 3. 监控和告警
```bash
# 配置Prometheus监控
# 设置Grafana仪表板
# 配置告警规则
# 启用日志聚合
# 设置健康检查
```

#### 4. 备份和恢复
```bash
# 配置数据库备份
# 设置文件备份
# 测试恢复流程
# 配置异地备份
# 设置备份监控
```

#### 5. 高可用配置
```bash
# 配置负载均衡
# 设置多实例部署
# 配置数据库主从
# 设置故障转移
# 配置自动恢复
```

### 部署检查清单

#### 部署前检查
- [ ] 环境变量配置正确
- [ ] 数据库连接正常
- [ ] Redis连接正常
- [ ] 文件存储权限正确
- [ ] SSL证书配置正确
- [ ] 防火墙规则配置正确

#### 部署后检查
- [ ] 应用启动正常
- [ ] 健康检查通过
- [ ] API接口正常
- [ ] 数据库迁移完成
- [ ] 监控系统正常
- [ ] 日志记录正常

#### 性能检查
- [ ] 响应时间正常
- [ ] 内存使用正常
- [ ] CPU使用正常
- [ ] 磁盘空间充足
- [ ] 网络连接正常
- [ ] 缓存命中率正常

### 故障排除

#### 常见问题
1. **应用启动失败**
   - 检查配置文件
   - 检查端口占用
   - 检查权限设置
   - 查看错误日志

2. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接配置
   - 检查网络连接
   - 验证用户权限

3. **Redis连接失败**
   - 检查Redis服务状态
   - 验证连接配置
   - 检查网络连接
   - 验证认证信息

4. **文件上传失败**
   - 检查存储目录权限
   - 验证文件大小限制
   - 检查文件类型限制
   - 查看错误日志

#### 调试方法
```bash
# 查看应用日志
docker-compose logs -f cloud-platform-api

# 进入容器调试
docker-compose exec cloud-platform-api bash

# 查看系统资源
docker stats

# 查看网络连接
netstat -tulpn | grep :8080

# 查看进程状态
ps aux | grep cloud-platform-api
```

## 🧪 测试指南

### 测试框架概览

项目使用Go内置的测试框架，支持多种测试类型：

| 测试类型 | 文件命名 | 用途 | 运行频率 |
|----------|----------|------|----------|
| 单元测试 | `*_test.go` | 测试单个函数/方法 | 每次提交 |
| 集成测试 | `*_integration_test.go` | 测试模块间交互 | 每次构建 |
| 性能测试 | `*_benchmark_test.go` | 测试性能指标 | 定期运行 |
| 端到端测试 | `*_e2e_test.go` | 测试完整流程 | 发布前 |

### 运行测试

#### 1. 运行所有测试
```bash
# 运行所有测试
go test ./...

# 运行测试并显示详细信息
go test -v ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### 2. 运行特定测试
```bash
# 运行特定包的测试
go test ./tests/User/

# 运行特定测试函数
go test -run TestUserService ./tests/User/

# 运行包含特定字符串的测试
go test -run "TestUser.*" ./tests/

# 运行特定测试文件
go test ./tests/User/user_test.go
```

#### 3. 运行集成测试
```bash
# 运行集成测试
go test -tags=integration ./tests/

# 运行集成测试并显示详细信息
go test -tags=integration -v ./tests/
```

#### 4. 运行性能测试
```bash
# 运行性能测试
go test -bench=. ./tests/

# 运行性能测试并显示内存分配
go test -bench=. -benchmem ./tests/

# 运行性能测试并生成CPU profile
go test -bench=. -cpuprofile=cpu.prof ./tests/

# 运行性能测试并生成内存 profile
go test -bench=. -memprofile=mem.prof ./tests/
```

### 测试配置

#### 1. 测试环境配置
```bash
# 设置测试环境变量
export GIN_MODE=test
export DB_DRIVER=sqlite
export DB_DATABASE=:memory:
export REDIS_HOST=
export LOG_LEVEL=error
```

#### 2. 测试数据库配置
```go
// 测试数据库配置
func setupTestDB() *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to test database:", err)
    }
    
    // 运行迁移
    db.AutoMigrate(&User{}, &Post{}, &Category{}, &Tag{})
    
    return db
}
```

#### 3. 测试数据准备
```go
// 测试数据工厂
func createTestUser() *User {
    return &User{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }
}

func createTestPost() *Post {
    return &Post{
        Title:   "Test Post",
        Content: "This is a test post",
        Status:  "published",
    }
}
```

### 测试类型详解

#### 1. 单元测试示例
```go
// user_test.go
package user

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser(t *testing.T) {
    // 准备测试数据
    userService := NewUserService(setupTestDB())
    userData := &CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }

    // 执行测试
    user, err := userService.CreateUser(userData)

    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "testuser", user.Username)
    assert.Equal(t, "test@example.com", user.Email)
    assert.NotEmpty(t, user.ID)
}

func TestUserService_GetUserByID(t *testing.T) {
    // 准备测试数据
    userService := NewUserService(setupTestDB())
    user, _ := userService.CreateUser(&CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    })

    // 执行测试
    foundUser, err := userService.GetUserByID(user.ID)

    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, foundUser)
    assert.Equal(t, user.ID, foundUser.ID)
    assert.Equal(t, "testuser", foundUser.Username)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
    // 准备测试数据
    userService := NewUserService(setupTestDB())

    // 执行测试
    user, err := userService.GetUserByID(999)

    // 验证结果
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Contains(t, err.Error(), "user not found")
}
```

#### 2. 集成测试示例
```go
// user_integration_test.go
package user

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestUserController_Register(t *testing.T) {
    // 准备测试环境
    gin.SetMode(gin.TestMode)
    router := gin.New()
    userController := NewUserController(setupTestDB())
    router.POST("/api/v1/auth/register", userController.Register)

    // 准备测试数据
    requestBody := `{
        "username": "testuser",
        "email": "test@example.com",
        "password": "password123",
        "password_confirmation": "password123"
    }`

    // 执行测试
    req, _ := http.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(requestBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // 验证结果
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "用户注册成功", response["message"])
    assert.NotNil(t, response["data"])
}
```

#### 3. 性能测试示例
```go
// user_benchmark_test.go
package user

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func BenchmarkUserService_CreateUser(b *testing.B) {
    userService := NewUserService(setupTestDB())
    userData := &CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        userData.Username = fmt.Sprintf("testuser%d", i)
        userData.Email = fmt.Sprintf("test%d@example.com", i)
        user, err := userService.CreateUser(userData)
        assert.NoError(b, err)
        assert.NotNil(b, user)
    }
}

func BenchmarkUserService_GetUserByID(b *testing.B) {
    userService := NewUserService(setupTestDB())
    user, _ := userService.CreateUser(&CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    })

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        foundUser, err := userService.GetUserByID(user.ID)
        assert.NoError(b, err)
        assert.NotNil(b, foundUser)
    }
}
```

### 测试工具和库

#### 1. 测试断言库
```go
// 使用testify/assert
import "github.com/stretchr/testify/assert"

// 基本断言
assert.Equal(t, expected, actual)
assert.NotEqual(t, expected, actual)
assert.True(t, condition)
assert.False(t, condition)
assert.Nil(t, value)
assert.NotNil(t, value)
assert.Error(t, err)
assert.NoError(t, err)
assert.Contains(t, str, substr)
assert.NotContains(t, str, substr)
```

#### 2. 测试模拟库
```go
// 使用testify/mock
import "github.com/stretchr/testify/mock"

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(user *User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uint) (*User, error) {
    args := m.Called(id)
    return args.Get(0).(*User), args.Error(1)
}
```

#### 3. 测试数据生成
```go
// 使用faker生成测试数据
import "github.com/jaswdr/faker"

func generateTestUser() *User {
    f := faker.New()
    return &User{
        Username: f.Person().Name(),
        Email:    f.Internet().Email(),
        Password: f.Internet().Password(),
    }
}
```

### 测试最佳实践

#### 1. 测试命名规范
```go
// 测试函数命名：Test[FunctionName]_[Scenario]_[ExpectedResult]
func TestUserService_CreateUser_ValidData_ReturnsUser(t *testing.T) {}
func TestUserService_CreateUser_InvalidEmail_ReturnsError(t *testing.T) {}
func TestUserService_CreateUser_DuplicateEmail_ReturnsError(t *testing.T) {}
```

#### 2. 测试结构
```go
func TestFunction(t *testing.T) {
    // 1. 准备测试数据 (Arrange)
    setupTestData()
    
    // 2. 执行测试 (Act)
    result, err := functionUnderTest()
    
    // 3. 验证结果 (Assert)
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

#### 3. 测试隔离
```go
// 每个测试都应该独立运行
func TestUserService_CreateUser(t *testing.T) {
    // 使用独立的测试数据库
    db := setupTestDB()
    defer cleanupTestDB(db)
    
    // 测试逻辑
}
```

#### 4. 测试覆盖率
```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 查看覆盖率统计
go tool cover -func=coverage.out
```

### 持续集成测试

#### 1. GitHub Actions配置
```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v1
      with:
        file: ./coverage.out
```

#### 2. 测试脚本
```bash
#!/bin/bash
# scripts/run_tests.sh

set -e

echo "Running tests..."

# 设置测试环境
export GIN_MODE=test
export DB_DRIVER=sqlite
export DB_DATABASE=:memory:
export REDIS_HOST=
export LOG_LEVEL=error

# 运行单元测试
echo "Running unit tests..."
go test -v -coverprofile=coverage.out ./...

# 运行集成测试
echo "Running integration tests..."
go test -tags=integration -v ./tests/

# 运行性能测试
echo "Running benchmark tests..."
go test -bench=. ./tests/

# 生成覆盖率报告
echo "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo "Tests completed successfully!"
```

### 测试数据管理

#### 1. 测试数据库
```go
// 使用内存数据库进行测试
func setupTestDB() *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to test database:", err)
    }
    
    // 运行迁移
    db.AutoMigrate(&User{}, &Post{}, &Category{}, &Tag{})
    
    return db
}
```

#### 2. 测试数据清理
```go
func cleanupTestDB(db *gorm.DB) {
    db.Exec("DELETE FROM users")
    db.Exec("DELETE FROM posts")
    db.Exec("DELETE FROM categories")
    db.Exec("DELETE FROM tags")
}
```

#### 3. 测试数据工厂
```go
type TestDataFactory struct {
    db *gorm.DB
}

func NewTestDataFactory(db *gorm.DB) *TestDataFactory {
    return &TestDataFactory{db: db}
}

func (f *TestDataFactory) CreateUser(overrides ...func(*User)) *User {
    user := &User{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }
    
    for _, override := range overrides {
        override(user)
    }
    
    f.db.Create(user)
    return user
}

func (f *TestDataFactory) CreatePost(overrides ...func(*Post)) *Post {
    post := &Post{
        Title:   "Test Post",
        Content: "This is a test post",
        Status:  "published",
    }
    
    for _, override := range overrides {
        override(post)
    }
    
    f.db.Create(post)
    return post
}
```

### 测试监控和报告

#### 1. 测试结果报告
```bash
# 生成测试报告
go test -v -json ./... > test-results.json

# 生成JUnit格式报告
go test -v -json ./... | go-junit-report > test-results.xml
```

#### 2. 性能监控
```bash
# 生成CPU profile
go test -bench=. -cpuprofile=cpu.prof ./tests/
go tool pprof cpu.prof

# 生成内存 profile
go test -bench=. -memprofile=mem.prof ./tests/
go tool pprof mem.prof
```

#### 3. 测试质量指标
- **测试覆盖率**: 目标 > 80%
- **测试通过率**: 目标 100%
- **测试执行时间**: 目标 < 5分钟
- **性能回归**: 目标 < 5%

## 📈 性能优化指南

### 性能优化概览

| 优化类型 | 影响范围 | 优化效果 | 实施难度 | 优先级 |
|----------|----------|----------|----------|--------|
| 数据库优化 | 高 | 高 | 中 | ⭐⭐⭐⭐⭐ |
| 缓存优化 | 高 | 高 | 低 | ⭐⭐⭐⭐⭐ |
| 并发优化 | 中 | 高 | 中 | ⭐⭐⭐⭐ |
| 内存优化 | 中 | 中 | 低 | ⭐⭐⭐ |
| 网络优化 | 中 | 中 | 低 | ⭐⭐⭐ |

### 数据库优化

#### 1. 连接池配置优化
```go
// 数据库连接池配置
func configureDB() *gorm.DB {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        // 连接池配置
        ConnPool: &sql.DB{
            MaxOpenConns:    100,  // 最大打开连接数
            MaxIdleConns:    10,   // 最大空闲连接数
            ConnMaxLifetime: time.Hour, // 连接最大生存时间
            ConnMaxIdleTime: time.Minute * 30, // 连接最大空闲时间
        },
    })
    
    return db
}
```

#### 2. 查询性能优化
```go
// 使用索引优化查询
func GetUsersByEmail(email string) ([]User, error) {
    var users []User
    
    // 使用索引字段查询
    err := db.Where("email = ?", email).Find(&users).Error
    if err != nil {
        return nil, err
    }
    
    return users, nil
}

// 使用预加载减少N+1查询
func GetUsersWithPosts() ([]User, error) {
    var users []User
    
    err := db.Preload("Posts").Find(&users).Error
    if err != nil {
        return nil, err
    }
    
    return users, nil
}

// 使用分页减少内存使用
func GetUsersPaginated(page, limit int) ([]User, error) {
    var users []User
    
    offset := (page - 1) * limit
    err := db.Limit(limit).Offset(offset).Find(&users).Error
    if err != nil {
        return nil, err
    }
    
    return users, nil
}
```

#### 3. 索引优化建议
```sql
-- 为常用查询字段创建索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_created_at ON posts(created_at);

-- 创建复合索引
CREATE INDEX idx_posts_status_created_at ON posts(status, created_at);
CREATE INDEX idx_posts_user_id_status ON posts(user_id, status);

-- 创建部分索引
CREATE INDEX idx_posts_published ON posts(created_at) WHERE status = 'published';
```

#### 4. 慢查询监控
```go
// 慢查询监控中间件
func SlowQueryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        if duration > time.Second {
            log.Printf("Slow query detected: %s %s took %v", 
                c.Request.Method, c.Request.URL.Path, duration)
        }
    }
}
```

### 缓存优化

#### 1. Redis缓存策略
```go
// 缓存服务
type CacheService struct {
    redis *redis.Client
}

func (c *CacheService) Get(key string) (string, error) {
    return c.redis.Get(context.Background(), key).Result()
}

func (c *CacheService) Set(key string, value interface{}, expiration time.Duration) error {
    return c.redis.Set(context.Background(), key, value, expiration).Err()
}

func (c *CacheService) Delete(key string) error {
    return c.redis.Del(context.Background(), key).Err()
}

// 缓存装饰器
func CacheDecorator(cache *CacheService, key string, expiration time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 尝试从缓存获取
        if cached, err := cache.Get(key); err == nil {
            c.JSON(200, gin.H{"data": cached, "from_cache": true})
            c.Abort()
            return
        }
        
        // 缓存未命中，继续处理
        c.Next()
        
        // 将结果存入缓存
        if c.Writer.Status() == 200 {
            response := c.Writer.Header().Get("X-Cache-Data")
            cache.Set(key, response, expiration)
        }
    }
}
```

#### 2. 内存缓存降级
```go
// 内存缓存服务
type MemoryCache struct {
    cache map[string]interface{}
    mutex sync.RWMutex
}

func (m *MemoryCache) Get(key string) (interface{}, bool) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    value, exists := m.cache[key]
    return value, exists
}

func (m *MemoryCache) Set(key string, value interface{}) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    m.cache[key] = value
}

// 多级缓存
func (s *Service) GetUserWithCache(userID uint) (*User, error) {
    // 1. 尝试从内存缓存获取
    if user, exists := s.memoryCache.Get(fmt.Sprintf("user:%d", userID)); exists {
        return user.(*User), nil
    }
    
    // 2. 尝试从Redis缓存获取
    if cached, err := s.redisCache.Get(fmt.Sprintf("user:%d", userID)); err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        s.memoryCache.Set(fmt.Sprintf("user:%d", userID), &user)
        return &user, nil
    }
    
    // 3. 从数据库获取
    user, err := s.GetUserFromDB(userID)
    if err != nil {
        return nil, err
    }
    
    // 4. 存入缓存
    s.memoryCache.Set(fmt.Sprintf("user:%d", userID), user)
    userJSON, _ := json.Marshal(user)
    s.redisCache.Set(fmt.Sprintf("user:%d", userID), string(userJSON), time.Hour)
    
    return user, nil
}
```

#### 3. 缓存预热
```go
// 缓存预热服务
func (s *Service) WarmupCache() error {
    // 预热热门用户数据
    users, err := s.GetHotUsers()
    if err != nil {
        return err
    }
    
    for _, user := range users {
        userJSON, _ := json.Marshal(user)
        s.redisCache.Set(fmt.Sprintf("user:%d", user.ID), string(userJSON), time.Hour)
    }
    
    // 预热热门文章数据
    posts, err := s.GetHotPosts()
    if err != nil {
        return err
    }
    
    for _, post := range posts {
        postJSON, _ := json.Marshal(post)
        s.redisCache.Set(fmt.Sprintf("post:%d", post.ID), string(postJSON), time.Hour)
    }
    
    return nil
}
```

#### 4. 缓存清理策略
```go
// 缓存清理服务
func (s *Service) CleanupCache() error {
    // 清理过期缓存
    keys, err := s.redisCache.Keys("user:*")
    if err != nil {
        return err
    }
    
    for _, key := range keys {
        if s.redisCache.TTL(key) < time.Minute {
            s.redisCache.Delete(key)
        }
    }
    
    return nil
}
```

### 并发优化

#### 1. Goroutine池
```go
// Goroutine池
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    workerPool chan chan Job
    quit       chan bool
}

type Job struct {
    ID   int
    Data interface{}
}

func NewWorkerPool(workers int, jobQueue chan Job) *WorkerPool {
    return &WorkerPool{
        workers:    workers,
        jobQueue:   jobQueue,
        workerPool: make(chan chan Job, workers),
        quit:       make(chan bool),
    }
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        worker := NewWorker(p.workerPool)
        worker.Start()
    }
    
    go p.dispatch()
}

func (p *WorkerPool) dispatch() {
    for {
        select {
        case job := <-p.jobQueue:
            go func(job Job) {
                jobChannel := <-p.workerPool
                jobChannel <- job
            }(job)
        case <-p.quit:
            return
        }
    }
}
```

#### 2. 连接复用
```go
// HTTP客户端连接池
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}

// 数据库连接复用
func (s *Service) GetUserWithConnection(userID uint) (*User, error) {
    // 使用连接池中的连接
    db := s.db.WithContext(context.Background())
    
    var user User
    err := db.First(&user, userID).Error
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}
```

#### 3. 异步处理
```go
// 异步任务处理
type AsyncTask struct {
    ID   string
    Data interface{}
}

func (s *Service) ProcessAsync(task AsyncTask) {
    go func() {
        // 异步处理任务
        err := s.processTask(task)
        if err != nil {
            log.Printf("Async task failed: %v", err)
        }
    }()
}

// 批量处理
func (s *Service) ProcessBatch(tasks []AsyncTask) {
    var wg sync.WaitGroup
    
    for _, task := range tasks {
        wg.Add(1)
        go func(task AsyncTask) {
            defer wg.Done()
            s.processTask(task)
        }(task)
    }
    
    wg.Wait()
}
```

### 内存优化

#### 1. 对象池
```go
// 对象池
var userPool = sync.Pool{
    New: func() interface{} {
        return &User{}
    },
}

func GetUser() *User {
    return userPool.Get().(*User)
}

func PutUser(user *User) {
    // 重置对象状态
    user.ID = 0
    user.Username = ""
    user.Email = ""
    user.Password = ""
    
    userPool.Put(user)
}
```

#### 2. 内存监控
```go
// 内存监控
func MonitorMemory() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    log.Printf("Memory usage: %d KB", m.Alloc/1024)
    log.Printf("GC cycles: %d", m.NumGC)
    log.Printf("GC pause: %v", time.Duration(m.PauseTotalNs))
}
```

#### 3. 垃圾回收优化
```go
// 垃圾回收优化
func OptimizeGC() {
    // 设置GC目标百分比
    debug.SetGCPercent(100)
    
    // 手动触发GC
    runtime.GC()
    
    // 设置内存限制
    debug.SetMemoryLimit(1024 * 1024 * 1024) // 1GB
}
```

### 网络优化

#### 1. HTTP/2支持
```go
// HTTP/2服务器配置
func createHTTPServer() *http.Server {
    return &http.Server{
        Addr:    ":8080",
        Handler: router,
        // 启用HTTP/2
        TLSConfig: &tls.Config{
            NextProtos: []string{"h2", "http/1.1"},
        },
    }
}
```

#### 2. 压缩优化
```go
// 压缩中间件
func CompressionMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查客户端是否支持压缩
        if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
            c.Header("Content-Encoding", "gzip")
            c.Header("Vary", "Accept-Encoding")
            
            gz := gzip.NewWriter(c.Writer)
            defer gz.Close()
            
            c.Writer = &gzipResponseWriter{Writer: gz, ResponseWriter: c.Writer}
        }
        
        c.Next()
    }
}
```

#### 3. 连接优化
```go
// 连接优化配置
func optimizeConnections() {
    // 设置TCP keep-alive
    net.ListenConfig{
        KeepAlive: 30 * time.Second,
    }
    
    // 设置连接超时
    net.Dialer{
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
    }
}
```

### 性能监控

#### 1. 性能指标收集
```go
// 性能指标
type PerformanceMetrics struct {
    RequestCount    int64
    ResponseTime    time.Duration
    ErrorCount      int64
    MemoryUsage     uint64
    GoroutineCount  int
}

func (m *PerformanceMetrics) RecordRequest(duration time.Duration) {
    atomic.AddInt64(&m.RequestCount, 1)
    atomic.StoreInt64((*int64)(&m.ResponseTime), int64(duration))
}
```

#### 2. 性能分析
```go
// 性能分析
func StartProfiling() {
    // CPU profiling
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
    
    // Memory profiling
    f2, _ := os.Create("mem.prof")
    defer f2.Close()
    runtime.GC()
    pprof.WriteHeapProfile(f2)
}
```

#### 3. 性能告警
```go
// 性能告警
func (s *Service) CheckPerformance() {
    if s.metrics.ResponseTime > time.Second {
        log.Printf("Performance warning: response time %v", s.metrics.ResponseTime)
    }
    
    if s.metrics.ErrorCount > 100 {
        log.Printf("Error rate warning: %d errors", s.metrics.ErrorCount)
    }
}
```

### 性能测试

#### 1. 压力测试
```bash
# 使用wrk进行压力测试
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/health

# 使用ab进行压力测试
ab -n 10000 -c 100 http://localhost:8080/api/v1/health
```

#### 2. 性能基准测试
```go
// 性能基准测试
func BenchmarkUserService_CreateUser(b *testing.B) {
    service := NewUserService(setupTestDB())
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        user := &CreateUserRequest{
            Username: fmt.Sprintf("user%d", i),
            Email:    fmt.Sprintf("user%d@example.com", i),
            Password: "password123",
        }
        
        _, err := service.CreateUser(user)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

#### 3. 性能回归测试
```bash
# 运行性能测试
go test -bench=. -benchmem ./tests/

# 比较性能结果
go test -bench=. -benchmem -count=5 ./tests/ > current.txt
benchcmp previous.txt current.txt
```

### 性能优化检查清单

#### 数据库优化
- [ ] 连接池配置合理
- [ ] 查询使用索引
- [ ] 避免N+1查询
- [ ] 使用分页查询
- [ ] 监控慢查询

#### 缓存优化
- [ ] 实现多级缓存
- [ ] 设置合理的过期时间
- [ ] 实现缓存预热
- [ ] 监控缓存命中率
- [ ] 实现缓存清理

#### 并发优化
- [ ] 使用Goroutine池
- [ ] 实现连接复用
- [ ] 使用异步处理
- [ ] 避免竞态条件
- [ ] 监控Goroutine数量

#### 内存优化
- [ ] 使用对象池
- [ ] 避免内存泄漏
- [ ] 监控内存使用
- [ ] 优化垃圾回收
- [ ] 使用内存分析工具

#### 网络优化
- [ ] 启用HTTP/2
- [ ] 使用压缩
- [ ] 优化连接配置
- [ ] 使用CDN
- [ ] 监控网络延迟

## 👨‍💻 开发指南

### 开发环境设置

#### 1. 环境要求
```bash
# Go 1.21+
go version

# Git
git --version

# Docker (可选)
docker --version

# 代码编辑器推荐
# - VS Code with Go extension
# - GoLand
# - Vim/Neovim with vim-go
```

#### 2. 项目结构
```
cloud-platform-api/
├── app/                    # 应用核心代码
│   ├── Config/            # 配置管理
│   ├── Controllers/       # 控制器
│   ├── Middleware/        # 中间件
│   ├── Models/            # 数据模型
│   ├── Services/          # 业务逻辑
│   └── Utils/             # 工具函数
├── bootstrap/             # 启动配置
├── docs/                  # 文档
├── k8s/                   # Kubernetes配置
├── scripts/               # 脚本
├── storage/               # 存储
├── tests/                 # 测试
└── main.go               # 入口文件
```

#### 3. 开发工作流
```bash
# 1. 克隆项目
git clone https://github.com/your-username/cloud-platform-api.git
cd cloud-platform-api

# 2. 创建开发分支
git checkout -b feature/your-feature-name

# 3. 安装依赖
go mod download
go mod tidy

# 4. 运行测试
go test ./...

# 5. 启动开发服务器
make dev

# 6. 提交代码
git add .
git commit -m "feat: add new feature"
git push origin feature/your-feature-name

# 7. 创建Pull Request
```

### 代码规范

#### 1. Go代码规范
```go
// 包注释
// Package user provides user management functionality.
package user

// 函数注释
// CreateUser creates a new user with the given data.
// It returns the created user and any error encountered.
func CreateUser(data *CreateUserRequest) (*User, error) {
    // 实现代码
}

// 变量命名
var (
    // 常量使用大写字母
    DefaultPageSize = 10
    
    // 变量使用驼峰命名
    userService *UserService
)

// 结构体注释
// User represents a user in the system.
type User struct {
    ID       uint   `json:"id" gorm:"primaryKey"`
    Username string `json:"username" gorm:"uniqueIndex"`
    Email    string `json:"email" gorm:"uniqueIndex"`
}
```

#### 2. 错误处理
```go
// 使用自定义错误类型
type UserError struct {
    Code    string
    Message string
    Err     error
}

func (e *UserError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

// 错误包装
func GetUser(id uint) (*User, error) {
    user, err := userRepo.GetByID(id)
    if err != nil {
        return nil, &UserError{
            Code:    "USER_NOT_FOUND",
            Message: "User not found",
            Err:     err,
        }
    }
    return user, nil
}
```

#### 3. 日志记录
```go
// 使用结构化日志
log.Info("User created successfully",
    zap.String("user_id", user.ID),
    zap.String("username", user.Username),
    zap.String("email", user.Email),
)

// 错误日志
log.Error("Failed to create user",
    zap.Error(err),
    zap.String("username", username),
    zap.String("email", email),
)
```

### 测试开发

#### 1. 单元测试
```go
func TestUserService_CreateUser(t *testing.T) {
    // 准备测试数据
    service := NewUserService(mockDB)
    userData := &CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }

    // 执行测试
    user, err := service.CreateUser(userData)

    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "testuser", user.Username)
}
```

#### 2. 集成测试
```go
func TestUserController_Register(t *testing.T) {
    // 设置测试环境
    gin.SetMode(gin.TestMode)
    router := setupTestRouter()
    
    // 准备测试数据
    requestBody := `{
        "username": "testuser",
        "email": "test@example.com",
        "password": "password123"
    }`
    
    // 执行测试
    req, _ := http.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(requestBody))
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // 验证结果
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### 数据库开发

#### 1. 迁移文件
```go
// CreateUsersTable.go
package Migrations

import (
    "gorm.io/gorm"
)

func CreateUsersTable(db *gorm.DB) error {
    return db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INT AUTO_INCREMENT PRIMARY KEY,
            username VARCHAR(255) UNIQUE NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            password VARCHAR(255) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
        )
    `).Error
}
```

#### 2. 模型定义
```go
// User.go
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Username  string    `json:"username" gorm:"uniqueIndex;size:255"`
    Email     string    `json:"email" gorm:"uniqueIndex;size:255"`
    Password  string    `json:"-" gorm:"size:255"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    // 关联
    Posts []Post `json:"posts,omitempty" gorm:"foreignKey:UserID"`
}
```

### API开发

#### 1. 控制器开发
```go
// UserController.go
type UserController struct {
    userService *UserService
}

func NewUserController(userService *UserService) *UserController {
    return &UserController{userService: userService}
}

func (c *UserController) GetUsers(ctx *gin.Context) {
    // 获取查询参数
    page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
    
    // 调用服务
    users, err := c.userService.GetUsers(page, limit)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // 返回响应
    ctx.JSON(http.StatusOK, gin.H{
        "data": users,
        "page": page,
        "limit": limit,
    })
}
```

#### 2. 中间件开发
```go
// AuthMiddleware.go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
            c.Abort()
            return
        }
        
        // 验证token
        user, err := validateToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Set("user", user)
        c.Next()
    }
}
```

### 调试技巧

#### 1. 日志调试
```go
// 设置日志级别
log.SetLevel(log.DebugLevel)

// 调试日志
log.Debug("Processing user request",
    zap.String("user_id", userID),
    zap.String("action", "create"),
)
```

#### 2. 性能调试
```go
// 性能分析
func profileHandler(c *gin.Context) {
    // 启用CPU分析
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
    
    // 业务逻辑
    processRequest()
}
```

#### 3. 数据库调试
```go
// 启用SQL日志
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})
```

## 🔧 故障排除

### 常见问题

#### 1. 应用启动失败
**问题**: 应用无法启动
**解决方案**:
```bash
# 检查端口占用
netstat -tulpn | grep :8080

# 检查配置文件
cat .env

# 检查日志
tail -f storage/logs/app.log

# 检查依赖
go mod verify
```

#### 2. 数据库连接失败
**问题**: 无法连接到数据库
**解决方案**:
```bash
# 检查数据库服务状态
systemctl status mysql

# 测试数据库连接
mysql -u root -p -h localhost

# 检查连接配置
grep -E "DB_" .env

# 检查网络连接
telnet localhost 3306
```

#### 3. Redis连接失败
**问题**: 无法连接到Redis
**解决方案**:
```bash
# 检查Redis服务状态
systemctl status redis

# 测试Redis连接
redis-cli ping

# 检查Redis配置
redis-cli config get "*"

# 检查网络连接
telnet localhost 6379
```

#### 4. 文件上传失败
**问题**: 文件上传失败
**解决方案**:
```bash
# 检查存储目录权限
ls -la storage/app/public/

# 检查磁盘空间
df -h

# 检查文件大小限制
grep STORAGE_MAX_FILE_SIZE .env

# 检查文件类型限制
grep STORAGE_ALLOWED_TYPES .env
```

#### 5. 性能问题
**问题**: 应用响应缓慢
**解决方案**:
```bash
# 检查系统资源
top
htop
iostat

# 检查数据库性能
mysql -e "SHOW PROCESSLIST;"
mysql -e "SHOW STATUS LIKE 'Slow_queries';"

# 检查应用性能
go tool pprof http://localhost:8080/debug/pprof/profile
```

### 调试工具

#### 1. 日志分析
```bash
# 查看错误日志
grep "ERROR" storage/logs/app.log

# 查看访问日志
tail -f storage/logs/access/access.log

# 查看SQL日志
tail -f storage/logs/sql/sql.log
```

#### 2. 性能分析
```bash
# CPU分析
go tool pprof http://localhost:8080/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:8080/debug/pprof/heap

# 协程分析
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

#### 3. 网络调试
```bash
# 检查端口监听
netstat -tulpn | grep :8080

# 检查网络连接
ss -tulpn | grep :8080

# 测试API接口
curl -v http://localhost:8080/api/v1/health
```

### 监控和告警

#### 1. 健康检查
```bash
# 基本健康检查
curl http://localhost:8080/api/v1/health

# 详细健康检查
curl http://localhost:8080/api/v1/health/detailed

# 系统指标
curl http://localhost:8080/api/v1/metrics
```

#### 2. 日志监控
```bash
# 实时日志监控
tail -f storage/logs/app.log | grep ERROR

# 日志统计
grep -c "ERROR" storage/logs/app.log

# 日志分析
awk '{print $1}' storage/logs/access/access.log | sort | uniq -c
```

#### 3. 性能监控
```bash
# 系统资源监控
htop
iotop
nethogs

# 应用性能监控
go tool pprof http://localhost:8080/debug/pprof/profile
```

### 恢复和备份

#### 1. 数据备份
```bash
# 数据库备份
mysqldump -u root -p cloud_platform > backup.sql

# 文件备份
tar -czf storage_backup.tar.gz storage/

# 配置备份
cp .env .env.backup
```

#### 2. 数据恢复
```bash
# 数据库恢复
mysql -u root -p cloud_platform < backup.sql

# 文件恢复
tar -xzf storage_backup.tar.gz

# 配置恢复
cp .env.backup .env
```

#### 3. 应用回滚
```bash
# 停止应用
systemctl stop cloud-platform-api

# 回滚代码
git checkout previous-version

# 重新构建
make build

# 启动应用
systemctl start cloud-platform-api
```

## ⚠️ 重要安全注意事项

### 生产环境配置
1. **必须修改JWT密钥** - 使用至少32字符的强密钥
2. **启用所有安全防护** - XSS、SQL注入、CSRF等
3. **配置HTTPS** - 使用SSL证书
4. **设置强密码策略** - 密码复杂度要求
5. **启用审计日志** - 记录所有重要操作
6. **定期备份数据** - 自动备份策略
7. **监控系统状态** - 实时监控和告警

### 安全最佳实践
1. 定期更新依赖包
2. 使用环境变量管理敏感配置
3. 实施最小权限原则
4. 定期进行安全扫描
5. 建立安全事件响应流程

## 🤝 贡献指南

### 贡献流程
1. Fork项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建Pull Request

### 代码规范
- 遵循Go官方代码规范
- 使用有意义的提交信息
- 添加必要的测试
- 更新相关文档

### 问题报告
- 使用GitHub Issues报告问题
- 提供详细的错误信息和复现步骤
- 包含系统环境信息

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

如果您遇到问题或有建议，请：
1. 查看 [文档](docs/)
2. 搜索 [Issues](../../issues)
3. 创建新的Issue
4. 联系维护者

## 🔄 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 完整的用户认证系统
- 内容管理功能
- 文件存储功能
- 监控和日志系统
- API文档生成
- 安全防护机制
- 自动化测试框架

### 计划中的功能
- WebSocket实时通信
- 微服务架构支持
- 更多数据库支持
- 高级缓存策略
- 机器学习集成
