# 性能监控系统文档

## 📋 概述

性能监控系统是Cloud Platform API项目的核心组成部分，提供全面的系统性能监控、告警和分析功能。该系统能够实时监控系统资源、应用性能和业务指标，并在发现异常时及时发出告警，帮助运维团队快速定位和解决问题。

## 🏗️ 系统架构

### 核心组件

```
app/Config/performance_monitoring.go           # 性能监控配置
app/Models/PerformanceMetric.go               # 性能指标数据模型
app/Services/PerformanceMonitoringService.go  # 性能监控核心服务
app/Http/Controllers/PerformanceMonitoringController.go  # API控制器
app/Http/Middleware/PerformanceMonitoringMiddleware.go   # 性能监控中间件
app/Http/Routes/performance_monitoring.go     # 路由配置
docs/PERFORMANCE_MONITORING_SYSTEM.md         # 本文档
```

### 系统功能模块

1. **系统资源监控** - CPU、内存、磁盘、网络使用情况
2. **应用性能监控** - HTTP请求、数据库、缓存、Go运行时指标
3. **业务指标监控** - 用户活跃度、API使用情况、自定义指标
4. **告警管理** - 规则配置、实时告警、通知发送
5. **数据分析** - 历史数据查询、趋势分析、报告生成

## 🚀 快速开始

### 1. 环境变量配置

在 `env.example` 中添加以下配置：

```bash
# 性能监控基础配置
PERF_MON_ENABLED=true
PERF_MON_INTERVAL=30s
PERF_MON_RETENTION=7d
PERF_MON_BATCH_SIZE=100
PERF_MON_VERBOSE=false

# 系统资源监控
PERF_MON_SYSTEM_ENABLED=true
PERF_MON_CPU_ENABLED=true
PERF_MON_CPU_THRESHOLD=80.0
PERF_MON_MEMORY_ENABLED=true
PERF_MON_MEMORY_THRESHOLD=85.0
PERF_MON_DISK_ENABLED=true
PERF_MON_DISK_THRESHOLD=90.0
PERF_MON_NETWORK_ENABLED=true

# 应用性能监控
PERF_MON_APP_ENABLED=true
PERF_MON_HTTP_ENABLED=true
PERF_MON_HTTP_RESPONSE_TIME=1s
PERF_MON_HTTP_ERROR_RATE=0.05
PERF_MON_DB_ENABLED=true
PERF_MON_DB_SLOW_THRESHOLD=1s
PERF_MON_CACHE_ENABLED=true
PERF_MON_GO_ENABLED=true

# 业务指标监控
PERF_MON_BUSINESS_ENABLED=true
PERF_MON_USER_ACTIVITY_ENABLED=true
PERF_MON_API_USAGE_ENABLED=true

# 告警配置
PERF_MON_ALERTS_ENABLED=true
PERF_MON_ALERT_MAX=10
PERF_MON_ALERT_WINDOW=1h
PERF_MON_ALERT_COOLDOWN=5m

# 存储配置
PERF_MON_STORAGE_TYPE=memory
PERF_MON_STORAGE_COMPRESSION=true
PERF_MON_STORAGE_BATCH_SIZE=100
```

### 2. 服务初始化

```go
// 在主应用中初始化性能监控服务
package main

import (
    "cloud_platform/api/back/app/Config"
    "cloud_platform/api/back/app/Services"
    "cloud_platform/api/back/app/Http/Controllers"
    "cloud_platform/api/back/app/Http/Routes"
    "cloud_platform/api/back/app/Database"
)

func main() {
    // 初始化数据库
    db := Database.GetDB()
    
    // 创建性能监控配置
    perfConfig := &Config.PerformanceMonitoringConfig{}
    perfConfig.SetDefaults()
    perfConfig.BindEnvs()
    
    // 初始化性能监控服务
    monitoringService := Services.NewPerformanceMonitoringService(db, perfConfig)
    
    // 创建控制器
    perfController := Controllers.NewPerformanceMonitoringController()
    perfController.SetPerformanceMonitoringService(monitoringService)
    
    // 初始化路由
    router := gin.New()
    
    // 注册性能监控中间件
    Routes.RegisterPerformanceMiddleware(router, monitoringService)
    
    // 注册性能监控路由
    Routes.RegisterPerformanceMonitoringRoutes(router, perfController)
    
    // 启动服务器
    router.Run(":8080")
}
```

### 3. 数据库迁移

确保数据库中包含以下表结构：

