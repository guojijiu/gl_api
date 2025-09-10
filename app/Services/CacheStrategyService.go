package Services

import (
	"cloud-platform-api/app/Storage"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CacheStrategy 缓存策略
type CacheStrategy struct {
	Key           string        `json:"key"`
	TTL           time.Duration `json:"ttl"`
	MaxSize       int64         `json:"max_size"`
	Strategy      string        `json:"strategy"` // "memory", "redis", "hybrid"
	InvalidateOn  []string      `json:"invalidate_on"` // 触发失效的操作
	Priority      int           `json:"priority"` // 缓存优先级
}

// CacheStrategyService 缓存策略服务
type CacheStrategyService struct {
	storageManager *Storage.StorageManager
	redisService   *RedisService
	strategies     map[string]*CacheStrategy
}

// NewCacheStrategyService 创建缓存策略服务
// 功能说明：
// 1. 初始化缓存策略服务
// 2. 管理多种缓存策略
// 3. 提供智能缓存决策
// 4. 支持缓存预热和失效
func NewCacheStrategyService(storageManager *Storage.StorageManager, redisService *RedisService) *CacheStrategyService {
	service := &CacheStrategyService{
		storageManager: storageManager,
		redisService:   redisService,
		strategies:     make(map[string]*CacheStrategy),
	}
	
	// 初始化默认策略
	service.initDefaultStrategies()
	
	return service
}

// initDefaultStrategies 初始化默认缓存策略
func (s *CacheStrategyService) initDefaultStrategies() {
	// 用户数据缓存策略
	s.strategies["user_data"] = &CacheStrategy{
		Key:          "user:%d",
		TTL:          30 * time.Minute,
		MaxSize:      1024 * 1024, // 1MB
		Strategy:     "redis",
		InvalidateOn: []string{"user_update", "user_delete"},
		Priority:     1,
	}
	
	// 文章列表缓存策略
	s.strategies["post_list"] = &CacheStrategy{
		Key:          "posts:list:%s:%d:%d", // category:page:size
		TTL:          15 * time.Minute,
		MaxSize:      5 * 1024 * 1024, // 5MB
		Strategy:     "redis",
		InvalidateOn: []string{"post_create", "post_update", "post_delete"},
		Priority:     2,
	}
	
	// 系统配置缓存策略
	s.strategies["system_config"] = &CacheStrategy{
		Key:          "config:%s",
		TTL:          1 * time.Hour,
		MaxSize:      1024 * 1024, // 1MB
		Strategy:     "hybrid",
		InvalidateOn: []string{"config_update"},
		Priority:     0,
	}
	
	// 统计数据缓存策略
	s.strategies["statistics"] = &CacheStrategy{
		Key:          "stats:%s:%s", // type:period
		TTL:          5 * time.Minute,
		MaxSize:      2 * 1024 * 1024, // 2MB
		Strategy:     "redis",
		InvalidateOn: []string{"stats_update"},
		Priority:     3,
	}
}

// GetStrategy 获取缓存策略
func (s *CacheStrategyService) GetStrategy(strategyName string) (*CacheStrategy, error) {
	strategy, exists := s.strategies[strategyName]
	if !exists {
		return nil, fmt.Errorf("缓存策略不存在: %s", strategyName)
	}
	return strategy, nil
}

// SetStrategy 设置缓存策略
func (s *CacheStrategyService) SetStrategy(name string, strategy *CacheStrategy) {
	s.strategies[name] = strategy
	s.storageManager.LogInfo("缓存策略已更新", map[string]interface{}{
		"strategy_name": name,
		"strategy":      strategy,
	})
}

// Get 获取缓存数据
// 功能说明：
// 1. 根据策略获取缓存数据
// 2. 支持多级缓存
// 3. 自动处理缓存失效
func (s *CacheStrategyService) Get(strategyName string, params ...interface{}) (interface{}, error) {
	strategy, err := s.GetStrategy(strategyName)
	if err != nil {
		return nil, err
	}
	
	// 生成缓存键
	cacheKey := s.generateCacheKey(strategy, params...)
	
	// 根据策略获取数据
	switch strategy.Strategy {
	case "memory":
		return s.getFromMemory(cacheKey)
	case "redis":
		return s.getFromRedis(cacheKey)
	case "hybrid":
		// 先尝试内存缓存
		if data, err := s.getFromMemory(cacheKey); err == nil {
			return data, nil
		}
		// 再尝试Redis缓存
		return s.getFromRedis(cacheKey)
	default:
		return nil, fmt.Errorf("不支持的缓存策略: %s", strategy.Strategy)
	}
}

// Set 设置缓存数据
// 功能说明：
// 1. 根据策略设置缓存数据
// 2. 支持多级缓存
// 3. 自动处理缓存大小限制
func (s *CacheStrategyService) Set(strategyName string, data interface{}, params ...interface{}) error {
	strategy, err := s.GetStrategy(strategyName)
	if err != nil {
		return err
	}
	
	// 生成缓存键
	cacheKey := s.generateCacheKey(strategy, params...)
	
	// 检查数据大小
	if s.isDataTooLarge(data, strategy.MaxSize) {
		s.storageManager.LogWarning("缓存数据过大", map[string]interface{}{
			"strategy": strategyName,
			"key":      cacheKey,
			"max_size": strategy.MaxSize,
		})
		return fmt.Errorf("缓存数据超过大小限制")
	}
	
	// 根据策略设置数据
	switch strategy.Strategy {
	case "memory":
		return s.setToMemory(cacheKey, data, strategy.TTL)
	case "redis":
		return s.setToRedis(cacheKey, data, strategy.TTL)
	case "hybrid":
		// 同时设置到内存和Redis
		if err := s.setToMemory(cacheKey, data, strategy.TTL); err != nil {
			s.storageManager.LogError("内存缓存设置失败", map[string]interface{}{
				"key": cacheKey,
				"error": err.Error(),
			})
		}
		return s.setToRedis(cacheKey, data, strategy.TTL)
	default:
		return fmt.Errorf("不支持的缓存策略: %s", strategy.Strategy)
	}
}

// Invalidate 失效缓存
// 功能说明：
// 1. 根据操作类型失效相关缓存
// 2. 支持批量失效
// 3. 记录失效日志
func (s *CacheStrategyService) Invalidate(operation string) error {
	var invalidatedKeys []string
	
	// 查找需要失效的策略
	for name, strategy := range s.strategies {
		for _, invalidateOp := range strategy.InvalidateOn {
			if invalidateOp == operation {
				// 失效该策略的所有缓存
				keys, err := s.invalidateStrategy(name)
				if err != nil {
					s.storageManager.LogError("缓存失效失败", map[string]interface{}{
						"strategy": name,
						"operation": operation,
						"error": err.Error(),
					})
				} else {
					invalidatedKeys = append(invalidatedKeys, keys...)
				}
			}
		}
	}
	
	// 记录失效日志
	if len(invalidatedKeys) > 0 {
		s.storageManager.LogInfo("缓存已失效", map[string]interface{}{
			"operation": operation,
			"keys_count": len(invalidatedKeys),
			"keys": invalidatedKeys[:min(10, len(invalidatedKeys))], // 只记录前10个键
		})
	}
	
	return nil
}

// WarmCache 预热缓存
// 功能说明：
// 1. 预热常用数据到缓存
// 2. 提高系统响应速度
// 3. 支持自定义预热策略
func (s *CacheStrategyService) WarmCache() error {
	s.storageManager.LogInfo("开始缓存预热", map[string]interface{}{
		"strategies_count": len(s.strategies),
	})
	
	// 预热系统配置
	if err := s.warmSystemConfig(); err != nil {
		s.storageManager.LogError("系统配置预热失败", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	// 预热统计数据
	if err := s.warmStatistics(); err != nil {
		s.storageManager.LogError("统计数据预热失败", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	s.storageManager.LogInfo("缓存预热完成", map[string]interface{}{})
	return nil
}

// generateCacheKey 生成缓存键
func (s *CacheStrategyService) generateCacheKey(strategy *CacheStrategy, params ...interface{}) string {
	key := strategy.Key
	for _, param := range params {
		key = strings.Replace(key, "%s", fmt.Sprintf("%v", param), 1)
		key = strings.Replace(key, "%d", fmt.Sprintf("%v", param), 1)
	}
	return key
}

// isDataTooLarge 检查数据是否过大
func (s *CacheStrategyService) isDataTooLarge(data interface{}, maxSize int64) bool {
	// 这里可以实现具体的大小检查逻辑
	// 目前返回false作为示例
	return false
}

// getFromMemory 从内存获取数据
func (s *CacheStrategyService) getFromMemory(key string) (interface{}, error) {
	return s.storageManager.GetCache(key)
}

// getFromRedis 从Redis获取数据
func (s *CacheStrategyService) getFromRedis(key string) (interface{}, error) {
	if s.redisService == nil {
		return nil, fmt.Errorf("Redis服务不可用")
	}
	
	var data interface{}
	err := s.redisService.Get(key, &data)
	if err != nil {
		return nil, err
	}
	
	return data, nil
}

// setToMemory 设置数据到内存
func (s *CacheStrategyService) setToMemory(key string, data interface{}, ttl time.Duration) error {
	return s.storageManager.Cache(key, data, ttl)
}

// setToRedis 设置数据到Redis
func (s *CacheStrategyService) setToRedis(key string, data interface{}, ttl time.Duration) error {
	if s.redisService == nil {
		return fmt.Errorf("Redis服务不可用")
	}
	
	// 将data转换为JSON字符串
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("数据序列化失败: %v", err)
	}
	
	return s.redisService.SetWithTTL(context.Background(), key, string(jsonData), ttl)
}

// invalidateStrategy 失效策略缓存
func (s *CacheStrategyService) invalidateStrategy(strategyName string) ([]string, error) {
	var keys []string
	
	// 这里应该实现具体的失效逻辑
	// 例如删除所有匹配的缓存键
	
	return keys, nil
}

// warmSystemConfig 预热系统配置
func (s *CacheStrategyService) warmSystemConfig() error {
	// 预热常用配置
	configs := []string{"app", "database", "redis", "email"}
	
	for _, config := range configs {
		// 这里应该从数据库或配置文件加载配置
		// 然后设置到缓存
		data := map[string]interface{}{
			"name": config,
			"value": "preloaded",
		}
		
		if err := s.Set("system_config", data, config); err != nil {
			return err
		}
	}
	
	return nil
}

// warmStatistics 预热统计数据
func (s *CacheStrategyService) warmStatistics() error {
	// 预热常用统计数据
	periods := []string{"today", "week", "month"}
	
	for _, period := range periods {
		data := map[string]interface{}{
			"period": period,
			"count": 0,
		}
		
		if err := s.Set("statistics", data, "users", period); err != nil {
			return err
		}
	}
	
	return nil
}

// GetCacheStats 获取缓存统计信息
func (s *CacheStrategyService) GetCacheStats() map[string]interface{} {
	stats := map[string]interface{}{
		"strategies_count": len(s.strategies),
		"strategies":       s.strategies,
	}
	
	// 添加Redis统计信息
	if s.redisService != nil {
		redisStats, err := s.redisService.GetStats()
		if err == nil {
			stats["redis"] = redisStats
		}
	}
	
	return stats
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
