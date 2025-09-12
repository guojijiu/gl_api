package Container

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Storage"
	"cloud-platform-api/app/Utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DatabaseProvider 数据库服务提供者
type DatabaseProvider struct{}

func (p *DatabaseProvider) Register(container *Container) error {
	// 注册数据库连接
	container.RegisterSingleton("database", func() interface{} {
		return Database.GetDB()
	})

	// 注册数据库配置
	container.RegisterSingleton("database_config", func() interface{} {
		return Config.GetDatabaseConfig()
	})

	return nil
}

// ConfigProvider 配置服务提供者
type ConfigProvider struct{}

func (p *ConfigProvider) Register(container *Container) error {
	// 注册全局配置
	container.RegisterSingleton("config", func() interface{} {
		return Config.GetConfig()
	})

	// 注册各种配置
	container.RegisterSingleton("server_config", func() interface{} {
		return Config.GetServerConfig()
	})

	container.RegisterSingleton("redis_config", func() interface{} {
		return Config.GetRedisConfig()
	})

	container.RegisterSingleton("storage_config", func() interface{} {
		return Config.GetStorageConfig()
	})

	return nil
}

// BusinessServiceProvider 业务服务提供者
type BusinessServiceProvider struct{}

func (p *BusinessServiceProvider) Register(container *Container) error {
	// 注册用户服务
	container.RegisterSingleton("user_service", func() interface{} {
		return Services.NewUserService()
	})

	// 注册认证服务
	container.RegisterSingleton("auth_service", func() interface{} {
		return Services.NewAuthService()
	})

	// 注册Redis服务
	container.RegisterSingleton("redis_service", func() interface{} {
		config, err := container.Get("redis_config")
		if err != nil || config == nil {
			// 如果配置不可用，返回默认配置
			return Services.NewRedisService(&Services.RedisConfig{
				Host:     "localhost",
				Port:     6379,
				Password: "",
				DB:       0,
			})
		}
		redisConfig := config.(*Config.RedisConfig)
		return Services.NewRedisService(&Services.RedisConfig{
			Host:     redisConfig.Host,
			Port:     redisConfig.Port,
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		})
	})

	// 注册缓存服务
	container.RegisterSingleton("cache_service", func() interface{} {
		return Services.NewOptimizedCacheService()
	})

	// 注册监控服务
	container.RegisterSingleton("monitoring_service", func() interface{} {
		return Services.NewOptimizedMonitoringService()
	})

	// 注册邮件服务
	container.RegisterSingleton("email_service", func() interface{} {
		config, _ := container.Get("config")
		emailConfig := config.(*Config.Config).Email
		return Services.NewEmailService(&Services.EmailConfig{
			Host:     emailConfig.Host,
			Port:     emailConfig.Port,
			Username: emailConfig.Username,
			Password: emailConfig.Password,
			From:     emailConfig.From,
			UseTLS:   emailConfig.UseTLS,
		})
	})

	// 注册日志服务
	container.RegisterSingleton("log_service", func() interface{} {
		config, _ := container.Get("config")
		return Services.NewLogManagerService(&config.(*Config.Config).Log)
	})

	// 注册安全服务
	container.RegisterSingleton("security_service", func() interface{} {
		db, _ := container.Get("database")
		config, _ := container.Get("config")
		return Services.NewSecurityService(db.(*gorm.DB), &config.(*Config.Config).Security)
	})

	// 注册审计服务
	container.RegisterSingleton("audit_service", func() interface{} {
		db, _ := container.Get("database")
		return Services.NewAuditService(db.(*gorm.DB))
	})

	// 注册备份服务
	container.RegisterSingleton("backup_service", func() interface{} {
		storageManager, _ := container.Get("storage_manager")
		config, _ := container.Get("config")
		storageConfig := config.(*Config.Config).Storage
		return Services.NewBackupService(storageManager.(*Storage.StorageManager), &Services.BackupConfig{
			EnableAutoBackup:    storageConfig.EnableAutoBackup,
			BackupInterval:      time.Duration(storageConfig.BackupInterval) * time.Minute,
			MaxBackupFiles:      storageConfig.MaxBackupFiles,
			BackupRetentionDays: storageConfig.BackupRetentionDays,
			BackupPath:          storageConfig.BackupPath,
			EnableCompression:   storageConfig.EnableCompression,
			EnableEncryption:    storageConfig.EnableEncryption,
			EncryptionKey:       storageConfig.EncryptionKey,
		})
	})

	// 注册WebSocket服务
	container.RegisterSingleton("websocket_service", func() interface{} {
		config, _ := container.Get("config")
		wsConfig := config.(*Config.Config).WebSocket
		return Services.NewWebSocketService(&wsConfig)
	})

	return nil
}

// StorageProvider 存储服务提供者
type StorageProvider struct{}

