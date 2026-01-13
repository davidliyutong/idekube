# idekube-housekeeper

idekube-housekeeper 是 idekube 平台的清理和维护组件，负责管理 Kubernetes 资源的生命周期，清理过期资源，归档数据，以及执行定期维护任务。

## 概述

idekube-housekeeper 通过监听 RabbitMQ 消息队列接收清理请求，与 PostgreSQL 数据库和 Kubernetes API 交互，执行以下任务：

- **资源清理**：删除过期的 Kubernetes 工作空间资源（Namespace、Deployment、Service、PVC 等）
- **数据归档**：将工作空间数据归档到持久化存储
- **定期维护**：执行数据库优化、镜像清理等定期维护任务
- **状态同步**：保持数据库记录与 Kubernetes 资源状态一致

## 核心功能

### 1. 工作空间清理
- 接收来自 RabbitMQ 的清理请求
- 删除 Kubernetes 中的工作空间资源
- 更新数据库中的工作空间状态
- 支持强制清理和安全清理模式

### 2. 数据归档
- 归档工作空间数据到 S3/NFS 等存储
- 压缩和打包工作空间文件
- 设置数据保留期限
- 支持增量备份和全量备份

### 3. 定期维护
- 清理超时未访问的工作空间
- 优化 PostgreSQL 数据库
- 清理未使用的容器镜像
- 检测和修复资源不一致问题

### 4. 监控和日志
- 详细的操作日志
- 清理任务统计和报告
- 错误追踪和告警
- 性能指标收集

## 架构说明

```
┌─────────────┐
│   RabbitMQ  │  ← 接收清理请求
└──────┬──────┘
       │
       ↓
┌─────────────────┐
│  Housekeeper    │
│                 │
│  ┌───────────┐  │
│  │  消息处理  │  │
│  └───────────┘  │
│        │        │
│        ↓        │
│  ┌───────────┐  │
│  │  资源清理  │  │
│  └───────────┘  │
│        │        │
│        ↓        │
│  ┌───────────┐  │
│  │  状态更新  │  │
│  └───────────┘  │
└────┬─────┬──────┘
     │     │
     ↓     ↓
┌─────────┐ ┌──────────┐
│Kubernetes│ │PostgreSQL│
└─────────┘ └──────────┘
```

## 快速开始

### 最简单的方式 - 使用 Docker Compose

```bash
# 克隆项目
git clone https://github.com/davidliyutong/idekube.git
cd idekube/components/housekeeper

# 配置环境变量（编辑 .env 文件）
cp .env.example .env

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f housekeeper
```

### 本地开发

```bash
# 安装依赖
go mod download

# 构建
make build

# 运行（需要先启动 PostgreSQL 和 RabbitMQ）
make run
```

### Kubernetes 部署

```bash
# 使用 Helm Chart（推荐）
helm upgrade --install idekube ../../helm \
  --namespace idekube-system \
  --create-namespace

# 或使用 kubectl
kubectl apply -f k8s/
```

详细的部署指南请参考 [快速开始文档](./QUICKSTART.md)。

## 文档

### 📚 完整文档导航

- **[快速开始指南 (QUICKSTART.md)](./QUICKSTART.md)** - 部署和配置指南
  - 系统要求
  - 本地开发环境搭建
  - Docker 部署
  - Kubernetes 部署
  - 配置说明
  - 故障排查

- **[API 定义文档 (API.md)](./API.md)** - RabbitMQ 消息格式和数据库模型
  - RabbitMQ 连接配置
  - 队列定义和消息格式
  - 数据库模型说明
  - 消息发送示例（Python、Go、cURL）
  - 监控和指标

- **[API 测试指南 (API_TESTING.md)](./API_TESTING.md)** - 功能测试方法
  - 测试环境准备
  - 功能测试用例
  - 集成测试
  - 性能测试
  - 故障恢复测试
  - 常见问题排查

## 配置概览

通过环境变量进行配置：

