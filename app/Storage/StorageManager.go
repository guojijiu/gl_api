package Storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

var globalStorageManager *StorageManager

// GetStorageManager 获取全局存储管理器实例
func GetStorageManager() *StorageManager {
	if globalStorageManager == nil {
		// 使用当前工作目录作为基础路径
		wd, _ := os.Getwd()
		globalStorageManager = NewStorageManager(wd)
	}
	return globalStorageManager
}

// StorageManager 存储管理器
// 功能说明：
// 1. 统一管理文件存储、日志、缓存、临时文件等服务
// 2. 提供公共和私有文件存储功能
// 3. 集成日志记录、缓存管理、临时文件管理
// 4. 提供存储信息查询和文件管理功能
type StorageManager struct {
	FileStorage  *FileStorage
	LogService   *LogService
	CacheService *CacheService
	TempService  *TempService

	// 存储路径配置
	BasePath    string
	PublicPath  string
	PrivatePath string
	LogPath     string
	CachePath   string
	TempPath    string
}

// NewStorageManager 创建新的存储管理器
// 功能说明：
// 1. 初始化各种存储路径（公共、私有、日志、缓存、临时）
// 2. 创建各种服务实例（文件存储、日志、缓存、临时文件）
// 3. 设置基础路径配置
func NewStorageManager(basePath string) *StorageManager {
	// 构建各种存储路径
	publicPath := filepath.Join(basePath, "app", "public")
	privatePath := filepath.Join(basePath, "app", "private")
	logPath := filepath.Join(basePath, "logs")
	cachePath := filepath.Join(basePath, "framework", "cache")
	tempPath := filepath.Join(basePath, "temp")

	// 创建存储管理器
	sm := &StorageManager{
		BasePath:    basePath,
		PublicPath:  publicPath,
		PrivatePath: privatePath,
		LogPath:     logPath,
		CachePath:   cachePath,
		TempPath:    tempPath,
	}

	// 初始化各种服务
	sm.FileStorage = NewFileStorage(basePath)
	sm.LogService = NewLogService(logPath)
	sm.CacheService = NewCacheService(cachePath)
	sm.TempService = NewTempService(tempPath)

	return sm
}

// StorePublic 存储到公共目录
// 功能说明：
// 1. 支持multipart.FileHeader和io.Reader两种文件类型
// 2. 将文件存储到app/public目录下
// 3. 返回存储后的文件路径
// 4. 自动处理文件打开和关闭
func (sm *StorageManager) StorePublic(file interface{}, filename string, path string) (string, error) {
	// 构建公共目录路径
	storagePath := filepath.Join("app", "public", path)

	// 处理multipart.FileHeader类型
	if multipartFile, ok := file.(*multipart.FileHeader); ok {
		src, err := multipartFile.Open()
		if err != nil {
			return "", fmt.Errorf("打开上传文件失败: %v", err)
		}
		defer src.Close()

		// 使用FileStorage存储文件
		return sm.FileStorage.Store(src, filename, storagePath)
	}

	// 处理io.Reader类型
	if reader, ok := file.(io.Reader); ok {
		return sm.FileStorage.Store(reader, filename, storagePath)
	}

	return "", fmt.Errorf("不支持的文件类型: %T", file)
}

// StorePrivate 存储到私有目录
// 功能说明：
// 1. 支持multipart.FileHeader和io.Reader两种文件类型
// 2. 将文件存储到app/private目录下
// 3. 返回存储后的文件路径
// 4. 自动处理文件打开和关闭
func (sm *StorageManager) StorePrivate(file interface{}, filename string, path string) (string, error) {
	// 构建私有目录路径
	storagePath := filepath.Join("app", "private", path)

	// 处理multipart.FileHeader类型
	if multipartFile, ok := file.(*multipart.FileHeader); ok {
		src, err := multipartFile.Open()
		if err != nil {
			return "", fmt.Errorf("打开上传文件失败: %v", err)
		}
		defer src.Close()

		// 使用FileStorage存储文件到私有目录
		return sm.FileStorage.Store(src, filename, storagePath)
	}

	// 处理io.Reader类型
	if reader, ok := file.(io.Reader); ok {
		return sm.FileStorage.Store(reader, filename, storagePath)
	}

	return "", fmt.Errorf("不支持的文件类型: %T", file)
}

