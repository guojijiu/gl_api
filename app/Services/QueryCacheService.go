package Services

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"cloud-platform-api/app/Models"
)

// QueryCacheService 查询缓存服务
type QueryCacheService struct {
	redisService *RedisService
}

// QueryKey 查询键
type QueryKey struct {
	Table    string                 `json:"table"`
	Where    map[string]interface{} `json:"where"`
	Order    string                 `json:"order"`
	Limit    int                    `json:"limit"`
	Offset   int                    `json:"offset"`
	Includes []string               `json:"includes"`
}

// QueryResult 查询结果
type QueryResult struct {
	Data      interface{} `json:"data"`
	Count     int64       `json:"count"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

// NewQueryCacheService 创建查询缓存服务
func NewQueryCacheService(redisService *RedisService) *QueryCacheService {
	return &QueryCacheService{
		redisService: redisService,
	}
}

// GenerateCacheKey 生成缓存键
func (qc *QueryCacheService) GenerateCacheKey(queryKey *QueryKey) string {
	data, _ := json.Marshal(queryKey)
	hash := md5.Sum(data)
	return fmt.Sprintf("query:%s:%x", queryKey.Table, hash)
}

// CacheQuery 缓存查询结果
func (qc *QueryCacheService) CacheQuery(queryKey *QueryKey, result *QueryResult, ttl time.Duration) error {
	cacheKey := qc.GenerateCacheKey(queryKey)
	
	result.Timestamp = time.Now()
	result.TTL = ttl
	
	return qc.redisService.Set(cacheKey, result, ttl)
}

// GetCachedQuery 获取缓存的查询结果
func (qc *QueryCacheService) GetCachedQuery(queryKey *QueryKey) (*QueryResult, error) {
	cacheKey := qc.GenerateCacheKey(queryKey)
	
	var result QueryResult
	err := qc.redisService.Get(cacheKey, &result)
	if err != nil {
		return nil, err
	}
	
	// 检查是否过期
	if time.Since(result.Timestamp) > result.TTL {
		qc.redisService.Delete(cacheKey)
		return nil, fmt.Errorf("缓存已过期")
	}
	
	return &result, nil
}

// InvalidateTableCache 清除表缓存
func (qc *QueryCacheService) InvalidateTableCache(tableName string) error {
	_ = fmt.Sprintf("query:%s:*", tableName)
	
	// 这里应该实现模式匹配删除
	// 暂时返回成功
	return nil
}

// CacheUsers 缓存用户列表
func (qc *QueryCacheService) CacheUsers(users []Models.User, ttl time.Duration) error {
	queryKey := &QueryKey{
		Table: "users",
	}
	
	result := &QueryResult{
		Data:  users,
		Count: int64(len(users)),
	}
	
	return qc.CacheQuery(queryKey, result, ttl)
}

// GetCachedUsers 获取缓存的用户列表
func (qc *QueryCacheService) GetCachedUsers() ([]Models.User, error) {
	queryKey := &QueryKey{
		Table: "users",
	}
	
	result, err := qc.GetCachedQuery(queryKey)
	if err != nil {
		return nil, err
	}
	
	if users, ok := result.Data.([]Models.User); ok {
		return users, nil
	}
	
	return nil, fmt.Errorf("缓存数据类型错误")
}

// CachePosts 缓存文章列表
func (qc *QueryCacheService) CachePosts(posts []Models.Post, ttl time.Duration) error {
	queryKey := &QueryKey{
		Table: "posts",
	}
	
	result := &QueryResult{
		Data:  posts,
		Count: int64(len(posts)),
	}
	
	return qc.CacheQuery(queryKey, result, ttl)
}

// GetCachedPosts 获取缓存的文章列表
func (qc *QueryCacheService) GetCachedPosts() ([]Models.Post, error) {
	queryKey := &QueryKey{
		Table: "posts",
	}
	
	result, err := qc.GetCachedQuery(queryKey)
	if err != nil {
		return nil, err
	}
	
	if posts, ok := result.Data.([]Models.Post); ok {
		return posts, nil
	}
	
	return nil, fmt.Errorf("缓存数据类型错误")
}

// CacheCategories 缓存分类列表
func (qc *QueryCacheService) CacheCategories(categories []Models.Category, ttl time.Duration) error {
	queryKey := &QueryKey{
		Table: "categories",
	}
	
	result := &QueryResult{
		Data:  categories,
		Count: int64(len(categories)),
	}
	
	return qc.CacheQuery(queryKey, result, ttl)
}

// GetCachedCategories 获取缓存的分类列表
func (qc *QueryCacheService) GetCachedCategories() ([]Models.Category, error) {
	queryKey := &QueryKey{
		Table: "categories",
	}
	
	result, err := qc.GetCachedQuery(queryKey)
	if err != nil {
		return nil, err
	}
	
	if categories, ok := result.Data.([]Models.Category); ok {
		return categories, nil
	}
	
	return nil, fmt.Errorf("缓存数据类型错误")
}

// CacheTags 缓存标签列表
func (qc *QueryCacheService) CacheTags(tags []Models.Tag, ttl time.Duration) error {
	queryKey := &QueryKey{
		Table: "tags",
	}
	
	result := &QueryResult{
		Data:  tags,
		Count: int64(len(tags)),
	}
	
	return qc.CacheQuery(queryKey, result, ttl)
}

// GetCachedTags 获取缓存的标签列表
func (qc *QueryCacheService) GetCachedTags() ([]Models.Tag, error) {
	queryKey := &QueryKey{
		Table: "tags",
	}
	
	result, err := qc.GetCachedQuery(queryKey)
	if err != nil {
		return nil, err
	}
	
	if tags, ok := result.Data.([]Models.Tag); ok {
		return tags, nil
	}
	
	return nil, fmt.Errorf("缓存数据类型错误")
}

// InvalidateUserCache 清除用户缓存
func (qc *QueryCacheService) InvalidateUserCache() error {
	return qc.InvalidateTableCache("users")
}

// InvalidatePostCache 清除文章缓存
func (qc *QueryCacheService) InvalidatePostCache() error {
	return qc.InvalidateTableCache("posts")
}

// InvalidateCategoryCache 清除分类缓存
func (qc *QueryCacheService) InvalidateCategoryCache() error {
	return qc.InvalidateTableCache("categories")
}

// InvalidateTagCache 清除标签缓存
func (qc *QueryCacheService) InvalidateTagCache() error {
	return qc.InvalidateTableCache("tags")
}

// ClearAllCache 清除所有缓存
func (qc *QueryCacheService) ClearAllCache() error {
	return qc.redisService.Clear()
}

// GetCacheStats 获取缓存统计
func (qc *QueryCacheService) GetCacheStats() (map[string]interface{}, error) {
	return qc.redisService.GetStats()
}
