# Cloud Platform API 部署指南

## 概述

本文档详细介绍了 Cloud Platform API 的部署流程，包括环境准备、配置设置、部署步骤和故障排除。

## 系统要求

### 最低要求
- **CPU**: 2 核心
- **内存**: 4GB RAM
- **存储**: 20GB 可用空间
- **操作系统**: Linux (Ubuntu 20.04+), macOS, Windows 10+

### 推荐配置
- **CPU**: 4 核心
- **内存**: 8GB RAM
- **存储**: 50GB SSD
- **操作系统**: Ubuntu 22.04 LTS

## 依赖服务

### 必需服务
1. **MySQL 8.0+**
   - 用于数据存储
   - 支持 InnoDB 引擎
   - 建议配置主从复制

2. **Go 1.21+**
   - 应用运行环境
   - 支持模块化开发

### 可选服务
1. **Redis 6.0+**
   - 用于缓存和会话存储
   - 提高应用性能

2. **Prometheus + Grafana**
   - 用于监控和可视化
   - 提供系统指标监控

## 环境准备

### 1. 安装 Go

#### Ubuntu/Debian
```bash
# 下载 Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# 解压
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### CentOS/RHEL
```bash
# 下载 Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# 解压
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bash_profile
echo 'export GOPATH=$HOME/go' >> ~/.bash_profile
source ~/.bash_profile

# 验证安装
go version
```

### 2. 安装 MySQL

#### Ubuntu/Debian
```bash
# 更新包列表
sudo apt update

# 安装 MySQL
sudo apt install mysql-server

# 启动 MySQL
sudo systemctl start mysql
sudo systemctl enable mysql

# 安全配置
sudo mysql_secure_installation
```

#### CentOS/RHEL
```bash
# 安装 MySQL
sudo yum install mysql-server

# 启动 MySQL
sudo systemctl start mysqld
sudo systemctl enable mysqld

# 获取临时密码
sudo grep 'temporary password' /var/log/mysqld.log

# 安全配置
sudo mysql_secure_installation
```

### 3. 安装 Redis (可选)

#### Ubuntu/Debian
```bash
# 安装 Redis
sudo apt install redis-server

# 启动 Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server

# 验证安装
redis-cli ping
```

#### CentOS/RHEL
```bash
# 安装 EPEL 仓库
sudo yum install epel-release

# 安装 Redis
sudo yum install redis

# 启动 Redis
sudo systemctl start redis
sudo systemctl enable redis

# 验证安装
redis-cli ping
```

## 配置设置

### 1. 数据库配置

#### 创建数据库和用户
```sql
-- 连接到 MySQL
mysql -u root -p

-- 创建数据库
CREATE DATABASE cloud_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户
CREATE USER 'cloud_platform'@'localhost' IDENTIFIED BY 'secure_password';

-- 授权
GRANT ALL PRIVILEGES ON cloud_platform.* TO 'cloud_platform'@'localhost';
FLUSH PRIVILEGES;

-- 退出
EXIT;
```

#### 配置 MySQL
```bash
# 编辑 MySQL 配置
sudo nano /etc/mysql/mysql.conf.d/mysqld.cnf

# 添加以下配置
[mysqld]
# 基本配置
port = 3306
bind-address = 0.0.0.0
max_connections = 200
max_connect_errors = 1000

# 字符集配置
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci

# InnoDB 配置
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
innodb_flush_log_at_trx_commit = 2

# 重启 MySQL
sudo systemctl restart mysql
```

### 2. 应用配置

#### 创建配置文件
```bash
# 复制示例配置
cp env.example .env

# 编辑配置
nano .env
```

#### 环境变量配置
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=cloud_platform
DB_PASSWORD=secure_password
DB_DATABASE=cloud_platform
DB_CHARSET=utf8mb4

# Redis配置 (可选)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT配置
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRE_HOURS=24

# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30
SERVER_WRITE_TIMEOUT=30

# 日志配置
LOG_LEVEL=info
LOG_FILE_PATH=./storage/logs
LOG_MAX_SIZE=100
LOG_MAX_BACKUPS=5
LOG_MAX_AGE=30

# 监控配置
ENABLE_MONITORING=true
PROMETHEUS_PORT=9090
```

## 部署步骤

