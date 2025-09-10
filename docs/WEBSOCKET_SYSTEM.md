# WebSocket 实时通信系统

## 📖 概述

WebSocket实时通信系统为云平台API提供了强大的实时通信能力，支持客户端与服务器之间的双向实时通信。系统采用现代化的架构设计，具备高性能、高可用性和强安全性。

## 🏗️ 系统架构

### 核心组件

1. **WebSocketService**: 核心服务，管理连接、房间和消息
2. **WebSocketController**: HTTP API控制器，提供REST接口
3. **Client**: 客户端连接管理
4. **Room**: 房间管理
5. **Message**: 消息结构定义

### 技术特性

- 基于Gorilla WebSocket库
- 支持JSON消息格式
- 连接池管理
- 心跳检测机制
- 异步消息处理
- 房间管理系统

## 🚀 功能特性

### 1. 实时通信
- 双向实时通信
- 低延迟消息传递
- 支持文本和二进制消息
- 自动重连机制

### 2. 房间管理
- 动态创建房间
- 房间加入/离开
- 房间消息广播
- 房间统计信息

### 3. 用户管理
- 在线用户统计
- 用户状态管理
- 用户权限控制
- 用户行为追踪

### 4. 消息系统
- 房间消息
- 私聊消息
- 全局广播
- 消息过滤和验证

### 5. 安全特性
- JWT认证
- IP白名单
- 速率限制
- 消息大小限制

## 📋 API接口

### WebSocket连接

#### 建立连接
```
GET /ws/connect?room_id={room_id}
```

**参数说明:**
- `room_id`: 房间ID（可选，默认为"general"）

**响应示例:**
```json
{
  "type": "welcome",
  "content": "欢迎连接到WebSocket服务",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 房间管理

#### 获取房间列表
```
GET /ws/rooms
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "rooms": [
      {
        "id": "general",
        "name": "General",
        "description": "通用聊天室",
        "client_count": 5,
        "message_count": 120,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 1
  }
}
```

#### 创建房间
```
POST /ws/rooms
```

**请求体:**
```json
{
  "id": "tech-support",
  "name": "技术支持",
  "description": "技术问题讨论"
}
```

#### 加入房间
```
POST /ws/rooms/{room_id}/join
```

#### 离开房间
```
POST /ws/rooms/{room_id}/leave
```

### 消息管理

#### 发送消息
```
POST /ws/messages
```

**请求体:**
```json
{
  "type": "room_message",
  "content": "大家好！",
  "room_id": "general",
  "data": {
    "emoji": "👋"
  }
}
```

**消息类型:**
- `room_message`: 房间消息
- `private_message`: 私聊消息
- `broadcast`: 全局广播

### 用户管理

#### 获取在线用户
```
GET /ws/users/online
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "online_users": 15,
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

### 统计信息

#### 获取系统统计
```
GET /ws/stats
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "total_rooms": 3,
    "online_users": 15,
    "total_messages": 1250,
    "rooms": {
      "general": {
        "name": "General",
        "client_count": 8,
        "message_count": 800
      }
    },
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

## ⚙️ 配置说明

### 环境变量配置

```bash
# WebSocket基础配置
WS_READ_BUFFER_SIZE=1024
WS_WRITE_BUFFER_SIZE=1024
WS_PING_PERIOD=60s
WS_PONG_WAIT=10s
WS_WRITE_WAIT=10s
WS_MAX_MESSAGE_SIZE=512
WS_ENABLE_COMPRESSION=true

# 连接管理
WS_MAX_CONNECTIONS=10000
WS_ENABLE_RATE_LIMIT=true
WS_RATE_LIMIT_PER_MINUTE=1000

# 安全配置
WS_ENABLE_IP_WHITELIST=false
WS_ENABLE_ORIGIN_CHECK=true

# 房间管理
WS_ENABLE_ROOM_LIMIT=true
WS_MAX_ROOMS=100

# 消息管理
WS_ENABLE_MESSAGE_PERSISTENCE=false
WS_MESSAGE_RETENTION_DAYS=30
WS_MAX_MESSAGE_LENGTH=1000
WS_ENABLE_MESSAGE_FILTER=true

# 用户状态
WS_ENABLE_USER_STATUS=true
WS_USER_STATUS_TIMEOUT=5m
WS_ENABLE_USER_PRESENCE=true

