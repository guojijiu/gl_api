package Config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// SecurityConfig 安全防护配置
type SecurityConfig struct {
	// 基础安全配置
	BaseSecurity BaseSecurityConfig `mapstructure:"base_security"`
	
	// 密码策略配置
	PasswordPolicy PasswordPolicyConfig `mapstructure:"password_policy"`
	
	// 访问控制配置
	AccessControl AccessControlConfig `mapstructure:"access_control"`
	
	// 异常检测配置
	AnomalyDetection AnomalyDetectionConfig `mapstructure:"anomaly_detection"`
	
	// 安全审计配置
	SecurityAudit SecurityAuditConfig `mapstructure:"security_audit"`
	
	// 威胁防护配置
	ThreatProtection ThreatProtectionConfig `mapstructure:"threat_protection"`
}

// BaseSecurityConfig 基础安全配置
type BaseSecurityConfig struct {
	Enabled                    bool          `mapstructure:"enabled"`                      // 是否启用安全防护
	SessionTimeout             time.Duration `mapstructure:"session_timeout"`             // 会话超时时间
	MaxLoginAttempts           int           `mapstructure:"max_login_attempts"`          // 最大登录尝试次数
	LoginLockoutDuration       time.Duration `mapstructure:"login_lockout_duration"`      // 登录锁定时间
	PasswordHistoryCount       int           `mapstructure:"password_history_count"`      // 密码历史记录数量
	ForcePasswordChange        bool          `mapstructure:"force_password_change"`       // 是否强制密码更改
	PasswordChangeInterval     time.Duration `mapstructure:"password_change_interval"`   // 密码更改间隔
	AccountLockoutThreshold    int           `mapstructure:"account_lockout_threshold"`   // 账户锁定阈值
	AccountLockoutDuration     time.Duration `mapstructure:"account_lockout_duration"`    // 账户锁定时间
	InactiveAccountTimeout     time.Duration `mapstructure:"inactive_account_timeout"`    // 非活跃账户超时
	ConcurrentSessionLimit     int           `mapstructure:"concurrent_session_limit"`     // 并发会话限制
	IPWhitelistEnabled         bool          `mapstructure:"ip_whitelist_enabled"`         // 是否启用IP白名单
	IPBlacklistEnabled         bool          `mapstructure:"ip_blacklist_enabled"`         // 是否启用IP黑名单
	RateLimitEnabled           bool          `mapstructure:"rate_limit_enabled"`          // 是否启用速率限制
	RateLimitRequests          int           `mapstructure:"rate_limit_requests"`         // 速率限制请求数
	RateLimitWindow            time.Duration `mapstructure:"rate_limit_window"`           // 速率限制时间窗口
}

// PasswordPolicyConfig 密码策略配置
type PasswordPolicyConfig struct {
	MinLength                  int           `mapstructure:"min_length"`                   // 最小长度
	MaxLength                  int           `mapstructure:"max_length"`                   // 最大长度
	RequireUppercase           bool          `mapstructure:"require_uppercase"`           // 要求大写字母
	RequireLowercase           bool          `mapstructure:"require_lowercase"`           // 要求小写字母
	RequireNumbers             bool          `mapstructure:"require_numbers"`             // 要求数字
	RequireSpecialChars        bool          `mapstructure:"require_special_chars"`      // 要求特殊字符
	SpecialCharsList          string        `mapstructure:"special_chars_list"`          // 特殊字符列表
	PreventCommonPasswords     bool          `mapstructure:"prevent_common_passwords"`    // 防止常见密码
	CommonPasswordsFile        string        `mapstructure:"common_passwords_file"`       // 常见密码文件
	PreventUsernameInPassword  bool          `mapstructure:"prevent_username_in_password"` // 防止用户名出现在密码中
	PreventSequentialChars     bool          `mapstructure:"prevent_sequential_chars"`    // 防止连续字符
	PreventRepeatedChars       bool          `mapstructure:"prevent_repeated_chars"`      // 防止重复字符
	MaxRepeatedChars           int           `mapstructure:"max_repeated_chars"`         // 最大重复字符数
	PasswordStrengthThreshold  int           `mapstructure:"password_strength_threshold"` // 密码强度阈值
}