| 类别 | 环境变量 | 默认值 | 说明 |
|------|---------|--------|------|
| **Kubernetes** | `KUBECONFIG` | - | kubeconfig 文件路径 |
| | `NAMESPACE` | 全部 | 监视的命名空间 |
| **PostgreSQL** | `POSTGRES_HOST` | localhost | 数据库主机 |
| | `POSTGRES_PORT` | 5432 | 数据库端口 |
| | `POSTGRES_USER` | - | 用户名（必填） |
| | `POSTGRES_PASSWORD` | - | 密码（必填） |
| | `POSTGRES_DB` | - | 数据库名（必填） |
| **RabbitMQ** | `RABBITMQ_HOST` | localhost | 消息队列主机 |
| | `RABBITMQ_PORT` | 5672 | 消息队列端口 |
| | `RABBITMQ_USER` | - | 用户名（必填） |
| | `RABBITMQ_PASSWORD` | - | 密码（必填） |
| | `RABBITMQ_VHOST` | / | 虚拟主机 |
| **应用** | `LOG_LEVEL` | info | 日志级别 |
| | `WORKER_THREADS` | 1 | 工作线程数 |

详细配置说明请参考 [快速开始指南](./QUICKSTART.md#配置说明)。

## 开发

### 前置条件
- Go 1.21+
- Docker（用于构建镜像）
- Kubernetes 集群（用于测试）
- PostgreSQL 12+
- RabbitMQ 3.8+

### 构建

```bash
# 构建二进制文件
make build

# 构建 Docker 镜像
make docker-build

# 运行测试
make test

# 清理构建产物
make clean
```

### 项目结构

```
.
├── cmd/
│   └── housekeeper/        # 主程序入口
├── internal/
│   ├── config/            # 配置管理
│   ├── housekeeper/       # 核心业务逻辑
│   └── models/            # 数据模型
├── pkg/
│   ├── database/          # 数据库客户端
│   ├── k8s/               # Kubernetes 客户端
│   ├── logger/            # 日志工具
│   └── queue/             # RabbitMQ 客户端
├── bin/                   # 编译输出目录
├── Dockerfile            # Docker 镜像构建文件
├── Makefile              # 构建脚本
└── README.md             # 本文件
```

### 开发工作流

1. **Fork 并克隆代码库**
2. **创建功能分支**：`git checkout -b feature/your-feature`
3. **编写代码和测试**
4. **运行测试**：`make test`
5. **提交更改**：`git commit -am 'Add some feature'`
6. **推送分支**：`git push origin feature/your-feature`
7. **创建 Pull Request**

## 技术栈

- **语言**：Go 1.21+
- **消息队列**：RabbitMQ (AMQP 0-9-1)
- **数据库**：PostgreSQL 12+
- **容器编排**：Kubernetes 1.20+
- **依赖管理**：Go Modules

### 主要依赖

- `k8s.io/client-go` - Kubernetes 客户端
- `github.com/rabbitmq/amqp091-go` - RabbitMQ 客户端
- `gorm.io/gorm` - ORM 框架
- `github.com/lib/pq` - PostgreSQL 驱动

## 贡献

欢迎贡献代码、报告问题或提出改进建议！

### 如何贡献

1. 查看 [GitHub Issues](https://github.com/davidliyutong/idekube/issues)
2. Fork 项目
3. 创建功能分支
4. 提交 Pull Request

### 代码规范

- 遵循 Go 代码规范和最佳实践
- 为新功能添加测试
- 更新相关文档
- 提交消息使用清晰的描述

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](../../LICENSE) 文件

## 相关链接

- **主项目**：[idekube](https://github.com/davidliyutong/idekube)
- **Controller 组件**：[../controller](../controller)
- **Frontend 组件**：[../frontend](../frontend)
- **RBAC 组件**：[../rbac](../rbac)

## 获取帮助

- **GitHub Issues**：[提交问题](https://github.com/davidliyutong/idekube/issues)
- **文档**：查看上方的文档导航
- **示例**：参考 [API 测试指南](./API_TESTING.md)

---

**维护者**：idekube 团队  
**最后更新**：2026-01-13