// GetPublicURL 获取公共文件的URL
func (sm *StorageManager) GetPublicURL(path string) string {
	// 这里可以根据配置返回完整的URL
	return "/storage/app/public/" + path
}

// GetPrivateURL 获取私有文件的URL（需要认证）
func (sm *StorageManager) GetPrivateURL(path string) string {
	// 这里可以根据配置返回完整的URL
	return "/storage/app/private/" + path
}

// LogInfo 记录信息日志
func (sm *StorageManager) LogInfo(message string, context map[string]interface{}) error {
	return sm.LogService.LogInfo(message, context)
}

// LogWarning 记录警告日志
func (sm *StorageManager) LogWarning(message string, context map[string]interface{}) error {
	return sm.LogService.LogWarning(message, context)
}

// LogError 记录错误日志
func (sm *StorageManager) LogError(message string, context map[string]interface{}) error {
	return sm.LogService.LogError(message, context)
}

// LogDebug 记录调试日志
func (sm *StorageManager) LogDebug(message string, context map[string]interface{}) error {
	return sm.LogService.LogDebug(message, context)
}

// Cache 设置缓存
func (sm *StorageManager) Cache(key string, value interface{}, ttl time.Duration) error {
	return sm.CacheService.Cache(key, value, ttl)
}

// GetCache 获取缓存
func (sm *StorageManager) GetCache(key string) (interface{}, error) {
	return sm.CacheService.GetCache(key)
}

// DeleteCache 删除缓存
func (sm *StorageManager) DeleteCache(key string) error {
	return sm.CacheService.DeleteCache(key)
}

// ClearCache 清空缓存
func (sm *StorageManager) ClearCache() error {
	return sm.CacheService.ClearCache()
}

// CreateTempFile 创建临时文件
func (sm *StorageManager) CreateTempFile(prefix string) (*os.File, error) {
	return sm.TempService.CreateTempFile(prefix)
}

// CreateTempFileWithExtension 创建带扩展名的临时文件
func (sm *StorageManager) CreateTempFileWithExtension(prefix string, extension string) (*os.File, error) {
	return sm.TempService.CreateTempFileWithExtension(prefix, extension)
}

// CleanTempFiles 清理临时文件
func (sm *StorageManager) CleanTempFiles() error {
	return sm.TempService.CleanTempFiles()
}

// GetStorageInfo 获取存储信息
// 功能说明：
// 1. 获取各种存储路径信息
// 2. 统计临时文件数量和大小
// 3. 统计缓存项目数量
// 4. 返回完整的存储系统状态
func (sm *StorageManager) GetStorageInfo() map[string]interface{} {
	// 获取临时文件信息
	tempCount, tempSize, _ := sm.TempService.GetTempFileInfo()

	// 获取缓存信息
	cacheCount := 0
	if sm.CacheService != nil {
		cacheCount = sm.CacheService.GetCacheCount()
	}

	return map[string]interface{}{
		"base_path":    sm.BasePath,
		"public_path":  sm.PublicPath,
		"private_path": sm.PrivatePath,
		"log_path":     sm.LogPath,
		"cache_path":   sm.CachePath,
		"temp_path":    sm.TempPath,
		"temp_files":   tempCount,
		"temp_size":    tempSize,
		"cache_items":  cacheCount,
	}
}

