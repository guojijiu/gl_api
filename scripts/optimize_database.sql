-- 云平台API数据库优化脚本
-- 包含索引创建、查询优化建议等

-- ==============================================
-- 用户表优化
-- ==============================================

-- 用户名索引（唯一）
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- 邮箱索引（唯一）
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- 状态索引
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- 创建时间索引
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- 复合索引：状态+创建时间
CREATE INDEX IF NOT EXISTS idx_users_status_created_at ON users(status, created_at);

-- 复合索引：角色+状态
CREATE INDEX IF NOT EXISTS idx_users_role_status ON users(role, status);

-- ==============================================
-- 文章表优化
-- ==============================================

-- 状态索引
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status);

-- 分类ID索引
CREATE INDEX IF NOT EXISTS idx_posts_category_id ON posts(category_id);

-- 作者ID索引
CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);

-- 创建时间索引
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);

-- 更新时间索引
CREATE INDEX IF NOT EXISTS idx_posts_updated_at ON posts(updated_at);

-- 复合索引：状态+创建时间（用于首页文章列表）
CREATE INDEX IF NOT EXISTS idx_posts_status_created_at ON posts(status, created_at DESC);

-- 复合索引：分类+状态+创建时间
CREATE INDEX IF NOT EXISTS idx_posts_category_status_created_at ON posts(category_id, status, created_at DESC);

-- 复合索引：作者+状态+创建时间
CREATE INDEX IF NOT EXISTS idx_posts_author_status_created_at ON posts(author_id, status, created_at DESC);

-- 标题全文索引（MySQL）
-- CREATE FULLTEXT INDEX idx_posts_title_fulltext ON posts(title);

-- 内容全文索引（MySQL）
-- CREATE FULLTEXT INDEX idx_posts_content_fulltext ON posts(content);

-- ==============================================
-- 分类表优化
-- ==============================================

-- 分类名索引
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);

-- 分类slug索引（唯一）
CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);

-- 状态索引
CREATE INDEX IF NOT EXISTS idx_categories_status ON categories(status);

-- 父分类ID索引
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);

-- 排序索引
CREATE INDEX IF NOT EXISTS idx_categories_sort_order ON categories(sort_order);

-- 复合索引：状态+排序
CREATE INDEX IF NOT EXISTS idx_categories_status_sort ON categories(status, sort_order);

-- ==============================================
-- 标签表优化
-- ==============================================

-- 标签名索引
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

-- 标签slug索引（唯一）
CREATE INDEX IF NOT EXISTS idx_tags_slug ON tags(slug);

-- 状态索引
CREATE INDEX IF NOT EXISTS idx_tags_status ON tags(status);

-- 使用次数索引
CREATE INDEX IF NOT EXISTS idx_tags_usage_count ON tags(usage_count);

-- 复合索引：状态+使用次数
CREATE INDEX IF NOT EXISTS idx_tags_status_usage ON tags(status, usage_count DESC);

-- ==============================================
-- 文章标签关联表优化
-- ==============================================

-- 文章ID索引
CREATE INDEX IF NOT EXISTS idx_post_tags_post_id ON post_tags(post_id);

-- 标签ID索引
CREATE INDEX IF NOT EXISTS idx_post_tags_tag_id ON post_tags(tag_id);

-- 复合索引：文章+标签（唯一）
CREATE UNIQUE INDEX IF NOT EXISTS idx_post_tags_post_tag ON post_tags(post_id, tag_id);

-- 复合索引：标签+文章
CREATE INDEX IF NOT EXISTS idx_post_tags_tag_post ON post_tags(tag_id, post_id);

-- ==============================================
-- 审计日志表优化
-- ==============================================

-- 用户ID索引
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);

-- 事件类型索引
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);

-- 创建时间索引
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- IP地址索引
CREATE INDEX IF NOT EXISTS idx_audit_logs_ip_address ON audit_logs(ip_address);

-- 复合索引：用户+时间
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created_at ON audit_logs(user_id, created_at DESC);

-- 复合索引：事件类型+时间
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_created_at ON audit_logs(event_type, created_at DESC);

