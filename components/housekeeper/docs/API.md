# API Definition

This document provides detailed specifications for RabbitMQ message formats, queue definitions, and database models used by the idekube-housekeeper component.

## Overview

idekube-housekeeper receives cleanup and maintenance tasks through RabbitMQ message queues, responsible for cleaning Kubernetes resources, archiving data, and executing periodic maintenance tasks.

## RabbitMQ Configuration

### Connection Parameters

| Parameter | Environment Variable | Default | Description |
|-----------|---------------------|---------|-------------|
| Host | `RABBITMQ_HOST` | localhost | RabbitMQ server address |
| Port | `RABBITMQ_PORT` | 5672 | RabbitMQ service port |
| Username | `RABBITMQ_USER` | - | RabbitMQ username (required) |
| Password | `RABBITMQ_PASSWORD` | - | RabbitMQ password (required) |
| Virtual Host | `RABBITMQ_VHOST` | / | RabbitMQ virtual host |

### Connection String Format

```
amqp://<user>:<password>@<host>:<port><vhost>
```

Examples:
```
amqp://guest:guest@localhost:5672/
amqp://idekube:secret@rabbitmq.example.com:5672/idekube
```

## Queue Definitions

### 1. housekeeper.cleanup

**Purpose**: Receive workspace cleanup requests

**Queue Properties**:
- **Durable**: true
- **Exclusive**: false
- **Auto-delete**: false
- **Message TTL**: 86400000 (24 hours)
- **Max Length**: 10000

**Message Format**:

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

**Field Descriptions**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| action | string | Yes | Operation type, fixed as "cleanup" |
| resource_type | string | Yes | Resource type, currently supports "workspace" |
| resource_id | string | Yes | Resource identifier (UUID or name) |
| owner_id | integer | No | Resource owner ID |
| owner_type | string | No | Owner type: "user" or "organization" |
| force | boolean | No | Whether to force cleanup (default: false) |
| timestamp | string | Yes | ISO 8601 formatted timestamp |
| metadata | object | No | Additional metadata |

**Response**: No direct response, feedback is provided through logs and database status updates

---

### 2. housekeeper.archive

**Purpose**: Receive data archiving requests

**Queue Properties**:
- **Durable**: true
- **Exclusive**: false
- **Auto-delete**: false
- **Message TTL**: 604800000 (7 days)

**Message Format**:

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

**Field Descriptions**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| action | string | Yes | Operation type, fixed as "archive" |
| resource_type | string | Yes | Resource type |
| resource_id | string | Yes | Resource identifier |
| archive_location | string | Yes | Archive storage location (S3/NFS/local path) |
| retention_days | integer | No | Retention days (default: 90) |
| timestamp | string | Yes | ISO 8601 formatted timestamp |
| metadata | object | No | Archive metadata |

---

### 3. housekeeper.maintenance

**Purpose**: Receive periodic maintenance tasks

**Queue Properties**:
- **Durable**: true
- **Exclusive**: false
- **Auto-delete**: false

**Message Format**:

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

**Field Descriptions**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| action | string | Yes | Operation type, fixed as "maintenance" |
| task_type | string | Yes | Maintenance task type |
| parameters | object | Yes | Task parameters (varies by task type) |
| scheduled_at | string | No | Scheduled execution time |
| timestamp | string | Yes | Message creation timestamp |

**Supported Maintenance Task Types**:

#### cleanup_expired
Clean up expired workspaces

Parameters:
```json
{
  "max_age_hours": 24,
  "batch_size": 100,
  "dry_run": false
}
```

#### optimize_database
Optimize database (VACUUM, ANALYZE, etc.)

Parameters:
```json
{
  "tables": ["workspaces", "workspace_logs"],
  "analyze_only": false
}
```

#### prune_images
Clean up unused container images

Parameters:
```json
{
  "age_days": 30,
  "keep_latest": 5
}
```

## Database Models

### Workspace Table

Main workspace table storing core workspace information.

**Table Name**: `workspaces`

**Field Definitions**:

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | bigint | PRIMARY KEY, AUTO_INCREMENT | Primary key ID |
| uuid | uuid | UNIQUE, NOT NULL | Globally unique identifier |
| name | varchar(255) | UNIQUE, NOT NULL | Workspace name |
| display_name | varchar(255) | NULLABLE | Display name |
| description | text | NULLABLE | Workspace description |
| owner_type | varchar(50) | NOT NULL | Owner type |
| owner_id | bigint | NOT NULL, INDEX | Owner ID |
| template_id | bigint | NOT NULL | Template ID |
| cpu_millicores | bigint | NOT NULL | CPU quota (millicores) |
| memory_mb | bigint | NOT NULL | Memory quota (MB) |
| storage_mb | bigint | NOT NULL | Storage quota (MB) |
| current_status | varchar(50) | NOT NULL, INDEX | Current status |
| target_status | varchar(50) | NOT NULL | Target status |
| k8s_namespace | varchar(255) | NULLABLE | Kubernetes namespace |
| k8s_deployment_name | varchar(255) | NULLABLE | Kubernetes Deployment name |
| k8s_service_name | varchar(255) | NULLABLE | Kubernetes Service name |
| timeout_seconds | bigint | DEFAULT 3600 | Timeout in seconds |
| accessed_at | timestamp with time zone | NULLABLE | Last access time |
| created_at | timestamp with time zone | NOT NULL, DEFAULT now() | Creation time |
| updated_at | timestamp with time zone | NOT NULL, DEFAULT now() | Update time |
| deleted_at | timestamp with time zone | NULLABLE, INDEX | Soft delete time |

