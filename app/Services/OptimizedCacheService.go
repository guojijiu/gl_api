package Services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// OptimizedCacheService 优化的缓存服务
type OptimizedCacheService struct {
	*ServiceBase

	// 缓存存储
	memoryCache map[string]*OptimizedCacheItem
	mu          sync.RWMutex

	// 配置
	config *OptimizedCacheConfig

	// 统计信息
	stats *CacheStats

	// 清理协程控制
	ctx    context.Context
	cancel context.CancelFunc

	// 性能优化
	shards     []*OptimizedCacheShard
	shardCount int
}

// OptimizedCacheConfig 优化缓存配置
type OptimizedCacheConfig struct {
	MaxSize         int           `json:"max_size"`
	DefaultTTL      time.Duration `json:"default_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	ShardCount      int           `json:"shard_count"`
	EnableStats     bool          `json:"enable_stats"`
}

// OptimizedCacheItem 优化缓存项
type OptimizedCacheItem struct {
	Value       interface{}
	ExpiresAt   time.Time
	CreatedAt   time.Time
	AccessCount int64
	LastAccess  time.Time
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits      int64 `json:"hits"`
	Misses    int64 `json:"misses"`
	Sets      int64 `json:"sets"`
	Deletes   int64 `json:"deletes"`
	Evictions int64 `json:"evictions"`
	Size      int   `json:"size"`
	mu        sync.RWMutex
}

// OptimizedCacheShard 优化缓存分片
type OptimizedCacheShard struct {
	items map[string]*OptimizedCacheItem
	mu    sync.RWMutex
}

// NewOptimizedCacheService 创建优化的缓存服务
func NewOptimizedCacheService() *OptimizedCacheService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &OptimizedCacheService{
		ServiceBase: NewServiceBase("optimized_cache_service"),
		memoryCache: make(map[string]*OptimizedCacheItem),
		config: &OptimizedCacheConfig{
			MaxSize:         10000,
			DefaultTTL:      1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			ShardCount:      16,
			EnableStats:     true,
		},
		stats:  &CacheStats{},
		ctx:    ctx,
		cancel: cancel,
	}

	// 初始化分片
	service.initializeShards()

	// 启动清理协程
	go service.cleanupLoop()

	// 注册到全局服务管理器
	RegisterGlobalService("optimized_cache_service", service)

	return service
}

// initializeShards 初始化分片
func (s *OptimizedCacheService) initializeShards() {
	s.shardCount = s.config.ShardCount
	s.shards = make([]*OptimizedCacheShard, s.shardCount)

	for i := 0; i < s.shardCount; i++ {
		s.shards[i] = &OptimizedCacheShard{
			items: make(map[string]*OptimizedCacheItem),
		}
	}
}

// getShard 获取分片
func (s *OptimizedCacheService) getShard(key string) *OptimizedCacheShard {
	hash := s.hash(key)
	return s.shards[hash%uint32(s.shardCount)]
}

// hash 计算哈希值
func (s *OptimizedCacheService) hash(key string) uint32 {
	hash := uint32(0)
	for _, c := range key {
		hash = hash*31 + uint32(c)
	}
	return hash
}

// Set 设置缓存
func (s *OptimizedCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = s.config.DefaultTTL
	}

	item := &OptimizedCacheItem{
		Value:       value,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
	}

	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	// 检查是否需要清理空间
	if len(shard.items) >= s.config.MaxSize/s.shardCount {
		s.evictLRU(shard)
	}

	shard.items[key] = item

	// 更新统计
	if s.config.EnableStats {
		s.stats.mu.Lock()
		s.stats.Sets++
		s.stats.Size = s.getTotalSize()
		s.stats.mu.Unlock()
	}

	return nil
}

// Get 获取缓存
func (s *OptimizedCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	shard := s.getShard(key)
	shard.mu.RLock()
	item, exists := shard.items[key]
	shard.mu.RUnlock()

	if !exists {
		// 更新统计
		if s.config.EnableStats {
			s.stats.mu.Lock()
			s.stats.Misses++
			s.stats.mu.Unlock()
		}
		return nil, fmt.Errorf("key not found: %s", key)
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		// 删除过期项
		shard.mu.Lock()
		delete(shard.items, key)
		shard.mu.Unlock()

		// 更新统计
		if s.config.EnableStats {
			s.stats.mu.Lock()
			s.stats.Misses++
			s.stats.mu.Unlock()
		}
		return nil, fmt.Errorf("key expired: %s", key)
	}

	// 更新访问信息
	shard.mu.Lock()
	item.AccessCount++
	item.LastAccess = time.Now()
	shard.mu.Unlock()

	// 更新统计
	if s.config.EnableStats {
		s.stats.mu.Lock()
		s.stats.Hits++
		s.stats.mu.Unlock()
	}

	return item.Value, nil
}

// Delete 删除缓存
func (s *OptimizedCacheService) Delete(ctx context.Context, key string) error {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, exists := shard.items[key]; exists {
		delete(shard.items, key)

		// 更新统计
		if s.config.EnableStats {
			s.stats.mu.Lock()
			s.stats.Deletes++
			s.stats.Size = s.getTotalSize()
			s.stats.mu.Unlock()
		}
	}

	return nil
}

// Exists 检查键是否存在
func (s *OptimizedCacheService) Exists(ctx context.Context, key string) bool {
	shard := s.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	item, exists := shard.items[key]
	if !exists {
		return false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		return false
	}

	return true
}

// Clear 清空缓存
func (s *OptimizedCacheService) Clear(ctx context.Context) error {
	for _, shard := range s.shards {
		shard.mu.Lock()
		shard.items = make(map[string]*OptimizedCacheItem)
		shard.mu.Unlock()
	}

	// 更新统计
	if s.config.EnableStats {
		s.stats.mu.Lock()
		s.stats.Size = 0
		s.stats.mu.Unlock()
	}

	return nil
}

// evictLRU 淘汰最近最少使用的项
func (s *OptimizedCacheService) evictLRU(shard *OptimizedCacheShard) {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range shard.items {
		if oldestKey == "" || item.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.LastAccess
		}
	}

	if oldestKey != "" {
		delete(shard.items, oldestKey)

		// 更新统计
		if s.config.EnableStats {
			s.stats.mu.Lock()
			s.stats.Evictions++
			s.stats.mu.Unlock()
		}
	}
}

// cleanupLoop 清理循环
func (s *OptimizedCacheService) cleanupLoop() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupExpired()
		}
	}
}

// cleanupExpired 清理过期项
func (s *OptimizedCacheService) cleanupExpired() {
	now := time.Now()

	for _, shard := range s.shards {
		shard.mu.Lock()
		for key, item := range shard.items {
			if now.After(item.ExpiresAt) {
				delete(shard.items, key)
			}
		}
		shard.mu.Unlock()
	}

	// 更新统计
	if s.config.EnableStats {
		s.stats.mu.Lock()
		s.stats.Size = s.getTotalSize()
		s.stats.mu.Unlock()
	}
}

// getTotalSize 获取总大小
func (s *OptimizedCacheService) getTotalSize() int {
	total := 0
	for _, shard := range s.shards {
		shard.mu.RLock()
		total += len(shard.items)
		shard.mu.RUnlock()
	}
	return total
}

// GetStats 获取统计信息
func (s *OptimizedCacheService) GetStats() *CacheStats {
	if !s.config.EnableStats {
		return nil
	}

	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	// 返回统计信息的副本
	return &CacheStats{
		Hits:      s.stats.Hits,
		Misses:    s.stats.Misses,
		Sets:      s.stats.Sets,
		Deletes:   s.stats.Deletes,
		Evictions: s.stats.Evictions,
		Size:      s.stats.Size,
	}
}

// GetHitRate 获取命中率
func (s *OptimizedCacheService) GetHitRate() float64 {
	if !s.config.EnableStats {
		return 0
	}

	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	total := s.stats.Hits + s.stats.Misses
	if total == 0 {
		return 0
	}

	return float64(s.stats.Hits) / float64(total) * 100
}

// UpdateConfig 更新配置
func (s *OptimizedCacheService) UpdateConfig(config *OptimizedCacheConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config

	// 如果分片数量发生变化，重新初始化
	if config.ShardCount != s.shardCount {
		s.initializeShards()
	}
}

// GetConfig 获取配置
func (s *OptimizedCacheService) GetConfig() *OptimizedCacheConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config
}

// Stop 停止服务
func (s *OptimizedCacheService) Stop() error {
	s.cancel()
	return nil
}

// WarmCache 预热缓存
func (s *OptimizedCacheService) WarmCache() error {
	// 这里可以实现缓存预热逻辑
	// 例如：预加载用户列表、分类列表等
	return nil
}
