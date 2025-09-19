# Cloud Platform API 开发指南

## 概述

本文档为 Cloud Platform API 项目的开发者提供详细的开发指南，包括项目结构、开发环境设置、编码规范、测试指南等。

## 项目架构

### 整体架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Layer    │    │  Business Layer │    │  Data Layer     │
│                 │    │                 │    │                 │
│  Controllers    │───▶│    Services     │───▶│     Models      │
│  Middleware     │    │                 │    │   Database      │
│  Routes         │    │                 │    │     Cache       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 分层说明
1. **HTTP Layer**: 处理 HTTP 请求和响应
2. **Business Layer**: 实现业务逻辑
3. **Data Layer**: 处理数据存储和访问

## 开发环境设置

### 1. 环境要求
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+ (可选)
- Git
- IDE (推荐 VS Code 或 GoLand)

### 2. 项目初始化
```bash
# 克隆项目
git clone https://github.com/your-org/cloud-platform-api.git
cd cloud-platform-api

# 安装依赖
go mod download

# 复制配置文件
cp env.example .env

# 编辑配置文件
nano .env
```

### 3. 数据库设置
```bash
# 创建数据库
mysql -u root -p
CREATE DATABASE cloud_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 运行迁移
go run main.go migrate
```

### 4. 启动开发服务器
```bash
# 启动应用
go run main.go

# 或使用 air 进行热重载
air
```

## 项目结构详解

### 目录结构
```
cloud-platform-api/
├── app/                          # 应用核心代码
│   ├── Config/                   # 配置管理
│   │   ├── config.go            # 配置结构体
│   │   └── loader.go            # 配置加载器
│   ├── Database/                 # 数据库相关
│   │   ├── database.go          # 数据库连接
│   │   ├── migrations/          # 数据库迁移
│   │   └── connection_pool.go   # 连接池管理
│   ├── Http/                    # HTTP 层
│   │   ├── Controllers/         # 控制器
│   │   │   ├── Controller.go    # 基础控制器
│   │   │   ├── AuthController.go # 认证控制器
│   │   │   └── UserController.go # 用户控制器
│   │   ├── Middleware/          # 中间件
│   │   │   ├── AuthMiddleware.go # 认证中间件
│   │   │   ├── CORSMiddleware.go # CORS 中间件
│   │   │   └── ...              # 其他中间件
│   │   └── Routes/              # 路由
│   │       └── routes.go        # 路由配置
│   ├── Models/                  # 数据模型
│   │   ├── User.go             # 用户模型
│   │   └── BaseModel.go        # 基础模型
│   ├── Services/               # 业务逻辑层
│   │   ├── BaseService.go      # 基础服务
│   │   ├── AuthService.go      # 认证服务
│   │   ├── UserService.go      # 用户服务
│   │   └── CacheService.go     # 缓存服务
│   └── Utils/                  # 工具函数
│       ├── password.go         # 密码工具
│       ├── jwt.go             # JWT 工具
│       └── validator.go       # 验证工具
├── docs/                       # 文档
├── monitoring/                 # 监控配置
├── tests/                      # 测试文件
├── storage/                    # 存储目录
│   └── logs/                   # 日志文件
├── main.go                     # 应用入口
├── go.mod                      # Go 模块文件
├── go.sum                      # 依赖校验文件
├── Dockerfile                  # Docker 配置
├── docker-compose.yml          # Docker Compose 配置
└── Makefile                    # 构建脚本
```

## 编码规范

### 1. Go 代码规范

#### 命名规范
```go
// 包名：小写，简短，有意义
package services

// 类型名：大驼峰
type UserService struct {}

// 方法名：大驼峰
func (s *UserService) GetUser(id uint) (*User, error) {}

// 变量名：小驼峰
var userName string

// 常量：全大写，下划线分隔
const MAX_RETRY_COUNT = 3

// 私有成员：小写开头
type userService struct {
    db *gorm.DB
}
```

