package Middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS中间件
type CORSMiddleware struct {
	BaseMiddleware
}

// NewCORSMiddleware 创建CORS中间件
// 功能说明：
// 1. 初始化CORS中间件实例
// 2. 处理跨域资源共享
// 3. 配置允许的请求方法和头部
// 4. 支持预检请求处理
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

// Handle 处理CORS请求
// 功能说明：
// 1. 设置允许的请求来源（Origin）
// 2. 配置允许的HTTP方法（GET, POST, PUT, DELETE, OPTIONS）
// 3. 设置允许的请求头部
// 4. 处理预检请求（OPTIONS）
// 5. 设置凭证支持
func (m *CORSMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置CORS头部
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
