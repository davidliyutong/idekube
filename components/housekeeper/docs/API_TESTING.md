# API Testing Guide

This document describes how to test the functionality of the idekube-housekeeper component.

## Overview

idekube-housekeeper is a background service responsible for cleaning and maintaining Kubernetes resources in the idekube platform. It receives cleanup tasks by listening to RabbitMQ message queues and interacts with PostgreSQL database and Kubernetes API.

## Test Environment Setup

### Prerequisites

- Kubernetes cluster (local or remote)
- PostgreSQL database
- RabbitMQ message queue
- Go 1.21+ (for local testing)
- kubectl command-line tool

### Environment Variables Configuration

Before testing, configure the following environment variables:

```bash
# Kubernetes Configuration
export KUBECONFIG=~/.kube/config
export NAMESPACE=idekube-system

# PostgreSQL Configuration
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=idekube
export POSTGRES_PASSWORD=your_password
export POSTGRES_DB=idekube

# RabbitMQ Configuration
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export RABBITMQ_USER=guest
export RABBITMQ_PASSWORD=guest
export RABBITMQ_VHOST=/

# Application Configuration
export LOG_LEVEL=debug
export WORKER_THREADS=1
```

### Start Dependent Services

#### PostgreSQL

Start PostgreSQL using Docker:

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

Start RabbitMQ using Docker:

```bash
docker run -d \
  --name rabbitmq-test \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management-alpine
```

Access RabbitMQ management interface: http://localhost:15672 (username: guest, password: guest)

## Local Testing

### 1. Build Application

```bash
make build
```

### 2. Run Application

```bash
make run
```

Or run the compiled binary directly:

```bash
./bin/idekube-housekeeper
```

### 3. View Logs

After the application starts, it will output logs:

```
INFO  Starting idekube-housekeeper
INFO  Connecting to PostgreSQL at localhost:5432
INFO  Connecting to RabbitMQ at localhost:5672
INFO  Housekeeper started
DEBUG Housekeeper heartbeat
```

## Functionality Testing

### 1. Connection Testing

Verify that housekeeper can successfully connect to all dependent services:

#### PostgreSQL Connection Test

Check logs for successful connection messages:

```
INFO  Connected to PostgreSQL successfully
```

Verify connection with the following SQL query:

```bash
docker exec -it postgres-test psql -U idekube -d idekube -c "SELECT version();"
```

#### RabbitMQ Connection Test

Check the connection list in RabbitMQ management interface to confirm connection from housekeeper.

Or use RabbitMQ command-line tool:

```bash
docker exec rabbitmq-test rabbitmqctl list_connections
```

#### Kubernetes Connection Test

Verify that housekeeper can access Kubernetes API:

```bash
kubectl get pods -n idekube-system
```

### 2. Message Queue Testing

#### Send Test Message

Send test messages through RabbitMQ management interface or command-line tools.

Using Python script to send test message:

```python
#!/usr/bin/env python3
import pika
import json

# Connect to RabbitMQ
connection = pika.BlockingConnection(
    pika.ConnectionParameters('localhost')
)
channel = connection.channel()

# Declare queue
channel.queue_declare(queue='housekeeper.cleanup', durable=True)

# Construct test message
message = {
    "action": "cleanup",
    "resource_type": "workspace",
    "resource_id": "test-workspace-123",
    "timestamp": "2026-01-13T10:00:00Z"
}

# Send message
channel.basic_publish(
    exchange='',
    routing_key='housekeeper.cleanup',
    body=json.dumps(message),
    properties=pika.BasicProperties(
        delivery_mode=2,  # Persistent message
    )
)

print(f"Sent message: {message}")
connection.close()
```

#### Verify Message Processing

Check housekeeper logs to confirm message has been received and processed:

```
DEBUG Received cleanup request for workspace: test-workspace-123
INFO  Processing cleanup task
DEBUG Cleanup completed successfully
```

### 3. Cleanup Task Testing

#### Create Test Resources

First create some test resources in Kubernetes:

```bash
kubectl create namespace test-workspace-123
kubectl create deployment test-app --image=nginx -n test-workspace-123
kubectl create service clusterip test-service --tcp=80:80 -n test-workspace-123
```

#### Trigger Cleanup Task

Send cleanup message (using the Python script above)

#### Verify Cleanup Results

Check if resources have been deleted:

```bash
kubectl get namespace test-workspace-123
kubectl get deployment test-app -n test-workspace-123
kubectl get service test-service -n test-workspace-123
```

