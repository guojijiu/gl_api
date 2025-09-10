package Services

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
)

var globalSecurityService *SecurityService

// Service 基础服务接口
type Service interface {
	// 可以在这里定义通用的服务方法
}

// BaseService 基础服务结构
type BaseService struct {
	// 可以在这里添加通用的服务依赖
}

// NewBaseService 创建基础服务
func NewBaseService() *BaseService {
	return &BaseService{}
}

// GetSecurityService 获取全局安全服务实例
func GetSecurityService() *SecurityService {
	if globalSecurityService == nil {
		db := Database.GetDB()
		config := &Config.SecurityConfig{}
		config.SetDefaults()
		globalSecurityService = NewSecurityService(db, config)
	}
	return globalSecurityService
}
