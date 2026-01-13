````markdown
# IDEKube Controller API Testing Guide

This document provides complete API testing examples, including curl commands and expected responses, to help developers quickly test and verify API functionality.

> 📖 **Related Documentation**
> - [API Documentation](./API.md) - View complete API specifications and data models
> - [Quick Start](./QUICKSTART.md) - Deployment and configuration guide
> - [Swagger UI](http://localhost:8080/swagger/index.html) - Interactive API documentation

## Test Environment Setup

### 1. Start Service

```bash
cd components/controller
make run
```

After service starts, you can access:
- API endpoint: http://localhost:8080/api/v1
- Swagger UI: http://localhost:8080/swagger/index.html
- Health check: http://localhost:8080/health

### 2. Install Testing Tools

Recommended tools for testing:

```bash
# jq - JSON processing tool
brew install jq  # macOS
apt install jq   # Ubuntu/Debian

# httpie - Friendly HTTP client (optional)
brew install httpie  # macOS
apt install httpie   # Ubuntu/Debian
```

## Testing Workflow

Follow this order to build a complete usage scenario:

1. **Authentication** → Register user and get token
2. **Organizations** → Create organization and manage members
3. **Templates** → Create workspace templates
4. **Volumes** → Create persistent storage
5. **Workspaces** → Create and manage workspaces

## 1. Authentication API

### 1.1 Register User

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "full_name": "Test User",
    "role": "user",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 1.2 User Login

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "full_name": "Test User",
      "role": "user"
    }
  }
}
```

**Save Token to Environment Variable:**
```bash
export TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}' | jq -r '.data.token')

echo "Token saved: $TOKEN"
```

### 1.3 Get Current User Info

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

## 2. Organization Management API

### 2.1 Create Organization
```bash
curl -X POST http://localhost:8080/api/v1/organizations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-org",
    "display_name": "Test Organization",
    "description": "A test organization"
  }'
```

### 2.2 List User's Organizations
```bash
curl -X GET http://localhost:8080/api/v1/organizations \
  -H "Authorization: Bearer $TOKEN"
```

### 2.3 Get Organization Details
```bash
curl -X GET http://localhost:8080/api/v1/organizations/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 2.4 Add Organization Member
```bash
curl -X POST http://localhost:8080/api/v1/organizations/1/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "role": "member"
  }'
```

## 3. Template Management API

### 3.1 Create Template
```bash
curl -X POST http://localhost:8080/api/v1/templates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "python-dev",
    "display_name": "Python Development",
    "description": "Python development environment with common tools",
    "owner_type": "user",
    "owner_id": 1,
    "image": "python:3.11-slim",
    "is_public": true,
    "manifest_yaml": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: python-dev\nspec:\n  containers:\n  - name: python\n    image: python:3.11-slim",
    "default_cpu_millicores": 1000,
    "default_memory_mb": 2048,
    "default_storage_mb": 10240
  }'
```

### 3.2 List Accessible Templates
```bash
# List all accessible templates
curl -X GET http://localhost:8080/api/v1/templates \
  -H "Authorization: Bearer $TOKEN"

# List only public templates
curl -X GET "http://localhost:8080/api/v1/templates?public_only=true" \
  -H "Authorization: Bearer $TOKEN"
```

### 3.3 Get Template Details
```bash
curl -X GET http://localhost:8080/api/v1/templates/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 3.4 Update Template
```bash
curl -X PUT http://localhost:8080/api/v1/templates/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Python Development (Updated)",
    "description": "Updated Python development environment",
    "default_cpu_millicores": 2000,
    "default_memory_mb": 4096
  }'
```

### 3.5 Delete Template
```bash
curl -X DELETE http://localhost:8080/api/v1/templates/1 \
  -H "Authorization: Bearer $TOKEN"
```

## 4. Volume Management API

### 4.1 Create Volume
```bash
curl -X POST http://localhost:8080/api/v1/volumes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "data-volume",
    "display_name": "Data Volume",
    "description": "Persistent data storage",
    "owner_type": "user",
    "owner_id": 1,
    "size_mb": 10240,
    "access_mode": "ReadWriteOnce"
  }'
```

### 4.2 List Volumes
```bash
curl -X GET "http://localhost:8080/api/v1/volumes?owner_type=user&owner_id=1" \
  -H "Authorization: Bearer $TOKEN"
```

### 4.3 Get Volume Details
```bash
curl -X GET http://localhost:8080/api/v1/volumes/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 4.4 Update Volume (Expand)
```bash
curl -X PUT http://localhost:8080/api/v1/volumes/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "size_mb": 20480
  }'
```

### 4.5 Sync Volume Status
```bash
curl -X POST http://localhost:8080/api/v1/volumes/1/sync \
  -H "Authorization: Bearer $TOKEN"
```

### 4.6 Delete Volume
```bash
curl -X DELETE http://localhost:8080/api/v1/volumes/1 \
  -H "Authorization: Bearer $TOKEN"
```

## 5. Workspace Management API

### 5.1 Create Workspace
```bash
curl -X POST http://localhost:8080/api/v1/workspaces \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-workspace",
    "display_name": "My Development Workspace",
    "description": "Python development workspace",
    "owner_type": "user",
    "owner_id": 1,
    "template_id": 1,
    "cpu_millicores": 2000,
    "memory_mb": 4096,
    "storage_mb": 20480
  }'
```

### 5.2 List Workspaces
```bash
curl -X GET "http://localhost:8080/api/v1/workspaces?owner_type=user&owner_id=1" \
  -H "Authorization: Bearer $TOKEN"
```

