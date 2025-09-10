package Config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// MonitoringConfig 监控告警系统配置
type MonitoringConfig struct {
	// 基础配置
	BaseConfig struct {
		Enabled           bool          `mapstructure:"enabled" json:"enabled"`
		CheckInterval     time.Duration `mapstructure:"check_interval" json:"check_interval"`
		RetentionPeriod   time.Duration `mapstructure:"retention_period" json:"retention_period"`
		MaxAlertsPerHour  int           `mapstructure:"max_alerts_per_hour" json:"max_alerts_per_hour"`
		AlertCooldown     time.Duration `mapstructure:"alert_cooldown" json:"alert_cooldown"`
		EnableDashboard   bool          `mapstructure:"enable_dashboard" json:"enable_dashboard"`
		DashboardPort     int           `mapstructure:"dashboard_port" json:"dashboard_port"`
		EnableMetrics     bool          `mapstructure:"enable_metrics" json:"enable_metrics"`
		MetricsPort       int           `mapstructure:"metrics_port" json:"metrics_port"`
	} `mapstructure:"base" json:"base"`

	// 系统监控配置
	SystemMonitoring struct {
		Enabled           bool          `mapstructure:"enabled" json:"enabled"`
		CheckInterval     time.Duration `mapstructure:"check_interval" json:"check_interval"`
		CPUThreshold      float64       `mapstructure:"cpu_threshold" json:"cpu_threshold"`
		MemoryThreshold   float64       `mapstructure:"memory_threshold" json:"memory_threshold"`
		DiskThreshold     float64       `mapstructure:"disk_threshold" json:"disk_threshold"`
		NetworkThreshold  float64       `mapstructure:"network_threshold" json:"network_threshold"`
		ProcessThreshold  int           `mapstructure:"process_threshold" json:"process_threshold"`
		LoadAverageThreshold float64    `mapstructure:"load_average_threshold" json:"load_average_threshold"`
	} `mapstructure:"system" json:"system"`

	// 应用监控配置
	ApplicationMonitoring struct {
		Enabled              bool          `mapstructure:"enabled" json:"enabled"`
		CheckInterval        time.Duration `mapstructure:"check_interval" json:"check_interval"`
		ResponseTimeThreshold time.Duration `mapstructure:"response_time_threshold" json:"response_time_threshold"`
		ErrorRateThreshold   float64       `mapstructure:"error_rate_threshold" json:"error_rate_threshold"`
		ThroughputThreshold  int           `mapstructure:"throughput_threshold" json:"throughput_threshold"`
		MemoryLeakThreshold  float64       `mapstructure:"memory_leak_threshold" json:"memory_leak_threshold"`
		GoroutineThreshold   int           `mapstructure:"goroutine_threshold" json:"goroutine_threshold"`
		GCThreshold          time.Duration `mapstructure:"gc_threshold" json:"gc_threshold"`
	} `mapstructure:"application" json:"application"`

	// 数据库监控配置
	DatabaseMonitoring struct {
		Enabled              bool          `mapstructure:"enabled" json:"enabled"`
		CheckInterval        time.Duration `mapstructure:"check_interval" json:"check_interval"`
		ConnectionThreshold  int           `mapstructure:"connection_threshold" json:"connection_threshold"`
		SlowQueryThreshold   time.Duration `mapstructure:"slow_query_threshold" json:"slow_query_threshold"`
		QueryTimeoutThreshold time.Duration `mapstructure:"query_timeout_threshold" json:"query_timeout_threshold"`
		DeadlockThreshold    int           `mapstructure:"deadlock_threshold" json:"deadlock_threshold"`
		LockWaitThreshold    time.Duration `mapstructure:"lock_wait_threshold" json:"lock_wait_threshold"`
		TableSizeThreshold   int64         `mapstructure:"table_size_threshold" json:"table_size_threshold"`
	} `mapstructure:"database" json:"database"`

	// 缓存监控配置
	CacheMonitoring struct {
		Enabled              bool          `mapstructure:"enabled" json:"enabled"`
		CheckInterval        time.Duration `mapstructure:"check_interval" json:"check_interval"`
		HitRateThreshold     float64       `mapstructure:"hit_rate_threshold" json:"hit_rate_threshold"`
		MemoryUsageThreshold float64       `mapstructure:"memory_usage_threshold" json:"memory_usage_threshold"`
		ConnectionThreshold  int           `mapstructure:"connection_threshold" json:"connection_threshold"`
		EvictionThreshold   int           `mapstructure:"eviction_threshold" json:"eviction_threshold"`
		ExpiredKeysThreshold int           `mapstructure:"expired_keys_threshold" json:"expired_keys_threshold"`
	} `mapstructure:"cache" json:"cache"`

	// 业务监控配置
	BusinessMonitoring struct {
		Enabled              bool          `mapstructure:"enabled" json:"enabled"`
		CheckInterval        time.Duration `mapstructure:"check_interval" json:"check_interval"`
		UserActivityThreshold int           `mapstructure:"user_activity_threshold" json:"user_activity_threshold"`
		APIUsageThreshold    int           `mapstructure:"api_usage_threshold" json:"api_usage_threshold"`
		ErrorLogThreshold    int           `mapstructure:"error_log_threshold" json:"error_log_threshold"`
		SecurityEventThreshold int         `mapstructure:"security_event_threshold" json:"security_event_threshold"`
		DataSyncThreshold    time.Duration `mapstructure:"data_sync_threshold" json:"data_sync_threshold"`
	} `mapstructure:"business" json:"business"`

	// 告警配置
	AlertConfig struct {
		Enabled              bool          `mapstructure:"enabled" json:"enabled"`
		DefaultSeverity      string        `mapstructure:"default_severity" json:"default_severity"`
		EscalationEnabled    bool          `mapstructure:"escalation_enabled" json:"escalation_enabled"`
		EscalationDelay      time.Duration `mapstructure:"escalation_delay" json:"escalation_delay"`
		MaxEscalationLevel   int           `mapstructure:"max_escalation_level" json:"max_escalation_level"`
		AutoResolveEnabled   bool          `mapstructure:"auto_resolve_enabled" json:"auto_resolve_enabled"`
		AutoResolveDelay     time.Duration `mapstructure:"auto_resolve_delay" json:"auto_resolve_delay"`
		SuppressionEnabled   bool          `mapstructure:"suppression_enabled" json:"suppression_enabled"`
		SuppressionWindow    time.Duration `mapstructure:"suppression_window" json:"suppression_window"`
	} `mapstructure:"alert" json:"alert"`

	// 通知配置
	NotificationConfig struct {
		Email struct {
			Enabled     bool   `mapstructure:"enabled" json:"enabled"`
			SMTPHost    string `mapstructure:"smtp_host" json:"smtp_host"`
			SMTPPort    int    `mapstructure:"smtp_port" json:"smtp_port"`
			Username    string `mapstructure:"username" json:"username"`
			Password    string `mapstructure:"password" json:"password"`
			FromAddress string `mapstructure:"from_address" json:"from_address"`
			ToAddresses string `mapstructure:"to_addresses" json:"to_addresses"`
			Subject     string `mapstructure:"subject" json:"subject"`
		} `mapstructure:"email" json:"email"`

		Webhook struct {
			Enabled     bool   `mapstructure:"enabled" json:"enabled"`
			URL         string `mapstructure:"url" json:"url"`
			Method      string `mapstructure:"method" json:"method"`
			Headers     string `mapstructure:"headers" json:"headers"`
			Timeout     time.Duration `mapstructure:"timeout" json:"timeout"`
			RetryCount  int    `mapstructure:"retry_count" json:"retry_count"`
		} `mapstructure:"webhook" json:"webhook"`

		Slack struct {
			Enabled     bool   `mapstructure:"enabled" json:"enabled"`
			WebhookURL  string `mapstructure:"webhook_url" json:"webhook_url"`
			Channel     string `mapstructure:"channel" json:"channel"`
			Username    string `mapstructure:"username" json:"username"`
			IconEmoji   string `mapstructure:"icon_emoji" json:"icon_emoji"`
		} `mapstructure:"slack" json:"slack"`

		DingTalk struct {
			Enabled     bool   `mapstructure:"enabled" json:"enabled"`
			WebhookURL  string `mapstructure:"webhook_url" json:"webhook_url"`
			Secret      string `mapstructure:"secret" json:"secret"`
			AtMobiles   string `mapstructure:"at_mobiles" json:"at_mobiles"`
		} `mapstructure:"dingtalk" json:"dingtalk"`

		SMS struct {
			Enabled     bool   `mapstructure:"enabled" json:"enabled"`
			Provider    string `mapstructure:"provider" json:"provider"`
			APIKey      string `mapstructure:"api_key" json:"api_key"`
			APISecret   string `mapstructure:"api_secret" json:"api_secret"`
			PhoneNumbers string `mapstructure:"phone_numbers" json:"phone_numbers"`
		} `mapstructure:"sms" json:"sms"`
	} `mapstructure:"notification" json:"notification"`

	// 存储配置
	StorageConfig struct {
		Type              string        `mapstructure:"type" json:"type"`
		Database struct {
			Enabled     bool   `mapstructure:"enabled" json:"enabled"`
			TablePrefix string `mapstructure:"table_prefix" json:"table_prefix"`
			Retention   time.Duration `mapstructure:"retention" json:"retention"`
		} `mapstructure:"database" json:"database"`
		
		File struct {
			Enabled     bool   `mapstructure:"enabled" json:"enabled"`
			Path        string `mapstructure:"path" json:"path"`
			Format      string `mapstructure:"format" json:"format"`
			MaxSize     int64  `mapstructure:"max_size" json:"max_size"`
			MaxAge      time.Duration `mapstructure:"max_age" json:"max_age"`
		} `mapstructure:"file" json:"file"`
		
		Redis struct {
			Enabled     bool   `mapstructure:"enabled" json:"enabled"`
			KeyPrefix   string `mapstructure:"key_prefix" json:"key_prefix"`
			TTL         time.Duration `mapstructure:"ttl" json:"ttl"`
		} `mapstructure:"redis" json:"redis"`
	} `mapstructure:"storage" json:"storage"`
}

