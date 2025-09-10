# 查询优化系统文档

## 📋 概述

查询优化系统是Cloud Platform API项目的重要组成部分，旨在提供全面的数据库查询性能监控、分析和优化功能。系统通过实时监控慢查询、统计查询性能、生成索引建议和性能报告，帮助开发者优化数据库性能。

## 🏗️ 系统架构

### 核心组件

```
app/Services/QueryOptimizationService.go     # 查询优化服务核心
app/Http/Controllers/QueryOptimizationController.go  # API控制器
app/Http/Routes/query_optimization.go       # 路由配置
docs/QUERY_OPTIMIZATION_SYSTEM.md          # 本文档
```

### 系统功能模块

1. **慢查询检测** - 实时监控和记录执行时间超过阈值的查询
2. **查询统计** - 收集和分析查询执行统计信息
3. **索引建议** - 基于查询模式自动生成索引优化建议
4. **性能监控** - 监控系统响应时间、吞吐量、错误率等指标
5. **报告生成** - 生成详细的优化分析报告

## 🚀 快速开始

### 1. 服务初始化

```go
// 在主应用中初始化查询优化服务
config := &Config.QueryOptimizationConfig{}
config.SetDefaults()

queryOptService := Services.NewQueryOptimizationService(db, config)

// 设置到控制器
controller := Controllers.NewQueryOptimizationController()
controller.SetQueryOptimizationService(queryOptService)

// 注册路由
Routes.RegisterQueryOptimizationRoutes(router, controller)
```

### 2. 基本使用

#### 记录慢查询
```go
// 在数据库操作后记录查询信息
queryOptService.RecordSlowQuery(sql, duration, affectedRows)
```

#### 记录性能指标
```go
// 在请求处理中记录性能指标
queryOptService.RecordPerformanceMetric(responseTime, isError)
```

## 📊 功能详解

### 1. 慢查询检测

#### 功能特性
- **阈值检测**: 可配置的查询时间阈值（默认1秒）
- **多级警告**: WARNING和CRITICAL两个级别
- **堆栈跟踪**: 记录查询调用的代码位置
- **执行计划**: 自动获取SQL执行计划
- **文件日志**: 将慢查询记录到专门的日志文件

#### 配置参数
```go
SlowQueryConfig {
    Threshold             time.Duration  // 慢查询阈值
    Enabled              bool           // 是否启用
    LogFile              string         // 日志文件路径
    MaxRecords           int            // 最大记录数
    RecordExecutionPlan  bool           // 是否记录执行计划
    RecordStackTrace     bool           // 是否记录堆栈跟踪
    NotificationThreshold time.Duration  // 通知阈值
}
```

#### API接口
- `GET /api/v1/query-optimization/slow-queries` - 获取慢查询列表
  - 参数: `limit` (数量限制), `warning_level` (警告级别)

### 2. 查询统计

#### 统计指标
- **执行次数**: 每个查询的总执行次数
- **平均耗时**: 平均执行时间
- **最小/最大耗时**: 执行时间范围
- **涉及表**: 查询涉及的数据库表
- **操作类型**: SELECT、INSERT、UPDATE、DELETE

#### API接口
- `GET /api/v1/query-optimization/query-statistics` - 获取查询统计信息

### 3. 索引建议

#### 智能分析
- **慢查询分析**: 基于慢查询自动生成索引建议
- **频率分析**: 考虑查询执行频率
- **影响评估**: 评估索引对性能的影响程度
- **SQL生成**: 自动生成CREATE INDEX语句

#### 建议状态
- `PENDING` - 待处理
- `APPLIED` - 已应用
- `REJECTED` - 已拒绝

#### API接口
- `GET /api/v1/query-optimization/index-suggestions` - 获取索引建议
- `POST /api/v1/query-optimization/index-suggestions/{id}/apply` - 应用建议
- `POST /api/v1/query-optimization/index-suggestions/{id}/reject` - 拒绝建议

### 4. 性能监控

#### 监控指标
- **请求总数**: 系统处理的总请求数
- **响应时间**: 平均、P95、P99响应时间
- **错误率**: 请求错误率
- **吞吐量**: 每秒处理请求数
- **系统运行时间**: 服务运行时长

#### 阈值检测
```go
PerformanceThresholds {
    AvgResponseTime  time.Duration  // 平均响应时间阈值
    P95ResponseTime  time.Duration  // P95响应时间阈值
    P99ResponseTime  time.Duration  // P99响应时间阈值
    ErrorRate        float64        // 错误率阈值
    Throughput       int            // 吞吐量阈值
}
```

#### API接口
- `GET /api/v1/query-optimization/performance-report` - 获取性能报告

### 5. 报告生成

#### 支持格式
- **JSON**: 结构化数据格式
- **HTML**: 可视化网页报告

#### 报告内容
- 慢查询统计和详情
- 查询性能统计
- 索引建议列表
- 系统性能指标
- 优化建议摘要

#### API接口
- `POST /api/v1/query-optimization/generate-report` - 生成优化报告
- `GET /api/v1/query-optimization/summary` - 获取优化摘要

## ⚙️ 配置说明

### 环境变量配置

