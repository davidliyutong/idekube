# 快速开始指南

本文档提供 idekube-housekeeper 组件的部署和配置指南。

## 概述

idekube-housekeeper 是 idekube 平台的清理和维护组件，负责：

- 清理过期的 Kubernetes 工作空间资源
- 归档工作空间数据
- 执行定期维护任务
- 通过 RabbitMQ 接收清理请求

## 系统要求

### 运行环境

- Kubernetes 1.20+
- PostgreSQL 12+
- RabbitMQ 3.8+
- Go 1.21+（仅用于开发）

### 资源配额

**最小配置**：
- CPU: 100m
- 内存: 128Mi
- 存储: 无特殊要求

**推荐配置**：
- CPU: 500m
- 内存: 512Mi
- 存储: 根据日志需求

**生产环境**：
- CPU: 1000m (1核)
- 内存: 1Gi
- 存储: 10Gi（用于日志和临时文件）

## 本地开发

### 1. 克隆代码

```bash
git clone https://github.com/davidliyutong/idekube.git
cd idekube/components/housekeeper
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置环境变量

创建 `.env` 文件：

```bash
# Kubernetes 配置
export KUBECONFIG=~/.kube/config
export NAMESPACE=default

# PostgreSQL 配置
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=idekube
export POSTGRES_PASSWORD=your_password
export POSTGRES_DB=idekube

# RabbitMQ 配置
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export RABBITMQ_USER=guest
export RABBITMQ_PASSWORD=guest
export RABBITMQ_VHOST=/

# 应用配置
export LOG_LEVEL=debug
export WORKER_THREADS=1
```

加载环境变量：

```bash
source .env
```

### 4. 启动依赖服务

#### 使用 Docker Compose

创建 `docker-compose.dev.yml`：

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: idekube-postgres
    environment:
      POSTGRES_USER: idekube
      POSTGRES_PASSWORD: your_password
      POSTGRES_DB: idekube
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U idekube"]
      interval: 10s
      timeout: 5s
      retries: 5

  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: idekube-rabbitmq
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  rabbitmq_data:
```

启动服务：

```bash
docker-compose -f docker-compose.dev.yml up -d
```

验证服务状态：

```bash
# 检查 PostgreSQL
docker exec idekube-postgres pg_isready -U idekube

# 检查 RabbitMQ
docker exec idekube-rabbitmq rabbitmq-diagnostics ping
```

### 5. 初始化数据库

运行数据库迁移（如果有迁移工具）：

```bash
# 示例：使用 migrate 工具
migrate -path ./migrations -database "postgresql://idekube:your_password@localhost:5432/idekube?sslmode=disable" up
```

或手动执行 SQL：

```bash
psql -h localhost -U idekube -d idekube -f ./migrations/000001_init_schema.up.sql
```

### 6. 构建和运行

```bash
# 构建
make build

# 运行
make run
```

或直接使用 go：

```bash
go run cmd/housekeeper/main.go
```

### 7. 验证运行状态

检查日志输出：

```
INFO  Starting idekube-housekeeper
INFO  Configuration loaded successfully
INFO  Connected to PostgreSQL at localhost:5432
INFO  Connected to RabbitMQ at localhost:5672
INFO  Kubernetes client initialized
INFO  Housekeeper started
DEBUG Housekeeper heartbeat
```

## Docker 部署

### 1. 构建镜像

```bash
# 使用 Makefile
make docker-build

# 或直接使用 docker
docker build -t idekube/housekeeper:latest .
```

### 2. 推送镜像

```bash
# 标记镜像
docker tag idekube/housekeeper:latest your-registry.com/idekube/housekeeper:latest

# 推送到镜像仓库
docker push your-registry.com/idekube/housekeeper:latest
```

### 3. 运行容器

```bash
docker run -d \
  --name idekube-housekeeper \
  --network host \
  -e POSTGRES_HOST=localhost \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=idekube \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=idekube \
  -e RABBITMQ_HOST=localhost \
  -e RABBITMQ_PORT=5672 \
  -e RABBITMQ_USER=guest \
  -e RABBITMQ_PASSWORD=guest \
  -e LOG_LEVEL=info \
  -v ~/.kube/config:/root/.kube/config:ro \
  idekube/housekeeper:latest
```

