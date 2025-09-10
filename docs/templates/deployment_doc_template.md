# 部署文档模板

## 📋 概述

本文档提供系统的详细部署说明，包括开发环境、测试环境和生产环境的部署方法。

## 🔧 环境要求

### 基础要求
- **操作系统**: Linux (推荐 Ubuntu 20.04+), macOS, Windows
- **Go版本**: 1.21+
- **内存**: 最少 2GB RAM
- **磁盘**: 最少 10GB 可用空间

### 数据库要求
- **MySQL**: 5.7+ 或 8.0+
- **PostgreSQL**: 10+
- **SQLite**: 3.x (仅开发环境)

### 可选组件
- **Redis**: 6.0+ (用于缓存和会话存储)
- **Nginx**: 1.18+ (反向代理)
- **Docker**: 20.10+ (容器化部署)

## 🚀 开发环境部署

### 1. 克隆项目
```bash
git clone <repository-url>
cd project-name
```

### 2. 安装依赖
```bash
# 安装Go依赖
go mod download

# 安装开发工具
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 3. 配置环境变量
```bash
# 复制环境变量模板
cp env.example .env

# 编辑配置文件
vim .env
```

**开发环境配置示例：**
```bash
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=debug

# 数据库配置
DB_DRIVER=sqlite
DB_DATABASE=:memory:

# JWT配置
JWT_SECRET=your-secret-key-at-least-32-characters-long

# 日志配置
LOG_LEVEL=debug
```

### 4. 初始化数据库
```bash
# 运行数据库迁移
go run scripts/migrate.go

# 可选：填充测试数据
go run scripts/seed.go
```

### 5. 启动应用
```bash
# 开发模式启动
go run main.go

# 或者使用Makefile
make dev
```

### 6. 验证部署
```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# API文档
open http://localhost:8080/swagger/index.html
```

## 🧪 测试环境部署

### 1. 环境准备
```bash
# 创建测试环境目录
mkdir -p /opt/project-test
cd /opt/project-test

# 下载最新代码
git clone <repository-url> .
git checkout develop
```

### 2. 构建应用
```bash
# 构建应用
make build

# 或手动构建
go build -o project-name main.go
```

### 3. 配置测试环境
```bash
# 创建测试环境配置
cat > .env << EOF
SERVER_PORT=8081
SERVER_MODE=test
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=test_user
DB_PASSWORD=test_password
DB_DATABASE=project_test
JWT_SECRET=test-jwt-secret-key-for-testing-only
REDIS_HOST=localhost
REDIS_PORT=6379
EOF
```

### 4. 数据库设置
```bash
# 创建测试数据库
mysql -u root -p -e "CREATE DATABASE project_test;"
mysql -u root -p -e "CREATE USER 'test_user'@'localhost' IDENTIFIED BY 'test_password';"
mysql -u root -p -e "GRANT ALL PRIVILEGES ON project_test.* TO 'test_user'@'localhost';"

# 运行迁移
./project-name migrate
```

### 5. 启动服务
```bash
# 使用systemd管理服务
sudo tee /etc/systemd/system/project-test.service << EOF
[Unit]
Description=Project Test
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/project-test
ExecStart=/opt/project-test/project-name
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable project-test
sudo systemctl start project-test
```

## 🏭 生产环境部署

### 1. 服务器准备
```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装必要软件
sudo apt install -y nginx mysql-server redis-server supervisor

# 创建应用用户
sudo useradd -r -s /bin/false project-user
```

### 2. 应用部署
```bash
# 创建应用目录
sudo mkdir -p /opt/project
sudo chown project-user:project-user /opt/project

