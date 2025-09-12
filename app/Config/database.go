package Config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver            string        `mapstructure:"driver"`
	Host              string        `mapstructure:"host"`
	Port              string        `mapstructure:"port"`
	Username          string        `mapstructure:"username"`
	Password          string        `mapstructure:"password"`
	Database          string        `mapstructure:"database"`
	Charset           string        `mapstructure:"charset"`
	MaxOpenConns      int           `mapstructure:"max_open_conns"`
	MaxIdleConns      int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime   time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime   time.Duration `mapstructure:"conn_max_idle_time"`
	ConnectionTimeout int           `mapstructure:"connection_timeout"` // 秒
	ReadTimeout       int           `mapstructure:"read_timeout"`       // 秒
	WriteTimeout      int           `mapstructure:"write_timeout"`      // 秒
}

// SetDefaults 设置数据库配置默认值
func (d *DatabaseConfig) SetDefaults() {
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.charset", "utf8mb4")

	// 连接池默认配置
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", "1h")   // 1小时
	viper.SetDefault("database.conn_max_idle_time", "10m") // 10分钟
	viper.SetDefault("database.connection_timeout", 30)    // 30秒
	viper.SetDefault("database.read_timeout", 30)          // 30秒
	viper.SetDefault("database.write_timeout", 30)         // 30秒
}

// BindEnvs 绑定数据库环境变量
func (d *DatabaseConfig) BindEnvs() {
	viper.BindEnv("database.driver", "DATABASE_DRIVER")
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.username", "DATABASE_USERNAME")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.database", "DATABASE_NAME")
	viper.BindEnv("database.charset", "DATABASE_CHARSET")

	// 连接池环境变量
	viper.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	viper.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")
	viper.BindEnv("database.conn_max_idle_time", "DATABASE_CONN_MAX_IDLE_TIME")
	viper.BindEnv("database.connection_timeout", "DATABASE_CONNECTION_TIMEOUT")
	viper.BindEnv("database.read_timeout", "DATABASE_READ_TIMEOUT")
	viper.BindEnv("database.write_timeout", "DATABASE_WRITE_TIMEOUT")
}

// GetDatabaseConfig 获取数据库配置
func GetDatabaseConfig() *DatabaseConfig {
	if globalConfig == nil {
		return nil
	}
	return &globalConfig.Database
}

// GetDSN 获取数据库连接字符串
func (d *DatabaseConfig) GetDSN() string {
	switch strings.ToLower(d.Driver) {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local&timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
			d.Username, d.Password, d.Host, d.Port, d.Database, d.Charset, d.ConnectionTimeout, d.ReadTimeout, d.WriteTimeout)
	case "postgres":
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai connect_timeout=%d",
			d.Host, d.Port, d.Username, d.Password, d.Database, d.ConnectionTimeout)
	case "sqlite":
		return d.Database
	default:
		return ""
	}
}

// IsSQLite 检查是否为SQLite数据库
func (d *DatabaseConfig) IsSQLite() bool {
	return strings.ToLower(d.Driver) == "sqlite"
}

// IsMySQL 检查是否为MySQL数据库
func (d *DatabaseConfig) IsMySQL() bool {
	return strings.ToLower(d.Driver) == "mysql"
}

// IsPostgreSQL 检查是否为PostgreSQL数据库
func (d *DatabaseConfig) IsPostgreSQL() bool {
	return strings.ToLower(d.Driver) == "postgres"
}

// Validate 验证数据库配置
func (d *DatabaseConfig) Validate() error {
	if d.Driver == "" {
		return fmt.Errorf("数据库驱动未配置")
	}

	if d.IsSQLite() {
		if d.Database == "" {
			return fmt.Errorf("SQLite数据库文件路径未配置")
		}
		return d.validateConnectionPool()
	}

	if d.Host == "" {
		return fmt.Errorf("数据库主机未配置")
	}

	if d.Port == "" {
		return fmt.Errorf("数据库端口未配置")
	}

	if d.Database == "" {
		return fmt.Errorf("数据库名称未配置")
	}

	if d.Charset == "" {
		return fmt.Errorf("数据库字符集未配置")
	}

	return d.validateConnectionPool()
}