Expected result: Resources should be deleted or marked for deletion.

### 4. Database Operation Testing

#### Query Workspace Status

```sql
SELECT id, name, current_status, target_status, accessed_at, deleted_at
FROM workspaces
WHERE name = 'test-workspace-123';
```

#### Verify Status Update

After cleanup task completes, workspace status should be updated:

- `deleted_at` field should have a value
- `current_status` should change to `stopped` or similar status

### 5. Periodic Cleanup Testing

Housekeeper executes a heartbeat check every 10 seconds by default. You can adjust the frequency by modifying the `ticker` interval in the code.

Verify periodic cleanup:

1. Create some expired workspaces
2. Wait for housekeeper to automatically clean up
3. Check logs and database status

## Integration Testing

### Docker Compose Testing

Start complete test environment using Docker Compose:

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

Start test environment:

```bash
docker-compose up -d
docker-compose logs -f housekeeper
```

## Performance Testing

### Concurrency Testing

Test housekeeper's ability to handle multiple concurrent cleanup requests:

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

# Send 100 cleanup requests concurrently
with ThreadPoolExecutor(max_workers=10) as executor:
    executor.map(send_cleanup_message, range(100))
```

Monitor system resource usage:

```bash
docker stats housekeeper
```

### Load Testing

Test housekeeper's performance under high load:

1. Increase `WORKER_THREADS` environment variable value
2. Send large number of cleanup requests
3. Monitor response time and error rate
4. Check database connection pool and RabbitMQ connection status

## Failure Recovery Testing

### 1. Database Failure Testing

Stop PostgreSQL:

```bash
docker stop postgres-test
```

Observe housekeeper's behavior and log output.

Restart PostgreSQL:

```bash
docker start postgres-test
```

Verify that housekeeper can automatically reconnect.

### 2. Message Queue Failure Testing

Stop RabbitMQ:

```bash
docker stop rabbitmq-test
```

Send some cleanup requests (should fail).

Restart RabbitMQ:

```bash
docker start rabbitmq-test
```

Verify that messages can be processed correctly.

### 3. Kubernetes API Failure Testing

Simulate Kubernetes API unavailability and verify housekeeper's error handling and retry mechanism.

## Log Analysis

### Key Log Examples

Successful startup:
```
INFO  Starting idekube-housekeeper
INFO  Configuration loaded successfully
INFO  Connected to PostgreSQL
INFO  Connected to RabbitMQ
INFO  Kubernetes client initialized
INFO  Housekeeper started
```

Processing cleanup task:
```
DEBUG Received cleanup message: {action: cleanup, resource_id: test-workspace-123}
INFO  Starting cleanup for workspace: test-workspace-123
DEBUG Deleting Kubernetes resources in namespace: test-workspace-123
DEBUG Updating database records
INFO  Cleanup completed successfully for: test-workspace-123
```

Error handling:
```
ERROR Failed to delete Kubernetes resources: namespace not found
WARN  Retrying cleanup task in 30 seconds
```

## Common Issues Troubleshooting

### Issue 1: Cannot Connect to PostgreSQL

**Symptoms**:
```
FATAL Failed to connect to PostgreSQL: connection refused
```

**Solutions**:
1. Check if PostgreSQL is running
2. Verify environment variable configuration
3. Check network connection and firewall rules

### Issue 2: RabbitMQ Messages Not Being Processed

**Symptoms**: Messages sent to queue but no log output

**Solutions**:
1. Check if queue name is correct
2. Verify that housekeeper is listening to the queue
3. Check if message format meets requirements

### Issue 3: Kubernetes Resource Deletion Failed

**Symptoms**:
```
ERROR Failed to delete namespace: forbidden
```

**Solutions**:
1. Check ServiceAccount permissions
2. Verify RBAC configuration
3. Confirm kubeconfig file is valid

## Best Practices

1. **Use Isolated Test Environment**: Avoid testing in production environment
2. **Preserve Test Logs**: Helpful for troubleshooting and performance analysis
3. **Automate Testing**: Use scripts to automate routine testing processes
4. **Monitor Resource Usage**: Monitor CPU, memory, and network usage during testing
5. **Simulate Real Scenarios**: Test data and scenarios should be as close to production as possible

## References

- [API Definition](./API.md) - Detailed RabbitMQ message format specification
- [Quick Start Guide](./QUICKSTART.md) - Deployment and configuration guide
- [Main README](./README.md) - Project overview and architecture description