```bash
# 慢查询配置
QUERY_OPT_SLOW_ENABLED=true
QUERY_OPT_SLOW_THRESHOLD=1000ms
QUERY_OPT_SLOW_LOG_FILE=logs/slow_query.log
QUERY_OPT_SLOW_MAX_RECORDS=1000
QUERY_OPT_SLOW_RECORD_PLAN=true
QUERY_OPT_SLOW_RECORD_STACK=false
QUERY_OPT_SLOW_NOTIFY_THRESHOLD=5000ms

# 索引优化配置
QUERY_OPT_INDEX_ENABLED=true
QUERY_OPT_INDEX_DEPTH=10
QUERY_OPT_INDEX_MIN_FREQ=5
QUERY_OPT_INDEX_MAX_SUGGESTIONS=50
QUERY_OPT_INDEX_AUTO_GENERATE=true
QUERY_OPT_INDEX_SUGGESTIONS_PATH=data/index_suggestions

# 性能监控配置
QUERY_OPT_PERF_ENABLED=true
QUERY_OPT_PERF_INTERVAL=1m
QUERY_OPT_PERF_RETENTION=720h
QUERY_OPT_PERF_DETAILED=true

# 性能阈值配置
QUERY_OPT_PERF_THRESHOLD_AVG=500ms
QUERY_OPT_PERF_THRESHOLD_P95=1000ms
QUERY_OPT_PERF_THRESHOLD_P99=2000ms
QUERY_OPT_PERF_THRESHOLD_ERROR_RATE=0.01
QUERY_OPT_PERF_THRESHOLD_THROUGHPUT=1000

# 报告生成配置
QUERY_OPT_REPORT_ENABLED=true
QUERY_OPT_REPORT_FREQUENCY=24h
QUERY_OPT_REPORT_OUTPUT_PATH=reports/query_optimization
QUERY_OPT_REPORT_FORMAT=html
QUERY_OPT_REPORT_INCLUDE_CHARTS=true
QUERY_OPT_REPORT_TEMPLATE_PATH=templates/reports
```

## 🔧 使用示例

### 1. 获取慢查询列表

```bash
# 获取最近50个慢查询
curl -X GET "http://localhost:8080/api/v1/query-optimization/slow-queries?limit=50" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 只获取严重级别的慢查询
curl -X GET "http://localhost:8080/api/v1/query-optimization/slow-queries?warning_level=CRITICAL" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 2. 获取查询统计

```bash
curl -X GET "http://localhost:8080/api/v1/query-optimization/query-statistics" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. 应用索引建议

```bash
curl -X POST "http://localhost:8080/api/v1/query-optimization/index-suggestions/idx_123/apply" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. 生成优化报告

```bash
curl -X POST "http://localhost:8080/api/v1/query-optimization/generate-report" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. 获取系统摘要

```bash
curl -X GET "http://localhost:8080/api/v1/query-optimization/summary" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 📈 性能优化建议

### 1. 慢查询优化
- 定期检查慢查询日志
- 为经常查询的字段添加索引
- 优化复杂的JOIN查询
- 避免在WHERE子句中使用函数

### 2. 索引优化
- 根据系统建议创建合适的索引
- 定期清理无用的索引
- 注意复合索引的字段顺序
- 监控索引的使用情况

### 3. 查询优化
- 使用EXPLAIN分析查询执行计划
- 避免SELECT *，只查询需要的字段
- 合理使用分页查询
- 优化子查询为JOIN

### 4. 监控告警
- 设置合理的性能阈值
- 建立性能监控告警机制
- 定期生成和分析优化报告
- 关注系统健康状态

## 🔍 故障排查

### 常见问题

#### 1. 慢查询检测不工作
**可能原因**:
- 慢查询检测未启用
- 阈值设置过高
- 数据库连接问题

**解决方案**:
```bash
# 检查配置
echo $QUERY_OPT_SLOW_ENABLED
echo $QUERY_OPT_SLOW_THRESHOLD

# 检查日志文件权限
ls -la logs/slow_query.log
```

#### 2. 索引建议不准确
**可能原因**:
- 查询样本数量不足
- SQL解析逻辑需要优化
- 业务场景特殊

**解决方案**:
- 增加查询样本收集时间
- 手动分析查询模式
- 根据业务需求调整建议算法

#### 3. 性能监控数据异常
**可能原因**:
- 监控间隔设置不当
- 数据retention时间过短
- 系统资源不足

**解决方案**:
```bash
# 检查监控配置
echo $QUERY_OPT_PERF_INTERVAL
echo $QUERY_OPT_PERF_RETENTION

# 检查系统资源
top
free -h
df -h
```

#### 4. 报告生成失败
**可能原因**:
- 输出目录权限不足
- 磁盘空间不足
- 模板文件缺失

**解决方案**:
```bash
# 检查目录权限
ls -ld reports/query_optimization

# 检查磁盘空间
df -h

# 创建必要目录
mkdir -p reports/query_optimization
chmod 755 reports/query_optimization
```

## 📝 最佳实践

### 1. 监控配置
- 根据系统规模合理设置阈值
- 启用详细的监控选项
- 定期备份监控数据

### 2. 索引管理
- 谨慎应用索引建议
- 在测试环境先验证效果
- 监控索引对写入性能的影响

### 3. 报告分析
- 定期生成和分析优化报告
- 建立性能基线
- 跟踪优化效果

### 4. 团队协作
- 建立代码审查机制
- 培训团队成员使用优化工具
- 制定数据库性能规范

## 🔮 未来规划

### 短期目标
- 增强SQL解析能力
- 优化索引建议算法
- 添加更多数据库类型支持

### 中期目标
- 集成机器学习预测
- 添加自动化优化功能
- 开发可视化仪表板

### 长期目标
- 支持分布式数据库
- 智能查询重写
- 全自动性能调优

## 📚 相关资源

- [数据库性能优化指南](https://dev.mysql.com/doc/refman/8.0/en/optimization.html)
- [SQL查询优化技巧](https://use-the-index-luke.com/)
- [数据库索引设计原则](https://www.postgresql.org/docs/current/indexes.html)
- [性能监控最佳实践](https://prometheus.io/docs/practices/)
