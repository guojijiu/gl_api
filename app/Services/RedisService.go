package Services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud-platform-api/app/Models"
	"github.com/redis/go-redis/v9"
)

// RedisService Redis缓存服务
type RedisService struct {
	client *redis.Client
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedisService 创建Redis服务
func NewRedisService(config *RedisConfig) *RedisService {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	return &RedisService{
		client: client,
	}
}

// Set 设置缓存
func (r *RedisService) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func (r *RedisService) Get(key string, dest interface{}) error {
	ctx := context.Background()
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("缓存不存在: %s", key)
		}
		return fmt.Errorf("获取缓存失败: %v", err)
	}

	return json.Unmarshal(data, dest)
}

// Delete 删除缓存
func (r *RedisService) Delete(key string) error {
	ctx := context.Background()
	return r.client.Del(ctx, key).Err()
}

// Clear 清空缓存
func (r *RedisService) Clear() error {
	ctx := context.Background()
	return r.client.FlushDB(ctx).Err()
}

// Exists 检查缓存是否存在
func (r *RedisService) Exists(key string) (bool, error) {
	ctx := context.Background()
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// TTL 获取缓存剩余时间
func (r *RedisService) TTL(key string) (time.Duration, error) {
	ctx := context.Background()
	return r.client.TTL(ctx, key).Result()
}

// CacheUser 缓存用户信息
func (r *RedisService) CacheUser(userID uint, user *Models.User) error {
	key := fmt.Sprintf("user:%d", userID)
	return r.Set(key, user, 1*time.Hour)
}

// GetCachedUser 获取缓存的用户信息
func (r *RedisService) GetCachedUser(userID uint) (*Models.User, error) {
	key := fmt.Sprintf("user:%d", userID)
	var user Models.User
	err := r.Get(key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CachePosts 缓存文章列表
func (r *RedisService) CachePosts(page, limit int, posts []Models.Post) error {
	key := fmt.Sprintf("posts:page:%d:limit:%d", page, limit)
	return r.Set(key, posts, 5*time.Minute)
}

// GetCachedPosts 获取缓存的文章列表
func (r *RedisService) GetCachedPosts(page, limit int) ([]Models.Post, error) {
	key := fmt.Sprintf("posts:page:%d:limit:%d", page, limit)
	var posts []Models.Post
	err := r.Get(key, &posts)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// CacheCategories 缓存分类列表
func (r *RedisService) CacheCategories(categories []Models.Category) error {
	key := "categories:all"
	return r.Set(key, categories, 10*time.Minute)
}

// GetCachedCategories 获取缓存的分类列表
func (r *RedisService) GetCachedCategories() ([]Models.Category, error) {
	key := "categories:all"
	var categories []Models.Category
	err := r.Get(key, &categories)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// CacheTags 缓存标签列表
func (r *RedisService) CacheTags(tags []Models.Tag) error {
	key := "tags:all"
	return r.Set(key, tags, 10*time.Minute)
}

// GetCachedTags 获取缓存的标签列表
func (r *RedisService) GetCachedTags() ([]Models.Tag, error) {
	key := "tags:all"
	var tags []Models.Tag
	err := r.Get(key, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// InvalidateUserCache 清除用户相关缓存
func (r *RedisService) InvalidateUserCache(userID uint) error {
	key := fmt.Sprintf("user:%d", userID)
	return r.Delete(key)
}

// InvalidatePostsCache 清除文章相关缓存
func (r *RedisService) InvalidatePostsCache() error {
	ctx := context.Background()
	pattern := "posts:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}

// InvalidateCategoriesCache 清除分类相关缓存
func (r *RedisService) InvalidateCategoriesCache() error {
	key := "categories:all"
	return r.Delete(key)
}

// InvalidateTagsCache 清除标签相关缓存
func (r *RedisService) InvalidateTagsCache() error {
	key := "tags:all"
	return r.Delete(key)
}

// GetStats 获取缓存统计信息
func (r *RedisService) GetStats() (map[string]interface{}, error) {
	ctx := context.Background()
	info := r.client.Info(ctx, "memory").Val()
	
	// 解析Redis信息
	stats := make(map[string]interface{})
	stats["info"] = info
	stats["db_size"] = r.client.DBSize(ctx).Val()
	
	return stats, nil
}

// Ping 测试Redis连接
func (r *RedisService) Ping() error {
	ctx := context.Background()
	_, err := r.client.Ping(ctx).Result()
	return err
}

// Close 关闭Redis连接
func (r *RedisService) Close() error {
	return r.client.Close()
}

// SetWithTTL 设置缓存（带TTL）
func (r *RedisService) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// GetString 获取字符串缓存
func (r *RedisService) GetString(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del 删除缓存（带上下文）
func (r *RedisService) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// ExistsWithContext 检查缓存是否存在（带上下文）
func (r *RedisService) ExistsWithContext(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Keys 获取匹配模式的键
func (r *RedisService) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}
