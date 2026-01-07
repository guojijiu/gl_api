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
// 
// 功能说明：
// 1. 为每个请求设置超时上下文，防止长时间运行的请求阻塞服务器
// 2. 使用goroutine并发处理请求，通过select监听完成或超时
// 3. 超时时返回408请求超时错误，并记录详细日志
// 4. 确保context的cancel函数被调用，避免资源泄漏
//
// 实现原理：
// - 使用context.WithTimeout创建带超时的上下文
// - 在独立的goroutine中执行请求处理（c.Next()）
// - 使用select同时监听请求完成信号和超时信号
// - 如果超时先发生，立即返回408错误并调用c.Abort()停止后续处理
//
// 注意事项：
// - defer cancel()确保即使函数提前返回也会释放context资源
// - done通道使用缓冲通道（容量1）避免goroutine阻塞
// - 超时后goroutine可能仍在运行，但c.Abort()会阻止响应写入
// - storageManager需要nil检查，避免在未初始化时panic
func (m *TimeoutMiddleware) Handle(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建超时上下文
		// 注意：context.WithTimeout会创建一个在指定时间后自动取消的context
		// 必须调用cancel()释放资源，即使context已经超时
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel() // 确保资源释放，避免context泄漏

		// 将超时上下文设置到请求中
		// 这样后续的中间件和处理器可以通过c.Request.Context()获取超时上下文
		// 如果它们支持context取消，可以在超时时自动停止
		c.Request = c.Request.WithContext(ctx)

		// 创建完成通道（使用缓冲通道避免goroutine阻塞）
		// 容量为1，确保即使没有接收者，发送操作也不会阻塞
		done := make(chan bool, 1)

		// 在独立的goroutine中处理请求
		// 这样可以并发监听请求完成和超时两个事件
		go func() {
			// 处理请求（执行后续中间件和处理器）
			// 注意：即使超时发生，这个goroutine可能仍在运行
			// 但c.Abort()会阻止响应写入，客户端会收到超时错误
			c.Next()
			// 请求处理完成，发送完成信号
			done <- true
		}()

		// 使用select同时监听两个事件：请求完成或超时
		select {
		case <-done:
			// 请求正常完成（在超时之前）
			// 正常返回，让响应继续发送给客户端
			return
		case <-ctx.Done():
			// 请求超时（超时发生在请求完成之前）
			// ctx.Done()会在超时时间到达时被关闭
			
			// 记录超时日志（需要nil检查，避免panic）
			if m.storageManager != nil {
				m.storageManager.LogWarning("请求超时", map[string]interface{}{
					"url":        c.Request.URL.String(),
					"method":     c.Request.Method,
					"client_ip":  c.ClientIP(),
					"timeout":    timeout.String(),
					"user_agent": c.Request.UserAgent(),
				})
			}

			// 返回超时错误响应
			// 使用408 Request Timeout状态码，符合HTTP标准
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
				"error":   "The request took too long to process",
			})
			// 中止后续处理，防止重复响应
			c.Abort()
			return
		}
	}
}
