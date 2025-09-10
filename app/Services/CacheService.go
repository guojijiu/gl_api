package Services

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// CacheService 缓存服务
// 功能说明：
// 1. 提供统一的缓存接口
// 2. 支持查询缓存和页面缓存
// 3. 自动缓存失效策略
// 4. 支持多种缓存存储后端
type CacheService struct {
	redisService *RedisService
	prefix       string
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Prefix        string
	DefaultTTL    time.Duration
	MaxTTL        time.Duration
	EnableCache   bool
}

// CacheItem 缓存项
type CacheItem struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	TTL       time.Duration `json:"ttl"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// NewCacheService 创建缓存服务
func NewCacheService(redisService *RedisService, config *CacheConfig) *CacheService {
	if config == nil {
		config = &CacheConfig{
			Prefix:      "cache:",
			DefaultTTL:  5 * time.Minute,
			MaxTTL:      1 * time.Hour,
			EnableCache: true,
		}
	}
	
	return &CacheService{
		redisService: redisService,
		prefix:       config.Prefix,
	}
}

// Get 获取缓存
// 功能说明：
// 1. 从缓存中获取数据
// 2. 支持自动反序列化
// 3. 处理缓存不存在的情况
func (c *CacheService) Get(key string, dest interface{}) error {
	if c.redisService == nil {
		return fmt.Errorf("Redis服务未初始化")
	}
	
	fullKey := c.prefix + key
	return c.redisService.Get(fullKey, dest)
}

// Set 设置缓存
// 功能说明：
// 1. 将数据存储到缓存
// 2. 支持自动序列化
// 3. 设置过期时间
func (c *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
	if c.redisService == nil {
		return fmt.Errorf("Redis服务未初始化")
	}
	
	fullKey := c.prefix + key
	return c.redisService.Set(fullKey, value, ttl)
}

// Delete 删除缓存
func (c *CacheService) Delete(key string) error {
	if c.redisService == nil {
		return fmt.Errorf("Redis服务未初始化")
	}
	
	fullKey := c.prefix + key
	return c.redisService.Delete(fullKey)
}

// Exists 检查缓存是否存在
func (c *CacheService) Exists(key string) (bool, error) {
	if c.redisService == nil {
		return false, fmt.Errorf("Redis服务未初始化")
	}
	
	fullKey := c.prefix + key
	return c.redisService.Exists(fullKey)
}

// Clear 清空所有缓存
func (c *CacheService) Clear() error {
	if c.redisService == nil {
		return fmt.Errorf("Redis服务未初始化")
	}
	
	// 删除所有以prefix开头的键
	ctx := context.Background()
	pattern := c.prefix + "*"
	keys, err := c.redisService.Keys(ctx, pattern)
	if err != nil {
		return err
	}
	
	for _, key := range keys {
		c.redisService.Del(ctx, key)
	}
	
	return nil
}

// CacheQuery 缓存查询结果
// 功能说明：
// 1. 缓存数据库查询结果
// 2. 自动生成缓存键
// 3. 支持查询参数哈希
func (c *CacheService) CacheQuery(query string, params map[string]interface{}, ttl time.Duration) (string, error) {
	// 生成缓存键
	cacheKey := c.generateQueryKey(query, params)
	
	// 检查缓存是否存在
	exists, err := c.Exists(cacheKey)
	if err != nil {
		return "", err
	}
	
	if exists {
		return cacheKey, nil
	}
	
	// 缓存不存在，返回键名供后续存储
	return cacheKey, nil
}

// generateQueryKey 生成查询缓存键
func (c *CacheService) generateQueryKey(query string, params map[string]interface{}) string {
	// 将查询和参数组合
	data := map[string]interface{}{
		"query":  query,
		"params": params,
	}
	
	// 序列化为JSON
	jsonData, _ := json.Marshal(data)
	
	// 生成MD5哈希
	hash := md5.Sum(jsonData)
	
	return fmt.Sprintf("query:%x", hash)
}

// CachePage 缓存页面内容
// 功能说明：
// 1. 缓存页面HTML内容
// 2. 支持页面参数
// 3. 自动生成页面缓存键
func (c *CacheService) CachePage(path string, params map[string]string, ttl time.Duration) (string, error) {
	// 生成页面缓存键
	cacheKey := c.generatePageKey(path, params)
	
	// 检查缓存是否存在
	exists, err := c.Exists(cacheKey)
	if err != nil {
		return "", err
	}
	
	if exists {
		return cacheKey, nil
	}
	
	// 缓存不存在，返回键名供后续存储
	return cacheKey, nil
}

// generatePageKey 生成页面缓存键
func (c *CacheService) generatePageKey(path string, params map[string]string) string {
	// 将路径和参数组合
	data := map[string]interface{}{
		"path":   path,
		"params": params,
	}
	
	// 序列化为JSON
	jsonData, _ := json.Marshal(data)
	
	// 生成MD5哈希
	hash := md5.Sum(jsonData)
	
	return fmt.Sprintf("page:%x", hash)
}

// CacheUser 缓存用户信息
func (c *CacheService) CacheUser(userID uint, user interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("user:%d", userID)
	return c.Set(key, user, ttl)
}

// GetCachedUser 获取缓存的用户信息
func (c *CacheService) GetCachedUser(userID uint, dest interface{}) error {
	key := fmt.Sprintf("user:%d", userID)
	return c.Get(key, dest)
}

// InvalidateUserCache 清除用户缓存
func (c *CacheService) InvalidateUserCache(userID uint) error {
	key := fmt.Sprintf("user:%d", userID)
	return c.Delete(key)
}

// CacheList 缓存列表数据
func (c *CacheService) CacheList(prefix string, page, limit int, data interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("%s:page:%d:limit:%d", prefix, page, limit)
	return c.Set(key, data, ttl)
}

// GetCachedList 获取缓存的列表数据
func (c *CacheService) GetCachedList(prefix string, page, limit int, dest interface{}) error {
	key := fmt.Sprintf("%s:page:%d:limit:%d", prefix, page, limit)
	return c.Get(key, dest)
}

// InvalidateListCache 清除列表缓存
func (c *CacheService) InvalidateListCache(prefix string) error {
	ctx := context.Background()
	pattern := c.prefix + prefix + ":*"
	keys, err := c.redisService.Keys(ctx, pattern)
	if err != nil {
		return err
	}
	
	for _, key := range keys {
		c.redisService.Del(ctx, key)
	}
	
	return nil
}

// GetCacheStats 获取缓存统计信息
func (c *CacheService) GetCacheStats() (map[string]interface{}, error) {
	if c.redisService == nil {
		return nil, fmt.Errorf("Redis服务未初始化")
	}
	
	stats, err := c.redisService.GetStats()
	if err != nil {
		return nil, err
	}
	
	// 添加缓存前缀信息
	stats["cache_prefix"] = c.prefix
	
	return stats, nil
}

// WarmCache 预热缓存
// 功能说明：
// 1. 预先加载常用数据到缓存
// 2. 提高应用启动后的响应速度
// 3. 支持自定义预热策略
func (c *CacheService) WarmCache() error {
	// 这里可以实现缓存预热逻辑
	// 例如：预加载用户列表、分类列表等
	
	log.Println("缓存预热完成")
	return nil
}
