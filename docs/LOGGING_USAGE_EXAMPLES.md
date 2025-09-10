# 日志系统使用示例

## 📖 概述

本文档提供了云平台API日志系统的详细使用方法和代码示例，包括基础日志记录、高级功能使用、监控配置等。

## 🚀 快速开始

### 1. 基础日志记录

#### 简单日志记录
```go
import (
    "cloud-platform-api/app/Services"
    "cloud-platform-api/app/Config"
)

// 创建日志管理器
config := Config.GetDefaultLogConfig()
config.SetDefaults()

logManager := Services.NewLogManagerService(config)

// 记录不同级别的日志
logManager.Info("business", "用户登录成功", map[string]interface{}{
    "user_id": 123,
    "ip":      "192.168.1.1",
})

logManager.Warning("security", "检测到异常登录尝试", map[string]interface{}{
    "user_id": 123,
    "ip":      "192.168.1.100",
    "reason":  "IP地址异常",
})

logManager.Error("error", "数据库连接失败", map[string]interface{}{
    "error":   "connection timeout",
    "retries": 3,
})
```

#### 带上下文的日志记录
```go
import (
    "context"
)

// 创建带请求信息的上下文
ctx := context.WithValue(context.Background(), "request_id", "req_123")
ctx = context.WithValue(ctx, "user_id", uint(123))
ctx = context.WithValue(ctx, "client_ip", "192.168.1.1")
ctx = context.WithValue(ctx, "user_agent", "Mozilla/5.0...")

// 记录带上下文的日志
logManager.LogWithContext(ctx, "business", Config.LogLevelInfo, "用户操作", map[string]interface{}{
    "action": "update_profile",
    "field":  "email",
})
```

### 2. 专用日志方法

#### 请求日志
```go
// 在HTTP中间件中记录请求日志
func RequestLogMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 处理请求
        c.Next()
        
        // 记录请求日志
        duration := time.Since(start)
        logManager.LogRequest(c.Request.Context(), 
            c.Request.Method, 
            c.Request.URL.Path, 
            c.Writer.Status(), 
            duration, 
            map[string]interface{}{
                "user_agent": c.Request.UserAgent(),
                "referer":    c.Request.Referer(),
            })
    }
}
```

#### SQL日志
```go
// 在数据库操作中记录SQL日志
func LogSQLQuery(ctx context.Context, sql string, duration time.Duration, rows int64, err error) {
    logManager.LogSQL(ctx, sql, duration, rows, err, map[string]interface{}{
        "table":     "users",
        "operation": "SELECT",
    })
}

// 使用示例
start := time.Now()
var users []User
result := db.Find(&users)
duration := time.Since(start)

LogSQLQuery(ctx, result.Statement.SQL.String(), duration, int64(len(users)), result.Error)
```

#### 审计日志
```go
// 记录用户操作审计日志
func LogUserAction(ctx context.Context, action, resource string, resourceID interface{}, details map[string]interface{}) {
    logManager.LogAudit(ctx, action, resource, resourceID, details)
}

// 使用示例
LogUserAction(ctx, "create_user", "user", user.ID, map[string]interface{}{
    "username": user.Username,
    "email":    user.Email,
    "role":     user.Role,
})
```

#### 安全日志
```go
// 记录安全相关事件
func LogSecurityEvent(ctx context.Context, event string, level Config.LogLevel, details map[string]interface{}) {
    logManager.LogSecurity(ctx, event, level, details)
}

// 使用示例
LogSecurityEvent(ctx, "failed_login", Config.LogLevelWarning, map[string]interface{}{
    "username": "admin",
    "ip":       "192.168.1.100",
    "attempts": 5,
})
```

#### 业务日志
```go
// 记录业务操作日志
func LogBusinessOperation(ctx context.Context, module, action, message string, details map[string]interface{}) {
    logManager.LogBusiness(ctx, module, action, message, details)
}

// 使用示例
LogBusinessOperation(ctx, "order", "create", "订单创建成功", map[string]interface{}{
    "order_id": order.ID,
    "amount":   order.Amount,
    "user_id":  order.UserID,
})
```

#### 访问日志
```go
// 记录访问日志
func LogAccess(ctx context.Context, method, path string, statusCode int, userAgent string, details map[string]interface{}) {
    logManager.LogAccess(ctx, method, path, statusCode, userAgent, details)
}

// 使用示例
LogAccess(ctx, "GET", "/api/users", 200, userAgent, map[string]interface{}{
    "response_time": "150ms",
    "cache_hit":     true,
})
```

## 🔧 高级功能使用

### 1. 日志监控和告警

