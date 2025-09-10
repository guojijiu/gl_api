# API文档模板

## 📋 概述

API接口文档，包含所有可用的接口、参数、响应格式等信息。

## 🔧 基础信息

### 基本信息
- **Base URL**: `https://api.example.com`
- **API版本**: `v1`
- **认证方式**: Bearer Token
- **数据格式**: JSON
- **字符编码**: UTF-8

### 认证说明

#### Bearer Token认证
```http
Authorization: Bearer <your-token>
```

#### 获取Token
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "token_type": "Bearer"
  }
}
```

## 📡 接口列表

### 认证相关

#### 用户登录
```http
POST /api/v1/auth/login
```

**请求参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**请求示例：**
```json
{
  "username": "admin",
  "password": "password123"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "登录成功",
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

#### 用户注册
```http
POST /api/v1/auth/register
```

**请求参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| username | string | 是 | 用户名 |
| email | string | 是 | 邮箱 |
| password | string | 是 | 密码 |

**请求示例：**
```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "password123"
}
```

**响应示例：**
```json
{
  "code": 201,
  "message": "注册成功",
  "data": {
    "user": {
      "id": 2,
      "username": "newuser",
      "email": "newuser@example.com",
      "role": "user"
    }
  }
}
```

### 用户管理

#### 获取用户列表
```http
GET /api/v1/users
```

**请求参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | 否 | 页码，默认1 |
| limit | int | 否 | 每页数量，默认20 |
| search | string | 否 | 搜索关键词 |

**请求示例：**
```http
GET /api/v1/users?page=1&limit=10&search=admin
Authorization: Bearer <your-token>
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "users": [
      {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "role": "admin",
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

**路径参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 用户ID |

**请求示例：**
```http
GET /api/v1/users/1
Authorization: Bearer <your-token>
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "role": "admin",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 更新用户信息
```http
PUT /api/v1/users/{id}
```

**路径参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 用户ID |

**请求参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| username | string | 否 | 用户名 |
| email | string | 否 | 邮箱 |
| role | string | 否 | 角色 |

**请求示例：**
```json
{
  "username": "newadmin",
  "email": "newadmin@example.com",
  "role": "admin"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "id": 1,
    "username": "newadmin",
    "email": "newadmin@example.com",
    "role": "admin",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 删除用户
```http
DELETE /api/v1/users/{id}
```

**路径参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 用户ID |

**请求示例：**
```http
DELETE /api/v1/users/1
Authorization: Bearer <your-token>
```

**响应示例：**
```json
{
  "code": 200,
  "message": "删除成功",
  "data": null
}
```

### 内容管理

#### 获取内容列表
```http
GET /api/v1/posts
```

**请求参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | 否 | 页码，默认1 |
| limit | int | 否 | 每页数量，默认20 |
| category | string | 否 | 分类 |
| status | string | 否 | 状态 |

**请求示例：**
```http
GET /api/v1/posts?page=1&limit=10&category=tech&status=published
Authorization: Bearer <your-token>
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "posts": [
      {
        "id": 1,
        "title": "文章标题",
        "content": "文章内容",
        "category": "tech",
        "status": "published",
        "author": {
          "id": 1,
          "username": "admin"
        },
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

#### 创建内容
```http
POST /api/v1/posts
```

**请求参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| title | string | 是 | 标题 |
| content | string | 是 | 内容 |
| category | string | 是 | 分类 |
| status | string | 否 | 状态，默认draft |

**请求示例：**
```json
{
  "title": "新文章标题",
  "content": "新文章内容",
  "category": "tech",
  "status": "published"
}
```

**响应示例：**
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": 2,
    "title": "新文章标题",
    "content": "新文章内容",
    "category": "tech",
    "status": "published",
    "author": {
      "id": 1,
      "username": "admin"
    },
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

## 📊 响应格式

### 成功响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 响应数据
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "请求参数错误",
  "error": "详细错误信息",
  "errors": {
    "field1": "字段1错误信息",
    "field2": "字段2错误信息"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 分页响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      // 数据项列表
    ],
    "pagination": {
      "current_page": 1,
      "per_page": 20,
      "total": 100,
      "total_pages": 5,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

## 🚨 错误码说明

| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| 200 | 200 | 成功 |
| 201 | 201 | 创建成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未授权 |
| 403 | 403 | 禁止访问 |
| 404 | 404 | 资源不存在 |
| 422 | 422 | 验证失败 |
| 500 | 500 | 服务器内部错误 |

## 🔒 安全说明

### 认证要求
- 大部分接口需要Bearer Token认证
- Token有效期通常为1小时
- 支持Token刷新机制

### 权限控制
- 不同角色有不同的访问权限
- 管理员可以访问所有接口
- 普通用户只能访问自己的资源

### 安全建议
- 使用HTTPS协议
- 定期更换Token
- 不要在客户端存储敏感信息
- 使用强密码

## 📚 相关资源

- [认证指南](AUTH.md)
- [错误处理](ERROR_HANDLING.md)
- [SDK文档](SDK.md)
- [Postman集合](postman_collection.json)

## 📞 技术支持

如有问题或建议，请通过以下方式联系：

- 项目Issues: GitHub Issues
- 技术讨论: GitHub Discussions
- 技术支持: support@example.com

---

**API版本**: 1.0.0  
**最后更新**: 2024年12月  
**维护者**: API团队