// SetDefaults 设置默认值
func (c *MonitoringConfig) SetDefaults() {
	// 基础配置默认值
	c.BaseConfig.Enabled = true
	c.BaseConfig.CheckInterval = 30 * time.Second
	c.BaseConfig.RetentionPeriod = 30 * 24 * time.Hour // 30天
	c.BaseConfig.MaxAlertsPerHour = 100
	c.BaseConfig.AlertCooldown = 5 * time.Minute
	c.BaseConfig.EnableDashboard = true
	c.BaseConfig.DashboardPort = 8081
	c.BaseConfig.EnableMetrics = true
	c.BaseConfig.MetricsPort = 8082

	// 系统监控默认值
	c.SystemMonitoring.Enabled = true
	c.SystemMonitoring.CheckInterval = 60 * time.Second
	c.SystemMonitoring.CPUThreshold = 80.0
	c.SystemMonitoring.MemoryThreshold = 85.0
	c.SystemMonitoring.DiskThreshold = 90.0
	c.SystemMonitoring.NetworkThreshold = 1000.0 // MB/s
	c.SystemMonitoring.ProcessThreshold = 1000
	c.SystemMonitoring.LoadAverageThreshold = 5.0

	// 应用监控默认值
	c.ApplicationMonitoring.Enabled = true
	c.ApplicationMonitoring.CheckInterval = 30 * time.Second
	c.ApplicationMonitoring.ResponseTimeThreshold = 2 * time.Second
	c.ApplicationMonitoring.ErrorRateThreshold = 5.0
	c.ApplicationMonitoring.ThroughputThreshold = 1000
	c.ApplicationMonitoring.MemoryLeakThreshold = 10.0
	c.ApplicationMonitoring.GoroutineThreshold = 10000
	c.ApplicationMonitoring.GCThreshold = 100 * time.Millisecond

	// 数据库监控默认值
	c.DatabaseMonitoring.Enabled = true
	c.DatabaseMonitoring.CheckInterval = 60 * time.Second
	c.DatabaseMonitoring.ConnectionThreshold = 100
	c.DatabaseMonitoring.SlowQueryThreshold = 1 * time.Second
	c.DatabaseMonitoring.QueryTimeoutThreshold = 30 * time.Second
	c.DatabaseMonitoring.DeadlockThreshold = 5
	c.DatabaseMonitoring.LockWaitThreshold = 10 * time.Second
	c.DatabaseMonitoring.TableSizeThreshold = 1024 * 1024 * 1024 // 1GB

	// 缓存监控默认值
	c.CacheMonitoring.Enabled = true
	c.CacheMonitoring.CheckInterval = 60 * time.Second
	c.CacheMonitoring.HitRateThreshold = 80.0
	c.CacheMonitoring.MemoryUsageThreshold = 85.0
	c.CacheMonitoring.ConnectionThreshold = 100
	c.CacheMonitoring.EvictionThreshold = 1000
	c.CacheMonitoring.ExpiredKeysThreshold = 10000

	// 业务监控默认值
	c.BusinessMonitoring.Enabled = true
	c.BusinessMonitoring.CheckInterval = 5 * time.Minute
	c.BusinessMonitoring.UserActivityThreshold = 100
	c.BusinessMonitoring.APIUsageThreshold = 1000
	c.BusinessMonitoring.ErrorLogThreshold = 50
	c.BusinessMonitoring.SecurityEventThreshold = 10
	c.BusinessMonitoring.DataSyncThreshold = 10 * time.Minute

	// 告警配置默认值
	c.AlertConfig.Enabled = true
	c.AlertConfig.DefaultSeverity = "warning"
	c.AlertConfig.EscalationEnabled = true
	c.AlertConfig.EscalationDelay = 10 * time.Minute
	c.AlertConfig.MaxEscalationLevel = 3
	c.AlertConfig.AutoResolveEnabled = true
	c.AlertConfig.AutoResolveDelay = 30 * time.Minute
	c.AlertConfig.SuppressionEnabled = true
	c.AlertConfig.SuppressionWindow = 1 * time.Hour

	// 通知配置默认值
	c.NotificationConfig.Email.Enabled = false
	c.NotificationConfig.Email.SMTPPort = 587
	c.NotificationConfig.Email.Subject = "[监控告警] Cloud Platform"

	c.NotificationConfig.Webhook.Enabled = false
	c.NotificationConfig.Webhook.Method = "POST"
	c.NotificationConfig.Webhook.Timeout = 10 * time.Second
	c.NotificationConfig.Webhook.RetryCount = 3

	c.NotificationConfig.Slack.Enabled = false
	c.NotificationConfig.Slack.Username = "监控告警"
	c.NotificationConfig.Slack.IconEmoji = ":warning:"

	c.NotificationConfig.DingTalk.Enabled = false

	c.NotificationConfig.SMS.Enabled = false

	// 存储配置默认值
	c.StorageConfig.Type = "database"
	c.StorageConfig.Database.Enabled = true
	c.StorageConfig.Database.TablePrefix = "monitoring_"
	c.StorageConfig.Database.Retention = 90 * 24 * time.Hour // 90天

	c.StorageConfig.File.Enabled = false
	c.StorageConfig.File.Path = "logs/monitoring"
	c.StorageConfig.File.Format = "json"
	c.StorageConfig.File.MaxSize = 100 * 1024 * 1024 // 100MB
	c.StorageConfig.File.MaxAge = 30 * 24 * time.Hour

	c.StorageConfig.Redis.Enabled = false
	c.StorageConfig.Redis.KeyPrefix = "monitoring:"
	c.StorageConfig.Redis.TTL = 24 * time.Hour
}

