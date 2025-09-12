package Storage

import (
	"cloud-platform-api/app/Config"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// StorageManager 存储管理器
type StorageManager struct {
	config      *Config.StorageConfig
	services    map[string]StorageService
	mu          sync.RWMutex
	tempFiles   map[string]*os.File
	tempMu      sync.RWMutex
	tempService *TempService
	cache       map[string]interface{}
	cacheMu     sync.RWMutex
}

// NewStorageManager 创建存储管理器
func NewStorageManager(config *Config.StorageConfig) *StorageManager {
	sm := &StorageManager{
		config:    config,
		services:  make(map[string]StorageService),
		tempFiles: make(map[string]*os.File),
		cache:     make(map[string]interface{}),
	}

	// 初始化默认服务
	sm.initDefaultServices()

	// 启动清理任务
	go sm.startCleanupTask()

	return sm
}

// initDefaultServices 初始化默认服务
func (sm *StorageManager) initDefaultServices() {
	// 文件存储服务
	sm.services["file"] = NewFileStorage(sm.config.BasePath)

	// 可以添加其他存储服务，如云存储等
}

// GetService 获取存储服务
func (sm *StorageManager) GetService(name string) (StorageService, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	service, exists := sm.services[name]
	if !exists {
		return nil, fmt.Errorf("存储服务不存在: %s", name)
	}

	return service, nil
}

// RegisterService 注册存储服务
func (sm *StorageManager) RegisterService(name string, service StorageService) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.services[name] = service
}

// StoreFile 存储文件
func (sm *StorageManager) StoreFile(file io.Reader, filename string, path string, serviceName string) (string, error) {
	service, err := sm.GetService(serviceName)
	if err != nil {
		return "", err
	}

	return service.Store(file, filename, path)
}

// GetFile 获取文件
func (sm *StorageManager) GetFile(path string, serviceName string) (io.ReadCloser, error) {
	service, err := sm.GetService(serviceName)
	if err != nil {
		return nil, err
	}

	return service.Get(path)
}

// DeleteFile 删除文件
func (sm *StorageManager) DeleteFile(path string, serviceName string) error {
	service, err := sm.GetService(serviceName)
	if err != nil {
		return err
	}

	return service.Delete(path)
}

// FileExists 检查文件是否存在
func (sm *StorageManager) FileExists(path string, serviceName string) bool {
	service, err := sm.GetService(serviceName)
	if err != nil {
		return false
	}

	return service.Exists(path)
}

// GetFileSize 获取文件大小
func (sm *StorageManager) GetFileSize(path string, serviceName string) (int64, error) {
	service, err := sm.GetService(serviceName)
	if err != nil {
		return 0, err
	}

	return service.Size(path)
}

// CreateTempFile 创建临时文件
func (sm *StorageManager) CreateTempFile(prefix string) (*os.File, error) {
	service, err := sm.GetService("file")
	if err != nil {
		return nil, err
	}

	file, err := service.CreateTempFile(prefix)
	if err != nil {
		return nil, err
	}

	// 记录临时文件
	sm.tempMu.Lock()
	sm.tempFiles[file.Name()] = file
	sm.tempMu.Unlock()

	return file, nil
}

// CleanTempFiles 清理临时文件
func (sm *StorageManager) CleanTempFiles() error {
	sm.tempMu.Lock()
	defer sm.tempMu.Unlock()

	var errors []error

	for name, file := range sm.tempFiles {
		if err := file.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭临时文件失败 %s: %v", name, err))
		}

		if err := os.Remove(name); err != nil {
			errors = append(errors, fmt.Errorf("删除临时文件失败 %s: %v", name, err))
		}

		delete(sm.tempFiles, name)
	}

	if len(errors) > 0 {
		return fmt.Errorf("清理临时文件时发生错误: %v", errors)
	}

	return nil
}

// startCleanupTask 启动清理任务
func (sm *StorageManager) startCleanupTask() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.cleanup()
		}
	}
}

// cleanup 执行清理任务
func (sm *StorageManager) cleanup() {
	// 清理临时文件
	if err := sm.CleanTempFiles(); err != nil {
		fmt.Printf("清理临时文件失败: %v\n", err)
	}

	// 清理各服务的临时文件
	sm.mu.RLock()
	for _, service := range sm.services {
		if err := service.CleanTempFiles(); err != nil {
			fmt.Printf("清理服务临时文件失败: %v\n", err)
		}
	}
	sm.mu.RUnlock()
}

// GetStorageStats 获取存储统计信息
func (sm *StorageManager) GetStorageStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 统计文件数量
	fileCount := 0
	fileSize := int64(0)

	sm.mu.RLock()
	for name, service := range sm.services {
		serviceStats := make(map[string]interface{})

		// 这里可以添加具体的统计逻辑
		serviceStats["name"] = name
		serviceStats["type"] = "file"

		// 使用 service 变量避免未使用警告
		_ = service

		stats[name] = serviceStats
	}
	sm.mu.RUnlock()

	stats["total_services"] = len(sm.services)
	stats["file_count"] = fileCount
	stats["total_size"] = fileSize

	return stats
}