#### 创建监控规则
```go
// 创建自定义监控规则
rule := &Services.LogRule{
    ID:          "custom_error_rule",
    Name:        "自定义错误监控",
    Description: "监控特定模块的错误日志",
    Enabled:     true,
    Logger:      "business",
    Level:       Config.LogLevelError,
    Keywords:    []string{"payment", "order"},
    Threshold:   5,
    TimeWindow:  2 * time.Minute,
    AlertLevel:  "warning",
    Message:     "业务模块错误过多",
    Actions:     []string{"email", "webhook"},
}

// 添加规则到监控服务
logMonitor := Services.NewLogMonitorService(logManager, config)
err := logMonitor.AddRule(rule)
if err != nil {
    log.Printf("添加监控规则失败: %v", err)
}
```

#### 查看监控统计
```go
// 获取监控统计信息
stats := logMonitor.GetStats()
fmt.Printf("总规则数: %d\n", stats.TotalRules)
fmt.Printf("活跃规则数: %d\n", stats.ActiveRules)
fmt.Printf("总告警数: %d\n", stats.TotalAlerts)
fmt.Printf("活跃告警数: %d\n", stats.ActiveAlerts)

// 获取告警列表
alerts := logMonitor.GetAlerts("active", 10)
for _, alert := range alerts {
    fmt.Printf("告警: %s - %s\n", alert.Level, alert.Message)
}
```

### 2. 日志统计和分析

#### 获取日志统计
```go
// 获取总体统计
stats := logManager.GetStats()
fmt.Printf("总日志数: %d\n", stats.TotalLogs)

// 按级别统计
for level, count := range stats.LogsByLevel {
    fmt.Printf("%s: %d\n", level, count)
}

// 按记录器统计
for logger, count := range stats.LogsByLogger {
    fmt.Printf("%s: %d\n", logger, count)
}

// 获取性能指标
for metric, value := range stats.Performance {
    fmt.Printf("%s: %.3f\n", metric, value)
}
```

#### 获取特定记录器统计
```go
// 获取错误日志记录器统计
errorStats := logManager.GetLoggerStats("error")
if errorStats != nil {
    fmt.Printf("错误日志总数: %d\n", errorStats.TotalLogs)
    fmt.Printf("最后记录时间: %s\n", errorStats.LastLog)
    fmt.Printf("错误计数: %d\n", errorStats.ErrorCount)
    
    // 计算平均写入延迟
    if len(errorStats.WriteLatency) > 0 {
        var total float64
        for _, latency := range errorStats.WriteLatency {
            total += latency
        }
        avgLatency := total / float64(len(errorStats.WriteLatency))
        fmt.Printf("平均写入延迟: %.3fs\n", avgLatency)
    }
}
```

### 3. 日志配置管理

#### 动态配置更新
```go
// 更新日志级别
config.Level = Config.LogLevelDebug
config.RequestLog.Level = Config.LogLevelInfo
config.SQLLog.Level = Config.LogLevelWarning

// 更新轮转配置
config.Rotation.MaxSize = 200    // 200MB
config.Rotation.MaxAge = 168 * time.Hour  // 7天
config.Rotation.MaxBackups = 20

// 更新特定日志配置
config.ErrorLog.Enabled = true
config.ErrorLog.IncludeStack = true
config.ErrorLog.NotifyEmail = "admin@example.com"

config.SecurityLog.Enabled = true
config.SecurityLog.RealTime = true
config.SecurityLog.AlertLevel = Config.LogLevelWarning
```

## 📊 监控和告警配置

### 1. 默认监控规则

系统预置了以下监控规则：

#### 错误日志阈值监控
```yaml
rule:
  id: "error_threshold"
  name: "错误日志阈值"
  description: "监控错误日志数量，超过阈值时告警"
  enabled: true
  logger: "error"
  level: "error"
  threshold: 10
  time_window: "5m"
  alert_level: "warning"
  message: "错误日志数量过多，请检查系统状态"
  actions: ["email", "webhook"]
```

#### 慢查询检测
```yaml
rule:
  id: "slow_query_detection"
  name: "慢查询检测"
  description: "检测SQL慢查询"
  enabled: true
  logger: "sql"
  level: "warning"
  keywords: ["slow_query"]
  threshold: 5
  time_window: "1m"
  alert_level: "warning"
  message: "检测到多个慢查询，请优化数据库性能"
  actions: ["email"]
```

#### 安全事件监控
```yaml
rule:
  id: "security_events"
  name: "安全事件监控"
  description: "监控安全相关日志"
  enabled: true
  logger: "security"
  level: "warning"
  threshold: 1
  time_window: "1m"
  alert_level: "critical"
  message: "检测到安全事件，请立即处理"
  actions: ["email", "slack", "webhook"]
```

### 2. 自定义监控规则

