# 配置模块说明

## 概述

配置模块已按功能拆分为多个独立的文件，每个配置类型都是完全自包含的模块，包含自己的默认值设置、环境变量绑定和验证逻辑。这种设计提供了更好的代码组织和维护性，同时保持了向后兼容性。

## 🚀 快速开始

### 基本使用
```go
package main

import (
    "log"
    "your-project/app/Config"
)

func main() {
    // 1. 加载配置
    Config.LoadConfig()
    
    // 2. 验证配置
    if err := Config.ValidateConfig(); err != nil {
        log.Fatal("配置验证失败:", err)
    }
    
    // 3. 获取配置
    config := Config.GetConfig()
    serverConfig := Config.GetServerConfig()
    dbConfig := Config.GetDatabaseConfig()
    
    // 4. 使用配置
    log.Printf("服务器端口: %s", serverConfig.Port)
    log.Printf("数据库驱动: %s", dbConfig.Driver)
}
```

### 高级使用
```go
// 服务器配置
serverConfig := Config.GetServerConfig()
if serverConfig.IsDebugMode() {
    log.Println("调试模式")
}

// 数据库配置
dbConfig := Config.GetDatabaseConfig()
dsn := dbConfig.GetDSN()
if dbConfig.IsSQLite() {
    log.Println("使用SQLite数据库")
}

// JWT配置
jwtConfig := Config.GetJWTConfig()
if jwtConfig.IsSecretDefault() {
    log.Println("警告：使用默认JWT密钥")
}

// 存储配置
storageConfig := Config.GetStorageConfig()
if storageConfig.IsFileTypeAllowed("jpg") {
    log.Println("允许上传jpg文件")
}
```

## 文件结构

```
app/Config/
├── base.go          # 主配置结构和协调器
├── server.go        # 服务器配置 (完全自包含)
├── database.go      # 数据库配置 (完全自包含)
├── jwt.go          # JWT认证配置 (完全自包含)
├── redis.go        # Redis缓存配置 (完全自包含)
├── storage.go      # 文件存储配置 (完全自包含)
├── config.go       # 向后兼容性文件
├── example.go      # 使用示例
└── README.md       # 本文档
```

## 架构设计

### 模块化设计原则

每个配置模块都是完全自包含的，包含以下三个核心方法：

1. **SetDefaults()**: 设置该配置类型的默认值
2. **BindEnvs()**: 绑定该配置类型的环境变量
3. **Validate()**: 验证该配置类型的有效性

这种设计确保了：
- **高内聚**: 每个配置模块的所有相关功能都在同一个文件中
- **低耦合**: 各个配置模块之间相互独立
- **易维护**: 修改某个配置类型时只需要修改对应的文件
- **易扩展**: 添加新的配置类型时只需要创建新的文件并实现三个核心方法

### 协调器模式

`base.go` 作为协调器，负责：
- 定义主配置结构
- 协调各个配置模块的加载
- 提供全局的配置访问接口
- 执行整体配置验证

## 配置类型

### 1. 服务器配置 (server.go)

**配置项：**
- `port`: 服务器端口
- `mode`: 运行模式 (debug/production)
- `base_url`: 基础URL

**环境变量：**
- `SERVER_PORT`
- `SERVER_MODE`
- `SERVER_BASE_URL`

**主要方法：**
- `SetDefaults()`: 设置默认值
- `BindEnvs()`: 绑定环境变量
- `Validate()`: 验证配置
- `IsDebugMode()`: 检查是否为调试模式
- `IsProductionMode()`: 检查是否为生产模式
- `GetFullURL(path)`: 获取完整URL

### 2. 数据库配置 (database.go)

**配置项：**
- `driver`: 数据库驱动 (sqlite/mysql/postgres)
- `host`: 数据库主机
- `port`: 数据库端口
- `username`: 用户名
- `password`: 密码
- `database`: 数据库名
- `charset`: 字符集

**环境变量：**
- `DB_DRIVER`
- `DB_HOST`
- `DB_PORT`
- `DB_USERNAME`
- `DB_PASSWORD`
- `DB_DATABASE`
- `DB_CHARSET`

