package bootstrap

import (
	"cloud-platform-api/app/Config"
	"net/http"

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

// ServeHTTP 实现http.Handler接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Engine.ServeHTTP(w, req)
}
