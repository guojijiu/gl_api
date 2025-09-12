# 配置热重载系统文档

## 📋 概述

配置热重载系统允许应用程序在运行时动态重新加载配置文件，无需重启服务。这对于生产环境中的配置更新、调试和运维非常有用。

## 🏗️ 系统架构

### 核心组件

```
app/Config/hot_reload.go          # 配置热重载管理器
app/Config/config.go              # 配置加载和管理
app/Http/Controllers/ConfigController.go  # 配置管理API控制器
app/Http/Routes/config.go         # 配置管理路由
```

### 功能特性

1. **文件监控** - 使用fsnotify监控配置文件变化
2. **回调机制** - 支持多个重载回调函数
3. **错误处理** - 完善的错误处理和恢复机制
4. **并发安全** - 使用读写锁保证并发安全
5. **优雅关闭** - 支持优雅停止监控

## 🚀 快速开始

### 1. 基本使用

```go
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
```

### 2. 在服务中使用

```go
// 在服务初始化时设置热重载
func initService() {
    // 创建热重载管理器
    hotReloadManager := Config.NewHotReloadManager("config/app.yaml")
    
    // 添加配置重载回调
    hotReloadManager.AddReloadCallback(func(config *Config.Config) {
        // 更新数据库配置
        updateDatabaseConfig(config.Database)
        
        // 更新Redis配置
        updateRedisConfig(config.Redis)
        
        // 更新日志配置
        updateLoggingConfig(config.Logging)
        
        log.Info("服务配置已更新")
    })
    
    // 开始监控
    if err := hotReloadManager.StartWatching(); err != nil {
        log.Fatal("配置热重载启动失败:", err)
    }
}
```

## 📡 API接口

### 配置管理接口

#### 获取当前配置
```http
GET /api/v1/config
Authorization: Bearer <token>
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "database": {
      "host": "localhost",
      "port": 3306,
      "username": "root",
      "password": "password"
    },
    "redis": {
      "host": "localhost",
      "port": 6379
    },
    "logging": {
      "level": "info",
      "format": "json"
    }
  }
}
```

#### 更新配置
```http
PUT /api/v1/config
Authorization: Bearer <token>
Content-Type: application/json

{
  "key": "database.host",
  "value": "new-db-host"
}
```

#### 重载配置
```http
POST /api/v1/config/reload
Authorization: Bearer <token>
```

**响应示例：**
```json
{
  "success": true,
  "message": "配置重载成功"
}
```

#### 获取配置状态
```http
GET /api/v1/config/status
Authorization: Bearer <token>
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "is_watching": true,
    "config_file": "config/app.yaml",
    "last_reload": "2024-12-20T10:30:00Z",
    "reload_count": 5
  }
}
```

## ⚙️ 配置说明

### 环境变量配置

```bash
# 配置热重载
CONFIG_HOT_RELOAD_ENABLED=true
CONFIG_FILE_PATH=config/app.yaml
CONFIG_WATCH_INTERVAL=1s
CONFIG_RELOAD_TIMEOUT=30s
```

### 配置文件示例

```yaml
# config/app.yaml
app:
  name: "Cloud Platform API"
  version: "1.3.0"
  port: 8080
  mode: "release"

database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  database: "cloud_platform"
  max_open_conns: 100
  max_idle_conns: 10

redis:
  host: "localhost"
  port: 6379
  password: ""
  database: 0
  pool_size: 10

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file_path: "logs/app.log"
```

## 🔧 高级功能

### 1. 自定义重载回调

```go
// 创建自定义重载回调
func createCustomReloadCallback(serviceName string) func(*Config.Config) {
    return func(config *Config.Config) {
        log.Printf("服务 %s 配置已重载", serviceName)
        
        // 自定义配置更新逻辑
        switch serviceName {
        case "database":
            updateDatabaseConnection(config.Database)
        case "redis":
            updateRedisConnection(config.Redis)
        case "logging":
            updateLoggingConfig(config.Logging)
        }
    }
}

// 使用自定义回调
hotReloadManager.AddReloadCallback(createCustomReloadCallback("database"))
```

### 2. 配置验证

