# 日志管理系统使用说明

## 📋 概述

Cloud Platform API 采用了全新的日志管理系统，支持多种日志类型、分类存储、可配置输出和自动轮转等功能。

## 🏗️ 系统架构

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   应用层        │    │   日志管理器     │    │   存储层        │
│                │───▶│                │───▶│                │
│ - 控制器       │    │ - 日志分类      │    │ - 文件存储      │
│ - 服务层       │    │ - 格式转换      │    │ - 控制台输出    │
│ - 中间件       │    │ - 级别控制      │    │ - 远程存储      │
│ - 模型层       │    │ - 异步处理      │    │ - 日志轮转      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 📁 日志分类

### 1. **请求日志 (Request Log)**
- **存储路径**: `./storage/logs/requests/`
- **内容**: HTTP请求和响应的详细信息
- **用途**: 访问分析、性能监控、安全审计

### 2. **SQL日志 (SQL Log)**
- **存储路径**: `./storage/logs/sql/`
- **内容**: 数据库查询语句、执行时间、慢查询
- **用途**: 性能优化、问题排查、安全监控

### 3. **错误日志 (Error Log)**
- **存储路径**: `./storage/logs/errors/`
- **内容**: 系统错误、异常信息、堆栈跟踪
- **用途**: 问题诊断、错误统计、告警通知

### 4. **审计日志 (Audit Log)**
- **存储路径**: `./storage/logs/audit/`
- **内容**: 用户操作、权限变更、系统配置
- **用途**: 合规审计、安全监控、操作追踪

### 5. **安全日志 (Security Log)**
- **存储路径**: `./storage/logs/security/`
- **内容**: 安全事件、攻击尝试、异常访问
- **用途**: 安全监控、威胁检测、实时告警

### 6. **业务日志 (Business Log)**
- **存储路径**: `./storage/logs/business/`
- **内容**: 业务操作、业务流程、业务指标
- **用途**: 业务分析、流程监控、决策支持

### 7. **访问日志 (Access Log)**
- **存储路径**: `./storage/logs/access/`
- **内容**: 用户访问、页面浏览、资源下载
- **用途**: 用户行为分析、资源使用统计

## ⚙️ 配置说明

### 环境变量配置

```bash
# 全局日志配置
LOG_LEVEL=info                    # 日志级别: debug, info, warning, error, fatal
LOG_FORMAT=json                   # 日志格式: json, text, custom
LOG_OUTPUT=both                   # 输出方式: file, console, both
LOG_TIMESTAMP=true                # 是否包含时间戳
LOG_CALLER=true                   # 是否包含调用者信息
LOG_STACKTRACE=false              # 是否包含堆栈跟踪
LOG_BASE_PATH=./storage/logs     # 日志基础路径
LOG_MAX_SIZE=100                  # 单个日志文件最大大小(MB)
LOG_MAX_AGE=720h                 # 日志文件保留时间
LOG_MAX_BACKUPS=10               # 保留的日志文件数量
LOG_COMPRESS=true                # 是否压缩旧日志文件

# 请求日志配置
REQUEST_LOG_ENABLED=true          # 是否启用请求日志
REQUEST_LOG_LEVEL=info            # 请求日志级别
REQUEST_LOG_PATH=requests         # 请求日志存储路径
REQUEST_LOG_FORMAT=json           # 请求日志格式
REQUEST_LOG_INCLUDE_BODY=false    # 是否包含请求/响应体
REQUEST_LOG_MAX_BODY_SIZE=1024   # 最大记录体大小(KB)
REQUEST_LOG_FILTER_PATHS=/health,/metrics  # 过滤的路径
REQUEST_LOG_MASK_FIELDS=password,token,secret  # 需要脱敏的字段

# SQL日志配置
SQL_LOG_ENABLED=true              # 是否启用SQL日志
SQL_LOG_LEVEL=info                # SQL日志级别
SQL_LOG_PATH=sql                  # SQL日志存储路径
SQL_LOG_FORMAT=json               # SQL日志格式
SQL_LOG_SLOW_THRESHOLD=1s         # 慢查询阈值
SQL_LOG_INCLUDE_PARAMS=true       # 是否包含SQL参数
SQL_LOG_INCLUDE_STACK=false       # 是否包含调用栈
SQL_LOG_MAX_QUERY_SIZE=2048      # 最大SQL记录大小(KB)
```

