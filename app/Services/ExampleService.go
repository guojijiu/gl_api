package Services

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Utils"
	"context"
	"fmt"
	"time"
)

// ExampleService 示例服务，展示依赖注入的使用
type ExampleService struct {
	*ServiceBase

	// 使用依赖注入标签
	DatabaseConfig *Config.DatabaseConfig `inject:"database_config"`
	CacheService   *CacheService          `inject:"cache_service"`
	Logger         *Utils.EnhancedLogger  `inject:"enhanced_logger"`
	ErrorHandler   *Utils.EnhancedError   `inject:"error_handler"`
}

// NewExampleService 创建示例服务
func NewExampleService() *ExampleService {
	service := &ExampleService{
		ServiceBase: NewServiceBase("example_service"),
	}

	// 注册到全局服务管理器
	RegisterGlobalService("example_service", service)

	return service
}

// ProcessData 处理数据的示例方法
func (s *ExampleService) ProcessData(ctx context.Context, data interface{}) error {
	// 使用注入的日志服务
	s.Logger.WithFields(map[string]interface{}{
		"service": s.GetName(),
		"data":    data,
	}).Info("开始处理数据")

	// 使用注入的数据库配置
	if s.DatabaseConfig != nil {
		s.Logger.WithFields(map[string]interface{}{
			"driver": s.DatabaseConfig.Driver,
			"host":   s.DatabaseConfig.Host,
		}).Info("数据库配置信息")
	}

	// 使用注入的缓存服务
	if s.CacheService != nil {
		cacheKey := fmt.Sprintf("data_%v", data)
		if err := s.CacheService.Set(cacheKey, data, 3600*time.Second); err != nil {
			s.Logger.WithFields(map[string]interface{}{
				"error": err.Error(),
				"key":   cacheKey,
			}).Error("缓存设置失败")
		}
	}

	// 模拟一些业务逻辑
	if err := s.validateData(data); err != nil {
		// 使用注入的错误处理器
		enhancedErr := Utils.WrapEnhancedError(err, "VALIDATION_ERROR", "数据验证失败")
		enhancedErr.WithContext("data", data)
		s.Logger.WithFields(map[string]interface{}{
			"error": enhancedErr.Error(),
		}).Error("数据验证失败")
		return enhancedErr
	}

	s.Logger.WithFields(map[string]interface{}{
		"service": s.GetName(),
	}).Info("数据处理完成")

	return nil
}

// validateData 验证数据
func (s *ExampleService) validateData(data interface{}) error {
	if data == nil {
		return fmt.Errorf("数据不能为空")
	}

	// 这里可以添加更多的验证逻辑
	return nil
}

// GetServiceInfo 获取服务信息
func (s *ExampleService) GetServiceInfo() map[string]interface{} {
	info := map[string]interface{}{
		"name":         s.GetName(),
		"initialized":  s.IsInitialized(),
		"dependencies": s.getDependencyInfo(),
	}

	return info
}

// getDependencyInfo 获取依赖信息
func (s *ExampleService) getDependencyInfo() map[string]interface{} {
	deps := make(map[string]interface{})

	if s.DatabaseConfig != nil {
		deps["database_config"] = "已注入"
	} else {
		deps["database_config"] = "未注入"
	}

	if s.CacheService != nil {
		deps["cache_service"] = "已注入"
	} else {
		deps["cache_service"] = "未注入"
	}

	if s.Logger != nil {
		deps["logger"] = "已注入"
	} else {
		deps["logger"] = "未注入"
	}

	if s.ErrorHandler != nil {
		deps["error_handler"] = "已注入"
	} else {
		deps["error_handler"] = "未注入"
	}

	return deps
}

// Initialize 重写初始化方法
func (s *ExampleService) Initialize() error {
	// 调用基类的初始化方法
	if err := s.ServiceBase.Initialize(); err != nil {
		return err
	}

	// 添加服务特定的初始化逻辑
	s.Logger.WithFields(map[string]interface{}{
		"service": s.GetName(),
	}).Info("示例服务初始化完成")

	return nil
}

// Shutdown 重写关闭方法
func (s *ExampleService) Shutdown() error {
	// 添加服务特定的关闭逻辑
	if s.Logger != nil {
		s.Logger.WithFields(map[string]interface{}{
			"service": s.GetName(),
		}).Info("示例服务正在关闭")
	}

	// 调用基类的关闭方法
	return s.ServiceBase.Shutdown()
}
