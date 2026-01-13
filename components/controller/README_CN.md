# IDEKube Controller

IDEKube Controller是云IDE平台的核心API服务器，提供工作区、模板、用户和组织的管理功能。

## 功能特性

- 🔐 **用户认证** - JWT令牌认证、用户注册与登录
- 👥 **组织管理** - 多租户组织、成员角色管理
- 📦 **模板系统** - 预定义工作区模板、自定义资源配置
- 💻 **工作区管理** - 创建、启动、停止、删除工作区实例
- 💾 **存储卷管理** - 持久化存储、动态卷挂载
- 🔄 **事件驱动** - RabbitMQ消息队列集成
- 📊 **资源配额** - CPU、内存、存储配额管理
- 🔍 **审计日志** - 完整的操作审计追踪
- 📖 **Swagger文档** - 交互式API文档

## 快速开始

### 前置条件

- Go 1.21+
- PostgreSQL 14+
- Kubernetes 1.24+ (可选)

### 本地运行

```bash
# 1. 安装依赖
go mod download

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 配置数据库等参数

# 3. 启动服务 (会自动运行数据库迁移)
make run
```

服务启动后访问：
- API: http://localhost:8080/api/v1
- Swagger UI: http://localhost:8080/swagger/index.html
- 健康检查: http://localhost:8080/health

> 📖 详细部署步骤请参考 [快速开始指南](./QUICKSTART.md)

## 文档导航

| 文档 | 描述 |
|------|------|
| [QUICKSTART.md](./QUICKSTART.md) | 快速开始 - 安装、配置、部署指南 |
| [API.md](./API.md) | API定义 - 完整的接口规范和数据模型 |
| [API_TESTING.md](./API_TESTING.md) | API测试 - 详细的测试示例和用例 |
| [DESIGN.md](./DESIGN.md) | 设计文档 - 架构设计和技术方案 |
| [Swagger UI](http://localhost:8080/swagger/index.html) | 交互式API文档 (需先启动服务) |

## API概览

### 认证
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录

### 用户管理
- `GET /api/v1/users/me` - 获取当前用户
- `POST /api/v1/users/me/password` - 修改密码
- `GET /api/v1/users` - 列出用户 (Admin)

### 组织管理
- `POST /api/v1/organizations` - 创建组织
- `GET /api/v1/organizations` - 列出组织
- `POST /api/v1/organizations/:id/members` - 添加成员

### 模板管理
- `POST /api/v1/templates` - 创建模板
- `GET /api/v1/templates` - 列出模板
- `GET /api/v1/templates/:id` - 获取模板详情

### 工作区管理
- `POST /api/v1/workspaces` - 创建工作区
- `GET /api/v1/workspaces` - 列出工作区
- `POST /api/v1/workspaces/:id/start` - 启动工作区
- `POST /api/v1/workspaces/:id/stop` - 停止工作区

### 存储卷管理
- `POST /api/v1/volumes` - 创建存储卷
- `GET /api/v1/volumes` - 列出存储卷
- `DELETE /api/v1/volumes/:id` - 删除存储卷

> 完整API列表请查看 [API文档](./API.md) 或 [Swagger UI](http://localhost:8080/swagger/index.html)

## 配置说明

通过环境变量配置服务：

### 基本配置

```bash
SERVER_ADDRESS=:8080                    # 服务监听地址
JWT_SECRET=your-secret-key              # JWT密钥
JWT_EXPIRATION_HOURS=720                # Token过期时间(小时)
```

### PostgreSQL配置

```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=idekube
POSTGRES_PASSWORD=idekube123
POSTGRES_DB=idekube
```

### Kubernetes配置 (可选)

```bash
KUBECONFIG=~/.kube/config               # kubeconfig路径
NAMESPACE=idekube                       # 工作区命名空间
```

### RabbitMQ配置 (可选)

```bash
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VHOST=/
```

## 开发

### Makefile命令

```bash
# 构建
make build              # 编译应用
make build-migrate      # 编译迁移工具

# 运行
make run                # 运行服务
make dev                # 热重载模式 (需要air)

# 测试
make test               # 运行测试
make test-coverage      # 生成覆盖率报告

# 代码质量
make fmt                # 格式化代码
make lint               # 代码检查

# Swagger文档
make swagger-gen        # 生成Swagger文档
make swagger-js-client  # 生成JavaScript客户端
make swagger-ts-client  # 生成TypeScript客户端

# 数据库迁移
make migrate-up         # 应用迁移
make migrate-down       # 回滚迁移
make migrate-version    # 查看版本

# Docker
make docker-build       # 构建镜像
make docker-push        # 推送镜像

# 清理
make clean              # 清理构建产物
```

### 生成Swagger文档

```bash
# 1. 安装swag工具
go install github.com/swaggo/swag/cmd/swag@latest

# 2. 生成文档
make swagger-gen

# 3. 重启服务查看
make run
# 访问 http://localhost:8080/swagger/index.html
```

### 生成API客户端

```bash
# JavaScript客户端 (需要swagger-codegen)
brew install swagger-codegen
make swagger-js-client

# TypeScript客户端 (需要openapi-generator)
npm install -g @openapitools/openapi-generator-cli
make swagger-ts-client
```

## 部署

### Docker

```bash
# 构建镜像
make docker-build

# 使用Docker Compose
docker-compose up -d
```

### Kubernetes

```bash
# 使用Helm
helm install idekube-controller ../../helm

# 使用kubectl
kubectl apply -f manifests/
```

### 数据库迁移

数据库迁移在服务启动时**自动执行**，无需手动运行。如需手动控制：

```bash
# 查看当前版本
make migrate-version

# 应用迁移
make migrate-up

# 回滚最后一次迁移
make migrate-down
```

## 架构

```
┌─────────────────┐
│   Frontend UI   │
└────────┬────────┘
         │ HTTP/REST
         ▼
┌─────────────────┐      ┌──────────────┐
│   Controller    │◄─────┤  PostgreSQL  │
│   (API Server)  │      └──────────────┘
└────────┬────────┘
         │ Events
         ▼
┌─────────────────┐      ┌──────────────┐
│   RabbitMQ      │─────►│ Housekeeper  │
└─────────────────┘      └──────┬───────┘
                                 │
                                 ▼
                         ┌──────────────┐
                         │  Kubernetes  │
                         └──────────────┘
```

- **Controller**: REST API服务器，处理用户请求
- **PostgreSQL**: 存储用户、组织、模板、工作区等数据
- **RabbitMQ**: 消息队列，异步事件处理
- **Housekeeper**: 后台服务，管理Kubernetes资源

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin (HTTP Router)
- **数据库**: PostgreSQL + GORM
- **消息队列**: RabbitMQ
- **容器编排**: Kubernetes
- **认证**: JWT
- **文档**: Swagger/OpenAPI

## 贡献

欢迎贡献！请查看 [贡献指南](../../CONTRIBUTING.md)。

## 许可证

Apache 2.0 - 详见 [LICENSE](../../LICENSE) 文件。

## 链接

- 项目主页: https://github.com/davidliyutong/idekube
- 问题跟踪: https://github.com/davidliyutong/idekube/issues
- 文档: https://github.com/davidliyutong/idekube/wiki
