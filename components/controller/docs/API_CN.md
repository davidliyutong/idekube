# IDEKube Controller API 定义文档

本文档描述IDEKube Controller的REST API接口定义、数据模型和使用说明。

## 概述

IDEKube Controller是一个云IDE平台的核心API服务器，提供工作区、模板、用户和组织管理功能。

- **版本**: v1.0
- **Base URL**: `/api/v1`
- **认证方式**: JWT Bearer Token

## 交互式文档

启动服务后，访问以下URL查看Swagger UI交互式文档：

```
http://localhost:8080/swagger/index.html
```

## 认证

大部分API需要JWT令牌认证。在请求头中添加：

```
Authorization: Bearer <your_jwt_token>
```

### 获取Token

通过登录接口获取JWT令牌：

**POST** `/api/v1/auth/login`

```json
{
  "username": "your_username",
  "password": "your_password"
}
```

响应：
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": { ... }
  }
}
```

## API 分组

### 1. 认证 (Authentication)

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| POST | `/auth/register` | 用户注册 | ❌ |
| POST | `/auth/login` | 用户登录 | ❌ |

### 2. 用户 (Users)

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| GET | `/users/me` | 获取当前用户信息 | ✅ |
| POST | `/users/me/password` | 修改密码 | ✅ |
| GET | `/users/:id` | 获取用户详情 | ✅ |
| GET | `/users` | 列出所有用户 (Admin) | ✅ |
| PUT | `/users/:id` | 更新用户信息 (Admin) | ✅ |
| DELETE | `/users/:id` | 删除用户 (Admin) | ✅ |

### 3. 组织 (Organizations)

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| POST | `/organizations` | 创建组织 | ✅ |
| GET | `/organizations` | 列出用户的组织 | ✅ |
| GET | `/organizations/:id` | 获取组织详情 | ✅ |
| PUT | `/organizations/:id` | 更新组织信息 | ✅ |
| DELETE | `/organizations/:id` | 删除组织 | ✅ |
| POST | `/organizations/:id/members` | 添加成员 | ✅ |
| DELETE | `/organizations/:id/members/:user_id` | 移除成员 | ✅ |
| PUT | `/organizations/:id/members/:user_id` | 更新成员角色 | ✅ |

### 4. 模板 (Templates)

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| POST | `/templates` | 创建模板 | ✅ |
| GET | `/templates` | 列出模板 | ✅ |
| GET | `/templates/:id` | 获取模板详情 | ✅ |
| PUT | `/templates/:id` | 更新模板 | ✅ |
| DELETE | `/templates/:id` | 删除模板 | ✅ |

### 5. 工作区 (Workspaces)

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| POST | `/workspaces` | 创建工作区 | ✅ |
| GET | `/workspaces` | 列出工作区 | ✅ |
| GET | `/workspaces/:id` | 获取工作区详情 | ✅ |
| PUT | `/workspaces/:id` | 更新工作区 | ✅ |
| DELETE | `/workspaces/:id` | 删除工作区 | ✅ |
| POST | `/workspaces/:id/start` | 启动工作区 | ✅ |
| POST | `/workspaces/:id/stop` | 停止工作区 | ✅ |
| POST | `/workspaces/:id/volumes/:volume_id` | 挂载存储卷 | ✅ |
| DELETE | `/workspaces/:id/volumes/:volume_id` | 卸载存储卷 | ✅ |

### 6. 存储卷 (Volumes)

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| POST | `/volumes` | 创建存储卷 | ✅ |
| GET | `/volumes` | 列出存储卷 | ✅ |
| GET | `/volumes/:id` | 获取存储卷详情 | ✅ |
| PUT | `/volumes/:id` | 更新存储卷 | ✅ |
| DELETE | `/volumes/:id` | 删除存储卷 | ✅ |
| POST | `/volumes/:id/sync` | 同步存储卷状态 | ✅ |

## 数据模型

### User (用户)

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

### Organization (组织)

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

### Template (模板)

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

### Workspace (工作区)

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

### Volume (存储卷)

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

## 通用响应格式

所有API响应都遵循统一的格式：

### 成功响应

```json
{
  "success": true,
  "data": { ... },
  "message": "操作成功"
}
```

### 错误响应

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "错误描述",
    "details": "详细错误信息"
  }
}
```

### 常见错误码

| 错误码 | HTTP状态码 | 描述 |
|--------|------------|------|
| INVALID_REQUEST | 400 | 请求参数错误 |
| UNAUTHORIZED | 401 | 未认证或Token无效 |
| FORBIDDEN | 403 | 无权限访问 |
| NOT_FOUND | 404 | 资源不存在 |
| CONFLICT | 409 | 资源冲突 |
| INTERNAL_ERROR | 500 | 服务器内部错误 |

## 生成客户端SDK

### JavaScript客户端

```bash
cd components/controller
make swagger-gen          # 先生成Swagger文档
make swagger-js-client    # 生成JavaScript客户端
```

生成的客户端代码位于 `./client-js/` 目录。

### TypeScript客户端

```bash
make swagger-ts-client    # 生成TypeScript客户端
```

生成的客户端代码位于 `./client-ts/` 目录。

## 更新Swagger文档

当API发生变更时，重新生成Swagger文档：

```bash
make swagger-gen
```

这将扫描代码中的Swagger注解并更新文档。

## 参考资源

- [Swagger UI](http://localhost:8080/swagger/index.html) - 交互式API文档
- [API测试指南](./API_TESTING.md) - 详细的API测试示例
- [快速开始](./QUICKSTART.md) - 部署和运行指南
