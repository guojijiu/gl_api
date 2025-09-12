-- Cloud Platform API 数据库初始化脚本
-- 功能说明：
-- 1. 创建数据库用户和权限
-- 2. 初始化基础数据
-- 3. 设置数据库配置

-- 创建数据库（如果不存在）
-- SELECT 'CREATE DATABASE cloud_platform' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'cloud_platform')\gexec

-- 创建扩展（PostgreSQL）
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 注意：此脚本适用于PostgreSQL
-- 如果使用MySQL，请使用相应的MySQL语法

-- 设置时区
SET timezone = 'UTC';

-- 创建用户表（如果不存在）
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    status INTEGER DEFAULT 1,
    avatar VARCHAR(255),
    email_verified_at TIMESTAMP,
    last_login_at TIMESTAMP,
    login_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建文章表（如果不存在）
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    excerpt VARCHAR(500),
    status INTEGER DEFAULT 1,
    user_id INTEGER REFERENCES users(id),
    category_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建分类表（如果不存在）
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    slug VARCHAR(100) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建标签表（如果不存在）
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) DEFAULT '#007bff',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建文章标签关联表（如果不存在）
CREATE TABLE IF NOT EXISTS post_tags (
    post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

-- 创建审计日志表（如果不存在）
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    table_name VARCHAR(100),
    record_id INTEGER,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建迁移记录表（如果不存在）
CREATE TABLE IF NOT EXISTS migrations (
    id SERIAL PRIMARY KEY,
    migration VARCHAR(255) UNIQUE NOT NULL,
    batch INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_category_id ON posts(category_id);
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- 插入默认管理员用户（如果不存在）
INSERT INTO users (username, email, password, role, status) 
VALUES (
    'admin',
    'admin@example.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password
    'admin',
    1
) ON CONFLICT (username) DO NOTHING;

-- 插入默认分类（如果不存在）
INSERT INTO categories (name, description, slug) VALUES
    ('技术', '技术相关文章', 'tech'),
    ('生活', '生活相关文章', 'life'),
    ('随笔', '随笔文章', 'essay')
ON CONFLICT (slug) DO NOTHING;

-- 插入默认标签（如果不存在）
INSERT INTO tags (name, color) VALUES
    ('Go', '#00ADD8'),
    ('Gin', '#00AC47'),
    ('PostgreSQL', '#336791'),
    ('Redis', '#DC382D'),
    ('Docker', '#2496ED'),
    ('微服务', '#FF6B6B'),
    ('API', '#4ECDC4')
ON CONFLICT (name) DO NOTHING;

-- 创建触发器函数（用于自动更新updated_at字段）
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为表添加更新时间触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_posts_updated_at ON posts;
CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_tags_updated_at ON tags;
CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 创建全文搜索索引
CREATE INDEX IF NOT EXISTS idx_posts_title_content_gin ON posts USING gin(to_tsvector('english', title || ' ' || COALESCE(content, '')));
CREATE INDEX IF NOT EXISTS idx_categories_name_description_gin ON categories USING gin(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- 设置权限
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO cloud_platform;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO cloud_platform;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO cloud_platform;

-- 完成初始化
SELECT 'Database initialization completed successfully!' as status;
