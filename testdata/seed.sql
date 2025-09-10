-- 测试数据种子文件
-- 用于初始化测试数据库的基础数据

-- 插入测试用户
INSERT INTO users (username, email, password, role, status, created_at, updated_at) VALUES
('admin', 'admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 'active', NOW(), NOW()),
('testuser1', 'testuser1@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user', 'active', NOW(), NOW()),
('testuser2', 'testuser2@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user', 'active', NOW(), NOW()),
('inactiveuser', 'inactiveuser@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user', 'inactive', NOW(), NOW());

-- 插入测试标签
INSERT INTO tags (name, description, status, created_at, updated_at) VALUES
('技术', '技术相关标签', 'active', NOW(), NOW()),
('设计', '设计相关标签', 'active', NOW(), NOW()),
('产品', '产品相关标签', 'active', NOW(), NOW()),
('运营', '运营相关标签', 'active', NOW(), NOW()),
('测试标签', '用于测试的标签', 'inactive', NOW(), NOW());

-- 插入测试API密钥
INSERT INTO api_keys (user_id, name, key_hash, permissions, status, expires_at, created_at, updated_at) VALUES
(1, 'Admin API Key', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '["read","write","admin"]', 'active', DATE_ADD(NOW(), INTERVAL 1 YEAR), NOW(), NOW()),
(2, 'User API Key', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '["read","write"]', 'active', DATE_ADD(NOW(), INTERVAL 1 YEAR), NOW(), NOW()),
(3, 'Read Only API Key', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '["read"]', 'active', DATE_ADD(NOW(), INTERVAL 1 YEAR), NOW(), NOW());

-- 插入测试审计日志
INSERT INTO audit_logs (operator_id, target_user_id, username, action, target_type, target_id, description, ip_address, user_agent, created_at) VALUES
(1, 2, 'admin', 'create_user', 'user', 2, '创建测试用户1', '127.0.0.1', 'Mozilla/5.0 (Test)', NOW()),
(1, 3, 'admin', 'create_user', 'user', 3, '创建测试用户2', '127.0.0.1', 'Mozilla/5.0 (Test)', NOW()),
(2, 2, 'testuser1', 'update_profile', 'user', 2, '更新个人资料', '127.0.0.1', 'Mozilla/5.0 (Test)', NOW()),
(3, 3, 'testuser2', 'update_profile', 'user', 3, '更新个人资料', '127.0.0.1', 'Mozilla/5.0 (Test)', NOW());

-- 插入测试日志监控规则
INSERT INTO log_monitor_rules (name, description, log_type, condition_type, condition_value, severity, enabled, created_at, updated_at) VALUES
('错误日志监控', '监控错误日志数量', 'error', 'count', '10', 'high', 1, NOW(), NOW()),
('慢查询监控', '监控慢查询响应时间', 'sql', 'duration', '1000', 'medium', 1, NOW(), NOW()),
('安全事件监控', '监控安全相关日志', 'security', 'count', '5', 'high', 1, NOW(), NOW());

-- 插入测试日志告警
INSERT INTO log_alerts (rule_id, name, description, severity, status, triggered_at, resolved_at, created_at, updated_at) VALUES
(1, '错误日志过多', '检测到大量错误日志', 'high', 'open', NOW(), NULL, NOW(), NOW()),
(2, '慢查询检测', '检测到慢查询', 'medium', 'acknowledged', NOW(), NULL, NOW(), NOW()),
(3, '安全事件', '检测到安全事件', 'high', 'resolved', NOW(), NOW(), NOW(), NOW());

-- 插入测试WebSocket房间
INSERT INTO websocket_rooms (name, description, max_users, status, created_at, updated_at) VALUES
('general', '通用聊天室', 100, 'active', NOW(), NOW()),
('tech', '技术讨论室', 50, 'active', NOW(), NOW()),
('test_room', '测试房间', 10, 'inactive', NOW(), NOW());

-- 插入测试WebSocket消息
INSERT INTO websocket_messages (room_id, user_id, username, message_type, content, created_at) VALUES
(1, 2, 'testuser1', 'text', 'Hello, everyone!', NOW()),
(1, 3, 'testuser2', 'text', 'Hi there!', NOW()),
(2, 2, 'testuser1', 'text', 'Any tech questions?', NOW()),
(2, 3, 'testuser2', 'text', 'Yes, I have one.', NOW());

-- 插入测试日志统计
INSERT INTO log_statistics (log_type, count, date, created_at) VALUES
('request', 1000, CURDATE(), NOW()),
('sql', 500, CURDATE(), NOW()),
('error', 10, CURDATE(), NOW()),
('audit', 200, CURDATE(), NOW()),
('security', 5, CURDATE(), NOW()),
('business', 300, CURDATE(), NOW()),
('access', 800, CURDATE(), NOW());

-- 插入测试性能指标
INSERT INTO performance_metrics (metric_name, metric_value, unit, timestamp, created_at) VALUES
('response_time_avg', 150, 'ms', NOW(), NOW()),
('response_time_p95', 300, 'ms', NOW(), NOW()),
('response_time_p99', 500, 'ms', NOW(), NOW()),
('throughput', 100, 'req/s', NOW(), NOW()),
('error_rate', 0.01, 'percent', NOW(), NOW()),
('cpu_usage', 45.5, 'percent', NOW(), NOW()),
('memory_usage', 67.2, 'percent', NOW(), NOW()),
('disk_usage', 23.8, 'percent', NOW(), NOW());

-- 插入测试系统配置
INSERT INTO system_configs (config_key, config_value, description, created_at, updated_at) VALUES
('app_name', 'Cloud Platform API', '应用程序名称', NOW(), NOW()),
('app_version', '1.0.0', '应用程序版本', NOW(), NOW()),
('debug_mode', 'false', '调试模式', NOW(), NOW()),
('maintenance_mode', 'false', '维护模式', NOW(), NOW()),
('max_file_size', '10485760', '最大文件大小（字节）', NOW(), NOW()),
('allowed_file_types', 'jpg,png,pdf,doc,docx', '允许的文件类型', NOW(), NOW()),
('session_timeout', '3600', '会话超时时间（秒）', NOW(), NOW()),
('rate_limit_enabled', 'true', '启用限流', NOW(), NOW()),
('rate_limit_requests', '100', '限流请求数', NOW(), NOW()),
('rate_limit_window', '3600', '限流时间窗口（秒）', NOW(), NOW());

-- 插入测试通知配置
INSERT INTO notification_configs (type, name, config, enabled, created_at, updated_at) VALUES
('email', 'SMTP配置', '{"host":"smtp.example.com","port":587,"username":"noreply@example.com","password":"password","encryption":"tls"}', 1, NOW(), NOW()),
('webhook', 'Slack通知', '{"url":"https://hooks.slack.com/services/xxx","channel":"#alerts","username":"API Monitor"}', 1, NOW(), NOW()),
('webhook', '钉钉通知', '{"url":"https://oapi.dingtalk.com/robot/send?access_token=xxx","secret":"secret"}', 0, NOW(), NOW());

-- 插入测试用户会话
INSERT INTO user_sessions (user_id, session_token, ip_address, user_agent, expires_at, created_at) VALUES
(2, 'session_token_1', '127.0.0.1', 'Mozilla/5.0 (Test)', DATE_ADD(NOW(), INTERVAL 1 HOUR), NOW()),
(3, 'session_token_2', '127.0.0.1', 'Mozilla/5.0 (Test)', DATE_ADD(NOW(), INTERVAL 1 HOUR), NOW()),
(1, 'session_token_3', '127.0.0.1', 'Mozilla/5.0 (Test)', DATE_ADD(NOW(), INTERVAL 1 HOUR), NOW());

-- 插入测试用户活动记录
INSERT INTO user_activities (user_id, activity_type, description, ip_address, user_agent, created_at) VALUES
(2, 'login', '用户登录', '127.0.0.1', 'Mozilla/5.0 (Test)', NOW()),
(2, 'profile_update', '更新个人资料', '127.0.0.1', 'Mozilla/5.0 (Test)', NOW()),
(3, 'login', '用户登录', '127.0.0.1', 'Mozilla/5.0 (Test)', NOW()),
(3, 'api_call', '调用API接口', '127.0.0.1', 'Mozilla/5.0 (Test)', NOW()),
(1, 'admin_action', '管理员操作', '127.0.0.1', 'Mozilla/5.0 (Test)', NOW());

-- 插入测试系统事件
INSERT INTO system_events (event_type, event_level, title, description, details, created_at) VALUES
('system_startup', 'info', '系统启动', '系统成功启动', '{"version":"1.0.0","timestamp":"2024-12-01T00:00:00Z"}', NOW()),
('database_backup', 'info', '数据库备份', '数据库备份完成', '{"backup_size":"100MB","duration":"5m30s"}', NOW()),
('security_alert', 'warning', '安全警告', '检测到异常登录尝试', '{"ip":"192.168.1.100","attempts":5}', NOW()),
('performance_alert', 'warning', '性能警告', '响应时间超过阈值', '{"avg_response_time":"800ms","threshold":"500ms"}', NOW()),
('maintenance_scheduled', 'info', '维护计划', '系统维护计划', '{"start_time":"2024-12-02T02:00:00Z","duration":"2h"}', NOW());