### 4. 查看日志

```bash
docker logs -f idekube-housekeeper
```

### 5. 停止和清理

```bash
# 停止容器
docker stop idekube-housekeeper

# 删除容器
docker rm idekube-housekeeper
```

## Kubernetes 部署

### 方式一：使用 kubectl

#### 1. 创建命名空间

```bash
kubectl create namespace idekube-system
```

#### 2. 创建 Secret

```bash
kubectl create secret generic housekeeper-secrets \
  --from-literal=postgres-user=idekube \
  --from-literal=postgres-password=your_password \
  --from-literal=rabbitmq-user=guest \
  --from-literal=rabbitmq-password=guest \
  -n idekube-system
```

#### 3. 创建 Deployment

创建 `deployment.yaml`：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: idekube-housekeeper
  namespace: idekube-system
  labels:
    app: idekube-housekeeper
spec:
  replicas: 1
  selector:
    matchLabels:
      app: idekube-housekeeper
  template:
    metadata:
      labels:
        app: idekube-housekeeper
    spec:
      serviceAccountName: idekube-housekeeper
      containers:
      - name: housekeeper
        image: idekube/housekeeper:latest
        imagePullPolicy: Always
        env:
        - name: NAMESPACE
          value: "idekube-system"
        - name: POSTGRES_HOST
          value: "postgresql"
        - name: POSTGRES_PORT
          value: "5432"
        - name: POSTGRES_DB
          value: "idekube"
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: housekeeper-secrets
              key: postgres-user
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: housekeeper-secrets
              key: postgres-password
        - name: RABBITMQ_HOST
          value: "rabbitmq"
        - name: RABBITMQ_PORT
          value: "5672"
        - name: RABBITMQ_VHOST
          value: "/"
        - name: RABBITMQ_USER
          valueFrom:
            secretKeyRef:
              name: housekeeper-secrets
              key: rabbitmq-user
        - name: RABBITMQ_PASSWORD
          valueFrom:
            secretKeyRef:
              name: housekeeper-secrets
              key: rabbitmq-password
        - name: LOG_LEVEL
          value: "info"
        - name: WORKER_THREADS
          value: "1"
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 1000m
            memory: 1Gi
        livenessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - pgrep -f idekube-housekeeper
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - pgrep -f idekube-housekeeper
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### 4. 创建 ServiceAccount 和 RBAC

创建 `rbac.yaml`：

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: idekube-housekeeper
  namespace: idekube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: idekube-housekeeper
