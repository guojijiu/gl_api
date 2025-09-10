package Config

import (
	"time"

	"github.com/spf13/viper"
)

// QueryOptimizationConfig 查询优化配置
type QueryOptimizationConfig struct {
	// 基础配置
	Enabled         bool          `mapstructure:"enabled" json:"enabled"`
	Interval        time.Duration `mapstructure:"interval" json:"interval"`
	RetentionPeriod time.Duration `mapstructure:"retention_period" json:"retention_period"`

	// 慢查询监控
	SlowQuery struct {
		Enabled               bool          `mapstructure:"enabled" json:"enabled"`
		Threshold             time.Duration `mapstructure:"threshold" json:"threshold"`
		NotificationThreshold time.Duration `mapstructure:"notification_threshold" json:"notification_threshold"`
		MaxRecords            int           `mapstructure:"max_records" json:"max_records"`
		RecordStackTrace      bool          `mapstructure:"record_stack_trace" json:"record_stack_trace"`
		RecordExecutionPlan   bool          `mapstructure:"record_execution_plan" json:"record_execution_plan"`
		LogFile               string        `mapstructure:"log_file" json:"log_file"`
		AlertEnabled          bool          `mapstructure:"alert_enabled" json:"alert_enabled"`
		AlertThreshold        time.Duration `mapstructure:"alert_threshold" json:"alert_threshold"`
	} `mapstructure:"slow_query" json:"slow_query"`

	// 索引优化
	IndexOptimization struct {
		Enabled            bool          `mapstructure:"enabled" json:"enabled"`
		AutoAnalyze        bool          `mapstructure:"auto_analyze" json:"auto_analyze"`
		AnalyzeInterval    time.Duration `mapstructure:"analyze_interval" json:"analyze_interval"`
		MaxSuggestions     int           `mapstructure:"max_suggestions" json:"max_suggestions"`
		MinQueryCount      int64         `mapstructure:"min_query_count" json:"min_query_count"`
		MinImprovement     float64       `mapstructure:"min_improvement" json:"min_improvement"`
		ApplyAutomatically bool          `mapstructure:"apply_automatically" json:"apply_automatically"`
		BackupBeforeApply  bool          `mapstructure:"backup_before_apply" json:"backup_before_apply"`
	} `mapstructure:"index_optimization" json:"index_optimization"`

	// 性能监控
	PerformanceMonitoring struct {
		Enabled         bool          `mapstructure:"enabled" json:"enabled"`
		Interval        time.Duration `mapstructure:"interval" json:"interval"`
		RetentionPeriod time.Duration `mapstructure:"retention_period" json:"retention_period"`
		Thresholds      struct {
			AvgResponseTime time.Duration `mapstructure:"avg_response_time" json:"avg_response_time"`
			P95ResponseTime time.Duration `mapstructure:"p95_response_time" json:"p95_response_time"`
			P99ResponseTime time.Duration `mapstructure:"p99_response_time" json:"p99_response_time"`
			ErrorRate       float64       `mapstructure:"error_rate" json:"error_rate"`
			Throughput      int           `mapstructure:"throughput" json:"throughput"`
		} `mapstructure:"thresholds" json:"thresholds"`
		AlertEnabled   bool    `mapstructure:"alert_enabled" json:"alert_enabled"`
		AlertThreshold float64 `mapstructure:"alert_threshold" json:"alert_threshold"`
	} `mapstructure:"performance_monitoring" json:"performance_monitoring"`

	// 报告配置
	Reporting struct {
		Enabled            bool          `mapstructure:"enabled" json:"enabled"`
		Interval           time.Duration `mapstructure:"interval" json:"interval"`
		Format             string        `mapstructure:"format" json:"format"` // json, csv, html
		OutputPath         string        `mapstructure:"output_path" json:"output_path"`
		IncludeStats       bool          `mapstructure:"include_stats" json:"include_stats"`
		IncludeQueries     bool          `mapstructure:"include_queries" json:"include_queries"`
		IncludeSuggestions bool          `mapstructure:"include_suggestions" json:"include_suggestions"`
		MaxFileSize        int64         `mapstructure:"max_file_size" json:"max_file_size"`
		Compression        bool          `mapstructure:"compression" json:"compression"`
	} `mapstructure:"reporting" json:"reporting"`

	// 数据库配置
	Database struct {
		ConnectionPool struct {
			MaxOpenConns    int           `mapstructure:"max_open_conns" json:"max_open_conns"`
			MaxIdleConns    int           `mapstructure:"max_idle_conns" json:"max_idle_conns"`
			ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" json:"conn_max_lifetime"`
		} `mapstructure:"connection_pool" json:"connection_pool"`
		QueryTimeout       time.Duration `mapstructure:"query_timeout" json:"query_timeout"`
		TransactionTimeout time.Duration `mapstructure:"transaction_timeout" json:"transaction_timeout"`
		MaxQuerySize       int           `mapstructure:"max_query_size" json:"max_query_size"`
	} `mapstructure:"database" json:"database"`

	// 缓存配置
	Cache struct {
		Enabled         bool          `mapstructure:"enabled" json:"enabled"`
		TTL             time.Duration `mapstructure:"ttl" json:"ttl"`
		MaxSize         int           `mapstructure:"max_size" json:"max_size"`
		CleanupInterval time.Duration `mapstructure:"cleanup_interval" json:"cleanup_interval"`
	} `mapstructure:"cache" json:"cache"`
}