// validateConnectionPool 验证连接池配置
func (d *DatabaseConfig) validateConnectionPool() error {
	// 验证最大打开连接数
	if d.MaxOpenConns <= 0 {
		return fmt.Errorf("最大打开连接数必须大于0，当前值: %d", d.MaxOpenConns)
	}

	if d.MaxOpenConns > 1000 {
		return fmt.Errorf("最大打开连接数过大，建议不超过1000，当前值: %d", d.MaxOpenConns)
	}

	// 验证最大空闲连接数
	if d.MaxIdleConns < 0 {
		return fmt.Errorf("最大空闲连接数不能为负数，当前值: %d", d.MaxIdleConns)
	}

	if d.MaxIdleConns > d.MaxOpenConns {
		return fmt.Errorf("最大空闲连接数不能大于最大打开连接数，当前值: %d > %d", d.MaxIdleConns, d.MaxOpenConns)
	}

	// 验证连接生命周期
	if d.ConnMaxLifetime <= 0 {
		return fmt.Errorf("连接最大生命周期必须大于0，当前值: %v", d.ConnMaxLifetime)
	}

	if d.ConnMaxLifetime > 24*time.Hour { // 超过24小时
		return fmt.Errorf("连接最大生命周期过长，建议不超过24小时，当前值: %v", d.ConnMaxLifetime)
	}

	// 验证空闲连接超时
	if d.ConnMaxIdleTime < 0 {
		return fmt.Errorf("连接最大空闲时间不能为负数，当前值: %v", d.ConnMaxIdleTime)
	}

	if d.ConnMaxIdleTime > time.Hour { // 超过1小时
		return fmt.Errorf("连接最大空闲时间过长，建议不超过1小时，当前值: %v", d.ConnMaxIdleTime)
	}

	// 验证超时配置
	if d.ConnectionTimeout <= 0 {
		return fmt.Errorf("连接超时时间必须大于0，当前值: %d秒", d.ConnectionTimeout)
	}

	if d.ReadTimeout <= 0 {
		return fmt.Errorf("读取超时时间必须大于0，当前值: %d秒", d.ReadTimeout)
	}

	if d.WriteTimeout <= 0 {
		return fmt.Errorf("写入超时时间必须大于0，当前值: %d秒", d.WriteTimeout)
	}

	return nil
}

// GetOptimizedConnectionPoolConfig 获取优化的连接池配置
func (d *DatabaseConfig) GetOptimizedConnectionPoolConfig() (maxOpen, maxIdle int, maxLifetime, maxIdleTime time.Duration) {
	// 根据环境自动优化连接池配置
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))

	switch env {
	case "production", "prod":
		// 生产环境：更高的连接数，更长的生命周期
		maxOpen = max(d.MaxOpenConns, 200)
		maxIdle = max(d.MaxIdleConns, 20)
		maxLifetime = maxDuration(d.ConnMaxLifetime, 2*time.Hour)    // 至少2小时
		maxIdleTime = maxDuration(d.ConnMaxIdleTime, 30*time.Minute) // 至少30分钟
	case "development", "dev":
		// 开发环境：较少的连接数，较短的生命周期
		maxOpen = min(d.MaxOpenConns, 50)
		maxIdle = min(d.MaxIdleConns, 5)
		maxLifetime = minDuration(d.ConnMaxLifetime, 30*time.Minute) // 最多30分钟
		maxIdleTime = minDuration(d.ConnMaxIdleTime, 5*time.Minute)  // 最多5分钟
	default:
		// 默认配置
		maxOpen = d.MaxOpenConns
		maxIdle = d.MaxIdleConns
		maxLifetime = d.ConnMaxLifetime
		maxIdleTime = d.ConnMaxIdleTime
	}

	return
}

// 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