// AccessControlConfig 访问控制配置
type AccessControlConfig struct {
	RBACEnabled                bool          `mapstructure:"rbac_enabled"`                 // 是否启用RBAC
	PermissionCacheEnabled     bool          `mapstructure:"permission_cache_enabled"`   // 是否启用权限缓存
	PermissionCacheTTL         time.Duration `mapstructure:"permission_cache_ttl"`        // 权限缓存TTL
	DefaultDenyPolicy          bool          `mapstructure:"default_deny_policy"`         // 默认拒绝策略
	ResourceLevelAccess        bool          `mapstructure:"resource_level_access"`       // 资源级访问控制
	TimeBasedAccess            bool          `mapstructure:"time_based_access"`           // 基于时间的访问控制
	LocationBasedAccess        bool          `mapstructure:"location_based_access"`      // 基于位置的访问控制
	DeviceBasedAccess          bool          `mapstructure:"device_based_access"`        // 基于设备的访问控制
	SessionBasedAccess         bool          `mapstructure:"session_based_access"`       // 基于会话的访问控制
	APIKeyPermissions          bool          `mapstructure:"api_key_permissions"`         // API密钥权限
	JWTClaimsValidation        bool          `mapstructure:"jwt_claims_validation"`      // JWT声明验证
	TokenRefreshEnabled        bool          `mapstructure:"token_refresh_enabled"`       // 是否启用令牌刷新
	TokenRefreshThreshold      time.Duration `mapstructure:"token_refresh_threshold"`     // 令牌刷新阈值
}

// AnomalyDetectionConfig 异常检测配置
type AnomalyDetectionConfig struct {
	Enabled                    bool          `mapstructure:"enabled"`                      // 是否启用异常检测
	LearningMode               bool          `mapstructure:"learning_mode"`               // 学习模式
	LearningPeriod             time.Duration `mapstructure:"learning_period"`             // 学习周期
	AnomalyThreshold           float64       `mapstructure:"anomaly_threshold"`          // 异常阈值
	BehavioralAnalysis         bool          `mapstructure:"behavioral_analysis"`        // 行为分析
	PatternRecognition         bool          `mapstructure:"pattern_recognition"`       // 模式识别
	MachineLearningEnabled     bool          `mapstructure:"machine_learning_enabled"`   // 机器学习
	MLModelPath                string        `mapstructure:"ml_model_path"`              // ML模型路径
	MLTrainingDataPath         string        `mapstructure:"ml_training_data_path"`      // ML训练数据路径
	RealTimeAnalysis           bool          `mapstructure:"real_time_analysis"`         // 实时分析
	BatchAnalysis              bool          `mapstructure:"batch_analysis"`             // 批量分析
	AnalysisInterval           time.Duration `mapstructure:"analysis_interval"`          // 分析间隔
	AlertOnAnomaly             bool          `mapstructure:"alert_on_anomaly"`           // 异常时告警
	AutoBlockOnAnomaly         bool          `mapstructure:"auto_block_on_anomaly"`      // 异常时自动阻止
	AnomalyScoreThreshold      float64       `mapstructure:"anomaly_score_threshold"`    // 异常分数阈值
}

