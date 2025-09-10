# 安全防护系统文档

## 📋 概述

安全防护系统是Cloud Platform API项目的核心安全组件，提供全面的安全防护功能，包括密码策略、访问控制、异常检测、威胁防护、安全审计等。系统通过多层次的安全防护机制，确保应用程序和用户数据的安全。

## 🏗️ 系统架构

### 核心组件

```
app/Config/security.go                    # 安全防护配置管理
app/Models/Security.go                    # 安全数据模型
app/Services/SecurityService.go           # 安全防护核心服务
app/Http/Middleware/SecurityMiddleware.go # 安全防护中间件
app/Http/Controllers/SecurityController.go # 安全防护API控制器
app/Http/Routes/security.go               # 安全防护路由配置
docs/SECURITY_SYSTEM.md                   # 本文档
```

### 系统功能模块

1. **密码策略管理** - 强密码策略、密码历史、密码强度评估
2. **访问控制** - RBAC权限控制、资源级访问、时间/位置/设备限制
3. **异常检测** - 行为分析、模式识别、机器学习异常检测
4. **威胁防护** - 威胁情报、恶意软件检测、钓鱼防护、SQL注入/XSS防护
5. **安全审计** - 事件记录、合规报告、实时监控、告警管理
6. **账户管理** - 登录尝试监控、账户锁定、会话管理

## 🚀 快速开始

### 1. 服务初始化

```go
// 在主应用中初始化安全防护服务
config := &Config.SecurityConfig{}
config.SetDefaults()
config.BindEnvs()

securityService := Services.NewSecurityService(db, config)

// 设置到控制器
controller := Controllers.NewSecurityController()
controller.SetSecurityService(securityService)

// 注册路由
Routes.RegisterSecurityRoutes(router, controller)
```

### 2. 中间件集成

```go
// 创建安全中间件
securityMiddleware := Middleware.NewSecurityMiddleware(securityService)

// 在路由中使用安全中间件
router.Use(securityMiddleware.SecurityCheck())
router.Use(securityMiddleware.XSSProtection())
router.Use(securityMiddleware.SQLInjectionProtection())
router.Use(securityMiddleware.CSRFProtection())
router.Use(securityMiddleware.FileUploadSecurity())
router.Use(securityMiddleware.RateLimitSecurity())
router.Use(securityMiddleware.ContentSecurityPolicy())
router.Use(securityMiddleware.SecurityHeaders())
```

## 🔧 功能特性

### 1. 密码策略管理

#### 密码强度验证
```go
// 验证密码强度
valid, errors := securityService.ValidatePassword(password, username)
if !valid {
    for _, err := range errors {
        fmt.Println(err)
    }
}
```

#### 密码历史检查
```go
// 检查密码历史
allowed := securityService.CheckPasswordHistory(userID, passwordHash)
if !allowed {
    return errors.New("密码不能与历史密码相同")
}
```

#### 密码更改记录
```go
// 记录密码更改
err := securityService.RecordPasswordChange(
    userID, 
    changedBy, 
    passwordHash, 
    "password_reset", 
    ipAddress, 
    userAgent,
)
```

### 2. 访问控制

#### 权限检查
```go
// 检查用户权限
allowed, reason := securityService.CheckAccessControl(userID, resource, action)
if !allowed {
    return errors.New(reason)
}
```

#### 时间限制访问
```go
// 检查时间限制
allowed := securityService.checkTimeRestriction(timeRestriction)
```

### 3. 异常检测

#### 行为分析
```go
// 检测异常行为
isAnomaly, score := securityService.DetectAnomaly(
    userID, 
    eventType, 
    resource, 
    action, 
    ipAddress, 
    userAgent,
)
if isAnomaly {
    // 处理异常
    securityService.RecordSecurityEvent(...)
}
```

#### 模式识别
```go
// 分析用户行为模式
patternScore := securityService.analyzePattern(userID, eventType, resource, action, ipAddress)
```

### 4. 威胁防护

#### 威胁情报检查
```go
// 检查威胁情报
allowed, reason := securityService.CheckThreatProtection(ipAddress, url, fileHash)
if !allowed {
    return errors.New(reason)
}
```

#### 恶意软件检测
```go
// 检查恶意文件
isMalware := securityService.isMalwareFile(fileHash)
if isMalware {
    return errors.New("文件被识别为恶意软件")
}
```

### 5. 安全审计

