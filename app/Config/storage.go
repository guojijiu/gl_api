package Config

import (
	"fmt"
	"path/filepath"
	"strings"
	"github.com/spf13/viper"
)

// StorageConfig 存储配置
type StorageConfig struct {
	UploadPath    string   `mapstructure:"upload_path"`
	MaxFileSize   int      `mapstructure:"max_file_size"`   // 最大文件大小（MB）
	AllowedTypes  []string `mapstructure:"allowed_types"`   // 允许的文件类型
	PrivatePath   string   `mapstructure:"private_path"`    // 私有文件路径
	PublicPath    string   `mapstructure:"public_path"`     // 公共文件路径
	TempPath      string   `mapstructure:"temp_path"`       // 临时文件路径
	LogPath       string   `mapstructure:"log_path"`        // 日志文件路径
	CachePath     string   `mapstructure:"cache_path"`      // 缓存文件路径
}

// SetDefaults 设置存储配置默认值
func (s *StorageConfig) SetDefaults() {
	viper.SetDefault("storage.upload_path", "./storage/app/public")
	viper.SetDefault("storage.max_file_size", 10)
	viper.SetDefault("storage.allowed_types", []string{"jpg", "jpeg", "png", "gif", "pdf", "doc", "docx"})
	viper.SetDefault("storage.private_path", "./storage/app/private")
	viper.SetDefault("storage.public_path", "./storage/app/public")
	viper.SetDefault("storage.temp_path", "./storage/temp")
	viper.SetDefault("storage.log_path", "./storage/logs")
	viper.SetDefault("storage.cache_path", "./storage/framework/cache")
}

// BindEnvs 绑定存储环境变量
func (s *StorageConfig) BindEnvs() {
	viper.BindEnv("storage.upload_path", "STORAGE_UPLOAD_PATH")
	viper.BindEnv("storage.max_file_size", "STORAGE_MAX_FILE_SIZE")
	viper.BindEnv("storage.allowed_types", "STORAGE_ALLOWED_TYPES")
	viper.BindEnv("storage.private_path", "STORAGE_PRIVATE_PATH")
	viper.BindEnv("storage.public_path", "STORAGE_PUBLIC_PATH")
	viper.BindEnv("storage.temp_path", "STORAGE_TEMP_PATH")
	viper.BindEnv("storage.log_path", "STORAGE_LOG_PATH")
	viper.BindEnv("storage.cache_path", "STORAGE_CACHE_PATH")
}

// GetStorageConfig 获取存储配置
func GetStorageConfig() *StorageConfig {
	if globalConfig == nil {
		return nil
	}
	return &globalConfig.Storage
}

// GetMaxFileSizeBytes 获取最大文件大小（字节）
func (s *StorageConfig) GetMaxFileSizeBytes() int64 {
	return int64(s.MaxFileSize * 1024 * 1024)
}

// IsFileTypeAllowed 检查文件类型是否允许
func (s *StorageConfig) IsFileTypeAllowed(fileType string) bool {
	fileType = strings.ToLower(strings.TrimPrefix(fileType, "."))
	for _, allowedType := range s.AllowedTypes {
		if strings.ToLower(allowedType) == fileType {
			return true
		}
	}
	return false
}

// GetPublicFilePath 获取公共文件完整路径
func (s *StorageConfig) GetPublicFilePath(filename string) string {
	return filepath.Join(s.PublicPath, filename)
}

// GetPrivateFilePath 获取私有文件完整路径
func (s *StorageConfig) GetPrivateFilePath(filename string) string {
	return filepath.Join(s.PrivatePath, filename)
}

// GetTempFilePath 获取临时文件完整路径
func (s *StorageConfig) GetTempFilePath(filename string) string {
	return filepath.Join(s.TempPath, filename)
}

// GetLogFilePath 获取日志文件完整路径
func (s *StorageConfig) GetLogFilePath(filename string) string {
	return filepath.Join(s.LogPath, filename)
}

// GetCacheFilePath 获取缓存文件完整路径
func (s *StorageConfig) GetCacheFilePath(filename string) string {
	return filepath.Join(s.CachePath, filename)
}

// Validate 验证存储配置
func (s *StorageConfig) Validate() error {
	if s.UploadPath == "" {
		return fmt.Errorf("上传路径未配置")
	}

	if s.MaxFileSize <= 0 {
		return fmt.Errorf("最大文件大小配置无效")
	}

	if s.MaxFileSize > 1000 { // 超过1GB
		return fmt.Errorf("最大文件大小过大，建议不超过1000MB")
	}

	if len(s.AllowedTypes) == 0 {
		return fmt.Errorf("允许的文件类型未配置")
	}

	return nil
}

// GetAllowedTypesString 获取允许的文件类型字符串
func (s *StorageConfig) GetAllowedTypesString() string {
	return strings.Join(s.AllowedTypes, ", ")
}
