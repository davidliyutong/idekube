# API定义文档

IDEKube RBAC API规范文档

## 版本信息

- **版本:** 1.0
- **协议:** HTTP, HTTPS
- **Base Path:** /api/v1
- **默认端口:** 8080

## 概述

IDEKube RBAC服务提供基于角色的访问控制（Role-Based Access Control）功能，用于管理idekube平台中的权限控制。该服务使用Casbin作为权限引擎，支持灵活的权限策略配置。

## 架构

```
┌─────────────┐
│  Controller │
│   Service   │
└──────┬──────┘
       │ HTTP API Calls
       ▼
┌─────────────┐
│    RBAC     │
│   Service   │
├─────────────┤
│  Casbin     │
│  Enforcer   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ PostgreSQL  │
└─────────────┘
```

## 认证

当前版本的API不需要认证。在生产环境中，建议添加以下认证机制：

- API Key认证
- JWT Token认证
- mTLS认证

## API端点

### 1. 健康检查

#### GET /healthz

检查服务是否正常运行。

**请求：**

```http
GET /healthz HTTP/1.1
Host: localhost:8080
```

**响应：**

- **200 OK**
  ```
  ok
  ```

**示例：**

```bash
curl http://localhost:8080/healthz
```

---

### 2. 检查权限

#### POST /api/v1/rbac/check

检查用户是否有权限对指定资源执行特定操作。

**请求：**

```http
POST /api/v1/rbac/check HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "user_id": 123,
  "resource_type": "workspace",
  "resource_id": "ws-001",
  "action": "read"
}
```

**请求体参数：**

| 字段 | 类型 | 必需 | 描述 | 示例 |
|------|------|------|------|------|
| user_id | integer | 是 | 用户ID | 123 |
| resource_type | string | 是 | 资源类型 | "workspace", "template", "volume" |
| resource_id | string | 否 | 特定资源的ID | "ws-001" |
| action | string | 是 | 操作类型 | "read", "write", "delete", "execute" |

**响应：**

- **200 OK**
  ```json
  {
    "allowed": true
  }
  ```
  
  或
  
  ```json
  {
    "allowed": false
  }
  ```

- **400 Bad Request**
  ```
  invalid request body: <error_message>
  ```
  
  或
  
  ```
  permission check failed: <error_message>
  ```

- **405 Method Not Allowed**
  
  当使用GET或其他非POST方法时返回。

**资源类型说明：**

| 资源类型 | 描述 | 支持的操作 |
|---------|------|-----------|
| workspace | 工作空间 | read, write, delete, execute |
| template | 模板 | read, write, delete, create |
| volume | 持久化卷 | read, write, delete, create |
| organization | 组织 | read, write, delete, manage |
| user | 用户 | read, write, delete, manage |

**权限检查逻辑：**

1. 如果提供了`resource_id`，检查用户对特定资源的权限
2. 如果只提供了`resource_type`，检查用户对该类型资源的通用权限
3. 权限检查通过Casbin enforcer进行，格式为：`(subject, object, action)`
   - subject: `user:<user_id>`
   - object: `<resource_type>` 或 `<resource_type>:<resource_id>`
   - action: 请求中的action

**示例：**

```bash
# 检查用户是否可以读取所有workspace
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "resource_type": "workspace",
    "action": "read"
  }'

# 检查用户是否可以删除特定workspace
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "resource_type": "workspace",
    "resource_id": "ws-12345",
    "action": "delete"
  }'
```

---

### 3. 分配角色

#### POST /api/v1/rbac/assign-role

为用户分配角色。

**请求：**

```http
POST /api/v1/rbac/assign-role HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "user_id": 123,
  "role": "admin"
}
```

**请求体参数：**

| 字段 | 类型 | 必需 | 描述 | 示例 |
|------|------|------|------|------|
| user_id | integer | 是 | 用户ID | 123 |
| role | string | 是 | 角色名称 | "admin", "editor", "viewer" |

