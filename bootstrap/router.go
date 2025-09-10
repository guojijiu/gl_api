package bootstrap

import (
	"cloud-platform-api/app/Config"
	"github.com/gin-gonic/gin"
)

// Router 路由器结构
type Router struct {
	Engine *gin.Engine
}

// NewRouter 创建新的路由器
func NewRouter() *Router {
	// 设置Gin模式
	gin.SetMode(Config.GetConfig().Server.Mode)
	
	// 创建Gin实例
	engine := gin.New()
	
	// 添加基础中间件
	engine.Use(gin.Recovery())
	
	return &Router{
		Engine: engine,
	}
}