**主要方法：**
- `SetDefaults()`: 设置默认值
- `BindEnvs()`: 绑定环境变量
- `Validate()`: 验证配置
- `GetDSN()`: 获取数据库连接字符串
- `IsSQLite()`: 检查是否为SQLite
- `IsMySQL()`: 检查是否为MySQL
- `IsPostgreSQL()`: 检查是否为PostgreSQL

### 3. JWT配置 (jwt.go)

**配置项：**
- `secret`: JWT密钥
- `expire_time`: 过期时间（小时）

**环境变量：**
- `JWT_SECRET`
- `JWT_EXPIRE_TIME`

**主要方法：**
- `SetDefaults()`: 设置默认值
- `BindEnvs()`: 绑定环境变量
- `Validate()`: 验证配置
- `GetExpireDuration()`: 获取过期时间间隔
- `GetExpireTime()`: 获取过期时间（秒）
- `IsSecretDefault()`: 检查是否为默认密钥

### 4. Redis配置 (redis.go)

**配置项：**
- `host`: Redis主机
- `port`: Redis端口
- `password`: Redis密码
- `database`: Redis数据库编号

**环境变量：**
- `REDIS_HOST`
- `REDIS_PORT`
- `REDIS_PASSWORD`
- `REDIS_DATABASE`

**主要方法：**
- `SetDefaults()`: 设置默认值
- `BindEnvs()`: 绑定环境变量
- `Validate()`: 验证配置
- `GetAddr()`: 获取Redis地址
- `GetConnectionString()`: 获取连接字符串
- `IsPasswordSet()`: 检查是否设置密码

### 5. 存储配置 (storage.go)

**配置项：**
- `upload_path`: 上传路径
- `max_file_size`: 最大文件大小（MB）
- `allowed_types`: 允许的文件类型
- `private_path`: 私有文件路径
- `public_path`: 公共文件路径
- `temp_path`: 临时文件路径
- `log_path`: 日志文件路径
- `cache_path`: 缓存文件路径

**环境变量：**
- `STORAGE_UPLOAD_PATH`
- `STORAGE_MAX_FILE_SIZE`
- `STORAGE_ALLOWED_TYPES`
- `STORAGE_PRIVATE_PATH`
- `STORAGE_PUBLIC_PATH`
- `STORAGE_TEMP_PATH`
- `STORAGE_LOG_PATH`
- `STORAGE_CACHE_PATH`

**主要方法：**
- `SetDefaults()`: 设置默认值
- `BindEnvs()`: 绑定环境变量
- `Validate()`: 验证配置

### 6. AI 网关与 ai_robot 上下文配置 (ai_gateway.go)

**AI 网关基础环境变量：**
- `CLOUD_PLATFORM_BASE_URL`
- `CLOUD_PLATFORM_INTRANET_BASE_URL`
- `AI_GATEWAY_LLM_BASE_URL`
- `AI_GATEWAY_LLM_API_KEY`
- `AI_GATEWAY_LLM_MODEL`
- `AI_GATEWAY_LLM_TIMEOUT_SEC`
- `AI_GATEWAY_EXPIRING_DAYS`
- `AI_GATEWAY_STRICT_MODE`

**公共 MongoDB 环境变量：**
- `MONGODB_URI`
- `MONGODB_HOST`
- `MONGODB_PORT`
- `MONGODB_DATABASE`
- `MONGODB_USERNAME`
- `MONGODB_PASSWORD`
- `MONGODB_AUTH_SOURCE`
- `MONGODB_TIMEOUT_SEC`

**ai_robot 会话存储环境变量：**
- `AI_ROBOT_CONVERSATION_STORE_DRIVER`
- `AI_ROBOT_CONVERSATION_MONGO_URI`
- `AI_ROBOT_CONVERSATION_MONGO_HOST`
- `AI_ROBOT_CONVERSATION_MONGO_PORT`
- `AI_ROBOT_CONVERSATION_MONGO_DATABASE`
- `AI_ROBOT_CONVERSATION_MONGO_COLLECTION`
- `AI_ROBOT_CONVERSATION_MONGO_USERNAME`
- `AI_ROBOT_CONVERSATION_MONGO_PASSWORD`
- `AI_ROBOT_CONVERSATION_MONGO_AUTH_SOURCE`
- `AI_ROBOT_CONVERSATION_MONGO_TIMEOUT_SEC`
- `AI_ROBOT_CONVERSATION_TTL_HOURS`

