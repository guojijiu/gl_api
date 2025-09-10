package Controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// DocumentationController API文档控制器
// 功能说明：
// 1. 提供API文档和说明
// 2. 生成Swagger文档
// 3. 提供API使用示例
// 4. 支持文档版本管理
type DocumentationController struct {
	Controller
}

// NewDocumentationController 创建文档控制器
func NewDocumentationController() *DocumentationController {
	return &DocumentationController{}
}

// GetAPIDocumentation 获取API文档
// 功能说明：
// 1. 返回API文档的基本信息
// 2. 包含API版本、端点列表
// 3. 提供文档链接和说明
func (c *DocumentationController) GetAPIDocumentation(ctx *gin.Context) {
	documentation := gin.H{
		"title":       "Cloud Platform API Documentation",
		"version":     "1.0.0",
		"description": "基于Gin + Laravel设计理念的Web开发框架API文档",
		"base_url":    "http://localhost:8080/api/v1",
		"swagger_url": "/swagger/index.html",
		"endpoints": gin.H{
			"auth": gin.H{
				"description": "用户认证相关接口",
				"endpoints": []gin.H{
					{"method": "POST", "path": "/auth/register", "description": "用户注册"},
					{"method": "POST", "path": "/auth/login", "description": "用户登录"},
					{"method": "POST", "path": "/auth/logout", "description": "用户登出"},
					{"method": "GET", "path": "/auth/profile", "description": "获取用户资料"},
					{"method": "PUT", "path": "/auth/profile", "description": "更新用户资料"},
					{"method": "POST", "path": "/auth/refresh", "description": "刷新Token"},
					{"method": "POST", "path": "/auth/password/reset-request", "description": "请求密码重置"},
					{"method": "POST", "path": "/auth/password/reset", "description": "重置密码"},
					{"method": "POST", "path": "/auth/email/verify-request", "description": "请求邮箱验证"},
					{"method": "POST", "path": "/auth/email/verify", "description": "验证邮箱"},
				},
			},
			"users": gin.H{
				"description": "用户管理接口",
				"endpoints": []gin.H{
					{"method": "GET", "path": "/users", "description": "获取用户列表"},
					{"method": "GET", "path": "/users/:id", "description": "获取用户详情"},
					{"method": "PUT", "path": "/users/:id", "description": "更新用户信息"},
					{"method": "DELETE", "path": "/users/:id", "description": "删除用户"},
					{"method": "GET", "path": "/users/:id/posts", "description": "获取用户的文章列表"},
				},
			},
			"posts": gin.H{
				"description": "文章管理接口",
				"endpoints": []gin.H{
					{"method": "GET", "path": "/posts", "description": "获取文章列表"},
					{"method": "GET", "path": "/posts/:id", "description": "获取文章详情"},
					{"method": "POST", "path": "/posts", "description": "创建文章"},
					{"method": "PUT", "path": "/posts/:id", "description": "更新文章"},
					{"method": "DELETE", "path": "/posts/:id", "description": "删除文章"},
				},
			},
			"categories": gin.H{
				"description": "分类管理接口",
				"endpoints": []gin.H{
					{"method": "GET", "path": "/categories", "description": "获取分类列表"},
					{"method": "GET", "path": "/categories/:id", "description": "获取分类详情"},
					{"method": "POST", "path": "/categories", "description": "创建分类"},
					{"method": "PUT", "path": "/categories/:id", "description": "更新分类"},
					{"method": "DELETE", "path": "/categories/:id", "description": "删除分类"},
				},
			},
			"tags": gin.H{
				"description": "标签管理接口",
				"endpoints": []gin.H{
					{"method": "GET", "path": "/tags", "description": "获取标签列表"},
					{"method": "GET", "path": "/tags/popular", "description": "获取热门标签"},
					{"method": "GET", "path": "/tags/:id", "description": "获取标签详情"},
					{"method": "POST", "path": "/tags", "description": "创建标签"},
					{"method": "PUT", "path": "/tags/:id", "description": "更新标签"},
					{"method": "DELETE", "path": "/tags/:id", "description": "删除标签"},
				},
			},
			"storage": gin.H{
				"description": "存储管理接口",
				"endpoints": []gin.H{
					{"method": "POST", "path": "/storage/upload", "description": "文件上传"},
					{"method": "GET", "path": "/storage/download/*path", "description": "文件下载"},
					{"method": "DELETE", "path": "/storage/delete/*path", "description": "删除文件"},
					{"method": "GET", "path": "/storage/list", "description": "获取文件列表"},
					{"method": "GET", "path": "/storage/info/*path", "description": "获取文件信息"},
					{"method": "GET", "path": "/storage/logs", "description": "获取日志"},
					{"method": "GET", "path": "/storage/request-logs", "description": "获取请求日志"},
					{"method": "GET", "path": "/storage/sql-logs", "description": "获取SQL日志"},
					{"method": "GET", "path": "/storage/info", "description": "获取存储信息"},
					{"method": "POST", "path": "/storage/cache/clear", "description": "清理缓存"},
					{"method": "POST", "path": "/storage/temp/clean", "description": "清理临时文件"},
					{"method": "POST", "path": "/storage/logs/cleanup", "description": "清理日志"},
					{"method": "GET", "path": "/storage/logs/stats", "description": "获取日志统计"},
				},
			},
			"admin": gin.H{
				"description": "管理员接口",
				"endpoints": []gin.H{
					{"method": "GET", "path": "/admin/dashboard", "description": "管理员仪表板"},
					{"method": "GET", "path": "/admin/stats", "description": "获取系统统计"},
				},
			},
			"monitoring": gin.H{
				"description": "监控接口",
				"endpoints": []gin.H{
					{"method": "GET", "path": "/health", "description": "健康检查"},
					{"method": "GET", "path": "/health/detailed", "description": "详细健康检查"},
					{"method": "GET", "path": "/metrics", "description": "获取指标"},
					{"method": "GET", "path": "/status", "description": "获取系统状态"},
					{"method": "GET", "path": "/stats/performance", "description": "获取性能统计"},
					{"method": "GET", "path": "/stats/errors", "description": "获取错误日志"},
					{"method": "POST", "path": "/stats/cache/clear", "description": "清理缓存"},
					{"method": "POST", "path": "/system/restart", "description": "重启系统"},
				},
			},
		},
		"authentication": gin.H{
			"type": "Bearer Token",
			"description": "使用JWT Bearer Token进行身份验证",
			"header": "Authorization: Bearer <token>",
		},
		"rate_limiting": gin.H{
			"description": "API请求频率限制",
			"limits": gin.H{
				"global": "100 requests per minute",
				"auth": gin.H{
					"register": "5 requests per hour",
					"login":    "10 requests per hour",
					"refresh":  "20 requests per hour",
				},
			},
		},
		"error_codes": []gin.H{
			{"code": 400, "message": "Bad Request", "description": "请求参数错误"},
			{"code": 401, "message": "Unauthorized", "description": "未授权访问"},
			{"code": 403, "message": "Forbidden", "description": "禁止访问"},
			{"code": 404, "message": "Not Found", "description": "资源不存在"},
			{"code": 429, "message": "Too Many Requests", "description": "请求频率超限"},
			{"code": 500, "message": "Internal Server Error", "description": "服务器内部错误"},
		},
		"examples": gin.H{
			"register": gin.H{
				"request": gin.H{
					"username": "testuser",
					"email":    "test@example.com",
					"password": "password123",
				},
				"response": gin.H{
					"success": true,
					"message": "注册成功",
					"data": gin.H{
						"id":       1,
						"username": "testuser",
						"email":    "test@example.com",
						"role":     "user",
						"status":   1,
					},
				},
			},
			"login": gin.H{
				"request": gin.H{
					"username": "testuser",
					"password": "password123",
				},
				"response": gin.H{
					"success": true,
					"message": "登录成功",
					"data": gin.H{
						"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
						"user": gin.H{
							"id":       1,
							"username": "testuser",
							"email":    "test@example.com",
							"role":     "user",
						},
					},
				},
			},
		},
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API文档获取成功",
		"data":    documentation,
	})
}