#### 事件记录
```go
// 记录安全事件
err := securityService.RecordSecurityEvent(
    userID,
    eventType,
    eventLevel,
    ipAddress,
    userAgent,
    resource,
    action,
    details,
    riskScore,
    anomalyScore,
    blocked,
    alerted,
    location,
    deviceInfo,
)
```

#### 报告生成
```go
// 生成安全报告
report, err := securityService.GenerateSecurityReport(
    reportType,
    period,
    startDate,
    endDate,
    generatedBy,
)
```

## 📡 API接口

### 安全事件管理

#### 获取安全事件列表
```http
GET /api/v1/security/events?page=1&limit=20&event_type=login&event_level=high
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "events": [
      {
        "id": 1,
        "event_type": "login",
        "event_level": "high",
        "user_id": 123,
        "username": "admin",
        "ip_address": "192.168.1.100",
        "user_agent": "Mozilla/5.0...",
        "resource": "/api/v1/auth/login",
        "action": "POST",
        "details": "登录成功",
        "risk_score": 85.5,
        "anomaly_score": 0.0,
        "blocked": false,
        "alerted": false,
        "created_at": "2024-12-01T10:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "limit": 20,
    "total_pages": 5
  }
}
```

### 威胁情报管理

#### 获取威胁情报列表
```http
GET /api/v1/security/threats?page=1&limit=20&threat_type=malware&severity=high
```

### 登录尝试管理

#### 获取登录尝试记录
```http
GET /api/v1/security/login-attempts?page=1&limit=20&username=admin&success=false
```

### 账户锁定管理

#### 获取账户锁定记录
```http
GET /api/v1/security/account-lockouts?page=1&limit=20&active=true
```

#### 解锁账户
```http
POST /api/v1/security/account-lockouts/123/unlock
Content-Type: application/json

{
  "reason": "管理员手动解锁"
}
```

### 安全告警管理

#### 获取安全告警列表
```http
GET /api/v1/security/alerts?page=1&limit=20&severity=high&status=open
```

#### 确认告警
```http
POST /api/v1/security/alerts/456/acknowledge
```

#### 解决告警
```http
POST /api/v1/security/alerts/456/resolve
Content-Type: application/json

{
  "resolution_notes": "已处理异常登录尝试"
}
```

### 安全报告管理

#### 获取安全报告列表
```http
GET /api/v1/security/reports?page=1&limit=20&report_type=login_attempts&status=published
```

#### 生成安全报告
```http
POST /api/v1/security/reports/generate
Content-Type: application/json

{
  "report_type": "security_events",
  "period": "weekly",
  "start_date": "2024-12-01T00:00:00Z",
  "end_date": "2024-12-07T23:59:59Z"
}
```

### 安全指标管理

#### 获取安全指标
```http
GET /api/v1/security/metrics?category=login&time_window=24h
```

### 安全仪表板

#### 获取仪表板数据
```http
GET /api/v1/security/dashboard
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "today_events": 150,
    "today_alerts": 5,
    "today_lockouts": 2,
    "active_threats": 25,
    "risk_distribution": [
      {"risk_level": "low", "count": 100},
      {"risk_level": "medium", "count": 40},
      {"risk_level": "high", "count": 10}
    ],
    "event_type_distribution": [
      {"event_type": "login", "count": 80},
      {"event_type": "api_access", "count": 50},
      {"event_type": "file_upload", "count": 20}
    ]
  }
}
```

## ⚙️ 配置说明

### 环境变量配置

