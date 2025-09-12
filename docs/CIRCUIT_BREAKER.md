# 熔断器系统文档

## 📋 概述

熔断器系统实现了熔断器模式，用于防止级联故障，提高系统的稳定性和可用性。当外部服务出现问题时，熔断器会自动开启，阻止请求继续发送到故障服务，从而保护系统整体稳定性。

## 🏗️ 系统架构

### 核心组件

```
app/Http/Middleware/CircuitBreakerMiddleware.go  # 熔断器中间件
app/Services/MonitoringIntegrationService.go     # 监控集成服务（包含熔断器）
app/Http/Controllers/CircuitBreakerController.go # 熔断器管理API控制器
app/Http/Routes/circuit_breaker.go              # 熔断器路由
```

### 熔断器状态

1. **关闭状态（Closed）** - 正常状态，请求正常通过
2. **开启状态（Open）** - 熔断器开启，请求被拒绝
3. **半开状态（Half-Open）** - 尝试恢复，允许少量请求通过

## 🚀 快速开始

### 1. 基本使用

```go
// 创建熔断器
circuitBreaker := NewCircuitBreaker("user-service", CircuitBreakerConfig{
    MaxRequests: 10,                    // 半开状态下最大请求数
    Interval:    time.Minute,           // 统计时间窗口
    Timeout:     time.Second * 30,      // 熔断器开启后的超时时间
    Threshold:   5,                     // 失败阈值
    SuccessRate: 0.5,                   // 成功率阈值
})

// 在服务调用中使用
func callExternalService() (interface{}, error) {
    if !circuitBreaker.AllowRequest() {
        return nil, errors.New("熔断器开启，请求被拒绝")
    }
    
    start := time.Now()
    result, err := actualServiceCall()
    
    // 记录结果
    circuitBreaker.RecordResult(err == nil, time.Since(start))
    
    return result, err
}
```

### 2. 中间件使用

```go
// 在路由中使用熔断器中间件
router.Use(Middleware.NewCircuitBreakerMiddleware().Handle())

// 或者为特定路由组使用
apiGroup := router.Group("/api/v1")
apiGroup.Use(Middleware.NewCircuitBreakerMiddleware().Handle())
```

## 📡 API接口

### 熔断器管理接口

#### 获取熔断器状态
```http
GET /api/v1/circuit-breaker/status
Authorization: Bearer <token>
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "circuit_breakers": [
      {
        "name": "user-service",
        "state": "closed",
        "requests": 100,
        "successes": 95,
        "failures": 5,
        "success_rate": 0.95,
        "last_failure": "2024-12-20T10:30:00Z",
        "next_attempt": "2024-12-20T10:35:00Z"
      }
    ]
  }
}
```

#### 获取特定熔断器状态
```http
GET /api/v1/circuit-breaker/status/{service_name}
Authorization: Bearer <token>
```

#### 重置熔断器
```http
POST /api/v1/circuit-breaker/reset
Authorization: Bearer <token>
Content-Type: application/json

{
  "service": "user-service"
}
```

#### 手动开启熔断器
```http
POST /api/v1/circuit-breaker/open
Authorization: Bearer <token>
Content-Type: application/json

{
  "service": "user-service",
  "reason": "手动开启熔断器"
}
```

#### 手动关闭熔断器
```http
POST /api/v1/circuit-breaker/close
Authorization: Bearer <token>
Content-Type: application/json

{
  "service": "user-service",
  "reason": "手动关闭熔断器"
}
```

## ⚙️ 配置说明

### 环境变量配置

```bash
# 熔断器配置
CIRCUIT_BREAKER_ENABLED=true
CIRCUIT_BREAKER_DEFAULT_MAX_REQUESTS=10
CIRCUIT_BREAKER_DEFAULT_INTERVAL=60s
CIRCUIT_BREAKER_DEFAULT_TIMEOUT=30s
CIRCUIT_BREAKER_DEFAULT_THRESHOLD=5
CIRCUIT_BREAKER_DEFAULT_SUCCESS_RATE=0.5
```

### 配置文件示例

```yaml
# config/circuit_breaker.yaml
circuit_breaker:
  enabled: true
  default_config:
    max_requests: 10
    interval: 60s
    timeout: 30s
    threshold: 5
    success_rate: 0.5
  
  services:
    user-service:
      max_requests: 20
      interval: 30s
      timeout: 15s
      threshold: 3
      success_rate: 0.8
    
    payment-service:
      max_requests: 5
      interval: 120s
      timeout: 60s
      threshold: 2
      success_rate: 0.9
```

## 🔧 高级功能

### 1. 自定义熔断器配置