```sql
-- 性能指标表
CREATE TABLE performance_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    metric_type VARCHAR(50) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    value DOUBLE NOT NULL,
    unit VARCHAR(20),
    labels JSON,
    timestamp DATETIME NOT NULL,
    source VARCHAR(50),
    severity VARCHAR(20),
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_metric_type (metric_type),
    INDEX idx_metric_name (metric_name),
    INDEX idx_timestamp (timestamp)
);

-- 系统资源指标表
CREATE TABLE system_resource_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    cpu_usage DOUBLE,
    cpu_load_1 DOUBLE,
    cpu_load_5 DOUBLE,
    cpu_load_15 DOUBLE,
    memory_usage DOUBLE,
    memory_total BIGINT,
    memory_used BIGINT,
    memory_free BIGINT,
    swap_usage DOUBLE,
    disk_usage DOUBLE,
    disk_total BIGINT,
    disk_used BIGINT,
    disk_free BIGINT,
    disk_read_rate BIGINT,
    disk_write_rate BIGINT,
    network_rx_rate BIGINT,
    network_tx_rate BIGINT,
    timestamp DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_timestamp (timestamp)
);

-- 应用指标表
CREATE TABLE application_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    request_count BIGINT,
    error_count BIGINT,
    error_rate DOUBLE,
    avg_response_time BIGINT,
    p50_response_time BIGINT,
    p95_response_time BIGINT,
    p99_response_time BIGINT,
    throughput DOUBLE,
    active_connections INT,
    database_connections INT,
    cache_hit_rate DOUBLE,
    cache_size BIGINT,
    go_routines INT,
    heap_alloc BIGINT,
    heap_sys BIGINT,
    gc_count INT,
    gc_pause BIGINT,
    timestamp DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_timestamp (timestamp)
);

-- 业务指标表
CREATE TABLE business_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    active_users INT,
    online_users INT,
    new_users INT,
    user_sessions INT,
    avg_session_duration BIGINT,
    api_call_count BIGINT,
    popular_endpoints JSON,
    business_operations BIGINT,
    revenue DOUBLE,
    conversion_rate DOUBLE,
    bounce_rate DOUBLE,
    custom_metrics JSON,
    timestamp DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_timestamp (timestamp)
);

-- 告警表
CREATE TABLE alerts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    rule_name VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    condition VARCHAR(10) NOT NULL,
    threshold DOUBLE NOT NULL,
    current_value DOUBLE NOT NULL,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    labels JSON,
    first_seen DATETIME NOT NULL,
    last_seen DATETIME NOT NULL,
    resolved_at DATETIME,
    acknowledged_at DATETIME,
    acknowledged_by VARCHAR(100),
    count INT DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_rule_name (rule_name),
    INDEX idx_metric_name (metric_name),
    INDEX idx_severity (severity),
    INDEX idx_status (status)
);

-- 告警规则表
CREATE TABLE alert_rules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) UNIQUE NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    condition VARCHAR(10) NOT NULL,
    threshold DOUBLE NOT NULL,
    duration BIGINT,
    severity VARCHAR(20) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    description TEXT,
    labels JSON,
    notification_channels JSON,
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_metric_name (metric_name)
);

-- 告警通知表
CREATE TABLE alert_notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    alert_id BIGINT NOT NULL,
    channel VARCHAR(50) NOT NULL,
    recipient VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    error TEXT,
    sent_at DATETIME,
    retry_count INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_alert_id (alert_id),
    INDEX idx_status (status)
);

-- 性能事件表
CREATE TABLE performance_events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    event_type VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    impact VARCHAR(20),
    status VARCHAR(20) NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    duration BIGINT,
    tags JSON,
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_event_type (event_type),
    INDEX idx_status (status)
);

-- 监控仪表板表
CREATE TABLE monitoring_dashboards (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    config JSON,
    is_public BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME
);
```

## 🔧 功能特性

### 1. 系统资源监控

监控服务器的基础资源使用情况：

- **CPU监控**：使用率、负载平均值、核心数
- **内存监控**：使用率、总内存、已用内存、空闲内存、交换分区
- **磁盘监控**：使用率、读写速率、IO操作统计
- **网络监控**：接收/发送速率、带宽使用情况

### 2. 应用性能监控

监控Go应用程序的运行状态：

- **HTTP指标**：请求数、响应时间、错误率、吞吐量
- **数据库指标**：连接池状态、查询性能、慢查询检测
- **缓存指标**：命中率、内存使用、操作统计
- **Go运行时**：Goroutine数量、内存分配、GC性能

### 3. 业务指标监控

监控业务相关的关键指标：

- **用户活动**：活跃用户数、在线用户数、会话时长
- **API使用**：调用次数、热门接口、速率限制状态
- **自定义指标**：支持业务特定的指标收集

