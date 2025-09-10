package Config

import (
	"fmt"
	"time"
	"github.com/spf13/viper"
)

// PerformanceMonitoringConfig 性能监控配置
type PerformanceMonitoringConfig struct {
	// 基础配置
	Enabled           bool          `mapstructure:"enabled" json:"enabled"`
	Interval          time.Duration `mapstructure:"interval" json:"interval"`
	RetentionPeriod   time.Duration `mapstructure:"retention_period" json:"retention_period"`
	BatchSize         int           `mapstructure:"batch_size" json:"batch_size"`
	
	// 系统资源监控
	SystemResourcesEnabled bool `mapstructure:"system_resources_enabled" json:"system_resources_enabled"`
	CPUEnabled            bool `mapstructure:"cpu_enabled" json:"cpu_enabled"`
	CPUUsageThreshold     float64 `mapstructure:"cpu_usage_threshold" json:"cpu_usage_threshold"`
	MemoryEnabled         bool `mapstructure:"memory_enabled" json:"memory_enabled"`
	MemoryUsageThreshold  float64 `mapstructure:"memory_usage_threshold" json:"memory_usage_threshold"`
	DiskEnabled           bool `mapstructure:"disk_enabled" json:"disk_enabled"`
	DiskUsageThreshold    float64 `mapstructure:"disk_usage_threshold" json:"disk_usage_threshold"`
	NetworkEnabled        bool `mapstructure:"network_enabled" json:"network_enabled"`
	NetworkBandwidthThreshold int64 `mapstructure:"network_bandwidth_threshold" json:"network_bandwidth_threshold"`
	
	// 应用监控
	ApplicationEnabled    bool `mapstructure:"application_enabled" json:"application_enabled"`
	HTTPEnabled           bool `mapstructure:"http_enabled" json:"http_enabled"`
	HTTPResponseTimeThreshold time.Duration `mapstructure:"http_response_time_threshold" json:"http_response_time_threshold"`
	HTTPErrorRateThreshold    float64 `mapstructure:"http_error_rate_threshold" json:"http_error_rate_threshold"`
	DatabaseEnabled       bool `mapstructure:"database_enabled" json:"database_enabled"`
	DatabaseConnectionThreshold int `mapstructure:"database_connection_threshold" json:"database_connection_threshold"`
	DatabaseQueryTimeThreshold time.Duration `mapstructure:"database_query_time_threshold" json:"database_query_time_threshold"`
	CacheEnabled          bool `mapstructure:"cache_enabled" json:"cache_enabled"`
	CacheHitRateThreshold float64 `mapstructure:"cache_hit_rate_threshold" json:"cache_hit_rate_threshold"`
	GoRuntimeEnabled      bool `mapstructure:"go_runtime_enabled" json:"go_runtime_enabled"`
	GoRuntimeGoroutineThreshold int `mapstructure:"go_runtime_goroutine_threshold" json:"go_runtime_goroutine_threshold"`
	GoRuntimeMemoryThreshold    int `mapstructure:"go_runtime_memory_threshold" json:"go_runtime_memory_threshold"`
	
	// 业务监控
	BusinessEnabled       bool `mapstructure:"business_enabled" json:"business_enabled"`
	UserActivityEnabled   bool `mapstructure:"user_activity_enabled" json:"user_activity_enabled"`
	UserActivityThreshold int `mapstructure:"user_activity_threshold" json:"user_activity_threshold"`
	APIUsageEnabled       bool `mapstructure:"api_usage_enabled" json:"api_usage_enabled"`
	APIUsageThreshold     int `mapstructure:"api_usage_threshold" json:"api_usage_threshold"`
	
	// 告警配置
	AlertsEnabled         bool `mapstructure:"alerts_enabled" json:"alerts_enabled"`
	AlertsSeverity        string `mapstructure:"alerts_severity" json:"alerts_severity"`
	AlertsCooldownPeriod  time.Duration `mapstructure:"alerts_cooldown_period" json:"alerts_cooldown_period"`
	
	// 存储配置
	StorageType           string `mapstructure:"storage_type" json:"storage_type"`
	StorageCompression    bool `mapstructure:"storage_compression" json:"storage_compression"`
	StorageBatchWriteSize int `mapstructure:"storage_batch_write_size" json:"storage_batch_write_size"`
}

