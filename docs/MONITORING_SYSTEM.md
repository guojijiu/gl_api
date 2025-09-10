# 监控告警系统文档

## 📋 概述

监控告警系统是Cloud Platform API项目的核心组件，提供全面的系统监控、性能分析、告警管理和通知功能。系统通过实时收集和分析各种监控指标，及时发现系统异常并自动触发告警通知，确保系统的稳定性和可靠性。

## 🏗️ 系统架构

### 核心组件

```
app/Config/monitoring.go                    # 监控告警配置管理
app/Models/Monitoring.go                    # 监控数据模型
app/Services/MonitoringService.go          # 监控告警核心服务
app/Http/Controllers/MonitoringController.go # API控制器
app/Http/Routes/monitoring.go              # 路由配置
docs/MONITORING_SYSTEM.md                 # 本文档
```

### 系统功能模块

1. **系统监控** - 实时监控CPU、内存、磁盘、网络等系统资源
2. **应用监控** - 监控应用性能、内存使用、Goroutine数量等
3. **数据库监控** - 监控数据库连接、慢查询、表大小等
4. **缓存监控** - 监控缓存命中率、内存使用等
5. **业务监控** - 监控用户活动、API调用、错误日志等
6. **告警管理** - 告警规则配置、告警触发、告警处理
7. **通知系统** - 多渠道通知（邮件、Webhook、Slack、钉钉、短信）
8. **数据管理** - 监控数据存储、清理、统计

## 🚀 快速开始

### 1. 服务初始化

```go
// 在主应用中初始化监控告警服务
config := &Config.MonitoringConfig{}
config.SetDefaults()
config.BindEnvs()

if err := config.Validate(); err != nil {
    log.Fatalf("监控配置验证失败: %v", err)
}

monitoringService := Services.NewMonitoringService(db, config)

// 设置到控制器
controller := Controllers.NewMonitoringController()
controller.SetMonitoringService(monitoringService)

// 注册路由
Routes.RegisterMonitoringRoutes(router, controller)
```

### 2. 基础配置

```yaml
# 监控告警基础配置
monitoring:
  base:
    enabled: true
    check_interval: 30s
    retention_period: 720h  # 30天
    max_alerts_per_hour: 100
    alert_cooldown: 5m
    enable_dashboard: true
    dashboard_port: 8081
    enable_metrics: true
    metrics_port: 8082
```

## 🔧 功能特性

### 1. 多层次监控

#### 系统监控
- **CPU使用率监控**: 实时监控CPU使用率，支持多核CPU
- **内存使用监控**: 监控物理内存和虚拟内存使用情况
- **磁盘使用监控**: 监控磁盘空间使用率和IO性能
- **网络流量监控**: 监控网络发送和接收流量
- **进程数量监控**: 监控系统进程总数
- **系统负载监控**: 监控系统1分钟、5分钟、15分钟负载

#### 应用监控
- **内存使用监控**: 监控应用内存分配和使用情况
- **Goroutine监控**: 监控Go协程数量
- **GC监控**: 监控垃圾回收性能和频率
- **内存泄漏检测**: 检测应用内存泄漏趋势

#### 数据库监控
- **连接数监控**: 监控数据库连接池使用情况
- **慢查询监控**: 检测和记录慢查询
- **表大小监控**: 监控数据库表大小增长
- **锁等待监控**: 监控数据库锁等待时间

#### 缓存监控
- **命中率监控**: 监控缓存命中率
- **内存使用监控**: 监控缓存内存使用情况
- **连接数监控**: 监控缓存连接数
- **过期键监控**: 监控过期键数量

#### 业务监控
- **用户活跃度**: 监控用户登录和活动情况
- **API调用统计**: 监控API调用次数和性能
- **错误日志监控**: 监控错误日志数量
- **安全事件监控**: 监控安全相关事件

### 2. 智能告警

#### 告警规则
- **阈值告警**: 基于固定阈值的告警
- **趋势告警**: 基于指标变化趋势的告警
- **异常告警**: 基于异常检测算法的告警

#### 告警特性
- **多级告警**: 支持info、warning、critical、emergency级别
- **告警抑制**: 防止告警风暴，支持时间窗口抑制
- **告警升级**: 支持告警自动升级机制
- **自动解决**: 支持告警自动解决功能