**ai_robot 上下文继承调优环境变量：**
- `AI_ROBOT_CONVERSATION_AUTO_FILL_MAX_AGE_MINUTES`
- `AI_ROBOT_CONVERSATION_AUTO_FILL_MAX_TURNS`
- `AI_ROBOT_CONVERSATION_SUMMARY_MAX_TURNS`

**推荐说明：**
- `AI_ROBOT_CONVERSATION_AUTO_FILL_MAX_AGE_MINUTES`：控制自动补参的时间窗口，默认 `10`
- `AI_ROBOT_CONVERSATION_AUTO_FILL_MAX_TURNS`：控制自动补参参考的最近连续轮次，默认 `3`
- `AI_ROBOT_CONVERSATION_SUMMARY_MAX_TURNS`：控制提供给模型的上下文摘要轮次，默认 `5`
- `ai_robot` 会优先读取 `AI_ROBOT_CONVERSATION_MONGO_*`；未单独配置时自动回退到公共 `MONGODB_*`
- 如果你们在 PHP/Laravel 里使用的是 `mongodb_cloud_platform_v2` 这类“连接名”，Go 侧这里仍然要填实际库名到 `MONGODB_DATABASE`
- `ai_robot` 上下文会按 `user_id + platform + conversation_id` 进行隔离；相同 `conversation_id` 在不同用户之间不会串话
- `user_id` 按字符串保存和查询，不要求必须是整型，兼容历史或外部系统的字符串用户标识
- `POST /api/v1/ai_robot/chat`、`GET/DELETE /api/v1/ai_robot/conversation`、`GET/DELETE /api/v1/ai_robot/conversations` 现在都要求登录态
- 普通用户只能查看/删除自己的会话；管理员可以查看和删除全量会话
- 当 `enable_context=true` 时，后端必须能拿到当前登录用户，否则不会启用上下文持久化
- 建议让 `SUMMARY_MAX_TURNS >= AUTO_FILL_MAX_TURNS`
- 建议生产默认先保持 `10 / 3 / 5`，再按真实对话体验微调

**一个建议配置示例：**

```env
CLOUD_PLATFORM_BASE_URL=http://local.cloud_platform_api.com
AI_GATEWAY_LLM_API_KEY=your-llm-api-key
MONGODB_HOST=121.43.37.234
MONGODB_PORT=27117
MONGODB_DATABASE=cloud_platform
MONGODB_USERNAME=root
MONGODB_PASSWORD=your-mongodb-password
MONGODB_AUTH_SOURCE=admin
AI_ROBOT_CONVERSATION_STORE_DRIVER=mongodb
AI_ROBOT_CONVERSATION_MONGO_COLLECTION=ai_robot_conversations
AI_ROBOT_CONVERSATION_AUTO_FILL_MAX_AGE_MINUTES=10
AI_ROBOT_CONVERSATION_AUTO_FILL_MAX_TURNS=3
AI_ROBOT_CONVERSATION_SUMMARY_MAX_TURNS=5
```
- `GetMaxFileSizeBytes()`: 获取最大文件大小（字节）
- `IsFileTypeAllowed()`: 检查文件类型是否允许
- `GetPublicFilePath()`: 获取公共文件路径
- `GetPrivateFilePath()`: 获取私有文件路径
- `GetAllowedTypesString()`: 获取允许的文件类型字符串

## 使用方法

### 基本使用

```go
package main

import (
    "log"
    "your-project/app/Config"
)

func main() {
    // 1. 加载配置
    Config.LoadConfig()
    
    // 2. 验证配置
    if err := Config.ValidateConfig(); err != nil {
        log.Fatal("配置验证失败:", err)
    }
    
    // 3. 获取配置
    config := Config.GetConfig()
    serverConfig := Config.GetServerConfig()
    dbConfig := Config.GetDatabaseConfig()
    
    // 4. 使用配置
    log.Printf("服务器端口: %s", serverConfig.Port)
    log.Printf("数据库驱动: %s", dbConfig.Driver)
}
```

### 高级使用

