#!/bin/bash

# 云平台API Docker部署脚本
# 功能：构建和部署Docker容器

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
IMAGE_NAME="cloud-platform-api"
IMAGE_TAG="${IMAGE_TAG:-latest}"
CONTAINER_NAME="cloud-platform-api"
PORT="${PORT:-8080}"
NETWORK_NAME="cloud-platform-network"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker环境
check_docker() {
    log_info "检查Docker环境..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装或未在PATH中"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装或未在PATH中"
        exit 1
    fi
    
    # 检查Docker是否运行
    if ! docker info &> /dev/null; then
        log_error "Docker未运行，请启动Docker服务"
        exit 1
    fi
    
    log_success "Docker环境检查通过"
}

# 构建Docker镜像
build_image() {
    log_info "构建Docker镜像..."
    
    # 构建镜像
    docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .
    
    log_success "Docker镜像构建完成: ${IMAGE_NAME}:${IMAGE_TAG}"
}

# 创建Docker网络
create_network() {
    log_info "创建Docker网络..."
    
    # 检查网络是否存在
    if ! docker network ls | grep -q ${NETWORK_NAME}; then
        docker network create ${NETWORK_NAME}
        log_success "Docker网络创建完成: ${NETWORK_NAME}"
    else
        log_info "Docker网络已存在: ${NETWORK_NAME}"
    fi
}

# 停止现有容器
stop_container() {
    log_info "停止现有容器..."
    
    if docker ps -q -f name=${CONTAINER_NAME} | grep -q .; then
        docker stop ${CONTAINER_NAME}
        log_success "容器已停止: ${CONTAINER_NAME}"
    else
        log_info "没有运行中的容器: ${CONTAINER_NAME}"
    fi
}

# 删除现有容器
remove_container() {
    log_info "删除现有容器..."
    
    if docker ps -aq -f name=${CONTAINER_NAME} | grep -q .; then
        docker rm ${CONTAINER_NAME}
        log_success "容器已删除: ${CONTAINER_NAME}"
    else
        log_info "没有容器: ${CONTAINER_NAME}"
    fi
}

# 运行容器
run_container() {
    log_info "运行容器..."
    
    # 创建数据卷
    docker volume create cloud-platform-data 2>/dev/null || true
    
    # 运行容器
    docker run -d \
        --name ${CONTAINER_NAME} \
        --network ${NETWORK_NAME} \
        -p ${PORT}:8080 \
        -v cloud-platform-data:/app/data \
        -e DATABASE_URL="${DATABASE_URL:-postgres://postgres:password@postgres:5432/cloud_platform?sslmode=disable}" \
        -e REDIS_URL="${REDIS_URL:-redis://redis:6379/0}" \
        -e LOG_LEVEL="${LOG_LEVEL:-info}" \
        --restart unless-stopped \
        ${IMAGE_NAME}:${IMAGE_TAG}
    
    log_success "容器已启动: ${CONTAINER_NAME}"
}

# 检查容器状态
check_container() {
    log_info "检查容器状态..."
    
    # 等待容器启动
    sleep 10
    
    # 检查容器是否运行
    if docker ps -q -f name=${CONTAINER_NAME} | grep -q .; then
        log_success "容器运行正常"
    else
        log_error "容器启动失败"
        docker logs ${CONTAINER_NAME}
        exit 1
    fi
    
    # 检查健康状态
    log_info "检查健康状态..."
    for i in {1..30}; do
        if curl -s http://localhost:${PORT}/api/v1/health > /dev/null; then
            log_success "API服务健康检查通过"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "API服务健康检查失败"
            docker logs ${CONTAINER_NAME}
            exit 1
        fi
        sleep 2
    done
}

# 显示容器信息
show_container_info() {
    log_info "容器信息:"
    echo "  容器名称: ${CONTAINER_NAME}"
    echo "  镜像标签: ${IMAGE_NAME}:${IMAGE_TAG}"
    echo "  端口映射: ${PORT}:8080"
    echo "  网络: ${NETWORK_NAME}"
    echo "  数据卷: cloud-platform-data"
    echo ""
    echo "访问地址:"
    echo "  API: http://localhost:${PORT}/api/v1"
    echo "  健康检查: http://localhost:${PORT}/api/v1/health"
    echo "  API文档: http://localhost:${PORT}/api/v1/docs"
    echo ""
    echo "管理命令:"
    echo "  查看日志: docker logs ${CONTAINER_NAME}"
    echo "  进入容器: docker exec -it ${CONTAINER_NAME} /bin/sh"
    echo "  停止容器: docker stop ${CONTAINER_NAME}"
    echo "  删除容器: docker rm ${CONTAINER_NAME}"
}

# 清理资源
cleanup() {
    log_info "清理资源..."
    
    # 停止容器
    stop_container
    
    # 删除容器
    remove_container
    
    # 删除镜像
    if [ "$1" = "full" ]; then
        log_info "删除Docker镜像..."
        docker rmi ${IMAGE_NAME}:${IMAGE_TAG} 2>/dev/null || true
    fi
    
    log_success "清理完成"
}

# 主函数
main() {
    case "${1:-deploy}" in
        "build")
            check_docker
            build_image
            ;;
        "deploy")
            check_docker
            build_image
            create_network
            stop_container
            remove_container
            run_container
            check_container
            show_container_info
            ;;
        "start")
            check_docker
            create_network
            run_container
            check_container
            show_container_info
            ;;
        "stop")
            stop_container
            ;;
        "restart")
            stop_container
            run_container
            check_container
            show_container_info
            ;;
        "logs")
            docker logs -f ${CONTAINER_NAME}
            ;;
        "shell")
            docker exec -it ${CONTAINER_NAME} /bin/sh
            ;;
        "status")
            docker ps -f name=${CONTAINER_NAME}
            ;;
        "cleanup")
            cleanup
            ;;
        "cleanup-full")
            cleanup full
            ;;
        *)
            echo "用法: $0 [build|deploy|start|stop|restart|logs|shell|status|cleanup|cleanup-full]"
            echo "  build        - 构建Docker镜像"
            echo "  deploy       - 构建并部署容器（默认）"
            echo "  start        - 启动容器"
            echo "  stop         - 停止容器"
            echo "  restart      - 重启容器"
            echo "  logs         - 查看容器日志"
            echo "  shell        - 进入容器shell"
            echo "  status       - 查看容器状态"
            echo "  cleanup      - 清理容器和网络"
            echo "  cleanup-full - 清理容器、网络和镜像"
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
