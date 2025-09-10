package Config

import (
	"time"
)

// ConfigError 配置错误类型
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}

// WebSocketConfig WebSocket配置
//
// 重要功能说明：
// 1. 连接配置：缓冲区大小、超时设置、压缩选项
// 2. 性能配置：消息大小限制、连接池大小、心跳间隔
// 3. 安全配置：来源检查、IP白名单、速率限制
// 4. 功能配置：房间管理、消息持久化、用户状态
// 5. 监控配置：统计信息、性能指标、告警设置
//
// 配置项说明：
// - ReadBufferSize: 读取缓冲区大小（字节）
// - WriteBufferSize: 写入缓冲区大小（字节）
// - PingPeriod: 心跳检测间隔
// - PongWait: Pong响应等待时间
// - WriteWait: 写入超时时间
// - MaxMessageSize: 最大消息大小
// - EnableCompression: 是否启用压缩
// - MaxConnections: 最大连接数
// - EnableRateLimit: 是否启用速率限制
// - RateLimitPerMinute: 每分钟速率限制
// - EnableIPWhitelist: 是否启用IP白名单
// - IPWhitelist: IP白名单列表
// - EnableRoomLimit: 是否启用房间数量限制
// - MaxRooms: 最大房间数量
// - EnableMessagePersistence: 是否启用消息持久化
// - MessageRetentionDays: 消息保留天数
// - EnableUserStatus: 是否启用用户状态管理
// - EnableStatistics: 是否启用统计信息
// - StatisticsInterval: 统计信息收集间隔
type WebSocketConfig struct {
	// 基础连接配置
	ReadBufferSize   int           `json:"read_buffer_size" yaml:"read_buffer_size"`
	WriteBufferSize  int           `json:"write_buffer_size" yaml:"write_buffer_size"`
	PingPeriod       time.Duration `json:"ping_period" yaml:"ping_period"`
	PongWait         time.Duration `json:"pong_wait" yaml:"pong_wait"`
	WriteWait        time.Duration `json:"write_wait" yaml:"write_wait"`
	MaxMessageSize   int64         `json:"max_message_size" yaml:"max_message_size"`
	EnableCompression bool         `json:"enable_compression" yaml:"enable_compression"`
	
	// 连接管理配置
	MaxConnections   int           `json:"max_connections" yaml:"max_connections"`
	EnableRateLimit  bool          `json:"enable_rate_limit" yaml:"enable_rate_limit"`
	RateLimitPerMinute int         `json:"rate_limit_per_minute" yaml:"rate_limit_per_minute"`
	
	// 安全配置
	EnableIPWhitelist bool         `json:"enable_ip_whitelist" yaml:"enable_ip_whitelist"`
	IPWhitelist       []string     `json:"ip_whitelist" yaml:"ip_whitelist"`
	EnableOriginCheck bool         `json:"enable_origin_check" yaml:"enable_origin_check"`
	AllowedOrigins    []string     `json:"allowed_origins" yaml:"allowed_origins"`
	
	// 房间管理配置
	EnableRoomLimit   bool         `json:"enable_room_limit" yaml:"enable_room_limit"`
	MaxRooms          int          `json:"max_rooms" yaml:"max_rooms"`
	DefaultRooms      []string     `json:"default_rooms" yaml:"default_rooms"`
	
	// 消息管理配置
	EnableMessagePersistence bool   `json:"enable_message_persistence" yaml:"enable_message_persistence"`
	MessageRetentionDays     int    `json:"message_retention_days" yaml:"message_retention_days"`
	MaxMessageLength         int    `json:"max_message_length" yaml:"max_message_length"`
	EnableMessageFilter      bool   `json:"enable_message_filter" yaml:"enable_message_filter"`
	
	// 用户状态配置
	EnableUserStatus         bool   `json:"enable_user_status" yaml:"enable_user_status"`
	UserStatusTimeout        time.Duration `json:"user_status_timeout" yaml:"user_status_timeout"`
	EnableUserPresence       bool   `json:"enable_user_presence" yaml:"enable_user_presence"`
	
	// 统计和监控配置
	EnableStatistics         bool   `json:"enable_statistics" yaml:"enable_statistics"`
	StatisticsInterval       time.Duration `json:"statistics_interval" yaml:"statistics_interval"`
	EnablePerformanceMetrics bool   `json:"enable_performance_metrics" yaml:"enable_performance_metrics"`
	EnableAlerts             bool   `json:"enable_alerts" yaml:"enable_alerts"`
	
	// 高级功能配置
	EnableLoadBalancing      bool   `json:"enable_load_balancing" yaml:"enable_load_balancing"`
	EnableMessageQueue       bool   `json:"enable_message_queue" yaml:"enable_message_queue"`
	EnableMessageBroadcast   bool   `json:"enable_message_broadcast" yaml:"enable_message_broadcast"`
	EnablePrivateMessaging   bool   `json:"enable_private_messaging" yaml:"enable_private_messaging"`
}