#### 注释规范
```go
// Package services 提供业务逻辑服务
// 包含用户管理、认证、缓存等核心功能
package services

// UserService 用户服务
// 功能说明：
// 1. 用户数据的业务逻辑处理
// 2. 用户CRUD操作
// 3. 用户权限和状态管理
// 4. 用户数据安全处理（密码字段过滤）
type UserService struct {
    BaseService
}

// GetUser 获取单个用户
// 功能说明：
// 1. 根据用户ID获取用户详细信息
// 2. 自动过滤敏感信息（密码字段）
// 3. 处理用户不存在的情况
// 4. 用于用户资料查看和编辑
// 参数：
//   - id: 用户ID
// 返回：
//   - *User: 用户信息
//   - error: 错误信息
func (s *UserService) GetUser(id uint) (*User, error) {
    // 实现代码
}
```

#### 错误处理
```go
// 使用 errors.New 创建简单错误
if user == nil {
    return nil, errors.New("user not found")
}

// 使用 fmt.Errorf 创建格式化错误
if err != nil {
    return fmt.Errorf("failed to get user: %w", err)
}

// 使用自定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
}
```

### 2. 项目特定规范

#### 控制器规范
```go
// 所有控制器都应该继承 BaseController
type UserController struct {
    BaseController
    userService *Services.UserService
}

// 使用统一的响应格式
func (c *UserController) GetUser(ctx *gin.Context) {
    // 获取参数
    userID, err := c.GetUserIDFromParam(ctx)
    if err != nil {
        c.ErrorResponse(ctx, http.StatusBadRequest, "invalid user ID", err)
        return
    }

    // 调用服务
    user, err := c.userService.GetUser(userID)
    if err != nil {
        c.ErrorResponse(ctx, http.StatusNotFound, "user not found", err)
        return
    }

    // 返回成功响应
    c.SuccessResponse(ctx, "user retrieved successfully", user)
}
```

#### 服务层规范
```go
// 所有服务都应该继承 BaseService
type UserService struct {
    BaseService
}

// 方法应该包含详细的注释
// 参数验证应该在服务层进行
func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
    // 参数验证
    if err := s.validateCreateUserRequest(req); err != nil {
        return nil, err
    }

    // 业务逻辑
    user := &User{
        Username: req.Username,
        Email:    req.Email,
    }

    // 数据库操作
    if err := s.getDB().Create(user).Error; err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}
```

## 开发流程

### 1. 功能开发流程

#### 创建新功能
```bash
# 1. 创建功能分支
git checkout -b feature/user-profile-management

# 2. 创建相关文件
# - 模型文件 (app/Models/UserProfile.go)
# - 服务文件 (app/Services/UserProfileService.go)
# - 控制器文件 (app/Http/Controllers/UserProfileController.go)
# - 路由配置 (在 app/Http/Routes/routes.go 中添加)

# 3. 编写代码
# 4. 编写测试
# 5. 提交代码
git add .
git commit -m "feat: add user profile management"

# 6. 推送分支
git push origin feature/user-profile-management

# 7. 创建 Pull Request
```

#### 代码审查清单
- [ ] 代码符合项目规范
- [ ] 包含必要的注释
- [ ] 错误处理完整
- [ ] 包含单元测试
- [ ] 通过所有测试
- [ ] 无安全漏洞
- [ ] 性能考虑合理

### 2. 测试流程

#### 单元测试
```go
// tests/services/user_service_test.go
package tests

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "cloud-platform-api/app/Services"
)

func TestUserService_GetUser(t *testing.T) {
    // 准备测试数据
    service := Services.NewUserService()
    
    // 执行测试
    user, err := service.GetUser(1)
    
    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, uint(1), user.ID)
}
```

#### 集成测试
```go
// tests/integration/user_api_test.go
package tests

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestUserAPI_GetUser(t *testing.T) {
    // 设置测试环境
    gin.SetMode(gin.TestMode)
    router := setupTestRouter()
    
    // 创建请求
    req, _ := http.NewRequest("GET", "/api/v1/users/1", nil)
    w := httptest.NewRecorder()
    
    // 执行请求
    router.ServeHTTP(w, req)
    
    // 验证响应
    assert.Equal(t, http.StatusOK, w.Code)
}
```

#### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./app/Services/...

# 运行测试并显示覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 运行基准测试
go test -bench=.
```

### 3. 调试技巧

#### 使用日志调试
```go
import "log"

func (s *UserService) GetUser(id uint) (*User, error) {
    log.Printf("Getting user with ID: %d", id)
    
    var user User
    if err := s.getDB().First(&user, id).Error; err != nil {
        log.Printf("Error getting user: %v", err)
        return nil, err
    }
    
    log.Printf("Successfully retrieved user: %s", user.Username)
    return &user, nil
}
```

#### 使用调试器
```bash
# 使用 delve 调试器
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug main.go

# 在 VS Code 中使用调试功能
# 创建 .vscode/launch.json 配置文件
```

## 数据库开发

### 1. 模型定义
```go
// app/Models/User.go
package Models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Username  string         `json:"username" gorm:"uniqueIndex;not null;size:50"`
    Email     string         `json:"email" gorm:"uniqueIndex;not null;size:100"`
    Password  string         `json:"-" gorm:"not null;size:255"`
    Status    int            `json:"status" gorm:"default:1;comment:用户状态 1:正常 0:禁用"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (User) TableName() string {
    return "users"
}
```

### 2. 数据库迁移
```go
// app/Database/migrations/001_create_users_table.go
package migrations

import (
    "gorm.io/gorm"
)

func CreateUsersTable(db *gorm.DB) error {
    return db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
            username VARCHAR(50) NOT NULL UNIQUE,
            email VARCHAR(100) NOT NULL UNIQUE,
            password VARCHAR(255) NOT NULL,
            status TINYINT DEFAULT 1 COMMENT '用户状态 1:正常 0:禁用',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP NULL,
            INDEX idx_username (username),
            INDEX idx_email (email),
            INDEX idx_status (status),
            INDEX idx_deleted_at (deleted_at)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
    `).Error
}
```

### 3. 查询优化
```go
// 使用 Select 只查询需要的字段
func (s *UserService) GetUsers() ([]User, error) {
    var users []User
    err := s.getDB().Select("id,username,email,status,created_at,updated_at").Find(&users).Error
    return users, err
}

// 使用 Preload 预加载关联数据
func (s *UserService) GetUserWithProfile(id uint) (*User, error) {
    var user User
    err := s.getDB().Preload("Profile").First(&user, id).Error
    return &user, err
}