#### 业务异常监控
```go
// 监控业务模块的异常情况
businessRule := &Services.LogRule{
    ID:          "business_exception",
    Name:        "业务异常监控",
    Description: "监控业务模块的异常和错误",
    Enabled:     true,
    Logger:      "business",
    Level:       Config.LogLevelError,
    Pattern:     `.*exception.*|.*error.*`,
    Threshold:   3,
    TimeWindow:  5 * time.Minute,
    AlertLevel:  "warning",
    Message:     "业务模块出现异常，请检查业务逻辑",
    Actions:     []string{"email", "slack"},
}
```

#### 性能监控规则
```go
// 监控系统性能指标
performanceRule := &Services.LogRule{
    ID:          "performance_monitor",
    Name:        "性能监控",
    Description: "监控系统性能相关日志",
    Enabled:     true,
    Logger:      "request",
    Level:       Config.LogLevelWarning,
    Keywords:    []string{"slow", "timeout", "performance"},
    Threshold:   10,
    TimeWindow:  10 * time.Minute,
    AlertLevel:  "warning",
    Message:     "系统性能下降，请检查资源使用情况",
    Actions:     []string{"email", "webhook"},
}
```

## 🔍 日志搜索和过滤

### 1. 基于条件的日志过滤

```go
// 根据日志记录器过滤
func FilterLogsByLogger(logs []LogEntry, loggerName string) []LogEntry {
    var filtered []LogEntry
    for _, log := range logs {
        if log.Logger == loggerName {
            filtered = append(filtered, log)
        }
    }
    return filtered
}

// 根据日志级别过滤
func FilterLogsByLevel(logs []LogEntry, level Config.LogLevel) []LogEntry {
    var filtered []LogEntry
    for _, log := range logs {
        if log.Level >= level {
            filtered = append(filtered, log)
        }
    }
    return filtered
}

// 根据时间范围过滤
func FilterLogsByTimeRange(logs []LogEntry, start, end time.Time) []LogEntry {
    var filtered []LogEntry
    for _, log := range logs {
        if log.Timestamp.After(start) && log.Timestamp.Before(end) {
            filtered = append(filtered, log)
        }
    }
    return filtered
}

// 根据关键词过滤
func FilterLogsByKeywords(logs []LogEntry, keywords []string) []LogEntry {
    var filtered []LogEntry
    for _, log := range logs {
        for _, keyword := range keywords {
            if strings.Contains(log.Message, keyword) {
                filtered = append(filtered, log)
                break
            }
        }
    }
    return filtered
}
```

### 2. 高级搜索功能

```go
// 复合条件搜索
func SearchLogs(logs []LogEntry, criteria LogSearchCriteria) []LogEntry {
    var filtered []LogEntry
    
    for _, log := range logs {
        // 检查是否匹配所有条件
        if matchesCriteria(log, criteria) {
            filtered = append(filtered, log)
        }
    }
    
    return filtered
}

type LogSearchCriteria struct {
    Logger    string
    Level     Config.LogLevel
    Keywords  []string
    StartTime time.Time
    EndTime   time.Time
    Fields    map[string]interface{}
}

func matchesCriteria(log LogEntry, criteria LogSearchCriteria) bool {
    // 检查日志记录器
    if criteria.Logger != "" && log.Logger != criteria.Logger {
        return false
    }
    
    // 检查日志级别
    if criteria.Level != "" && log.Level < criteria.Level {
        return false
    }
    
    // 检查关键词
    if len(criteria.Keywords) > 0 {
        matched := false
        for _, keyword := range criteria.Keywords {
            if strings.Contains(log.Message, keyword) {
                matched = true
                break
            }
        }
        if !matched {
            return false
        }
    }
    
    // 检查时间范围
    if !criteria.StartTime.IsZero() && log.Timestamp.Before(criteria.StartTime) {
        return false
    }
    if !criteria.EndTime.IsZero() && log.Timestamp.After(criteria.EndTime) {
        return false
    }
    
    // 检查字段匹配
    if len(criteria.Fields) > 0 {
        for key, value := range criteria.Fields {
            if log.Fields[key] != value {
                return false
            }
        }
    }
    
    return true
}
```

## 📈 性能优化建议

### 1. 异步日志记录

```go
// 使用异步日志记录提高性能
func AsyncLogExample() {
    // 日志会自动异步处理，无需等待
    logManager.Info("business", "异步日志记录", map[string]interface{}{
        "timestamp": time.Now(),
        "data":      "大量数据",
    })
    
    // 继续执行其他操作
    fmt.Println("日志已记录，继续执行...")
}
```

### 2. 批量日志记录

```go
// 批量记录日志以提高效率
func BatchLogExample() {
    logs := []LogEntry{
        {Logger: "business", Level: Config.LogLevelInfo, Message: "操作1"},
        {Logger: "business", Level: Config.LogLevelInfo, Message: "操作2"},
        {Logger: "business", Level: Config.LogLevelInfo, Message: "操作3"},
    }
    
    for _, log := range logs {
        logManager.Log(log.Logger, log.Level, log.Message, nil)
    }
}
```

