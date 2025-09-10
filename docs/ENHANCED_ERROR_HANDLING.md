# 增强错误处理和日志记录使用指南

## 概述

本项目实现了增强的错误处理和日志记录系统，提供了更强大的错误管理、分类、追踪和日志记录功能。

## 主要特性

### 1. 增强错误处理

- **错误分类**：按业务逻辑、技术类型等分类错误
- **严重程度**：低、中、高、严重四个级别
- **错误链**：支持错误原因追踪
- **上下文信息**：包含用户ID、请求ID、IP地址等
- **可恢复性**：标记错误是否可恢复或重试
- **堆栈跟踪**：自动记录错误发生位置

### 2. 增强日志记录

- **结构化日志**：JSON格式，便于分析
- **上下文感知**：自动提取请求上下文信息
- **性能监控**：记录请求耗时和慢请求
- **多级别日志**：Debug、Info、Warn、Error、Fatal
- **彩色输出**：支持终端彩色显示
- **文件轮转**：支持日志文件自动轮转

## 使用方法

### 1. 基础错误处理

```go
package main

import (
    "cloud-platform-api/app/Utils"
    "net/http"
)

func main() {
    // 创建基础错误
    err := Utils.NewEnhancedError("USER_NOT_FOUND", "User not found", http.StatusNotFound).
        WithCategory(Utils.CategoryBusiness).
        WithSeverity(Utils.SeverityLow).
        WithDetails("User ID: 12345").
        WithUserID("12345").
        WithRequestID("req-123").
        WithStackTrace()
    
    // 记录错误
    Utils.LogError(err)
}
```

### 2. 错误构建器

```go
// 使用错误构建器
err := Utils.NewErrorBuilder("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest).
    Category(Utils.CategoryValidation).
    Severity(Utils.SeverityLow).
    Details("Email format is invalid").
    WithContext("field", "email").
    WithContext("value", "invalid-email").
    UserID("user-123").
    RequestID("req-456").
    StackTrace().
    Build()

Utils.LogError(err)
```

### 3. 错误包装

```go
// 包装现有错误
originalErr := errors.New("database connection failed")
enhancedErr := Utils.WrapEnhancedError(originalErr, "DB_CONNECTION_FAILED", "Database connection failed").
    WithCategory(Utils.CategoryDatabase).
    WithSeverity(Utils.SeverityHigh).
    SetRetryable(true)

Utils.LogError(enhancedErr)
```

### 4. 上下文错误包装

```go
func handleRequest(c *gin.Context) {
    ctx := c.Request.Context()
    
    // 执行业务逻辑
    err := someBusinessLogic()
    if err != nil {
        // 使用上下文包装错误
        enhancedErr := Utils.WrapWithContext(ctx, err, "BUSINESS_ERROR", "Business logic failed")
        Utils.LogError(enhancedErr)
        
        c.JSON(enhancedErr.Status, gin.H{
            "success": false,
            "message": enhancedErr.Message,
            "code":    enhancedErr.Code,
        })
        return
    }
}
```

### 5. 错误收集器

```go
// 创建错误收集器
collector := Utils.NewErrorCollector()

// 添加多个错误
collector.AddError(err1, "ERROR_1", "First error")
collector.AddError(err2, "ERROR_2", "Second error")
collector.Add(enhancedErr)

// 检查是否有错误
if collector.HasErrors() {
    // 获取所有错误
    errors := collector.GetErrors()
    
    // 获取严重错误
    criticalErrors := collector.GetCriticalErrors()
    
    // 获取可重试错误
    retryableErrors := collector.GetRetryableErrors()
    
    // 转换为JSON
    jsonErrors := collector.ToJSON()
    fmt.Println(jsonErrors)
}
```

### 6. 增强日志记录

```go
// 创建日志记录器
logger := Utils.NewEnhancedLogger(&Utils.LoggerConfig{
    Level:       Utils.LogLevelInfo,
    EnableJSON:  true,
    EnableColor: false,
    Output:      "stdout",
})

// 基础日志
logger.Info("Application started")
logger.Error("Something went wrong")

// 带字段的日志
logger.WithField("user_id", "123").Info("User logged in")
logger.WithFields(map[string]interface{}{
    "user_id": "123",
    "action":  "login",
    "ip":      "192.168.1.1",
}).Info("User action")

// 带上下文的日志
ctx := context.WithValue(context.Background(), "user_id", "123")
logger.WithContextValue(ctx).Info("Context-aware logging")

// 记录增强错误
logger.LogError(enhancedErr)
```

### 7. 中间件使用