// SetDefaults 设置默认值
func (c *PerformanceMonitoringConfig) SetDefaults() {
	c.Enabled = true
	c.Interval = 30 * time.Second
	c.RetentionPeriod = 7 * 24 * time.Hour // 7天
	c.BatchSize = 100
	
	// 系统资源监控默认值
	c.SystemResourcesEnabled = true
	c.CPUEnabled = true
	c.CPUUsageThreshold = 80.0 // 80%
	c.MemoryEnabled = true
	c.MemoryUsageThreshold = 85.0 // 85%
	c.DiskEnabled = true
	c.DiskUsageThreshold = 90.0 // 90%
	c.NetworkEnabled = true
	c.NetworkBandwidthThreshold = 100 * 1024 * 1024 // 100MB/s
	
	// 应用监控默认值
	c.ApplicationEnabled = true
	c.HTTPEnabled = true
	c.HTTPResponseTimeThreshold = 1 * time.Second
	c.HTTPErrorRateThreshold = 0.05 // 5%
	c.DatabaseEnabled = true
	c.DatabaseConnectionThreshold = 100
	c.DatabaseQueryTimeThreshold = 1 * time.Second
	c.CacheEnabled = true
	c.CacheHitRateThreshold = 80.0 // 80%
	c.GoRuntimeEnabled = true
	c.GoRuntimeGoroutineThreshold = 10000
	c.GoRuntimeMemoryThreshold = 100 * 1024 * 1024 // 100MB
	
	// 业务监控默认值
	c.BusinessEnabled = true
	c.UserActivityEnabled = true
	c.UserActivityThreshold = 100
	c.APIUsageEnabled = true
	c.APIUsageThreshold = 1000
	
	// 告警默认值
	c.AlertsEnabled = true
	c.AlertsSeverity = "warning"
	c.AlertsCooldownPeriod = 5 * time.Minute
	
	// 存储默认值
	c.StorageType = "memory"
	c.StorageCompression = false
	c.StorageBatchWriteSize = 100
}