// 使用事务
func (s *UserService) CreateUserWithProfile(user *User, profile *UserProfile) error {
    return s.getDB().Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        profile.UserID = user.ID
        if err := tx.Create(profile).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

## 缓存开发

### 1. 缓存策略
```go
// 缓存键命名规范
const (
    UserCacheKey    = "user:%d"           // 用户缓存
    UserListCacheKey = "users:list:%d"    // 用户列表缓存
    SessionCacheKey = "session:%s"        // 会话缓存
)

// 缓存 TTL 设置
const (
    UserCacheTTL    = 15 * time.Minute    // 用户缓存 15 分钟
    UserListCacheTTL = 5 * time.Minute    // 用户列表缓存 5 分钟
    SessionCacheTTL  = 30 * time.Minute   // 会话缓存 30 分钟
)
```

### 2. 缓存实现
```go
func (s *UserService) GetUser(id uint) (*User, error) {
    // 尝试从缓存获取
    cacheKey := fmt.Sprintf(UserCacheKey, id)
    if cached, exists := s.cache.Get(cacheKey); exists {
        if user, ok := cached.(*User); ok {
            return user, nil
        }
    }

    // 从数据库查询
    user, err := s.getUserFromDB(id)
    if err != nil {
        return nil, err
    }

    // 存入缓存
    s.cache.Set(cacheKey, user, UserCacheTTL)
    return user, nil
}
```

## 中间件开发

### 1. 中间件结构
```go
// app/Http/Middleware/CustomMiddleware.go
package Middleware

import (
    "github.com/gin-gonic/gin"
)

type CustomMiddleware struct {
    BaseMiddleware
    // 中间件特定字段
}

// NewCustomMiddleware 创建中间件
func NewCustomMiddleware() *CustomMiddleware {
    return &CustomMiddleware{}
}

// Handle 处理请求
func (m *CustomMiddleware) Handle() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 请求前处理
        c.Next()
        // 请求后处理
    }
}
```

### 2. 中间件注册
```go
// app/Http/Routes/routes.go
func RegisterRoutes(engine *gin.Engine) {
    // 创建中间件
    customMiddleware := Middleware.NewCustomMiddleware()
    
    // 注册中间件
    engine.Use(customMiddleware.Handle())
    
    // 注册路由
    api := engine.Group("/api/v1")
    {
        api.GET("/users", userController.GetUsers)
    }
}
```

## 性能优化

### 1. 数据库优化
```go
// 使用连接池
func initDB() {
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetConnMaxLifetime(time.Hour)
}

// 使用索引
// 在模型定义中添加索引标签
type User struct {
    Username string `gorm:"index:idx_username"`
    Email    string `gorm:"index:idx_email"`
}
```

### 2. 内存优化
```go
// 使用对象池
var userPool = sync.Pool{
    New: func() interface{} {
        return &User{}
    },
}

func (s *UserService) GetUser(id uint) (*User, error) {
    user := userPool.Get().(*User)
    defer userPool.Put(user)
    
    // 使用 user 对象
    return user, nil
}
```

### 3. 并发优化
```go
// 使用 goroutine 池
func (s *UserService) ProcessUsers(users []User) error {
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) // 限制并发数
    
    for _, user := range users {
        wg.Add(1)
        go func(u User) {
            defer wg.Done()
            semaphore <- struct{}{} // 获取信号量
            defer func() { <-semaphore }() // 释放信号量
            
            // 处理用户
            s.processUser(u)
        }(user)
    }
    
    wg.Wait()
    return nil
}
```

## 安全开发

### 1. 输入验证
```go
// 使用验证器
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

func (c *UserController) CreateUser(ctx *gin.Context) {
    var req CreateUserRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.ErrorResponse(ctx, http.StatusBadRequest, "invalid request", err)
        return
    }
    
    // 验证请求
    if err := c.validator.Struct(req); err != nil {
        c.ErrorResponse(ctx, http.StatusBadRequest, "validation failed", err)
        return
    }
    
    // 处理请求
}
```

### 2. SQL 注入防护
```go
// 使用参数化查询
func (s *UserService) GetUserByUsername(username string) (*User, error) {
    var user User
    err := s.getDB().Where("username = ?", username).First(&user).Error
    return &user, err
}

// 避免字符串拼接
// 错误示例
// query := "SELECT * FROM users WHERE username = '" + username + "'"

// 正确示例
err := s.getDB().Where("username = ?", username).First(&user).Error
```

### 3. XSS 防护
```go
// 输入过滤
func sanitizeInput(input string) string {
    // 移除 HTML 标签
    re := regexp.MustCompile(`<[^>]*>`)
    return re.ReplaceAllString(input, "")
}

// 输出编码
func (c *UserController) GetUser(ctx *gin.Context) {
    user, err := c.userService.GetUser(userID)
    if err != nil {
        c.ErrorResponse(ctx, http.StatusNotFound, "user not found", err)
        return
    }
    
    // 使用 JSON 编码自动处理 XSS
    ctx.JSON(http.StatusOK, gin.H{"user": user})
}
```

## 测试开发

### 1. 单元测试
```go
func TestUserService_GetUser(t *testing.T) {
    // 准备测试数据
    mockDB := setupMockDB()
    service := &UserService{db: mockDB}
    
    // 测试用例
    tests := []struct {
        name     string
        userID   uint
        wantUser *User
        wantErr  bool
    }{
        {
            name:     "valid user ID",
            userID:   1,
            wantUser: &User{ID: 1, Username: "testuser"},
            wantErr:  false,
        },
        {
            name:     "invalid user ID",
            userID:   999,
            wantUser: nil,
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user, err := service.GetUser(tt.userID)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.wantUser, user)
            }
        })
    }
}
```

### 2. 集成测试
```go
func TestUserAPI_Integration(t *testing.T) {
    // 设置测试环境
    gin.SetMode(gin.TestMode)
    router := setupTestRouter()
    
    // 测试用户创建
    t.Run("create user", func(t *testing.T) {
        userData := map[string]interface{}{
            "username": "testuser",
            "email":    "test@example.com",
            "password": "password123",
        }
        
        jsonData, _ := json.Marshal(userData)
        req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusCreated, w.Code)
    })
}
```

### 3. 性能测试
```go
func BenchmarkUserService_GetUser(b *testing.B) {
    service := setupUserService()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.GetUser(1)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 部署开发

### 1. Docker 开发
```dockerfile
# Dockerfile.dev
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/.env .

CMD ["./main"]
```

### 2. 本地开发环境
```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis
    depends_on:
      - mysql
      - redis
    volumes:
      - .:/app
      - /app/vendor

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: cloud_platform
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"

volumes:
  mysql_data:
```

## 常见问题

### 1. 开发环境问题
**问题**: 数据库连接失败
**解决**: 检查数据库服务是否启动，配置文件是否正确

**问题**: 依赖包下载失败
**解决**: 设置 Go 代理 `go env -w GOPROXY=https://goproxy.cn,direct`

### 2. 代码问题
**问题**: 循环导入
**解决**: 重新设计包结构，避免循环依赖

**问题**: 内存泄漏
**解决**: 检查 goroutine 是否正确关闭，资源是否正确释放

### 3. 测试问题
**问题**: 测试数据污染
**解决**: 使用事务回滚或独立的测试数据库

**问题**: 并发测试失败
**解决**: 使用 `t.Parallel()` 和适当的同步机制

## 最佳实践

### 1. 代码组织
- 按功能模块组织代码
- 保持包的内聚性
- 避免过深的嵌套

### 2. 错误处理
- 使用有意义的错误信息
- 记录详细的错误日志
- 提供用户友好的错误响应

### 3. 性能考虑
- 避免 N+1 查询问题
- 使用适当的缓存策略
- 监控内存和 CPU 使用

### 4. 安全考虑
- 验证所有输入
- 使用参数化查询
- 实施适当的权限控制

## 贡献指南

### 1. 提交规范
```
feat: 新功能
fix: 修复问题
docs: 文档更新
style: 代码格式调整
refactor: 代码重构
test: 测试相关
chore: 构建过程或辅助工具的变动
```

### 2. Pull Request 流程
1. Fork 项目
2. 创建功能分支
3. 提交代码
4. 创建 Pull Request
5. 代码审查
6. 合并代码

### 3. 代码审查要点
- 代码质量和规范
- 功能完整性
- 测试覆盖率
- 性能影响
- 安全性考虑

## 学习资源

### 1. Go 语言学习
- [Go 官方文档](https://golang.org/doc/)
- [Go 语言圣经](https://gopl-zh.github.io/)
- [Go 语言实战](https://www.manning.com/books/go-in-action)

### 2. Web 开发
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [RESTful API 设计指南](https://restfulapi.net/)

### 3. 数据库
- [MySQL 官方文档](https://dev.mysql.com/doc/)
- [Redis 官方文档](https://redis.io/documentation)
- [数据库设计最佳实践](https://www.databasedesign.com/)

### 4. 测试
- [Go 测试文档](https://golang.org/pkg/testing/)
- [Testify 库](https://github.com/stretchr/testify)
- [测试驱动开发](https://en.wikipedia.org/wiki/Test-driven_development)

## 联系方式

- **项目维护者**: 开发团队
- **技术支持**: support@yourdomain.com
- **问题反馈**: https://github.com/your-org/cloud-platform-api/issues
- **文档更新**: 请提交 Pull Request
