# idekube-rbac

基于角色的访问控制（RBAC）服务，用于管理idekube平台中的权限控制。

## 特性

- ✅ Kubernetes CRD操作支持
- ✅ PostgreSQL集成
- ✅ 基于OPA (Open Policy Agent)的灵活权限引擎
- ✅ Rego策略语言支持
- ✅ RESTful API
- ✅ 结构化日志（zap）
- ✅ 云原生配置（环境变量）
- ✅ 数据库迁移工具

## 快速开始

详细的部署指南请参考 [快速开始文档](QUICKSTART.md)。

### 本地运行

```bash
# 安装依赖
make deps

# 配置环境变量（参考QUICKSTART.md）
source .env

# 构建
make build

# 运行
make run
```

服务将在 `http://localhost:8080` 启动。

## 文档

- **[API定义文档](API.md)** - 完整的API规范、数据模型和集成指南
- **[API测试指南](API_TESTING.md)** - API测试方法和示例
- **[快速开始](QUICKSTART.md)** - 详细的部署和配置指南

## API概览

### 健康检查
```bash
GET /healthz
```

### 权限检查
```bash
POST /api/v1/rbac/check
{
  "user_id": 123,
  "resource_type": "workspace",
  "resource_id": "ws-001",
  "action": "read"
}
```

### 角色分配
```bash
POST /api/v1/rbac/assign-role
{
  "user_id": 123,
  "role": "admin"
}
```

完整API文档请查看 [API.md](API.md)。

## 开发

### 前置要求

- Go 1.21+
- Docker (用于构建镜像)
- Kubernetes集群 (用于测试)
- PostgreSQL 12+
- RabbitMQ 3.9+

### 构建

```bash
# 构建二进制文件
make build

# 构建Docker镜像
make docker-build

# 推送Docker镜像
make docker-push
```

### 测试

```bash
# 运行测试
make test

# 运行带覆盖率的测试
make test-coverage
```

### 生成文档和客户端

```bash
# 生成Swagger文档
make swagger-gen

# 生成Golang客户端（供Controller使用）
make client-gen

# 查看Swagger UI（需要先运行服务）
make swagger-serve
```

生成的文件：
- Swagger文档: `docs/swagger.json`, `docs/swagger.yaml`
- Golang客户端: `client/client.go`

### 代码格式化和检查

```bash
# 格式化代码
make fmt

# 代码检查
make lint
```

## 配置

服务通过环境变量进行配置。详细配置说明请参考 [QUICKSTART.md](QUICKSTART.md#配置说明)。

### 核心配置

#### Kubernetes
- `KUBECONFIG`: kubeconfig文件路径（集群内运行时可选）
- `NAMESPACE`: 监听的命名空间（默认：所有命名空间）

#### PostgreSQL
- `POSTGRES_HOST`: PostgreSQL主机（默认：localhost）
- `POSTGRES_PORT`: PostgreSQL端口（默认：5432）
- `POSTGRES_USER`: PostgreSQL用户
- `POSTGRES_PASSWORD`: PostgreSQL密码
- `POSTGRES_DB`: PostgreSQL数据库名

#### RabbitMQ
- `RABBITMQ_HOST`: RabbitMQ主机（默认：localhost）
- `RABBITMQ_PORT`: RabbitMQ端口（默认：5672）
- `RABBITMQ_USER`: RabbitMQ用户
- `RABBITMQ_PASSWORD`: RabbitMQ密码
- `RABBITMQ_VHOST`: RabbitMQ虚拟主机（默认：/）

#### 应用配置
- `HTTP_PORT`: HTTP服务端口（默认：8080）
- `LOG_LEVEL`: 日志级别（默认：info）
- `WORKER_THREADS`: 工作线程数（默认：1）
- `CASBIN_MODEL_PATH`: Casbin模型文件路径
- `CASBIN_POLICY_PATH`: Casbin策略文件路径

## 部署

### Docker

```bash
docker run -d \
  -p 8080:8080 \
  -e POSTGRES_HOST=postgres \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=idekube_rbac \
  -e RABBITMQ_HOST=rabbitmq \
  -e RABBITMQ_PORT=5672 \
  -e RABBITMQ_USER=guest \
  -e RABBITMQ_PASSWORD=guest \
  davidliyutong/idekube-rbac:latest
```

### Kubernetes

使用项目根目录的Helm Chart进行部署：

```bash
helm install idekube-rbac ../../helm \
  -n idekube \
  --create-namespace
```

详细部署说明请参考 [QUICKSTART.md](QUICKSTART.md#kubernetes部署)。

## 项目结构

```
.
├── cmd/
│   └── rbac/           # 主程序入口
├── internal/
│   ├── api/            # HTTP API服务器
│   ├── config/         # 配置管理
│   ├── permission/     # 权限检查服务
│   └── rbac/           # RBAC核心服务
├── pkg/
│   ├── database/       # 数据库客户端
│   ├── k8s/            # Kubernetes客户端
│   ├── logger/         # 日志工具
│   └── queue/          # 消息队列客户端
├── configs/            # 配置文件
├── docs/               # Swagger文档（自动生成）
├── client/             # Golang客户端（自动生成）
├── Dockerfile
├── Makefile
└── README.md
```

## 技术栈

- **编程语言:** Go 1.21+
- **权限引擎:** [Casbin](https://casbin.org/) - 强大的权限管理框架
- **数据库:** PostgreSQL - 存储Casbin策略
- **消息队列:** RabbitMQ - 异步消息处理
- **API文档:** Swagger/OpenAPI - 自动生成API文档
- **容器化:** Docker - 容器化部署
- **编排:** Kubernetes - 容器编排

## Casbin模型

服务使用RBAC模型进行权限控制：

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

详细说明请参考 [API.md#casbin模型](API.md#casbin模型)。

## 与Controller集成

Controller服务可以使用生成的Golang客户端与RBAC服务集成：

```go
import (
    rbac "github.com/davidliyutong/idekube-rbac/client"
)

// 创建客户端
client, err := rbac.NewClient("http://rbac-service:8080")

// 检查权限
allowed, err := client.CheckPermission(ctx, &rbac.CheckPermissionRequest{
    UserID:       userID,
    ResourceType: "workspace",
    ResourceID:   workspaceID,
    Action:       "delete",
})
```

详细集成指南请参考 [API.md#集成指南](API.md#集成指南)。

## 监控和观测

### 健康检查

```bash
curl http://localhost:8080/healthz
```

### 日志

服务输出结构化日志，支持不同日志级别：

```bash
# 启用debug日志
export LOG_LEVEL=debug
```

### Metrics

建议集成Prometheus进行指标收集：

- HTTP请求延迟
- 权限检查成功/失败率
- 数据库连接池状态
- RabbitMQ队列深度

## 贡献

欢迎贡献！请遵循以下步骤：

1. Fork项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启Pull Request

## 许可证

本项目使用Apache 2.0许可证 - 详见LICENSE文件

## 支持

- 📧 Email: support@idekube.io
- 🐛 Issues: [GitHub Issues](https://github.com/davidliyutong/idekube/issues)
- 📖 Wiki: [项目Wiki](https://github.com/davidliyutong/idekube/wiki)

## 相关项目

- [idekube-controller](../controller/) - IDEKube控制器服务
- [idekube-housekeeper](../housekeeper/) - IDEKube清理服务
- [idekube-frontend](../frontend/) - IDEKube前端应用