### 方法一: 源码部署

#### 1. 获取源码
```bash
# 克隆仓库
git clone https://github.com/your-org/cloud-platform-api.git
cd cloud-platform-api

# 切换到稳定版本
git checkout v1.0.0
```

#### 2. 安装依赖
```bash
# 下载依赖
go mod download

# 验证依赖
go mod verify
```

#### 3. 构建应用
```bash
# 构建应用
go build -o cloud-platform-api .

# 验证构建
./cloud-platform-api --version
```

#### 4. 运行应用
```bash
# 直接运行
./cloud-platform-api

# 后台运行
nohup ./cloud-platform-api > app.log 2>&1 &

# 使用 systemd 管理
sudo cp cloud-platform-api.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable cloud-platform-api
sudo systemctl start cloud-platform-api
```

### 方法二: Docker 部署

#### 1. 构建镜像
```bash
# 构建 Docker 镜像
docker build -t cloud-platform-api:latest .

# 验证镜像
docker images | grep cloud-platform-api
```

#### 2. 运行容器
```bash
# 运行容器
docker run -d \
  --name cloud-platform-api \
  -p 8080:8080 \
  -e DB_HOST=mysql \
  -e DB_USERNAME=cloud_platform \
  -e DB_PASSWORD=secure_password \
  -e DB_DATABASE=cloud_platform \
  --link mysql:mysql \
  cloud-platform-api:latest
```

#### 3. 使用 Docker Compose
```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 方法三: Kubernetes 部署

#### 1. 创建命名空间
```bash
kubectl create namespace cloud-platform
```

#### 2. 创建 ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloud-platform-config
  namespace: cloud-platform
data:
  DB_HOST: "mysql-service"
  DB_PORT: "3306"
  DB_DATABASE: "cloud_platform"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  SERVER_PORT: "8080"
```

#### 3. 创建 Secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloud-platform-secret
  namespace: cloud-platform
type: Opaque
data:
  DB_USERNAME: Y2xvdWRfcGxhdGZvcm0=  # base64 encoded
  DB_PASSWORD: c2VjdXJlX3Bhc3N3b3Jk  # base64 encoded
  JWT_SECRET: eW91ci1zdXBlci1zZWNyZXQ=  # base64 encoded
```

#### 4. 创建 Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloud-platform-api
  namespace: cloud-platform
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cloud-platform-api
  template:
    metadata:
      labels:
        app: cloud-platform-api
    spec:
      containers:
      - name: cloud-platform-api
        image: cloud-platform-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: cloud-platform-config
              key: DB_HOST
        - name: DB_USERNAME
          valueFrom:
            secretKeyRef:
              name: cloud-platform-secret
              key: DB_USERNAME
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## 监控配置

### 1. Prometheus 配置

#### 安装 Prometheus
```bash
# 下载 Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz

# 解压
tar xzf prometheus-2.45.0.linux-amd64.tar.gz
cd prometheus-2.45.0.linux-amd64

# 配置 prometheus.yml
cat > prometheus.yml << EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'cloud-platform-api'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
EOF

# 启动 Prometheus
./prometheus --config.file=prometheus.yml
```

### 2. Grafana 配置

#### 安装 Grafana
```bash
# Ubuntu/Debian
sudo apt-get install -y adduser libfontconfig1
wget https://dl.grafana.com/oss/release/grafana_10.1.0_amd64.deb
sudo dpkg -i grafana_10.1.0_amd64.deb

# 启动 Grafana
sudo systemctl start grafana-server
sudo systemctl enable grafana-server
```

#### 配置数据源
1. 访问 Grafana: http://localhost:3000
2. 默认用户名/密码: admin/admin
3. 添加 Prometheus 数据源
4. URL: http://localhost:9090

## 安全配置

### 1. 防火墙配置
```bash
# Ubuntu/Debian (ufw)
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 8080/tcp  # API
sudo ufw allow 3306/tcp  # MySQL (仅内网)
sudo ufw enable

# CentOS/RHEL (firewalld)
sudo firewall-cmd --permanent --add-port=22/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=3306/tcp
sudo firewall-cmd --reload
```

### 2. SSL/TLS 配置
```bash
# 使用 Let's Encrypt
sudo apt install certbot