```bash
# 基础安全配置
SECURITY_ENABLED=true
SECURITY_SESSION_TIMEOUT=30m
SECURITY_MAX_LOGIN_ATTEMPTS=5
SECURITY_LOGIN_LOCKOUT_DURATION=15m
SECURITY_PASSWORD_HISTORY_COUNT=5
SECURITY_FORCE_PASSWORD_CHANGE=false
SECURITY_PASSWORD_CHANGE_INTERVAL=2160h
SECURITY_ACCOUNT_LOCKOUT_THRESHOLD=10
SECURITY_ACCOUNT_LOCKOUT_DURATION=1h
SECURITY_INACTIVE_ACCOUNT_TIMEOUT=4320h
SECURITY_CONCURRENT_SESSION_LIMIT=3
SECURITY_IP_WHITELIST_ENABLED=false
SECURITY_IP_BLACKLIST_ENABLED=true
SECURITY_RATE_LIMIT_ENABLED=true
SECURITY_RATE_LIMIT_REQUESTS=100
SECURITY_RATE_LIMIT_WINDOW=1m

# 密码策略配置
SECURITY_PASSWORD_MIN_LENGTH=8
SECURITY_PASSWORD_MAX_LENGTH=128
SECURITY_PASSWORD_REQUIRE_UPPERCASE=true
SECURITY_PASSWORD_REQUIRE_LOWERCASE=true
SECURITY_PASSWORD_REQUIRE_NUMBERS=true
SECURITY_PASSWORD_REQUIRE_SPECIAL_CHARS=true
SECURITY_PASSWORD_SPECIAL_CHARS_LIST="!@#$%^&*()_+-=[]{}|;:,.<>?"
SECURITY_PASSWORD_PREVENT_COMMON=true
SECURITY_PASSWORD_COMMON_FILE="config/common_passwords.txt"
SECURITY_PASSWORD_PREVENT_USERNAME=true
SECURITY_PASSWORD_PREVENT_SEQUENTIAL=true
SECURITY_PASSWORD_PREVENT_REPEATED=true
SECURITY_PASSWORD_MAX_REPEATED_CHARS=3
SECURITY_PASSWORD_STRENGTH_THRESHOLD=70

# 访问控制配置
SECURITY_RBAC_ENABLED=true
SECURITY_PERMISSION_CACHE_ENABLED=true
SECURITY_PERMISSION_CACHE_TTL=5m
SECURITY_DEFAULT_DENY_POLICY=true
SECURITY_RESOURCE_LEVEL_ACCESS=true
SECURITY_TIME_BASED_ACCESS=false
SECURITY_LOCATION_BASED_ACCESS=false
SECURITY_DEVICE_BASED_ACCESS=false
SECURITY_SESSION_BASED_ACCESS=true
SECURITY_API_KEY_PERMISSIONS=true
SECURITY_JWT_CLAIMS_VALIDATION=true
SECURITY_TOKEN_REFRESH_ENABLED=true
SECURITY_TOKEN_REFRESH_THRESHOLD=5m

# 异常检测配置
SECURITY_ANOMALY_DETECTION_ENABLED=true
SECURITY_ANOMALY_LEARNING_MODE=true
SECURITY_ANOMALY_LEARNING_PERIOD=168h
SECURITY_ANOMALY_THRESHOLD=0.8
SECURITY_ANOMALY_BEHAVIORAL_ANALYSIS=true
SECURITY_ANOMALY_PATTERN_RECOGNITION=true
SECURITY_ANOMALY_ML_ENABLED=false
SECURITY_ANOMALY_ML_MODEL_PATH="models/anomaly_detection.model"
SECURITY_ANOMALY_ML_TRAINING_DATA_PATH="data/training/"
SECURITY_ANOMALY_REAL_TIME_ANALYSIS=true
SECURITY_ANOMALY_BATCH_ANALYSIS=true
SECURITY_ANOMALY_ANALYSIS_INTERVAL=5m
SECURITY_ANOMALY_ALERT_ON_ANOMALY=true
SECURITY_ANOMALY_AUTO_BLOCK=false
SECURITY_ANOMALY_SCORE_THRESHOLD=0.7

# 安全审计配置
SECURITY_AUDIT_ENABLED=true
SECURITY_AUDIT_LEVEL=medium
SECURITY_AUDIT_EVENTS=login,logout,password_change,permission_change,data_access,admin_action
SECURITY_AUDIT_DATA_RETENTION=8760h
SECURITY_AUDIT_ENCRYPTION_ENABLED=true
SECURITY_AUDIT_ENCRYPTION_KEY=""
SECURITY_AUDIT_COMPRESSION_ENABLED=true
SECURITY_AUDIT_REAL_TIME_MONITORING=true
SECURITY_AUDIT_ALERT_ON_SUSPICIOUS=true
SECURITY_AUDIT_COMPLIANCE_REPORTING=true
SECURITY_AUDIT_REPORT_GENERATION=true
SECURITY_AUDIT_REPORT_SCHEDULE=weekly
SECURITY_AUDIT_DATA_EXPORT_ENABLED=true
SECURITY_AUDIT_DATA_EXPORT_FORMAT=json

# 威胁防护配置
SECURITY_THREAT_PROTECTION_ENABLED=true
SECURITY_THREAT_INTELLIGENCE=true
SECURITY_THREAT_TI_UPDATE_INTERVAL=24h
SECURITY_THREAT_TI_SOURCE_URLS=https://api.abuseipdb.com/api/v2/blacklist,https://api.blocklist.de/get.php
SECURITY_THREAT_MALWARE_SCANNING=true
SECURITY_THREAT_SCAN_INTERVAL=1h
SECURITY_THREAT_VIRUS_TOTAL_API_KEY=""
SECURITY_THREAT_PHISHING_PROTECTION=true
SECURITY_THREAT_PHISHING_URLS_FILE="config/phishing_urls.txt"
SECURITY_THREAT_SQL_INJECTION_PROTECTION=true
SECURITY_THREAT_XSS_PROTECTION=true
SECURITY_THREAT_CSRF_PROTECTION=true
SECURITY_THREAT_CSRF_TOKEN_EXPIRY=30m
SECURITY_THREAT_FILE_UPLOAD_SCANNING=true
SECURITY_THREAT_ALLOWED_FILE_TYPES=.jpg,.jpeg,.png,.gif,.pdf,.doc,.docx,.txt
SECURITY_THREAT_MAX_FILE_SIZE=10485760
SECURITY_THREAT_BLOCKED_FILE_TYPES=.exe,.bat,.cmd,.com,.pif,.scr,.vbs,.js
SECURITY_THREAT_CONTENT_SECURITY_POLICY=true
SECURITY_THREAT_CSP_DIRECTIVES="default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';"
```

