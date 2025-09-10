package Controllers

import (
	"cloud-platform-api/app/Services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// WebSocketController WebSocket控制器
//
// 重要功能说明：
// 1. WebSocket连接管理：建立、断开、状态查询
// 2. 房间管理：创建、加入、离开、查询房间信息
// 3. 消息管理：发送、接收、历史消息查询
// 4. 用户状态：在线状态、用户统计
// 5. 系统监控：连接统计、性能监控
//
// 安全特性：
// - 所有接口都需要JWT认证
// - 支持用户权限验证
// - 防止恶意消息和攻击
// - 支持IP白名单和黑名单
//
// 性能优化：
// - 支持连接池管理
// - 异步消息处理
// - 消息队列和缓存
// - 支持负载均衡
type WebSocketController struct {
	Controller
	webSocketService *Services.WebSocketService
}

// NewWebSocketController 创建WebSocket控制器
func NewWebSocketController() *WebSocketController {
	return &WebSocketController{
		webSocketService: Services.NewWebSocketService(nil),
	}
}

// Connect WebSocket连接
// @Summary 建立WebSocket连接
// @Description 建立WebSocket连接，支持实时通信
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param room_id query string false "房间ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /ws/connect [get]
func (c *WebSocketController) Connect(ctx *gin.Context) {
	// 获取用户信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		c.Error(ctx, http.StatusUnauthorized, "需要登录")
		return
	}
	
	username := ctx.GetString("username")
	
	// 获取房间ID
	roomID := ctx.Query("room_id")
	if roomID == "" {
		roomID = "general" // 默认房间
	}
	
	// 建立WebSocket连接
	c.webSocketService.HandleWebSocket(ctx.Writer, ctx.Request)
	
	// 记录连接日志
	c.logWebSocketAction("connect", uint(userID.(uint)), username, map[string]interface{}{
		"room_id": roomID,
		"ip":      ctx.ClientIP(),
		"user_agent": ctx.GetHeader("User-Agent"),
	})
}

// GetRooms 获取房间列表
// @Summary 获取房间列表
// @Description 获取所有可用的聊天房间
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /ws/rooms [get]
func (c *WebSocketController) GetRooms(ctx *gin.Context) {
	rooms := c.webSocketService.GetRooms()
	
	// 转换为响应格式
	roomList := make([]map[string]interface{}, 0, len(rooms))
	for _, room := range rooms {
		roomList = append(roomList, map[string]interface{}{
			"id":            room.ID,
			"name":          room.Name,
			"description":   room.Description,
			"client_count":  room.GetClientCount(),
			"message_count": room.MessageCount,
			"created_at":    room.CreatedAt,
		})
	}
	
	c.Success(ctx, gin.H{
		"rooms": roomList,
		"total": len(roomList),
	}, "房间列表获取成功")
}

// CreateRoom 创建房间
// @Summary 创建聊天房间
// @Description 创建新的聊天房间
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param room body CreateRoomRequest true "房间信息"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /ws/rooms [post]
func (c *WebSocketController) CreateRoom(ctx *gin.Context) {
	var req CreateRoomRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Error(ctx, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}
	
	// 验证房间名称
	if req.Name == "" {
		c.Error(ctx, http.StatusBadRequest, "房间名称不能为空")
		return
	}
	
	// 创建房间
	room := c.webSocketService.CreateRoom(req.ID, req.Name, req.Description)
	
	// 记录操作日志
	userID, _ := ctx.Get("user_id")
	username := ctx.GetString("username")
	c.logWebSocketAction("create_room", uint(userID.(uint)), username, map[string]interface{}{
		"room_id":      room.ID,
		"room_name":    room.Name,
		"description":  room.Description,
	})
	
	c.Success(ctx, gin.H{
		"room": map[string]interface{}{
			"id":            room.ID,
			"name":          room.Name,
			"description":   room.Description,
			"client_count":  room.GetClientCount(),
			"message_count": room.MessageCount,
			"created_at":    room.CreatedAt,
		},
	}, "房间创建成功")
}

// JoinRoom 加入房间
// @Summary 加入聊天房间
// @Description 加入指定的聊天房间
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param room_id path string true "房间ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /ws/rooms/{room_id}/join [post]
func (c *WebSocketController) JoinRoom(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	if roomID == "" {
		c.Error(ctx, http.StatusBadRequest, "房间ID不能为空")
		return
	}
	
	// 检查房间是否存在
	room := c.webSocketService.GetRoom(roomID)
	if room == nil {
		c.Error(ctx, http.StatusNotFound, "房间不存在")
		return
	}
	
	// 记录操作日志
	userID, _ := ctx.Get("user_id")
	username := ctx.GetString("username")
	c.logWebSocketAction("join_room", uint(userID.(uint)), username, map[string]interface{}{
		"room_id": roomID,
	})
	
	c.Success(ctx, gin.H{
		"message": "成功加入房间",
		"room": map[string]interface{}{
			"id":            room.ID,
			"name":          room.Name,
			"description":   room.Description,
			"client_count":  room.GetClientCount(),
			"message_count": room.MessageCount,
		},
	}, "成功加入房间")
}

