package Middleware

import (
	"cloud-platform-api/app/Storage"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware 错误恢复中间件
type RecoveryMiddleware struct {
	BaseMiddleware
	storageManager *Storage.StorageManager
}

// NewRecoveryMiddleware 创建错误恢复中间件
// 功能说明：
// 1. 初始化错误恢复中间件实例
// 2. 用于捕获和处理panic异常
// 3. 记录详细的错误信息和堆栈跟踪
// 4. 返回友好的错误响应
func NewRecoveryMiddleware(storageManager *Storage.StorageManager) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		storageManager: storageManager,
	}
}

// Handle 处理错误恢复
// 功能说明：
// 1. 使用defer recover()捕获panic
// 2. 记录详细的错误信息和堆栈跟踪
// 3. 返回500内部服务器错误响应
// 4. 防止应用因panic而崩溃
func (m *RecoveryMiddleware) Handle() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录错误信息
		m.storageManager.LogError("应用发生panic", map[string]interface{}{
			"error":        recovered,
			"url":          c.Request.URL.String(),
			"method":       c.Request.Method,
			"client_ip":    c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
			"stack_trace":  string(debug.Stack()),
		})

		// 返回错误响应
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Internal server error",
			"error":   "An unexpected error occurred",
		})
	})
}