### 配置文件示例

```yaml
# config/security.yaml
security:
  base_security:
    enabled: true
    session_timeout: 30m
    max_login_attempts: 5
    login_lockout_duration: 15m
    password_history_count: 5
    force_password_change: false
    password_change_interval: 2160h
    account_lockout_threshold: 10
    account_lockout_duration: 1h
    inactive_account_timeout: 4320h
    concurrent_session_limit: 3
    ip_whitelist_enabled: false
    ip_blacklist_enabled: true
    rate_limit_enabled: true
    rate_limit_requests: 100
    rate_limit_window: 1m

  password_policy:
    min_length: 8
    max_length: 128
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_special_chars: true
    special_chars_list: "!@#$%^&*()_+-=[]{}|;:,.<>?"
    prevent_common_passwords: true
    common_passwords_file: "config/common_passwords.txt"
    prevent_username_in_password: true
    prevent_sequential_chars: true
    prevent_repeated_chars: true
    max_repeated_chars: 3
    password_strength_threshold: 70

  access_control:
    rbac_enabled: true
    permission_cache_enabled: true
    permission_cache_ttl: 5m
    default_deny_policy: true
    resource_level_access: true
    time_based_access: false
    location_based_access: false
    device_based_access: false
    session_based_access: true
    api_key_permissions: true
    jwt_claims_validation: true
    token_refresh_enabled: true
    token_refresh_threshold: 5m

  anomaly_detection:
    enabled: true
    learning_mode: true
    learning_period: 168h
    anomaly_threshold: 0.8
    behavioral_analysis: true
    pattern_recognition: true
    machine_learning_enabled: false
    ml_model_path: "models/anomaly_detection.model"
    ml_training_data_path: "data/training/"
    real_time_analysis: true
    batch_analysis: true
    analysis_interval: 5m
    alert_on_anomaly: true
    auto_block_on_anomaly: false
    anomaly_score_threshold: 0.7

  security_audit:
    enabled: true
    audit_level: medium
    audit_events:
      - login
      - logout
      - password_change
      - permission_change
      - data_access
      - admin_action
    data_retention: 8760h
    encryption_enabled: true
    encryption_key: ""
    compression_enabled: true
    real_time_monitoring: true
    alert_on_suspicious_activity: true
    compliance_reporting: true
    report_generation: true
    report_schedule: weekly
    data_export_enabled: true
    data_export_format: json

  threat_protection:
    enabled: true
    threat_intelligence: true
    ti_update_interval: 24h
    ti_source_urls:
      - https://api.abuseipdb.com/api/v2/blacklist
      - https://api.blocklist.de/get.php
    malware_scanning: true
    scan_interval: 1h
    virus_total_api_key: ""
    phishing_protection: true
    phishing_urls_file: "config/phishing_urls.txt"
    sql_injection_protection: true
    xss_protection: true
    csrf_protection: true
    csrf_token_expiry: 30m
    file_upload_scanning: true
    allowed_file_types:
      - .jpg
      - .jpeg
      - .png
      - .gif
      - .pdf
      - .doc
      - .docx
      - .txt
    max_file_size: 10485760
    blocked_file_types:
      - .exe
      - .bat
      - .cmd
      - .com
      - .pif
      - .scr
      - .vbs
      - .js
    content_security_policy: true
    csp_directives: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';"
```

