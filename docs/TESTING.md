# 测试指南

## 概述

本项目提供了完整的测试框架，包括单元测试、集成测试和性能测试。特别针对JWT token认证相关的测试，提供了专门的TokenHelper工具。

## 测试环境设置

### 1. 环境变量配置

测试环境会自动设置以下环境变量：

```bash
SERVER_MODE=test
DB_DRIVER=sqlite
DB_DATABASE=:memory:
JWT_SECRET=test-jwt-secret-key-for-testing-only-32-chars
REDIS_HOST=
LOG_LEVEL=error
TEST_ENVIRONMENT=true
```

### 2. 数据库配置

测试使用SQLite内存数据库，确保测试的隔离性和速度。

## TokenHelper 使用指南

### 1. 基本用法

TokenHelper是专门为测试中处理JWT token而设计的工具类。

```go
// 在你的测试套件中继承TestSuite
type YourTestSuite struct {
    testsetup.TestSuite
}

// 创建测试用户并获取token
userInfo, err := ts.CreateTestUserWithToken("testuser", "test@example.com", "password123", "user")
ts.Require().NoError(err)

// 使用token
token := userInfo.Token
user := userInfo.User
```

### 2. 创建不同类型的用户

#### 创建普通用户
```go
userInfo, err := ts.CreateNormalUserWithToken()
ts.Require().NoError(err)
// userInfo.User.Username = "user"
// userInfo.User.Role = "user"
```

#### 创建管理员用户
```go
adminInfo, err := ts.CreateAdminUserWithToken()
ts.Require().NoError(err)
// adminInfo.User.Username = "admin"
// adminInfo.User.Role = "admin"
```

#### 创建自定义用户
```go
userInfo, err := ts.CreateTestUserWithToken("customuser", "custom@example.com", "password123", "user")
ts.Require().NoError(err)
```

### 3. 获取请求头

#### 获取标准请求头
```go
headers := ts.GetTestTokenHeaders(token)
// 包含: Authorization, Content-Type, Accept
```

#### 获取管理员请求头
```go
headers, err := ts.GetAdminTokenHeaders()
ts.Require().NoError(err)
// 包含: Authorization, Content-Type, Accept, X-Test-User-ID, X-Test-User-Role
```

#### 获取普通用户请求头
```go
headers, err := ts.GetUserTokenHeaders()
ts.Require().NoError(err)
// 包含: Authorization, Content-Type, Accept, X-Test-User-ID, X-Test-User-Role
```

### 4. Token验证

#### 验证token有效性
```go
ts.AssertTokenValid(token)
```

#### 验证token无效
```go
ts.AssertTokenInvalid("invalid.token")
```

#### 手动验证token
```go
err := ts.ValidateToken(token)
ts.Require().NoError(err)
```

### 5. 创建特殊token

#### 创建过期token（用于测试）
```go
expiredToken, err := ts.tokenHelper.CreateExpiredToken(user)
ts.Require().NoError(err)
ts.AssertTokenInvalid(expiredToken)
```

#### 创建无效token（用于测试）
```go
invalidToken := ts.tokenHelper.CreateInvalidToken()
ts.AssertTokenInvalid(invalidToken)
```

## HTTP测试示例

### 1. 基本HTTP测试

```go
func (ts *YourTestSuite) TestAuthenticatedEndpoint() {
    // 创建测试用户并获取token
    userInfo, err := ts.CreateTestUserWithToken("testuser", "test@example.com", "password123", "user")
    ts.Require().NoError(err)
    
    // 创建HTTP请求
    req, _ := http.NewRequest("GET", "/api/v1/auth/profile", nil)
    req.Header.Set("Authorization", "Bearer "+userInfo.Token)
    req.Header.Set("Content-Type", "application/json")
    
    // 执行请求
    w := httptest.NewRecorder()
    testsetup.Router.ServeHTTP(w, req)
    
    // 验证响应
    ts.Require().Equal(http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err = json.Unmarshal(w.Body.Bytes(), &response)
    ts.Require().NoError(err)
    
    ts.AssertResponseSuccess(response)
}
```

### 2. 使用预定义请求头