rules:
- apiGroups: [""]
  resources: ["namespaces", "pods", "services", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets", "replicasets"]
  verbs: ["get", "list", "watch", "delete"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list", "watch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: idekube-housekeeper
subjects:
- kind: ServiceAccount
  name: idekube-housekeeper
  namespace: idekube-system
roleRef:
  kind: ClusterRole
  name: idekube-housekeeper
  apiGroup: rbac.authorization.k8s.io
```

#### 5. 应用配置

```bash
# 应用 RBAC
kubectl apply -f rbac.yaml

# 应用 Deployment
kubectl apply -f deployment.yaml
```

#### 6. 验证部署

```bash
# 检查 Pod 状态
kubectl get pods -n idekube-system -l app=idekube-housekeeper

# 查看日志
kubectl logs -f deployment/idekube-housekeeper -n idekube-system

# 检查 ServiceAccount
kubectl get serviceaccount idekube-housekeeper -n idekube-system
```

### 方式二：使用 Helm Chart

#### 1. 准备 values.yaml

创建 `values.yaml`（或使用项目根目录的 Helm chart）：

```yaml
housekeeper:
  enabled: true
  replicas: 1
  
  image:
    repository: idekube/housekeeper
    tag: latest
    pullPolicy: Always
  
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 1000m
      memory: 1Gi
  
  env:
    logLevel: info
    workerThreads: 1
  
  postgres:
    host: postgresql
    port: 5432
    database: idekube
    existingSecret: housekeeper-secrets
    userKey: postgres-user
    passwordKey: postgres-password
  
  rabbitmq:
    host: rabbitmq
    port: 5672
    vhost: /
    existingSecret: housekeeper-secrets
    userKey: rabbitmq-user
    passwordKey: rabbitmq-password
  
  serviceAccount:
    create: true
    name: idekube-housekeeper
```

#### 2. 安装 Helm Chart

```bash
# 从项目根目录
cd /path/to/idekube

# 安装或升级
helm upgrade --install idekube ./helm \
  --namespace idekube-system \
  --create-namespace \
  -f values.yaml
```

#### 3. 验证安装

```bash
# 检查发布状态
helm status idekube -n idekube-system

# 查看所有资源
kubectl get all -n idekube-system -l app=idekube-housekeeper
```

#### 4. 卸载

```bash
helm uninstall idekube -n idekube-system
```

## 配置说明

### 环境变量详解

#### Kubernetes 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| KUBECONFIG | - | kubeconfig 文件路径（集群内运行时可省略） |
| NAMESPACE | 空（所有命名空间） | 要监视的命名空间 |

#### PostgreSQL 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| POSTGRES_HOST | localhost | PostgreSQL 主机地址 |
| POSTGRES_PORT | 5432 | PostgreSQL 端口 |
| POSTGRES_USER | - | 数据库用户名（必填） |
| POSTGRES_PASSWORD | - | 数据库密码（必填） |
| POSTGRES_DB | - | 数据库名称（必填） |

连接字符串格式：
```
postgresql://<user>:<password>@<host>:<port>/<database>?sslmode=disable
```

#### RabbitMQ 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| RABBITMQ_HOST | localhost | RabbitMQ 主机地址 |
| RABBITMQ_PORT | 5672 | RabbitMQ 端口 |
| RABBITMQ_USER | - | RabbitMQ 用户名（必填） |
| RABBITMQ_PASSWORD | - | RabbitMQ 密码（必填） |
| RABBITMQ_VHOST | / | RabbitMQ 虚拟主机 |

#### 应用配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| LOG_LEVEL | info | 日志级别：debug, info, warn, error |
| WORKER_THREADS | 1 | 工作线程数量 |

### 日志级别

- **debug**：详细的调试信息，包括所有操作细节
- **info**：常规操作信息，包括启动、关闭和重要事件
- **warn**：警告信息，表示潜在问题但不影响运行
- **error**：错误信息，表示操作失败需要关注

推荐配置：
- 开发环境：`debug`
- 测试环境：`info`
- 生产环境：`info` 或 `warn`

## 健康检查

### Liveness Probe

检查进程是否存活：

```yaml
livenessProbe:
  exec:
    command:
    - /bin/sh
    - -c
    - pgrep -f idekube-housekeeper
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Readiness Probe

检查服务是否就绪：

```yaml
readinessProbe:
  exec:
    command:
    - /bin/sh
    - -c
    - pgrep -f idekube-housekeeper
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

## 监控和日志

### 日志收集

#### 使用 kubectl

```bash
# 实时查看日志
kubectl logs -f deployment/idekube-housekeeper -n idekube-system

# 查看最近 100 行日志
kubectl logs --tail=100 deployment/idekube-housekeeper -n idekube-system

# 查看特定 Pod 的日志
kubectl logs <pod-name> -n idekube-system
```

#### 集成日志系统

**Fluentd/Fluent Bit**：
```yaml
# 示例：添加日志标签
metadata:
  annotations:
    fluentd.io/parse: "json"
    fluentd.io/exclude: "false"
```

**ELK Stack**：
- 配置 Filebeat 收集容器日志
- 使用 JSON 格式便于解析

### 监控指标

推荐监控的指标：

1. **应用指标**：
   - CPU 使用率
   - 内存使用率
   - Goroutine 数量
   - 消息处理速率

2. **业务指标**：
   - 清理任务成功/失败数
   - 平均清理时间
   - 队列积压数量
   - 数据库连接数

3. **依赖服务**：
   - PostgreSQL 连接状态
   - RabbitMQ 连接状态
   - Kubernetes API 响应时间

## 故障排查

### 常见问题

#### 1. 无法连接到 PostgreSQL

**错误信息**：
```
FATAL Failed to connect to PostgreSQL: connection refused
```

**排查步骤**：
```bash
# 检查 PostgreSQL 服务
kubectl get svc postgresql -n idekube-system

# 测试连接
kubectl run -it --rm debug --image=postgres:15 --restart=Never -- \
  psql -h postgresql -U idekube -d idekube -c "SELECT 1;"

# 检查网络策略
kubectl get networkpolicies -n idekube-system
```

#### 2. 无法连接到 RabbitMQ

**错误信息**：
```
FATAL Failed to connect to RabbitMQ: connection refused
```

**排查步骤**：
```bash
# 检查 RabbitMQ 服务
kubectl get svc rabbitmq -n idekube-system

# 测试连接
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl -u guest:guest http://rabbitmq:15672/api/overview

# 检查 RabbitMQ 日志
kubectl logs deployment/rabbitmq -n idekube-system
```

#### 3. 权限不足

**错误信息**：
```
ERROR Failed to delete namespace: forbidden
```

**排查步骤**：
```bash
# 检查 ServiceAccount
kubectl get serviceaccount idekube-housekeeper -n idekube-system

# 检查 ClusterRoleBinding
kubectl get clusterrolebinding idekube-housekeeper

# 验证权限
kubectl auth can-i delete namespaces \
  --as=system:serviceaccount:idekube-system:idekube-housekeeper
```

#### 4. Pod 启动失败

```bash
# 查看 Pod 状态
kubectl describe pod -l app=idekube-housekeeper -n idekube-system

# 查看事件
kubectl get events -n idekube-system --sort-by='.lastTimestamp'

# 检查镜像拉取
kubectl get pods -n idekube-system -o jsonpath='{.items[*].status.containerStatuses[*].state}'
```

### 调试模式

启用调试日志：

```bash
# 临时修改环境变量
kubectl set env deployment/idekube-housekeeper LOG_LEVEL=debug -n idekube-system

# 查看详细日志
kubectl logs -f deployment/idekube-housekeeper -n idekube-system
```

## 升级和回滚

### 升级

```bash
# 使用新镜像
kubectl set image deployment/idekube-housekeeper \
  housekeeper=idekube/housekeeper:v1.1.0 \
  -n idekube-system

# 检查升级状态
kubectl rollout status deployment/idekube-housekeeper -n idekube-system
```

### 回滚

```bash
# 查看版本历史
kubectl rollout history deployment/idekube-housekeeper -n idekube-system

# 回滚到上一个版本
kubectl rollout undo deployment/idekube-housekeeper -n idekube-system

# 回滚到特定版本
kubectl rollout undo deployment/idekube-housekeeper --to-revision=2 -n idekube-system
```

## 安全建议

1. **使用 Secret 管理敏感信息**：
   - 不要在配置文件中硬编码密码
   - 使用 Kubernetes Secret 或外部密钥管理系统

2. **最小权限原则**：
   - ServiceAccount 只授予必要的权限
   - 使用 RBAC 限制资源访问范围

3. **网络隔离**：
   - 使用 NetworkPolicy 限制网络访问
   - 只允许必要的服务间通信

4. **镜像安全**：
   - 使用官方基础镜像
   - 定期更新依赖和基础镜像
   - 扫描镜像漏洞

5. **日志安全**：
   - 不要在日志中输出敏感信息
   - 设置合适的日志保留策略

## 性能优化

1. **资源配额**：
   - 根据实际负载调整 CPU 和内存限制
   - 使用 HPA 进行水平扩展（如果需要）

2. **并发处理**：
   - 增加 WORKER_THREADS 提高并发处理能力
   - 注意数据库连接池大小

3. **批量操作**：
   - 批量删除 Kubernetes 资源
   - 批量更新数据库记录

4. **缓存策略**：
   - 缓存常用的 Kubernetes 资源信息
   - 减少数据库查询频率

## 下一步

- 阅读 [API 定义文档](./API.md) 了解 RabbitMQ 消息格式
- 阅读 [API 测试指南](./API_TESTING.md) 学习如何测试功能
- 查看 [主 README](./README.md) 了解项目架构

## 获取帮助

- GitHub Issues: https://github.com/davidliyutong/idekube/issues
- 项目文档: https://github.com/davidliyutong/idekube