**响应：**

- **200 OK**
  ```json
  {
    "message": "role assigned successfully"
  }
  ```

- **400 Bad Request**
  ```
  invalid request body: <error_message>
  ```
  
  或
  
  ```
  role assignment failed: <error_message>
  ```

- **405 Method Not Allowed**
  
  当使用GET或其他非POST方法时返回。

**预定义角色：**

| 角色 | 描述 | 权限范围 |
|------|------|---------|
| admin | 管理员 | 所有资源的所有操作 |
| editor | 编辑者 | 读取和写入权限，无删除权限 |
| viewer | 查看者 | 只有读取权限 |
| workspace-admin | 工作空间管理员 | 工作空间的所有操作 |
| template-admin | 模板管理员 | 模板的所有操作 |

**示例：**

```bash
# 将admin角色分配给用户1
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "role": "admin"
  }'

# 将viewer角色分配给用户2
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "role": "viewer"
  }'
```

---

## 数据模型

### CheckPermissionRequest

权限检查请求模型。

```json
{
  "user_id": 123,
  "resource_type": "workspace",
  "resource_id": "ws-001",
  "action": "read"
}
```

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| user_id | integer | 是 | 用户唯一标识符 |
| resource_type | string | 是 | 资源类型（workspace, template, volume等） |
| resource_id | string | 否 | 资源的唯一标识符，用于检查特定资源的权限 |
| action | string | 是 | 要执行的操作（read, write, delete, execute等） |

### AssignRoleRequest

角色分配请求模型。

```json
{
  "user_id": 123,
  "role": "admin"
}
```

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| user_id | integer | 是 | 用户唯一标识符 |
| role | string | 是 | 要分配的角色名称 |

---

## 错误码

### HTTP状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误或业务逻辑错误 |
| 405 | 请求方法不允许 |
| 500 | 服务器内部错误 |

### 错误消息格式

错误响应以纯文本形式返回：

```
<error_type>: <error_message>
```

**常见错误消息：**

| 错误消息 | 原因 |
|---------|------|
| `invalid request body: ...` | JSON解析失败或格式错误 |
| `user_id is required` | 缺少user_id参数 |
| `resource_type is required` | 缺少resource_type参数 |
| `action is required` | 缺少action参数 |
| `permission check failed: ...` | 权限检查过程中发生错误 |
| `role assignment failed: ...` | 角色分配过程中发生错误 |

---

## Casbin模型

服务使用Casbin作为权限引擎，采用RBAC模型。

### 模型定义

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

### 策略示例

```csv
p, role:admin, workspace, read
p, role:admin, workspace, write
p, role:admin, workspace, delete
p, role:admin, template, read
p, role:admin, template, write
p, role:editor, workspace, read
p, role:editor, workspace, write
p, role:viewer, workspace, read

g, user:1, role:admin
g, user:2, role:editor
g, user:3, role:viewer
```

### 权限继承

Casbin支持角色继承，例如：

```
role:super-admin -> role:admin -> role:editor -> role:viewer
```

---

## 集成指南

### Controller服务集成

在Controller服务中使用RBAC API的示例：

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type RBACClient struct {
    baseURL string
    client  *http.Client
}

func NewRBACClient(baseURL string) *RBACClient {
    return &RBACClient{
        baseURL: baseURL,
        client:  &http.Client{},
    }
}

func (c *RBACClient) CheckPermission(userID int64, resourceType, resourceID, action string) (bool, error) {
    req := map[string]interface{}{
        "user_id":       userID,
        "resource_type": resourceType,
        "resource_id":   resourceID,
        "action":        action,
    }

    body, _ := json.Marshal(req)
    resp, err := c.client.Post(
        c.baseURL+"/api/v1/rbac/check",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return false, fmt.Errorf("permission check failed with status: %d", resp.StatusCode)
    }

    var result map[string]bool
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return false, err
    }

    return result["allowed"], nil
}

