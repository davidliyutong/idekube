# API 测试指南

本文档描述如何测试 idekube-housekeeper 组件的功能。

## 概述

idekube-housekeeper 是一个后台服务，负责清理和维护 idekube 平台中的 Kubernetes 资源。它通过监听 RabbitMQ 消息队列来接收清理任务，并与 PostgreSQL 数据库和 Kubernetes API 进行交互。

## 测试环境准备

### 前置条件

- Kubernetes 集群（本地或远程）
- PostgreSQL 数据库
- RabbitMQ 消息队列
- Go 1.21+（用于本地测试）
- kubectl 命令行工具

### 环境变量配置

在测试之前，需要配置以下环境变量：

```bash
# Kubernetes 配置
export KUBECONFIG=~/.kube/config
export NAMESPACE=idekube-system

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

### 启动依赖服务

#### PostgreSQL

使用 Docker 启动 PostgreSQL：

```bash
docker run -d \
  --name postgres-test \
  -e POSTGRES_USER=idekube \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=idekube \
  -p 5432:5432 \
  postgres:15-alpine
```

#### RabbitMQ

使用 Docker 启动 RabbitMQ：

```bash
docker run -d \
  --name rabbitmq-test \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management-alpine
```

访问 RabbitMQ 管理界面：http://localhost:15672（用户名：guest，密码：guest）

## 本地测试

### 1. 构建应用

```bash
make build
```

### 2. 运行应用

```bash
make run
```

或直接运行编译后的二进制文件：

```bash
./bin/idekube-housekeeper
```

### 3. 查看日志

应用启动后会输出日志：

```
INFO  Starting idekube-housekeeper
INFO  Connecting to PostgreSQL at localhost:5432
INFO  Connecting to RabbitMQ at localhost:5672
INFO  Housekeeper started
DEBUG Housekeeper heartbeat
```

## 功能测试

### 1. 连接测试

验证 housekeeper 是否能够成功连接到所有依赖服务：

#### PostgreSQL 连接测试

检查日志中是否有成功连接的消息：

```
INFO  Connected to PostgreSQL successfully
```

可以通过以下 SQL 查询验证连接：

```bash
docker exec -it postgres-test psql -U idekube -d idekube -c "SELECT version();"
```

#### RabbitMQ 连接测试

检查 RabbitMQ 管理界面的连接列表，确认有来自 housekeeper 的连接。

或使用 RabbitMQ 命令行工具：

```bash
docker exec rabbitmq-test rabbitmqctl list_connections
```

#### Kubernetes 连接测试

验证 housekeeper 能否访问 Kubernetes API：

```bash
kubectl get pods -n idekube-system
```

### 2. 消息队列测试

#### 发送测试消息

通过 RabbitMQ 管理界面或命令行工具发送测试消息。

使用 Python 脚本发送测试消息：

```python
#!/usr/bin/env python3
import pika
import json

# 连接到 RabbitMQ
connection = pika.BlockingConnection(
    pika.ConnectionParameters('localhost')
)
channel = connection.channel()

# 声明队列
channel.queue_declare(queue='housekeeper.cleanup', durable=True)

# 构造测试消息
message = {
    "action": "cleanup",
    "resource_type": "workspace",
    "resource_id": "test-workspace-123",
    "timestamp": "2026-01-13T10:00:00Z"
}

# 发送消息
channel.basic_publish(
    exchange='',
    routing_key='housekeeper.cleanup',
    body=json.dumps(message),
    properties=pika.BasicProperties(
        delivery_mode=2,  # 持久化消息
    )
)

print(f"Sent message: {message}")
connection.close()
```

#### 验证消息处理

检查 housekeeper 日志，确认消息已被接收和处理：

```
DEBUG Received cleanup request for workspace: test-workspace-123
INFO  Processing cleanup task
DEBUG Cleanup completed successfully
```

### 3. 清理任务测试

#### 创建测试资源

首先在 Kubernetes 中创建一些测试资源：

```bash
kubectl create namespace test-workspace-123
kubectl create deployment test-app --image=nginx -n test-workspace-123
kubectl create service clusterip test-service --tcp=80:80 -n test-workspace-123
```

#### 触发清理任务

发送清理消息（使用上面的 Python 脚本）

#### 验证清理结果

检查资源是否已被删除：

```bash
kubectl get namespace test-workspace-123
kubectl get deployment test-app -n test-workspace-123
kubectl get service test-service -n test-workspace-123
```

预期结果：资源应该已被删除或标记为删除。

### 4. 数据库操作测试

#### 查询工作空间状态

```sql
SELECT id, name, current_status, target_status, accessed_at, deleted_at
FROM workspaces
WHERE name = 'test-workspace-123';
```

#### 验证状态更新

清理任务完成后，工作空间状态应该更新：

- `deleted_at` 字段应该有值
- `current_status` 应该变为 `stopped` 或类似状态

### 5. 定期清理测试

housekeeper 默认每 10 秒执行一次心跳检查。可以通过修改代码中的 `ticker` 时间间隔来调整频率。

验证定期清理：

1. 创建一些过期的工作空间
2. 等待 housekeeper 自动清理
3. 检查日志和数据库状态

## 集成测试

### Docker Compose 测试

使用 Docker Compose 启动完整的测试环境：

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: idekube
      POSTGRES_PASSWORD: test_password
      POSTGRES_DB: idekube
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

  housekeeper:
    build: .
    depends_on:
      - postgres
      - rabbitmq
    environment:
      POSTGRES_HOST: postgres
      POSTGRES_PORT: 5432
      POSTGRES_USER: idekube
      POSTGRES_PASSWORD: test_password
      POSTGRES_DB: idekube
      RABBITMQ_HOST: rabbitmq
      RABBITMQ_PORT: 5672
      RABBITMQ_USER: guest
      RABBITMQ_PASSWORD: guest
      RABBITMQ_VHOST: /
      LOG_LEVEL: debug

volumes:
  postgres_data:
  rabbitmq_data:
```

