package Services

import (
	"cloud-platform-api/app/Storage"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheService 缓存服务
// 功能说明：
// 1. 提供内存缓存功能
// 2. 支持TTL（生存时间）管理
// 3. 提供缓存统计和监控
// 4. 支持缓存预热和清理
type CacheService struct {
	BaseService
	storageManager *Storage.StorageManager
	cache          map[string]*CacheItem
	mutex          sync.RWMutex
	stats          *CacheStats
	config         *CacheConfig
}

// CacheItem 缓存项
type CacheItem struct {
	Value       interface{} `json:"value"`
	ExpiresAt   time.Time   `json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
	AccessCount int64       `json:"access_count"`
}

// CacheStats 缓存统计
type CacheStats struct {
	HitCount    int64 `json:"hit_count"`    // 命中次数
	MissCount   int64 `json:"miss_count"`   // 未命中次数
	SetCount    int64 `json:"set_count"`    // 设置次数
	DeleteCount int64 `json:"delete_count"` // 删除次数
	EvictCount  int64 `json:"evict_count"`  // 淘汰次数
	TotalItems  int64 `json:"total_items"`  // 总项目数
	MemoryUsage int64 `json:"memory_usage"` // 内存使用量（字节）
}

// CacheConfig 缓存配置
type CacheConfig struct {
	MaxSize         int           `json:"max_size"`         // 最大缓存项数
	DefaultTTL      time.Duration `json:"default_ttl"`      // 默认TTL
	CleanupInterval time.Duration `json:"cleanup_interval"` // 清理间隔
	EnableStats     bool          `json:"enable_stats"`     // 启用统计
}

// NewCacheService 创建缓存服务
func NewCacheService(storageManager *Storage.StorageManager) *CacheService {
	config := &CacheConfig{
		MaxSize:         10000,            // 最大10000个缓存项
		DefaultTTL:      30 * time.Minute, // 默认30分钟TTL
		CleanupInterval: 5 * time.Minute,  // 每5分钟清理一次
		EnableStats:     true,             // 启用统计
	}

	service := &CacheService{
		BaseService:    *NewBaseService(),
		storageManager: storageManager,
		cache:          make(map[string]*CacheItem),
		stats:          &CacheStats{},
		config:         config,
	}

	// 启动清理协程
	go service.startCleanup()

	return service
}

// Set 设置缓存
func (c *CacheService) Set(key string, value interface{}, ttl ...time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查缓存大小限制
	if len(c.cache) >= c.config.MaxSize {
		c.evictOldest()
	}

	// 确定TTL
	duration := c.config.DefaultTTL
	if len(ttl) > 0 {
		duration = ttl[0]
	}

	// 创建缓存项
	item := &CacheItem{
		Value:       value,
		ExpiresAt:   time.Now().Add(duration),
		CreatedAt:   time.Now(),
		AccessCount: 0,
	}

	c.cache[key] = item

	// 更新统计
	if c.config.EnableStats {
		c.stats.SetCount++
		c.stats.TotalItems = int64(len(c.cache))
	}

	return nil
}

// Get 获取缓存
func (c *CacheService) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	item, exists := c.cache[key]
	c.mutex.RUnlock()

	if !exists {
		if c.config.EnableStats {
			c.mutex.Lock()
			c.stats.MissCount++
			c.mutex.Unlock()
		}
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		c.Delete(key)
		if c.config.EnableStats {
			c.mutex.Lock()
			c.stats.MissCount++
			c.mutex.Unlock()
		}
		return nil, false
	}

	// 更新访问计数
	c.mutex.Lock()
	item.AccessCount++
	if c.config.EnableStats {
		c.stats.HitCount++
	}
	c.mutex.Unlock()

	return item.Value, true
}

// Delete 删除缓存
func (c *CacheService) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.cache[key]; exists {
		delete(c.cache, key)
		if c.config.EnableStats {
			c.stats.DeleteCount++
			c.stats.TotalItems = int64(len(c.cache))
		}
	}
}

// Clear 清空所有缓存
func (c *CacheService) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*CacheItem)
	if c.config.EnableStats {
		c.stats.TotalItems = 0
	}
}

// GetStats 获取缓存统计
func (c *CacheService) GetStats() *CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// 创建统计副本
	stats := *c.stats
	stats.TotalItems = int64(len(c.cache))

	return &stats
}

// GetHitRate 获取命中率
func (c *CacheService) GetHitRate() float64 {
	stats := c.GetStats()
	total := stats.HitCount + stats.MissCount
	if total == 0 {
		return 0
	}
	return float64(stats.HitCount) / float64(total)
}

// evictOldest 淘汰最旧的缓存项
func (c *CacheService) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.cache {
		if oldestKey == "" || item.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
		if c.config.EnableStats {
			c.stats.EvictCount++
		}
	}
}

// startCleanup 启动清理协程
func (c *CacheService) startCleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup 清理过期缓存
func (c *CacheService) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, item := range c.cache {
		if now.After(item.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(c.cache, key)
		if c.config.EnableStats {
			c.stats.EvictCount++
		}
	}

	if len(expiredKeys) > 0 {
		c.stats.TotalItems = int64(len(c.cache))
		c.storageManager.LogInfo("缓存清理完成", map[string]interface{}{
			"expired_count":   len(expiredKeys),
			"remaining_count": len(c.cache),
		})
	}
}

// SetWithJSON 使用JSON序列化设置缓存
func (c *CacheService) SetWithJSON(key string, value interface{}, ttl ...time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %v", err)
	}

	return c.Set(key, jsonData, ttl...)
}

// GetWithJSON 使用JSON反序列化获取缓存
func (c *CacheService) GetWithJSON(key string, dest interface{}) error {
	value, exists := c.Get(key)
	if !exists {
		return fmt.Errorf("缓存未找到: %s", key)
	}

	jsonData, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("缓存数据类型错误")
	}

	return json.Unmarshal(jsonData, dest)
}

// Exists 检查缓存是否存在
func (c *CacheService) Exists(key string) bool {
	c.mutex.RLock()
	item, exists := c.cache[key]
	c.mutex.RUnlock()

	if !exists {
		return false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		c.Delete(key)
		return false
	}

	return true
}

// GetTTL 获取剩余TTL
func (c *CacheService) GetTTL(key string) time.Duration {
	c.mutex.RLock()
	item, exists := c.cache[key]
	c.mutex.RUnlock()

	if !exists {
		return 0
	}

	remaining := time.Until(item.ExpiresAt)
	if remaining <= 0 {
		c.Delete(key)
		return 0
	}

	return remaining
}

// SetConfig 更新缓存配置
func (c *CacheService) SetConfig(config *CacheConfig) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if config != nil {
		c.config = config
	}
}

// Keys 获取所有缓存键
func (c *CacheService) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	return keys
}

// Size 获取缓存项数量
func (c *CacheService) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

// WarmCache 预热缓存
func (c *CacheService) WarmCache() error {
	// 这里可以添加缓存预热逻辑
	// 例如：预加载常用数据、配置信息等
	c.storageManager.LogInfo("缓存预热开始", map[string]interface{}{
		"cache_size": c.Size(),
	})

	// 示例：预热一些基础数据
	// 实际项目中可以根据需要预热具体的数据

	c.storageManager.LogInfo("缓存预热完成", map[string]interface{}{
		"cache_size": c.Size(),
	})

	return nil
}

// HealthCheck 健康检查
func (c *CacheService) HealthCheck() map[string]interface{} {
	stats := c.GetStats()
	hitRate := c.GetHitRate()

	health := map[string]interface{}{
		"status":      "healthy",
		"hit_rate":    fmt.Sprintf("%.2f%%", hitRate*100),
		"total_items": stats.TotalItems,
		"hit_count":   stats.HitCount,
		"miss_count":  stats.MissCount,
	}

	// 判断健康状态
	if hitRate < 0.5 {
		health["status"] = "warning"
		health["message"] = "缓存命中率过低"
	} else if stats.TotalItems >= int64(c.config.MaxSize) {
		health["status"] = "warning"
		health["message"] = "缓存已满"
	}

	return health
}
