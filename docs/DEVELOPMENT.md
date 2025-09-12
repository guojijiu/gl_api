# 开发指南

本文档为Cloud Platform API项目提供详细的开发指南，包括环境设置、代码规范、开发流程和最佳实践。

## 📋 目录

- [开发环境设置](#开发环境设置)
- [项目结构](#项目结构)
- [代码规范](#代码规范)
- [开发流程](#开发流程)
- [测试指南](#测试指南)
- [API开发](#api开发)
- [数据库操作](#数据库操作)
- [调试技巧](#调试技巧)
- [性能优化](#性能优化)
- [贡献指南](#贡献指南)

## 🛠️ 开发环境设置

### 1. 系统要求

- **操作系统**: Linux, macOS, Windows
- **Go版本**: 1.21+
- **Git**: 最新版本
- **IDE**: VS Code, GoLand, Vim等
- **数据库**: MySQL 8.0+, PostgreSQL 13+, SQLite 3.x
- **Redis**: 6.0+ (可选)

### 2. 安装Go

```bash
# Linux/macOS
wget https://golang.org/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# macOS (使用Homebrew)
brew install go

# Windows
# 下载并安装 https://golang.org/dl/go1.21.windows-amd64.msi
```

### 3. 安装开发工具

```bash
# 安装代码格式化工具
go install golang.org/x/tools/cmd/goimports@latest

# 安装代码检查工具
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装API文档生成工具
go install github.com/swaggo/swag/cmd/swag@latest

# 安装测试工具
go install github.com/stretchr/testify@latest

# 安装热重载工具
go install github.com/cosmtrek/air@latest

# 安装性能分析工具
go install github.com/google/pprof@latest

# 安装依赖检查工具
go install github.com/golang/dep/cmd/dep@latest
```

### 4. 项目初始化

```bash
# 克隆项目
git clone <repository-url>
cd cloud-platform-api

# 安装依赖
go mod download
go mod tidy

# 复制环境配置
cp env.example .env

# 初始化数据库
go run scripts/migrate.go
```

## 📁 项目结构

```
cloud-platform-api/
├── app/                          # 应用核心代码
│   ├── Config/                   # 配置管理
│   │   ├── base.go              # 主配置结构
│   │   ├── config.go            # 配置加载
│   │   ├── database.go          # 数据库配置
│   │   ├── jwt.go               # JWT配置
│   │   ├── redis.go             # Redis配置
│   │   ├── email.go             # 邮件配置
│   │   └── storage.go           # 存储配置
│   ├── Database/                # 数据库相关
│   │   ├── database.go          # 数据库连接
│   │   ├── connection_pool.go   # 连接池管理
│   │   ├── models.go            # 模型定义
│   │   └── Migrations/          # 数据库迁移
│   ├── Http/                    # HTTP层
│   │   ├── Controllers/         # 控制器
│   │   ├── Middleware/          # 中间件
│   │   ├── Requests/            # 请求验证
│   │   └── Routes/              # 路由定义
│   ├── Models/                  # 数据模型
│   ├── Services/                # 业务逻辑层
│   ├── Storage/                 # 存储管理
│   └── Utils/                   # 工具函数
├── bootstrap/                   # 应用启动
├── docs/                        # 文档
├── scripts/                     # 脚本文件
├── storage/                     # 存储目录
├── tests/                       # 测试文件
├── main.go                      # 应用入口
├── go.mod                       # Go模块文件
├── go.sum                       # 依赖校验
├── Dockerfile                   # Docker配置
├── docker-compose.yml           # Docker Compose配置
├── Makefile                     # 构建脚本
└── README.md                    # 项目说明
```

## 📝 代码规范

### 1. Go代码规范

#### 命名规范
```go
// 包名：小写，简短
package controllers

// 变量名：驼峰命名
var userName string
var isActive bool

// 常量名：大写，下划线分隔
const (
    MAX_RETRY_COUNT = 3
    DEFAULT_TIMEOUT = 30 * time.Second
)

// 函数名：驼峰命名
func getUserByID(id uint) (*User, error) {
    // 实现
}

// 结构体名：驼峰命名，首字母大写
type UserController struct {
    userService *Services.UserService
}

// 接口名：驼峰命名，通常以er结尾
type UserService interface {
    GetUser(id uint) (*User, error)
    CreateUser(user *User) error
}
```

#### 注释规范
```go
// UserController 用户控制器
// 功能说明：
// 1. 处理用户相关的HTTP请求
// 2. 提供用户CRUD操作
// 3. 处理用户认证和授权
type UserController struct {
    userService *Services.UserService
}

// GetUser 获取用户信息
// 功能说明：
// 1. 根据用户ID获取用户详细信息
// 2. 验证用户权限
// 3. 返回用户数据
// 参数：
//   - ctx: Gin上下文
// 返回：
//   - 用户信息或错误
func (c *UserController) GetUser(ctx *gin.Context) {
    // 实现
}
```

#### 错误处理
```go
// 使用errors包创建错误
import "errors"

// 定义错误常量
var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidInput = errors.New("invalid input")
)

// 错误处理示例
func (s *UserService) GetUser(id uint) (*User, error) {
    if id == 0 {
        return nil, ErrInvalidInput
    }
    
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    if user == nil {
        return nil, ErrUserNotFound
    }
    
    return user, nil
}
```

### 2. 项目特定规范

#### 控制器规范
```go
// 控制器应该：
// 1. 只处理HTTP请求和响应
// 2. 调用Service层处理业务逻辑
// 3. 进行输入验证
// 4. 返回标准化的响应格式

func (c *UserController) CreateUser(ctx *gin.Context) {
    // 1. 绑定和验证请求
    var request Requests.CreateUserRequest
    if err := ctx.ShouldBindJSON(&request); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request data",
            "errors":  err.Error(),
        })
        return
    }
    
    // 2. 调用Service层
    user, err := c.userService.CreateUser(request)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to create user",
            "error":   err.Error(),
        })
        return
    }
    
    // 3. 返回成功响应
    ctx.JSON(http.StatusCreated, gin.H{
        "success": true,
        "message": "User created successfully",
        "data":    user,
    })
}
```

#### Service层规范
```go
// Service层应该：
// 1. 包含业务逻辑
// 2. 调用Repository层
// 3. 处理事务
// 4. 返回业务错误

func (s *UserService) CreateUser(request Requests.CreateUserRequest) (*Models.User, error) {
    // 1. 业务验证
    if err := s.validateCreateUserRequest(request); err != nil {
        return nil, err
    }
    
    // 2. 检查重复
    existingUser, err := s.userRepo.FindByEmail(request.Email)
    if err != nil {
        return nil, fmt.Errorf("failed to check existing user: %w", err)
    }
    
    if existingUser != nil {
        return nil, ErrUserAlreadyExists
    }
    
    // 3. 创建用户
    user := &Models.User{
        Username: request.Username,
        Email:    request.Email,
        Password: request.Password,
        Role:     "user",
        Status:   1,
    }
    
    if err := s.userRepo.Create(user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}
```

## 🔄 开发流程

### 1. 功能开发流程

```bash
# 1. 创建功能分支
git checkout -b feature/user-management

# 2. 开发功能
# - 编写测试
# - 实现功能
# - 更新文档

# 3. 运行测试
go test ./...

# 4. 代码检查
golangci-lint run

# 5. 提交代码
git add .
git commit -m "feat: add user management functionality"

# 6. 推送分支
git push origin feature/user-management

# 7. 创建Pull Request
```

### 2. 数据库迁移流程

```bash
# 1. 创建迁移文件
go run scripts/migrate.go create create_users_table

# 2. 编辑迁移文件
# 在 app/Database/Migrations/ 目录下编辑生成的迁移文件

# 3. 运行迁移
go run scripts/migrate.go migrate

# 4. 回滚迁移（如果需要）
go run scripts/migrate.go rollback
```

### 3. API开发流程

```bash
# 1. 定义API规范
# 在docs/API.md中定义API接口

# 2. 创建请求验证结构
# 在app/Http/Requests/中创建验证结构

# 3. 实现控制器
# 在app/Http/Controllers/中实现控制器

# 4. 添加路由
# 在app/Http/Routes/routes.go中添加路由

# 5. 实现Service层
# 在app/Services/中实现业务逻辑

# 6. 编写测试
# 在tests/中编写测试用例

# 7. 更新API文档
swag init
```

## 🧪 测试指南

### 1. 单元测试

```go
// 测试文件命名：*_test.go
package Services

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

// UserServiceTestSuite 用户服务测试套件
type UserServiceTestSuite struct {
    suite.Suite
    userService *UserService
    mockRepo    *MockUserRepository
}

// SetupSuite 测试套件初始化
func (suite *UserServiceTestSuite) SetupSuite() {
    suite.mockRepo = NewMockUserRepository()
    suite.userService = NewUserService(suite.mockRepo)
}

// TestCreateUser 测试创建用户
func (suite *UserServiceTestSuite) TestCreateUser() {
    // 准备测试数据
    request := Requests.CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }
    
    // 设置Mock期望
    suite.mockRepo.On("FindByEmail", request.Email).Return(nil, nil)
    suite.mockRepo.On("Create", mock.AnythingOfType("*Models.User")).Return(nil)
    
    // 执行测试
    user, err := suite.userService.CreateUser(request)
    
    // 验证结果
    suite.NoError(err)
    suite.NotNil(user)
    suite.Equal(request.Username, user.Username)
    suite.Equal(request.Email, user.Email)
    
    // 验证Mock调用
    suite.mockRepo.AssertExpectations(suite.T())
}
```

### 2. 集成测试

```go
// 集成测试示例
func TestUserControllerIntegration(t *testing.T) {
    // 设置测试数据库
    db := setupTestDatabase()
    defer cleanupTestDatabase(db)
    
    // 创建测试应用
    app := setupTestApp(db)
    
    // 创建测试请求
    requestBody := `{
        "username": "testuser",
        "email": "test@example.com",
        "password": "password123"
    }`
    
    // 发送请求
    req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(requestBody))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    app.ServeHTTP(w, req)
    
    // 验证响应
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
}
```

### 3. 性能测试

```go
// 性能测试示例
func BenchmarkUserService_CreateUser(b *testing.B) {
    service := setupBenchmarkService()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        request := Requests.CreateUserRequest{
            Username: fmt.Sprintf("user%d", i),
            Email:    fmt.Sprintf("user%d@example.com", i),
            Password: "password123",
        }
        
        _, err := service.CreateUser(request)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 🔌 API开发

### 1. API设计原则

- **RESTful设计**: 使用标准HTTP方法和状态码
- **版本控制**: API版本通过URL路径控制
- **标准化响应**: 统一的响应格式
- **错误处理**: 详细的错误信息和状态码
- **文档化**: 完整的API文档

### 2. 响应格式

```go
// 成功响应
{
    "success": true,
    "message": "Operation completed successfully",
    "data": {
        // 响应数据
    },
    "timestamp": "2024-01-01T12:00:00Z"
}

// 错误响应
{
    "success": false,
    "message": "Operation failed",
    "error": "Detailed error message",
    "errors": {
        "field": "Field validation error"
    },
    "timestamp": "2024-01-01T12:00:00Z"
}
```

### 3. 分页响应

```go
// 分页响应格式
{
    "success": true,
    "message": "Data retrieved successfully",
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

### 4. API文档生成

```go
// 使用Swagger注释
// @Summary 创建用户
// @Description 创建新用户账户
// @Tags users
// @Accept json
// @Produce json
// @Param user body Requests.CreateUserRequest true "用户信息"
// @Success 201 {object} Responses.UserResponse
// @Failure 400 {object} Responses.ErrorResponse
// @Failure 500 {object} Responses.ErrorResponse
// @Router /api/v1/users [post]
func (c *UserController) CreateUser(ctx *gin.Context) {
    // 实现
}
```

## 🗄️ 数据库操作

### 1. 模型定义

```go
// 用户模型示例
type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Username  string         `json:"username" gorm:"size:50;uniqueIndex;not null"`
    Email     string         `json:"email" gorm:"size:100;uniqueIndex;not null"`
    Password  string         `json:"-" gorm:"size:255;not null"`
    Role      string         `json:"role" gorm:"size:20;default:'user'"`
    Status    int            `json:"status" gorm:"default:1"`
    Avatar    string         `json:"avatar" gorm:"size:255"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
    
    // 关联关系
    Posts []Post `json:"posts,omitempty" gorm:"foreignKey:UserID"`
}

// 表名
func (User) TableName() string {
    return "users"
}

// 钩子方法
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 密码加密
    hashedPassword, err := Utils.HashPassword(u.Password)
    if err != nil {
        return err
    }
    u.Password = hashedPassword
    return nil
}
```

### 2. 查询操作

```go
// 基础查询
func (r *UserRepository) FindByID(id uint) (*User, error) {
    var user User
    err := r.db.First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// 条件查询
func (r *UserRepository) FindByEmail(email string) (*User, error) {
    var user User
    err := r.db.Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// 分页查询
func (r *UserRepository) FindWithPagination(page, perPage int) ([]User, int64, error) {
    var users []User
    var total int64
    
    // 获取总数
    if err := r.db.Model(&User{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // 获取分页数据
    offset := (page - 1) * perPage
    err := r.db.Offset(offset).Limit(perPage).Find(&users).Error
    if err != nil {
        return nil, 0, err
    }
    
    return users, total, nil
}

// 关联查询
func (r *UserRepository) FindWithPosts(id uint) (*User, error) {
    var user User
    err := r.db.Preload("Posts").First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

### 3. 事务处理

```go
// 事务示例
func (s *UserService) CreateUserWithProfile(userData *User, profileData *Profile) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 创建用户
        if err := tx.Create(userData).Error; err != nil {
            return err
        }
        
        // 创建用户资料
        profileData.UserID = userData.ID
        if err := tx.Create(profileData).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

## 🐛 调试技巧

### 1. 日志调试

```go
// 使用结构化日志
log.Info("User created successfully", map[string]interface{}{
    "user_id": user.ID,
    "username": user.Username,
    "email": user.Email,
})

// 使用不同日志级别
log.Debug("Processing user request", map[string]interface{}{
    "request_data": request,
})

log.Warning("User login failed", map[string]interface{}{
    "username": username,
    "ip_address": ctx.ClientIP(),
})

log.Error("Database connection failed", map[string]interface{}{
    "error": err.Error(),
})
```

### 2. 性能分析

```go
// 使用pprof进行性能分析
import _ "net/http/pprof"

// 在main.go中添加
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// 分析CPU使用
go tool pprof http://localhost:6060/debug/pprof/profile

// 分析内存使用
go tool pprof http://localhost:6060/debug/pprof/heap

// 分析goroutine
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### 3. 热重载开发

```bash
# 使用air进行热重载
# 安装air
go install github.com/cosmtrek/air@latest

# 创建.air.toml配置文件
cat > .air.toml << EOF
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false
EOF

# 启动热重载
air
```

### 4. 配置热重载开发
```go
// 配置热重载示例
func setupConfigHotReload() {
    // 创建热重载管理器
    hotReloadManager := Config.NewHotReloadManager("config.yaml")
    
    // 添加重载回调
    hotReloadManager.AddReloadCallback(func(config *Config.Config) {
        log.Println("配置已重载")
        // 更新相关服务配置
        updateServiceConfigs(config)
    })
    
    // 开始监控
    if err := hotReloadManager.StartWatching(); err != nil {
        log.Fatal("启动配置热重载失败:", err)
    }
    
    // 优雅关闭
    defer hotReloadManager.StopWatching()
}
```

### 5. 熔断器开发
```go
// 熔断器使用示例
func setupCircuitBreaker() {
    // 创建熔断器
    circuitBreaker := NewCircuitBreaker("user-service", CircuitBreakerConfig{
        MaxRequests: 10,
        Interval:    time.Minute,
        Timeout:     time.Second * 30,
    })
    
    // 在服务调用中使用
    if circuitBreaker.AllowRequest() {
        result, err := callExternalService()
        circuitBreaker.RecordResult(err == nil, time.Since(start))
        return result, err
    } else {
        return nil, errors.New("熔断器开启，请求被拒绝")
    }
}
```

## ⚡ 性能优化

### 1. 数据库优化

```go
// 使用索引
// 在模型中定义索引
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Email    string `gorm:"uniqueIndex"`
    Username string `gorm:"index"`
}

// 使用连接池
func InitDB() {
    sqlDB, err := DB.DB()
    if err != nil {
        log.Fatal(err)
    }
    
    // 设置连接池参数
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
}

// 使用预加载避免N+1问题
func (r *UserRepository) FindUsersWithPosts() ([]User, error) {
    var users []User
    err := r.db.Preload("Posts").Find(&users).Error
    return users, err
}
```

### 2. 缓存优化

```go
// Redis缓存示例
func (s *UserService) GetUserWithCache(id uint) (*User, error) {
    // 尝试从缓存获取
    cacheKey := fmt.Sprintf("user:%d", id)
    cachedUser, err := s.redis.Get(cacheKey)
    if err == nil {
        var user User
        json.Unmarshal([]byte(cachedUser), &user)
        return &user, nil
    }
    
    // 从数据库获取
    user, err := s.userRepo.FindByID(id)
    if err != nil {
        return nil, err
    }
    
    // 缓存到Redis
    userJSON, _ := json.Marshal(user)
    s.redis.Set(cacheKey, string(userJSON), time.Hour)
    
    return user, nil
}
```

### 3. 并发优化

```go
// 使用goroutine处理并发任务
func (s *UserService) ProcessUsers(users []User) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(users))
    
    for _, user := range users {
        wg.Add(1)
        go func(u User) {
            defer wg.Done()
            if err := s.processUser(u); err != nil {
                errChan <- err
            }
        }(user)
    }
    
    wg.Wait()
    close(errChan)
    
    // 检查错误
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## 🤝 贡献指南

### 1. 贡献流程

1. **Fork项目**: 在GitHub上fork项目到你的账户
2. **创建分支**: 创建功能分支或修复分支
3. **开发功能**: 按照代码规范开发功能
4. **编写测试**: 为新增功能编写测试用例
5. **提交代码**: 使用规范的提交信息
6. **创建PR**: 创建Pull Request并描述变更

### 2. 提交信息规范

```bash
# 提交信息格式
<type>(<scope>): <subject>

# 类型说明
feat:     新功能
fix:      修复bug
docs:     文档更新
style:    代码格式调整
refactor: 代码重构
test:     测试相关
chore:    构建过程或辅助工具的变动

# 示例
feat(user): add user registration functionality
fix(auth): resolve JWT token validation issue
docs(api): update API documentation
style(controller): format code according to standards
```

### 3. 代码审查

- 所有代码变更都需要通过代码审查
- 确保代码符合项目规范
- 测试覆盖率不低于80%
- 性能影响评估
- 安全性检查

### 4. 问题报告

报告问题时请包含：

- 问题描述
- 复现步骤
- 期望行为
- 实际行为
- 环境信息
- 错误日志

## 📚 学习资源

### 1. Go语言学习

- [Go官方文档](https://golang.org/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://golang.org/doc/effective_go.html)

### 2. Web开发

- [Gin框架文档](https://gin-gonic.com/docs/)
- [GORM文档](https://gorm.io/docs/)
- [JWT认证](https://jwt.io/)

### 3. 最佳实践

- [Go项目结构](https://github.com/golang-standards/project-layout)
- [Go代码规范](https://github.com/golang/go/wiki/CodeReviewComments)
- [RESTful API设计](https://restfulapi.net/)

## 📞 获取帮助

- **文档**: 查看项目文档
- **Issues**: 在GitHub上提交Issue
- **讨论**: 参与项目讨论
- **邮件**: 联系项目维护者

---

感谢您为Cloud Platform API项目做出贡献！
