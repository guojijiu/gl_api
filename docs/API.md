# Cloud Platform API 文档

## 📋 概述

Cloud Platform API 是一个基于 Go + Gin + GORM 构建的现代化云平台API服务，提供完整的用户管理、内容管理、文件存储、实时通信、性能监控、安全防护等功能。

## 🚀 核心特性

- **用户管理**: 完整的用户注册、登录、权限管理
- **内容管理**: 文章、分类、标签管理
- **文件存储**: 安全的文件上传、下载、管理
- **实时通信**: WebSocket实时消息推送
- **性能监控**: 系统资源、应用性能、业务指标监控
- **安全防护**: 多层安全防护、威胁检测、审计日志
- **查询优化**: 数据库性能监控和优化建议
- **缓存管理**: Redis缓存和内存缓存

## 📊 基础信息

- **基础URL**: `http://localhost:8080`
- **API版本**: v1
- **认证方式**: Bearer Token (JWT)
- **数据格式**: JSON
- **支持协议**: HTTP/HTTPS, WebSocket
- **字符编码**: UTF-8

## 认证

### Bearer Token 认证

在需要认证的API请求中，需要在请求头中添加 `Authorization` 字段：

```
Authorization: Bearer <your-jwt-token>
```

### 获取Token

通过登录接口获取JWT Token：

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