# 统计监控
WS_ENABLE_STATISTICS=true
WS_STATISTICS_INTERVAL=1m
WS_ENABLE_PERFORMANCE_METRICS=true
WS_ENABLE_ALERTS=false
```

### 配置文件示例

```yaml
websocket:
  # 基础连接配置
  read_buffer_size: 1024
  write_buffer_size: 1024
  ping_period: 60s
  pong_wait: 10s
  write_wait: 10s
  max_message_size: 512
  enable_compression: true
  
  # 连接管理配置
  max_connections: 10000
  enable_rate_limit: true
  rate_limit_per_minute: 1000
  
  # 安全配置
  enable_ip_whitelist: false
  enable_origin_check: true
  
  # 房间管理配置
  enable_room_limit: true
  max_rooms: 100
  default_rooms:
    - "general"
    - "announcements"
    - "support"
  
  # 消息管理配置
  enable_message_persistence: false
  message_retention_days: 30
  max_message_length: 1000
  enable_message_filter: true
  
  # 用户状态配置
  enable_user_status: true
  user_status_timeout: 5m
  enable_user_presence: true
  
  # 统计和监控配置
  enable_statistics: true
  statistics_interval: 1m
  enable_performance_metrics: true
  enable_alerts: false
  
  # 高级功能配置
  enable_load_balancing: false
  enable_message_queue: false
  enable_message_broadcast: true
  enable_private_messaging: true
```

## 🔧 使用方法

### 1. 初始化WebSocket服务

```go
import (
    "cloud-platform-api/app/Services"
    "cloud-platform-api/app/Config"
)

// 创建配置
config := Config.GetDefaultWebSocketConfig()
config.SetDefaults()

// 创建服务
wsService := Services.NewWebSocketService(config)
```

### 2. 在路由中注册

```go
import (
    "cloud-platform-api/app/Http/Routes"
)

// 注册WebSocket路由
Routes.RegisterWebSocketRoutes(router)
```

### 3. 客户端连接示例

```javascript
// 建立WebSocket连接
const ws = new WebSocket('ws://localhost:8080/ws/connect?room_id=general');

// 连接建立
ws.onopen = function() {
    console.log('WebSocket连接已建立');
    
    // 发送消息
    ws.send(JSON.stringify({
        type: 'room_message',
        content: '大家好！',
        room_id: 'general'
    }));
};

// 接收消息
ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('收到消息:', message);
};

// 连接关闭
ws.onclose = function() {
    console.log('WebSocket连接已关闭');
};

// 错误处理
ws.onerror = function(error) {
    console.error('WebSocket错误:', error);
};
```

### 4. 消息格式

#### 客户端发送消息
```json
{
  "type": "room_message",
  "content": "消息内容",
  "room_id": "房间ID",
  "to": "接收者ID",
  "data": {
    "extra": "额外数据"
  }
}
```

#### 服务器发送消息
```json
{
  "type": "message_type",
  "from": "发送者ID",
  "to": "接收者ID",
  "room_id": "房间ID",
  "content": "消息内容",
  "data": {
    "extra": "额外数据"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 📊 监控和统计

### 性能指标

- 连接数量
- 消息吞吐量
- 响应时间
- 错误率
- 内存使用

### 告警设置

- 连接数超限
- 消息延迟过高
- 错误率超标
- 内存使用过高

## 🔒 安全考虑

### 1. 认证和授权
- 所有API接口都需要JWT认证
- 支持用户权限验证
- 防止未授权访问

### 2. 输入验证
- 消息内容长度限制
- 消息类型验证
- 防止恶意消息

### 3. 速率限制
- 每分钟消息数量限制
- 连接频率限制
- 防止DoS攻击

### 4. IP控制
- 支持IP白名单
- 支持IP黑名单
- 地理位置限制

## 🚨 故障排除

### 常见问题

1. **连接失败**
   - 检查服务器状态
   - 验证认证信息
   - 检查网络连接

2. **消息丢失**
   - 检查连接状态
   - 验证消息格式
   - 查看错误日志

3. **性能问题**
   - 检查连接数量
   - 监控消息吞吐量
   - 优化配置参数

### 日志分析

```bash
# 查看WebSocket日志
tail -f logs/websocket.log

# 查看错误日志
grep "ERROR" logs/websocket.log

# 查看性能日志
grep "PERFORMANCE" logs/websocket.log
```

## 🔮 未来规划

### 短期目标
- 消息持久化存储
- 离线消息推送
- 消息搜索功能
- 文件传输支持

### 长期目标
- 分布式部署支持
- 消息队列集成
- 实时数据分析
- AI智能助手

## 📚 相关资源

- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [WebSocket协议规范](https://tools.ietf.org/html/rfc6455)
- [实时通信最佳实践](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)
- [性能优化指南](https://websocket.org/echo.html)

## 🤝 贡献指南

欢迎贡献代码和提出建议！请遵循以下步骤：

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 📄 许可证

本项目采用MIT许可证，详见LICENSE文件。