// Close 关闭存储管理器
func (sm *StorageManager) Close() error {
	// 清理所有临时文件
	return sm.CleanTempFiles()
}

// LogInfo 记录信息日志
func (sm *StorageManager) LogInfo(message string, context map[string]interface{}) {
	// 这里可以集成日志服务
	fmt.Printf("[INFO] %s: %v\n", message, context)
}

// LogError 记录错误日志
func (sm *StorageManager) LogError(message string, context map[string]interface{}) {
	// 这里可以集成日志服务
	fmt.Printf("[ERROR] %s: %v\n", message, context)
}

// LogWarning 记录警告日志
func (sm *StorageManager) LogWarning(message string, context map[string]interface{}) {
	// 这里可以集成日志服务
	fmt.Printf("[WARNING] %s: %v\n", message, context)
}

// StorePublic 存储公共文件
func (sm *StorageManager) StorePublic(file io.Reader, filename string, path string) (string, error) {
	return sm.StoreFile(file, filename, path, "file")
}

// StorePrivate 存储私有文件
func (sm *StorageManager) StorePrivate(file io.Reader, filename string, path string) (string, error) {
	return sm.StoreFile(file, filename, path, "file")
}

// FileStorage 获取文件存储服务
func (sm *StorageManager) FileStorage() StorageService {
	service, _ := sm.GetService("file")
	return service
}

// LogService 获取日志服务
func (sm *StorageManager) LogService() *LogService {
	// 这里应该返回实际的日志服务实例
	return nil
}

// GetStorageInfo 获取存储信息
func (sm *StorageManager) GetStorageInfo() map[string]interface{} {
	return sm.GetStorageStats()
}

// CheckHealth 检查存储健康状态
func (sm *StorageManager) CheckHealth() error {
	// 检查存储服务是否可用
	_, err := sm.GetService("file")
	return err
}

// Cache 缓存数据
func (sm *StorageManager) Cache(key string, value interface{}, ttl time.Duration) error {
	sm.cacheMu.Lock()
	defer sm.cacheMu.Unlock()
	sm.cache[key] = value
	return nil
}

// GetCache 获取缓存数据
func (sm *StorageManager) GetCache(key string) (interface{}, error) {
	sm.cacheMu.RLock()
	defer sm.cacheMu.RUnlock()
	value, exists := sm.cache[key]
	if !exists {
		return nil, nil
	}
	return value, nil
}

// DeleteCache 删除缓存数据
func (sm *StorageManager) DeleteCache(key string) error {
	sm.cacheMu.Lock()
	defer sm.cacheMu.Unlock()
	delete(sm.cache, key)
	return nil
}

// ClearCache 清空缓存
func (sm *StorageManager) ClearCache() error {
	sm.cacheMu.Lock()
	defer sm.cacheMu.Unlock()
	sm.cache = make(map[string]interface{})
	return nil
}

// TempService 获取临时文件服务
func (sm *StorageManager) TempService() *TempService {
	if sm.tempService == nil {
		sm.tempService = NewTempService(sm.config)
	}
	return sm.tempService
}

// BasePath 获取基础路径
func (sm *StorageManager) BasePath() string {
	return sm.config.BasePath
}

// GetStorageManager 获取全局存储管理器实例
var globalStorageManager *StorageManager

// SetGlobalStorageManager 设置全局存储管理器实例
func SetGlobalStorageManager(sm *StorageManager) {
	globalStorageManager = sm
}

// GetStorageManager 获取全局存储管理器实例
func GetStorageManager() *StorageManager {
	if globalStorageManager == nil {
		// 如果没有设置全局实例，尝试从配置创建
		config := Config.GetStorageConfig()
		if config != nil {
			globalStorageManager = NewStorageManager(config)
		}
	}
	return globalStorageManager
}

// GetFileList 获取文件列表
func (sm *StorageManager) GetFileList(path string) ([]map[string]interface{}, error) {
	_, err := sm.GetService("file")
	if err != nil {
		return nil, err
	}

	// 这里应该实现文件列表获取逻辑
	// 暂时返回空列表
	return []map[string]interface{}{}, nil
}

// GetFileInfo 获取文件信息
func (sm *StorageManager) GetFileInfo(path string) (map[string]interface{}, error) {
	service, err := sm.GetService("file")
	if err != nil {
		return nil, err
	}

	// 这里应该实现文件信息获取逻辑
	// 暂时返回基本信息
	info := map[string]interface{}{
		"path":   path,
		"exists": service.Exists(path),
	}

	if size, err := service.Size(path); err == nil {
		info["size"] = size
	}

	return info, nil
}

// CleanupLogs 清理日志
func (sm *StorageManager) CleanupLogs() error {
	// 这里应该实现日志清理逻辑
	return nil
}

// GetLogStats 获取日志统计
func (sm *StorageManager) GetLogStats() map[string]interface{} {
	// 这里应该实现日志统计逻辑
	return map[string]interface{}{
		"total_logs": 0,
		"log_levels": []string{"debug", "info", "warning", "error", "fatal"},
	}
}
