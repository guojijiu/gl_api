package Middleware

import (
	"cloud-platform-api/app/Storage"
)

// BaseMiddleware 中间件基类
// 功能说明：
// 1. 提供中间件的通用功能
// 2. 包含存储管理器引用
// 3. 提供通用的日志记录方法
// 4. 支持中间件配置管理
type BaseMiddleware struct {
	storageManager *Storage.StorageManager
}

// SetStorageManager 设置存储管理器
func (m *BaseMiddleware) SetStorageManager(storageManager *Storage.StorageManager) {
	m.storageManager = storageManager
}

// GetStorageManager 获取存储管理器
func (m *BaseMiddleware) GetStorageManager() *Storage.StorageManager {
	return m.storageManager
}

// LogInfo 记录信息日志
func (m *BaseMiddleware) LogInfo(message string, data map[string]interface{}) {
	if m.storageManager != nil {
		m.storageManager.LogInfo(message, data)
	}
}

// LogWarning 记录警告日志
func (m *BaseMiddleware) LogWarning(message string, data map[string]interface{}) {
	if m.storageManager != nil {
		m.storageManager.LogWarning(message, data)
	}
}

// LogError 记录错误日志
func (m *BaseMiddleware) LogError(message string, data map[string]interface{}) {
	if m.storageManager != nil {
		m.storageManager.LogError(message, data)
	}
}