// SetDefaults 设置默认值
func (c *WebSocketConfig) SetDefaults() {
	if c.ReadBufferSize == 0 {
		c.ReadBufferSize = 1024
	}
	if c.WriteBufferSize == 0 {
		c.WriteBufferSize = 1024
	}
	if c.PingPeriod == 0 {
		c.PingPeriod = 60 * time.Second
	}
	if c.PongWait == 0 {
		c.PongWait = 10 * time.Second
	}
	if c.WriteWait == 0 {
		c.WriteWait = 10 * time.Second
	}
	if c.MaxMessageSize == 0 {
		c.MaxMessageSize = 512
	}
	if c.EnableCompression == false {
		c.EnableCompression = true
	}
	if c.MaxConnections == 0 {
		c.MaxConnections = 10000
	}
	if c.EnableRateLimit == false {
		c.EnableRateLimit = true
	}
	if c.RateLimitPerMinute == 0 {
		c.RateLimitPerMinute = 1000
	}
	if c.EnableIPWhitelist == false {
		c.EnableIPWhitelist = false
	}
	if c.EnableOriginCheck == false {
		c.EnableOriginCheck = true
	}
	if c.EnableRoomLimit == false {
		c.EnableRoomLimit = true
	}
	if c.MaxRooms == 0 {
		c.MaxRooms = 100
	}
	if c.DefaultRooms == nil {
		c.DefaultRooms = []string{"general", "announcements", "support"}
	}
	if c.EnableMessagePersistence == false {
		c.EnableMessagePersistence = false
	}
	if c.MessageRetentionDays == 0 {
		c.MessageRetentionDays = 30
	}
	if c.MaxMessageLength == 0 {
		c.MaxMessageLength = 1000
	}
	if c.EnableMessageFilter == false {
		c.EnableMessageFilter = true
	}
	if c.EnableUserStatus == false {
		c.EnableUserStatus = true
	}
	if c.UserStatusTimeout == 0 {
		c.UserStatusTimeout = 5 * time.Minute
	}
	if c.EnableUserPresence == false {
		c.EnableUserPresence = true
	}
	if c.EnableStatistics == false {
		c.EnableStatistics = true
	}
	if c.StatisticsInterval == 0 {
		c.StatisticsInterval = 1 * time.Minute
	}
	if c.EnablePerformanceMetrics == false {
		c.EnablePerformanceMetrics = true
	}
	if c.EnableAlerts == false {
		c.EnableAlerts = false
	}
	if c.EnableLoadBalancing == false {
		c.EnableLoadBalancing = false
	}
	if c.EnableMessageQueue == false {
		c.EnableMessageQueue = false
	}
	if c.EnableMessageBroadcast == false {
		c.EnableMessageBroadcast = true
	}
	if c.EnablePrivateMessaging == false {
		c.EnablePrivateMessaging = true
	}
}

// BindEnvs 绑定环境变量
func (c *WebSocketConfig) BindEnvs() {
	// 这里可以绑定环境变量
	// 例如：viper.BindEnv("websocket.read_buffer_size", "WS_READ_BUFFER_SIZE")
}

// Validate 验证配置
func (c *WebSocketConfig) Validate() error {
	// 验证配置项的有效性
	if c.ReadBufferSize <= 0 {
		return &ConfigError{Field: "ReadBufferSize", Message: "读取缓冲区大小必须大于0"}
	}
	if c.WriteBufferSize <= 0 {
		return &ConfigError{Field: "WriteBufferSize", Message: "写入缓冲区大小必须大于0"}
	}
	if c.PingPeriod <= 0 {
		return &ConfigError{Field: "PingPeriod", Message: "心跳间隔必须大于0"}
	}
	if c.PongWait <= 0 {
		return &ConfigError{Field: "PongWait", Message: "Pong等待时间必须大于0"}
	}
	if c.WriteWait <= 0 {
		return &ConfigError{Field: "WriteWait", Message: "写入超时时间必须大于0"}
	}
	if c.MaxMessageSize <= 0 {
		return &ConfigError{Field: "MaxMessageSize", Message: "最大消息大小必须大于0"}
	}
	if c.MaxConnections <= 0 {
		return &ConfigError{Field: "MaxConnections", Message: "最大连接数必须大于0"}
	}
	if c.RateLimitPerMinute <= 0 {
		return &ConfigError{Field: "RateLimitPerMinute", Message: "速率限制必须大于0"}
	}
	if c.MaxRooms <= 0 {
		return &ConfigError{Field: "MaxRooms", Message: "最大房间数必须大于0"}
	}
	if c.MessageRetentionDays <= 0 {
		return &ConfigError{Field: "MessageRetentionDays", Message: "消息保留天数必须大于0"}
	}
	if c.MaxMessageLength <= 0 {
		return &ConfigError{Field: "MaxMessageLength", Message: "最大消息长度必须大于0"}
	}
	if c.UserStatusTimeout <= 0 {
		return &ConfigError{Field: "UserStatusTimeout", Message: "用户状态超时必须大于0"}
	}
	if c.StatisticsInterval <= 0 {
		return &ConfigError{Field: "StatisticsInterval", Message: "统计间隔必须大于0"}
	}
	
	return nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultWebSocketConfig() *WebSocketConfig {
	config := &WebSocketConfig{}
	config.SetDefaults()
	return config
}