-- ==============================================
-- 监控指标表优化
-- ==============================================

-- 指标类型索引
CREATE INDEX IF NOT EXISTS idx_monitoring_metrics_type ON monitoring_metrics(type);

-- 指标名称索引
CREATE INDEX IF NOT EXISTS idx_monitoring_metrics_name ON monitoring_metrics(name);

-- 时间戳索引
CREATE INDEX IF NOT EXISTS idx_monitoring_metrics_timestamp ON monitoring_metrics(timestamp);

-- 状态索引
CREATE INDEX IF NOT EXISTS idx_monitoring_metrics_status ON monitoring_metrics(status);

-- 复合索引：类型+时间
CREATE INDEX IF NOT EXISTS idx_monitoring_metrics_type_timestamp ON monitoring_metrics(type, timestamp DESC);

-- 复合索引：名称+时间
CREATE INDEX IF NOT EXISTS idx_monitoring_metrics_name_timestamp ON monitoring_metrics(name, timestamp DESC);

-- ==============================================
-- 告警表优化
-- ==============================================

-- 告警规则名索引
CREATE INDEX IF NOT EXISTS idx_alerts_rule_name ON alerts(rule_name);

-- 告警状态索引
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);

-- 严重程度索引
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);

-- 触发时间索引
CREATE INDEX IF NOT EXISTS idx_alerts_fired_at ON alerts(fired_at);

-- 复合索引：状态+触发时间
CREATE INDEX IF NOT EXISTS idx_alerts_status_fired_at ON alerts(status, fired_at DESC);

-- 复合索引：严重程度+状态
CREATE INDEX IF NOT EXISTS idx_alerts_severity_status ON alerts(severity, status);

-- ==============================================
-- 安全事件表优化
-- ==============================================

-- 事件类型索引
CREATE INDEX IF NOT EXISTS idx_security_events_event_type ON security_events(event_type);

-- 事件级别索引
CREATE INDEX IF NOT EXISTS idx_security_events_event_level ON security_events(event_level);

-- 用户ID索引
CREATE INDEX IF NOT EXISTS idx_security_events_user_id ON security_events(user_id);

-- IP地址索引
CREATE INDEX IF NOT EXISTS idx_security_events_ip_address ON security_events(ip_address);

-- 创建时间索引
CREATE INDEX IF NOT EXISTS idx_security_events_created_at ON security_events(created_at);

-- 风险分数索引
CREATE INDEX IF NOT EXISTS idx_security_events_risk_score ON security_events(risk_score);

-- 复合索引：事件类型+时间
CREATE INDEX IF NOT EXISTS idx_security_events_type_created_at ON security_events(event_type, created_at DESC);

-- 复合索引：IP+时间
CREATE INDEX IF NOT EXISTS idx_security_events_ip_created_at ON security_events(ip_address, created_at DESC);

-- 复合索引：用户+时间
CREATE INDEX IF NOT EXISTS idx_security_events_user_created_at ON security_events(user_id, created_at DESC);

-- ==============================================
-- 登录尝试表优化
-- ==============================================

-- 用户名索引
CREATE INDEX IF NOT EXISTS idx_login_attempts_username ON login_attempts(username);

-- IP地址索引
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_address ON login_attempts(ip_address);

-- 成功状态索引
CREATE INDEX IF NOT EXISTS idx_login_attempts_success ON login_attempts(success);

-- 尝试时间索引
CREATE INDEX IF NOT EXISTS idx_login_attempts_attempt_time ON login_attempts(attempt_time);

-- 复合索引：用户名+时间
CREATE INDEX IF NOT EXISTS idx_login_attempts_username_time ON login_attempts(username, attempt_time DESC);

-- 复合索引：IP+时间
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_time ON login_attempts(ip_address, attempt_time DESC);

-- 复合索引：成功状态+时间
CREATE INDEX IF NOT EXISTS idx_login_attempts_success_time ON login_attempts(success, attempt_time DESC);

-- ==============================================
-- 查询优化建议
-- ==============================================

-- 1. 定期分析表统计信息
-- ANALYZE TABLE users, posts, categories, tags, audit_logs, monitoring_metrics, alerts, security_events, login_attempts;