func (p *StorageProvider) Register(container *Container) error {
	// 注册存储服务
	container.RegisterSingleton("storage_service", func() interface{} {
		config, _ := container.Get("storage_config")
		storageConfig := config.(*Config.StorageConfig)
		return Storage.NewStorageService(&Config.StorageConfig{
			BasePath:            storageConfig.BasePath,
			EnableAutoBackup:    storageConfig.EnableAutoBackup,
			BackupInterval:      storageConfig.BackupInterval,
			MaxBackupFiles:      storageConfig.MaxBackupFiles,
			BackupRetentionDays: storageConfig.BackupRetentionDays,
			BackupPath:          storageConfig.BackupPath,
			EnableCompression:   storageConfig.EnableCompression,
			EnableEncryption:    storageConfig.EnableEncryption,
			EncryptionKey:       storageConfig.EncryptionKey,
		})
	})

	// 注册存储管理器
	container.RegisterSingleton("storage_manager", func() interface{} {
		config, _ := container.Get("storage_config")
		storageConfig := config.(*Config.StorageConfig)
		return Storage.NewStorageManager(&Config.StorageConfig{
			BasePath:            storageConfig.BasePath,
			EnableAutoBackup:    storageConfig.EnableAutoBackup,
			BackupInterval:      storageConfig.BackupInterval,
			MaxBackupFiles:      storageConfig.MaxBackupFiles,
			BackupRetentionDays: storageConfig.BackupRetentionDays,
			BackupPath:          storageConfig.BackupPath,
			EnableCompression:   storageConfig.EnableCompression,
			EnableEncryption:    storageConfig.EnableEncryption,
			EncryptionKey:       storageConfig.EncryptionKey,
		})
	})

	// 注册日志服务
	container.RegisterSingleton("storage_log_service", func() interface{} {
		config, _ := container.Get("config")
		return Storage.NewLogService(&config.(*Config.Config).Log)
	})

	// 注册临时服务
	container.RegisterSingleton("temp_service", func() interface{} {
		config, _ := container.Get("storage_config")
		storageConfig := config.(*Config.StorageConfig)
		return Storage.NewTempService(&Config.StorageConfig{
			BasePath:            storageConfig.BasePath,
			EnableAutoBackup:    storageConfig.EnableAutoBackup,
			BackupInterval:      storageConfig.BackupInterval,
			MaxBackupFiles:      storageConfig.MaxBackupFiles,
			BackupRetentionDays: storageConfig.BackupRetentionDays,
			BackupPath:          storageConfig.BackupPath,
			EnableCompression:   storageConfig.EnableCompression,
			EnableEncryption:    storageConfig.EnableEncryption,
			EncryptionKey:       storageConfig.EncryptionKey,
		})
	})

	return nil
}

// UtilsProvider 工具服务提供者
type UtilsProvider struct{}

func (p *UtilsProvider) Register(container *Container) error {
	// 注册JWT工具
	container.RegisterSingleton("jwt_utils", func() interface{} {
		config, _ := container.Get("config")
		return Utils.NewJWTUtils(&config.(*Config.Config).JWT)
	})

	// 注册密码工具
	container.RegisterSingleton("password_utils", func() interface{} {
		return Utils.NewPasswordUtils()
	})

	// 注册增强错误处理
	container.RegisterSingleton("error_handler", func() interface{} {
		return Utils.NewEnhancedErrorHandler()
	})

	// 注册增强日志
	container.RegisterSingleton("enhanced_logger", func() interface{} {
		return Utils.NewEnhancedLogger()
	})

	return nil
}

// RegisterAllProviders 注册所有服务提供者
func RegisterAllProviders(container *Container) error {
	providers := []ServiceProvider{
		&ConfigProvider{},
		&DatabaseProvider{},
		&BusinessServiceProvider{},
		&StorageProvider{},
		&UtilsProvider{},
	}

	for _, provider := range providers {
		if err := container.RegisterProvider(provider); err != nil {
			return fmt.Errorf("failed to register provider: %v", err)
		}
	}

	return nil
}

// InitializeContainer 初始化容器
func InitializeContainer() (*Container, error) {
	container := NewContainer()

	// 注册所有服务提供者
	if err := RegisterAllProviders(container); err != nil {
		return nil, fmt.Errorf("failed to register providers: %v", err)
	}

	// 设置全局容器
	SetGlobalContainer(container)

	return container, nil
}

// GetService 从全局容器获取服务
func GetService(name string) (interface{}, error) {
	container := GetGlobalContainer()
	if container == nil {
		return nil, fmt.Errorf("global container not initialized")
	}
	return container.Get(name)
}

// ResolveService 解析服务并注入依赖
func ResolveService(target interface{}) error {
	container := GetGlobalContainer()
	if container == nil {
		return fmt.Errorf("global container not initialized")
	}
	return container.Resolve(target)
}