# 获取证书
sudo certbot certonly --standalone -d api.yourdomain.com

# 配置 Nginx 反向代理
sudo nano /etc/nginx/sites-available/cloud-platform-api

server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 故障排除

### 1. 常见问题

#### 应用启动失败
```bash
# 检查端口占用
netstat -tlnp | grep :8080

# 检查配置文件
./cloud-platform-api --config-check

# 查看详细日志
./cloud-platform-api --log-level=debug
```

#### 数据库连接失败
```bash
# 检查 MySQL 状态
sudo systemctl status mysql

# 检查数据库连接
mysql -h localhost -u cloud_platform -p cloud_platform

# 检查防火墙
sudo ufw status
```

#### Redis 连接失败
```bash
# 检查 Redis 状态
sudo systemctl status redis

# 测试连接
redis-cli ping

# 检查配置
redis-cli config get "*"
```

### 2. 日志分析

#### 应用日志
```bash
# 查看应用日志
tail -f storage/logs/app.log

# 查看错误日志
tail -f storage/logs/error.log

# 查看业务日志
tail -f storage/logs/business.log
```

#### 系统日志
```bash
# 查看系统日志
sudo journalctl -u cloud-platform-api -f

# 查看 MySQL 日志
sudo tail -f /var/log/mysql/error.log

# 查看 Redis 日志
sudo tail -f /var/log/redis/redis-server.log
```

### 3. 性能调优

#### 数据库优化
```sql
-- 检查慢查询
SHOW VARIABLES LIKE 'slow_query_log';
SHOW VARIABLES LIKE 'long_query_time';

-- 优化查询
EXPLAIN SELECT * FROM users WHERE username = 'testuser';

-- 添加索引
CREATE INDEX idx_username ON users(username);
CREATE INDEX idx_email ON users(email);
```

#### 应用优化
```bash
# 调整 Go 运行时参数
export GOGC=100
export GOMAXPROCS=4

# 监控内存使用
go tool pprof http://localhost:8080/debug/pprof/heap

# 监控 CPU 使用
go tool pprof http://localhost:8080/debug/pprof/profile
```

## 备份与恢复

### 1. 数据库备份
```bash
# 创建备份脚本
cat > backup_db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/backup/mysql"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# 备份数据库
mysqldump -u cloud_platform -p cloud_platform > $BACKUP_DIR/cloud_platform_$DATE.sql

# 压缩备份
gzip $BACKUP_DIR/cloud_platform_$DATE.sql

# 删除7天前的备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
EOF

chmod +x backup_db.sh

# 设置定时任务
crontab -e
# 添加: 0 2 * * * /path/to/backup_db.sh
```

### 2. 应用备份
```bash
# 备份应用文件
tar -czf cloud-platform-api-backup-$(date +%Y%m%d).tar.gz \
  cloud-platform-api \
  storage/ \
  .env

# 备份到远程存储
rsync -avz cloud-platform-api-backup-*.tar.gz user@backup-server:/backup/
```

## 更新升级

### 1. 应用更新
```bash
# 停止应用
sudo systemctl stop cloud-platform-api

# 备份当前版本
cp cloud-platform-api cloud-platform-api.backup

# 下载新版本
wget https://github.com/your-org/cloud-platform-api/releases/download/v1.1.0/cloud-platform-api

# 更新配置（如有需要）
# 编辑 .env 文件

# 启动应用
sudo systemctl start cloud-platform-api

# 验证更新
curl http://localhost:8080/api/v1/health
```

### 2. 数据库迁移
```bash
# 运行数据库迁移
./cloud-platform-api migrate

# 验证迁移结果
mysql -u cloud_platform -p cloud_platform -e "SHOW TABLES;"
```

## 维护计划

### 日常维护
- 监控系统资源使用情况
- 检查应用日志
- 验证备份完整性
- 更新安全补丁

### 定期维护
- 每周检查系统性能
- 每月更新依赖包
- 每季度进行安全审计
- 每年进行灾难恢复测试

## 支持与联系

- **文档**: https://docs.yourdomain.com
- **问题反馈**: https://github.com/your-org/cloud-platform-api/issues
- **技术支持**: support@yourdomain.com
- **紧急联系**: +1-xxx-xxx-xxxx