# 下载和构建应用
cd /opt/project
git clone <repository-url> .
make build
```

### 3. 数据库配置
```bash
# 创建生产数据库
sudo mysql -e "CREATE DATABASE project_prod;"
sudo mysql -e "CREATE USER 'project_user'@'localhost' IDENTIFIED BY 'strong_password_here';"
sudo mysql -e "GRANT ALL PRIVILEGES ON project_prod.* TO 'project_user'@'localhost';"
sudo mysql -e "FLUSH PRIVILEGES;"
```

### 4. 环境配置
```bash
# 创建生产环境配置
sudo tee /opt/project/.env << EOF
SERVER_PORT=8080
SERVER_MODE=release
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=project_user
DB_PASSWORD=strong_password_here
DB_DATABASE=project_prod
JWT_SECRET=your-super-secret-jwt-key-change-in-production-must-be-at-least-32-characters-long
REDIS_HOST=localhost
REDIS_PORT=6379
EMAIL_HOST=smtp.yourdomain.com
EMAIL_PORT=587
EMAIL_USERNAME=noreply@yourdomain.com
EMAIL_PASSWORD=your-email-password
EMAIL_FROM=noreply@yourdomain.com
EMAIL_USE_TLS=true
EOF

sudo chown project-user:project-user /opt/project/.env
sudo chmod 600 /opt/project/.env
```

### 5. 运行数据库迁移
```bash
cd /opt/project
sudo -u project-user ./project-name migrate
```

### 6. Nginx配置
```bash
# 创建Nginx配置
sudo tee /etc/nginx/sites-available/project << EOF
server {
    listen 80;
    server_name api.yourdomain.com;

    # 重定向到HTTPS
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    # SSL配置
    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # 客户端最大请求大小
    client_max_body_size 10M;

    # 代理配置
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # 静态文件缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
EOF

# 启用站点
sudo ln -s /etc/nginx/sites-available/project /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 7. SSL证书配置
```bash
# 安装Certbot
sudo apt install -y certbot python3-certbot-nginx

# 获取SSL证书
sudo certbot --nginx -d api.yourdomain.com

# 设置自动续期
sudo crontab -e
# 添加以下行：
# 0 12 * * * /usr/bin/certbot renew --quiet
```

### 8. 服务管理
```bash
# 创建systemd服务
sudo tee /etc/systemd/system/project.service << EOF
[Unit]
Description=Project
After=network.target mysql.service redis.service
Requires=mysql.service redis.service

[Service]
Type=simple
User=project-user
Group=project-user
WorkingDirectory=/opt/project
ExecStart=/opt/project/project-name
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=project

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/project/storage

# 环境变量
Environment=GIN_MODE=release
Environment=GOMAXPROCS=4

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable project
sudo systemctl start project
```

## 🐳 Docker部署

### 1. 构建镜像
```bash
# 构建生产镜像
docker build -t project:latest .

# 或使用多阶段构建
docker build -f Dockerfile.prod -t project:latest .
```

### 2. 创建Docker网络
```bash
docker network create project-network
```

### 3. 启动数据库容器
```bash
# MySQL容器
docker run -d \
  --name mysql \
  --network project-network \
  -e MYSQL_ROOT_PASSWORD=rootpassword \
  -e MYSQL_DATABASE=project \
  -e MYSQL_USER=project_user \
  -e MYSQL_PASSWORD=password \
  -v mysql_data:/var/lib/mysql \
  mysql:8.0

# Redis容器
docker run -d \
  --name redis \
  --network project-network \
  -v redis_data:/data \
  redis:6-alpine
```

### 4. 启动应用容器
```bash
docker run -d \
  --name project \
  --network project-network \
  -p 8080:8080 \
  -e DB_HOST=mysql \
  -e REDIS_HOST=redis \
  -e JWT_SECRET=your-secret-key \
  -v app_storage:/app/storage \
  project:latest
```

### 5. Docker Compose部署
```bash
# 使用docker-compose.yml
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f project
```

## ⚡ 性能优化

### 1. 数据库优化
```sql
-- 创建索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_created_at ON posts(created_at);

-- 优化查询
EXPLAIN SELECT * FROM users WHERE email = 'test@example.com';
```

### 2. 应用优化
```bash
# 设置环境变量
export GOMAXPROCS=4
export GIN_MODE=release

# 启用连接池
export DB_MAX_OPEN_CONNS=100
export DB_MAX_IDLE_CONNS=10
```

### 3. 缓存优化
```bash
# Redis配置优化
sudo tee /etc/redis/redis.conf << EOF
maxmemory 256mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
EOF
```

### 4. Nginx优化
```nginx
# 启用gzip压缩
gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

# 启用缓存
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

## 📊 监控和日志

### 1. 日志配置
```bash
# 创建日志目录
sudo mkdir -p /var/log/project
sudo chown project-user:project-user /var/log/project

# 配置日志轮转
sudo tee /etc/logrotate.d/project << EOF
/var/log/project/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 project-user project-user
    postrotate
        systemctl reload project
    endscript
}
EOF
```

### 2. 监控配置
```bash
# 安装Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.37.0/prometheus-2.37.0.linux-amd64.tar.gz
tar xvf prometheus-*.tar.gz
cd prometheus-*