### 配置文件示例

```yaml
# config/logging.yaml
logging:
  level: info
  format: json
  output: both
  timestamp: true
  caller: true
  stacktrace: false
  base_path: ./storage/logs
  
  rotation:
    max_size: 100
    max_age: 720h
    max_backups: 10
    compress: true
  
  request_log:
    enabled: true
    level: info
    path: requests
    format: json
    include_body: false
    max_body_size: 1024
    filter_paths:
      - /health
      - /metrics
    mask_fields:
      - password
      - token
      - secret
  
  sql_log:
    enabled: true
    level: info
    path: sql
    format: json
    slow_threshold: 1s
    include_params: true
    include_stack: false
    max_query_size: 2048
```

## 🚀 使用方法

### 1. 初始化日志管理器

```go
package main

import (
    "cloud-platform-api/app/Config"
    "cloud-platform-api/app/Services"
)

func main() {
    // 加载配置
    Config.LoadConfig()
    
    // 创建日志管理器
    logManager := Services.NewLogManagerService(Config.GetConfig().Logging)
    
    // 在应用中使用
    // ...
    
    // 应用结束时关闭日志管理器
    defer logManager.Close()
}
```

### 2. 在控制器中使用

```go
package Controllers

import (
    "cloud-platform-api/app/Services"
    "github.com/gin-gonic/gin"
)

type UserController struct {
    Controller
    logManager *Services.LogManagerService
}

func (c *UserController) CreateUser(ctx *gin.Context) {
    // 记录业务日志
    c.logManager.LogBusiness("创建用户", map[string]interface{}{
        "user_id":   user.ID,
        "username":  user.Username,
        "email":     user.Email,
        "ip":        ctx.ClientIP(),
        "user_agent": ctx.Request.UserAgent(),
    })
    
    // 业务逻辑...
}
```

### 3. 在服务层中使用

```go
package Services

import (
    "cloud-platform-api/app/Config"
)

type UserService struct {
    logManager *Services.LogManagerService
}

func (s *UserService) CreateUser(user *Models.User) error {
    // 记录操作日志
    s.logManager.LogAudit("用户创建", map[string]interface{}{
        "action":     "create_user",
        "user_id":    user.ID,
        "username":   user.Username,
        "timestamp":  time.Now(),
    })
    
    // 业务逻辑...
    
    return nil
}
```

### 4. 在中间件中使用

```go
package Middleware

import (
    "cloud-platform-api/app/Services"
)

type AuthMiddleware struct {
    logManager *Services.LogManagerService
}

func (m *AuthMiddleware) Handle() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 记录安全日志
        m.logManager.LogSecurity("用户认证", map[string]interface{}{
            "ip":         c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
            "path":       c.Request.URL.Path,
            "method":     c.Request.Method,
        })
        
        // 认证逻辑...
        c.Next()
    }
}
```

## 📊 日志格式

### JSON格式示例

```json
{
  "logger": "request",
  "level": "info",
  "message": "HTTP POST /api/v1/users 201 - 45.2ms",
  "timestamp": "2024-12-01T10:30:45.123Z",
  "fields": {
    "method": "POST",
    "path": "/api/v1/users",
    "status_code": 201,
    "duration_ms": 45,
    "client_ip": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "user_id": 123,
    "username": "john_doe"
  },
  "caller": {
    "file": "UserController.go",
    "line": 45,
    "function": "CreateUser"
  }
}
```

