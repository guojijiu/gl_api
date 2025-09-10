package Config

import (
	"errors"
	"time"
	"github.com/spf13/viper"
)

// TestConfig 测试环境配置
type TestConfig struct {
	// 基础测试配置
	Base BaseTestConfig `mapstructure:"base"`
	// 数据库测试配置
	Database DatabaseTestConfig `mapstructure:"database"`
	// 缓存测试配置
	Cache CacheTestConfig `mapstructure:"cache"`
	// 性能测试配置
	Performance PerformanceTestConfig `mapstructure:"performance"`
	// 覆盖率测试配置
	Coverage CoverageTestConfig `mapstructure:"coverage"`
}

// BaseTestConfig 基础测试配置
type BaseTestConfig struct {
	// 测试模式
	Mode string `mapstructure:"mode"` // unit, integration, performance, all
	// 测试超时时间
	Timeout time.Duration `mapstructure:"timeout"`
	// 是否并行执行
	Parallel bool `mapstructure:"parallel"`
	// 测试数据目录
	TestDataDir string `mapstructure:"test_data_dir"`
	// 是否清理测试数据
	CleanupTestData bool `mapstructure:"cleanup_test_data"`
	// 测试日志级别
	LogLevel string `mapstructure:"log_level"`
	// 是否显示详细输出
	Verbose bool `mapstructure:"verbose"`
}

// DatabaseTestConfig 数据库测试配置
type DatabaseTestConfig struct {
	// 测试数据库类型
	Type string `mapstructure:"type"` // sqlite, mysql, postgres
	// 测试数据库连接字符串
	DSN string `mapstructure:"dsn"`
	// 是否使用内存数据库
	InMemory bool `mapstructure:"in_memory"`
	// 测试数据种子文件
	SeedFile string `mapstructure:"seed_file"`
	// 是否重置数据库
	ResetOnStart bool `mapstructure:"reset_on_start"`
	// 测试事务回滚
	RollbackTransactions bool `mapstructure:"rollback_transactions"`
}

// CacheTestConfig 缓存测试配置
type CacheTestConfig struct {
	// 测试缓存类型
	Type string `mapstructure:"type"` // memory, redis
	// 测试Redis连接
	Redis RedisTestConfig `mapstructure:"redis"`
	// 是否清理缓存
	CleanupOnFinish bool `mapstructure:"cleanup_on_finish"`
}

