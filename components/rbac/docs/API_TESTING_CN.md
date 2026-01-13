# API测试指南

本文档介绍如何测试idekube-rbac API。

## 目录

- [环境准备](#环境准备)
- [健康检查](#健康检查)
- [权限检查API](#权限检查api)
- [角色分配API](#角色分配api)
- [使用Swagger UI](#使用swagger-ui)

## 环境准备

### 启动服务

在测试API之前，确保服务正在运行：

```bash
# 本地开发模式运行
make run

# 或者构建并运行
make build
./bin/idekube-rbac
```

默认情况下，服务会在端口8080上启动。确保在配置中设置了正确的环境变量。

### 必需的环境变量

```bash
# PostgreSQL配置
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=password
export POSTGRES_DB=idekube_rbac

# RabbitMQ配置
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export RABBITMQ_USER=guest
export RABBITMQ_PASSWORD=guest

# 应用配置
export HTTP_PORT=8080
export LOG_LEVEL=debug
```

## 健康检查

验证服务是否正常运行：

```bash
curl -X GET http://localhost:8080/healthz
```

**预期响应：**
```
ok
```

## 权限检查API

### 端点
`POST /api/v1/rbac/check`

### 描述
检查用户是否有权限对指定资源执行特定操作。

### 请求示例

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

### 请求参数

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| user_id | integer | 是 | 用户ID |
| resource_type | string | 是 | 资源类型（如：workspace, template, volume） |
| resource_id | string | 否 | 资源ID（可选，用于特定资源的权限检查） |
| action | string | 是 | 操作类型（如：read, write, delete, execute） |

### 响应示例

**成功 (200 OK)：**
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

**错误 (400 Bad Request)：**
```
invalid request body: <error message>
```

或

```
permission check failed: <error message>
```

### 测试场景

#### 场景1：检查用户是否有读取工作空间的权限

```bash
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "resource_type": "workspace",
    "action": "read"
  }'
```

#### 场景2：检查用户是否有删除特定工作空间的权限

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

#### 场景3：检查用户是否有创建模板的权限

```bash
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "resource_type": "template",
    "action": "create"
  }'
```

## 角色分配API

### 端点
`POST /api/v1/rbac/assign-role`

### 描述
为用户分配角色。

### 请求示例

```bash
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "role": "admin"
  }'
```

### 请求参数

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| user_id | integer | 是 | 用户ID |
| role | string | 是 | 角色名称（如：admin, editor, viewer） |

### 响应示例

**成功 (200 OK)：**
```json
{
  "message": "role assigned successfully"
}
```

**错误 (400 Bad Request)：**
```
invalid request body: <error message>
```

或

```
role assignment failed: <error message>
```

### 测试场景

#### 场景1：将管理员角色分配给用户

```bash
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "role": "admin"
  }'
```

#### 场景2：将编辑者角色分配给用户

```bash
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "role": "editor"
  }'
```

#### 场景3：将查看者角色分配给用户

```bash
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 3,
    "role": "viewer"
  }'
```

## 使用Swagger UI

服务提供了Swagger UI界面，可以更方便地测试API。

### 访问Swagger UI

1. 启动服务：
   ```bash
   make run
   ```

2. 在浏览器中打开：
   ```
   http://localhost:8080/swagger/
   ```

3. 你将看到所有可用的API端点及其详细信息。

### 使用Swagger UI测试

1. 选择你想测试的API端点
2. 点击"Try it out"按钮
3. 填写请求参数
4. 点击"Execute"按钮
5. 查看响应结果

## 常见问题

### 1. 连接被拒绝

**问题：** `curl: (7) Failed to connect to localhost port 8080`

**解决方案：**
- 确保服务正在运行
- 检查端口是否正确（默认8080）
- 检查防火墙设置

### 2. 数据库连接失败

**问题：** 服务启动失败，提示数据库连接错误

**解决方案：**
- 检查PostgreSQL是否在运行
- 验证数据库配置环境变量
- 确保数据库已创建

### 3. 无效的请求参数

**问题：** `400 Bad Request: user_id is required`

**解决方案：**
- 确保所有必需参数都已提供
- 检查JSON格式是否正确
- 验证参数类型是否匹配

## 自动化测试

可以使用以下工具进行自动化API测试：

### 使用Newman (Postman CLI)

1. 导出Postman集合
2. 运行测试：
   ```bash
   newman run rbac-api-tests.json
   ```

### 使用Go测试

创建集成测试：

```go
func TestCheckPermissionAPI(t *testing.T) {
    // 准备请求
    payload := map[string]interface{}{
        "user_id": 123,
        "resource_type": "workspace",
        "action": "read",
    }
    
    // 发送请求
    resp, err := http.Post("http://localhost:8080/api/v1/rbac/check", 
        "application/json", 
        bytes.NewBuffer(jsonPayload))
    
    // 验证响应
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## 日志和调试

### 启用详细日志

设置日志级别为debug：

```bash
export LOG_LEVEL=debug
make run
```

### 查看日志

日志会输出到标准输出，包含：
- 请求详情
- 权限检查结果
- 错误信息

### 示例日志输出

```
INFO[0000] Starting idekube-rbac
INFO[0001] RBAC HTTP server listening on :8080
DEBUG[0005] permission check sub=user:123 obj=workspace act=read allowed=true
```

## 性能测试

### 使用Apache Bench

```bash
ab -n 1000 -c 10 -p check.json -T application/json \
  http://localhost:8080/api/v1/rbac/check
```

### 使用wrk

```bash
wrk -t 4 -c 100 -d 30s --latency \
  -s check.lua http://localhost:8080/api/v1/rbac/check
```

## 安全注意事项

1. **生产环境：** 确保启用HTTPS
2. **认证：** 在生产环境中添加认证中间件
3. **速率限制：** 考虑添加API速率限制
4. **输入验证：** API已包含基本验证，但应在客户端也进行验证

## 相关文档

- [API定义文档](API.md) - 完整的API规范
- [快速开始](QUICKSTART.md) - 部署指南
- [README](README.md) - 项目概述