#### 告警处理
- **告警确认**: 支持告警确认操作
- **告警解决**: 支持告警解决操作
- **告警历史**: 完整的告警历史记录
- **告警统计**: 告警统计和分析

### 3. 多渠道通知

#### 通知渠道
- **邮件通知**: 支持SMTP邮件发送
- **Webhook通知**: 支持HTTP Webhook回调
- **Slack通知**: 支持Slack消息推送
- **钉钉通知**: 支持钉钉群消息推送
- **短信通知**: 支持短信发送（需配置服务商）

#### 通知特性
- **重试机制**: 失败自动重试
- **通知记录**: 完整的通知发送记录
- **通知状态**: 实时通知状态跟踪
- **通知模板**: 支持自定义通知模板

### 4. 数据管理

#### 数据存储
- **数据库存储**: 主要数据存储在关系型数据库
- **文件存储**: 支持日志文件存储
- **Redis缓存**: 支持Redis缓存存储

#### 数据清理
- **自动清理**: 定期清理过期数据
- **数据保留**: 可配置的数据保留策略
- **数据压缩**: 支持数据压缩存储

## 📡 API接口

### 监控指标接口

#### 获取监控指标
```http
GET /api/v1/monitoring/metrics
```

**查询参数:**
- `type`: 指标类型 (system, application, database, cache, business)
- `name`: 指标名称
- `limit`: 限制返回数量 (默认100)
- `start_time`: 开始时间 (ISO 8601格式)
- `end_time`: 结束时间 (ISO 8601格式)

**响应示例:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "metrics": [
      {
        "id": 1,
        "type": "system",
        "name": "cpu_usage",
        "value": 75.5,
        "unit": "%",
        "threshold": 80.0,
        "status": "normal",
        "severity": "info",
        "timestamp": "2024-12-01T10:00:00Z"
      }
    ],
    "total": 1,
    "type": "system",
    "name": "cpu_usage",
    "limit": 100
  }
}
```

### 告警管理接口

#### 获取告警列表
```http
GET /api/v1/monitoring/alerts
```

**查询参数:**
- `status`: 告警状态 (active, acknowledged, resolved, suppressed)
- `severity`: 严重程度 (info, warning, critical, emergency)
- `limit`: 限制返回数量 (默认50)
- `page`: 页码 (默认1)
- `page_size`: 每页数量 (默认20)

#### 确认告警
```http
POST /api/v1/monitoring/alerts/{id}/acknowledge
```

#### 解决告警
```http
POST /api/v1/monitoring/alerts/{id}/resolve
```

### 告警规则接口

#### 获取告警规则
```http
GET /api/v1/monitoring/alert-rules
```

#### 创建告警规则
```http
POST /api/v1/monitoring/alert-rules
```

**请求体:**
```json
{
  "name": "CPU告警",
  "description": "CPU使用率超过80%时告警",
  "type": "threshold",
  "metric_type": "system",
  "metric_name": "cpu_usage",
  "condition": ">",
  "threshold": 80.0,
  "duration": 1,
  "severity": "warning",
  "enabled": true,
  "suppression": true,
  "suppression_window": 3600,
  "escalation": true,
  "escalation_delay": 600,
  "max_escalation_level": 3,
  "notification_channels": "[\"email\", \"slack\"]"
}
```

#### 更新告警规则
```http
PUT /api/v1/monitoring/alert-rules/{id}
```

#### 删除告警规则
```http
DELETE /api/v1/monitoring/alert-rules/{id}
```

### 系统健康接口

#### 获取系统健康状态
```http
GET /api/v1/monitoring/health
```

**响应示例:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "health": {
      "status": "healthy",
      "timestamp": "2024-12-01T10:00:00Z",
      "metrics": {
        "latest": [...]
      },
      "alerts": {
        "active_count": 0,
        "critical_count": 0
      }
    }
  }
}
```

### 通知记录接口

#### 获取通知记录
```http
GET /api/v1/monitoring/notifications
```

### 统计信息接口

#### 获取监控统计
```http
GET /api/v1/monitoring/stats
```

## ⚙️ 配置说明

### 环境变量配置

