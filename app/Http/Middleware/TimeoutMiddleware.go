package Middleware

import (
	"cloud-platform-api/app/Storage"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// TimeoutMiddleware 请求超时中间件
type TimeoutMiddleware struct {
	BaseMiddleware
	storageManager *Storage.StorageManager
}

// NewTimeoutMiddleware 创建请求超时中间件
// 功能说明：
// 1. 初始化请求超时中间件实例
// 2. 设置请求超时时间
// 3. 防止长时间运行的请求阻塞服务器
// 4. 记录超时请求日志
func NewTimeoutMiddleware(storageManager *Storage.StorageManager) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		storageManager: storageManager,
	}
}

// Handle 处理请求超时
// 功能说明：
// 1. 为每个请求设置超时上下文
// 2. 默认超时时间为30秒
// 3. 超时时返回408请求超时错误
// 4. 记录超时请求的详细信息
func (m *TimeoutMiddleware) Handle(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建超时上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 将超时上下文设置到请求中
		c.Request = c.Request.WithContext(ctx)

		// 创建完成通道
		done := make(chan bool, 1)

		go func() {
			// 处理请求
			c.Next()
			done <- true
		}()

		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			m.storageManager.LogWarning("请求超时", map[string]interface{}{
				"url":        c.Request.URL.String(),
				"method":     c.Request.Method,
				"client_ip":  c.ClientIP(),
				"timeout":    timeout.String(),
				"user_agent": c.Request.UserAgent(),
			})

			// 返回超时错误
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
				"error":   "The request took too long to process",
			})
			c.Abort()
			return
		}
	}
}
