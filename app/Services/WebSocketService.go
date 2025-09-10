package Services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketService WebSocket服务
//
// 重要功能说明：
// 1. 实时通信：支持客户端与服务器之间的实时双向通信
// 2. 房间管理：支持多房间、私聊、群聊等通信模式
// 3. 用户管理：在线用户统计、用户状态管理
// 4. 消息广播：支持全局广播、房间广播、私聊消息
// 5. 连接管理：连接池管理、心跳检测、断线重连
//
// 技术特性：
// - 基于Gorilla WebSocket库，支持标准WebSocket协议
// - 支持JSON消息格式，易于扩展和调试
// - 支持连接池管理，高效处理大量并发连接
// - 支持心跳检测，自动清理无效连接
// - 支持消息队列，防止消息丢失
type WebSocketService struct {
	// 连接管理
	clients    map[*Client]bool
	clientsMu  sync.RWMutex
	
	// 房间管理
	rooms      map[string]*Room
	roomsMu    sync.RWMutex
	
	// 消息处理
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	
	// 配置
	upgrader   websocket.Upgrader
	config     *WebSocketConfig
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	ReadBufferSize   int           `json:"read_buffer_size"`
	WriteBufferSize  int           `json:"write_buffer_size"`
	PingPeriod       time.Duration `json:"ping_period"`
	PongWait         time.Duration `json:"pong_wait"`
	WriteWait        time.Duration `json:"write_wait"`
	MaxMessageSize   int64         `json:"max_message_size"`
	EnableCompression bool         `json:"enable_compression"`
}

// Client WebSocket客户端
type Client struct {
	ID       string          `json:"id"`
	UserID   uint            `json:"user_id"`
	Username string          `json:"username"`
	Conn     *websocket.Conn `json:"-"`
	Service  *WebSocketService `json:"-"`
	
	// 消息通道
	send chan []byte
	
	// 房间信息
	rooms map[string]*Room
	
	// 连接状态
	connected bool
	mu       sync.Mutex
}

// Room 聊天房间
type Room struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Clients     map[*Client]bool `json:"-"`
	mu          sync.RWMutex
	
	// 房间统计
	CreatedAt   time.Time `json:"created_at"`
	MessageCount int64     `json:"message_count"`
}