#### 基础配置
```bash
# 监控告警基础配置
MONITORING_ENABLED=true                    # 是否启用监控告警
MONITORING_CHECK_INTERVAL=30s              # 检查间隔
MONITORING_RETENTION_PERIOD=720h           # 数据保留时间(30天)
MONITORING_MAX_ALERTS_PER_HOUR=100         # 每小时最大告警数
MONITORING_ALERT_COOLDOWN=5m               # 告警冷却时间
MONITORING_ENABLE_DASHBOARD=true           # 是否启用仪表板
MONITORING_DASHBOARD_PORT=8081             # 仪表板端口
MONITORING_ENABLE_METRICS=true             # 是否启用指标收集
MONITORING_METRICS_PORT=8082               # 指标端口
```

#### 系统监控配置
```bash
# 系统监控配置
MONITORING_SYSTEM_ENABLED=true             # 是否启用系统监控
MONITORING_SYSTEM_CHECK_INTERVAL=60s        # 检查间隔
MONITORING_SYSTEM_CPU_THRESHOLD=80.0        # CPU阈值
MONITORING_SYSTEM_MEMORY_THRESHOLD=85.0     # 内存阈值
MONITORING_SYSTEM_DISK_THRESHOLD=90.0       # 磁盘阈值
MONITORING_SYSTEM_NETWORK_THRESHOLD=1000.0  # 网络阈值(MB/s)
MONITORING_SYSTEM_PROCESS_THRESHOLD=1000     # 进程阈值
MONITORING_SYSTEM_LOAD_AVERAGE_THRESHOLD=5.0 # 负载阈值
```

#### 应用监控配置
```bash
# 应用监控配置
MONITORING_APP_ENABLED=true                # 是否启用应用监控
MONITORING_APP_CHECK_INTERVAL=30s          # 检查间隔
MONITORING_APP_RESPONSE_TIME_THRESHOLD=2s   # 响应时间阈值
MONITORING_APP_ERROR_RATE_THRESHOLD=5.0     # 错误率阈值
MONITORING_APP_THROUGHPUT_THRESHOLD=1000    # 吞吐量阈值
MONITORING_APP_MEMORY_LEAK_THRESHOLD=10.0   # 内存泄漏阈值
MONITORING_APP_GOROUTINE_THRESHOLD=10000    # Goroutine阈值
MONITORING_APP_GC_THRESHOLD=100ms           # GC阈值
```

#### 数据库监控配置
```bash
# 数据库监控配置
MONITORING_DB_ENABLED=true                 # 是否启用数据库监控
MONITORING_DB_CHECK_INTERVAL=60s           # 检查间隔
MONITORING_DB_CONNECTION_THRESHOLD=100     # 连接数阈值
MONITORING_DB_SLOW_QUERY_THRESHOLD=1s      # 慢查询阈值
MONITORING_DB_QUERY_TIMEOUT_THRESHOLD=30s  # 查询超时阈值
MONITORING_DB_DEADLOCK_THRESHOLD=5         # 死锁阈值
MONITORING_DB_LOCK_WAIT_THRESHOLD=10s      # 锁等待阈值
MONITORING_DB_TABLE_SIZE_THRESHOLD=1073741824 # 表大小阈值(1GB)
```

#### 缓存监控配置
```bash
# 缓存监控配置
MONITORING_CACHE_ENABLED=true              # 是否启用缓存监控
MONITORING_CACHE_CHECK_INTERVAL=60s        # 检查间隔
MONITORING_CACHE_HIT_RATE_THRESHOLD=80.0    # 命中率阈值
MONITORING_CACHE_MEMORY_USAGE_THRESHOLD=85.0 # 内存使用阈值
MONITORING_CACHE_CONNECTION_THRESHOLD=100   # 连接数阈值
MONITORING_CACHE_EVICTION_THRESHOLD=1000   # 驱逐阈值
MONITORING_CACHE_EXPIRED_KEYS_THRESHOLD=10000 # 过期键阈值
```

#### 业务监控配置
```bash
# 业务监控配置
MONITORING_BUSINESS_ENABLED=true           # 是否启用业务监控
MONITORING_BUSINESS_CHECK_INTERVAL=5m      # 检查间隔
MONITORING_BUSINESS_USER_ACTIVITY_THRESHOLD=100 # 用户活动阈值
MONITORING_BUSINESS_API_USAGE_THRESHOLD=1000 # API使用阈值
MONITORING_BUSINESS_ERROR_LOG_THRESHOLD=50 # 错误日志阈值
MONITORING_BUSINESS_SECURITY_EVENT_THRESHOLD=10 # 安全事件阈值
MONITORING_BUSINESS_DATA_SYNC_THRESHOLD=10m # 数据同步阈值
```

