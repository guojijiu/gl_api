# Cloud Platform API 文档

## 概述

Cloud Platform API 是一个基于 Go 和 Gin 框架构建的云平台后端服务，提供用户管理、认证、监控等核心功能。

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin
- **数据库**: MySQL 8.0+
- **ORM**: GORM
- **缓存**: Redis (可选)
- **监控**: Prometheus + Grafana
- **日志**: 结构化日志系统

## 项目结构

```
cloud-platform-api/
├── app/
│   ├── Config/              # 配置管理
│   ├── Database/            # 数据库相关
│   ├── Http/               # HTTP层
│   │   ├── Controllers/    # 控制器
│   │   ├── Middleware/     # 中间件
│   │   └── Routes/         # 路由
│   ├── Models/             # 数据模型
│   ├── Services/           # 业务逻辑层
│   └── Utils/              # 工具函数
├── docs/                   # 文档
├── monitoring/             # 监控配置
├── tests/                  # 测试文件
└── storage/               # 存储目录
```

## 核心功能

### 1. 用户管理

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "SecurePass123!",
  "confirm_password": "SecurePass123!"
}
```

**响应:**
```json
{
  "success": true,
  "message": "用户注册成功",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "status": 1,
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "SecurePass123!"
}
```

**响应:**
```json
{
  "success": true,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "status": 1
    }
  }
}
```

#### 获取用户信息
```http
GET /api/v1/users/profile
Authorization: Bearer <token>
```

**响应:**
```json
{
  "success": true,
  "message": "获取用户信息成功",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "status": 1,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

### 2. 认证与授权

#### JWT Token 管理
- Token 有效期: 24小时
- 刷新机制: 支持 token 刷新
- 黑名单: 支持 token 黑名单管理

#### 权限控制
- 基于角色的访问控制 (RBAC)
- 中间件级别的权限验证
- API 级别的权限控制

### 3. 监控与日志

#### 健康检查
```http
GET /api/v1/health
```

**响应:**
```json
{
  "success": true,
  "message": "服务健康",
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T00:00:00Z",
    "version": "1.0.0",
    "uptime": "24h30m15s"
  }
}
```

#### 系统监控
```http
GET /api/v1/monitor/system
Authorization: Bearer <token>
```

**响应:**
```json
{
  "success": true,
  "message": "系统监控数据",
  "data": {
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 23.1,
    "request_count": 1250,
    "error_count": 5,
    "response_time": "150ms"
  }
}
```

## 中间件

### 1. 认证中间件
- JWT token 验证
- 用户身份确认
- 权限检查

### 2. 安全中间件
- CORS 跨域处理
- 请求大小限制
- SQL 注入防护
- XSS 攻击防护
- 速率限制

### 3. 性能中间件
- 请求统计
- 响应时间监控
- 内存使用监控
- 自动性能优化

### 4. 日志中间件
- 请求日志记录
- SQL 查询日志
- 错误日志记录
- 业务日志记录

## 错误处理

### 标准错误响应格式
```json
{
  "success": false,
  "message": "错误描述",
  "error": "ERROR_CODE",
  "details": {
    "field": "具体错误信息"
  }
}
```

### 常见错误码
- `400`: 请求参数错误
- `401`: 未授权
- `403`: 禁止访问
- `404`: 资源不存在
- `429`: 请求过于频繁
- `500`: 服务器内部错误

## 配置

### 环境变量
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=password
DB_DATABASE=cloud_platform

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=your-secret-key
JWT_EXPIRE_HOURS=24

# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
```

## 部署

### Docker 部署
```bash
# 构建镜像
docker build -t cloud-platform-api .

# 运行容器
docker run -d \
  --name cloud-platform-api \
  -p 8080:8080 \
  -e DB_HOST=mysql \
  -e REDIS_HOST=redis \
  cloud-platform-api
```

### Docker Compose 部署
```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## 测试

### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test ./tests/...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 测试覆盖
- 单元测试: 业务逻辑测试
- 集成测试: API 接口测试
- 安全测试: 安全漏洞测试
- 性能测试: 负载和压力测试

## 监控

### Prometheus 指标
- `http_requests_total`: HTTP 请求总数
- `http_request_duration_seconds`: 请求持续时间
- `database_connections_active`: 数据库连接数
- `cache_hits_total`: 缓存命中数

### Grafana 仪表板
- 系统性能监控
- 应用性能监控
- 数据库性能监控
- 错误率监控

## 安全

### 安全措施
1. **密码安全**: bcrypt 加密存储
2. **JWT 安全**: 安全的 token 生成和验证
3. **输入验证**: 严格的输入参数验证
4. **SQL 注入防护**: 参数化查询
5. **XSS 防护**: 输入过滤和输出编码
6. **CSRF 防护**: CSRF token 验证
7. **速率限制**: API 调用频率限制

### 安全建议
1. 定期更新依赖包
2. 使用 HTTPS 协议
3. 定期进行安全扫描
4. 监控异常访问模式
5. 实施最小权限原则

## 性能优化

### 数据库优化
1. 连接池配置优化
2. 查询优化和索引
3. 读写分离
4. 缓存策略

### 应用优化
1. 内存缓存
2. 响应压缩
3. 静态资源优化
4. 并发处理优化

## 故障排除

### 常见问题
1. **数据库连接失败**: 检查数据库配置和网络连接
2. **Redis 连接失败**: 检查 Redis 服务状态
3. **JWT 验证失败**: 检查 token 格式和密钥
4. **内存使用过高**: 检查缓存配置和内存泄漏

### 日志分析
```bash
# 查看应用日志
tail -f storage/logs/app.log

# 查看错误日志
tail -f storage/logs/error.log

# 查看业务日志
tail -f storage/logs/business.log
```

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 基础用户管理功能
- JWT 认证系统
- 监控和日志系统
- Docker 支持

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License

## 联系方式

- 项目地址: https://github.com/your-org/cloud-platform-api
- 问题反馈: https://github.com/your-org/cloud-platform-api/issues
- 邮箱: support@example.com