```go
// 创建增强错误处理中间件
config := &Middleware.EnhancedErrorHandlingConfig{
    EnableDetailedErrors:    true,
    EnableErrorTracking:     true,
    EnablePerformanceLog:    true,
    SlowRequestThreshold:    5 * time.Second,
    EnableRequestID:         true,
    EnableUserTracking:      true,
    EnableSecurityLogging:   true,
    IncludeStackTrace:       false,
}

middleware := Middleware.NewEnhancedErrorHandlingMiddleware(config)

// 在Gin路由中使用
router.Use(middleware.Handle())
```

### 8. 自定义错误处理器

```go
// 创建自定义错误处理器
handler := Middleware.NewCustomErrorHandler()

// 注册特定分类的错误处理器
handler.RegisterHandler(Utils.CategoryAuth, func(c *gin.Context, err *Utils.EnhancedError) {
    // 自定义认证错误处理
    c.JSON(err.Status, gin.H{
        "success": false,
        "message": "Authentication failed",
        "code":    err.Code,
        "redirect": "/login",
    })
})

handler.RegisterHandler(Utils.CategoryValidation, func(c *gin.Context, err *Utils.EnhancedError) {
    // 自定义验证错误处理
    c.JSON(err.Status, gin.H{
        "success": false,
        "message": "Validation failed",
        "code":    err.Code,
        "errors":  err.Context,
    })
})
```

## 错误分类

### 1. 认证相关 (CategoryAuth)
- 无效token
- token过期
- 未授权访问
- 权限不足

### 2. 验证相关 (CategoryValidation)
- 输入验证失败
- 格式错误
- 必填字段缺失

### 3. 数据库相关 (CategoryDatabase)
- 连接失败
- 查询错误
- 事务失败

### 4. 网络相关 (CategoryNetwork)
- 连接超时
- 网络不可达
- 服务不可用

### 5. 业务相关 (CategoryBusiness)
- 业务逻辑错误
- 资源不存在
- 资源冲突

### 6. 系统相关 (CategorySystem)
- 内部错误
- 配置错误
- 服务不可用

### 7. 安全相关 (CategorySecurity)
- 安全违规
- 频率限制
- 恶意请求

## 严重程度

### 1. 低 (SeverityLow)
- 用户输入错误
- 资源不存在
- 业务规则违反

### 2. 中 (SeverityMedium)
- 认证失败
- 权限不足
- 网络超时

### 3. 高 (SeverityHigh)
- 数据库错误
- 外部服务错误
- 系统配置错误

### 4. 严重 (SeverityCritical)
- 系统panic
- 致命错误
- 安全漏洞

## 最佳实践

### 1. 错误处理
- 使用适当的错误分类和严重程度
- 提供有意义的错误消息
- 包含足够的上下文信息
- 标记可重试的错误

### 2. 日志记录
- 使用结构化日志格式
- 包含请求ID和用户ID
- 记录关键业务事件
- 避免记录敏感信息

### 3. 性能监控
- 监控慢请求
- 记录关键性能指标
- 设置合理的阈值
- 定期分析日志

### 4. 安全考虑
- 不在日志中记录敏感信息
- 使用安全的日志存储
- 定期清理旧日志
- 监控异常访问模式

## 配置示例

```go
// 错误处理配置
errorConfig := &Middleware.EnhancedErrorHandlingConfig{
    EnableDetailedErrors:    os.Getenv("ENVIRONMENT") == "development",
    EnableErrorTracking:     true,
    EnablePerformanceLog:    true,
    SlowRequestThreshold:    5 * time.Second,
    EnableRequestID:         true,
    EnableUserTracking:      true,
    EnableSecurityLogging:   true,
    IncludeStackTrace:       os.Getenv("ENVIRONMENT") == "development",
    EnableErrorMetrics:      true,
}

// 日志配置
logConfig := &Utils.LoggerConfig{
    Level:       Utils.LogLevelInfo,
    EnableJSON:  true,
    EnableColor: os.Getenv("ENVIRONMENT") == "development",
    Output:      "stdout",
    FilePath:    "./logs/app.log",
}
```

## 监控和告警

### 1. 错误监控
- 监控错误频率和趋势
- 设置错误阈值告警
- 跟踪错误恢复时间
- 分析错误模式

### 2. 性能监控
- 监控请求响应时间
- 识别慢请求
- 跟踪资源使用情况
- 分析性能瓶颈

### 3. 安全监控
- 监控异常访问模式
- 检测安全威胁
- 跟踪认证失败
- 分析攻击尝试

这个增强的错误处理和日志记录系统为应用提供了强大的错误管理能力，有助于提高应用的可靠性、可维护性和安全性。
