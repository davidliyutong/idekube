# IDEKube Controller API 测试指南

本文档提供完整的API测试示例，包含curl命令和预期响应，帮助开发者快速测试和验证API功能。

> 📖 **相关文档**
> - [API定义文档](./API.md) - 查看完整的API规范和数据模型
> - [快速开始](./QUICKSTART.md) - 部署和配置指南  
> - [Swagger UI](http://localhost:8080/swagger/index.html) - 交互式API文档

## 测试环境准备

### 1. 启动服务

```bash
cd components/controller
make run
```

服务启动后可访问：
- API端点: http://localhost:8080/api/v1
- Swagger UI: http://localhost:8080/swagger/index.html
- 健康检查: http://localhost:8080/health

### 2. 安装测试工具

建议安装以下工具以便测试：

```bash
# jq - JSON处理工具
brew install jq  # macOS
apt install jq   # Ubuntu/Debian

# httpie - 友好的HTTP客户端 (可选)
brew install httpie  # macOS
apt install httpie   # Ubuntu/Debian
```

## 测试流程

按以下顺序进行测试可以构建完整的使用场景：

1. **认证** → 注册用户并获取Token
2. **组织** → 创建组织和管理成员
3. **模板** → 创建工作区模板
4. **存储卷** → 创建持久化存储
5. **工作区** → 创建和管理工作区

## 1. 认证 API

### 1.1 注册用户

**请求:**
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

**预期响应:**
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

### 1.2 用户登录

**请求:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

**预期响应:**
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

**保存Token到环境变量:**
```bash
export TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}' | jq -r '.data.token')

echo "Token saved: $TOKEN"
```

### 1.3 获取当前用户信息

**请求:**
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

## 2. 组织管理 API

### 2.1 创建组织
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

### 2.2 列出用户的组织
```bash
curl -X GET http://localhost:8080/api/v1/organizations \
  -H "Authorization: Bearer $TOKEN"
```

### 2.3 获取组织详情
```bash
curl -X GET http://localhost:8080/api/v1/organizations/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 2.4 添加组织成员
```bash
curl -X POST http://localhost:8080/api/v1/organizations/1/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "role": "member"
  }'
```

## 3. 模板管理 API

### 3.1 创建模板
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

### 3.2 列出可访问的模板
```bash
# 列出所有可访问的模板
curl -X GET http://localhost:8080/api/v1/templates \
  -H "Authorization: Bearer $TOKEN"

# 仅列出公开模板
curl -X GET "http://localhost:8080/api/v1/templates?public_only=true" \
  -H "Authorization: Bearer $TOKEN"
```

### 3.3 获取模板详情
```bash
curl -X GET http://localhost:8080/api/v1/templates/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 3.4 更新模板
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

### 3.5 删除模板
```bash
curl -X DELETE http://localhost:8080/api/v1/templates/1 \
  -H "Authorization: Bearer $TOKEN"
```

## 4. Volume 管理 API

### 4.1 创建 Volume
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

### 4.2 列出 Volumes
```bash
curl -X GET "http://localhost:8080/api/v1/volumes?owner_type=user&owner_id=1" \
  -H "Authorization: Bearer $TOKEN"
```

### 4.3 获取 Volume 详情
```bash
curl -X GET http://localhost:8080/api/v1/volumes/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 4.4 更新 Volume (扩容)
```bash
curl -X PUT http://localhost:8080/api/v1/volumes/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "size_mb": 20480
  }'
```

### 4.5 同步 Volume 状态
```bash
curl -X POST http://localhost:8080/api/v1/volumes/1/sync \
  -H "Authorization: Bearer $TOKEN"
```

### 4.6 删除 Volume
```bash
curl -X DELETE http://localhost:8080/api/v1/volumes/1 \
  -H "Authorization: Bearer $TOKEN"
```

## 5. Workspace 管理 API

### 5.1 创建 Workspace
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

### 5.2 列出 Workspaces
```bash
curl -X GET "http://localhost:8080/api/v1/workspaces?owner_type=user&owner_id=1" \
  -H "Authorization: Bearer $TOKEN"
