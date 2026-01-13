# API Testing Guide

This document describes how to test the idekube-rbac API.

## Table of Contents

- [Environment Setup](#environment-setup)
- [Health Check](#health-check)
- [Permission Check API](#permission-check-api)
- [Role Assignment API](#role-assignment-api)
- [Using Swagger UI](#using-swagger-ui)

## Environment Setup

### Start the Service

Before testing the API, ensure the service is running:

```bash
# Run in local development mode
make run

# Or build and run
make build
./bin/idekube-rbac
```

By default, the service will start on port 8080. Ensure the correct environment variables are set in the configuration.

### Required Environment Variables

```bash
# PostgreSQL configuration
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=password
export POSTGRES_DB=idekube_rbac

# RabbitMQ configuration
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export RABBITMQ_USER=guest
export RABBITMQ_PASSWORD=guest

# Application configuration
export HTTP_PORT=8080
export LOG_LEVEL=debug
```

## Health Check

Verify the service is running properly:

```bash
curl -X GET http://localhost:8080/healthz
```

**Expected Response:**
```
ok
```

## Permission Check API

### Endpoint
`POST /api/v1/rbac/check`

### Description
Check if a user has permission to perform a specific action on a specified resource.

### Request Example

```bash
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "resource_type": "workspace",
    "resource_id": "ws-001",
    "action": "read"
  }'
```

### Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| user_id | integer | Yes | User ID |
| resource_type | string | Yes | Resource type (e.g., workspace, template, volume) |
| resource_id | string | No | Resource ID (optional, for specific resource permission check) |
| action | string | Yes | Action type (e.g., read, write, delete, execute) |

### Response Examples

**Success (200 OK):**
```json
{
  "allowed": true
}
```

or

```json
{
  "allowed": false
}
```

**Error (400 Bad Request):**
```
invalid request body: <error message>
```

or

```
permission check failed: <error message>
```

### Test Scenarios

#### Scenario 1: Check if User Has Read Permission on Workspace

```bash
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "resource_type": "workspace",
    "action": "read"
  }'
```

#### Scenario 2: Check if User Has Delete Permission on Specific Workspace

```bash
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "resource_type": "workspace",
    "resource_id": "ws-12345",
    "action": "delete"
  }'
```

#### Scenario 3: Check if User Has Permission to Create Template

```bash
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "resource_type": "template",
    "action": "create"
  }'
```

## Role Assignment API

### Endpoint
`POST /api/v1/rbac/assign-role`

### Description
Assign a role to a user.

### Request Example

```bash
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "role": "admin"
  }'
```

### Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| user_id | integer | Yes | User ID |
| role | string | Yes | Role name (e.g., admin, editor, viewer) |

### Response Examples

**Success (200 OK):**
```json
{
  "message": "role assigned successfully"
}
```

**Error (400 Bad Request):**
```
invalid request body: <error message>
```

or

```
role assignment failed: <error message>
```

### Test Scenarios

#### Scenario 1: Assign Admin Role to User

```bash
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "role": "admin"
  }'
```

#### Scenario 2: Assign Editor Role to User

```bash
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "role": "editor"
  }'
```

#### Scenario 3: Assign Viewer Role to User

```bash
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 3,
    "role": "viewer"
  }'
```

## Using Swagger UI

The service provides a Swagger UI interface for easier API testing.

### Access Swagger UI

1. Start the service:
   ```bash
   make run
   ```

2. Open in browser:
   ```
   http://localhost:8080/swagger/
   ```

3. You will see all available API endpoints with detailed information.

### Testing with Swagger UI

1. Select the API endpoint you want to test
2. Click the "Try it out" button
3. Fill in the request parameters
4. Click the "Execute" button
5. View the response results

## Common Issues

### 1. Connection Refused

**Problem:** `curl: (7) Failed to connect to localhost port 8080`

**Solution:**
- Ensure the service is running
- Check if the port is correct (default 8080)
- Check firewall settings

### 2. Database Connection Failed

**Problem:** Service fails to start with database connection error

**Solution:**
- Check if PostgreSQL is running
- Verify database configuration environment variables
- Ensure database has been created

### 3. Invalid Request Parameters

**Problem:** `400 Bad Request: user_id is required`

**Solution:**
- Ensure all required parameters are provided
- Check if JSON format is correct
- Verify parameter types match

## Automated Testing

You can use the following tools for automated API testing:

### Using Newman (Postman CLI)

1. Export Postman collection
2. Run tests:
   ```bash
   newman run rbac-api-tests.json
   ```

### Using Go Tests

Create integration tests:

```go
func TestCheckPermissionAPI(t *testing.T) {
    // Prepare request
    payload := map[string]interface{}{
        "user_id": 123,
        "resource_type": "workspace",
        "action": "read",
    }
    
    // Send request
    resp, err := http.Post("http://localhost:8080/api/v1/rbac/check", 
        "application/json", 
        bytes.NewBuffer(jsonPayload))
    
    // Verify response
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Logging and Debugging

### Enable Verbose Logging

Set log level to debug:

```bash
export LOG_LEVEL=debug
make run
```

### View Logs

Logs are output to standard output, including:
- Request details
- Permission check results
- Error messages

### Example Log Output

```
INFO[0000] Starting idekube-rbac
INFO[0001] RBAC HTTP server listening on :8080
DEBUG[0005] permission check sub=user:123 obj=workspace act=read allowed=true
```

## Performance Testing

### Using Apache Bench

```bash
ab -n 1000 -c 10 -p check.json -T application/json \
  http://localhost:8080/api/v1/rbac/check
```

### Using wrk

```bash
wrk -t 4 -c 100 -d 30s --latency \
  -s check.lua http://localhost:8080/api/v1/rbac/check
```

## Security Considerations

1. **Production Environment:** Ensure HTTPS is enabled
2. **Authentication:** Add authentication middleware in production
3. **Rate Limiting:** Consider adding API rate limiting
4. **Input Validation:** API includes basic validation, but should also validate on client side

## Related Documentation

- [API Definition Documentation](API.md) - Complete API specification
- [Quick Start](QUICKSTART.md) - Deployment guide
- [README](README.md) - Project overview
