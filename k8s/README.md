# Kubernetes 部署配置

本目录包含云平台API的Kubernetes部署配置文件，支持生产环境的高可用部署。

## 📋 目录

- [文件说明](#-文件说明)
- [部署步骤](#-部署步骤)
- [配置说明](#-配置说明)
- [监控和日志](#-监控和日志)
- [安全配置](#-安全配置)
- [故障排除](#-故障排除)
- [更新部署](#-更新部署)
- [最佳实践](#-最佳实践)

## 📁 文件说明

### 核心部署文件
- `namespace.yaml` - 命名空间配置，定义资源隔离
- `deployment.yaml` - 应用部署配置，定义Pod副本和更新策略
- `service.yaml` - 服务配置，定义内部网络访问
- `ingress.yaml` - 入口配置，定义外部访问规则

### 配置管理
- `configmap.yaml` - 配置映射，存储非敏感配置
- `secret.yaml` - 敏感信息配置，存储密码、密钥等

### 扩展功能
- `hpa.yaml` - 水平Pod自动扩缩容，根据CPU/内存自动调整副本数
- `pdb.yaml` - Pod中断预算，确保服务可用性
- `networkpolicy.yaml` - 网络策略，控制Pod间通信
- `rbac.yaml` - 基于角色的访问控制，定义权限管理

### 监控和告警
- `monitoring/` - 监控相关配置
  - `prometheus.yaml` - Prometheus配置
  - `grafana/` - Grafana仪表板配置
  - `alertmanager.yaml` - 告警管理配置

## 🚀 部署步骤

### 前置要求

#### 1. 环境准备
```bash
# 安装kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# 安装Helm（可选）
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# 验证安装
kubectl version --client
helm version
```

#### 2. 集群准备
```bash
# 检查集群状态
kubectl cluster-info

# 检查节点状态
kubectl get nodes

# 检查存储类
kubectl get storageclass
```

### 部署流程

#### 1. 创建命名空间
```bash
# 创建命名空间
kubectl apply -f namespace.yaml

# 验证命名空间
kubectl get namespace cloud-platform
```

#### 2. 创建配置
```bash
# 创建ConfigMap
kubectl apply -f configmap.yaml

# 创建Secret（需要先修改secret.yaml中的敏感信息）
kubectl apply -f secret.yaml

# 验证配置
kubectl get configmap -n cloud-platform
kubectl get secret -n cloud-platform
```

#### 3. 创建RBAC
```bash
# 创建RBAC资源
kubectl apply -f rbac.yaml

# 验证RBAC
kubectl get serviceaccount -n cloud-platform
kubectl get role -n cloud-platform
kubectl get rolebinding -n cloud-platform
```

#### 4. 部署应用
```bash
# 部署应用
kubectl apply -f deployment.yaml

# 创建服务
kubectl apply -f service.yaml

# 验证部署
kubectl get pods -n cloud-platform
kubectl get svc -n cloud-platform
```

#### 5. 配置网络
```bash
# 创建网络策略
kubectl apply -f networkpolicy.yaml

# 创建Ingress
kubectl apply -f ingress.yaml

# 验证网络配置
kubectl get networkpolicy -n cloud-platform
kubectl get ingress -n cloud-platform
```

#### 6. 配置自动扩缩容
```bash
# 创建HPA
kubectl apply -f hpa.yaml

# 创建PDB
kubectl apply -f pdb.yaml

# 验证自动扩缩容
kubectl get hpa -n cloud-platform
kubectl get pdb -n cloud-platform
```

### 一键部署脚本

```bash
#!/bin/bash
# deploy.sh - 一键部署脚本

set -e

echo "开始部署Cloud Platform API到Kubernetes..."

# 检查kubectl是否可用
if ! command -v kubectl &> /dev/null; then
    echo "错误: kubectl未安装"
    exit 1
fi

# 检查集群连接
if ! kubectl cluster-info &> /dev/null; then
    echo "错误: 无法连接到Kubernetes集群"
    exit 1
fi

# 部署顺序
echo "1. 创建命名空间..."
kubectl apply -f namespace.yaml

echo "2. 创建配置..."
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml

echo "3. 创建RBAC..."
kubectl apply -f rbac.yaml

echo "4. 部署应用..."
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

echo "5. 配置网络..."
kubectl apply -f networkpolicy.yaml
kubectl apply -f ingress.yaml

echo "6. 配置自动扩缩容..."
kubectl apply -f hpa.yaml
kubectl apply -f pdb.yaml

echo "7. 等待Pod就绪..."
kubectl wait --for=condition=ready pod -l app=cloud-platform-api -n cloud-platform --timeout=300s

echo "8. 验证部署..."
kubectl get pods -n cloud-platform
kubectl get svc -n cloud-platform
kubectl get ingress -n cloud-platform

echo "部署完成！"
echo "访问地址: http://your-domain.com"
echo "健康检查: http://your-domain.com/api/v1/health"
```

### 使用Helm部署

```bash
# 创建Helm Chart
helm create cloud-platform-api

# 安装应用
helm install cloud-platform-api ./cloud-platform-api -n cloud-platform

# 升级应用
helm upgrade cloud-platform-api ./cloud-platform-api -n cloud-platform

# 查看状态
helm status cloud-platform-api -n cloud-platform

# 卸载应用
helm uninstall cloud-platform-api -n cloud-platform
```

## ⚙️ 配置说明

### 环境变量配置

#### ConfigMap配置
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloud-platform-config
  namespace: cloud-platform
data:
  # 服务器配置
  SERVER_PORT: "8080"
  SERVER_MODE: "production"
  SERVER_BASE_URL: "https://api.yourdomain.com"
  
  # 数据库配置
  DB_DRIVER: "mysql"
  DB_HOST: "mysql-service"
  DB_PORT: "3306"
  DB_DATABASE: "cloud_platform"
  DB_CHARSET: "utf8mb4"
  
  # Redis配置
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  REDIS_DATABASE: "0"
  
  # 监控配置
  MONITORING_ENABLE_METRICS: "true"
  MONITORING_ENABLE_HEALTH_CHECK: "true"
  MONITORING_ENABLE_PROMETHEUS: "true"
  
  # 日志配置
  LOG_LEVEL: "info"
  LOG_FORMAT: "json"
  LOG_OUTPUT: "stdout"
```

#### Secret配置
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloud-platform-secret
  namespace: cloud-platform
type: Opaque
data:
  # 数据库密码 (base64编码)
  DB_PASSWORD: <base64-encoded-password>
  
  # JWT密钥 (base64编码)
  JWT_SECRET: <base64-encoded-jwt-secret>
  
  # Redis密码 (base64编码)
  REDIS_PASSWORD: <base64-encoded-redis-password>
  
  # 邮件配置 (base64编码)
  EMAIL_USERNAME: <base64-encoded-email-username>
  EMAIL_PASSWORD: <base64-encoded-email-password>
```

### 资源限制和请求

#### 资源配额
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: cloud-platform-quota
  namespace: cloud-platform
spec:
  hard:
    requests.cpu: "2"
    requests.memory: "4Gi"
    limits.cpu: "4"
    limits.memory: "8Gi"
    persistentvolumeclaims: "10"
    services: "5"
    secrets: "10"
    configmaps: "10"
```

#### Pod资源限制
```yaml
resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"
```

### 健康检查配置

#### 存活探针
```yaml
livenessProbe:
  httpGet:
    path: /api/v1/health/live
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

#### 就绪探针
```yaml
readinessProbe:
  httpGet:
    path: /api/v1/health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

#### 启动探针
```yaml
startupProbe:
  httpGet:
    path: /api/v1/health/startup
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 30
```

### 自动扩缩容配置

#### HPA配置
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: cloud-platform-hpa
  namespace: cloud-platform
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: cloud-platform-api
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
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
```

#### PDB配置
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: cloud-platform-pdb
  namespace: cloud-platform
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: cloud-platform-api
```

### 网络配置

#### Service配置
```yaml
apiVersion: v1
kind: Service
metadata:
  name: cloud-platform-service
  namespace: cloud-platform
spec:
  selector:
    app: cloud-platform-api
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  type: ClusterIP
```

#### Ingress配置
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cloud-platform-ingress
  namespace: cloud-platform
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - api.yourdomain.com
    secretName: cloud-platform-tls
  rules:
  - host: api.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: cloud-platform-service
            port:
              number: 80
```

### 存储配置

#### PVC配置
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cloud-platform-storage
  namespace: cloud-platform
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: fast-ssd
```

#### 存储挂载
```yaml
volumeMounts:
- name: storage
  mountPath: /app/storage
volumes:
- name: storage
  persistentVolumeClaim:
    claimName: cloud-platform-storage
```

## 📊 监控和日志

### 指标收集

#### Prometheus配置
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: cloud-platform
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
    
    rule_files:
      - "alert_rules.yml"
    
    scrape_configs:
    - job_name: 'cloud-platform-api'
      static_configs:
      - targets: ['cloud-platform-service:80']
      metrics_path: '/api/v1/metrics'
      scrape_interval: 5s
      scrape_timeout: 3s
```

#### ServiceMonitor配置
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cloud-platform-monitor
  namespace: cloud-platform
spec:
  selector:
    matchLabels:
      app: cloud-platform-api
  endpoints:
  - port: http
    path: /api/v1/metrics
    interval: 30s
```

#### 告警规则
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cloud-platform-alerts
  namespace: cloud-platform
spec:
  groups:
  - name: cloud-platform.rules
    rules:
    - alert: HighErrorRate
      expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High error rate detected"
        description: "Error rate is {{ $value }} errors per second"
    
    - alert: HighCPUUsage
      expr: cpu_usage_percent > 80
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High CPU usage detected"
        description: "CPU usage is {{ $value }}%"
    
    - alert: HighMemoryUsage
      expr: memory_usage_percent > 80
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High memory usage detected"
        description: "Memory usage is {{ $value }}%"
```

### 日志配置

#### 日志收集
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: cloud-platform
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush         1
        Log_Level     info
        Daemon        off
        Parsers_File  parsers.conf
        HTTP_Server   On
        HTTP_Listen   0.0.0.0
        HTTP_Port     2020
    
    [INPUT]
        Name              tail
        Path              /var/log/containers/*cloud-platform*.log
        Parser            docker
        Tag               kube.*
        Refresh_Interval  5
        Mem_Buf_Limit     50MB
        Skip_Long_Lines   On
    
    [FILTER]
        Name                kubernetes
        Match               kube.*
        Kube_URL            https://kubernetes.default.svc:443
        Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
        Kube_Tag_Prefix     kube.var.log.containers.
        Merge_Log           On
        Keep_Log            Off
        K8S-Logging.Parser  On
        K8S-Logging.Exclude Off
    
    [OUTPUT]
        Name  es
        Match *
        Host  elasticsearch-service
        Port  9200
        Index cloud-platform-logs
        Type  _doc
```

#### 日志轮转
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: logrotate-config
  namespace: cloud-platform
data:
  logrotate.conf: |
    /var/log/containers/*.log {
        daily
        missingok
        rotate 30
        compress
        delaycompress
        notifempty
        create 0644 root root
        postrotate
            /bin/kill -USR1 $(cat /var/run/fluent-bit.pid 2>/dev/null) 2>/dev/null || true
        endscript
    }
```

### Grafana仪表板

#### 仪表板配置
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboard
  namespace: cloud-platform
data:
  dashboard.json: |
    {
      "dashboard": {
        "title": "Cloud Platform API Dashboard",
        "panels": [
          {
            "title": "Request Rate",
            "type": "graph",
            "targets": [
              {
                "expr": "rate(http_requests_total[5m])",
                "legendFormat": "{{method}} {{endpoint}}"
              }
            ]
          },
          {
            "title": "Response Time",
            "type": "graph",
            "targets": [
              {
                "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
                "legendFormat": "95th percentile"
              }
            ]
          },
          {
            "title": "Error Rate",
            "type": "graph",
            "targets": [
              {
                "expr": "rate(http_requests_total{status=~\"5..\"}[5m])",
                "legendFormat": "5xx errors"
              }
            ]
          }
        ]
      }
    }
```

### 健康检查

#### 健康检查端点
```yaml
# 基本健康检查
GET /api/v1/health
Response: {"status": "ok", "timestamp": "2024-01-01T00:00:00Z"}

# 详细健康检查
GET /api/v1/health/detailed
Response: {
  "status": "ok",
  "database": {"status": "ok", "response_time": "5ms"},
  "redis": {"status": "ok", "response_time": "2ms"},
  "storage": {"status": "ok", "free_space": "50GB"}
}

# 就绪检查
GET /api/v1/health/ready
Response: {"status": "ready", "checks": ["database", "redis", "storage"]}

# 存活检查
GET /api/v1/health/live
Response: {"status": "alive", "uptime": "1h30m45s"}
```

### 监控指标

#### 应用指标
- `http_requests_total` - HTTP请求总数
- `http_request_duration_seconds` - HTTP请求持续时间
- `http_requests_in_flight` - 正在处理的请求数
- `cpu_usage_percent` - CPU使用率
- `memory_usage_bytes` - 内存使用量
- `database_connections_active` - 活跃数据库连接数
- `redis_connections_active` - 活跃Redis连接数

#### 系统指标
- `node_cpu_seconds_total` - 节点CPU使用时间
- `node_memory_MemTotal_bytes` - 节点总内存
- `node_filesystem_size_bytes` - 文件系统大小
- `node_network_receive_bytes_total` - 网络接收字节数
- `node_network_transmit_bytes_total` - 网络发送字节数

## 安全配置

### 网络策略
- 限制Pod间通信
- 允许必要的出站流量
- 阻止不必要的入站流量

### RBAC
- 最小权限原则
- 只允许必要的Kubernetes API访问

## 故障排除

### 查看Pod状态
```bash
kubectl get pods -n cloud-platform
```

### 查看日志
```bash
kubectl logs -f deployment/cloud-platform-api -n cloud-platform
```

### 查看服务状态
```bash
kubectl get svc -n cloud-platform
```

### 查看Ingress状态
```bash
kubectl get ingress -n cloud-platform
```

## 更新部署

### 更新镜像
```bash
kubectl set image deployment/cloud-platform-api cloud-platform-api=your-registry/cloud-platform-api:v1.1.0 -n cloud-platform
```

### 滚动更新
```bash
kubectl rollout status deployment/cloud-platform-api -n cloud-platform
```

### 回滚
```bash
kubectl rollout undo deployment/cloud-platform-api -n cloud-platform
```
