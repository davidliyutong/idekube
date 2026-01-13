# 快速开始指南

本指南将帮助您快速部署和运行idekube-rbac服务。

## 目录

- [前置要求](#前置要求)
- [本地开发](#本地开发)
- [Docker部署](#docker部署)
- [Kubernetes部署](#kubernetes部署)
- [配置说明](#配置说明)
- [验证部署](#验证部署)
- [故障排查](#故障排查)

## 前置要求

### 必需组件

- **Go 1.21+** (本地开发)
- **Docker** (容器化部署)
- **Kubernetes 1.24+** (K8s部署)
- **PostgreSQL 12+** (数据库)
- **RabbitMQ 3.9+** (消息队列)

### 可选组件

- **Helm 3.0+** (用于Kubernetes部署)
- **Make** (构建工具)
- **kubectl** (Kubernetes命令行工具)

## 本地开发

### 1. 克隆仓库

```bash
git clone https://github.com/davidliyutong/idekube.git
cd idekube/components/rbac
```

### 2. 安装依赖

```bash
make deps
```

### 3. 准备数据库

启动PostgreSQL（使用Docker）：

```bash
docker run -d \
  --name idekube-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=idekube_rbac \
  -p 5432:5432 \
  postgres:14-alpine
```

### 4. 准备消息队列

启动RabbitMQ（使用Docker）：

```bash
docker run -d \
  --name idekube-rabbitmq \
  -e RABBITMQ_DEFAULT_USER=guest \
  -e RABBITMQ_DEFAULT_PASS=guest \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3.12-management-alpine
```

### 5. 配置环境变量

创建 `.env` 文件：

```bash
# PostgreSQL配置
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=password
export POSTGRES_DB=idekube_rbac

# RabbitMQ配置
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export RABBITMQ_USER=guest
export RABBITMQ_PASSWORD=guest
export RABBITMQ_VHOST=/

# 应用配置
export HTTP_PORT=8080
export LOG_LEVEL=info
export WORKER_THREADS=1

# Casbin配置
export CASBIN_MODEL_PATH=configs/model.conf
export CASBIN_POLICY_PATH=configs/policy.csv
```

加载环境变量：

```bash
source .env
```

### 6. 初始化Casbin配置

创建 `configs/model.conf`：

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

创建 `configs/policy.csv` (示例策略)：

```csv
p, role:admin, workspace, read
p, role:admin, workspace, write
p, role:admin, workspace, delete
p, role:admin, template, read
p, role:admin, template, write
p, role:admin, template, delete
p, role:editor, workspace, read
p, role:editor, workspace, write
p, role:viewer, workspace, read

g, user:1, role:admin
```

### 7. 构建并运行

```bash
# 构建
make build

# 运行
make run
```

服务将在 `http://localhost:8080` 上启动。

### 8. 验证

```bash
# 健康检查
curl http://localhost:8080/healthz

# 访问Swagger UI
open http://localhost:8080/swagger/
```

## Docker部署

### 1. 构建Docker镜像

```bash
make docker-build
```

这将创建 `davidliyutong/idekube-rbac:latest` 镜像。

### 2. 使用Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: idekube_rbac
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  rabbitmq:
    image: rabbitmq:3.12-management-alpine
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
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

  rbac:
    image: davidliyutong/idekube-rbac:latest
    depends_on:
      postgres:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
    environment:
      POSTGRES_HOST: postgres
      POSTGRES_PORT: 5432
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: idekube_rbac
      RABBITMQ_HOST: rabbitmq
      RABBITMQ_PORT: 5672
      RABBITMQ_USER: guest
      RABBITMQ_PASSWORD: guest
      RABBITMQ_VHOST: /
      HTTP_PORT: 8080
      LOG_LEVEL: info
      CASBIN_MODEL_PATH: /app/configs/model.conf
      CASBIN_POLICY_PATH: /app/configs/policy.csv
    ports:
      - "8080:8080"
    volumes:
      - ./configs:/app/configs

volumes:
  postgres_data:
  rabbitmq_data:
```

### 3. 启动服务

```bash
docker-compose up -d
```

### 4. 查看日志

```bash
docker-compose logs -f rbac
```

### 5. 停止服务

```bash
docker-compose down
```

## Kubernetes部署

### 使用Helm Chart

idekube提供了完整的Helm Chart用于部署。

#### 1. 准备values文件

创建 `values-rbac.yaml`：

```yaml
rbac:
  enabled: true
  image:
    repository: davidliyutong/idekube-rbac
    tag: latest
    pullPolicy: IfNotPresent
  
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
  
  replicas: 2
  
  service:
    type: ClusterIP
    port: 8080
  
  env:
    LOG_LEVEL: info
    WORKER_THREADS: "1"
    HTTP_PORT: "8080"

postgresql:
  enabled: true
  auth:
    username: postgres
    password: password
    database: idekube_rbac
  primary:
    persistence:
      enabled: true
      size: 10Gi

rabbitmq:
  enabled: true
  auth:
    username: guest
    password: guest
  persistence:
    enabled: true
    size: 8Gi
```

#### 2. 安装Chart

```bash
# 从项目根目录
cd /path/to/idekube

# 安装
helm install idekube-rbac ./helm \
  -f values-rbac.yaml \
  -n idekube \
  --create-namespace
```

#### 3. 验证部署

```bash
# 检查Pod状态
kubectl get pods -n idekube

# 检查服务
kubectl get svc -n idekube

# 查看日志
kubectl logs -n idekube -l app=idekube-rbac -f
```

#### 4. 访问服务

```bash
# 端口转发
kubectl port-forward -n idekube svc/idekube-rbac 8080:8080

# 访问
curl http://localhost:8080/healthz
```

### 使用Manifest文件

如果不使用Helm，可以使用原始Kubernetes manifest：

#### 1. 创建Namespace

```bash
kubectl create namespace idekube
```

#### 2. 创建ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: rbac-config
  namespace: idekube
data:
  model.conf: |
    [request_definition]
    r = sub, obj, act

    [policy_definition]
    p = sub, obj, act

    [role_definition]
    g = _, _

    [policy_effect]
    e = some(where (p.eft == allow))

    [matchers]
    m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
  
  policy.csv: |
    p, role:admin, workspace, read
    p, role:admin, workspace, write
    p, role:admin, workspace, delete
```

```bash
kubectl apply -f rbac-configmap.yaml
```

#### 3. 创建Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: rbac-secrets
  namespace: idekube
type: Opaque
stringData:
  POSTGRES_PASSWORD: password
  RABBITMQ_PASSWORD: guest
```

```bash
kubectl apply -f rbac-secret.yaml
```

#### 4. 创建Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: idekube-rbac
  namespace: idekube
spec:
  replicas: 2
  selector:
    matchLabels:
      app: idekube-rbac
  template:
    metadata:
      labels:
        app: idekube-rbac
    spec:
      containers:
      - name: rbac
        image: davidliyutong/idekube-rbac:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: POSTGRES_HOST
          value: postgresql
        - name: POSTGRES_PORT
          value: "5432"
        - name: POSTGRES_USER
          value: postgres
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rbac-secrets
              key: POSTGRES_PASSWORD
        - name: POSTGRES_DB
          value: idekube_rbac
        - name: RABBITMQ_HOST
          value: rabbitmq
        - name: RABBITMQ_PORT
          value: "5672"
        - name: RABBITMQ_USER
          value: guest
        - name: RABBITMQ_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rbac-secrets
              key: RABBITMQ_PASSWORD
        - name: HTTP_PORT
          value: "8080"
        - name: LOG_LEVEL
          value: info
        - name: CASBIN_MODEL_PATH
          value: /etc/rbac/model.conf
        - name: CASBIN_POLICY_PATH
          value: /etc/rbac/policy.csv
        volumeMounts:
        - name: config
          mountPath: /etc/rbac
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: rbac-config
```

```bash
kubectl apply -f rbac-deployment.yaml
```

#### 5. 创建Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: idekube-rbac
  namespace: idekube
spec:
  selector:
    app: idekube-rbac
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
    name: http
  type: ClusterIP
```

```bash
kubectl apply -f rbac-service.yaml
```

## 配置说明

### 环境变量

| 变量名 | 必需 | 默认值 | 描述 |
|--------|------|--------|------|
| POSTGRES_HOST | 是 | localhost | PostgreSQL主机地址 |
| POSTGRES_PORT | 是 | 5432 | PostgreSQL端口 |
| POSTGRES_USER | 是 | - | PostgreSQL用户名 |
| POSTGRES_PASSWORD | 是 | - | PostgreSQL密码 |
| POSTGRES_DB | 是 | - | PostgreSQL数据库名 |
| RABBITMQ_HOST | 是 | localhost | RabbitMQ主机地址 |
| RABBITMQ_PORT | 是 | 5672 | RabbitMQ端口 |
| RABBITMQ_USER | 是 | - | RabbitMQ用户名 |
| RABBITMQ_PASSWORD | 是 | - | RabbitMQ密码 |
| RABBITMQ_VHOST | 否 | / | RabbitMQ虚拟主机 |
| HTTP_PORT | 否 | 8080 | HTTP服务端口 |
| LOG_LEVEL | 否 | info | 日志级别 (debug/info/warn/error) |
| WORKER_THREADS | 否 | 1 | 工作线程数 |
| CASBIN_MODEL_PATH | 是 | - | Casbin模型配置文件路径 |
| CASBIN_POLICY_PATH | 否 | - | Casbin策略文件路径（可选） |

### Casbin配置

#### 模型文件 (model.conf)

定义RBAC规则的结构。推荐使用上面提供的默认配置。

#### 策略文件 (policy.csv)

定义具体的权限策略。格式：

```csv
# 策略规则: p, subject, object, action
p, role:admin, workspace, read
p, role:admin, workspace, write

# 角色继承: g, user, role
g, user:1, role:admin
g, user:2, role:editor
```

**注意：** 策略也可以存储在PostgreSQL数据库中，这样可以动态管理而无需重启服务。

## 验证部署

### 1. 健康检查

```bash
curl http://localhost:8080/healthz
```

期望输出：`ok`

### 2. 测试权限检查

```bash
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "resource_type": "workspace",
    "action": "read"
  }'
```

期望输出：`{"allowed":true}` 或 `{"allowed":false}`

### 3. 访问Swagger UI

在浏览器中打开：`http://localhost:8080/swagger/`

### 4. 检查日志

#### Docker
```bash
docker logs idekube-rbac
```

#### Kubernetes
```bash
kubectl logs -n idekube -l app=idekube-rbac
```

## 故障排查

### 服务无法启动

**症状：** 服务启动失败或立即退出

**检查清单：**

1. 验证PostgreSQL连接
   ```bash
   psql -h localhost -U postgres -d idekube_rbac
   ```

2. 验证RabbitMQ连接
   ```bash
   curl http://localhost:15672/api/overview
   ```

3. 检查环境变量是否正确设置

4. 查看详细日志（设置 `LOG_LEVEL=debug`）

### 数据库连接失败

**错误：** `Failed to connect to PostgreSQL`

**解决方案：**

1. 确保PostgreSQL正在运行
   ```bash
   docker ps | grep postgres
   ```

2. 检查连接字符串
   ```bash
   psql "postgresql://postgres:password@localhost:5432/idekube_rbac"
   ```

3. 验证网络连通性
   ```bash
   telnet localhost 5432
   ```

### RabbitMQ连接失败

**错误：** `Failed to connect to RabbitMQ`

**解决方案：**

1. 确保RabbitMQ正在运行
   ```bash
   docker ps | grep rabbitmq
   ```

2. 检查管理界面
   ```bash
   curl http://localhost:15672
   ```

3. 验证凭据
   ```bash
   rabbitmqctl authenticate_user guest guest
   ```

### Casbin初始化失败

**错误：** `Failed to initialize RBAC service`

**解决方案：**

1. 检查模型文件是否存在
   ```bash
   ls -la configs/model.conf
   ```

2. 验证模型文件语法

3. 检查文件权限

### Kubernetes Pod一直处于CrashLoopBackOff

**解决方案：**

1. 查看Pod详情
   ```bash
   kubectl describe pod -n idekube <pod-name>
   ```

2. 查看日志
   ```bash
   kubectl logs -n idekube <pod-name> --previous
   ```

3. 检查ConfigMap和Secret是否正确挂载

4. 验证依赖服务（PostgreSQL, RabbitMQ）是否就绪

### API响应缓慢

**可能原因：**

1. 数据库连接池不足
2. 数据库缺少索引
3. 策略规则过多

**优化建议：**

1. 增加数据库连接池大小
2. 为Casbin策略表添加索引
3. 启用结果缓存
4. 增加副本数（Kubernetes）

## 生产环境建议

### 1. 安全

- 使用强密码
- 启用TLS/HTTPS
- 配置网络策略
- 定期更新依赖

### 2. 高可用

- 部署多个副本（至少2个）
- 配置Pod反亲和性
- 使用HPA自动扩缩容
- 配置健康检查

### 3. 监控

- 配置Prometheus监控
- 设置告警规则
- 收集日志到中心化日志系统
- 监控关键指标（延迟、错误率、吞吐量）

### 4. 备份

- 定期备份PostgreSQL数据库
- 备份Casbin配置文件
- 测试恢复流程

## 下一步

- 阅读 [API定义文档](API.md) 了解API详情
- 查看 [API测试指南](API_TESTING.md) 学习如何测试
- 浏览 [README](README.md) 了解项目概述

## 获取帮助

- 查看项目Issues
- 阅读项目Wiki
- 联系开发团队

---

**祝您使用愉快！**
