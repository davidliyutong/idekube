# Casbin配置文件

这个目录包含Casbin RBAC的配置文件。

## 文件说明

### model.conf

Casbin模型定义文件，定义了RBAC的规则结构：

- `[request_definition]`: 定义请求的格式 (subject, object, action)
- `[policy_definition]`: 定义策略的格式
- `[role_definition]`: 定义角色继承关系
  - `g`: 用户-角色关系
  - `g2`: 资源-资源组关系（可选）
- `[policy_effect]`: 定义策略效果（允许或拒绝）
- `[matchers]`: 定义匹配规则
  - 支持通配符 `*`
  - 支持 `keyMatch2` 进行路径匹配

**注意：** 此文件通常不需要修改，除非要改变权限模型的基本结构。

### policy.csv

Casbin策略文件，定义具体的权限规则和角色分配：

格式：
```csv
# 策略规则
p, subject, object, action

# 角色分配
g, user, role
```

示例：
```csv
# 管理员角色可以对workspace进行所有操作
p, role:admin, workspace, read
p, role:admin, workspace, write
p, role:admin, workspace, delete

# 用户1被分配为管理员
g, user:1, role:admin
```

## 使用方式

### 数据库模式policy.csv`加载初始策略。

### 文件模式

如果需要始终从文件加载策略（例如在测试环境），可以：

1. 修改`

如果需要始终从文件加载策略（例如在测试环境），可以：

1. 修改`policy.csv`
2. 清空数据库中的策略表
3. 重启服务

## 修改策略

### 通过API修改

推荐使用API动态修改策略：

```bash
# 为用户分配角色
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123, "role": "admin"}'
```

### 通过文件修改

1. 编辑`policy.csv`
2. 清空数据库策略表：
   ```sql
   DELETE FROM casbin_rule;
   ```
3. 重启服务

## 预定义角色

### admin（管理员）
- 所有资源的完全访问权限
- 可以管理用户和组织

### editor（编辑者）
- workspace: read, write, execute
- template: read, write
- volume: read, write

### viewer（查看者）
- workspace: read
- template: read
- volume: read

### workspace-admin（工作空间管理员）
- workspace的所有权限
- 其他资源的有限权限

### template-admin（模板管理员）
- template的所有权限
- 其他资源的有限权限

## 资源类型

- `workspace`: 工作空间
- `template`: 模板
- `volume`: 持久化卷
- `organization`: 组织
- `user`: 用户

## 操作类型

- `read`: 读取/查看
- `write`: 写入/更新
- `delete`: 删除
- `execute`: 执行
- `create`: 创建
- `manage`: 管理（包括所有操作）

## 示例场景

### 场景1：为新用户分配viewer角色

添加到`policy.csv`：
```csv
g, user:100, role:viewer
```

或通过API：
```bash
curl -X POST http://localhost:8080/api/v1/rbac/assign-role \
  -H "Content-Type: application/json" \
  -d '{"user_id": 100, "role": "viewer"}'
```

### 场景2：创建自定义角色

在`policy.csv`中定义新角色：
```csv
# 定义data-scientist角色的权限
p, role:data-scientist, workspace, read
p, role:data-scientist, workspace, execute
p, role:data-scientist, template, read
p, role:data-scientist, volume, read
p, role:data-scientist, volume, write

# 分配角色
g, user:200, role:data-scientist
```

### 场景3：特定资源的权限

给予用户对特定workspace的权限：
```csv
# 用户201可以访问workspace ws-001
p, user:201, workspace:ws-001, read
p, user:201, workspace:ws-001, write
```

## 最佳实践

1. **使用角色而非直接分配权限**：通过角色管理权限更易维护
2. **遵循最小权限原则**：只授予必要的权限
3. **使用角色继承**：利用Casbin的角色继承特性简化管理
4. **定期审计**：定期检查和更新权限策略
5. **测试策略**：在生产环境应用前充分测试权限策略

## 调试

### 检查策略是否生效

```bash
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "resource_type": "workspace",
    "action": "read"
  }'
```

### 查看数据库中的策略

```sql
SELECT * FROM casbin_rule;
```

## 参考资源

- [Casbin官方文档](https://casbin.org/docs/overview)
- [RBAC模型](https://casbin.org/docs/rbac)
- [策略语法](https://casbin.org/docs/syntax-for-models)
