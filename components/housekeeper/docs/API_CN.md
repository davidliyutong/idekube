# API 定义文档

本文档详细说明 idekube-housekeeper 组件使用的 RabbitMQ 消息格式、队列定义和数据库模型。

## 概述

idekube-housekeeper 通过 RabbitMQ 消息队列接收清理和维护任务，负责清理 Kubernetes 资源、归档数据和执行定期维护任务。

## RabbitMQ 配置

### 连接参数

| 参数 | 环境变量 | 默认值 | 说明 |
|------|---------|--------|------|
| 主机 | `RABBITMQ_HOST` | localhost | RabbitMQ 服务器地址 |
| 端口 | `RABBITMQ_PORT` | 5672 | RabbitMQ 服务端口 |
| 用户名 | `RABBITMQ_USER` | - | RabbitMQ 用户名（必填） |
| 密码 | `RABBITMQ_PASSWORD` | - | RabbitMQ 密码（必填） |
| 虚拟主机 | `RABBITMQ_VHOST` | / | RabbitMQ 虚拟主机 |

### 连接字符串格式

```
amqp://<user>:<password>@<host>:<port><vhost>
```

示例：
```
amqp://guest:guest@localhost:5672/
amqp://idekube:secret@rabbitmq.example.com:5672/idekube
```

## 队列定义

### 1. housekeeper.cleanup

**用途**：接收工作空间清理请求

**队列属性**：
- **持久化 (Durable)**：true
- **独占 (Exclusive)**：false
- **自动删除 (Auto-delete)**：false
- **消息 TTL**：86400000 (24小时)
- **最大长度**：10000

**消息格式**：

```json
{
  "action": "cleanup",
  "resource_type": "workspace",
  "resource_id": "workspace-uuid-or-name",
  "owner_id": 123,
  "owner_type": "user",
  "force": false,
  "timestamp": "2026-01-13T10:00:00Z",
  "metadata": {
    "reason": "user_deleted",
    "requested_by": "user@example.com"
  }
}
```

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | 操作类型，固定为 "cleanup" |
| resource_type | string | 是 | 资源类型，目前支持 "workspace" |
| resource_id | string | 是 | 资源标识符（UUID 或名称） |
| owner_id | integer | 否 | 资源所有者 ID |
| owner_type | string | 否 | 所有者类型："user" 或 "organization" |
| force | boolean | 否 | 是否强制清理（默认：false） |
| timestamp | string | 是 | ISO 8601 格式的时间戳 |
| metadata | object | 否 | 附加元数据 |

**响应**：无直接响应，通过日志和数据库状态更新反馈结果

---

### 2. housekeeper.archive

**用途**：接收数据归档请求

**队列属性**：
- **持久化 (Durable)**：true
- **独占 (Exclusive)**：false
- **自动删除 (Auto-delete)**：false
- **消息 TTL**：604800000 (7天)

**消息格式**：

```json
{
  "action": "archive",
  "resource_type": "workspace",
  "resource_id": "workspace-uuid-or-name",
  "archive_location": "s3://backups/workspace-123",
  "retention_days": 90,
  "timestamp": "2026-01-13T10:00:00Z",
  "metadata": {
    "size_mb": 1024,
    "file_count": 150
  }
}
```

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | 操作类型，固定为 "archive" |
| resource_type | string | 是 | 资源类型 |
| resource_id | string | 是 | 资源标识符 |
| archive_location | string | 是 | 归档存储位置（S3/NFS/本地路径） |
| retention_days | integer | 否 | 保留天数（默认：90） |
| timestamp | string | 是 | ISO 8601 格式的时间戳 |
| metadata | object | 否 | 归档元数据 |

---

### 3. housekeeper.maintenance

**用途**：接收定期维护任务

**队列属性**：
- **持久化 (Durable)**：true
- **独占 (Exclusive)**：false
- **自动删除 (Auto-delete)**：false

**消息格式**：

```json
{
  "action": "maintenance",
  "task_type": "cleanup_expired",
  "parameters": {
    "max_age_hours": 24,
    "batch_size": 100
  },
  "scheduled_at": "2026-01-13T02:00:00Z",
  "timestamp": "2026-01-13T02:00:00Z"
}
```

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | 操作类型，固定为 "maintenance" |
| task_type | string | 是 | 维护任务类型 |
| parameters | object | 是 | 任务参数（因任务类型而异） |
| scheduled_at | string | 否 | 计划执行时间 |
| timestamp | string | 是 | 消息创建时间戳 |

**支持的维护任务类型**：

#### cleanup_expired
清理过期的工作空间

参数：
```json
{
  "max_age_hours": 24,
  "batch_size": 100,
  "dry_run": false
}
```

#### optimize_database
优化数据库（VACUUM、ANALYZE等）

参数：
```json
{
  "tables": ["workspaces", "workspace_logs"],
  "analyze_only": false
}
```

#### prune_images
清理未使用的容器镜像

参数：
```json
{
  "age_days": 30,
  "keep_latest": 5
}
```

