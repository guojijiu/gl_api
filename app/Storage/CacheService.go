package Storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Value      interface{} `json:"value"`
	Expiration time.Time   `json:"expiration"`
}

// CacheService 缓存服务
type CacheService struct {
	cachePath string
	memory    map[string]CacheItem
	mutex     sync.RWMutex
}

// NewCacheService 创建新的缓存服务实例
func NewCacheService(cachePath string) *CacheService {
	cs := &CacheService{
		cachePath: cachePath,
		memory:    make(map[string]CacheItem),
	}
	
	// 启动清理过期缓存的goroutine
	go cs.cleanupExpiredCache()
	
	return cs
}

// Cache 设置缓存
func (cs *CacheService) Cache(key string, value interface{}, ttl time.Duration) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	expiration := time.Now().Add(ttl)
	item := CacheItem{
		Value:      value,
		Expiration: expiration,
	}
	
	// 内存缓存
	cs.memory[key] = item
	
	// 持久化到文件
	return cs.persistToFile(key, item)
}

// GetCache 获取缓存
func (cs *CacheService) GetCache(key string) (interface{}, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	
	// 先从内存获取
	if item, exists := cs.memory[key]; exists {
		if time.Now().Before(item.Expiration) {
			return item.Value, nil
		}
		// 过期了，从内存删除
		delete(cs.memory, key)
	}
	
	// 从文件获取
	item, err := cs.loadFromFile(key)
	if err != nil {
		return nil, err
	}
	
	// 检查是否过期
	if time.Now().After(item.Expiration) {
		cs.DeleteCache(key)
		return nil, fmt.Errorf("缓存已过期")
	}
	
	// 加载到内存
	cs.memory[key] = item
	return item.Value, nil
}

// DeleteCache 删除缓存
func (cs *CacheService) DeleteCache(key string) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	// 从内存删除
	delete(cs.memory, key)
	
	// 从文件删除
	return cs.deleteFromFile(key)
}

// ClearCache 清空所有缓存
func (cs *CacheService) ClearCache() error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	// 清空内存缓存
	cs.memory = make(map[string]CacheItem)
	
	// 清空文件缓存
	return cs.clearFileCache()
}

// persistToFile 持久化缓存到文件
func (cs *CacheService) persistToFile(key string, item CacheItem) error {
	// 确保缓存目录存在
	if err := os.MkdirAll(cs.cachePath, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %v", err)
	}
	
	// 序列化缓存项
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("序列化缓存失败: %v", err)
	}
	
	// 写入文件
	cacheFile := filepath.Join(cs.cachePath, key+".cache")
	return os.WriteFile(cacheFile, data, 0644)
}

// loadFromFile 从文件加载缓存
func (cs *CacheService) loadFromFile(key string) (CacheItem, error) {
	var item CacheItem
	
	cacheFile := filepath.Join(cs.cachePath, key+".cache")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return item, fmt.Errorf("读取缓存文件失败: %v", err)
	}
	
	if err := json.Unmarshal(data, &item); err != nil {
		return item, fmt.Errorf("反序列化缓存失败: %v", err)
	}
	
	return item, nil
}

// deleteFromFile 从文件删除缓存
func (cs *CacheService) deleteFromFile(key string) error {
	cacheFile := filepath.Join(cs.cachePath, key+".cache")
	return os.Remove(cacheFile)
}

// clearFileCache 清空文件缓存
func (cs *CacheService) clearFileCache() error {
	// 删除所有.cache文件
	files, err := filepath.Glob(filepath.Join(cs.cachePath, "*.cache"))
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	
	return nil
}

// cleanupExpiredCache 清理过期的缓存
// 功能说明：
// 1. 定期清理过期的缓存项
// 2. 每5分钟执行一次清理
// 3. 同时清理内存和文件中的过期缓存
// 4. 防止内存泄漏和磁盘空间浪费
func (cs *CacheService) cleanupExpiredCache() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()
	
	for range ticker.C {
		cs.mutex.Lock()
		now := time.Now()
		
		// 清理内存中过期的缓存
		for key, item := range cs.memory {
			if now.After(item.Expiration) {
				delete(cs.memory, key)
				// 忽略文件删除错误，避免影响清理流程
				_ = cs.deleteFromFile(key)
			}
		}
		
		cs.mutex.Unlock()
	}
}

// GetCacheCount 获取缓存数量
func (cs *CacheService) GetCacheCount() int {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	return len(cs.memory)
}
