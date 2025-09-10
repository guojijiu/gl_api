#!/bin/bash

# Cloud Platform API 部署脚本
# 功能说明：
# 1. 自动构建和部署应用
# 2. 支持环境配置
# 3. 数据库迁移
# 4. 服务管理

set -e

# 配置变量
APP_NAME="cloud-platform-api"
APP_VERSION="1.0.0"
BUILD_DIR="build"
DEPLOY_DIR="/opt/cloud-platform-api"
SERVICE_NAME="cloud-platform-api"
ENVIRONMENT=${1:-production}

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查系统依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go未安装，请先安装Go 1.19+"
        exit 1
    fi
    
    # 检查Git
    if ! command -v git &> /dev/null; then
        log_error "Git未安装，请先安装Git"
        exit 1
    fi
    
    # 检查Docker（可选）
    if command -v docker &> /dev/null; then
        log_info "Docker已安装"
    else
        log_warn "Docker未安装，将跳过容器化部署"
    fi
    
    log_info "依赖检查完成"
}

# 清理构建目录
clean_build() {
    log_info "清理构建目录..."
    rm -rf $BUILD_DIR
    mkdir -p $BUILD_DIR
}

# 构建应用
build_app() {
    log_info "构建应用..."
    
    # 设置环境变量
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=amd64
    
    # 构建
    go build -ldflags="-s -w" -o $BUILD_DIR/$APP_NAME main.go
    
    if [ $? -eq 0 ]; then
        log_info "应用构建成功"
    else
        log_error "应用构建失败"
        exit 1
    fi
}

# 创建部署目录
create_deploy_dir() {
    log_info "创建部署目录..."
    sudo mkdir -p $DEPLOY_DIR
    sudo mkdir -p $DEPLOY_DIR/{config,logs,storage}
    sudo chown -R $USER:$USER $DEPLOY_DIR
}

# 复制文件
copy_files() {
    log_info "复制应用文件..."
    
    # 复制可执行文件
    cp $BUILD_DIR/$APP_NAME $DEPLOY_DIR/
    
    # 复制配置文件
    if [ -f "env.example" ]; then
        cp env.example $DEPLOY_DIR/config/
    fi
    
    # 复制其他必要文件
    cp -r storage/* $DEPLOY_DIR/storage/ 2>/dev/null || true
    
    log_info "文件复制完成"
}

# 创建系统服务
create_service() {
    log_info "创建系统服务..."
    
    cat > /tmp/$SERVICE_NAME.service << EOF
[Unit]
Description=Cloud Platform API
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/$APP_NAME
Restart=always
RestartSec=5
Environment=GIN_MODE=release
Environment=ENVIRONMENT=$ENVIRONMENT

[Install]
WantedBy=multi-user.target
EOF
    
    sudo cp /tmp/$SERVICE_NAME.service /etc/systemd/system/
    sudo systemctl daemon-reload
    sudo systemctl enable $SERVICE_NAME
    
    log_info "系统服务创建完成"
}

# 数据库迁移
run_migrations() {
    log_info "执行数据库迁移..."
    
    cd $DEPLOY_DIR
    ./$APP_NAME migrate
    
    if [ $? -eq 0 ]; then
        log_info "数据库迁移完成"
    else
        log_warn "数据库迁移失败，请手动检查"
    fi
}

# 启动服务
start_service() {
    log_info "启动服务..."
    
    sudo systemctl start $SERVICE_NAME
    sudo systemctl status $SERVICE_NAME --no-pager
    
    log_info "服务启动完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 等待服务启动
    sleep 5
    
    # 检查服务状态
    if sudo systemctl is-active --quiet $SERVICE_NAME; then
        log_info "服务运行正常"
    else
        log_error "服务启动失败"
        sudo systemctl status $SERVICE_NAME --no-pager
        exit 1
    fi
    
    # 检查API健康端点
    if command -v curl &> /dev/null; then
        if curl -f http://localhost:8080/api/v1/health > /dev/null 2>&1; then
            log_info "API健康检查通过"
        else
            log_warn "API健康检查失败，请手动验证"
        fi
    fi
}

# 清理临时文件
cleanup() {
    log_info "清理临时文件..."
    rm -rf $BUILD_DIR
    rm -f /tmp/$SERVICE_NAME.service
}

# 显示部署信息
show_deployment_info() {
    log_info "部署完成！"
    echo ""
    echo "部署信息："
    echo "  应用名称: $APP_NAME"
    echo "  版本: $APP_VERSION"
    echo "  部署目录: $DEPLOY_DIR"
    echo "  服务名称: $SERVICE_NAME"
    echo "  环境: $ENVIRONMENT"
    echo ""
    echo "常用命令："
    echo "  查看服务状态: sudo systemctl status $SERVICE_NAME"
    echo "  启动服务: sudo systemctl start $SERVICE_NAME"
    echo "  停止服务: sudo systemctl stop $SERVICE_NAME"
    echo "  重启服务: sudo systemctl restart $SERVICE_NAME"
    echo "  查看日志: sudo journalctl -u $SERVICE_NAME -f"
    echo ""
    echo "API文档: http://localhost:8080/api/documentation"
    echo "健康检查: http://localhost:8080/api/v1/health"
}

# 主函数
main() {
    log_info "开始部署 Cloud Platform API..."
    log_info "环境: $ENVIRONMENT"
    
    check_dependencies
    clean_build
    build_app
    create_deploy_dir
    copy_files
    create_service
    run_migrations
    start_service
    health_check
    cleanup
    show_deployment_info
    
    log_info "部署完成！"
}

# 错误处理
trap 'log_error "部署过程中发生错误，请检查日志"; exit 1' ERR

# 执行主函数
main "$@"