## 数据库模型

### Workspace 表

工作空间主表，存储工作空间的核心信息。

**表名**：`workspaces`

**字段定义**：

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | bigint | PRIMARY KEY, AUTO_INCREMENT | 主键 ID |
| uuid | uuid | UNIQUE, NOT NULL | 全局唯一标识符 |
| name | varchar(255) | UNIQUE, NOT NULL | 工作空间名称 |
| display_name | varchar(255) | NULLABLE | 显示名称 |
| description | text | NULLABLE | 工作空间描述 |
| owner_type | varchar(50) | NOT NULL | 所有者类型 |
| owner_id | bigint | NOT NULL, INDEX | 所有者 ID |
| template_id | bigint | NOT NULL | 模板 ID |
| cpu_millicores | bigint | NOT NULL | CPU 配额（毫核） |
| memory_mb | bigint | NOT NULL | 内存配额（MB） |
| storage_mb | bigint | NOT NULL | 存储配额（MB） |
| current_status | varchar(50) | NOT NULL, INDEX | 当前状态 |
| target_status | varchar(50) | NOT NULL | 目标状态 |
| k8s_namespace | varchar(255) | NULLABLE | Kubernetes 命名空间 |
| k8s_deployment_name | varchar(255) | NULLABLE | Kubernetes Deployment 名称 |
| k8s_service_name | varchar(255) | NULLABLE | Kubernetes Service 名称 |
| timeout_seconds | bigint | DEFAULT 3600 | 超时时间（秒） |
| accessed_at | timestamp with time zone | NULLABLE | 最后访问时间 |
| created_at | timestamp with time zone | NOT NULL, DEFAULT now() | 创建时间 |
| updated_at | timestamp with time zone | NOT NULL, DEFAULT now() | 更新时间 |
| deleted_at | timestamp with time zone | NULLABLE, INDEX | 软删除时间 |

**工作空间状态枚举**：

| 状态 | 说明 |
|------|------|
| pending | 待创建 |
| starting | 启动中 |
| running | 运行中 |
| stopped | 已停止 |
| failed | 失败 |

**所有者类型枚举**：

| 类型 | 说明 |
|------|------|
| user | 用户 |
| organization | 组织 |

### 索引定义

```sql
-- 主键索引
CREATE UNIQUE INDEX idx_workspaces_pkey ON workspaces(id);

-- UUID 唯一索引
CREATE UNIQUE INDEX idx_workspaces_uuid ON workspaces(uuid);

-- 名称唯一索引
CREATE UNIQUE INDEX idx_workspaces_name ON workspaces(name);

-- 所有者索引
CREATE INDEX idx_workspaces_owner ON workspaces(owner_type, owner_id);

-- 状态索引
CREATE INDEX idx_workspaces_status ON workspaces(current_status);

-- 软删除索引
CREATE INDEX idx_workspaces_deleted_at ON workspaces(deleted_at);

-- 访问时间索引（用于清理过期工作空间）
CREATE INDEX idx_workspaces_accessed_at ON workspaces(accessed_at) WHERE deleted_at IS NULL;
```

## 消息处理流程

### 清理流程

```
1. 接收清理消息
   ↓
2. 验证消息格式
   ↓
3. 查询数据库获取工作空间信息
   ↓
4. 删除 Kubernetes 资源
   - 删除 Deployment
   - 删除 Service
   - 删除 PVC（可选）
   - 删除 Namespace（可选）
   ↓
5. 更新数据库状态
   - 设置 deleted_at 时间戳
   - 更新 current_status 为 stopped
   ↓
6. 记录清理日志
   ↓
7. 发送确认消息（ACK）
```

### 错误处理

当消息处理失败时：

1. **重试机制**：
   - 自动重试 3 次
   - 每次重试间隔：5 秒、30 秒、2 分钟
   - 使用指数退避策略

2. **死信队列**：
   - 队列名：`housekeeper.dlq`
   - 超过最大重试次数的消息会被发送到死信队列
   - 需要人工介入处理

3. **错误日志**：
   - 记录详细的错误信息
   - 包含消息内容、错误原因、堆栈跟踪
   - 按错误级别分类：ERROR、WARN

## 消息发送示例

### Python 示例

```python
#!/usr/bin/env python3
import pika
import json
from datetime import datetime, timezone

def send_cleanup_message(workspace_id):
    # 连接到 RabbitMQ
    credentials = pika.PlainCredentials('guest', 'guest')
    connection = pika.BlockingConnection(
        pika.ConnectionParameters(
            host='localhost',
            port=5672,
            virtual_host='/',
            credentials=credentials
        )
    )
    channel = connection.channel()
    
    # 声明队列
    channel.queue_declare(
        queue='housekeeper.cleanup',
        durable=True,
        arguments={
            'x-message-ttl': 86400000,  # 24小时
            'x-max-length': 10000
        }
    )
    
    # 构造消息
    message = {
        "action": "cleanup",
        "resource_type": "workspace",
        "resource_id": workspace_id,
        "force": False,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "metadata": {
            "reason": "user_deleted",
            "requested_by": "admin@example.com"
        }
    }
    
    # 发送消息
    channel.basic_publish(
        exchange='',
        routing_key='housekeeper.cleanup',
        body=json.dumps(message),
        properties=pika.BasicProperties(
            delivery_mode=2,  # 持久化消息
            content_type='application/json',
            timestamp=int(datetime.now(timezone.utc).timestamp())
        )
    )
    
    print(f"✓ Sent cleanup message for workspace: {workspace_id}")
    connection.close()

if __name__ == '__main__':
    send_cleanup_message('workspace-123')
```