// BindEnvs 绑定环境变量
func (c *MonitoringConfig) BindEnvs() {
	viper.SetDefault("MONITORING_ENABLED", c.BaseConfig.Enabled)
	viper.SetDefault("MONITORING_CHECK_INTERVAL", c.BaseConfig.CheckInterval)
	viper.SetDefault("MONITORING_RETENTION_PERIOD", c.BaseConfig.RetentionPeriod)
	viper.SetDefault("MONITORING_MAX_ALERTS_PER_HOUR", c.BaseConfig.MaxAlertsPerHour)
	viper.SetDefault("MONITORING_ALERT_COOLDOWN", c.BaseConfig.AlertCooldown)
	viper.SetDefault("MONITORING_ENABLE_DASHBOARD", c.BaseConfig.EnableDashboard)
	viper.SetDefault("MONITORING_DASHBOARD_PORT", c.BaseConfig.DashboardPort)
	viper.SetDefault("MONITORING_ENABLE_METRICS", c.BaseConfig.EnableMetrics)
	viper.SetDefault("MONITORING_METRICS_PORT", c.BaseConfig.MetricsPort)

	// 系统监控环境变量
	viper.SetDefault("MONITORING_SYSTEM_ENABLED", c.SystemMonitoring.Enabled)
	viper.SetDefault("MONITORING_SYSTEM_CHECK_INTERVAL", c.SystemMonitoring.CheckInterval)
	viper.SetDefault("MONITORING_SYSTEM_CPU_THRESHOLD", c.SystemMonitoring.CPUThreshold)
	viper.SetDefault("MONITORING_SYSTEM_MEMORY_THRESHOLD", c.SystemMonitoring.MemoryThreshold)
	viper.SetDefault("MONITORING_SYSTEM_DISK_THRESHOLD", c.SystemMonitoring.DiskThreshold)
	viper.SetDefault("MONITORING_SYSTEM_NETWORK_THRESHOLD", c.SystemMonitoring.NetworkThreshold)
	viper.SetDefault("MONITORING_SYSTEM_PROCESS_THRESHOLD", c.SystemMonitoring.ProcessThreshold)
	viper.SetDefault("MONITORING_SYSTEM_LOAD_AVERAGE_THRESHOLD", c.SystemMonitoring.LoadAverageThreshold)

	// 应用监控环境变量
	viper.SetDefault("MONITORING_APP_ENABLED", c.ApplicationMonitoring.Enabled)
	viper.SetDefault("MONITORING_APP_CHECK_INTERVAL", c.ApplicationMonitoring.CheckInterval)
	viper.SetDefault("MONITORING_APP_RESPONSE_TIME_THRESHOLD", c.ApplicationMonitoring.ResponseTimeThreshold)
	viper.SetDefault("MONITORING_APP_ERROR_RATE_THRESHOLD", c.ApplicationMonitoring.ErrorRateThreshold)
	viper.SetDefault("MONITORING_APP_THROUGHPUT_THRESHOLD", c.ApplicationMonitoring.ThroughputThreshold)
	viper.SetDefault("MONITORING_APP_MEMORY_LEAK_THRESHOLD", c.ApplicationMonitoring.MemoryLeakThreshold)
	viper.SetDefault("MONITORING_APP_GOROUTINE_THRESHOLD", c.ApplicationMonitoring.GoroutineThreshold)
	viper.SetDefault("MONITORING_APP_GC_THRESHOLD", c.ApplicationMonitoring.GCThreshold)

	// 数据库监控环境变量
	viper.SetDefault("MONITORING_DB_ENABLED", c.DatabaseMonitoring.Enabled)
	viper.SetDefault("MONITORING_DB_CHECK_INTERVAL", c.DatabaseMonitoring.CheckInterval)
	viper.SetDefault("MONITORING_DB_CONNECTION_THRESHOLD", c.DatabaseMonitoring.ConnectionThreshold)
	viper.SetDefault("MONITORING_DB_SLOW_QUERY_THRESHOLD", c.DatabaseMonitoring.SlowQueryThreshold)
	viper.SetDefault("MONITORING_DB_QUERY_TIMEOUT_THRESHOLD", c.DatabaseMonitoring.QueryTimeoutThreshold)
	viper.SetDefault("MONITORING_DB_DEADLOCK_THRESHOLD", c.DatabaseMonitoring.DeadlockThreshold)
	viper.SetDefault("MONITORING_DB_LOCK_WAIT_THRESHOLD", c.DatabaseMonitoring.LockWaitThreshold)
	viper.SetDefault("MONITORING_DB_TABLE_SIZE_THRESHOLD", c.DatabaseMonitoring.TableSizeThreshold)

	// 缓存监控环境变量
	viper.SetDefault("MONITORING_CACHE_ENABLED", c.CacheMonitoring.Enabled)
	viper.SetDefault("MONITORING_CACHE_CHECK_INTERVAL", c.CacheMonitoring.CheckInterval)
	viper.SetDefault("MONITORING_CACHE_HIT_RATE_THRESHOLD", c.CacheMonitoring.HitRateThreshold)
	viper.SetDefault("MONITORING_CACHE_MEMORY_USAGE_THRESHOLD", c.CacheMonitoring.MemoryUsageThreshold)
	viper.SetDefault("MONITORING_CACHE_CONNECTION_THRESHOLD", c.CacheMonitoring.ConnectionThreshold)
	viper.SetDefault("MONITORING_CACHE_EVICTION_THRESHOLD", c.CacheMonitoring.EvictionThreshold)
	viper.SetDefault("MONITORING_CACHE_EXPIRED_KEYS_THRESHOLD", c.CacheMonitoring.ExpiredKeysThreshold)

	// 业务监控环境变量
	viper.SetDefault("MONITORING_BUSINESS_ENABLED", c.BusinessMonitoring.Enabled)
	viper.SetDefault("MONITORING_BUSINESS_CHECK_INTERVAL", c.BusinessMonitoring.CheckInterval)
	viper.SetDefault("MONITORING_BUSINESS_USER_ACTIVITY_THRESHOLD", c.BusinessMonitoring.UserActivityThreshold)
	viper.SetDefault("MONITORING_BUSINESS_API_USAGE_THRESHOLD", c.BusinessMonitoring.APIUsageThreshold)
	viper.SetDefault("MONITORING_BUSINESS_ERROR_LOG_THRESHOLD", c.BusinessMonitoring.ErrorLogThreshold)
	viper.SetDefault("MONITORING_BUSINESS_SECURITY_EVENT_THRESHOLD", c.BusinessMonitoring.SecurityEventThreshold)
	viper.SetDefault("MONITORING_BUSINESS_DATA_SYNC_THRESHOLD", c.BusinessMonitoring.DataSyncThreshold)

	// 告警配置环境变量
	viper.SetDefault("MONITORING_ALERT_ENABLED", c.AlertConfig.Enabled)
	viper.SetDefault("MONITORING_ALERT_DEFAULT_SEVERITY", c.AlertConfig.DefaultSeverity)
	viper.SetDefault("MONITORING_ALERT_ESCALATION_ENABLED", c.AlertConfig.EscalationEnabled)
	viper.SetDefault("MONITORING_ALERT_ESCALATION_DELAY", c.AlertConfig.EscalationDelay)
	viper.SetDefault("MONITORING_ALERT_MAX_ESCALATION_LEVEL", c.AlertConfig.MaxEscalationLevel)
	viper.SetDefault("MONITORING_ALERT_AUTO_RESOLVE_ENABLED", c.AlertConfig.AutoResolveEnabled)
	viper.SetDefault("MONITORING_ALERT_AUTO_RESOLVE_DELAY", c.AlertConfig.AutoResolveDelay)
	viper.SetDefault("MONITORING_ALERT_SUPPRESSION_ENABLED", c.AlertConfig.SuppressionEnabled)
	viper.SetDefault("MONITORING_ALERT_SUPPRESSION_WINDOW", c.AlertConfig.SuppressionWindow)

	// 通知配置环境变量
	viper.SetDefault("MONITORING_NOTIFICATION_EMAIL_ENABLED", c.NotificationConfig.Email.Enabled)
	viper.SetDefault("MONITORING_NOTIFICATION_EMAIL_SMTP_HOST", c.NotificationConfig.Email.SMTPHost)
	viper.SetDefault("MONITORING_NOTIFICATION_EMAIL_SMTP_PORT", c.NotificationConfig.Email.SMTPPort)
	viper.SetDefault("MONITORING_NOTIFICATION_EMAIL_USERNAME", c.NotificationConfig.Email.Username)
	viper.SetDefault("MONITORING_NOTIFICATION_EMAIL_PASSWORD", c.NotificationConfig.Email.Password)
	viper.SetDefault("MONITORING_NOTIFICATION_EMAIL_FROM_ADDRESS", c.NotificationConfig.Email.FromAddress)
	viper.SetDefault("MONITORING_NOTIFICATION_EMAIL_TO_ADDRESSES", c.NotificationConfig.Email.ToAddresses)
	viper.SetDefault("MONITORING_NOTIFICATION_EMAIL_SUBJECT", c.NotificationConfig.Email.Subject)

	viper.SetDefault("MONITORING_NOTIFICATION_WEBHOOK_ENABLED", c.NotificationConfig.Webhook.Enabled)
	viper.SetDefault("MONITORING_NOTIFICATION_WEBHOOK_URL", c.NotificationConfig.Webhook.URL)
	viper.SetDefault("MONITORING_NOTIFICATION_WEBHOOK_METHOD", c.NotificationConfig.Webhook.Method)
	viper.SetDefault("MONITORING_NOTIFICATION_WEBHOOK_HEADERS", c.NotificationConfig.Webhook.Headers)
	viper.SetDefault("MONITORING_NOTIFICATION_WEBHOOK_TIMEOUT", c.NotificationConfig.Webhook.Timeout)
	viper.SetDefault("MONITORING_NOTIFICATION_WEBHOOK_RETRY_COUNT", c.NotificationConfig.Webhook.RetryCount)

	viper.SetDefault("MONITORING_NOTIFICATION_SLACK_ENABLED", c.NotificationConfig.Slack.Enabled)
	viper.SetDefault("MONITORING_NOTIFICATION_SLACK_WEBHOOK_URL", c.NotificationConfig.Slack.WebhookURL)
	viper.SetDefault("MONITORING_NOTIFICATION_SLACK_CHANNEL", c.NotificationConfig.Slack.Channel)
	viper.SetDefault("MONITORING_NOTIFICATION_SLACK_USERNAME", c.NotificationConfig.Slack.Username)
	viper.SetDefault("MONITORING_NOTIFICATION_SLACK_ICON_EMOJI", c.NotificationConfig.Slack.IconEmoji)

	viper.SetDefault("MONITORING_NOTIFICATION_DINGTALK_ENABLED", c.NotificationConfig.DingTalk.Enabled)
	viper.SetDefault("MONITORING_NOTIFICATION_DINGTALK_WEBHOOK_URL", c.NotificationConfig.DingTalk.WebhookURL)
	viper.SetDefault("MONITORING_NOTIFICATION_DINGTALK_SECRET", c.NotificationConfig.DingTalk.Secret)
	viper.SetDefault("MONITORING_NOTIFICATION_DINGTALK_AT_MOBILES", c.NotificationConfig.DingTalk.AtMobiles)

	viper.SetDefault("MONITORING_NOTIFICATION_SMS_ENABLED", c.NotificationConfig.SMS.Enabled)
	viper.SetDefault("MONITORING_NOTIFICATION_SMS_PROVIDER", c.NotificationConfig.SMS.Provider)
	viper.SetDefault("MONITORING_NOTIFICATION_SMS_API_KEY", c.NotificationConfig.SMS.APIKey)
	viper.SetDefault("MONITORING_NOTIFICATION_SMS_API_SECRET", c.NotificationConfig.SMS.APISecret)
	viper.SetDefault("MONITORING_NOTIFICATION_SMS_PHONE_NUMBERS", c.NotificationConfig.SMS.PhoneNumbers)

	// 存储配置环境变量
	viper.SetDefault("MONITORING_STORAGE_TYPE", c.StorageConfig.Type)
	viper.SetDefault("MONITORING_STORAGE_DATABASE_ENABLED", c.StorageConfig.Database.Enabled)
	viper.SetDefault("MONITORING_STORAGE_DATABASE_TABLE_PREFIX", c.StorageConfig.Database.TablePrefix)
	viper.SetDefault("MONITORING_STORAGE_DATABASE_RETENTION", c.StorageConfig.Database.Retention)
	viper.SetDefault("MONITORING_STORAGE_FILE_ENABLED", c.StorageConfig.File.Enabled)
	viper.SetDefault("MONITORING_STORAGE_FILE_PATH", c.StorageConfig.File.Path)
	viper.SetDefault("MONITORING_STORAGE_FILE_FORMAT", c.StorageConfig.File.Format)
	viper.SetDefault("MONITORING_STORAGE_FILE_MAX_SIZE", c.StorageConfig.File.MaxSize)
	viper.SetDefault("MONITORING_STORAGE_FILE_MAX_AGE", c.StorageConfig.File.MaxAge)
	viper.SetDefault("MONITORING_STORAGE_REDIS_ENABLED", c.StorageConfig.Redis.Enabled)
	viper.SetDefault("MONITORING_STORAGE_REDIS_KEY_PREFIX", c.StorageConfig.Redis.KeyPrefix)
	viper.SetDefault("MONITORING_STORAGE_REDIS_TTL", c.StorageConfig.Redis.TTL)
}

