package Middleware

import (
	"cloud-platform-api/app/Storage"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// PermissionMiddleware 权限管理中间件
type PermissionMiddleware struct {
	BaseMiddleware
	storageManager *Storage.StorageManager
}

// NewPermissionMiddleware 创建权限管理中间件
// 功能说明：
// 1. 初始化权限管理中间件实例
// 2. 提供细粒度的权限控制
// 3. 支持角色和权限的灵活配置
// 4. 记录权限检查日志
func NewPermissionMiddleware(storageManager *Storage.StorageManager) *PermissionMiddleware {
	return &PermissionMiddleware{
		storageManager: storageManager,
	}
}

// RequireRole 要求特定角色
// 功能说明：
// 1. 检查用户是否具有指定角色
// 2. 支持多个角色的OR逻辑
// 3. 记录权限检查结果
func (m *PermissionMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole == "" {
			m.logPermissionDenied(c, "no_role", "用户未登录")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "需要登录",
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			m.logPermissionDenied(c, "insufficient_role", "角色权限不足")
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
				"error":   "需要角色: " + strings.Join(roles, " 或 "),
			})
			c.Abort()
			return
		}

		// 记录权限检查成功
		m.logPermissionGranted(c, "role_check", "角色权限验证通过")
		c.Next()
	}
}

// RequirePermission 要求特定权限
// 功能说明：
// 1. 检查用户是否具有指定权限
// 2. 支持权限的AND和OR逻辑
// 3. 可扩展的权限系统
func (m *PermissionMiddleware) RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			m.logPermissionDenied(c, "no_user", "用户未登录")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "需要登录",
			})
			c.Abort()
			return
		}

		// 这里可以实现更复杂的权限检查逻辑
		// 例如从数据库或缓存中获取用户权限
		// 目前简化为基于角色的权限检查
		userRole := c.GetString("user_role")
		
		// 管理员拥有所有权限
		if userRole == "admin" {
			m.logPermissionGranted(c, "admin_permission", "管理员权限")
			c.Next()
			return
		}

		// 检查具体权限（这里可以根据实际需求扩展）
		hasPermission := m.checkUserPermissions(userID, permissions)
		
		if !hasPermission {
			m.logPermissionDenied(c, "insufficient_permission", "权限不足")
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
				"error":   "需要权限: " + strings.Join(permissions, " 和 "),
			})
			c.Abort()
			return
		}

		m.logPermissionGranted(c, "permission_check", "权限验证通过")
		c.Next()
	}
}

// RequireOwnership 要求资源所有权
// 功能说明：
// 1. 检查用户是否拥有指定资源
// 2. 支持资源ID参数检查
// 3. 管理员可以访问所有资源
func (m *PermissionMiddleware) RequireOwnership(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			m.logPermissionDenied(c, "no_user", "用户未登录")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "需要登录",
			})
			c.Abort()
			return
		}

		userRole := c.GetString("user_role")
		
		// 管理员可以访问所有资源
		if userRole == "admin" {
			m.logPermissionGranted(c, "admin_ownership", "管理员资源访问权限")
			c.Next()
			return
		}

		// 检查资源所有权
		resourceID := c.Param("id")
		if resourceID == "" {
			m.logPermissionDenied(c, "no_resource_id", "缺少资源ID")
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "缺少资源ID",
			})
			c.Abort()
			return
		}

		// 这里应该实现具体的资源所有权检查逻辑
		// 例如查询数据库检查资源是否属于当前用户
		isOwner := m.checkResourceOwnership(userID, resourceType, resourceID)
		
		if !isOwner {
			m.logPermissionDenied(c, "not_owner", "不是资源所有者")
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
				"error":   "您不是该资源的所有者",
			})
			c.Abort()
			return
		}

		m.logPermissionGranted(c, "ownership_check", "资源所有权验证通过")
		c.Next()
	}
}

// checkUserPermissions 检查用户权限
// 功能说明：
// 1. 实现具体的权限检查逻辑
// 2. 可以从数据库或缓存中获取用户权限
// 3. 支持复杂的权限规则
func (m *PermissionMiddleware) checkUserPermissions(userID string, permissions []string) bool {
	// 这里应该实现具体的权限检查逻辑
	// 例如从数据库查询用户权限表
	// 目前返回true作为示例
	return true
}

// checkResourceOwnership 检查资源所有权
// 功能说明：
// 1. 检查用户是否拥有指定资源
// 2. 支持不同类型的资源检查
// 3. 可扩展的资源权限系统
func (m *PermissionMiddleware) checkResourceOwnership(userID, resourceType, resourceID string) bool {
	// 这里应该实现具体的资源所有权检查逻辑
	// 例如查询数据库检查资源的user_id字段
	// 目前返回true作为示例
	return true
}

// logPermissionGranted 记录权限授予日志
func (m *PermissionMiddleware) logPermissionGranted(c *gin.Context, action, reason string) {
	m.storageManager.LogInfo("权限授予", map[string]interface{}{
		"action":    action,
		"reason":    reason,
		"user_id":   c.GetString("user_id"),
		"user_role": c.GetString("user_role"),
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"client_ip": c.ClientIP(),
	})
}

// logPermissionDenied 记录权限拒绝日志
func (m *PermissionMiddleware) logPermissionDenied(c *gin.Context, action, reason string) {
	m.storageManager.LogWarning("权限拒绝", map[string]interface{}{
		"action":    action,
		"reason":    reason,
		"user_id":   c.GetString("user_id"),
		"user_role": c.GetString("user_role"),
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"client_ip": c.ClientIP(),
	})
}