#### 告警配置
```bash
# 告警配置
MONITORING_ALERT_ENABLED=true              # 是否启用告警
MONITORING_ALERT_DEFAULT_SEVERITY=warning  # 默认严重程度
MONITORING_ALERT_ESCALATION_ENABLED=true   # 是否启用升级
MONITORING_ALERT_ESCALATION_DELAY=10m      # 升级延迟
MONITORING_ALERT_MAX_ESCALATION_LEVEL=3    # 最大升级级别
MONITORING_ALERT_AUTO_RESOLVE_ENABLED=true # 是否启用自动解决
MONITORING_ALERT_AUTO_RESOLVE_DELAY=30m   # 自动解决延迟
MONITORING_ALERT_SUPPRESSION_ENABLED=true # 是否启用抑制
MONITORING_ALERT_SUPPRESSION_WINDOW=1h     # 抑制窗口
```

#### 通知配置
```bash
# 邮件通知配置
MONITORING_NOTIFICATION_EMAIL_ENABLED=false # 是否启用邮件通知
MONITORING_NOTIFICATION_EMAIL_SMTP_HOST=smtp.example.com
MONITORING_NOTIFICATION_EMAIL_SMTP_PORT=587
MONITORING_NOTIFICATION_EMAIL_USERNAME=user@example.com
MONITORING_NOTIFICATION_EMAIL_PASSWORD=password
MONITORING_NOTIFICATION_EMAIL_FROM_ADDRESS=monitoring@example.com
MONITORING_NOTIFICATION_EMAIL_TO_ADDRESSES=admin@example.com
MONITORING_NOTIFICATION_EMAIL_SUBJECT="[监控告警] Cloud Platform"

# Webhook通知配置
MONITORING_NOTIFICATION_WEBHOOK_ENABLED=false # 是否启用Webhook通知
MONITORING_NOTIFICATION_WEBHOOK_URL=https://api.example.com/webhook
MONITORING_NOTIFICATION_WEBHOOK_METHOD=POST
MONITORING_NOTIFICATION_WEBHOOK_HEADERS=Content-Type:application/json
MONITORING_NOTIFICATION_WEBHOOK_TIMEOUT=10s
MONITORING_NOTIFICATION_WEBHOOK_RETRY_COUNT=3

# Slack通知配置
MONITORING_NOTIFICATION_SLACK_ENABLED=false # 是否启用Slack通知
MONITORING_NOTIFICATION_SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx
MONITORING_NOTIFICATION_SLACK_CHANNEL=#monitoring
MONITORING_NOTIFICATION_SLACK_USERNAME=监控告警
MONITORING_NOTIFICATION_SLACK_ICON_EMOJI=:warning:

# 钉钉通知配置
MONITORING_NOTIFICATION_DINGTALK_ENABLED=false # 是否启用钉钉通知
MONITORING_NOTIFICATION_DINGTALK_WEBHOOK_URL=https://oapi.dingtalk.com/robot/send?access_token=xxx
MONITORING_NOTIFICATION_DINGTALK_SECRET=secret
MONITORING_NOTIFICATION_DINGTALK_AT_MOBILES=13800138000

# 短信通知配置
MONITORING_NOTIFICATION_SMS_ENABLED=false # 是否启用短信通知
MONITORING_NOTIFICATION_SMS_PROVIDER=aliyun
MONITORING_NOTIFICATION_SMS_API_KEY=your_api_key
MONITORING_NOTIFICATION_SMS_API_SECRET=your_api_secret
MONITORING_NOTIFICATION_SMS_PHONE_NUMBERS=13800138000
```

#### 存储配置
```bash
# 存储配置
MONITORING_STORAGE_TYPE=database           # 存储类型
MONITORING_STORAGE_DATABASE_ENABLED=true  # 是否启用数据库存储
MONITORING_STORAGE_DATABASE_TABLE_PREFIX=monitoring_ # 表前缀
MONITORING_STORAGE_DATABASE_RETENTION=2160h # 数据保留时间(90天)
MONITORING_STORAGE_FILE_ENABLED=false     # 是否启用文件存储
MONITORING_STORAGE_FILE_PATH=logs/monitoring # 文件路径
MONITORING_STORAGE_FILE_FORMAT=json       # 文件格式
MONITORING_STORAGE_FILE_MAX_SIZE=104857600 # 最大文件大小(100MB)
MONITORING_STORAGE_FILE_MAX_AGE=720h      # 最大文件年龄(30天)
MONITORING_STORAGE_REDIS_ENABLED=false    # 是否启用Redis存储
MONITORING_STORAGE_REDIS_KEY_PREFIX=monitoring: # 键前缀
MONITORING_STORAGE_REDIS_TTL=24h          # TTL时间
```