### 文本格式示例

```
2024-12-01 10:30:45.123 INFO [request] (UserController.go:45:CreateUser) HTTP POST /api/v1/users 201 - 45.2ms {"method":"POST","path":"/api/v1/users","status_code":201,"duration_ms":45,"client_ip":"192.168.1.100","user_agent":"Mozilla/5.0...","user_id":123,"username":"john_doe"}
```

## 🔧 高级功能

### 1. 日志轮转

系统自动按以下规则进行日志轮转：
- **大小轮转**: 单个文件超过配置的最大大小
- **时间轮转**: 按天、小时等时间间隔
- **数量控制**: 保留指定数量的日志文件
- **自动压缩**: 压缩旧的日志文件

### 2. 敏感数据脱敏

自动识别和脱敏以下敏感字段：
- 密码相关: `password`, `passwd`, `pwd`
- 认证相关: `token`, `auth`, `secret`
- 个人信息: `phone`, `id_card`, `ssn`

### 3. 性能优化

- **异步写入**: 日志异步写入，不阻塞主流程
- **缓冲管理**: 智能缓冲管理，平衡性能和可靠性
- **并发安全**: 支持高并发环境下的安全写入

### 4. 监控告警

- **日志级别监控**: 实时监控各类型日志的级别分布
- **异常检测**: 自动检测异常日志模式
- **实时告警**: 支持邮件、Webhook等告警方式

## 📈 性能指标

### 日志性能基准

| 指标 | 值 | 说明 |
|------|-----|------|
| 写入延迟 | < 1ms | 单条日志写入延迟 |
| 吞吐量 | > 10,000/s | 每秒可处理的日志条数 |
| 内存占用 | < 100MB | 日志管理器内存占用 |
| 磁盘I/O | < 10MB/s | 日志写入磁盘I/O |

### 资源使用建议

- **日志级别**: 生产环境建议使用 `info` 级别
- **文件大小**: 建议单个日志文件不超过100MB
- **保留时间**: 根据合规要求设置，一般30-90天
- **压缩策略**: 启用压缩可节省50-70%存储空间

## 🚨 故障排查

### 常见问题

1. **日志文件过大**
   - 检查 `LOG_MAX_SIZE` 配置
   - 检查日志轮转是否正常工作
   - 检查是否有大量错误日志

2. **日志写入失败**
   - 检查磁盘空间是否充足
   - 检查文件权限是否正确
   - 检查日志目录是否存在

3. **性能问题**
   - 检查日志级别是否过高
   - 检查是否启用了过多日志类型
   - 检查磁盘I/O性能

### 调试命令

```bash
# 查看日志文件大小
du -sh ./storage/logs/*/

# 查看最新日志
tail -f ./storage/logs/requests/request.log

# 查看错误日志
grep "ERROR" ./storage/logs/errors/error.log

# 查看慢查询日志
grep "slow_query" ./storage/logs/sql/sql.log
```

## 🔮 未来规划

### 短期目标 (1-2个月)
- [ ] 支持远程日志存储 (ELK Stack)
- [ ] 添加日志分析工具
- [ ] 实现日志搜索功能
- [ ] 支持结构化日志查询

### 中期目标 (3-6个月)
- [ ] 集成机器学习异常检测
- [ ] 支持多租户日志隔离
- [ ] 实现日志流式处理
- [ ] 添加日志可视化界面

### 长期目标 (6-12个月)
- [ ] 支持分布式日志收集
- [ ] 实现日志智能分析
- [ ] 支持日志合规审计
- [ ] 集成第三方日志服务

## 📞 技术支持

如有问题或建议，请通过以下方式联系：

- **项目Issues**: GitHub Issues
- **技术讨论**: GitHub Discussions
- **文档反馈**: Pull Request

---

**文档版本**: 1.0  
**最后更新**: 2024年12月  
**维护状态**: 活跃维护