// Validate 验证配置
func (c *MonitoringConfig) Validate() error {
	// 基础配置验证
	if c.BaseConfig.CheckInterval < time.Second {
		return fmt.Errorf("check interval must be at least 1 second")
	}
	if c.BaseConfig.MaxAlertsPerHour <= 0 {
		return fmt.Errorf("max alerts per hour must be positive")
	}
	if c.BaseConfig.DashboardPort <= 0 || c.BaseConfig.DashboardPort > 65535 {
		return fmt.Errorf("dashboard port must be between 1 and 65535")
	}
	if c.BaseConfig.MetricsPort <= 0 || c.BaseConfig.MetricsPort > 65535 {
		return fmt.Errorf("metrics port must be between 1 and 65535")
	}

	// 系统监控验证
	if c.SystemMonitoring.CPUThreshold <= 0 || c.SystemMonitoring.CPUThreshold > 100 {
		return fmt.Errorf("CPU threshold must be between 0 and 100")
	}
	if c.SystemMonitoring.MemoryThreshold <= 0 || c.SystemMonitoring.MemoryThreshold > 100 {
		return fmt.Errorf("memory threshold must be between 0 and 100")
	}
	if c.SystemMonitoring.DiskThreshold <= 0 || c.SystemMonitoring.DiskThreshold > 100 {
		return fmt.Errorf("disk threshold must be between 0 and 100")
	}

	// 应用监控验证
	if c.ApplicationMonitoring.ErrorRateThreshold < 0 || c.ApplicationMonitoring.ErrorRateThreshold > 100 {
		return fmt.Errorf("error rate threshold must be between 0 and 100")
	}
	if c.ApplicationMonitoring.ThroughputThreshold <= 0 {
		return fmt.Errorf("throughput threshold must be positive")
	}

	// 数据库监控验证
	if c.DatabaseMonitoring.ConnectionThreshold <= 0 {
		return fmt.Errorf("connection threshold must be positive")
	}
	if c.DatabaseMonitoring.SlowQueryThreshold <= 0 {
		return fmt.Errorf("slow query threshold must be positive")
	}

	// 缓存监控验证
	if c.CacheMonitoring.HitRateThreshold < 0 || c.CacheMonitoring.HitRateThreshold > 100 {
		return fmt.Errorf("hit rate threshold must be between 0 and 100")
	}
	if c.CacheMonitoring.MemoryUsageThreshold < 0 || c.CacheMonitoring.MemoryUsageThreshold > 100 {
		return fmt.Errorf("memory usage threshold must be between 0 and 100")
	}

	// 告警配置验证
	if c.AlertConfig.MaxEscalationLevel <= 0 {
		return fmt.Errorf("max escalation level must be positive")
	}

	// 通知配置验证
	if c.NotificationConfig.Email.Enabled {
		if c.NotificationConfig.Email.SMTPHost == "" {
			return fmt.Errorf("SMTP host is required when email notification is enabled")
		}
		if c.NotificationConfig.Email.FromAddress == "" {
			return fmt.Errorf("from address is required when email notification is enabled")
		}
		if c.NotificationConfig.Email.ToAddresses == "" {
			return fmt.Errorf("to addresses is required when email notification is enabled")
		}
	}

	if c.NotificationConfig.Webhook.Enabled {
		if c.NotificationConfig.Webhook.URL == "" {
			return fmt.Errorf("webhook URL is required when webhook notification is enabled")
		}
	}

	if c.NotificationConfig.Slack.Enabled {
		if c.NotificationConfig.Slack.WebhookURL == "" {
			return fmt.Errorf("Slack webhook URL is required when Slack notification is enabled")
		}
	}

	if c.NotificationConfig.DingTalk.Enabled {
		if c.NotificationConfig.DingTalk.WebhookURL == "" {
			return fmt.Errorf("DingTalk webhook URL is required when DingTalk notification is enabled")
		}
	}

	if c.NotificationConfig.SMS.Enabled {
		if c.NotificationConfig.SMS.APIKey == "" {
			return fmt.Errorf("SMS API key is required when SMS notification is enabled")
		}
		if c.NotificationConfig.SMS.PhoneNumbers == "" {
			return fmt.Errorf("phone numbers are required when SMS notification is enabled")
		}
	}

	return nil
}
