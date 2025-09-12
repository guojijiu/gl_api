# 云平台API文档 - 增强版

## 概述

云平台API是一个功能完整的企业级API服务，提供用户管理、认证授权、监控告警、配置管理等核心功能。

## 版本信息

- **API版本**: 1.0.0
- **构建时间**: {{BUILD_TIME}}
- **Git提交**: {{GIT_COMMIT}}
- **最后更新**: {{LAST_UPDATE}}

## 基础信息

### 基础URL

- **开发环境**: `http://localhost:8080`
- **生产环境**: `https://api.cloudplatform.com`

### 认证方式

API支持多种认证方式：

1. **JWT Token认证**
   - Header: `Authorization: Bearer <token>`
   - 有效期: 24小时
   - 刷新: 支持token刷新

2. **API Key认证**
   - Header: `X-API-Key: <api_key>`
   - 适用于服务间调用

3. **Session认证**
   - Cookie: `session_id=<session_id>`
   - 适用于Web应用

### 响应格式

所有API响应都遵循统一格式：

```json
{
  "success": true,
  "message": "操作成功",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "req_123456789"
}
```

### 错误处理

错误响应格式：

```json
{
  "success": false,
  "message": "错误描述",
  "error": "ERROR_CODE",
  "details": {},
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "req_123456789"
}
```

## 核心功能

### 1. 健康检查

#### 基础健康检查
```http
GET /api/v1/health
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T00:00:00Z",
    "version": "1.0.0",
    "uptime": "2d 5h 30m 15s",
    "services": {
      "database": {
        "status": "healthy",
        "response_time": "5ms",
        "message": "OK"
      },
      "redis": {
        "status": "healthy",
        "response_time": "2ms",
        "message": "OK"
      }
    },
    "metrics": {
      "memory": {
        "alloc": 1024000,
        "total_alloc": 2048000,
        "sys": 4096000
      },
      "cpu": {
        "usage": 15.5
      },
      "goroutines": 150
    }
  }
}
```

#### 详细健康检查
```http
GET /api/v1/health/detailed
```

#### 就绪检查
```http
GET /api/v1/health/ready
```

#### 存活检查
```http
GET /api/v1/health/live
```

### 2. 用户认证

#### 用户登录
```http
POST /api/v1/auth/login
```

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "refresh_token_here",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "name": "张三",
      "email": "user@example.com",
      "role": "user"
    }
  }
}
```

#### 用户注册
```http
POST /api/v1/auth/register
```

#### 刷新Token
```http
POST /api/v1/auth/refresh
```

#### 用户登出
```http
POST /api/v1/auth/logout
```

### 3. 用户管理

#### 获取用户列表
```http
GET /api/v1/users?page=1&limit=10&search=张三
```

**查询参数**:
- `page`: 页码 (默认: 1)
- `limit`: 每页数量 (默认: 10, 最大: 100)
- `search`: 搜索关键词
- `role`: 角色筛选
- `status`: 状态筛选

#### 获取用户详情
```http
GET /api/v1/users/{id}
```

#### 创建用户
```http
POST /api/v1/users
```

**请求体**:
```json
{
  "name": "张三",
  "email": "user@example.com",
  "password": "password123",
  "role": "user"
}
```

#### 更新用户
```http
PUT /api/v1/users/{id}
```

#### 删除用户
```http
DELETE /api/v1/users/{id}
```

### 4. 监控告警

#### 获取监控状态
```http
GET /api/v1/monitoring/status
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "circuit_breakers": {
      "api": {
        "state": "closed",
        "failures": 0,
        "successes": 100,
        "requests": 100
      }
    },
    "performance_metrics": {
      "response_time": 150.5,
      "error_rate": 0.01,
      "cpu_usage": 25.5,
      "memory_usage": 60.2
    },
    "alert_thresholds": {
      "response_time": 1000,
      "error_rate": 0.05,
      "cpu_usage": 80.0,
      "memory_usage": 80.0
    }
  }
}
```

#### 获取告警历史
```http
GET /api/v1/monitoring/alerts?limit=50&severity=high
```

#### 解决告警
```http
POST /api/v1/monitoring/alerts/{id}/resolve
```

### 5. 配置管理

#### 获取配置信息
```http
GET /api/v1/config
```

#### 更新配置
```http
PUT /api/v1/config
```

#### 重载配置
```http
POST /api/v1/config/reload
```

### 6. API文档

#### 获取API文档
```http
GET /api/v1/docs
```

#### 获取Swagger UI
```http
GET /api/v1/docs/ui
```

#### 导出API文档
```http
GET /api/v1/docs/export
```

## 高级功能

### 1. 熔断器

系统内置熔断器保护，自动监控API调用失败率：

- **状态**: closed, open, half-open
- **阈值**: 可配置失败率和请求量阈值
- **恢复**: 自动恢复机制

### 2. 配置热重载

支持配置文件实时重载：

- **监控**: 自动监控配置文件变化
- **回调**: 支持重载回调函数
- **验证**: 配置验证和回滚

### 3. 国际化支持

支持多语言：

- **语言**: 中文(zh), 英文(en)
- **回退**: 自动回退到默认语言
- **动态**: 支持运行时语言切换

### 4. 性能监控

全面的性能监控：

- **指标**: 响应时间、错误率、资源使用
- **告警**: 可配置阈值告警
- **历史**: 指标历史数据

## 错误代码

| 代码 | 描述 | HTTP状态码 |
|------|------|------------|
| `AUTH_INVALID_CREDENTIALS` | 认证失败 | 401 |
| `AUTH_TOKEN_EXPIRED` | Token过期 | 401 |
| `AUTH_PERMISSION_DENIED` | 权限不足 | 403 |
| `VALIDATION_FAILED` | 验证失败 | 400 |
| `RESOURCE_NOT_FOUND` | 资源不存在 | 404 |
| `RESOURCE_ALREADY_EXISTS` | 资源已存在 | 409 |
| `RATE_LIMIT_EXCEEDED` | 请求频率过高 | 429 |
| `CIRCUIT_BREAKER_OPEN` | 熔断器开启 | 503 |
| `SYSTEM_ERROR` | 系统错误 | 500 |

## 限流策略

### 全局限流
- **限制**: 100请求/分钟
- **范围**: 所有API端点
- **响应**: 429 Too Many Requests

### 用户限流
- **限制**: 1000请求/小时
- **范围**: 单个用户
- **响应**: 429 Too Many Requests

### IP限流
- **限制**: 500请求/小时
- **范围**: 单个IP
- **响应**: 429 Too Many Requests

## 安全特性

### 1. 输入验证
- **类型检查**: 严格的数据类型验证
- **长度限制**: 字符串和数组长度限制
- **格式验证**: 邮箱、URL等格式验证

### 2. SQL注入防护
- **参数化查询**: 使用参数化查询防止SQL注入
- **输入过滤**: 特殊字符过滤和转义

### 3. XSS防护
- **输出编码**: 所有输出都进行HTML编码
- **CSP头**: 内容安全策略头

### 4. CSRF防护
- **Token验证**: CSRF Token验证
- **同源检查**: 检查请求来源

## 部署指南

### 环境要求
- **Go版本**: 1.21+
- **数据库**: PostgreSQL 12+
- **缓存**: Redis 6+
- **内存**: 最小2GB，推荐4GB+
- **CPU**: 最小2核，推荐4核+

### 配置文件
```yaml
server:
  port: 8080
  host: "0.0.0.0"
  timeout: 30s