### 5.3 Get Workspace Details
```bash
curl -X GET http://localhost:8080/api/v1/workspaces/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 5.4 Update Workspace Resources
```bash
curl -X PUT http://localhost:8080/api/v1/workspaces/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cpu_millicores": 4000,
    "memory_mb": 8192
  }'
```

### 5.5 Start Workspace
```bash
curl -X POST http://localhost:8080/api/v1/workspaces/1/start \
  -H "Authorization: Bearer $TOKEN"
```

### 5.6 Stop Workspace
```bash
curl -X POST http://localhost:8080/api/v1/workspaces/1/stop \
  -H "Authorization: Bearer $TOKEN"
```

### 5.7 Mount Volume to Workspace
```bash
curl -X POST http://localhost:8080/api/v1/workspaces/1/volumes/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "mount_path": "/data"
  }'
```

### 5.8 Unmount Volume
```bash
curl -X DELETE http://localhost:8080/api/v1/workspaces/1/volumes/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 5.9 Delete Workspace
```bash
curl -X DELETE http://localhost:8080/api/v1/workspaces/1 \
  -H "Authorization: Bearer $TOKEN"
```

## 6. Complete Workflow Example

### 6.1 Create Complete Development Environment
```bash
# 1. Create template
TEMPLATE_ID=$(curl -s -X POST http://localhost:8080/api/v1/templates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "python-ml",
    "display_name": "Python ML Environment",
    "description": "Python with ML libraries",
    "owner_type": "user",
    "owner_id": 1,
    "image": "python:3.11",
    "is_public": false,
    "manifest_yaml": "apiVersion: v1\nkind: Pod",
    "default_cpu_millicores": 2000,
    "default_memory_mb": 4096,
    "default_storage_mb": 20480
  }' | jq -r '.data.id')

echo "Created template: $TEMPLATE_ID"

# 2. Create Volume
VOLUME_ID=$(curl -s -X POST http://localhost:8080/api/v1/volumes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ml-data",
    "display_name": "ML Data Volume",
    "description": "Storage for ML datasets",
    "owner_type": "user",
    "owner_id": 1,
    "size_mb": 51200,
    "access_mode": "ReadWriteOnce"
  }' | jq -r '.data.id')

echo "Created volume: $VOLUME_ID"

# Wait for Volume to be ready
sleep 5

# 3. Create Workspace
WORKSPACE_ID=$(curl -s -X POST http://localhost:8080/api/v1/workspaces \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"ml-workspace\",
    \"display_name\": \"ML Development\",
    \"description\": \"Machine learning workspace\",
    \"owner_type\": \"user\",
    \"owner_id\": 1,
    \"template_id\": $TEMPLATE_ID,
    \"cpu_millicores\": 4000,
    \"memory_mb\": 8192,
    \"storage_mb\": 20480
  }" | jq -r '.data.id')

echo "Created workspace: $WORKSPACE_ID"

# 4. Mount Volume
curl -X POST http://localhost:8080/api/v1/workspaces/$WORKSPACE_ID/volumes/$VOLUME_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "mount_path": "/data"
  }'

echo "Attached volume to workspace"

# 5. Start Workspace
curl -X POST http://localhost:8080/api/v1/workspaces/$WORKSPACE_ID/start \
  -H "Authorization: Bearer $TOKEN"

echo "Started workspace"

# 6. Check status
curl -X GET http://localhost:8080/api/v1/workspaces/$WORKSPACE_ID \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### 6.2 Clean Up Resources
```bash
# Stop and delete Workspace
curl -X POST http://localhost:8080/api/v1/workspaces/$WORKSPACE_ID/stop \
  -H "Authorization: Bearer $TOKEN"
sleep 2
curl -X DELETE http://localhost:8080/api/v1/workspaces/$WORKSPACE_ID \
  -H "Authorization: Bearer $TOKEN"

# Delete Volume
curl -X DELETE http://localhost:8080/api/v1/volumes/$VOLUME_ID \
  -H "Authorization: Bearer $TOKEN"

# Delete Template
curl -X DELETE http://localhost:8080/api/v1/templates/$TEMPLATE_ID \
  -H "Authorization: Bearer $TOKEN"
```

## 7. Error Handling

All API responses follow a unified format:

Success response:
```json
{
  "success": true,
  "data": {...},
  "message": "Optional success message"
}
```

Error response:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": "Additional error details"
  }
}
```

## 8. Kubernetes Resource Naming Rules

- **PVC**: `volume-{volume_id}-{uuid_prefix}` (e.g., `volume-1-abc12345`)
- **Deployment**: `workspace-{workspace_id}-{uuid_prefix}` (e.g., `workspace-1-def67890`)
- **Service**: `workspace-{workspace_id}-{uuid_prefix}` (e.g., `workspace-1-def67890`)

## 9. Status Descriptions

### Volume Status
- `pending`: PVC being created
- `bound`: PVC bound to PV
- `failed`: Creation failed

### Workspace Status
- `pending`: Initial creation in progress
- `starting`: Deployment starting
- `running`: Pod running
- `stopped`: Stopped (replicas=0)
- `failed`: Startup failed

## 10. Quota Limits

Each user/organization has quota limits:
- Total CPU (millicores)
- Total memory (MB)
- Total storage (MB)
- GPU count
- Maximum workspace count
- Maximum volume count

Quotas are automatically checked when creating resources; exceeding limits will return an error.

## 11. Next Steps (Phase 2)

The following features will be implemented in Phase 2:
- OIDC single sign-on
- Email verification and password reset
- Multi-factor authentication (MFA)
- Webhook event notifications
- API key management
- Audit log query API

````