```go
// 服务器配置
serverConfig := Config.GetServerConfig()
if serverConfig.IsDebugMode() {
    log.Println("调试模式")
}

// 数据库配置
dbConfig := Config.GetDatabaseConfig()
dsn := dbConfig.GetDSN()
if dbConfig.IsSQLite() {
    log.Println("使用SQLite数据库")
}

// JWT配置
jwtConfig := Config.GetJWTConfig()
if jwtConfig.IsSecretDefault() {
    log.Println("警告：使用默认JWT密钥")
}

// 存储配置
storageConfig := Config.GetStorageConfig()
if storageConfig.IsFileTypeAllowed("jpg") {
    log.Println("允许上传jpg文件")
}
```

## 环境变量配置

### 完整配置示例

创建 `.env` 文件：

```env
# ===========================================
# 服务器配置
# ===========================================
SERVER_PORT=8080
SERVER_MODE=debug
SERVER_BASE_URL=http://localhost:8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=120s
SERVER_MAX_HEADER_BYTES=1048576

# ===========================================
# 数据库配置
# ===========================================
DB_DRIVER=sqlite
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=
DB_DATABASE=cloud_platform.db
DB_CHARSET=utf8mb4
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=3600s

# ===========================================
# JWT配置
# ===========================================
JWT_SECRET=your-super-secret-jwt-key-change-in-production-must-be-at-least-32-characters-long
JWT_EXPIRE_TIME=24
JWT_REFRESH_EXPIRE_TIME=168
JWT_ISSUER=cloud-platform-api
JWT_AUDIENCE=cloud-platform-users

# ===========================================
# Redis配置
# ===========================================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DATABASE=0
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5
REDIS_MAX_RETRIES=3
REDIS_DIAL_TIMEOUT=5s
REDIS_READ_TIMEOUT=3s
REDIS_WRITE_TIMEOUT=3s

# ===========================================
# 存储配置
# ===========================================
STORAGE_UPLOAD_PATH=./storage/app/public
STORAGE_MAX_FILE_SIZE=10
STORAGE_ALLOWED_TYPES=jpg,jpeg,png,gif,pdf,doc,docx
STORAGE_PRIVATE_PATH=./storage/app/private
STORAGE_PUBLIC_PATH=./storage/app/public
STORAGE_TEMP_PATH=./storage/temp
STORAGE_LOG_PATH=./storage/logs
STORAGE_CACHE_PATH=./storage/framework/cache

# ===========================================
# 安全配置
# ===========================================
SECURITY_ENABLE_XSS_PROTECTION=true
SECURITY_ENABLE_SQL_INJECTION_CHECK=true
SECURITY_ENABLE_CSRF_PROTECTION=true
SECURITY_ENABLE_RATE_LIMIT=true
SECURITY_MAX_LOGIN_ATTEMPTS=5
SECURITY_LOCKOUT_DURATION=15m
SECURITY_PASSWORD_MIN_LENGTH=8
SECURITY_PASSWORD_REQUIRE_UPPERCASE=true
SECURITY_PASSWORD_REQUIRE_LOWERCASE=true
SECURITY_PASSWORD_REQUIRE_NUMBER=true
SECURITY_PASSWORD_REQUIRE_SYMBOL=true

# ===========================================
# 监控配置
# ===========================================
MONITORING_ENABLE_METRICS=true
MONITORING_ENABLE_HEALTH_CHECK=true
MONITORING_ENABLE_PROMETHEUS=true
MONITORING_METRICS_PATH=/metrics
MONITORING_HEALTH_PATH=/health
MONITORING_LOG_LEVEL=info
MONITORING_LOG_FORMAT=json

# ===========================================
# 邮件配置
# ===========================================
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM_NAME=Cloud Platform API
EMAIL_FROM_ADDRESS=noreply@example.com
EMAIL_USE_TLS=true
EMAIL_USE_SSL=false

# ===========================================
# 日志配置
# ===========================================
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_FILE_PATH=./storage/logs/app.log
LOG_MAX_SIZE=100
LOG_MAX_AGE=30
LOG_MAX_BACKUPS=10
LOG_COMPRESS=true
```

### 环境特定配置