### 4. 告警管理

提供灵活的告警配置和管理：

- **告警规则**：支持多种条件和阈值配置
- **告警级别**：Critical、Warning、Info三个级别
- **通知渠道**：邮件、Webhook、Slack等多种方式
- **告警状态**：触发、确认、解决的完整生命周期

## 📊 API接口

### 指标查询接口

#### 获取当前指标
```http
GET /api/v1/performance/current
Authorization: Bearer <token>
```

响应示例：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "metrics": {
      "system_resources": {
        "cpu_usage": 45.2,
        "memory_usage": 68.5,
        "disk_usage": 42.1
      },
      "application": {
        "request_count": 1250,
        "error_rate": 0.02,
        "avg_response_time": 125
      },
      "business": {
        "active_users": 342,
        "online_users": 89
      }
    },
    "timestamp": "2024-12-20T10:30:00Z"
  }
}
```

#### 按时间范围获取指标
```http
GET /api/v1/performance/metrics?metric_type=system_resources&start=2024-12-20T00:00:00Z&end=2024-12-20T23:59:59Z
Authorization: Bearer <token>
```

### 告警管理接口

#### 获取活跃告警
```http
GET /api/v1/performance/alerts/active
Authorization: Bearer <token>
```

#### 创建告警规则
```http
POST /api/v1/performance/alert-rules
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "CPU使用率过高",
  "metric_name": "cpu_usage",
  "condition": ">",
  "threshold": 80.0,
  "duration": 300,
  "severity": "warning",
  "enabled": true,
  "description": "CPU使用率超过80%时触发告警"
}
```

#### 确认告警
```http
POST /api/v1/performance/alerts/{alert_id}/acknowledge
Authorization: Bearer <token>
```

### 自定义指标接口

#### 记录自定义指标
```http
POST /api/v1/performance/custom-metrics
Authorization: Bearer <token>
Content-Type: application/json

{
  "metric_type": "business",
  "metric_name": "order_count",
  "value": 125.0,
  "labels": {
    "source": "web",
    "category": "electronics"
  }
}
```

### 系统健康接口

#### 获取系统健康状态
```http
GET /api/v1/performance/health
Authorization: Bearer <token>
```

```http
GET /health
# 公开接口，不需要认证
```

## ⚙️ 配置说明

### 基础监控配置

```yaml
performance_monitoring:
  base:
    enabled: true              # 是否启用监控
    interval: 30s              # 监控间隔
    retention_period: 168h     # 数据保留时间（7天）
    batch_size: 100            # 批量处理大小
    verbose_logging: false     # 详细日志
```

### 系统资源监控配置

```yaml
performance_monitoring:
  system_resources:
    enabled: true
    cpu:
      enabled: true
      usage_threshold: 80.0    # CPU使用率阈值
      cores: -1                # 监控核心数（-1为自动检测）
    memory:
      enabled: true
      usage_threshold: 85.0    # 内存使用率阈值
      leak_detection: true     # 内存泄漏检测
    disk:
      enabled: true
      usage_threshold: 90.0    # 磁盘使用率阈值
      paths: ["/", "/tmp"]     # 监控路径
      io_monitoring: true      # IO监控
    network:
      enabled: true
      interfaces: ["eth0", "en0"]        # 监控网络接口
      bandwidth_threshold: 104857600     # 带宽阈值（100MB/s）
```

### 应用性能监控配置

```yaml
performance_monitoring:
  application:
    enabled: true
    http:
      enabled: true
      response_time_threshold: 1s        # 响应时间阈值
      error_rate_threshold: 0.05         # 错误率阈值
      record_request_details: true       # 记录请求详情
      exclude_paths: ["/health", "/metrics"]  # 排除路径
    database:
      enabled: true
      connection_pool: true              # 连接池监控
      query_performance: true            # 查询性能监控
      slow_query_threshold: 1s           # 慢查询阈值
    cache:
      enabled: true
      hit_rate_threshold: 0.8            # 命中率阈值
      memory_usage: true                 # 内存使用监控
    go_runtime:
      enabled: true
      gc_monitoring: true                # GC监控
      goroutine_monitoring: true         # Goroutine监控
      heap_monitoring: true              # 堆内存监控
