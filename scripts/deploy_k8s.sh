#!/bin/bash

# 云平台API Kubernetes部署脚本
# 功能：部署到Kubernetes集群

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
NAMESPACE="${NAMESPACE:-cloud-platform}"
APP_NAME="cloud-platform-api"
IMAGE_NAME="cloud-platform-api"
IMAGE_TAG="${IMAGE_TAG:-latest}"
REPLICAS="${REPLICAS:-3}"

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

# 检查Kubernetes环境
check_k8s() {
    log_info "检查Kubernetes环境..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl未安装或未在PATH中"
        exit 1
    fi
    
    # 检查集群连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到Kubernetes集群"
        exit 1
    fi
    
    log_success "Kubernetes环境检查通过"
}

# 创建命名空间
create_namespace() {
    log_info "创建命名空间: ${NAMESPACE}"
    
    kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "命名空间已创建: ${NAMESPACE}"
}

# 创建ConfigMap
create_configmap() {
    log_info "创建ConfigMap..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${APP_NAME}-config
  namespace: ${NAMESPACE}
data:
  LOG_LEVEL: "info"
  PORT: "8080"
  API_VERSION: "v1"
  CORS_ENABLED: "true"
  RATE_LIMIT_ENABLED: "true"
  MONITORING_ENABLED: "true"
EOF
    
    log_success "ConfigMap已创建"
}

# 创建Secret
create_secret() {
    log_info "创建Secret..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: ${APP_NAME}-secret
  namespace: ${NAMESPACE}
type: Opaque
data:
  DATABASE_URL: $(echo -n "${DATABASE_URL:-postgres://postgres:password@postgres:5432/cloud_platform?sslmode=disable}" | base64)
  REDIS_URL: $(echo -n "${REDIS_URL:-redis://redis:6379/0}" | base64)
  JWT_SECRET: $(echo -n "${JWT_SECRET:-your-secret-key}" | base64)
EOF
    
    log_success "Secret已创建"
}

# 创建Deployment
create_deployment() {
    log_info "创建Deployment..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${APP_NAME}
  namespace: ${NAMESPACE}
  labels:
    app: ${APP_NAME}
spec:
  replicas: ${REPLICAS}
  selector:
    matchLabels:
      app: ${APP_NAME}
  template:
    metadata:
      labels:
        app: ${APP_NAME}
    spec:
      containers:
      - name: ${APP_NAME}
        image: ${IMAGE_NAME}:${IMAGE_TAG}
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: ${APP_NAME}-secret
              key: DATABASE_URL
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: ${APP_NAME}-secret
              key: REDIS_URL
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: ${APP_NAME}-secret
              key: JWT_SECRET
        envFrom:
        - configMapRef:
            name: ${APP_NAME}-config
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/v1/health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: data
          mountPath: /app/data
      volumes:
      - name: data
        emptyDir: {}
EOF
    
    log_success "Deployment已创建"
}

# 创建Service
create_service() {
    log_info "创建Service..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: ${APP_NAME}-service
  namespace: ${NAMESPACE}
  labels:
    app: ${APP_NAME}
spec:
  selector:
    app: ${APP_NAME}
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
  type: ClusterIP
EOF
    
    log_success "Service已创建"
}

# 创建Ingress
create_ingress() {
    log_info "创建Ingress..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ${APP_NAME}-ingress
  namespace: ${NAMESPACE}
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  rules:
  - host: ${INGRESS_HOST:-api.cloudplatform.com}
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: ${APP_NAME}-service
            port:
              number: 80
EOF
    
    log_success "Ingress已创建"
}

# 创建HPA
create_hpa() {
    log_info "创建HPA..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ${APP_NAME}-hpa
  namespace: ${NAMESPACE}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ${APP_NAME}
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
EOF
    
    log_success "HPA已创建"
}

# 等待部署完成
wait_for_deployment() {
    log_info "等待部署完成..."
    
    kubectl wait --for=condition=available --timeout=300s deployment/${APP_NAME} -n ${NAMESPACE}
    
    log_success "部署完成"
}