// GetFileList 获取指定目录下的文件列表
// 功能说明：
// 1. 支持公共和私有存储目录
// 2. 返回文件的详细信息（名称、路径、大小、修改时间等）
// 3. 区分文件和目录
// 4. 返回权限信息
func (sm *StorageManager) GetFileList(path string, storageType string) ([]map[string]interface{}, error) {
	var basePath string
	if storageType == "private" {
		basePath = sm.PrivatePath
	} else {
		basePath = sm.PublicPath
	}

	fullPath := filepath.Join(basePath, path)

	// 检查目录是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return []map[string]interface{}{}, nil
	}

	// 读取目录内容
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}

	var files []map[string]interface{}
	for _, entry := range entries {
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(path, entry.Name())
		files = append(files, map[string]interface{}{
			"name":        entry.Name(),
			"path":        filePath,
			"size":        fileInfo.Size(),
			"is_dir":      entry.IsDir(),
			"mod_time":    fileInfo.ModTime().Format("2006-01-02 15:04:05"),
			"permissions": fileInfo.Mode().String(),
		})
	}

	return files, nil
}

// GetFileInfo 获取文件详细信息
// 功能说明：
// 1. 获取单个文件的详细信息
// 2. 支持公共和私有存储
// 3. 返回文件大小、修改时间、权限等信息
// 4. 包含存储类型和完整路径信息
func (sm *StorageManager) GetFileInfo(filePath string, storageType string) (map[string]interface{}, error) {
	var basePath string
	if storageType == "private" {
		basePath = sm.PrivatePath
	} else {
		basePath = sm.PublicPath
	}

	fullPath := filepath.Join(basePath, filePath)

	// 检查文件是否存在
	if !sm.FileStorage.Exists(fullPath) {
		return nil, fmt.Errorf("文件不存在")
	}

	// 获取文件信息
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 获取文件大小
	size, err := sm.FileStorage.Size(fullPath)
	if err != nil {
		size = 0
	}

	return map[string]interface{}{
		"name":         info.Name(),
		"path":         filePath,
		"size":         size,
		"is_dir":       info.IsDir(),
		"mod_time":     info.ModTime().Format("2006-01-02 15:04:05"),
		"create_time":  info.ModTime().Format("2006-01-02 15:04:05"), // 注意：Go的os.Stat不提供创建时间
		"permissions":  info.Mode().String(),
		"storage_type": storageType,
		"full_path":    fullPath,
	}, nil
}

// CleanupLogs 清理过期和过多的日志文件
// 功能说明：
// 1. 清理超过指定天数的日志文件
// 2. 限制日志文件的总大小
// 3. 保留最近的日志文件
// 4. 返回清理的文件数量和大小
func (sm *StorageManager) CleanupLogs(maxDays int, maxSizeMB int64) (int, int64, error) {
	return sm.LogService.CleanupLogs(maxDays, maxSizeMB)
}

// GetLogStats 获取日志统计信息
// 功能说明：
// 1. 统计日志文件数量和总大小
// 2. 按级别统计日志数量
// 3. 获取最近的日志文件信息
// 4. 返回详细的统计信息
func (sm *StorageManager) GetLogStats() (map[string]interface{}, error) {
	return sm.LogService.GetLogStats()
}

// CheckHealth 检查存储管理器健康状态
// 功能说明：
// 1. 检查各个存储组件是否正常工作
// 2. 验证存储路径是否可访问
// 3. 检查服务实例是否已初始化
// 4. 返回健康状态和错误信息
func (sm *StorageManager) CheckHealth() error {
	// 检查基础路径是否存在
	if _, err := os.Stat(sm.BasePath); os.IsNotExist(err) {
		return fmt.Errorf("基础存储路径不存在: %s", sm.BasePath)
	}

	// 检查各个存储路径
	paths := []string{sm.PublicPath, sm.PrivatePath, sm.LogPath, sm.CachePath, sm.TempPath}
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// 尝试创建目录
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("无法创建存储路径 %s: %v", path, err)
			}
		}
	}

	// 检查各个服务是否已初始化
	if sm.FileStorage == nil {
		return fmt.Errorf("文件存储服务未初始化")
	}
	if sm.LogService == nil {
		return fmt.Errorf("日志服务未初始化")
	}
	if sm.CacheService == nil {
		return fmt.Errorf("缓存服务未初始化")
	}
	if sm.TempService == nil {
		return fmt.Errorf("临时文件服务未初始化")
	}

	return nil
}
