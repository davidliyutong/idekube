````markdown
# IDEKube Controller API Documentation

This document describes the REST API interfaces, data models, and usage instructions for IDEKube Controller.

## Overview

IDEKube Controller is the core API server for a cloud IDE platform, providing workspace, template, user, and organization management capabilities.

- **Version**: v1.0
- **Base URL**: `/api/v1`
- **Authentication**: JWT Bearer Token

## Interactive Documentation

After starting the service, visit the following URL to view the Swagger UI interactive documentation:

```
http://localhost:8080/swagger/index.html
```

## Authentication

Most APIs require JWT token authentication. Add the following to the request header:

```
Authorization: Bearer <your_jwt_token>
```

### Getting Token

Obtain a JWT token through the login endpoint:

**POST** `/api/v1/auth/login`

```json
{
  "username": "your_username",
  "password": "your_password"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": { ... }
  }
}
```

## API Groups

### 1. Authentication

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/auth/register` | User registration | ❌ |
| POST | `/auth/login` | User login | ❌ |

### 2. Users

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/users/me` | Get current user info | ✅ |
| POST | `/users/me/password` | Change password | ✅ |
| GET | `/users/:id` | Get user details | ✅ |
| GET | `/users` | List all users (Admin) | ✅ |
| PUT | `/users/:id` | Update user info (Admin) | ✅ |
| DELETE | `/users/:id` | Delete user (Admin) | ✅ |

### 3. Organizations

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/organizations` | Create organization | ✅ |
| GET | `/organizations` | List user's organizations | ✅ |
| GET | `/organizations/:id` | Get organization details | ✅ |
| PUT | `/organizations/:id` | Update organization info | ✅ |
| DELETE | `/organizations/:id` | Delete organization | ✅ |
| POST | `/organizations/:id/members` | Add member | ✅ |
| DELETE | `/organizations/:id/members/:user_id` | Remove member | ✅ |
| PUT | `/organizations/:id/members/:user_id` | Update member role | ✅ |

### 4. Templates

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/templates` | Create template | ✅ |
| GET | `/templates` | List templates | ✅ |
| GET | `/templates/:id` | Get template details | ✅ |
| PUT | `/templates/:id` | Update template | ✅ |
| DELETE | `/templates/:id` | Delete template | ✅ |

### 5. Workspaces

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/workspaces` | Create workspace | ✅ |
| GET | `/workspaces` | List workspaces | ✅ |
| GET | `/workspaces/:id` | Get workspace details | ✅ |
| PUT | `/workspaces/:id` | Update workspace | ✅ |
| DELETE | `/workspaces/:id` | Delete workspace | ✅ |
| POST | `/workspaces/:id/start` | Start workspace | ✅ |
| POST | `/workspaces/:id/stop` | Stop workspace | ✅ |
| POST | `/workspaces/:id/volumes/:volume_id` | Mount volume | ✅ |
| DELETE | `/workspaces/:id/volumes/:volume_id` | Unmount volume | ✅ |

### 6. Volumes

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/volumes` | Create volume | ✅ |
| GET | `/volumes` | List volumes | ✅ |
| GET | `/volumes/:id` | Get volume details | ✅ |
| PUT | `/volumes/:id` | Update volume | ✅ |
| DELETE | `/volumes/:id` | Delete volume | ✅ |
| POST | `/volumes/:id/sync` | Sync volume status | ✅ |

## Data Models

### User

```json
{
  "id": 1,
  "username": "alice",
  "email": "alice@example.com",
  "full_name": "Alice Smith",
  "role": "user",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Organization

```json
{
  "id": 1,
  "name": "my-org",
  "display_name": "My Organization",
  "description": "Organization description",
  "owner_id": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Template

```json
{
  "id": 1,
  "name": "python-dev",
  "display_name": "Python Development",
  "description": "Python development environment",
  "owner_type": "user",
  "owner_id": 1,
  "image": "python:3.11",
  "cpu_request": "500m",
  "memory_request": "1Gi",
  "cpu_limit": "2",
  "memory_limit": "4Gi",
  "storage_class": "standard",
  "storage_size": "10Gi",
  "is_public": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Workspace

```json
{
  "id": 1,
  "name": "my-workspace",
  "display_name": "My Workspace",
  "template_id": 1,
  "owner_id": 1,
  "status": "running",
  "cpu_request": "500m",
  "memory_request": "1Gi",
  "cpu_limit": "2",
  "memory_limit": "4Gi",
  "ports": [
    {
      "name": "http",
      "port": 8080,
      "target_port": 8080,
      "protocol": "TCP"
    }
  ],
  "environment": {
    "KEY": "value"
  },
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Volume

```json
{
  "id": 1,
  "name": "my-volume",
  "display_name": "My Volume",
  "owner_id": 1,
  "size": "10Gi",
  "storage_class": "standard",
  "status": "bound",
  "access_modes": ["ReadWriteOnce"],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

## Common Response Format

All API responses follow a unified format:

### Success Response

```json
{
  "success": true,
  "data": { ... },
  "message": "Operation successful"
}
```

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description",
    "details": "Detailed error information"
  }
}
```

### Common Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| INVALID_REQUEST | 400 | Invalid request parameters |
| UNAUTHORIZED | 401 | Unauthenticated or invalid token |
| FORBIDDEN | 403 | No permission to access |
| NOT_FOUND | 404 | Resource not found |
| CONFLICT | 409 | Resource conflict |
| INTERNAL_ERROR | 500 | Internal server error |

## Generate Client SDK

### JavaScript Client

```bash
cd components/controller
make swagger-gen          # First generate Swagger documentation
make swagger-js-client    # Generate JavaScript client
```

The generated client code is located in the `./client-js/` directory.

### TypeScript Client

```bash
make swagger-ts-client    # Generate TypeScript client
```

The generated client code is located in the `./client-ts/` directory.

## Update Swagger Documentation

When the API changes, regenerate the Swagger documentation:

```bash
make swagger-gen
```

This will scan the Swagger annotations in the code and update the documentation.

## Reference Resources

- [Swagger UI](http://localhost:8080/swagger/index.html) - Interactive API documentation
- [API Testing Guide](./API_TESTING.md) - Detailed API testing examples
- [Quick Start](./QUICKSTART.md) - Deployment and operation guide

````