// SetDefaults 设置默认值
func (c *QueryOptimizationConfig) SetDefaults() {
	c.Enabled = true
	c.Interval = 30 * time.Second
	c.RetentionPeriod = 7 * 24 * time.Hour // 7天

	// 慢查询默认值
	c.SlowQuery.Enabled = true
	c.SlowQuery.Threshold = 1 * time.Second
	c.SlowQuery.NotificationThreshold = 5 * time.Second
	c.SlowQuery.MaxRecords = 1000
	c.SlowQuery.RecordStackTrace = true
	c.SlowQuery.RecordExecutionPlan = true
	c.SlowQuery.LogFile = "logs/slow_queries.log"
	c.SlowQuery.AlertEnabled = true
	c.SlowQuery.AlertThreshold = 10 * time.Second

	// 索引优化默认值
	c.IndexOptimization.Enabled = true
	c.IndexOptimization.AutoAnalyze = false
	c.IndexOptimization.AnalyzeInterval = 24 * time.Hour
	c.IndexOptimization.MaxSuggestions = 100
	c.IndexOptimization.MinQueryCount = 10
	c.IndexOptimization.MinImprovement = 0.1 // 10%
	c.IndexOptimization.ApplyAutomatically = false
	c.IndexOptimization.BackupBeforeApply = true

	// 性能监控默认值
	c.PerformanceMonitoring.Enabled = true
	c.PerformanceMonitoring.Interval = 1 * time.Minute
	c.PerformanceMonitoring.RetentionPeriod = 30 * 24 * time.Hour // 30天
	c.PerformanceMonitoring.Thresholds.AvgResponseTime = 100 * time.Millisecond
	c.PerformanceMonitoring.Thresholds.P95ResponseTime = 500 * time.Millisecond
	c.PerformanceMonitoring.Thresholds.P99ResponseTime = 1 * time.Second
	c.PerformanceMonitoring.Thresholds.ErrorRate = 0.01 // 1%
	c.PerformanceMonitoring.Thresholds.Throughput = 1000
	c.PerformanceMonitoring.AlertEnabled = true
	c.PerformanceMonitoring.AlertThreshold = 0.05 // 5%

	// 报告默认值
	c.Reporting.Enabled = true
	c.Reporting.Interval = 24 * time.Hour
	c.Reporting.Format = "json"
	c.Reporting.OutputPath = "reports/query_optimization"
	c.Reporting.IncludeStats = true
	c.Reporting.IncludeQueries = true
	c.Reporting.IncludeSuggestions = true
	c.Reporting.MaxFileSize = 100 * 1024 * 1024 // 100MB
	c.Reporting.Compression = true

	// 数据库默认值
	c.Database.ConnectionPool.MaxOpenConns = 100
	c.Database.ConnectionPool.MaxIdleConns = 10
	c.Database.ConnectionPool.ConnMaxLifetime = 1 * time.Hour
	c.Database.QueryTimeout = 30 * time.Second
	c.Database.TransactionTimeout = 5 * time.Minute
	c.Database.MaxQuerySize = 1024 * 1024 // 1MB

	// 缓存默认值
	c.Cache.Enabled = true
	c.Cache.TTL = 1 * time.Hour
	c.Cache.MaxSize = 1000
	c.Cache.CleanupInterval = 10 * time.Minute
}

// BindEnvs 绑定环境变量
func (c *QueryOptimizationConfig) BindEnvs() {
	// 基础配置
	bindEnv("QUERY_OPTIMIZATION_ENABLED", &c.Enabled)
	bindEnv("QUERY_OPTIMIZATION_INTERVAL", &c.Interval)
	bindEnv("QUERY_OPTIMIZATION_RETENTION_PERIOD", &c.RetentionPeriod)
}

// LoadFromViper 从Viper加载配置
func (c *QueryOptimizationConfig) LoadFromViper(v *viper.Viper) error {
	if err := v.UnmarshalKey("query_optimization", c); err != nil {
		return err
	}
	return nil
}

// Validate 验证配置
func (c *QueryOptimizationConfig) Validate() error {
	if c.Interval <= 0 {
		c.Interval = 30 * time.Second
	}

	if c.RetentionPeriod <= 0 {
		c.RetentionPeriod = 7 * 24 * time.Hour
	}

	if c.SlowQuery.Threshold <= 0 {
		c.SlowQuery.Threshold = 1 * time.Second
	}

	if c.PerformanceMonitoring.Interval <= 0 {
		c.PerformanceMonitoring.Interval = 1 * time.Minute
	}

	if c.Reporting.Interval <= 0 {
		c.Reporting.Interval = 24 * time.Hour
	}

	return nil
}