```

### 告警配置

```yaml
performance_monitoring:
  alerts:
    enabled: true
    channels:
      - name: "email"
        type: "email"
        enabled: true
        config:
          smtp_host: "smtp.example.com"
          smtp_port: 587
          username: "alerts@example.com"
          password: "password"
          recipients: ["admin@example.com"]
      - name: "webhook"
        type: "webhook"
        enabled: true
        config:
          url: "https://hooks.slack.com/services/..."
          timeout: "30s"
    rules:
      - name: "CPU使用率过高"
        metric: "cpu_usage"
        condition: ">"
        threshold: 80.0
        duration: 300s
        severity: "warning"
        enabled: true
    rate_limit:
      max_alerts: 10           # 最大告警次数
      time_window: 1h          # 时间窗口
      cooldown_period: 5m      # 冷却时间
```

## 🔍 使用示例

### 1. 基础监控使用

```go
// 在业务代码中记录自定义指标
func (s *OrderService) CreateOrder(order *Order) error {
    startTime := time.Now()
    
    // 执行业务逻辑
    err := s.processOrder(order)
    
    // 记录处理时间
    duration := time.Since(startTime)
    s.monitoringService.RecordCustomMetric(
        "business",
        "order_processing_time",
        float64(duration.Milliseconds()),
        map[string]string{
            "status": getOrderStatus(err),
            "category": order.Category,
        },
    )
    
    // 记录订单数
    s.monitoringService.RecordCustomMetric(
        "business",
        "order_count",
        1.0,
        map[string]string{
            "status": getOrderStatus(err),
        },
    )
    
    return err
}
```

### 2. 中间件使用

```go
// 在路由中使用性能监控中间件
func setupRoutes(router *gin.Engine, monitoringService *Services.PerformanceMonitoringService) {
    // 全局性能监控中间件
    perfMiddleware := Middleware.NewPerformanceMonitoringMiddleware(
        monitoringService,
        []string{"/health", "/metrics"}, // 排除路径
    )
    router.Use(perfMiddleware.Handler())
    
    // 业务指标中间件
    router.Use(Middleware.BusinessMetricsMiddleware(monitoringService))
    
    // 注册业务路由
    api := router.Group("/api/v1")
    {
        api.POST("/orders", orderController.CreateOrder)
        api.GET("/orders", orderController.GetOrders)
    }
}
```

### 3. 告警规则管理

```go
// 动态创建告警规则
func createAlertRules(service *Services.PerformanceMonitoringService) {
    rules := []*Models.AlertRule{
        {
            Name:        "高错误率告警",
            MetricName:  "error_rate",
            Condition:   ">",
            Threshold:   0.05, // 5%
            Duration:    5 * time.Minute,
            Severity:    "critical",
            Enabled:     true,
            Description: "API错误率超过5%时触发",
        },
        {
            Name:        "响应时间过长",
            MetricName:  "avg_response_time",
            Condition:   ">",
            Threshold:   2000, // 2秒
            Duration:    3 * time.Minute,
            Severity:    "warning",
            Enabled:     true,
            Description: "平均响应时间超过2秒时触发",
        },
    }
    
    for _, rule := range rules {
        if err := service.CreateAlertRule(rule); err != nil {
            log.Printf("创建告警规则失败: %v", err)
        }
    }
}
```

## 📈 监控指标详解

### 系统资源指标

| 指标名称 | 描述 | 单位 | 阈值建议 |
|---------|------|------|----------|
| cpu_usage | CPU使用率 | % | 80% |
| memory_usage | 内存使用率 | % | 85% |
| disk_usage | 磁盘使用率 | % | 90% |
| disk_read_rate | 磁盘读取速率 | bytes/s | - |
| disk_write_rate | 磁盘写入速率 | bytes/s | - |
| network_rx_rate | 网络接收速率 | bytes/s | - |
| network_tx_rate | 网络发送速率 | bytes/s | - |

### 应用性能指标

| 指标名称 | 描述 | 单位 | 阈值建议 |
|---------|------|------|----------|
| request_count | 请求总数 | count | - |
| error_count | 错误总数 | count | - |
| error_rate | 错误率 | % | 5% |
| avg_response_time | 平均响应时间 | ms | 1000ms |
| p95_response_time | 95%响应时间 | ms | 2000ms |
| p99_response_time | 99%响应时间 | ms | 5000ms |
| throughput | 吞吐量 | req/s | - |
| active_connections | 活跃连接数 | count | - |
| cache_hit_rate | 缓存命中率 | % | 80% |
| go_routines | Goroutine数量 | count | 10000 |

### 业务指标

| 指标名称 | 描述 | 单位 | 说明 |
|---------|------|------|------|
| active_users | 活跃用户数 | count | 24小时内活跃用户 |
| online_users | 在线用户数 | count | 当前在线用户 |
| api_call_count | API调用次数 | count | 总调用次数 |
| user_sessions | 用户会话数 | count | 活跃会话数 |
| conversion_rate | 转换率 | % | 业务转换率 |

## 🚨 告警管理

### 告警级别

- **Critical（严重）**：需要立即处理的问题，如系统宕机、严重错误
- **Warning（警告）**：需要关注的问题，如性能下降、资源使用率高
- **Info（信息）**：一般信息，如配置变更、系统事件

### 告警状态

- **Triggered（触发）**：告警已触发，等待处理
- **Acknowledged（已确认）**：告警已被确认，正在处理
- **Resolved（已解决）**：告警条件已恢复正常

### 通知渠道

支持多种通知方式：

1. **邮件通知**
   - SMTP配置
   - 支持多个收件人
   - 模板化邮件内容

2. **Webhook通知**
   - HTTP POST请求
   - 自定义请求格式
   - 重试机制

3. **Slack通知**
   - Slack Bot集成
   - 频道消息
   - 富文本格式

## 🔧 故障排查

### 常见问题

#### 1. 监控服务无法启动

**问题症状**：
- 服务启动失败
- 日志显示配置错误

**解决方案**：
```bash
# 检查配置文件
cat .env | grep PERF_MON

