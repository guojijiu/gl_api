package Storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TempService 临时文件服务
type TempService struct {
	tempPath string
}

// NewTempService 创建新的临时文件服务实例
func NewTempService(tempPath string) *TempService {
	return &TempService{
		tempPath: tempPath,
	}
}

// CreateTempFile 创建临时文件
func (ts *TempService) CreateTempFile(prefix string) (*os.File, error) {
	// 确保临时目录存在
	if err := os.MkdirAll(ts.tempPath, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	
	// 生成唯一的文件名
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%s_%d.tmp", prefix, timestamp)
	filePath := filepath.Join(ts.tempPath, filename)
	
	// 创建临时文件
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %v", err)
	}
	
	return file, nil
}

// CreateTempFileWithExtension 创建带扩展名的临时文件
func (ts *TempService) CreateTempFileWithExtension(prefix string, extension string) (*os.File, error) {
	// 确保临时目录存在
	if err := os.MkdirAll(ts.tempPath, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	
	// 清理扩展名
	extension = strings.TrimPrefix(extension, ".")
	
	// 生成唯一的文件名
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%s_%d.%s", prefix, timestamp, extension)
	filePath := filepath.Join(ts.tempPath, filename)
	
	// 创建临时文件
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %v", err)
	}
	
	return file, nil
}

// CleanTempFiles 清理临时文件
func (ts *TempService) CleanTempFiles() error {
	// 获取所有临时文件
	files, err := filepath.Glob(filepath.Join(ts.tempPath, "*.tmp"))
	if err != nil {
		return fmt.Errorf("获取临时文件列表失败: %v", err)
	}
	
	// 删除所有临时文件
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("删除临时文件失败 %s: %v", file, err)
		}
	}
	
	return nil
}

// CleanTempFilesOlderThan 清理指定时间之前的临时文件
func (ts *TempService) CleanTempFilesOlderThan(age time.Duration) error {
	// 获取所有临时文件
	files, err := filepath.Glob(filepath.Join(ts.tempPath, "*.tmp"))
	if err != nil {
		return fmt.Errorf("获取临时文件列表失败: %v", err)
	}
	
	cutoffTime := time.Now().Add(-age)
	
	// 删除过期的临时文件
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue // 跳过无法获取信息的文件
		}
		
		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(file); err != nil {
				return fmt.Errorf("删除过期临时文件失败 %s: %v", file, err)
			}
		}
	}
	
	return nil
}

// GetTempFileInfo 获取临时文件信息
func (ts *TempService) GetTempFileInfo() (int, int64, error) {
	// 获取所有临时文件
	files, err := filepath.Glob(filepath.Join(ts.tempPath, "*.tmp"))
	if err != nil {
		return 0, 0, fmt.Errorf("获取临时文件列表失败: %v", err)
	}
	
	var totalSize int64
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		totalSize += info.Size()
	}
	
	return len(files), totalSize, nil
}

// CleanTempFilesByPattern 按模式清理临时文件
func (ts *TempService) CleanTempFilesByPattern(pattern string) error {
	// 构建完整的文件模式
	fullPattern := filepath.Join(ts.tempPath, pattern)
	
	// 获取匹配的文件
	files, err := filepath.Glob(fullPattern)
	if err != nil {
		return fmt.Errorf("获取匹配文件列表失败: %v", err)
	}
	
	// 删除匹配的文件
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("删除匹配文件失败 %s: %v", file, err)
		}
	}
	
	return nil
}