database:
  host: "localhost"
  port: 5432
  name: "cloud_platform"
  user: "postgres"
  password: "password"
  ssl_mode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

monitoring:
  enabled: true
  metrics_interval: 30s
  alert_thresholds:
    response_time: 1000
    error_rate: 0.05
    cpu_usage: 80.0
    memory_usage: 80.0

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

### Docker部署
```bash
# 构建镜像
docker build -t cloud-platform-api .

# 运行容器
docker run -d \
  --name cloud-platform-api \
  -p 8080:8080 \
  -e DATABASE_URL=postgres://user:pass@host:5432/db \
  -e REDIS_URL=redis://host:6379 \
  cloud-platform-api
```

### Kubernetes部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloud-platform-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cloud-platform-api
  template:
    metadata:
      labels:
        app: cloud-platform-api
    spec:
      containers:
      - name: api
        image: cloud-platform-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: url
```

## 监控和告警

### 监控指标
- **响应时间**: 平均、P95、P99响应时间
- **错误率**: 4xx、5xx错误率
- **吞吐量**: 每秒请求数
- **资源使用**: CPU、内存、磁盘使用率

### 告警规则
- **响应时间**: > 1秒持续5分钟
- **错误率**: > 5%持续2分钟
- **CPU使用率**: > 80%持续5分钟
- **内存使用率**: > 80%持续5分钟

### 通知方式
- **邮件**: 支持HTML格式邮件通知
- **Webhook**: 支持自定义Webhook通知
- **日志**: 结构化日志记录

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接配置
   - 检查网络连通性

2. **Redis连接失败**
   - 检查Redis服务状态
   - 验证连接配置
   - 检查内存使用情况

3. **API响应慢**
   - 检查数据库查询性能
   - 查看缓存命中率
   - 分析系统资源使用

4. **熔断器频繁开启**
   - 检查下游服务状态
   - 调整熔断器阈值
   - 优化API性能

### 日志分析
```bash
# 查看错误日志
tail -f logs/errors/app.log | grep ERROR

# 查看访问日志
tail -f logs/access/app.log

# 查看性能日志
tail -f logs/performance/app.log
```

### 性能调优
1. **数据库优化**
   - 添加适当的索引
   - 优化查询语句
   - 调整连接池配置

2. **缓存优化**
   - 增加缓存命中率
   - 调整缓存过期时间
   - 优化缓存策略

3. **API优化**
   - 减少不必要的数据库查询
   - 使用异步处理
   - 优化响应格式

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 基础用户管理功能
- JWT认证支持
- 健康检查端点
- 基础监控功能

### v1.1.0 (2024-01-15)
- 添加熔断器支持
- 配置热重载功能
- 增强健康检查
- API文档生成
- 国际化支持

## 支持

- **文档**: [https://docs.cloudplatform.com](https://docs.cloudplatform.com)
- **问题反馈**: [https://github.com/cloudplatform/api/issues](https://github.com/cloudplatform/api/issues)
- **技术支持**: support@cloudplatform.com
- **紧急联系**: +86-400-123-4567

---

*最后更新: 2024-01-01*
*文档版本: 1.1.0*