// SecurityAuditConfig 安全审计配置
type SecurityAuditConfig struct {
	Enabled                    bool          `mapstructure:"enabled"`                      // 是否启用安全审计
	AuditLevel                 string        `mapstructure:"audit_level"`                  // 审计级别
	AuditEvents                []string      `mapstructure:"audit_events"`                // 审计事件
	DataRetention              time.Duration `mapstructure:"data_retention"`              // 数据保留时间
	EncryptionEnabled          bool          `mapstructure:"encryption_enabled"`          // 是否启用加密
	EncryptionKey              string        `mapstructure:"encryption_key"`              // 加密密钥
	CompressionEnabled         bool          `mapstructure:"compression_enabled"`         // 是否启用压缩
	RealTimeMonitoring         bool          `mapstructure:"real_time_monitoring"`        // 实时监控
	AlertOnSuspiciousActivity  bool          `mapstructure:"alert_on_suspicious_activity"` // 可疑活动告警
	ComplianceReporting        bool          `mapstructure:"compliance_reporting"`        // 合规报告
	ReportGeneration           bool          `mapstructure:"report_generation"`            // 报告生成
	ReportSchedule             string        `mapstructure:"report_schedule"`              // 报告计划
	DataExportEnabled          bool          `mapstructure:"data_export_enabled"`         // 数据导出
	DataExportFormat           string        `mapstructure:"data_export_format"`         // 数据导出格式
}

// ThreatProtectionConfig 威胁防护配置
type ThreatProtectionConfig struct {
	Enabled                    bool          `mapstructure:"enabled"`                      // 是否启用威胁防护
	ThreatIntelligence         bool          `mapstructure:"threat_intelligence"`         // 威胁情报
	TIUpdateInterval           time.Duration `mapstructure:"ti_update_interval"`          // TI更新间隔
	TISourceURLs               []string      `mapstructure:"ti_source_urls"`              // TI源URL
	MalwareScanning            bool          `mapstructure:"malware_scanning"`            // 恶意软件扫描
	ScanInterval               time.Duration `mapstructure:"scan_interval"`               // 扫描间隔
	VirusTotalAPIKey           string        `mapstructure:"virus_total_api_key"`        // VirusTotal API密钥
	PhishingProtection         bool          `mapstructure:"phishing_protection"`         // 钓鱼防护
	PhishingURLsFile           string        `mapstructure:"phishing_urls_file"`          // 钓鱼URL文件
	SQLInjectionProtection     bool          `mapstructure:"sql_injection_protection"`    // SQL注入防护
	XSSProtection              bool          `mapstructure:"xss_protection"`               // XSS防护
	CSRFProtection             bool          `mapstructure:"csrf_protection"`             // CSRF防护
	CSRFTokenExpiry            time.Duration `mapstructure:"csrf_token_expiry"`           // CSRF令牌过期时间
	FileUploadScanning         bool          `mapstructure:"file_upload_scanning"`        // 文件上传扫描
	AllowedFileTypes           []string      `mapstructure:"allowed_file_types"`         // 允许的文件类型
	MaxFileSize                int64         `mapstructure:"max_file_size"`              // 最大文件大小
	BlockedFileTypes           []string      `mapstructure:"blocked_file_types"`         // 阻止的文件类型
	ContentSecurityPolicy      bool          `mapstructure:"content_security_policy"`     // 内容安全策略
	CSPDirectives              string        `mapstructure:"csp_directives"`              // CSP指令
}