## 📊 使用示例

### 1. 创建CPU告警规则

```bash
curl -X POST http://localhost:8080/api/v1/monitoring/alert-rules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "CPU告警",
    "description": "CPU使用率超过80%时告警",
    "type": "threshold",
    "metric_type": "system",
    "metric_name": "cpu_usage",
    "condition": ">",
    "threshold": 80.0,
    "duration": 1,
    "severity": "warning",
    "enabled": true,
    "suppression": true,
    "suppression_window": 3600,
    "escalation": true,
    "escalation_delay": 600,
    "max_escalation_level": 3,
    "notification_channels": "[\"email\", \"slack\"]"
  }'
```

### 2. 查看系统健康状态

```bash
curl -X GET http://localhost:8080/api/v1/monitoring/health \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. 获取监控指标

```bash
curl -X GET "http://localhost:8080/api/v1/monitoring/metrics?type=system&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. 确认告警

```bash
curl -X POST http://localhost:8080/api/v1/monitoring/alerts/1/acknowledge \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🔧 故障排查

### 常见问题

#### 1. 监控服务无法启动
**问题**: 监控服务启动失败
**解决方案**:
- 检查配置文件是否正确
- 检查数据库连接是否正常
- 检查依赖包是否正确安装

#### 2. 监控指标收集失败
**问题**: 无法收集系统监控指标
**解决方案**:
- 检查gopsutil包是否正确安装
- 检查系统权限是否足够
- 检查监控配置是否正确

#### 3. 告警无法触发
**问题**: 告警规则配置正确但无法触发
**解决方案**:
- 检查告警规则是否启用
- 检查指标值是否达到阈值
- 检查告警抑制配置

#### 4. 通知发送失败
**问题**: 告警通知无法发送
**解决方案**:
- 检查通知渠道配置
- 检查网络连接
- 检查认证信息

### 日志分析

监控告警系统会记录详细的日志信息，可以通过以下方式查看：

```bash
# 查看应用日志
tail -f logs/app.log | grep monitoring

# 查看错误日志
tail -f logs/error.log | grep monitoring
```

### 性能优化

#### 1. 数据存储优化
- 定期清理过期数据
- 使用数据库索引优化查询
- 考虑使用时间序列数据库

#### 2. 监控频率优化
- 根据业务需求调整监控频率
- 避免过于频繁的监控检查
- 使用缓存减少重复计算

#### 3. 告警优化
- 合理设置告警阈值
- 避免告警风暴
- 使用告警抑制机制

## 🔮 未来规划

### 短期目标（1个月）
- [ ] 添加更多监控指标
- [ ] 优化告警算法
- [ ] 增加更多通知渠道
- [ ] 完善监控仪表板

### 中期目标（3个月）
- [ ] 实现机器学习异常检测
- [ ] 添加监控数据可视化
- [ ] 支持监控数据导出
- [ ] 实现监控API开放

### 长期目标（6个月）
- [ ] 支持分布式监控
- [ ] 实现监控数据备份
- [ ] 添加监控报告生成
- [ ] 支持监控插件系统

## 📚 相关资源

### 文档链接
- [gopsutil文档](https://github.com/shirou/gopsutil)
- [Gin框架文档](https://gin-gonic.com/docs/)
- [GORM文档](https://gorm.io/docs/)

### 最佳实践
- [监控告警最佳实践](https://prometheus.io/docs/practices/)
- [系统监控指南](https://www.datadoghq.com/blog/monitoring-101-collecting-data/)
- [告警管理指南](https://www.pagerduty.com/resources/learn/incident-response/)

### 示例代码
- [监控指标收集示例](examples/metrics_collection.go)
- [告警规则配置示例](examples/alert_rules.go)
- [通知发送示例](examples/notifications.go)