```go
// 为不同服务创建不同的熔断器配置
func createServiceCircuitBreakers() map[string]*CircuitBreaker {
    breakers := make(map[string]*CircuitBreaker)
    
    // 用户服务熔断器（较宽松的配置）
    breakers["user-service"] = NewCircuitBreaker("user-service", CircuitBreakerConfig{
        MaxRequests: 20,
        Interval:    time.Minute,
        Timeout:     time.Second * 15,
        Threshold:   10,
        SuccessRate: 0.7,
    })
    
    // 支付服务熔断器（较严格的配置）
    breakers["payment-service"] = NewCircuitBreaker("payment-service", CircuitBreakerConfig{
        MaxRequests: 5,
        Interval:    time.Minute * 2,
        Timeout:     time.Second * 60,
        Threshold:   2,
        SuccessRate: 0.9,
    })
    
    return breakers
}
```

### 2. 熔断器监控和告警

```go
// 熔断器状态监控
func monitorCircuitBreakers() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            for name, breaker := range circuitBreakers {
                state := breaker.GetState()
                stats := breaker.GetStats()
                
                // 检查是否需要告警
                if state == "open" {
                    sendAlert(fmt.Sprintf("熔断器 %s 已开启", name), stats)
                }
                
                // 记录监控指标
                recordMetrics(name, state, stats)
            }
        }
    }
}

// 发送告警
func sendAlert(message string, stats map[string]interface{}) {
    alert := MonitoringAlert{
        Type:      "circuit_breaker",
        Severity:  "warning",
        Title:     "熔断器告警",
        Message:   message,
        Timestamp: time.Now(),
        Metadata:  stats,
    }
    
    // 发送到告警系统
    alertService.SendAlert(alert)
}
```

### 3. 熔断器恢复策略

```go
// 自定义恢复策略
func createRecoveryStrategy(breaker *CircuitBreaker) {
    // 指数退避恢复
    go func() {
        for {
            if breaker.GetState() == "open" {
                // 等待超时时间
                time.Sleep(breaker.GetTimeout())
                
                // 尝试恢复
                breaker.AttemptReset()
                
                // 如果恢复失败，增加等待时间
                if breaker.GetState() == "open" {
                    time.Sleep(breaker.GetTimeout() * 2)
                }
            }
            time.Sleep(time.Second)
        }
    }()
}
```

## 🛠️ 故障排除

### 常见问题

#### 1. 熔断器频繁开启
**问题：** 熔断器频繁在开启和关闭之间切换
**解决方案：**
- 调整失败阈值和成功率阈值
- 增加统计时间窗口
- 检查外部服务的稳定性

#### 2. 熔断器无法恢复
**问题：** 熔断器开启后无法自动恢复
**解决方案：**
- 检查超时时间配置
- 验证外部服务是否已恢复
- 手动重置熔断器

#### 3. 熔断器状态不一致
**问题：** 多个实例的熔断器状态不一致
**解决方案：**
- 使用共享存储同步状态
- 实现熔断器状态广播
- 添加状态同步机制

### 调试方法

```go
// 启用熔断器调试日志
func enableCircuitBreakerDebug() {
    log.SetLevel(log.DebugLevel)
}

// 监控熔断器状态变化
func monitorStateChanges(breaker *CircuitBreaker) {
    go func() {
        lastState := breaker.GetState()
        for {
            currentState := breaker.GetState()
            if currentState != lastState {
                log.Printf("熔断器状态变化: %s -> %s", lastState, currentState)
                lastState = currentState
            }
            time.Sleep(time.Second)
        }
    }()
}
```

## 📊 性能考虑

### 1. 内存使用
- 熔断器状态存储在内存中
- 定期清理过期的统计数据
- 使用对象池减少GC压力

### 2. 并发性能
- 使用读写锁保证并发安全
- 避免在关键路径上进行复杂计算
- 使用原子操作更新计数器

### 3. 监控开销
- 监控频率不应过高
- 使用异步方式发送告警
- 缓存监控数据减少重复计算

## 🔮 未来规划

### 短期目标（1-3个月）
- [ ] 支持熔断器状态持久化
- [ ] 添加熔断器配置热重载
- [ ] 实现熔断器集群同步
- [ ] 添加熔断器性能监控

### 中期目标（3-6个月）
- [ ] 支持自定义熔断器策略
- [ ] 添加熔断器A/B测试
- [ ] 实现熔断器智能恢复
- [ ] 添加熔断器可视化界面

### 长期目标（6-12个月）
- [ ] 集成机器学习预测
- [ ] 支持熔断器自动调优
- [ ] 实现熔断器策略推荐
- [ ] 添加熔断器效果分析

## 📚 相关资源

- [熔断器模式](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Hystrix熔断器](https://github.com/Netflix/Hystrix)
- [Go熔断器实现](https://github.com/sony/gobreaker)

## 📞 技术支持

如有问题或建议，请通过以下方式联系：
- 项目Issues: GitHub Issues
- 代码审查: Pull Request
- 技术讨论: GitHub Discussions

---

**文档版本**: 1.0.0  
**最后更新**: 2024年12月  
**维护者**: 开发团队
