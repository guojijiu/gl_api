package Storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// StorageService 存储服务接口
type StorageService interface {
	// 文件操作
	Store(file io.Reader, filename string, path string) (string, error)
	Get(path string) (io.ReadCloser, error)
	Delete(path string) error
	Exists(path string) bool
	Size(path string) (int64, error)
	
	// 日志操作
	Log(level string, message string, context map[string]interface{}) error
	GetLogs(level string, date string) ([]string, error)
	
	// 缓存操作
	Cache(key string, value interface{}, ttl time.Duration) error
	GetCache(key string) (interface{}, error)
	DeleteCache(key string) error
	ClearCache() error
	
	// 临时文件操作
	CreateTempFile(prefix string) (*os.File, error)
	CleanTempFiles() error
}

// FileStorage 文件存储实现
type FileStorage struct {
	basePath string
}

// NewFileStorage 创建新的文件存储实例
func NewFileStorage(basePath string) *FileStorage {
	return &FileStorage{
		basePath: basePath,
	}
}

// Store 存储文件
func (fs *FileStorage) Store(file io.Reader, filename string, path string) (string, error) {
	// 构建完整路径，path应该包含app/public或app/private等子目录
	fullPath := filepath.Join(fs.basePath, path)
	
	// 确保目录存在
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}
	
	filePath := filepath.Join(fullPath, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()
	
	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("复制文件内容失败: %v", err)
	}
	
	return filePath, nil
}

// Get 获取文件
func (fs *FileStorage) Get(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(fs.basePath, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	return file, nil
}

// Delete 删除文件
func (fs *FileStorage) Delete(path string) error {
	fullPath := filepath.Join(fs.basePath, path)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}
	return nil
}

// Exists 检查文件是否存在
func (fs *FileStorage) Exists(path string) bool {
	fullPath := filepath.Join(fs.basePath, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// Size 获取文件大小
func (fs *FileStorage) Size(path string) (int64, error) {
	fullPath := filepath.Join(fs.basePath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("获取文件信息失败: %v", err)
	}
	return info.Size(), nil
}