# 配置Prometheus
cat > prometheus.yml << EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'project'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
EOF
```

### 3. 健康检查
```bash
# 创建健康检查脚本
sudo tee /usr/local/bin/health-check.sh << 'EOF'
#!/bin/bash
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/health)
if [ $response -eq 200 ]; then
    echo "OK"
    exit 0
else
    echo "ERROR"
    exit 1
fi
EOF

sudo chmod +x /usr/local/bin/health-check.sh
```

## 🔧 故障排除

### 1. 常见问题

#### 应用无法启动
```bash
# 检查端口占用
sudo netstat -tlnp | grep :8080

# 检查日志
sudo journalctl -u project -f

# 检查权限
sudo chown -R project-user:project-user /opt/project
```

#### 数据库连接失败
```bash
# 检查MySQL服务状态
sudo systemctl status mysql

# 检查数据库连接
mysql -u project_user -p -h localhost project_prod

# 检查防火墙
sudo ufw status
```

#### Redis连接失败
```bash
# 检查Redis服务状态
sudo systemctl status redis

# 测试Redis连接
redis-cli ping

# 检查Redis配置
sudo cat /etc/redis/redis.conf | grep bind
```

### 2. 性能问题

#### 高CPU使用率
```bash
# 查看进程状态
top -p $(pgrep project)

# 分析goroutine
curl http://localhost:8080/debug/pprof/goroutine?debug=1

# 查看内存使用
curl http://localhost:8080/debug/pprof/heap?debug=1
```

#### 高内存使用率
```bash
# 查看内存使用
free -h

# 分析内存泄漏
go tool pprof http://localhost:8080/debug/pprof/heap

# 强制垃圾回收
curl http://localhost:8080/debug/pprof/heap?debug=1
```

### 3. 安全问题

#### 检查安全配置
```bash
# 检查SSL证书
openssl s_client -connect api.yourdomain.com:443 -servername api.yourdomain.com

# 检查安全头
curl -I https://api.yourdomain.com

# 检查端口开放
sudo nmap -sT localhost
```

#### 更新安全补丁
```bash
# 更新系统
sudo apt update && sudo apt upgrade

# 更新Go版本
wget https://golang.org/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz

# 更新依赖
go mod tidy
go mod download
```

## 📞 支持

如果您在部署过程中遇到问题，请：

1. 查看应用日志：`sudo journalctl -u project -f`
2. 检查系统资源：`htop` 或 `iotop`
3. 查看网络连接：`netstat -tlnp`
4. 提交Issue到项目仓库
5. 联系技术支持团队

## 📚 相关文档

- [API文档](API.md)
- [配置说明](CONFIG.md)
- [开发指南](DEVELOPMENT.md)
- [测试指南](TESTING.md)

---

**文档版本**: 1.0.0  
**最后更新**: 2024年12月  
**维护者**: 运维团队