启动测试环境：

```bash
docker-compose up -d
docker-compose logs -f housekeeper
```

## 性能测试

### 并发测试

测试 housekeeper 处理多个并发清理请求的能力：

```python
#!/usr/bin/env python3
import pika
import json
import time
from concurrent.futures import ThreadPoolExecutor

def send_cleanup_message(workspace_id):
    connection = pika.BlockingConnection(
        pika.ConnectionParameters('localhost')
    )
    channel = connection.channel()
    channel.queue_declare(queue='housekeeper.cleanup', durable=True)
    
    message = {
        "action": "cleanup",
        "resource_type": "workspace",
        "resource_id": f"test-workspace-{workspace_id}",
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    }
    
    channel.basic_publish(
        exchange='',
        routing_key='housekeeper.cleanup',
        body=json.dumps(message),
        properties=pika.BasicProperties(delivery_mode=2)
    )
    
    connection.close()
    print(f"Sent message for workspace {workspace_id}")

# 并发发送 100 个清理请求
with ThreadPoolExecutor(max_workers=10) as executor:
    executor.map(send_cleanup_message, range(100))
```

监控系统资源使用情况：

```bash
docker stats housekeeper
```

### 压力测试

测试 housekeeper 在高负载下的表现：

1. 增加 `WORKER_THREADS` 环境变量值
2. 发送大量清理请求
3. 监控响应时间和错误率
4. 检查数据库连接池和 RabbitMQ 连接状态

## 故障恢复测试

### 1. 数据库故障测试

停止 PostgreSQL：

```bash
docker stop postgres-test
```

观察 housekeeper 的行为和日志输出。

重启 PostgreSQL：

```bash
docker start postgres-test
```

验证 housekeeper 是否能够自动重连。

### 2. 消息队列故障测试

停止 RabbitMQ：

```bash
docker stop rabbitmq-test
```

发送一些清理请求（应该失败）。

重启 RabbitMQ：

```bash
docker start rabbitmq-test
```

验证消息是否能够被正确处理。

### 3. Kubernetes API 故障测试

模拟 Kubernetes API 不可用的情况，验证 housekeeper 的错误处理和重试机制。

## 日志分析

### 关键日志示例

成功启动：
```
INFO  Starting idekube-housekeeper
INFO  Configuration loaded successfully
INFO  Connected to PostgreSQL
INFO  Connected to RabbitMQ
INFO  Kubernetes client initialized
INFO  Housekeeper started
```

处理清理任务：
```
DEBUG Received cleanup message: {action: cleanup, resource_id: test-workspace-123}
INFO  Starting cleanup for workspace: test-workspace-123
DEBUG Deleting Kubernetes resources in namespace: test-workspace-123
DEBUG Updating database records
INFO  Cleanup completed successfully for: test-workspace-123
```

错误处理：
```
ERROR Failed to delete Kubernetes resources: namespace not found
WARN  Retrying cleanup task in 30 seconds
```

## 常见问题排查

### 问题 1: 无法连接到 PostgreSQL

**症状**：
```
FATAL Failed to connect to PostgreSQL: connection refused
```

**解决方案**：
1. 检查 PostgreSQL 是否正在运行
2. 验证环境变量配置
3. 检查网络连接和防火墙规则

### 问题 2: RabbitMQ 消息未被处理

**症状**：消息发送到队列但没有日志输出

**解决方案**：
1. 检查队列名称是否正确
2. 验证 housekeeper 是否正在监听队列
3. 检查消息格式是否符合要求

### 问题 3: Kubernetes 资源删除失败

**症状**：
```
ERROR Failed to delete namespace: forbidden
```

**解决方案**：
1. 检查 ServiceAccount 权限
2. 验证 RBAC 配置
3. 确认 kubeconfig 文件有效

## 最佳实践

1. **使用独立的测试环境**：避免在生产环境中进行测试
2. **保留测试日志**：便于问题排查和性能分析
3. **自动化测试**：使用脚本自动化常规测试流程
4. **监控资源使用**：测试时监控 CPU、内存和网络使用情况
5. **模拟真实场景**：测试数据和场景应尽可能接近生产环境

## 参考资料

- [API 定义文档](./API.md) - RabbitMQ 消息格式详细说明
- [快速开始指南](./QUICKSTART.md) - 部署和配置指南
- [主 README](./README.md) - 项目概述和架构说明