**Workspace Status Enums**:

| Status | Description |
|--------|-------------|
| pending | Pending creation |
| starting | Starting |
| running | Running |
| stopped | Stopped |
| failed | Failed |

**Owner Type Enums**:

| Type | Description |
|------|-------------|
| user | User |
| organization | Organization |

### Index Definitions

```sql
-- Primary key index
CREATE UNIQUE INDEX idx_workspaces_pkey ON workspaces(id);

-- UUID unique index
CREATE UNIQUE INDEX idx_workspaces_uuid ON workspaces(uuid);

-- Name unique index
CREATE UNIQUE INDEX idx_workspaces_name ON workspaces(name);

-- Owner index
CREATE INDEX idx_workspaces_owner ON workspaces(owner_type, owner_id);

-- Status index
CREATE INDEX idx_workspaces_status ON workspaces(current_status);

-- Soft delete index
CREATE INDEX idx_workspaces_deleted_at ON workspaces(deleted_at);

-- Access time index (for cleaning expired workspaces)
CREATE INDEX idx_workspaces_accessed_at ON workspaces(accessed_at) WHERE deleted_at IS NULL;
```

## Message Processing Flow

### Cleanup Flow

```
1. Receive cleanup message
   ↓
2. Validate message format
   ↓
3. Query database for workspace information
   ↓
4. Delete Kubernetes resources
   - Delete Deployment
   - Delete Service
   - Delete PVC (optional)
   - Delete Namespace (optional)
   ↓
5. Update database status
   - Set deleted_at timestamp
   - Update current_status to stopped
   ↓
6. Record cleanup logs
   ↓
7. Send acknowledgment (ACK)
```

### Error Handling

When message processing fails:

1. **Retry Mechanism**:
   - Automatically retry 3 times
   - Retry intervals: 5 seconds, 30 seconds, 2 minutes
   - Use exponential backoff strategy

2. **Dead Letter Queue**:
   - Queue name: `housekeeper.dlq`
   - Messages exceeding max retry count are sent to dead letter queue
   - Requires manual intervention

3. **Error Logging**:
   - Record detailed error information
   - Include message content, error reason, stack trace
   - Categorize by error level: ERROR, WARN

## Message Sending Examples

### Python Example

```python
#!/usr/bin/env python3
import pika
import json
from datetime import datetime, timezone

def send_cleanup_message(workspace_id):
    # Connect to RabbitMQ
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
    
    # Declare queue
    channel.queue_declare(
        queue='housekeeper.cleanup',
        durable=True,
        arguments={
            'x-message-ttl': 86400000,  # 24 hours
            'x-max-length': 10000
        }
    )
    
    # Construct message
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
    
    # Send message
    channel.basic_publish(
        exchange='',
        routing_key='housekeeper.cleanup',
        body=json.dumps(message),
        properties=pika.BasicProperties(
            delivery_mode=2,  # Persistent message
            content_type='application/json',
            timestamp=int(datetime.now(timezone.utc).timestamp())
        )
    )
    
    print(f"✓ Sent cleanup message for workspace: {workspace_id}")
    connection.close()

if __name__ == '__main__':
    send_cleanup_message('workspace-123')
```

### Go Example

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
    // Connect to RabbitMQ
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

    // Declare queue
    _, err = channel.QueueDeclare(
        "housekeeper.cleanup", // Queue name
        true,                  // Durable
        false,                 // Not auto-delete
        false,                 // Not exclusive
        false,                 // No wait
        amqp.Table{
            "x-message-ttl": int32(86400000),
            "x-max-length":  int32(10000),
        },
    )
    if err != nil {
        return fmt.Errorf("failed to declare queue: %w", err)
    }

    // Construct message
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

    // Send message
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
# Send cleanup message
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

## Monitoring and Metrics

### Recommended Monitoring Metrics

1. **Queue Metrics**:
   - Queue length
   - Message processing rate
   - Message retry count
   - Dead letter queue message count

2. **Processing Metrics**:
   - Cleanup task success rate
   - Average processing time
   - Failed task count
   - Concurrent processing count

3. **Resource Metrics**:
   - CPU usage
   - Memory usage
   - Database connection count
   - RabbitMQ connection count

### RabbitMQ Management Commands

```bash
# View queue status
rabbitmqctl list_queues name messages consumers

# View queue details
rabbitmqctl list_queues name messages_ready messages_unacknowledged

# View connections
rabbitmqctl list_connections

# Purge queue (use with caution)
rabbitmqctl purge_queue housekeeper.cleanup
```

## Best Practices

1. **Message Idempotency**: Ensure repeated processing of the same message doesn't produce side effects
2. **Message Ordering**: Strict ordering is not guaranteed, design should consider unordered processing
3. **Timeout Settings**: Set reasonable message TTL and task timeout values
4. **Batch Processing**: For large cleanup tasks, consider batch processing for efficiency
5. **Monitoring Alerts**: Set alerts for queue length and processing delays
6. **Dead Letter Handling**: Regularly check and process messages in dead letter queue
7. **Version Compatibility**: Maintain backward compatibility when changing message formats

## References

- [API Testing Guide](./API_TESTING.md) - How to test API and message queues
- [Quick Start Guide](./QUICKSTART.md) - Deployment and configuration guide
- [RabbitMQ Official Documentation](https://www.rabbitmq.com/documentation.html)
- [AMQP 0-9-1 Protocol Specification](https://www.rabbitmq.com/amqp-0-9-1-reference.html)