// SetDefaults 设置默认值
func (c *SecurityConfig) SetDefaults() {
	// 基础安全配置默认值
	c.BaseSecurity.Enabled = true
	c.BaseSecurity.SessionTimeout = 30 * time.Minute
	c.BaseSecurity.MaxLoginAttempts = 5
	c.BaseSecurity.LoginLockoutDuration = 15 * time.Minute
	c.BaseSecurity.PasswordHistoryCount = 5
	c.BaseSecurity.ForcePasswordChange = false
	c.BaseSecurity.PasswordChangeInterval = 90 * 24 * time.Hour // 90天
	c.BaseSecurity.AccountLockoutThreshold = 10
	c.BaseSecurity.AccountLockoutDuration = 1 * time.Hour
	c.BaseSecurity.InactiveAccountTimeout = 180 * 24 * time.Hour // 180天
	c.BaseSecurity.ConcurrentSessionLimit = 3
	c.BaseSecurity.IPWhitelistEnabled = false
	c.BaseSecurity.IPBlacklistEnabled = true
	c.BaseSecurity.RateLimitEnabled = true
	c.BaseSecurity.RateLimitRequests = 100
	c.BaseSecurity.RateLimitWindow = 1 * time.Minute

	// 密码策略配置默认值
	c.PasswordPolicy.MinLength = 8
	c.PasswordPolicy.MaxLength = 128
	c.PasswordPolicy.RequireUppercase = true
	c.PasswordPolicy.RequireLowercase = true
	c.PasswordPolicy.RequireNumbers = true
	c.PasswordPolicy.RequireSpecialChars = true
	c.PasswordPolicy.SpecialCharsList = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	c.PasswordPolicy.PreventCommonPasswords = true
	c.PasswordPolicy.CommonPasswordsFile = "config/common_passwords.txt"
	c.PasswordPolicy.PreventUsernameInPassword = true
	c.PasswordPolicy.PreventSequentialChars = true
	c.PasswordPolicy.PreventRepeatedChars = true
	c.PasswordPolicy.MaxRepeatedChars = 3
	c.PasswordPolicy.PasswordStrengthThreshold = 70

	// 访问控制配置默认值
	c.AccessControl.RBACEnabled = true
	c.AccessControl.PermissionCacheEnabled = true
	c.AccessControl.PermissionCacheTTL = 5 * time.Minute
	c.AccessControl.DefaultDenyPolicy = true
	c.AccessControl.ResourceLevelAccess = true
	c.AccessControl.TimeBasedAccess = false
	c.AccessControl.LocationBasedAccess = false
	c.AccessControl.DeviceBasedAccess = false
	c.AccessControl.SessionBasedAccess = true
	c.AccessControl.APIKeyPermissions = true
	c.AccessControl.JWTClaimsValidation = true
	c.AccessControl.TokenRefreshEnabled = true
	c.AccessControl.TokenRefreshThreshold = 5 * time.Minute

	// 异常检测配置默认值
	c.AnomalyDetection.Enabled = true
	c.AnomalyDetection.LearningMode = true
	c.AnomalyDetection.LearningPeriod = 7 * 24 * time.Hour // 7天
	c.AnomalyDetection.AnomalyThreshold = 0.8
	c.AnomalyDetection.BehavioralAnalysis = true
	c.AnomalyDetection.PatternRecognition = true
	c.AnomalyDetection.MachineLearningEnabled = false
	c.AnomalyDetection.MLModelPath = "models/anomaly_detection.model"
	c.AnomalyDetection.MLTrainingDataPath = "data/training/"
	c.AnomalyDetection.RealTimeAnalysis = true
	c.AnomalyDetection.BatchAnalysis = true
	c.AnomalyDetection.AnalysisInterval = 5 * time.Minute
	c.AnomalyDetection.AlertOnAnomaly = true
	c.AnomalyDetection.AutoBlockOnAnomaly = false
	c.AnomalyDetection.AnomalyScoreThreshold = 0.7

	// 安全审计配置默认值
	c.SecurityAudit.Enabled = true
	c.SecurityAudit.AuditLevel = "medium"
	c.SecurityAudit.AuditEvents = []string{"login", "logout", "password_change", "permission_change", "data_access", "admin_action"}
	c.SecurityAudit.DataRetention = 365 * 24 * time.Hour // 1年
	c.SecurityAudit.EncryptionEnabled = true
	c.SecurityAudit.EncryptionKey = ""
	c.SecurityAudit.CompressionEnabled = true
	c.SecurityAudit.RealTimeMonitoring = true
	c.SecurityAudit.AlertOnSuspiciousActivity = true
	c.SecurityAudit.ComplianceReporting = true
	c.SecurityAudit.ReportGeneration = true
	c.SecurityAudit.ReportSchedule = "weekly"
	c.SecurityAudit.DataExportEnabled = true
	c.SecurityAudit.DataExportFormat = "json"

	// 威胁防护配置默认值
	c.ThreatProtection.Enabled = true
	c.ThreatProtection.ThreatIntelligence = true
	c.ThreatProtection.TIUpdateInterval = 24 * time.Hour
	c.ThreatProtection.TISourceURLs = []string{"https://api.abuseipdb.com/api/v2/blacklist", "https://api.blocklist.de/get.php"}
	c.ThreatProtection.MalwareScanning = true
	c.ThreatProtection.ScanInterval = 1 * time.Hour
	c.ThreatProtection.VirusTotalAPIKey = ""
	c.ThreatProtection.PhishingProtection = true
	c.ThreatProtection.PhishingURLsFile = "config/phishing_urls.txt"
	c.ThreatProtection.SQLInjectionProtection = true
	c.ThreatProtection.XSSProtection = true
	c.ThreatProtection.CSRFProtection = true
	c.ThreatProtection.CSRFTokenExpiry = 30 * time.Minute
	c.ThreatProtection.FileUploadScanning = true
	c.ThreatProtection.AllowedFileTypes = []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".txt"}
	c.ThreatProtection.MaxFileSize = 10 * 1024 * 1024 // 10MB
	c.ThreatProtection.BlockedFileTypes = []string{".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js"}
	c.ThreatProtection.ContentSecurityPolicy = true
	c.ThreatProtection.CSPDirectives = "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';"
}