// Message WebSocket消息
type Message struct {
	Type      string                 `json:"type"`
	From      string                 `json:"from"`
	To        string                 `json:"to,omitempty"`
	RoomID    string                 `json:"room_id,omitempty"`
	Content   string                 `json:"content"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewWebSocketService 创建WebSocket服务
func NewWebSocketService(config *WebSocketConfig) *WebSocketService {
	if config == nil {
		config = &WebSocketConfig{
			ReadBufferSize:   1024,
			WriteBufferSize:  1024,
			PingPeriod:       60 * time.Second,
			PongWait:         10 * time.Second,
			WriteWait:        10 * time.Second,
			MaxMessageSize:   512,
			EnableCompression: true,
		}
	}
	
	service := &WebSocketService{
		clients:   make(map[*Client]bool),
		rooms:     make(map[string]*Room),
		broadcast: make(chan *Message, 100),
		register:  make(chan *Client, 100),
		unregister: make(chan *Client, 100),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源，生产环境应该限制
			},
			EnableCompression: config.EnableCompression,
		},
		config: config,
	}
	
	// 启动消息处理
	go service.run()
	
	// 启动心跳检测
	go service.heartbeat()
	
	return service
}

// HandleWebSocket 处理WebSocket连接
func (s *WebSocketService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接为WebSocket连接
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	
	// 创建客户端
	client := &Client{
		ID:       generateClientID(),
		Conn:     conn,
		Service:  s,
		send:     make(chan []byte, 256),
		rooms:    make(map[string]*Room),
		connected: true,
	}
	
	// 注册客户端
	s.register <- client
	
	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// run 运行消息处理循环
func (s *WebSocketService) run() {
	for {
		select {
		case client := <-s.register:
			s.clientsMu.Lock()
			s.clients[client] = true
			s.clientsMu.Unlock()
			
			// 发送欢迎消息
			welcomeMsg := &Message{
				Type:      "welcome",
				Content:   "欢迎连接到WebSocket服务",
				Timestamp: time.Now(),
			}
			client.sendMessage(welcomeMsg)
			
			// 更新在线用户统计
			s.broadcastUserCount()
			
		case client := <-s.unregister:
			s.clientsMu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.clientsMu.Unlock()
			
			// 从所有房间中移除
			client.leaveAllRooms()
			
			// 更新在线用户统计
			s.broadcastUserCount()
			
		case message := <-s.broadcast:
			s.broadcastMessage(message)
		}
	}
}

// broadcastMessage 广播消息
func (s *WebSocketService) broadcastMessage(message *Message) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	
	for client := range s.clients {
		select {
		case client.send <- message.ToJSON():
		default:
			close(client.send)
			delete(s.clients, client)
		}
	}
}

// broadcastToRoom 向房间广播消息
func (s *WebSocketService) broadcastToRoom(roomID string, message *Message) {
	s.roomsMu.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMu.RUnlock()
	
	if !exists {
		return
	}
	
	room.broadcast(message)
}

// CreateRoom 创建房间
func (s *WebSocketService) CreateRoom(id, name, description string) *Room {
	s.roomsMu.Lock()
	defer s.roomsMu.Unlock()
	
	room := &Room{
		ID:          id,
		Name:        name,
		Description: description,
		Clients:     make(map[*Client]bool),
		CreatedAt:   time.Now(),
	}
	
	s.rooms[id] = room
	return room
}

// GetRoom 获取房间
func (s *WebSocketService) GetRoom(id string) *Room {
	s.roomsMu.RLock()
	defer s.roomsMu.RUnlock()
	
	return s.rooms[id]
}

// GetRooms 获取所有房间
func (s *WebSocketService) GetRooms() []*Room {
	s.roomsMu.RLock()
	defer s.roomsMu.RUnlock()
	
	rooms := make([]*Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}
	
	return rooms
}

// GetOnlineUsers 获取在线用户数量
func (s *WebSocketService) GetOnlineUsers() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	
	return len(s.clients)
}

// broadcastUserCount 广播在线用户数量
func (s *WebSocketService) broadcastUserCount() {
	count := s.GetOnlineUsers()
	msg := &Message{
		Type: "user_count",
		Data: map[string]interface{}{
			"count": count,
		},
		Timestamp: time.Now(),
	}
	
	s.broadcastMessage(msg)
}

// heartbeat 心跳检测
func (s *WebSocketService) heartbeat() {
	ticker := time.NewTicker(s.config.PingPeriod)
	defer ticker.Stop()
	
	for range ticker.C {
		s.clientsMu.RLock()
		for client := range s.clients {
			client.mu.Lock()
			if client.connected {
				if err := client.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(s.config.WriteWait)); err != nil {
					client.connected = false
				}
			}
			client.mu.Unlock()
		}
		s.clientsMu.RUnlock()
	}
}

// Client方法

// readPump 读取消息泵
func (c *Client) readPump() {
	defer func() {
		c.Service.unregister <- c
		c.Conn.Close()
	}()
	
	c.Conn.SetReadLimit(c.Service.config.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.Service.config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.Service.config.PongWait))
		return nil
	})
	
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket读取错误: %v", err)
			}
			break
		}
		
		// 处理消息
		c.handleMessage(message)
	}
}

// writePump 写入消息泵
func (c *Client) writePump() {
	ticker := time.NewTicker(c.Service.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Service.config.WriteWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			
			w.Write(message)
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Service.config.WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理消息
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		c.sendError("消息格式错误")
		return
	}
	
	msg.From = c.ID
	msg.Timestamp = time.Now()
	
	switch msg.Type {
	case "join_room":
		c.joinRoom(msg.RoomID)
	case "leave_room":
		c.leaveRoom(msg.RoomID)
	case "room_message":
		c.sendRoomMessage(msg.RoomID, &msg)
	case "private_message":
		c.sendPrivateMessage(msg.To, &msg)
	case "broadcast":
		c.Service.broadcastMessage(&msg)
	default:
		c.sendError("未知消息类型")
	}
}

// joinRoom 加入房间
func (c *Client) joinRoom(roomID string) {
	room := c.Service.GetRoom(roomID)
	if room == nil {
		room = c.Service.CreateRoom(roomID, "Room "+roomID, "")
	}
	
	room.addClient(c)
	c.rooms[roomID] = room
	
	// 发送加入确认
	msg := &Message{
		Type:      "room_joined",
		RoomID:    roomID,
		Content:   "成功加入房间",
		Timestamp: time.Now(),
	}
	c.sendMessage(msg)
}

// leaveRoom 离开房间
func (c *Client) leaveRoom(roomID string) {
	room := c.Service.GetRoom(roomID)
	if room != nil {
		room.removeClient(c)
		delete(c.rooms, roomID)
		
		// 发送离开确认
		msg := &Message{
			Type:      "room_left",
			RoomID:    roomID,
			Content:   "已离开房间",
			Timestamp: time.Now(),
		}
		c.sendMessage(msg)
	}
}

// leaveAllRooms 离开所有房间
func (c *Client) leaveAllRooms() {
	for roomID := range c.rooms {
		c.leaveRoom(roomID)
	}
}

// sendRoomMessage 发送房间消息
func (c *Client) sendRoomMessage(roomID string, message *Message) {
	room := c.Service.GetRoom(roomID)
	if room != nil {
		room.broadcast(message)
	}
}

// sendPrivateMessage 发送私聊消息
func (c *Client) sendPrivateMessage(to string, message *Message) {
	// 查找目标客户端
	c.Service.clientsMu.RLock()
	for client := range c.Service.clients {
		if client.ID == to {
			client.sendMessage(message)
			break
		}
	}
	c.Service.clientsMu.RUnlock()
}

// sendMessage 发送消息
func (c *Client) sendMessage(message *Message) {
	select {
	case c.send <- message.ToJSON():
	default:
		close(c.send)
		delete(c.Service.clients, c)
	}
}

// sendError 发送错误消息
func (c *Client) sendError(content string) {
	msg := &Message{
		Type:      "error",
		Content:   content,
		Timestamp: time.Now(),
	}
	c.sendMessage(msg)
}

// Room方法

// addClient 添加客户端到房间
func (r *Room) addClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.Clients[client] = true
}

// removeClient 从房间移除客户端
func (r *Room) removeClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.Clients, client)
}

// broadcast 广播消息到房间
func (r *Room) broadcast(message *Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for client := range r.Clients {
		select {
		case client.send <- message.ToJSON():
		default:
			close(client.send)
			delete(r.Clients, client)
		}
	}
	
	r.MessageCount++
}

// GetClientCount 获取房间客户端数量
func (r *Room) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return len(r.Clients)
}

// Message方法

// ToJSON 转换为JSON
func (m *Message) ToJSON() []byte {
	data, _ := json.Marshal(m)
	return data
}

// 辅助函数

// generateClientID 生成客户端ID
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