#### 开发环境 (.env.development)
```env
SERVER_MODE=debug
LOG_LEVEL=debug
LOG_FORMAT=text
DB_DRIVER=sqlite
REDIS_HOST=
MONITORING_ENABLE_METRICS=false
SECURITY_ENABLE_RATE_LIMIT=false
```

#### 测试环境 (.env.testing)
```env
SERVER_MODE=debug
LOG_LEVEL=info
LOG_FORMAT=json
DB_DRIVER=sqlite
REDIS_HOST=localhost
MONITORING_ENABLE_METRICS=true
SECURITY_ENABLE_RATE_LIMIT=true
```

#### 生产环境 (.env.production)
```env
SERVER_MODE=production
LOG_LEVEL=warn
LOG_FORMAT=json
DB_DRIVER=mysql
REDIS_HOST=redis-server
MONITORING_ENABLE_METRICS=true
SECURITY_ENABLE_RATE_LIMIT=true
# 生产环境必须修改所有默认密码和密钥
```

## 向后兼容性

原有的配置使用方式仍然有效：

```go
// 这些调用仍然有效
config := Config.GetConfig()
serverPort := config.Server.Port
dbDriver := config.Database.Driver
```

## 最佳实践

### 1. 配置验证
```go
// 在应用启动时验证配置
func main() {
    // 加载配置
    Config.LoadConfig()
    
    // 验证配置
    if err := Config.ValidateConfig(); err != nil {
        log.Fatal("配置验证失败:", err)
    }
    
    // 启动应用
    startServer()
}
```

### 2. 环境变量管理
```go
// 使用环境变量而不是硬编码配置
func getDatabaseConfig() *DatabaseConfig {
    return &DatabaseConfig{
        Driver:   os.Getenv("DB_DRIVER"),
        Host:     os.Getenv("DB_HOST"),
        Port:     os.Getenv("DB_PORT"),
        Username: os.Getenv("DB_USERNAME"),
        Password: os.Getenv("DB_PASSWORD"),
        Database: os.Getenv("DB_DATABASE"),
    }
}
```

### 3. 默认值设置
```go
// 为所有配置项提供合理的默认值
func (s *ServerConfig) SetDefaults() {
    viper.SetDefault("server.port", "8080")
    viper.SetDefault("server.mode", "debug")
    viper.SetDefault("server.base_url", "http://localhost:8080")
    viper.SetDefault("server.read_timeout", "30s")
    viper.SetDefault("server.write_timeout", "30s")
    viper.SetDefault("server.idle_timeout", "120s")
}
```

### 4. 类型检查
```go
// 使用提供的类型检查方法
func setupDatabase() {
    dbConfig := Config.GetDatabaseConfig()
    
    if dbConfig.IsSQLite() {
        // SQLite特定配置
        setupSQLite()
    } else if dbConfig.IsMySQL() {
        // MySQL特定配置
        setupMySQL()
    } else if dbConfig.IsPostgreSQL() {
        // PostgreSQL特定配置
        setupPostgreSQL()
    }
}
```

### 5. 错误处理
```go
// 正确处理配置验证错误
func validateConfig() error {
    var errors []string
    
    // 验证服务器配置
    if err := validateServerConfig(); err != nil {
        errors = append(errors, fmt.Sprintf("服务器配置错误: %v", err))
    }
    
    // 验证数据库配置
    if err := validateDatabaseConfig(); err != nil {
        errors = append(errors, fmt.Sprintf("数据库配置错误: %v", err))
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("配置验证失败:\n%s", strings.Join(errors, "\n"))
    }
    
    return nil
}
```

### 6. 安全配置
```go
// 生产环境使用强密钥和安全的配置
func validateSecurityConfig() error {
    jwtConfig := Config.GetJWTConfig()
    
    if jwtConfig.IsSecretDefault() {
        return fmt.Errorf("生产环境必须修改JWT密钥")
    }
    
    if len(jwtConfig.Secret) < 32 {
        return fmt.Errorf("JWT密钥长度必须至少32个字符")
    }
    
    return nil
}
```

### 7. 配置热重载
```go
// 支持配置热重载
func setupConfigReload() {
    viper.WatchConfig()
    viper.OnConfigChange(func(e fsnotify.Event) {
        log.Println("配置文件已更改，重新加载配置...")
        Config.LoadConfig()
    })
}
```