// BindEnvs 绑定环境变量
func (c *SecurityConfig) BindEnvs() {
	// 基础安全配置
	viper.BindEnv("security.base_security.enabled", "SECURITY_ENABLED")
	viper.BindEnv("security.base_security.session_timeout", "SECURITY_SESSION_TIMEOUT")
	viper.BindEnv("security.base_security.max_login_attempts", "SECURITY_MAX_LOGIN_ATTEMPTS")
	viper.BindEnv("security.base_security.login_lockout_duration", "SECURITY_LOGIN_LOCKOUT_DURATION")
	viper.BindEnv("security.base_security.password_history_count", "SECURITY_PASSWORD_HISTORY_COUNT")
	viper.BindEnv("security.base_security.force_password_change", "SECURITY_FORCE_PASSWORD_CHANGE")
	viper.BindEnv("security.base_security.password_change_interval", "SECURITY_PASSWORD_CHANGE_INTERVAL")
	viper.BindEnv("security.base_security.account_lockout_threshold", "SECURITY_ACCOUNT_LOCKOUT_THRESHOLD")
	viper.BindEnv("security.base_security.account_lockout_duration", "SECURITY_ACCOUNT_LOCKOUT_DURATION")
	viper.BindEnv("security.base_security.inactive_account_timeout", "SECURITY_INACTIVE_ACCOUNT_TIMEOUT")
	viper.BindEnv("security.base_security.concurrent_session_limit", "SECURITY_CONCURRENT_SESSION_LIMIT")
	viper.BindEnv("security.base_security.ip_whitelist_enabled", "SECURITY_IP_WHITELIST_ENABLED")
	viper.BindEnv("security.base_security.ip_blacklist_enabled", "SECURITY_IP_BLACKLIST_ENABLED")
	viper.BindEnv("security.base_security.rate_limit_enabled", "SECURITY_RATE_LIMIT_ENABLED")
	viper.BindEnv("security.base_security.rate_limit_requests", "SECURITY_RATE_LIMIT_REQUESTS")
	viper.BindEnv("security.base_security.rate_limit_window", "SECURITY_RATE_LIMIT_WINDOW")

	// 密码策略配置
	viper.BindEnv("security.password_policy.min_length", "SECURITY_PASSWORD_MIN_LENGTH")
	viper.BindEnv("security.password_policy.max_length", "SECURITY_PASSWORD_MAX_LENGTH")
	viper.BindEnv("security.password_policy.require_uppercase", "SECURITY_PASSWORD_REQUIRE_UPPERCASE")
	viper.BindEnv("security.password_policy.require_lowercase", "SECURITY_PASSWORD_REQUIRE_LOWERCASE")
	viper.BindEnv("security.password_policy.require_numbers", "SECURITY_PASSWORD_REQUIRE_NUMBERS")
	viper.BindEnv("security.password_policy.require_special_chars", "SECURITY_PASSWORD_REQUIRE_SPECIAL_CHARS")
	viper.BindEnv("security.password_policy.special_chars_list", "SECURITY_PASSWORD_SPECIAL_CHARS_LIST")
	viper.BindEnv("security.password_policy.prevent_common_passwords", "SECURITY_PASSWORD_PREVENT_COMMON")
	viper.BindEnv("security.password_policy.common_passwords_file", "SECURITY_PASSWORD_COMMON_FILE")
	viper.BindEnv("security.password_policy.prevent_username_in_password", "SECURITY_PASSWORD_PREVENT_USERNAME")
	viper.BindEnv("security.password_policy.prevent_sequential_chars", "SECURITY_PASSWORD_PREVENT_SEQUENTIAL")
	viper.BindEnv("security.password_policy.prevent_repeated_chars", "SECURITY_PASSWORD_PREVENT_REPEATED")
	viper.BindEnv("security.password_policy.max_repeated_chars", "SECURITY_PASSWORD_MAX_REPEATED_CHARS")
	viper.BindEnv("security.password_policy.password_strength_threshold", "SECURITY_PASSWORD_STRENGTH_THRESHOLD")

	// 访问控制配置
	viper.BindEnv("security.access_control.rbac_enabled", "SECURITY_RBAC_ENABLED")
	viper.BindEnv("security.access_control.permission_cache_enabled", "SECURITY_PERMISSION_CACHE_ENABLED")
	viper.BindEnv("security.access_control.permission_cache_ttl", "SECURITY_PERMISSION_CACHE_TTL")
	viper.BindEnv("security.access_control.default_deny_policy", "SECURITY_DEFAULT_DENY_POLICY")
	viper.BindEnv("security.access_control.resource_level_access", "SECURITY_RESOURCE_LEVEL_ACCESS")
	viper.BindEnv("security.access_control.time_based_access", "SECURITY_TIME_BASED_ACCESS")
	viper.BindEnv("security.access_control.location_based_access", "SECURITY_LOCATION_BASED_ACCESS")
	viper.BindEnv("security.access_control.device_based_access", "SECURITY_DEVICE_BASED_ACCESS")
	viper.BindEnv("security.access_control.session_based_access", "SECURITY_SESSION_BASED_ACCESS")
	viper.BindEnv("security.access_control.api_key_permissions", "SECURITY_API_KEY_PERMISSIONS")
	viper.BindEnv("security.access_control.jwt_claims_validation", "SECURITY_JWT_CLAIMS_VALIDATION")
	viper.BindEnv("security.access_control.token_refresh_enabled", "SECURITY_TOKEN_REFRESH_ENABLED")
	viper.BindEnv("security.access_control.token_refresh_threshold", "SECURITY_TOKEN_REFRESH_THRESHOLD")

	// 异常检测配置
	viper.BindEnv("security.anomaly_detection.enabled", "SECURITY_ANOMALY_DETECTION_ENABLED")
	viper.BindEnv("security.anomaly_detection.learning_mode", "SECURITY_ANOMALY_LEARNING_MODE")
	viper.BindEnv("security.anomaly_detection.learning_period", "SECURITY_ANOMALY_LEARNING_PERIOD")
	viper.BindEnv("security.anomaly_detection.anomaly_threshold", "SECURITY_ANOMALY_THRESHOLD")
	viper.BindEnv("security.anomaly_detection.behavioral_analysis", "SECURITY_ANOMALY_BEHAVIORAL_ANALYSIS")
	viper.BindEnv("security.anomaly_detection.pattern_recognition", "SECURITY_ANOMALY_PATTERN_RECOGNITION")
	viper.BindEnv("security.anomaly_detection.machine_learning_enabled", "SECURITY_ANOMALY_ML_ENABLED")
	viper.BindEnv("security.anomaly_detection.ml_model_path", "SECURITY_ANOMALY_ML_MODEL_PATH")
	viper.BindEnv("security.anomaly_detection.ml_training_data_path", "SECURITY_ANOMALY_ML_TRAINING_DATA_PATH")
	viper.BindEnv("security.anomaly_detection.real_time_analysis", "SECURITY_ANOMALY_REAL_TIME_ANALYSIS")
	viper.BindEnv("security.anomaly_detection.batch_analysis", "SECURITY_ANOMALY_BATCH_ANALYSIS")
	viper.BindEnv("security.anomaly_detection.analysis_interval", "SECURITY_ANOMALY_ANALYSIS_INTERVAL")
	viper.BindEnv("security.anomaly_detection.alert_on_anomaly", "SECURITY_ANOMALY_ALERT_ON_ANOMALY")
	viper.BindEnv("security.anomaly_detection.auto_block_on_anomaly", "SECURITY_ANOMALY_AUTO_BLOCK")
	viper.BindEnv("security.anomaly_detection.anomaly_score_threshold", "SECURITY_ANOMALY_SCORE_THRESHOLD")

	// 安全审计配置
	viper.BindEnv("security.security_audit.enabled", "SECURITY_AUDIT_ENABLED")
	viper.BindEnv("security.security_audit.audit_level", "SECURITY_AUDIT_LEVEL")
	viper.BindEnv("security.security_audit.audit_events", "SECURITY_AUDIT_EVENTS")
	viper.BindEnv("security.security_audit.data_retention", "SECURITY_AUDIT_DATA_RETENTION")
	viper.BindEnv("security.security_audit.encryption_enabled", "SECURITY_AUDIT_ENCRYPTION_ENABLED")
	viper.BindEnv("security.security_audit.encryption_key", "SECURITY_AUDIT_ENCRYPTION_KEY")
	viper.BindEnv("security.security_audit.compression_enabled", "SECURITY_AUDIT_COMPRESSION_ENABLED")
	viper.BindEnv("security.security_audit.real_time_monitoring", "SECURITY_AUDIT_REAL_TIME_MONITORING")
	viper.BindEnv("security.security_audit.alert_on_suspicious_activity", "SECURITY_AUDIT_ALERT_ON_SUSPICIOUS")
	viper.BindEnv("security.security_audit.compliance_reporting", "SECURITY_AUDIT_COMPLIANCE_REPORTING")
	viper.BindEnv("security.security_audit.report_generation", "SECURITY_AUDIT_REPORT_GENERATION")
	viper.BindEnv("security.security_audit.report_schedule", "SECURITY_AUDIT_REPORT_SCHEDULE")
	viper.BindEnv("security.security_audit.data_export_enabled", "SECURITY_AUDIT_DATA_EXPORT_ENABLED")
	viper.BindEnv("security.security_audit.data_export_format", "SECURITY_AUDIT_DATA_EXPORT_FORMAT")

	// 威胁防护配置
	viper.BindEnv("security.threat_protection.enabled", "SECURITY_THREAT_PROTECTION_ENABLED")
	viper.BindEnv("security.threat_protection.threat_intelligence", "SECURITY_THREAT_INTELLIGENCE")
	viper.BindEnv("security.threat_protection.ti_update_interval", "SECURITY_THREAT_TI_UPDATE_INTERVAL")
	viper.BindEnv("security.threat_protection.ti_source_urls", "SECURITY_THREAT_TI_SOURCE_URLS")
	viper.BindEnv("security.threat_protection.malware_scanning", "SECURITY_THREAT_MALWARE_SCANNING")
	viper.BindEnv("security.threat_protection.scan_interval", "SECURITY_THREAT_SCAN_INTERVAL")
	viper.BindEnv("security.threat_protection.virus_total_api_key", "SECURITY_THREAT_VIRUS_TOTAL_API_KEY")
	viper.BindEnv("security.threat_protection.phishing_protection", "SECURITY_THREAT_PHISHING_PROTECTION")
	viper.BindEnv("security.threat_protection.phishing_urls_file", "SECURITY_THREAT_PHISHING_URLS_FILE")
	viper.BindEnv("security.threat_protection.sql_injection_protection", "SECURITY_THREAT_SQL_INJECTION_PROTECTION")
	viper.BindEnv("security.threat_protection.xss_protection", "SECURITY_THREAT_XSS_PROTECTION")
	viper.BindEnv("security.threat_protection.csrf_protection", "SECURITY_THREAT_CSRF_PROTECTION")
	viper.BindEnv("security.threat_protection.csrf_token_expiry", "SECURITY_THREAT_CSRF_TOKEN_EXPIRY")
	viper.BindEnv("security.threat_protection.file_upload_scanning", "SECURITY_THREAT_FILE_UPLOAD_SCANNING")
	viper.BindEnv("security.threat_protection.allowed_file_types", "SECURITY_THREAT_ALLOWED_FILE_TYPES")
	viper.BindEnv("security.threat_protection.max_file_size", "SECURITY_THREAT_MAX_FILE_SIZE")
	viper.BindEnv("security.threat_protection.blocked_file_types", "SECURITY_THREAT_BLOCKED_FILE_TYPES")
	viper.BindEnv("security.threat_protection.content_security_policy", "SECURITY_THREAT_CONTENT_SECURITY_POLICY")
	viper.BindEnv("security.threat_protection.csp_directives", "SECURITY_THREAT_CSP_DIRECTIVES")
}