```

### 5.3 获取 Workspace 详情
```bash
curl -X GET http://localhost:8080/api/v1/workspaces/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 5.4 更新 Workspace 资源
```bash
curl -X PUT http://localhost:8080/api/v1/workspaces/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cpu_millicores": 4000,
    "memory_mb": 8192
  }'
```

### 5.5 启动 Workspace
```bash
curl -X POST http://localhost:8080/api/v1/workspaces/1/start \
  -H "Authorization: Bearer $TOKEN"
```

### 5.6 停止 Workspace
```bash
curl -X POST http://localhost:8080/api/v1/workspaces/1/stop \
  -H "Authorization: Bearer $TOKEN"
```

### 5.7 挂载 Volume 到 Workspace
```bash
curl -X POST http://localhost:8080/api/v1/workspaces/1/volumes/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "mount_path": "/data"
  }'
```

### 5.8 卸载 Volume
```bash
curl -X DELETE http://localhost:8080/api/v1/workspaces/1/volumes/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 5.9 删除 Workspace
```bash
curl -X DELETE http://localhost:8080/api/v1/workspaces/1 \
  -H "Authorization: Bearer $TOKEN"
```

## 6. 完整工作流示例

### 6.1 创建完整开发环境
```bash
# 1. 创建模板
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

# 2. 创建 Volume
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

# 等待 Volume 就绪
sleep 5

# 3. 创建 Workspace
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

# 4. 挂载 Volume
curl -X POST http://localhost:8080/api/v1/workspaces/$WORKSPACE_ID/volumes/$VOLUME_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "mount_path": "/data"
  }'

echo "Attached volume to workspace"

# 5. 启动 Workspace
curl -X POST http://localhost:8080/api/v1/workspaces/$WORKSPACE_ID/start \
  -H "Authorization: Bearer $TOKEN"

echo "Started workspace"

# 6. 检查状态
curl -X GET http://localhost:8080/api/v1/workspaces/$WORKSPACE_ID \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### 6.2 清理资源
```bash
# 停止并删除 Workspace
curl -X POST http://localhost:8080/api/v1/workspaces/$WORKSPACE_ID/stop \
  -H "Authorization: Bearer $TOKEN"
sleep 2
curl -X DELETE http://localhost:8080/api/v1/workspaces/$WORKSPACE_ID \
  -H "Authorization: Bearer $TOKEN"

# 删除 Volume
curl -X DELETE http://localhost:8080/api/v1/volumes/$VOLUME_ID \
  -H "Authorization: Bearer $TOKEN"

# 删除 Template
curl -X DELETE http://localhost:8080/api/v1/templates/$TEMPLATE_ID \
  -H "Authorization: Bearer $TOKEN"
```

## 7. 错误处理

所有API响应遵循统一格式:

成功响应:
```json
{
  "success": true,
  "data": {...},
  "message": "Optional success message"
}
```

错误响应:
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

## 8. Kubernetes 资源命名规则

- **PVC**: `volume-{volume_id}-{uuid_prefix}` (例如: `volume-1-abc12345`)
- **Deployment**: `workspace-{workspace_id}-{uuid_prefix}` (例如: `workspace-1-def67890`)
- **Service**: `workspace-{workspace_id}-{uuid_prefix}` (例如: `workspace-1-def67890`)

## 9. 状态说明

### Volume 状态
- `pending`: PVC创建中
- `bound`: PVC已绑定到PV
- `failed`: 创建失败

### Workspace 状态
- `pending`: 初始创建中
- `starting`: Deployment启动中
- `running`: Pod运行中
- `stopped`: 已停止(replicas=0)
- `failed`: 启动失败

## 10. 配额限制

每个用户/组织都有配额限制:
- CPU总量 (millicores)
- 内存总量 (MB)
- 存储总量 (MB)
- GPU数量
- 最大Workspace数量
- 最大Volume数量

创建资源时会自动检查配额,超出限制将返回错误。

## 11. 下一步实现 (Phase 2)

以下功能在Phase 2中实现:
- OIDC单点登录
- 邮箱验证和密码重置
- 多因素认证(MFA)
- Webhooks事件通知
- API Keys管理
- 审计日志查询API
