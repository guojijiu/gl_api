package Config

import (
	"fmt"
	"strings"
	"github.com/spf13/viper"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
}

// SetDefaults 设置数据库配置默认值
func (d *DatabaseConfig) SetDefaults() {
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.charset", "utf8mb4")
}

// BindEnvs 绑定数据库环境变量
func (d *DatabaseConfig) BindEnvs() {
	viper.BindEnv("database.driver", "DB_DRIVER")
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.username", "DB_USERNAME")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.database", "DB_DATABASE")
	viper.BindEnv("database.charset", "DB_CHARSET")
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
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
			d.Username, d.Password, d.Host, d.Port, d.Database, d.Charset)
	case "postgres":
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			d.Host, d.Port, d.Username, d.Password, d.Database)
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
		return nil
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

	return nil
}