### 3. 日志级别控制

```go
// 在生产环境中使用适当的日志级别
func ProductionLogExample() {
    // 生产环境通常使用 Info 级别
    config.Level = Config.LogLevelInfo
    
    // Debug 日志在生产环境中不会记录
    logManager.Debug("business", "调试信息", nil)  // 不会记录
    
    // Info 及以上级别的日志会记录
    logManager.Info("business", "重要信息", nil)   // 会记录
    logManager.Warning("business", "警告信息", nil) // 会记录
    logManager.Error("business", "错误信息", nil)   // 会记录
}
```

## 🚨 故障排除

### 1. 常见问题

#### 日志文件权限问题
```bash
# 检查日志目录权限
ls -la storage/logs/

# 修复权限问题
chmod 755 storage/logs/
chown -R www-data:www-data storage/logs/
```

#### 磁盘空间不足
```bash
# 检查磁盘空间
df -h

# 清理旧日志文件
find storage/logs/ -name "*.log.*" -mtime +30 -delete

# 压缩旧日志
find storage/logs/ -name "*.log.*" -exec gzip {} \;
```

#### 日志轮转问题
```go
// 检查轮转配置
config.Rotation.MaxSize = 100      // 100MB
config.Rotation.MaxAge = 168 * time.Hour  // 7天
config.Rotation.MaxBackups = 10
config.Rotation.Compress = true

// 手动触发日志轮转
if closer, ok := logger.writer.(io.Closer); ok {
    closer.Close()
}
```

### 2. 性能问题诊断

```go
// 检查日志写入性能
stats := logManager.GetLoggerStats("system")
if stats != nil && len(stats.WriteLatency) > 0 {
    var total float64
    var max float64
    for _, latency := range stats.WriteLatency {
        total += latency
        if latency > max {
            max = latency
        }
    }
    avgLatency := total / float64(len(stats.WriteLatency))
    
    fmt.Printf("平均写入延迟: %.3fs\n", avgLatency)
    fmt.Printf("最大写入延迟: %.3fs\n", max)
    
    // 如果延迟过高，可能需要优化
    if avgLatency > 0.1 { // 100ms
        fmt.Println("警告: 日志写入延迟过高")
    }
}
```

### 3. 监控告警问题

```go
// 检查监控规则状态
rules := logMonitor.GetRules()
for _, rule := range rules {
    fmt.Printf("规则: %s, 状态: %v, 触发次数: %d\n", 
        rule.Name, rule.Enabled, rule.TriggerCount)
}

// 检查告警状态
alerts := logMonitor.GetAlerts("", 100)
for _, alert := range alerts {
    fmt.Printf("告警: %s - %s, 状态: %s\n", 
        alert.Level, alert.Message, alert.Status)
}
```

## 📚 最佳实践

### 1. 日志记录原则

- **结构化**: 使用结构化的字段记录日志
- **可搜索**: 包含便于搜索的关键信息
- **可操作**: 记录足够的信息以便问题诊断
- **性能友好**: 避免记录过多不必要的信息

### 2. 日志级别使用

- **Debug**: 详细的调试信息，仅在开发环境使用
- **Info**: 一般信息，记录重要的业务操作
- **Warning**: 警告信息，需要注意但不影响系统运行
- **Error**: 错误信息，系统出现问题
- **Fatal**: 致命错误，系统无法继续运行

### 3. 敏感信息处理

```go
// 避免记录敏感信息
func SafeLogExample() {
    user := getUser()
    
    // 错误示例：记录密码
    // logManager.Info("auth", "用户登录", map[string]interface{}{
    //     "username": user.Username,
    //     "password": user.Password,  // 不要记录密码
    // })
    
    // 正确示例：隐藏敏感信息
    logManager.Info("auth", "用户登录", map[string]interface{}{
        "username": user.Username,
        "user_id":  user.ID,
        "ip":       getClientIP(),
        "success":  true,
    })
}
```

### 4. 日志文件管理

- 定期清理旧日志文件
- 使用压缩减少存储空间
- 监控日志文件大小和数量
- 设置合理的轮转策略

## 🔮 未来扩展

### 1. 日志聚合和分析

- 集成ELK Stack (Elasticsearch, Logstash, Kibana)
- 支持日志的实时搜索和分析
- 提供可视化的日志仪表板

### 2. 机器学习集成

- 异常模式自动检测
- 智能告警阈值调整
- 日志趋势预测

### 3. 分布式日志

- 支持多节点日志收集
- 日志的统一存储和查询
- 跨服务的日志关联分析

---

通过以上示例和最佳实践，您可以充分利用云平台API的日志系统，实现高效的日志管理、监控和告警功能。