### Go 示例

```go
package main

import (
    "encoding/json"
    "fmt"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

type CleanupMessage struct {
    Action       string                 `json:"action"`
    ResourceType string                 `json:"resource_type"`
    ResourceID   string                 `json:"resource_id"`
    Force        bool                   `json:"force"`
    Timestamp    time.Time              `json:"timestamp"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

func sendCleanupMessage(workspaceID string) error {
    // 连接到 RabbitMQ
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    defer conn.Close()

    channel, err := conn.Channel()
    if err != nil {
        return fmt.Errorf("failed to open channel: %w", err)
    }
    defer channel.Close()

    // 声明队列
    _, err = channel.QueueDeclare(
        "housekeeper.cleanup", // 队列名
        true,                  // 持久化
        false,                 // 不自动删除
        false,                 // 非独占
        false,                 // 不等待
        amqp.Table{
            "x-message-ttl": int32(86400000),
            "x-max-length":  int32(10000),
        },
    )
    if err != nil {
        return fmt.Errorf("failed to declare queue: %w", err)
    }

    // 构造消息
    msg := CleanupMessage{
        Action:       "cleanup",
        ResourceType: "workspace",
        ResourceID:   workspaceID,
        Force:        false,
        Timestamp:    time.Now().UTC(),
        Metadata: map[string]interface{}{
            "reason":       "user_deleted",
            "requested_by": "admin@example.com",
        },
    }

    body, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    // 发送消息
    err = channel.Publish(
        "",                    // exchange
        "housekeeper.cleanup", // routing key
        false,                 // mandatory
        false,                 // immediate
        amqp.Publishing{
            DeliveryMode: amqp.Persistent,
            ContentType:  "application/json",
            Body:         body,
            Timestamp:    time.Now().UTC(),
        },
    )
    if err != nil {
        return fmt.Errorf("failed to publish message: %w", err)
    }

    fmt.Printf("✓ Sent cleanup message for workspace: %s\n", workspaceID)
    return nil
}

func main() {
    if err := sendCleanupMessage("workspace-123"); err != nil {
        panic(err)
    }
}
```

### cURL + RabbitMQ HTTP API

```bash
# 发送清理消息
curl -i -u guest:guest -H "content-type:application/json" \
  -X POST http://localhost:15672/api/exchanges/%2F/amq.default/publish \
  -d '{
    "properties": {
      "delivery_mode": 2,
      "content_type": "application/json"
    },
    "routing_key": "housekeeper.cleanup",
    "payload": "{\"action\":\"cleanup\",\"resource_type\":\"workspace\",\"resource_id\":\"workspace-123\",\"timestamp\":\"2026-01-13T10:00:00Z\"}",
    "payload_encoding": "string"
  }'
```

## 监控和指标

### 推荐监控指标

1. **队列指标**：
   - 队列长度
   - 消息处理速率
   - 消息重试次数
   - 死信队列消息数

2. **处理指标**：
   - 清理任务成功率
   - 平均处理时间
   - 失败任务数量
   - 并发处理数

3. **资源指标**：
   - CPU 使用率
   - 内存使用率
   - 数据库连接数
   - RabbitMQ 连接数

### RabbitMQ 管理命令

```bash
# 查看队列状态
rabbitmqctl list_queues name messages consumers

# 查看队列详细信息
rabbitmqctl list_queues name messages_ready messages_unacknowledged

# 查看连接
rabbitmqctl list_connections

# 清空队列（慎用）
rabbitmqctl purge_queue housekeeper.cleanup
```

## 最佳实践

1. **消息幂等性**：确保重复处理相同消息不会产生副作用
2. **消息顺序**：不保证严格顺序，设计时应考虑无序处理
3. **超时设置**：合理设置消息 TTL 和任务超时时间
4. **批量处理**：对于大量清理任务，考虑批量处理以提高效率
5. **监控告警**：设置队列长度和处理延迟告警
6. **死信处理**：定期检查和处理死信队列中的消息
7. **版本兼容**：消息格式变更时保持向后兼容

## 参考资料

- [API 测试指南](./API_TESTING.md) - 如何测试 API 和消息队列
- [快速开始指南](./QUICKSTART.md) - 部署和配置指南
- [RabbitMQ 官方文档](https://www.rabbitmq.com/documentation.html)
- [AMQP 0-9-1 协议规范](https://www.rabbitmq.com/amqp-0-9-1-reference.html)