// Validate 验证配置
func (c *SecurityConfig) Validate() error {
	// 基础安全配置验证
	if c.BaseSecurity.MaxLoginAttempts <= 0 {
		return fmt.Errorf("max_login_attempts must be greater than 0")
	}
	if c.BaseSecurity.PasswordHistoryCount < 0 {
		return fmt.Errorf("password_history_count must be non-negative")
	}
	if c.BaseSecurity.AccountLockoutThreshold <= 0 {
		return fmt.Errorf("account_lockout_threshold must be greater than 0")
	}
	if c.BaseSecurity.ConcurrentSessionLimit <= 0 {
		return fmt.Errorf("concurrent_session_limit must be greater than 0")
	}
	if c.BaseSecurity.RateLimitRequests <= 0 {
		return fmt.Errorf("rate_limit_requests must be greater than 0")
	}

	// 密码策略配置验证
	if c.PasswordPolicy.MinLength < 6 {
		return fmt.Errorf("password min_length must be at least 6")
	}
	if c.PasswordPolicy.MaxLength < c.PasswordPolicy.MinLength {
		return fmt.Errorf("password max_length must be greater than min_length")
	}
	if c.PasswordPolicy.MaxRepeatedChars < 1 {
		return fmt.Errorf("max_repeated_chars must be at least 1")
	}
	if c.PasswordPolicy.PasswordStrengthThreshold < 0 || c.PasswordPolicy.PasswordStrengthThreshold > 100 {
		return fmt.Errorf("password_strength_threshold must be between 0 and 100")
	}

	// 异常检测配置验证
	if c.AnomalyDetection.AnomalyThreshold < 0 || c.AnomalyDetection.AnomalyThreshold > 1 {
		return fmt.Errorf("anomaly_threshold must be between 0 and 1")
	}
	if c.AnomalyDetection.AnomalyScoreThreshold < 0 || c.AnomalyDetection.AnomalyScoreThreshold > 1 {
		return fmt.Errorf("anomaly_score_threshold must be between 0 and 1")
	}

	// 威胁防护配置验证
	if c.ThreatProtection.MaxFileSize <= 0 {
		return fmt.Errorf("max_file_size must be greater than 0")
	}

	return nil
}
