# IDEKube Controller 快速开始指南

本指南帮助你快速部署和运行IDEKube Controller服务。

> 📖 **相关文档**
> - [API定义文档](./API.md) - 完整的API规范
> - [API测试指南](./API_TESTING.md) - API测试示例
> - [Swagger UI](http://localhost:8080/swagger/index.html) - 交互式API文档

## 系统要求

### 必需组件

- **Go**: 1.21 或更高版本
- **PostgreSQL**: 14 或更高版本
- **Kubernetes**: 1.24+ (可选，用于完整工作区功能)

### 可选组件

- **RabbitMQ**: 3.12+ (用于异步事件处理)
- **Docker**: 用于容器化部署

## 方式一：本地开发运行

### 1. 安装PostgreSQL

#### macOS
```bash
brew install postgresql@14
brew services start postgresql@14
```

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install postgresql-14
sudo systemctl start postgresql
```

### 2. 创建数据库

```bash
# 切换到postgres用户
sudo -u postgres psql

# 在psql中执行
CREATE DATABASE idekube;
CREATE USER idekube WITH PASSWORD 'idekube123';
GRANT ALL PRIVILEGES ON DATABASE idekube TO idekube;
\q
```

或使用命令行：
```bash
createdb idekube
createuser -P idekube  # 输入密码: idekube123
psql -c "GRANT ALL PRIVILEGES ON DATABASE idekube TO idekube;"
```

> ⚠️ **注意**: 生产环境请使用强密码！

### 3. 配置环境变量

```bash
cd components/controller
cp .env.example .env
```

编辑 `.env` 文件：

```bash
# 服务器配置
SERVER_ADDRESS=:8080

# JWT配置
JWT_SECRET=your-super-secret-key-change-this-in-production
JWT_EXPIRATION_HOURS=720

# PostgreSQL配置
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=idekube
POSTGRES_PASSWORD=idekube123
POSTGRES_DB=idekube

# Kubernetes配置 (可选)
KUBECONFIG=~/.kube/config
NAMESPACE=idekube

# RabbitMQ配置 (可选)
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
```

### 4. 安装依赖

```bash
go mod download
```

### 5. 运行服务

```bash
# 使用 make (推荐)
make run

# 或直接使用 go run
go run cmd/controller/main.go
```

服务启动后会：
1. 连接PostgreSQL数据库
2. **自动运行数据库迁移** (无需手动执行SQL)
3. 启动HTTP服务器: http://localhost:8080

### 6. 验证服务

**健康检查:**
```bash
curl http://localhost:8080/health
```

**查看Swagger文档:**

浏览器访问: http://localhost:8080/swagger/index.html

**注册测试用户:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "admin123",
    "full_name": "Admin User"
  }'
```

## 方式二：Docker运行

### 1. 构建镜像

```bash
make docker-build
```

### 2. 使用Docker运行

**单独运行controller (需要外部PostgreSQL):**
```bash
docker run --rm \
  -e POSTGRES_HOST=host.docker.internal \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=idekube \
  -e POSTGRES_PASSWORD=idekube123 \
  -e POSTGRES_DB=idekube \
  -e JWT_SECRET=your-secret-key \
  -p 8080:8080 \
  davidliyutong/idekube-controller:latest
```

**使用Docker Compose (推荐):**

创建 `docker-compose.yml`:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: idekube
      POSTGRES_USER: idekube
      POSTGRES_PASSWORD: idekube123
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U idekube"]
      interval: 5s
      timeout: 5s
      retries: 5

  controller:
    image: davidliyutong/idekube-controller:latest
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      SERVER_ADDRESS: ":8080"
      JWT_SECRET: "your-secret-key-change-in-production"
      POSTGRES_HOST: postgres
      POSTGRES_PORT: 5432
      POSTGRES_USER: idekube
      POSTGRES_PASSWORD: idekube123
      POSTGRES_DB: idekube
    ports:
      - "8080:8080"

volumes:
  postgres-data:
```

启动服务：
```bash
docker-compose up -d
```

## 方式三：Kubernetes部署

### 使用Helm Chart

```bash
helm install idekube-controller ../../helm \
  --set controller.image.tag=latest \
  --set postgresql.enabled=true
```

### 使用Kubectl

```bash
kubectl apply -f manifests/
kubectl get pods -n idekube
```

## 生成Swagger文档

```bash
# 生成Swagger文档
make swagger-gen

# 生成JavaScript客户端
make swagger-js-client

# 生成TypeScript客户端
make swagger-ts-client
```

## 开发工具

### 热重载开发

```bash
go install github.com/cosmtrek/air@latest
make dev
```

### 代码质量

```bash
make fmt          # 格式化代码
make lint         # 代码检查
make test         # 运行测试
make test-coverage # 测试覆盖率
```

## 常见问题

### PostgreSQL连接失败

检查PostgreSQL是否运行: `pg_isready`

Docker用户使用 `host.docker.internal` 连接宿主机。

### 端口已被占用

```bash
lsof -i :8080  # 查找占用进程
export SERVER_ADDRESS=:8081  # 或更换端口
```

### Swagger文档未生成

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swagger-gen
```

## 下一步

- 📖 [API定义文档](./API.md) - 完整API规范
- 🧪 [API测试指南](./API_TESTING.md) - 测试示例
- 🔍 [Swagger UI](http://localhost:8080/swagger/index.html) - 交互式文档
### Phase 1 Implementation (TODO)

The following features are designed but need handlers and services to be implemented:

1. **Organization Management** - Create and manage organizations
2. **Template Management** - Define workspace templates
3. **Workspace Management** - Create and orchestrate workspaces in K8s
4. **Volume Management** - Manage persistent volumes

Refer to [DESIGN.md](DESIGN.md) for detailed API specifications and implementation guidance.

### Development Workflow

1. Implement service layer in `internal/services/`
2. Implement handler layer in `internal/handlers/`
3. Register routes in `internal/api/server.go`
4. Add K8s orchestration logic in `pkg/k8s/`
5. Test the endpoints

## Troubleshooting

### Database connection failed
- Check PostgreSQL is running: `pg_isready`
- Verify credentials in `.env`
- Check database exists: `psql -l`

### JWT token invalid
- Check `JWT_SECRET` is set correctly
- Verify token hasn't expired (default 24 hours)
- Ensure Bearer token format: `Authorization: Bearer <token>`

### Permission denied
- Check user role (admin/super_admin required for some endpoints)
- Verify JWT token is valid and not expired

## Docker Deployment

```bash
# Build image
make docker-build

# Run container
docker run -d \
  --name idekube-controller \
  -p 8080:8080 \
  --env-file .env \
  idekube/controller:latest
```

## Production Deployment Checklist

- [ ] Change default admin password
- [ ] Set strong JWT_SECRET (32+ random characters)
- [ ] Use environment-specific database credentials
- [ ] Enable HTTPS/TLS
- [ ] Configure proper CORS origins
- [ ] Set up database backups
- [ ] Configure log aggregation
- [ ] Set up monitoring and alerts
- [ ] Review and adjust quotas
- [ ] Enable audit log retention policies

## Support

For issues and questions:
- Check [DESIGN.md](DESIGN.md) for architecture details
- Review API documentation in code comments
- Check logs for error messages