```go
// 配置验证回调
func validateConfig(config *Config.Config) error {
    // 验证数据库配置
    if config.Database.Host == "" {
        return errors.New("数据库主机不能为空")
    }
    
    // 验证Redis配置
    if config.Redis.Host == "" {
        return errors.New("Redis主机不能为空")
    }
    
    return nil
}

// 添加验证回调
hotReloadManager.AddReloadCallback(func(config *Config.Config) {
    if err := validateConfig(config); err != nil {
        log.Error("配置验证失败:", err)
        return
    }
    
    // 配置验证通过，执行重载
    log.Info("配置重载成功")
})
```

### 3. 配置备份

```go
// 配置备份功能
func backupConfig(config *Config.Config) error {
    backupPath := fmt.Sprintf("config/backup/config_%s.yaml", time.Now().Format("20060102_150405"))
    
    // 创建备份目录
    if err := os.MkdirAll("config/backup", 0755); err != nil {
        return err
    }
    
    // 备份配置文件
    data, err := yaml.Marshal(config)
    if err != nil {
        return err
    }
    
    return os.WriteFile(backupPath, data, 0644)
}

// 在重载前备份配置
hotReloadManager.AddReloadCallback(func(config *Config.Config) {
    if err := backupConfig(config); err != nil {
        log.Error("配置备份失败:", err)
    }
    
    // 继续重载逻辑
    log.Info("配置重载完成")
})
```

## 🛠️ 故障排除

### 常见问题

#### 1. 配置文件监控失败
**问题：** 无法监控配置文件变化
**解决方案：**
- 检查配置文件路径是否正确
- 确认文件权限是否足够
- 检查文件系统是否支持inotify

#### 2. 配置重载失败
**问题：** 配置重载时出现错误
**解决方案：**
- 检查配置文件格式是否正确
- 验证配置值是否有效
- 查看错误日志获取详细信息

#### 3. 回调函数执行失败
**问题：** 重载回调函数执行时出错
**解决方案：**
- 在回调函数中添加错误处理
- 使用goroutine异步执行回调
- 添加重试机制

### 调试方法

```go
// 启用详细日志
func enableDebugLogging() {
    log.SetLevel(log.DebugLevel)
}

// 监控文件变化事件
func monitorFileEvents() {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()
    
    done := make(chan bool)
    go func() {
        for {
            select {
            case event := <-watcher.Events:
                log.Printf("文件事件: %s", event)
            case err := <-watcher.Errors:
                log.Printf("监控错误: %s", err)
            }
        }
    }()
    
    err = watcher.Add("config/app.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    <-done
}
```

## 📊 性能考虑

### 1. 监控频率
- 默认监控间隔：1秒
- 可根据需要调整监控频率
- 避免过于频繁的监控影响性能

### 2. 回调函数优化
- 回调函数应尽量轻量
- 避免在回调中执行耗时操作
- 使用异步处理复杂逻辑

### 3. 内存使用
- 定期清理过期的配置备份
- 避免在回调中创建大量对象
- 使用对象池减少GC压力

## 🔮 未来规划

### 短期目标（1-3个月）
- [ ] 支持多配置文件监控
- [ ] 添加配置变更历史记录
- [ ] 实现配置回滚功能
- [ ] 添加配置模板支持

### 中期目标（3-6个月）
- [ ] 支持远程配置中心集成
- [ ] 添加配置加密支持
- [ ] 实现配置版本管理
- [ ] 添加配置审计功能

### 长期目标（6-12个月）
- [ ] 支持配置热重载的集群同步
- [ ] 添加配置变更的实时通知
- [ ] 实现配置的A/B测试
- [ ] 添加配置性能监控

## 📚 相关资源

- [fsnotify文档](https://github.com/fsnotify/fsnotify)
- [Viper配置管理](https://github.com/spf13/viper)
- [Go文件监控最佳实践](https://golang.org/pkg/os/signal/)

## 📞 技术支持

如有问题或建议，请通过以下方式联系：
- 项目Issues: GitHub Issues
- 代码审查: Pull Request
- 技术讨论: GitHub Discussions

---

**文档版本**: 1.0.0  
**最后更新**: 2024年12月  
**维护者**: 开发团队