func (c *RBACClient) AssignRole(userID int64, role string) error {
    req := map[string]interface{}{
        "user_id": userID,
        "role":    role,
    }

    body, _ := json.Marshal(req)
    resp, err := c.client.Post(
        c.baseURL+"/api/v1/rbac/assign-role",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("role assignment failed with status: %d", resp.StatusCode)
    }

    return nil
}

// 使用示例
func main() {
    client := NewRBACClient("http://localhost:8080")

    // 检查权限
    allowed, err := client.CheckPermission(1, "workspace", "ws-001", "read")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Permission allowed: %v\n", allowed)

    // 分配角色
    if err := client.AssignRole(2, "editor"); err != nil {
        panic(err)
    }
    fmt.Println("Role assigned successfully")
}
```

### 使用生成的Go客户端

使用 `make client-gen` 生成的客户端：

```go
import (
    "context"
    rbac "github.com/davidliyutong/idekube-rbac/client"
)

func main() {
    // 创建客户端
    client, err := rbac.NewClient("http://localhost:8080")
    if err != nil {
        panic(err)
    }

    // 使用客户端
    ctx := context.Background()
    
    // 检查权限
    checkReq := &rbac.CheckPermissionRequest{
        UserID:       123,
        ResourceType: "workspace",
        ResourceID:   "ws-001",
        Action:       "read",
    }
    
    resp, err := client.CheckPermission(ctx, checkReq)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Allowed: %v\n", resp.Allowed)
}
```

---

## Swagger/OpenAPI

### 访问Swagger UI

服务启动后，可以通过以下URL访问Swagger UI：

```
http://localhost:8080/swagger/
```

Swagger UI提供：
- 交互式API文档
- 在线API测试工具
- 模型定义查看
- 请求/响应示例

### 下载OpenAPI规范

OpenAPI规范文件位于：
- JSON格式: `docs/swagger.json`
- YAML格式: `docs/swagger.yaml`

可以使用这些文件：
- 生成客户端SDK
- 集成到API网关
- 自动化测试
- 文档生成

---

## 最佳实践

### 1. 权限检查

- **细粒度权限：** 在检查特定资源权限时，始终提供`resource_id`
- **缓存策略：** 考虑缓存权限检查结果以提高性能
- **失败处理：** 默认拒绝访问，只有明确允许时才授权

### 2. 角色管理

- **最小权限原则：** 为用户分配执行任务所需的最小权限
- **角色层次：** 利用角色继承简化权限管理
- **定期审计：** 定期检查用户角色和权限

### 3. 集成建议

- **中间件：** 在Controller服务中实现RBAC中间件
- **错误处理：** 优雅处理RBAC服务不可用的情况
- **重试机制：** 实现指数退避的重试策略
- **监控：** 监控权限检查的延迟和错误率

---

## 性能考虑

### 响应时间

- 健康检查: < 1ms
- 权限检查: < 10ms (无缓存)
- 角色分配: < 50ms

### 吞吐量

在典型硬件上的性能指标：
- 权限检查: ~5000 req/s
- 角色分配: ~1000 req/s

### 优化建议

1. **数据库索引：** 确保Casbin策略表有适当的索引
2. **连接池：** 配置适当的数据库连接池大小
3. **读写分离：** 考虑使用只读副本处理权限检查
4. **缓存：** 在应用层实现缓存机制

---

## 变更日志

### v1.0 (当前版本)

- 初始版本发布
- 实现基本的权限检查API
- 实现角色分配API
- 集成Swagger UI
- 支持Casbin RBAC模型

---

## 相关资源

- [API测试指南](API_TESTING.md)
- [快速开始指南](QUICKSTART.md)
- [Casbin文档](https://casbin.org/docs/overview)
- [OpenAPI规范](https://swagger.io/specification/)

## 支持

如有问题或建议，请：
- 提交Issue到项目仓库
- 联系: support@idekube.io
- 查看项目Wiki
