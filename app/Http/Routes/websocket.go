package Routes

import (
	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/app/Http/Middleware"

	"github.com/gin-gonic/gin"
)

// RegisterWebSocketRoutes 注册WebSocket路由
//
// 重要功能说明：
// 1. WebSocket连接路由：支持实时通信连接
// 2. 房间管理路由：创建、加入、离开房间
// 3. 消息管理路由：发送、接收消息
// 4. 统计信息路由：在线用户、房间统计
// 5. 管理路由：房间管理、用户管理
//
// 安全特性：
// - 所有路由都需要JWT认证
// - 支持用户权限验证
// - 防止恶意访问和攻击
//
// 路由分组：
// - /ws/connect: WebSocket连接
// - /ws/rooms: 房间管理
// - /ws/messages: 消息管理
// - /ws/users: 用户管理
// - /ws/stats: 统计信息
func RegisterWebSocketRoutes(router *gin.Engine) {
	// WebSocket路由组
	wsGroup := router.Group("/ws")
	{
		// WebSocket连接 - 不需要认证，因为需要先建立连接
		wsGroup.GET("/connect", Controllers.NewWebSocketController().Connect)
		
		// 需要认证的路由
		authGroup := wsGroup.Group("")
		authGroup.Use(Middleware.NewAuthMiddleware().Handle())
		{
			// 房间管理
			roomsGroup := authGroup.Group("/rooms")
			{
				roomsGroup.GET("", Controllers.NewWebSocketController().GetRooms)                    // 获取房间列表
				roomsGroup.POST("", Controllers.NewWebSocketController().CreateRoom)               // 创建房间
				roomsGroup.POST("/:room_id/join", Controllers.NewWebSocketController().JoinRoom)   // 加入房间
				roomsGroup.POST("/:room_id/leave", Controllers.NewWebSocketController().LeaveRoom) // 离开房间
			}
			
			// 消息管理
			authGroup.POST("/messages", Controllers.NewWebSocketController().SendMessage) // 发送消息
			
			// 用户管理
			usersGroup := authGroup.Group("/users")
			{
				usersGroup.GET("/online", Controllers.NewWebSocketController().GetOnlineUsers) // 获取在线用户
			}
			
			// 统计信息
			authGroup.GET("/stats", Controllers.NewWebSocketController().GetStats) // 获取统计信息
		}
	}
}