# 检查数据库连接
mysql -h localhost -u username -p database_name

# 查看服务日志
tail -f logs/application.log
```

#### 2. 指标数据收集异常

**问题症状**：
- 指标数据缺失
- 收集器报错

**解决方案**：
```go
// 检查收集器状态
stats := monitoringService.GetMonitoringStats()
fmt.Printf("成功收集: %d, 失败收集: %d\n", 
    stats.SuccessfulCollections, 
    stats.FailedCollections)

// 检查收集器配置
config := &Config.PerformanceMonitoringConfig{}
config.SetDefaults()
if err := config.Validate(); err != nil {
    log.Printf("配置验证失败: %v", err)
}
```

#### 3. 告警不触发

**问题症状**：
- 满足条件但告警未触发
- 告警规则无效

**解决方案**：
```sql
-- 检查告警规则
SELECT * FROM alert_rules WHERE enabled = 1;

-- 检查告警历史
SELECT * FROM alerts ORDER BY created_at DESC LIMIT 10;

-- 检查指标数据
SELECT * FROM performance_metrics 
WHERE metric_name = 'cpu_usage' 
ORDER BY timestamp DESC LIMIT 10;
```

#### 4. 性能影响

**问题症状**：
- 监控导致应用性能下降
- 资源消耗过高

**解决方案**：
```yaml
# 调整监控配置
performance_monitoring:
  base:
    interval: 60s              # 增加监控间隔
    batch_size: 50             # 减少批量大小
    verbose_logging: false     # 关闭详细日志
```

### 性能优化建议

1. **合理设置监控间隔**
   - 生产环境：30-60秒
   - 开发环境：可以更频繁

2. **优化数据存储**
   - 使用批量写入
   - 定期清理历史数据
   - 考虑使用时间序列数据库

3. **告警规则优化**
   - 避免过于敏感的阈值
   - 设置合理的持续时间
   - 使用告警静默期

## 🔮 未来规划

### 短期目标（1个月）

1. **指标扩展**
   - 添加更多系统指标
   - 支持自定义业务指标
   - 集成第三方监控系统

2. **告警增强**
   - 智能告警降噪
   - 告警关联分析
   - 自动告警升级

### 中期目标（3个月）

1. **可视化仪表板**
   - 实时监控图表
   - 自定义仪表板
   - 移动端支持

2. **机器学习集成**
   - 异常检测算法
   - 预测性告警
   - 性能趋势分析

### 长期目标（6个月）

1. **分布式监控**
   - 多节点监控
   - 服务拓扑图
   - 分布式追踪

2. **智能运维**
   - 自动故障诊断
   - 智能容量规划
   - 自动扩缩容建议

## 📚 相关资源

### 文档链接

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM ORM Library](https://gorm.io/)
- [Prometheus Monitoring](https://prometheus.io/)
- [Grafana Visualization](https://grafana.com/)

### 示例项目

- [监控系统示例](https://github.com/example/monitoring-demo)
- [性能测试工具](https://github.com/example/perf-testing)
- [告警配置模板](https://github.com/example/alert-templates)

### 社区资源

- [性能监控最佳实践](https://docs.example.com/monitoring-best-practices)
- [Go应用性能优化](https://docs.example.com/go-performance)
- [分布式系统监控](https://docs.example.com/distributed-monitoring)

---

**注意**：本文档会随着系统功能的完善而持续更新，请关注最新版本。