// RedisTestConfig Redis测试配置
type RedisTestConfig struct {
	// 测试Redis地址
	Addr string `mapstructure:"addr"`
	// 测试Redis密码
	Password string `mapstructure:"password"`
	// 测试Redis数据库
	DB int `mapstructure:"db"`
	// 连接超时
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
	// 读取超时
	ReadTimeout time.Duration `mapstructure:"read_timeout"`
	// 写入超时
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// PerformanceTestConfig 性能测试配置
type PerformanceTestConfig struct {
	// 并发用户数
	ConcurrentUsers int `mapstructure:"concurrent_users"`
	// 测试持续时间
	Duration time.Duration `mapstructure:"duration"`
	// 请求间隔
	RequestInterval time.Duration `mapstructure:"request_interval"`
	// 是否记录响应时间
	RecordResponseTime bool `mapstructure:"record_response_time"`
	// 性能阈值配置
	Thresholds PerformanceThresholds `mapstructure:"thresholds"`
}

// PerformanceThresholds 性能阈值配置
type PerformanceThresholds struct {
	// 平均响应时间阈值
	AvgResponseTime time.Duration `mapstructure:"avg_response_time"`
	// 95%响应时间阈值
	P95ResponseTime time.Duration `mapstructure:"p95_response_time"`
	// 99%响应时间阈值
	P99ResponseTime time.Duration `mapstructure:"p99_response_time"`
	// 错误率阈值
	ErrorRate float64 `mapstructure:"error_rate"`
	// 吞吐量阈值
	Throughput int `mapstructure:"throughput"`
}

// CoverageTestConfig 覆盖率测试配置
type CoverageTestConfig struct {
	// 是否启用覆盖率测试
	Enabled bool `mapstructure:"enabled"`
	// 覆盖率输出目录
	OutputDir string `mapstructure:"output_dir"`
	// 覆盖率输出格式
	OutputFormat string `mapstructure:"output_format"` // html, xml, text
	// 覆盖率阈值
	Threshold float64 `mapstructure:"threshold"`
	// 排除的包
	ExcludePackages []string `mapstructure:"exclude_packages"`
	// 排除的文件
	ExcludeFiles []string `mapstructure:"exclude_files"`
}

// SetDefaults 设置默认值
func (c *TestConfig) SetDefaults() {
	// 基础测试配置默认值
	c.Base.Mode = "unit"
	c.Base.Timeout = 30 * time.Second
	c.Base.Parallel = true
	c.Base.TestDataDir = "./testdata"
	c.Base.CleanupTestData = true
	c.Base.LogLevel = "info"
	c.Base.Verbose = false

	// 数据库测试配置默认值
	c.Database.Type = "sqlite"
	c.Database.DSN = "file:test.db?cache=shared&mode=memory"
	c.Database.InMemory = true
	c.Database.SeedFile = "./testdata/seed.sql"
	c.Database.ResetOnStart = true
	c.Database.RollbackTransactions = true

	// 缓存测试配置默认值
	c.Cache.Type = "memory"
	c.Cache.CleanupOnFinish = true
	c.Cache.Redis.Addr = "localhost:6379"
	c.Cache.Redis.Password = ""
	c.Cache.Redis.DB = 1
	c.Cache.Redis.DialTimeout = 5 * time.Second
	c.Cache.Redis.ReadTimeout = 3 * time.Second
	c.Cache.Redis.WriteTimeout = 3 * time.Second

	// 性能测试配置默认值
	c.Performance.ConcurrentUsers = 10
	c.Performance.Duration = 60 * time.Second
	c.Performance.RequestInterval = 100 * time.Millisecond
	c.Performance.RecordResponseTime = true
	c.Performance.Thresholds.AvgResponseTime = 100 * time.Millisecond
	c.Performance.Thresholds.P95ResponseTime = 200 * time.Millisecond
	c.Performance.Thresholds.P99ResponseTime = 500 * time.Millisecond
	c.Performance.Thresholds.ErrorRate = 0.01
	c.Performance.Thresholds.Throughput = 100

	// 覆盖率测试配置默认值
	c.Coverage.Enabled = true
	c.Coverage.OutputDir = "./coverage"
	c.Coverage.OutputFormat = "html"
	c.Coverage.Threshold = 80.0
	c.Coverage.ExcludePackages = []string{"main", "vendor"}
	c.Coverage.ExcludeFiles = []string{"*_test.go", "test_*.go"}
}

// BindEnvs 绑定环境变量
func (c *TestConfig) BindEnvs() {
	// 基础测试配置环境变量
	viper.BindEnv("test.base.mode", "TEST_MODE")
	viper.BindEnv("test.base.timeout", "TEST_TIMEOUT")
	viper.BindEnv("test.base.parallel", "TEST_PARALLEL")
	viper.BindEnv("test.base.test_data_dir", "TEST_DATA_DIR")
	viper.BindEnv("test.base.cleanup_test_data", "TEST_CLEANUP_DATA")
	viper.BindEnv("test.base.log_level", "TEST_LOG_LEVEL")
	viper.BindEnv("test.base.verbose", "TEST_VERBOSE")

	// 数据库测试配置环境变量
	viper.BindEnv("test.database.type", "TEST_DB_TYPE")
	viper.BindEnv("test.database.dsn", "TEST_DB_DSN")
	viper.BindEnv("test.database.in_memory", "TEST_DB_IN_MEMORY")
	viper.BindEnv("test.database.seed_file", "TEST_DB_SEED_FILE")
	viper.BindEnv("test.database.reset_on_start", "TEST_DB_RESET")
	viper.BindEnv("test.database.rollback_transactions", "TEST_DB_ROLLBACK")

	// 缓存测试配置环境变量
	viper.BindEnv("test.cache.type", "TEST_CACHE_TYPE")
	viper.BindEnv("test.cache.cleanup_on_finish", "TEST_CACHE_CLEANUP")
	viper.BindEnv("test.cache.redis.addr", "TEST_REDIS_ADDR")
	viper.BindEnv("test.cache.redis.password", "TEST_REDIS_PASSWORD")
	viper.BindEnv("test.cache.redis.db", "TEST_REDIS_DB")

	// 性能测试配置环境变量
	viper.BindEnv("test.performance.concurrent_users", "TEST_PERF_USERS")
	viper.BindEnv("test.performance.duration", "TEST_PERF_DURATION")
	viper.BindEnv("test.performance.request_interval", "TEST_PERF_INTERVAL")
	viper.BindEnv("test.performance.record_response_time", "TEST_PERF_RECORD_TIME")
	viper.BindEnv("test.performance.thresholds.avg_response_time", "TEST_PERF_AVG_THRESHOLD")
	viper.BindEnv("test.performance.thresholds.p95_response_time", "TEST_PERF_P95_THRESHOLD")
	viper.BindEnv("test.performance.thresholds.p99_response_time", "TEST_PERF_P99_THRESHOLD")
	viper.BindEnv("test.performance.thresholds.error_rate", "TEST_PERF_ERROR_THRESHOLD")
	viper.BindEnv("test.performance.thresholds.throughput", "TEST_PERF_THROUGHPUT_THRESHOLD")

	// 覆盖率测试配置环境变量
	viper.BindEnv("test.coverage.enabled", "TEST_COVERAGE_ENABLED")
	viper.BindEnv("test.coverage.output_dir", "TEST_COVERAGE_OUTPUT_DIR")
	viper.BindEnv("test.coverage.output_format", "TEST_COVERAGE_FORMAT")
	viper.BindEnv("test.coverage.threshold", "TEST_COVERAGE_THRESHOLD")
}

// Validate 验证配置
func (c *TestConfig) Validate() error {
	// 验证基础测试配置
	if c.Base.Timeout <= 0 {
		return errors.New("test timeout must be positive")
	}

	// 验证数据库测试配置
	if c.Database.Type == "" {
		return errors.New("database type is required")
	}

	// 验证性能测试配置
	if c.Performance.ConcurrentUsers <= 0 {
		return errors.New("concurrent users must be positive")
	}
	if c.Performance.Duration <= 0 {
		return errors.New("test duration must be positive")
	}

	// 验证覆盖率测试配置
	if c.Coverage.Enabled {
		if c.Coverage.Threshold < 0 || c.Coverage.Threshold > 100 {
			return errors.New("coverage threshold must be between 0 and 100")
		}
	}

	return nil
}