-- 2. 定期优化表
-- OPTIMIZE TABLE users, posts, categories, tags, audit_logs, monitoring_metrics, alerts, security_events, login_attempts;

-- 3. 检查索引使用情况
-- SELECT 
--     TABLE_NAME,
--     INDEX_NAME,
--     CARDINALITY
-- FROM information_schema.STATISTICS 
-- WHERE TABLE_SCHEMA = 'cloud_platform'
-- ORDER BY TABLE_NAME, CARDINALITY DESC;

-- 4. 查找未使用的索引
-- SELECT 
--     s.TABLE_NAME,
--     s.INDEX_NAME,
--     s.CARDINALITY
-- FROM information_schema.STATISTICS s
-- LEFT JOIN information_schema.INDEX_STATISTICS i 
--     ON s.TABLE_NAME = i.TABLE_NAME 
--     AND s.INDEX_NAME = i.INDEX_NAME
-- WHERE s.TABLE_SCHEMA = 'cloud_platform'
--     AND i.INDEX_NAME IS NULL
--     AND s.INDEX_NAME != 'PRIMARY';

-- ==============================================
-- 性能监控查询
-- ==============================================

-- 查看慢查询
-- SELECT 
--     query_time,
--     lock_time,
--     rows_sent,
--     rows_examined,
--     sql_text
-- FROM mysql.slow_log 
-- WHERE start_time > DATE_SUB(NOW(), INTERVAL 1 HOUR)
-- ORDER BY query_time DESC
-- LIMIT 10;

-- 查看表大小
-- SELECT 
--     TABLE_NAME,
--     ROUND(((DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024), 2) AS 'Size (MB)',
--     TABLE_ROWS
-- FROM information_schema.TABLES 
-- WHERE TABLE_SCHEMA = 'cloud_platform'
-- ORDER BY (DATA_LENGTH + INDEX_LENGTH) DESC;

-- 查看索引大小
-- SELECT 
--     TABLE_NAME,
--     INDEX_NAME,
--     ROUND(((STAT_VALUE * @@innodb_page_size) / 1024 / 1024), 2) AS 'Size (MB)'
-- FROM information_schema.INNODB_SYS_TABLESTATS 
-- WHERE TABLE_SCHEMA = 'cloud_platform'
-- ORDER BY STAT_VALUE DESC;

-- ==============================================
-- 分区建议（适用于大表）
-- ==============================================

-- 对于audit_logs表，可以考虑按时间分区
-- ALTER TABLE audit_logs 
-- PARTITION BY RANGE (YEAR(created_at)) (
--     PARTITION p2023 VALUES LESS THAN (2024),
--     PARTITION p2024 VALUES LESS THAN (2025),
--     PARTITION p2025 VALUES LESS THAN (2026),
--     PARTITION p_future VALUES LESS THAN MAXVALUE
-- );

-- 对于monitoring_metrics表，可以考虑按时间分区
-- ALTER TABLE monitoring_metrics 
-- PARTITION BY RANGE (YEAR(timestamp)) (
--     PARTITION p2023 VALUES LESS THAN (2024),
--     PARTITION p2024 VALUES LESS THAN (2025),
--     PARTITION p2025 VALUES LESS THAN (2026),
--     PARTITION p_future VALUES LESS THAN MAXVALUE
-- );

-- ==============================================
-- 配置优化建议
-- ==============================================

-- MySQL配置优化建议：
-- innodb_buffer_pool_size = 1G  # 设置为可用内存的70-80%
-- innodb_log_file_size = 256M
-- innodb_flush_log_at_trx_commit = 2
-- max_connections = 200
-- query_cache_size = 64M
-- query_cache_type = 1
-- tmp_table_size = 64M
-- max_heap_table_size = 64M

-- PostgreSQL配置优化建议：
-- shared_buffers = 256MB
-- effective_cache_size = 1GB
-- maintenance_work_mem = 64MB
-- checkpoint_completion_target = 0.9
-- wal_buffers = 16MB
-- default_statistics_target = 100