### 8. 配置加密
```go
// 敏感配置加密存储
func decryptConfig(key string) (string, error) {
    // 使用AES加密解密敏感配置
    encrypted := os.Getenv(key)
    if encrypted == "" {
        return "", fmt.Errorf("配置项 %s 未设置", key)
    }
    
    // 解密逻辑
    decrypted, err := aesDecrypt(encrypted, getEncryptionKey())
    if err != nil {
        return "", fmt.Errorf("解密配置失败: %v", err)
    }
    
    return decrypted, nil
}
```

### 9. 配置验证规则
```go
// 自定义配置验证规则
func validateServerPort(port string) error {
    portNum, err := strconv.Atoi(port)
    if err != nil {
        return fmt.Errorf("端口号必须是数字")
    }
    
    if portNum < 1 || portNum > 65535 {
        return fmt.Errorf("端口号必须在1-65535范围内")
    }
    
    return nil
}
```

### 10. 配置文档生成
```go
// 自动生成配置文档
func generateConfigDocs() {
    config := Config.GetConfig()
    
    // 生成Markdown格式的配置文档
    doc := generateMarkdownDoc(config)
    
    // 写入文件
    err := ioutil.WriteFile("CONFIG.md", []byte(doc), 0644)
    if err != nil {
        log.Printf("生成配置文档失败: %v", err)
    }
}
```

## 扩展配置

如需添加新的配置类型：

1. 创建新的配置文件（如 `email.go`）
2. 定义配置结构体并实现三个核心方法：
   - `SetDefaults()`: 设置默认值
   - `BindEnvs()`: 绑定环境变量
   - `Validate()`: 验证配置
3. 在 `base.go` 的 `Config` 结构体中添加新字段
4. 在 `setDefaults()` 和 `bindEnvs()` 中调用新配置的方法
5. 在 `ValidateConfig()` 中添加新配置的验证
6. 在 `env.example` 中添加环境变量示例

### 示例：添加邮件配置

```go
// email.go
package Config

import (
    "fmt"
    "github.com/spf13/viper"
)

type EmailConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
}

func (e *EmailConfig) SetDefaults() {
    viper.SetDefault("email.host", "smtp.gmail.com")
    viper.SetDefault("email.port", 587)
}

func (e *EmailConfig) BindEnvs() {
    viper.BindEnv("email.host", "EMAIL_HOST")
    viper.BindEnv("email.port", "EMAIL_PORT")
    viper.BindEnv("email.username", "EMAIL_USERNAME")
    viper.BindEnv("email.password", "EMAIL_PASSWORD")
}

func (e *EmailConfig) Validate() error {
    if e.Host == "" {
        return fmt.Errorf("邮件服务器主机未配置")
    }
    return nil
}
```

## 配置验证规则

### 服务器配置验证
- 端口必须在有效范围内（1-65535）
- 运行模式必须是有效的值（debug/production）
- 基础URL必须是有效的URL格式

### 数据库配置验证
- 驱动必须是支持的类型（sqlite/mysql/postgres）
- 主机和端口不能为空（SQLite除外）
- 数据库名不能为空

### JWT配置验证
- 密钥不能为空
- 过期时间必须大于0

### Redis配置验证
- 主机和端口不能为空
- 数据库编号必须在有效范围内（0-15）

### 存储配置验证
- 所有路径必须有效
- 文件大小限制必须大于0
- 允许的文件类型不能为空

## 故障排除

### 常见问题

1. **配置加载失败**
   - 检查环境变量名称是否正确
   - 检查配置文件格式是否正确
   - 检查文件权限

2. **配置验证失败**
   - 查看具体的验证错误信息
   - 检查配置值的格式和范围
   - 确保所有必需配置都已设置

3. **环境变量不生效**
   - 检查环境变量名称是否与代码中的绑定一致
   - 确保环境变量已正确设置
   - 重启应用以重新加载配置

### 调试方法

1. **启用调试模式**
```go
viper.SetDebug(true)
```

2. **打印配置信息**
```go
config := Config.GetConfig()
fmt.Printf("%+v\n", config)
```

3. **检查环境变量**
```bash
env | grep -E "(SERVER|DB|JWT|REDIS|STORAGE)"
```