响应示例：
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin"
    }
  }
}
```

## 📡 API 端点

### 🔐 认证相关

#### 用户注册
```http
POST /api/v1/auth/register
```

**请求参数**:
```json
{
  "username": "string (required, min:3, max:50)",
  "email": "string (required, email format)",
  "password": "string (required, min:6)"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "newuser",
    "email": "newuser@example.com",
    "role": "user",
    "status": 1,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 用户登录
```http
POST /api/v1/auth/login
```

**请求参数**:
```json
{
  "username": "string (required)",
  "password": "string (required)"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin"
    }
  }
}
```

#### 用户登出
```http
POST /api/v1/auth/logout
```

**认证**: 需要

**响应示例**:
```json
{
  "success": true,
  "message": "登出成功"
}
```

#### 获取用户资料
```http
GET /api/v1/auth/profile
```

**认证**: 需要

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "role": "admin",
    "status": 1,
    "avatar": "https://example.com/avatar.jpg",
    "email_verified_at": "2024-01-01T00:00:00Z",
    "last_login_at": "2024-01-01T00:00:00Z",
    "login_count": 10,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 更新用户资料
```http
PUT /api/v1/auth/profile
```

**认证**: 需要

**请求参数**:
```json
{
  "username": "string (optional, min:3, max:50)",
  "email": "string (optional, email format)",
  "avatar": "string (optional, url)"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "updateduser",
    "email": "updated@example.com",
    "avatar": "https://example.com/new-avatar.jpg"
  }
}
```

#### 刷新Token
```http
POST /api/v1/auth/refresh
```

**认证**: 需要

**响应示例**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 👥 用户管理 (管理员)

#### 获取用户列表
```http
GET /api/v1/users?page=1&limit=10&search=keyword
```

**认证**: 需要 (管理员)

**查询参数**:
- `page`: 页码 (默认: 1)
- `limit`: 每页数量 (默认: 10, 最大: 100)
- `search`: 搜索关键词 (可选)

**响应示例**:
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "role": "admin",
        "status": 1,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "per_page": 10,
      "total": 1,
      "total_pages": 1
    }
  }
}
```

#### 获取用户详情
```http
GET /api/v1/users/{id}
```

**认证**: 需要 (管理员)

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "role": "admin",
    "status": 1,
    "avatar": "https://example.com/avatar.jpg",
    "email_verified_at": "2024-01-01T00:00:00Z",
    "last_login_at": "2024-01-01T00:00:00Z",
    "login_count": 10,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 更新用户信息
```http
PUT /api/v1/users/{id}
```

**认证**: 需要 (管理员)

**请求参数**:
```json
{
  "username": "string (optional)",
  "email": "string (optional, email format)",
  "role": "string (optional, admin|user)",
  "status": "integer (optional, 1|0)"
}
```

#### 删除用户
```http
DELETE /api/v1/users/{id}
```

**认证**: 需要 (管理员)

**响应示例**:
```json
{
  "success": true,
  "message": "用户删除成功"
}
```

### 📝 文章管理

#### 获取文章列表
```http
GET /api/v1/posts?page=1&limit=10&category_id=1&search=keyword
```

**查询参数**:
- `page`: 页码 (默认: 1)
- `limit`: 每页数量 (默认: 10, 最大: 100)
- `category_id`: 分类ID (可选)
- `search`: 搜索关键词 (可选)

**响应示例**:
```json
{
  "success": true,
  "data": {
    "posts": [
      {
        "id": 1,
        "title": "示例文章",
        "excerpt": "文章摘要",
        "status": 1,
        "user": {
          "id": 1,
          "username": "admin"
        },
        "category": {
          "id": 1,
          "name": "技术"
        },
        "tags": [
          {
            "id": 1,
            "name": "Go",
            "color": "#00ADD8"
          }
        ],
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "per_page": 10,
      "total": 1,
      "total_pages": 1
    }
  }
}
```

#### 获取文章详情
```http
GET /api/v1/posts/{id}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "示例文章",
    "content": "文章内容...",
    "excerpt": "文章摘要",
    "status": 1,
    "user": {
      "id": 1,
      "username": "admin"
    },
    "category": {
      "id": 1,
      "name": "技术"
    },
    "tags": [
      {
        "id": 1,
        "name": "Go",
        "color": "#00ADD8"
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 创建文章
```http
POST /api/v1/posts
```

**认证**: 需要

**请求参数**:
```json
{
  "title": "string (required, max:200)",
  "content": "string (required)",
  "excerpt": "string (optional, max:500)",
  "category_id": "integer (required)",
  "tag_ids": "array (optional)"
}
```

#### 更新文章
```http
PUT /api/v1/posts/{id}
```

**认证**: 需要

**请求参数**:
```json
{
  "title": "string (optional, max:200)",
  "content": "string (optional)",
  "excerpt": "string (optional, max:500)",
  "category_id": "integer (optional)",
  "tag_ids": "array (optional)"
}
```

#### 删除文章
```http
DELETE /api/v1/posts/{id}
```

**认证**: 需要

### 📂 分类管理

#### 获取分类列表
```http
GET /api/v1/categories
```

**响应示例**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "技术",
      "description": "技术相关文章",
      "slug": "tech",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 创建分类 (管理员)
```http
POST /api/v1/categories
```

**认证**: 需要 (管理员)

**请求参数**:
```json
{
  "name": "string (required, max:100)",
  "description": "string (optional)",
  "slug": "string (optional, unique)"
}
```

### 🏷️ 标签管理

#### 获取标签列表
```http
GET /api/v1/tags
```

**响应示例**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Go",
      "color": "#00ADD8",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 创建标签 (管理员)
```http
POST /api/v1/tags
```

**认证**: 需要 (管理员)

**请求参数**:
```json
{
  "name": "string (required, max:50)",
  "color": "string (optional, hex color)"
}
```

### 📁 文件管理

#### 文件上传
```http
POST /api/v1/storage/upload
```

**认证**: 需要

**请求参数**: `multipart/form-data`
- `file`: 文件 (required)
- `folder`: 文件夹路径 (optional)

**响应示例**:
```json
{
  "success": true,
  "data": {
    "filename": "example.jpg",
    "path": "/uploads/example.jpg",
    "size": 1024,
    "mime_type": "image/jpeg",
    "url": "http://localhost:8080/uploads/example.jpg"
  }
}
```

#### 文件下载
```http
GET /api/v1/storage/download/{filename}
```

**响应**: 文件流

#### 删除文件
```http
DELETE /api/v1/storage/delete/{filename}
```

**认证**: 需要

### 📊 监控和健康检查

#### 健康检查
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
    "uptime": "1h30m",
    "version": "1.0.0"
  }
}
```

#### 系统指标 (管理员)
```http
GET /api/v1/monitoring/metrics
```

**认证**: 需要 (管理员)

**响应示例**:
```json
{
  "success": true,
  "data": {
    "timestamp": "2024-01-01T00:00:00Z",
    "uptime": "1h30m",
    "memory": {
      "alloc_mb": 10.5,
      "total_alloc_mb": 50.2,
      "sys_mb": 100.0
    },
    "cpu": {
      "num_cpu": 8,
      "usage": 15.5
    },
    "goroutines": 25,
    "gc_stats": {
      "num_gc": 5,
      "pause_total_ms": 10.5
    }
  }
}
```

#### 性能统计 (管理员)
```http
GET /api/v1/monitoring/stats
```

**认证**: 需要 (管理员)

## 错误处理

### 错误响应格式

```json
{
  "success": false,
  "message": "错误描述",
  "errors": {
    "field": ["具体错误信息"]
  }
}
```

### 常见HTTP状态码

- `200`: 请求成功
- `201`: 创建成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `422`: 验证失败
- `500`: 服务器内部错误

### 错误代码

| 代码 | 说明 |
|------|------|
| `VALIDATION_ERROR` | 参数验证失败 |
| `AUTHENTICATION_FAILED` | 认证失败 |
| `PERMISSION_DENIED` | 权限不足 |
| `RESOURCE_NOT_FOUND` | 资源不存在 |
| `RESOURCE_ALREADY_EXISTS` | 资源已存在 |
| `INTERNAL_SERVER_ERROR` | 服务器内部错误 |

## 分页

支持分页的API使用以下查询参数：

- `page`: 页码 (默认: 1)
- `limit`: 每页数量 (默认: 10, 最大: 100)

分页响应格式：

```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "current_page": 1,
      "per_page": 10,
      "total": 100,
      "total_pages": 10,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

## 搜索

支持搜索的API使用 `search` 查询参数：

```http
GET /api/v1/posts?search=关键词
```

搜索会在标题、内容等字段中进行模糊匹配。

## 排序

支持排序的API使用 `sort` 和 `order` 查询参数：

```http
GET /api/v1/posts?sort=created_at&order=desc
```

- `sort`: 排序字段 (如: created_at, title)
- `order`: 排序方向 (asc, desc)

## 过滤

支持过滤的API使用相应的查询参数：

```http
GET /api/v1/posts?category_id=1&status=1
```

## 速率限制

API 实现了速率限制，默认限制：

- 认证接口: 5次/分钟
- 其他接口: 100次/分钟

超过限制会返回 `429 Too Many Requests` 状态码。

## 版本控制

API 使用 URL 路径进行版本控制：

- 当前版本: `/api/v1/`
- 未来版本: `/api/v2/`

### 🔒 安全防护

#### 获取安全事件
```http
GET /api/v1/security/events?page=1&limit=20&event_type=login&event_level=high
```

#### 获取威胁情报
```http
GET /api/v1/security/threats?page=1&limit=20&threat_type=malware&severity=high
```

#### 获取登录尝试记录
```http
GET /api/v1/security/login-attempts?page=1&limit=20&username=admin&success=false
```

### 📈 性能监控

#### 获取当前系统指标
```http
GET /api/v1/performance/current
```

#### 获取性能报告
```http
GET /api/v1/performance/metrics?metric_type=system_resources&start=2024-12-20T00:00:00Z&end=2024-12-20T23:59:59Z
```

#### 获取告警列表
```http
GET /api/v1/performance/alerts/active
```

### 🔍 查询优化

#### 获取慢查询列表
```http
GET /api/v1/query-optimization/slow-queries?limit=50&warning_level=CRITICAL
```

#### 获取查询统计
```http
GET /api/v1/query-optimization/query-statistics
```

#### 获取索引建议
```http
GET /api/v1/query-optimization/index-suggestions
```

### 💬 WebSocket 实时通信

#### 建立WebSocket连接
```
GET /ws/connect?room_id={room_id}
```

#### 获取房间列表
```http
GET /ws/rooms
```

#### 获取在线用户
```http
GET /ws/users/online
```

#### 获取系统统计
```http
GET /ws/stats
```

## 📝 更新日志

### v1.2.0 (最新)
- 新增性能监控系统
- 新增安全防护系统
- 新增查询优化系统
- 新增WebSocket实时通信
- 优化API响应格式
- 增强错误处理机制

### v1.1.0
- 新增文件管理功能
- 新增分类和标签管理
- 新增监控和健康检查
- 优化用户管理功能

### v1.0.0
- 初始版本发布
- 完整的用户认证系统
- 文章管理系统
- 基础API功能

## 🔗 相关文档

- [部署指南](DEPLOYMENT.md) - 详细的部署说明
- [开发指南](DEVELOPMENT.md) - 开发环境设置和代码规范
- [测试指南](TESTING.md) - 测试框架使用说明
- [日志系统](LOGGING_SYSTEM.md) - 日志管理使用说明
- [监控系统](MONITORING_SYSTEM.md) - 监控告警系统文档
- [性能监控](PERFORMANCE_MONITORING_SYSTEM.md) - 性能监控系统文档
- [安全系统](SECURITY_SYSTEM.md) - 安全防护系统文档
- [查询优化](QUERY_OPTIMIZATION_SYSTEM.md) - 查询优化系统文档
- [WebSocket系统](WEBSOCKET_SYSTEM.md) - 实时通信系统文档

---

更多信息请访问：[项目主页](https://github.com/your-username/cloud-platform-api)