```go
func (ts *YourTestSuite) TestWithPredefinedHeaders() {
    // 获取管理员请求头
    headers, err := ts.GetAdminTokenHeaders()
    ts.Require().NoError(err)
    
    // 创建HTTP请求
    req, _ := http.NewRequest("GET", "/api/v1/admin/dashboard", nil)
    for key, value := range headers {
        req.Header.Set(key, value)
    }
    
    // 执行请求
    w := httptest.NewRecorder()
    testsetup.Router.ServeHTTP(w, req)
    
    // 验证响应
    ts.Require().Equal(http.StatusOK, w.Code)
}
```

### 3. 测试权限控制

```go
func (ts *YourTestSuite) TestPermissionControl() {
    // 创建普通用户
    userInfo, err := ts.CreateNormalUserWithToken()
    ts.Require().NoError(err)
    
    // 尝试访问管理员接口
    req, _ := http.NewRequest("GET", "/api/v1/admin/dashboard", nil)
    req.Header.Set("Authorization", "Bearer "+userInfo.Token)
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    testsetup.Router.ServeHTTP(w, req)
    
    // 应该被拒绝访问
    ts.Require().Equal(http.StatusForbidden, w.Code)
}
```

## 测试最佳实践

### 1. 测试隔离

每个测试都应该在干净的环境中运行：

```go
func (ts *YourTestSuite) SetupTest() {
    // 清理数据库
    ts.cleanupDatabase()
    
    // 清理测试文件
    ts.cleanupTestFiles()
}
```

### 2. 测试数据管理

使用有意义的测试数据：

```go
// 好的做法
userInfo, err := ts.CreateTestUserWithToken("testuser", "test@example.com", "password123", "user")

// 避免的做法
userInfo, err := ts.CreateTestUserWithToken("user", "user@test.com", "123", "user")
```

### 3. 错误处理

始终检查错误：

```go
userInfo, err := ts.CreateTestUserWithToken("testuser", "test@example.com", "password123", "user")
ts.Require().NoError(err) // 使用Require而不是Assert
```

### 4. 断言使用

使用合适的断言方法：

```go
// 验证成功响应
ts.AssertResponseSuccess(response)

// 验证错误响应
ts.AssertResponseError(response)

// 验证token
ts.AssertTokenValid(token)
ts.AssertTokenInvalid(invalidToken)
```

## 运行测试

### 1. 运行所有测试

```bash
go test ./...
```

### 2. 运行特定测试

```bash
# 运行TokenHelper示例测试
go test ./tests -run TestTokenHelperExample

# 运行认证控制器测试
go test ./tests/Controllers -run TestAuthController

# 运行认证服务测试
go test ./tests/Services -run TestAuthService
```

### 3. 运行测试并显示详细信息

```bash
go test -v ./...
```

### 4. 运行测试并生成覆盖率报告

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 常见问题

### 1. Token失效问题

如果遇到token失效的问题，检查：

- JWT密钥是否一致
- 测试环境变量是否正确设置
- 数据库是否正确初始化

### 2. 测试数据清理

确保每个测试后都清理数据：

```go
func (ts *YourTestSuite) TearDownTest() {
    ts.cleanupDatabase()
}
```

### 3. 并发测试

避免在测试中使用共享状态，每个测试应该独立运行。

### 4. 性能测试

对于性能测试，使用基准测试：

```go
func BenchmarkTokenGeneration(b *testing.B) {
    for i := 0; i < b.N; i++ {
        userInfo, err := ts.CreateTestUserWithToken("benchuser", "bench@test.com", "password123", "user")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 扩展TokenHelper

如果需要扩展TokenHelper功能，可以在`tests/testsetup/token_helper.go`中添加新方法：

```go
// 添加新的辅助方法
func (th *TokenHelper) CreateUserWithCustomRole(username, email, password, role string) (*UserTokenInfo, error) {
    // 实现自定义逻辑
}
```

然后在测试套件中添加对应的方法：

```go
func (ts *TestSuite) CreateUserWithCustomRole(username, email, password, role string) (*UserTokenInfo, error) {
    return ts.tokenHelper.CreateUserWithCustomRole(username, email, password, role)
}
```