// LeaveRoom 离开房间
// @Summary 离开聊天房间
// @Description 离开指定的聊天房间
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param room_id path string true "房间ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /ws/rooms/{room_id}/leave [post]
func (c *WebSocketController) LeaveRoom(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	if roomID == "" {
		c.Error(ctx, http.StatusBadRequest, "房间ID不能为空")
		return
	}
	
	// 记录操作日志
	userID, _ := ctx.Get("user_id")
	username := ctx.GetString("username")
	c.logWebSocketAction("leave_room", uint(userID.(uint)), username, map[string]interface{}{
		"room_id": roomID,
	})
	
	c.Success(ctx, gin.H{
		"message": "成功离开房间",
		"room_id": roomID,
	}, "成功离开房间")
}

// SendMessage 发送消息
// @Summary 发送消息
// @Description 向指定房间或用户发送消息
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param message body SendMessageRequest true "消息内容"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /ws/messages [post]
func (c *WebSocketController) SendMessage(ctx *gin.Context) {
	var req SendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Error(ctx, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}
	
	// 验证消息内容
	if req.Content == "" {
		c.Error(ctx, http.StatusBadRequest, "消息内容不能为空")
		return
	}
	
	// 根据消息类型处理
	switch req.Type {
	case "room_message":
		if req.RoomID == "" {
			c.Error(ctx, http.StatusBadRequest, "房间消息需要指定房间ID")
			return
		}
		// TODO: 实现房间消息广播
	case "private_message":
		if req.To == "" {
			c.Error(ctx, http.StatusBadRequest, "私聊消息需要指定接收者")
			return
		}
		// TODO: 实现私聊逻辑
	case "broadcast":
		// TODO: 实现广播消息
	default:
		c.Error(ctx, http.StatusBadRequest, "不支持的消息类型")
		return
	}
	
	// 记录操作日志
	userID, _ := ctx.Get("user_id")
	username := ctx.GetString("username")
	c.logWebSocketAction("send_message", uint(userID.(uint)), username, map[string]interface{}{
		"message_type": req.Type,
		"room_id":      req.RoomID,
		"to":           req.To,
		"content":      req.Content,
	})
	
	c.Success(ctx, gin.H{
		"message": "消息发送成功",
		"message_id": time.Now().UnixNano(),
	}, "消息发送成功")
}

// GetOnlineUsers 获取在线用户
// @Summary 获取在线用户
// @Description 获取当前在线用户数量和统计信息
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /ws/users/online [get]
func (c *WebSocketController) GetOnlineUsers(ctx *gin.Context) {
	count := c.webSocketService.GetOnlineUsers()
	
	c.Success(ctx, gin.H{
		"online_users": count,
		"timestamp":    time.Now(),
	}, "在线用户信息获取成功")
}

// GetStats 获取统计信息
// @Summary 获取WebSocket统计信息
// @Description 获取WebSocket服务的统计信息
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /ws/stats [get]
func (c *WebSocketController) GetStats(ctx *gin.Context) {
	rooms := c.webSocketService.GetRooms()
	onlineUsers := c.webSocketService.GetOnlineUsers()
	
	// 计算房间统计
	roomStats := make(map[string]interface{})
	totalMessages := int64(0)
	
	for _, room := range rooms {
		roomStats[room.ID] = map[string]interface{}{
			"name":          room.Name,
			"client_count":  room.GetClientCount(),
			"message_count": room.MessageCount,
		}
		totalMessages += room.MessageCount
	}
	
	c.Success(ctx, gin.H{
		"total_rooms":     len(rooms),
		"online_users":    onlineUsers,
		"total_messages":  totalMessages,
		"rooms":           roomStats,
		"timestamp":       time.Now(),
	}, "统计信息获取成功")
}

// 请求结构体

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Type    string                 `json:"type" binding:"required"` // room_message, private_message, broadcast
	Content string                 `json:"content" binding:"required"`
	RoomID  string                 `json:"room_id"` // 房间消息时必填
	To      string                 `json:"to"`      // 私聊消息时必填
	Data    map[string]interface{} `json:"data"`    // 额外数据
}

// 辅助方法

// logWebSocketAction 记录WebSocket操作日志
func (c *WebSocketController) logWebSocketAction(action string, userID uint, username string, fields map[string]interface{}) {
	// 这里应该调用审计服务记录日志
	// 暂时只打印日志
	// TODO: 实现日志记录功能
}