// BindEnvs 绑定环境变量
func (c *PerformanceMonitoringConfig) BindEnvs() {
	viper.SetDefault("PERFORMANCE_MONITORING_ENABLED", c.Enabled)
	viper.SetDefault("PERFORMANCE_MONITORING_INTERVAL", c.Interval)
	viper.SetDefault("PERFORMANCE_MONITORING_RETENTION_PERIOD", c.RetentionPeriod)
	viper.SetDefault("PERFORMANCE_MONITORING_BATCH_SIZE", c.BatchSize)
	
	// 系统资源监控环境变量
	viper.SetDefault("PERFORMANCE_MONITORING_SYSTEM_ENABLED", c.SystemResourcesEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_CPU_ENABLED", c.CPUEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_CPU_THRESHOLD", c.CPUUsageThreshold)
	viper.SetDefault("PERFORMANCE_MONITORING_MEMORY_ENABLED", c.MemoryEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_MEMORY_THRESHOLD", c.MemoryUsageThreshold)
	viper.SetDefault("PERFORMANCE_MONITORING_DISK_ENABLED", c.DiskEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_DISK_THRESHOLD", c.DiskUsageThreshold)
	viper.SetDefault("PERFORMANCE_MONITORING_NETWORK_ENABLED", c.NetworkEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_NETWORK_THRESHOLD", c.NetworkBandwidthThreshold)
	
	// 应用监控环境变量
	viper.SetDefault("PERFORMANCE_MONITORING_APP_ENABLED", c.ApplicationEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_HTTP_ENABLED", c.HTTPEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_HTTP_RESPONSE_TIME_THRESHOLD", c.HTTPResponseTimeThreshold)
	viper.SetDefault("PERFORMANCE_MONITORING_HTTP_ERROR_RATE_THRESHOLD", c.HTTPErrorRateThreshold)
	viper.SetDefault("PERFORMANCE_MONITORING_DB_ENABLED", c.DatabaseEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_DB_CONNECTION_THRESHOLD", c.DatabaseConnectionThreshold)
	viper.SetDefault("PERFORMANCE_MONITORING_DB_QUERY_TIME_THRESHOLD", c.DatabaseQueryTimeThreshold)
	viper.SetDefault("PERFORMANCE_MONITORING_CACHE_ENABLED", c.CacheEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_CACHE_HIT_RATE_THRESHOLD", c.CacheHitRateThreshold)
	viper.SetDefault("PERFORMANCE_MONITORING_GO_RUNTIME_ENABLED", c.GoRuntimeEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_GO_RUNTIME_GOROUTINE_THRESHOLD", c.GoRuntimeGoroutineThreshold)
	viper.SetDefault("PERFORMANCE_MONITORING_GO_RUNTIME_MEMORY_THRESHOLD", c.GoRuntimeMemoryThreshold)
	
	// 业务监控环境变量
	viper.SetDefault("PERFORMANCE_MONITORING_BUSINESS_ENABLED", c.BusinessEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_USER_ACTIVITY_ENABLED", c.UserActivityEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_USER_ACTIVITY_THRESHOLD", c.UserActivityThreshold)
	viper.SetDefault("PERFORMANCE_MONITORING_API_USAGE_ENABLED", c.APIUsageEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_API_USAGE_THRESHOLD", c.APIUsageThreshold)
	
	// 告警环境变量
	viper.SetDefault("PERFORMANCE_MONITORING_ALERTS_ENABLED", c.AlertsEnabled)
	viper.SetDefault("PERFORMANCE_MONITORING_ALERTS_SEVERITY", c.AlertsSeverity)
	viper.SetDefault("PERFORMANCE_MONITORING_ALERTS_COOLDOWN_PERIOD", c.AlertsCooldownPeriod)
	
	// 存储环境变量
	viper.SetDefault("PERFORMANCE_MONITORING_STORAGE_TYPE", c.StorageType)
	viper.SetDefault("PERFORMANCE_MONITORING_STORAGE_COMPRESSION", c.StorageCompression)
	viper.SetDefault("PERFORMANCE_MONITORING_STORAGE_BATCH_WRITE_SIZE", c.StorageBatchWriteSize)
}

// Validate 验证配置
func (c *PerformanceMonitoringConfig) Validate() error {
	if c.Interval < time.Second {
		return fmt.Errorf("监控间隔必须至少1秒")
	}
	if c.RetentionPeriod < time.Hour {
		return fmt.Errorf("数据保留时间必须至少1小时")
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("批处理大小必须大于0")
	}
	
	if c.CPUUsageThreshold <= 0 || c.CPUUsageThreshold > 100 {
		return fmt.Errorf("CPU使用率阈值必须在0-100之间")
	}
	if c.MemoryUsageThreshold <= 0 || c.MemoryUsageThreshold > 100 {
		return fmt.Errorf("内存使用率阈值必须在0-100之间")
	}
	if c.DiskUsageThreshold <= 0 || c.DiskUsageThreshold > 100 {
		return fmt.Errorf("磁盘使用率阈值必须在0-100之间")
	}
	
	if c.HTTPResponseTimeThreshold <= 0 {
		return fmt.Errorf("HTTP响应时间阈值必须大于0")
	}
	if c.HTTPErrorRateThreshold < 0 || c.HTTPErrorRateThreshold > 1 {
		return fmt.Errorf("HTTP错误率阈值必须在0-1之间")
	}
	
	if c.DatabaseConnectionThreshold <= 0 {
		return fmt.Errorf("数据库连接阈值必须大于0")
	}
	if c.DatabaseQueryTimeThreshold <= 0 {
		return fmt.Errorf("数据库查询时间阈值必须大于0")
	}
	
	if c.CacheHitRateThreshold < 0 || c.CacheHitRateThreshold > 100 {
		return fmt.Errorf("缓存命中率阈值必须在0-100之间")
	}
	
	if c.GoRuntimeGoroutineThreshold <= 0 {
		return fmt.Errorf("Goroutine阈值必须大于0")
	}
	if c.GoRuntimeMemoryThreshold <= 0 {
		return fmt.Errorf("内存阈值必须大于0")
	}
	
	if c.UserActivityThreshold <= 0 {
		return fmt.Errorf("用户活动阈值必须大于0")
	}
	if c.APIUsageThreshold <= 0 {
		return fmt.Errorf("API使用阈值必须大于0")
	}
	
	if c.AlertsCooldownPeriod <= 0 {
		return fmt.Errorf("告警冷却时间必须大于0")
	}
	
	if c.StorageBatchWriteSize <= 0 {
		return fmt.Errorf("批量写入大小必须大于0")
	}
	
	return nil
}