## 🔍 使用示例

### 1. 密码验证示例

```go
// 在用户注册或密码更改时验证密码
func validateUserPassword(password, username string) error {
    valid, errors := securityService.ValidatePassword(password, username)
    if !valid {
        return fmt.Errorf("密码不符合要求: %v", errors)
    }
    return nil
}
```

### 2. 登录安全检查示例

```go
// 在登录处理中集成安全检查
func handleLogin(username, password, ipAddress, userAgent string) error {
    // 检查登录尝试
    allowed, reason := securityService.CheckLoginAttempts(username, ipAddress)
    if !allowed {
        return errors.New(reason)
    }

    // 验证用户凭据
    user, err := userService.ValidateUser(username, password)
    if err != nil {
        // 记录失败的登录尝试
        securityService.RecordLoginAttempt(username, ipAddress, userAgent, err.Error(), false, "", "")
        return err
    }

    // 记录成功的登录尝试
    securityService.RecordLoginAttempt(username, ipAddress, userAgent, "", true, "", "")

    // 检测异常
    isAnomaly, score := securityService.DetectAnomaly(
        user.ID, 
        "login", 
        "/api/v1/auth/login", 
        "POST", 
        ipAddress, 
        userAgent,
    )
    if isAnomaly {
        // 处理异常登录
        log.Printf("检测到异常登录，评分: %.2f", score)
    }

    return nil
}
```

### 3. 访问控制示例

```go
// 在API处理中检查访问权限
func handleAPIRequest(userID uint, resource, action string) error {
    allowed, reason := securityService.CheckAccessControl(userID, resource, action)
    if !allowed {
        return errors.New(reason)
    }
    return nil
}
```

### 4. 威胁防护示例

```go
// 在文件上传时检查威胁
func handleFileUpload(ipAddress, fileHash string) error {
    allowed, reason := securityService.CheckThreatProtection(ipAddress, "", fileHash)
    if !allowed {
        return errors.New(reason)
    }
    return nil
}
```

## 🛠️ 故障排除

### 常见问题

#### 1. 密码验证失败
**问题：** 密码验证总是失败
**解决方案：**
- 检查密码策略配置是否正确
- 确认密码长度、字符类型要求
- 检查常见密码文件是否存在

#### 2. 异常检测误报
**问题：** 正常操作被误判为异常
**解决方案：**
- 调整异常检测阈值
- 增加学习模式时间
- 检查用户行为模式

#### 3. 威胁情报更新失败
**问题：** 威胁情报无法更新
**解决方案：**
- 检查网络连接
- 验证威胁情报源URL
- 检查API密钥配置

#### 4. 性能问题
**问题：** 安全检查影响性能
**解决方案：**
- 启用权限缓存
- 调整检查频率
- 优化数据库查询

### 日志分析

#### 安全事件日志
```bash
# 查看安全事件日志
tail -f logs/security/security_events.log

# 搜索特定事件
grep "anomaly_detected" logs/security/security_events.log
```

#### 错误日志
```bash
# 查看错误日志
tail -f logs/security/error.log

# 搜索特定错误
grep "password_validation" logs/security/error.log
```

## 🔮 未来规划

### 短期目标（1-3个月）
- [ ] 集成更多威胁情报源
- [ ] 增强机器学习异常检测
- [ ] 添加地理位置异常检测
- [ ] 实现设备指纹识别

### 中期目标（3-6个月）
- [ ] 集成SIEM系统
- [ ] 添加安全评分系统
- [ ] 实现自动化响应
- [ ] 支持多租户安全隔离

### 长期目标（6-12个月）
- [ ] 集成AI安全助手
- [ ] 实现预测性安全分析
- [ ] 添加零信任架构支持
- [ ] 实现安全态势感知

## 📚 相关资源

- [OWASP安全指南](https://owasp.org/)
- [NIST网络安全框架](https://www.nist.gov/cyberframework)
- [ISO 27001信息安全标准](https://www.iso.org/isoiec-27001-information-security.html)
- [CIS安全控制](https://www.cisecurity.org/controls/)

## 📞 技术支持

如有问题或建议，请通过以下方式联系：
- 项目Issues: GitHub Issues
- 代码审查: Pull Request
- 技术讨论: GitHub Discussions

---

**文档版本**: 1.0.0  
**最后更新**: 2024年12月  
**维护者**: 安全团队