# 检查部署状态
check_deployment() {
    log_info "检查部署状态..."
    
    # 显示Pod状态
    kubectl get pods -n ${NAMESPACE} -l app=${APP_NAME}
    
    # 显示Service状态
    kubectl get service -n ${NAMESPACE}
    
    # 显示Ingress状态
    kubectl get ingress -n ${NAMESPACE}
    
    # 显示HPA状态
    kubectl get hpa -n ${NAMESPACE}
    
    log_success "部署状态检查完成"
}

# 显示访问信息
show_access_info() {
    log_info "访问信息:"
    
    # 获取Service信息
    SERVICE_IP=$(kubectl get service ${APP_NAME}-service -n ${NAMESPACE} -o jsonpath='{.spec.clusterIP}')
    echo "  Service IP: ${SERVICE_IP}"
    
    # 获取Ingress信息
    INGRESS_IP=$(kubectl get ingress ${APP_NAME}-ingress -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [ -n "${INGRESS_IP}" ]; then
        echo "  Ingress IP: ${INGRESS_IP}"
    fi
    
    # 获取Ingress Host
    INGRESS_HOST=$(kubectl get ingress ${APP_NAME}-ingress -n ${NAMESPACE} -o jsonpath='{.spec.rules[0].host}')
    if [ -n "${INGRESS_HOST}" ]; then
        echo "  Ingress Host: ${INGRESS_HOST}"
        echo "  API URL: http://${INGRESS_HOST}/api/v1"
        echo "  健康检查: http://${INGRESS_HOST}/api/v1/health"
        echo "  API文档: http://${INGRESS_HOST}/api/v1/docs"
    fi
    
    echo ""
    echo "管理命令:"
    echo "  查看Pod: kubectl get pods -n ${NAMESPACE}"
    echo "  查看日志: kubectl logs -f deployment/${APP_NAME} -n ${NAMESPACE}"
    echo "  进入Pod: kubectl exec -it deployment/${APP_NAME} -n ${NAMESPACE} -- /bin/sh"
    echo "  扩缩容: kubectl scale deployment ${APP_NAME} --replicas=5 -n ${NAMESPACE}"
    echo "  删除部署: kubectl delete namespace ${NAMESPACE}"
}

# 清理资源
cleanup() {
    log_info "清理Kubernetes资源..."
    
    # 删除命名空间（会删除所有相关资源）
    kubectl delete namespace ${NAMESPACE} --ignore-not-found=true
    
    log_success "清理完成"
}

# 主函数
main() {
    case "${1:-deploy}" in
        "deploy")
            check_k8s
            create_namespace
            create_configmap
            create_secret
            create_deployment
            create_service
            create_ingress
            create_hpa
            wait_for_deployment
            check_deployment
            show_access_info
            ;;
        "update")
            check_k8s
            create_configmap
            create_secret
            kubectl rollout restart deployment/${APP_NAME} -n ${NAMESPACE}
            wait_for_deployment
            check_deployment
            ;;
        "scale")
            if [ -z "$2" ]; then
                log_error "请指定副本数"
                exit 1
            fi
            kubectl scale deployment ${APP_NAME} --replicas=$2 -n ${NAMESPACE}
            wait_for_deployment
            check_deployment
            ;;
        "status")
            check_deployment
            ;;
        "logs")
            kubectl logs -f deployment/${APP_NAME} -n ${NAMESPACE}
            ;;
        "shell")
            kubectl exec -it deployment/${APP_NAME} -n ${NAMESPACE} -- /bin/sh
            ;;
        "cleanup")
            cleanup
            ;;
        *)
            echo "用法: $0 [deploy|update|scale|status|logs|shell|cleanup]"
            echo "  deploy  - 部署应用（默认）"
            echo "  update  - 更新应用"
            echo "  scale   - 扩缩容应用 (需要指定副本数)"
            echo "  status  - 查看部署状态"
            echo "  logs    - 查看应用日志"
            echo "  shell   - 进入应用Pod"
            echo "  cleanup - 清理所有资源"
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
