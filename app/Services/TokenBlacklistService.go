package Services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TokenBlacklistService Token黑名单服务
// 功能说明：
// 1. 管理已登出的JWT token
// 2. 防止登出后的token继续使用
// 3. 支持Redis和内存两种存储方式
// 4. 自动清理过期的token
type TokenBlacklistService struct {
	redisService *RedisService
	blacklist    map[string]time.Time // 内存黑名单（备用）
}

// BlacklistedToken 黑名单中的token信息
type BlacklistedToken struct {
	Token     string    `json:"token"`
	UserID    uint      `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	AddedAt   time.Time `json:"added_at"`
}

// NewTokenBlacklistService 创建Token黑名单服务
// 功能说明：
// 1. 初始化Token黑名单服务实例
// 2. 配置Redis服务用于持久化存储
// 3. 初始化内存黑名单作为备用
func NewTokenBlacklistService(redisService *RedisService) *TokenBlacklistService {
	return &TokenBlacklistService{
		redisService: redisService,
		blacklist:    make(map[string]time.Time),
	}
}

// AddToBlacklist 将token添加到黑名单
// 功能说明：
// 1. 将token添加到Redis黑名单中
// 2. 设置过期时间与token的过期时间一致
// 3. 同时添加到内存黑名单作为备用
// 4. 记录添加时间和用户信息
func (s *TokenBlacklistService) AddToBlacklist(token string, userID uint, expiresAt time.Time) error {
	// 计算token的剩余有效期
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}

	// 创建黑名单记录
	blacklistedToken := BlacklistedToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
		AddedAt:   time.Now(),
	}

	// 序列化为JSON
	tokenData, err := json.Marshal(blacklistedToken)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %v", err)
	}

	// 添加到Redis黑名单
	key := fmt.Sprintf("blacklist:%s", token)
	if s.redisService != nil {
		err = s.redisService.SetWithTTL(context.Background(), key, string(tokenData), ttl)
		if err != nil {
			// Redis失败时使用内存存储
			s.blacklist[token] = expiresAt
			return fmt.Errorf("redis failed, using memory storage: %v", err)
		}
	} else {
		// 没有Redis时使用内存存储
		s.blacklist[token] = expiresAt
	}

	return nil
}

// IsBlacklisted 检查token是否在黑名单中
// 功能说明：
// 1. 首先检查Redis黑名单
// 2. Redis不可用时检查内存黑名单
// 3. 清理过期的内存黑名单记录
// 4. 返回token是否在黑名单中
func (s *TokenBlacklistService) IsBlacklisted(token string) bool {
	// 首先检查Redis黑名单
	if s.redisService != nil {
		key := fmt.Sprintf("blacklist:%s", token)
		exists, err := s.redisService.Exists(key)
		if err == nil && exists {
			return true
		}
	}

	// 检查内存黑名单
	if expiresAt, exists := s.blacklist[token]; exists {
		// 检查是否过期
		if time.Now().Before(expiresAt) {
			return true
		} else {
			// 清理过期的记录
			delete(s.blacklist, token)
		}
	}

	return false
}

// RemoveFromBlacklist 从黑名单中移除token
// 功能说明：
// 1. 从Redis黑名单中移除token
// 2. 从内存黑名单中移除token
// 3. 用于token重新激活的场景
func (s *TokenBlacklistService) RemoveFromBlacklist(token string) error {
	// 从Redis黑名单中移除
	if s.redisService != nil {
		key := fmt.Sprintf("blacklist:%s", token)
		err := s.redisService.Del(context.Background(), key)
		if err != nil {
			return fmt.Errorf("failed to remove from redis blacklist: %v", err)
		}
	}

	// 从内存黑名单中移除
	delete(s.blacklist, token)

	return nil
}

// GetBlacklistStats 获取黑名单统计信息
// 功能说明：
// 1. 获取Redis黑名单中的token数量
// 2. 获取内存黑名单中的token数量
// 3. 返回黑名单统计信息
func (s *TokenBlacklistService) GetBlacklistStats() map[string]interface{} {
	stats := map[string]interface{}{
		"memory_count": len(s.blacklist),
		"redis_count":  0,
	}

	// 获取Redis黑名单数量
	if s.redisService != nil {
		keys, err := s.redisService.Keys(context.Background(), "blacklist:*")
		if err == nil {
			stats["redis_count"] = len(keys)
		}
	}

	return stats
}

// CleanupExpiredTokens 清理过期的token
// 功能说明：
// 1. 清理内存黑名单中过期的token
// 2. Redis中的token会自动过期
// 3. 定期调用以释放内存
func (s *TokenBlacklistService) CleanupExpiredTokens() int {
	cleaned := 0
	now := time.Now()

	for token, expiresAt := range s.blacklist {
		if now.After(expiresAt) {
			delete(s.blacklist, token)
			cleaned++
		}
	}

	return cleaned
}

// GetBlacklistedTokenInfo 获取黑名单中的token信息
// 功能说明：
// 1. 从Redis或内存中获取token的详细信息
// 2. 返回token的用户ID、过期时间等信息
// 3. 用于审计和调试
func (s *TokenBlacklistService) GetBlacklistedTokenInfo(token string) (*BlacklistedToken, error) {
	// 首先从Redis获取
	if s.redisService != nil {
		key := fmt.Sprintf("blacklist:%s", token)
		var data string
		err := s.redisService.Get(key, &data)
		if err == nil && data != "" {
			var blacklistedToken BlacklistedToken
			if err := json.Unmarshal([]byte(data), &blacklistedToken); err == nil {
				return &blacklistedToken, nil
			}
		}
	}

	// 从内存获取
	if expiresAt, exists := s.blacklist[token]; exists {
		return &BlacklistedToken{
			Token:     token,
			ExpiresAt: expiresAt,
			AddedAt:   time.Now(), // 内存中不存储AddedAt，使用当前时间
		}, nil
	}

	return nil, fmt.Errorf("token not found in blacklist")
}
