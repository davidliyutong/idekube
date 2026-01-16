# IDEKube Controller 重构计划

## 1. 问题分析

### 1.1 缺失的API功能

参考 `references/controller/docs/api.md`（Pod等同于Workspace），当前controller实现中缺失以下关键功能：

#### 1.1.1 用户管理（User Management）
- **已有功能：**
  - ✅ GET /api/v1/users/me - 获取当前用户信息
  - ✅ POST /api/v1/users/me/password - 修改密码
  - ✅ GET /api/v1/users/:id - 获取用户信息（仅限管理员）
  - ✅ GET /api/v1/users - 列出所有用户（仅限管理员）
  - ✅ PUT /api/v1/users/:id - 更新用户（仅限管理员）
  - ✅ DELETE /api/v1/users/:id - 删除用户（仅限管理员）

- **缺失功能：**
  - ❌ POST /api/v1/users - 管理员创建用户（当前只有register）
  - ❌ PUT /api/v1/users/me - 用户编辑自己的profile（邮箱、头像、昵称）
  - ❌ GET /api/v1/users/check - power_user查询用户是否存在
  - ❌ 字段级别的权限控制（用户只能修改自己的profile字段，不能修改role等）
  - ❌ 基于RBAC的细粒度权限控制（当前使用简单的角色中间件）

#### 1.1.2 工作区管理（Workspace Management）
- **已有功能：**
  - ✅ POST /api/v1/workspaces - 创建工作区
  - ✅ GET /api/v1/workspaces/:id - 获取工作区详情
  - ✅ PUT /api/v1/workspaces/:id - 更新工作区
  - ✅ DELETE /api/v1/workspaces/:id - 删除工作区
  - ✅ GET /api/v1/workspaces - 列出工作区（需要owner_type和owner_id参数）

- **缺失功能：**
  - ❌ 普通用户列出自己的工作区（无需指定owner参数）
  - ❌ System Admin列出所有用户的工作区
  - ❌ Organization Owner/Admin列出组织内的所有工作区
  - ❌ Organization Member列出自己的工作区和组织内共享的工作区
  - ❌ 工作区共享功能（is_shared字段）
  - ❌ 基于RBAC的访问控制（而非基于owner_type/owner_id）

#### 1.1.3 模板管理（Template Management）
- **已有功能：**
  - ✅ GET /api/v1/templates - 列出模板
  - ✅ POST /api/v1/templates - 创建模板
  - ✅ GET /api/v1/templates/:id - 获取模板详情
  - ✅ PUT /api/v1/templates/:id - 更新模板
  - ✅ DELETE /api/v1/templates/:id - 删除模板

- **缺失功能：**
  - ❌ System Admin列出所有模板（包括私有模板）
  - ❌ Organization Owner列出组织内的模板
  - ❌ 普通用户只能看到公开模板和自己可访问的模板
  - ❌ 基于RBAC的细粒度权限控制

#### 1.1.4 存储卷管理（Volume Management）
- **已有功能：**
  - ✅ POST /api/v1/volumes - 创建存储卷
  - ✅ GET /api/v1/volumes/:id - 获取存储卷详情
  - ✅ PUT /api/v1/volumes/:id - 更新存储卷
  - ✅ DELETE /api/v1/volumes/:id - 删除存储卷
  - ✅ GET /api/v1/volumes - 列出存储卷（需要owner参数）

- **缺失功能：**
  - ❌ 普通用户列出自己的存储卷
  - ❌ System Admin列出所有存储卷
  - ❌ Organization Owner列出组织内的存储卷
  - ❌ 基于RBAC的访问控制

#### 1.1.5 组织管理（Organization Management）
- **已有功能：**
  - ✅ POST /api/v1/organizations - 创建组织
  - ✅ GET /api/v1/organizations - 列出当前用户的组织
  - ✅ GET /api/v1/organizations/:id - 获取组织详情
  - ✅ PUT /api/v1/organizations/:id - 更新组织
  - ✅ DELETE /api/v1/organizations/:id - 删除组织
  - ✅ POST /api/v1/organizations/:id/members - 添加成员
  - ✅ DELETE /api/v1/organizations/:id/members/:user_id - 移除成员
  - ✅ PUT /api/v1/organizations/:id/members/:user_id - 更新成员角色

- **缺失功能：**
  - ❌ System Admin列出所有组织
  - ❌ 基于RBAC的细粒度权限控制

### 1.2 架构问题

#### 1.2.1 当前权限控制方式
当前实现使用简单的角色中间件（`RequireRole`）和owner检查：
- 通过JWT中的`user_role`判断是否为admin
- 通过`owner_type`和`owner_id`过滤资源
- **问题：** 不够灵活，无法支持复杂的权限场景

#### 1.2.2 缺少RBAC集成
- 当前没有与RBAC服务集成
- 无法实现细粒度的资源级权限控制
- 无法动态管理权限策略

#### 1.2.3 资源标签缺失
- 当前资源（Workspace、Volume等）没有标签机制
- 无法基于标签进行权限控制
- 难以实现组织级别的资源管理

## 2. 重构目标

### 2.1 统一API路径，使用RBAC进行权限控制
- ❌ **不使用** `/admin` 分层路由
- ✅ **使用** 统一路径，通过RBAC中间件控制访问
- 示例：
  - `GET /api/v1/workspaces` - 根据用户权限返回不同范围的工作区
  - `GET /api/v1/users` - 根据用户权限返回不同范围的用户

### 2.2 实现RBAC集成
- 与独立的RBAC服务集成
- 为资源添加标签和元数据
- 实现基于标签的权限控制

### 2.3 完善API功能
- 实现所有缺失的API端点
- 支持不同角色的用户访问相应范围的资源

### 2.4 权限层级
```
super_admin (系统超级管理员)
  └─ 可以访问所有资源
  └─ 可以管理所有用户、组织、模板、工作区、存储卷
  └─ 可以修改用户角色和权限
  └─ 可以删除任何用户

admin (系统管理员)
  └─ 可以访问所有用户资源
  └─ 可以管理用户（除了修改角色和删除用户）
  └─ 可以查看所有工作区和存储卷

power_user (高级用户)
  └─ 可以创建组织
  └─ 可以查询用户是否存在（用于邀请成员）
  └─ 可以管理自己的资源
  └─ 可以使用公开模板

organization:owner (组织所有者)
  └─ 拥有组织的所有权限
  └─ 可以管理组织信息
  └─ 可以添加/移除组织成员
  └─ 可以指派/解除组织管理员（organization:admin）
  └─ 可以搜索用户并添加到组织（而不是列出所有用户）
  └─ 可以查看和管理组织内的所有资源（模板、工作区、存储卷）
  └─ 可以创建和管理组织级别的模板
  └─ 可以查看组织内所有工作区（包括私有和共享的）
  └─ 可以删除组织内任何成员的工作区和存储卷

organization:admin (组织管理员)
  └─ 可以管理组织的普通成员（不能修改organization:owner和其他admin）
  └─ 可以添加/移除普通成员
  └─ 可以查看和管理组织内的所有资源
  └─ 可以查看组织内所有工作区（包括私有和共享的）
  └─ 可以管理组织级别的模板
  └─ 不能指派/解除管理员
  └─ 不能修改organization:owner

organization:member (组织成员)
  └─ 可以使用组织的模板
  └─ 只能查看和管理自己创建的工作区
  └─ 可以查看组织内标记为共享的工作区
  └─ 可以查看和管理自己的存储卷

user (普通用户)
  └─ 只能管理自己的资源
  └─ 只能使用公开模板
  └─ 不能创建组织
```

## 3. 实现计划

### 3.1 阶段一：数据模型增强

#### 3.1.1 为资源添加标签系统
```go
// 在 internal/models/common.go 中添加
type ResourceLabels map[string]string

// 为每个资源添加标签字段
type Workspace struct {
    // ... 现有字段
    Labels         ResourceLabels `json:"labels,omitempty" db:"labels"`
    IsShared       bool           `json:"is_shared" db:"is_shared"`              // 是否为组织内共享工作区
    OrganizationID *int64         `json:"organization_id,omitempty" db:"organization_id"` // 所属组织ID（如果属于组织）
}

type Volume struct {
    // ... 现有字段
    Labels ResourceLabels `json:"labels,omitempty" db:"labels"`
}

type Template struct {
    // ... 现有字段
    Labels ResourceLabels `json:"labels,omitempty" db:"labels"`
    Visibility TemplateVisibility `json:"visibility" db:"visibility"` // public, organization, private
}
```

#### 3.1.2 数据库迁移
直接修改现有的 `migrations/000001_init_schema.up.sql` 和 `migrations/000001_init_schema.down.sql`：

**需要添加的字段：**
1. workspaces 表：
   - `labels` JSONB 列，用于存储资源标签
   - `is_shared` BOOLEAN 列，默认 false，标识工作区是否在组织内共享
   - `organization_id` BIGINT 列，可为空，引用 organizations(id)，表示所属组织

2. volumes 表：
   - `labels` JSONB 列，用于存储资源标签

3. templates 表：
   - `labels` JSONB 列，用于存储资源标签
   - `visibility` VARCHAR(20) 列，默认 'private'，可选值：'public', 'organization', 'private'

4. users 表（如果需要）：
   - 确保有 `display_name`, `avatar_url`, `email` 字段

**初始标签规则：**
- 创建资源时自动添加：
  - `owner_type`: user/organization
  - `owner_id`: 所有者ID
  - `organization_id`: 所属组织ID（如果适用）

### 3.2 阶段二：RBAC集成

#### 3.2.1 创建RBAC客户端
```go
// 在 pkg/rbac/ 中创建
package rbac

type Client struct {
    baseURL string
    httpClient *http.Client
}

type CheckPermissionRequest struct {
    UserID       int64  `json:"user_id"`
    ResourceType string `json:"resource_type"` // workspace, volume, template, user, organization
    ResourceID   string `json:"resource_id"`
    Action       string `json:"action"` // create, read, update, delete, list
}

func (c *Client) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (bool, error)
func (c *Client) AssignRoleToUser(ctx context.Context, userID int64, role string) error
func (c *Client) AssignRoleForResource(ctx context.Context, userID int64, resourceType, resourceID, role string) error
```

#### 3.2.2 创建RBAC中间件
```go
// 在 internal/middleware/rbac.go 中创建
package middleware

func RBACMiddleware(rbacClient *rbac.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetInt64("user_id")
        resourceType := inferResourceType(c.Request.URL.Path)
        action := inferAction(c.Request.Method)
        resourceID := c.Param("id")
        
        allowed, err := rbacClient.CheckPermission(c.Request.Context(), &rbac.CheckPermissionRequest{
            UserID:       userID,
            ResourceType: resourceType,
            ResourceID:   resourceID,
            Action:       action,
        })
        
        if err != nil || !allowed {
            c.JSON(http.StatusForbidden, models.APIResponse{
                Success: false,
                Error: &models.APIError{
                    Code:    "PERMISSION_DENIED",
                    Message: "You don't have permission to perform this action",
                },
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

#### 3.2.3 更新RBAC服务策略模型
在 `components/rbac/configs/` 中定义新的策略：

**model.conf:**
```conf
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _
g2 = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && g2(r.obj, p.obj) && r.act == p.act
```

**policy.csv (示例):**
```csv
# 系统级角色
p, role:super_admin, *, *

# 管理员角色
p, role:admin, user, list
p, role:admin, user, read
p, role:admin, user, create
p, role:admin, user, update
p, role:admin, workspace, list
p, role:admin, workspace, read
p, role:admin, volume, list
p, role:admin, volume, read
p, role:admin, template, list
p, role:admin, template, read
p, role:admin, organization, list
p, role:admin, organization, read

# Power User角色
p, role:power_user, organization, create
p, role:power_user, user, check
p, role:power_user, workspace:own, *
p, role:power_user, volume:own, *
p, role:power_user, template:public, read

# 组织级角色
p, role:org:owner, organization, manage
p, role:org:owner, organization, update
p, role:org:owner, organization, delete
p, role:org:owner, organization, manage_members
p, role:org:owner, organization, manage_admins
p, role:org:owner, user, search
p, role:org:owner, template:org, *
p, role:org:owner, workspace:org, *
p, role:org:owner, volume:org, *

p, role:org:admin, organization, manage_members
p, role:org:admin, template:org, *
p, role:org:admin, workspace:org, *
p, role:org:admin, volume:org, list
p, role:org:admin, volume:org, read

p, role:org:member, template:org, read
p, role:org:member, workspace:own, *
p, role:org:member, volume:own, *

# 普通用户角色
p, role:user, workspace:own, *
p, role:user, volume:own, *
p, role:user, template:public, read
```

### 3.3 阶段三：Service层增强

#### 3.3.1 WorkspaceService 增强
```go
// internal/services/workspace_service.go

// ListWorkspaces 根据用户权限列出工作区
func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID int64, userRole models.UserRole, orgRole *models.OrganizationMemberRole, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
    var workspaces []*models.Workspace
    var total int64
    var err error
    
    switch userRole {
    case models.UserRoleSuperAdmin:
        // 超级管理员可以看到所有工作区
        workspaces, total, err = s.workspaceRepo.ListAll(ctx, opts)
    case models.UserRoleAdmin:
        // 管理员可以看到所有工作区
        workspaces, total, err = s.workspaceRepo.ListAll(ctx, opts)
    default:
        // 根据组织角色决定可见范围
        if orgRole != nil && (*orgRole == models.OrgRoleOwner || *orgRole == models.OrgRoleAdmin) {
            // organization:owner 和 organization:admin 可以看到组织内所有工作区
            workspaces, total, err = s.workspaceRepo.ListByOrganizationAll(ctx, userID, opts)
        } else {
            // 普通用户只能看到：1.自己创建的 2.自己所属的 3.组织内共享的工作区
            workspaces, total, err = s.workspaceRepo.ListAccessibleByUser(ctx, userID, opts)
        }
    }
    
    return workspaces, total, err
}

// ListWorkspacesByOrganization 列出组织的工作区
func (s *WorkspaceService) ListWorkspacesByOrganization(ctx context.Context, orgID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
    return s.workspaceRepo.ListByLabel(ctx, map[string]string{
        "organization_id": fmt.Sprintf("%d", orgID),
    }, opts)
}
```

#### 3.3.2 VolumeService 增强
```go
// internal/services/volume_service.go

// ListVolumes 根据用户权限列出存储卷
func (s *VolumeService) ListVolumes(ctx context.Context, userID int64, userRole models.UserRole, opts *models.ListOptions) ([]*models.Volume, int64, error) {
    // 类似 WorkspaceService
}

// ListVolumesByOrganization 列出组织的存储卷
func (s *VolumeService) ListVolumesByOrganization(ctx context.Context, orgID int64, opts *models.ListOptions) ([]*models.Volume, int64, error) {
    return s.volumeRepo.ListByLabel(ctx, map[string]string{
        "organization_id": fmt.Sprintf("%d", orgID),
    }, opts)
}
```

#### 3.3.3 TemplateService 增强
```go
// internal/services/template_service.go

// ListTemplates 根据用户权限列出模板
func (s *TemplateService) ListTemplates(ctx context.Context, userID int64, userRole models.UserRole, orgIDs []int64, opts *models.ListOptions) ([]*models.Template, int64, error) {
    switch userRole {
    case models.UserRoleSuperAdmin, models.UserRoleAdmin:
        // 管理员可以看到所有模板
        return s.templateRepo.ListAll(ctx, opts)
    default:
        // 普通用户可以看到：1. 公开模板 2. 自己创建的模板 3. 所属组织的模板
        return s.templateRepo.ListAccessibleByUser(ctx, userID, orgIDs, opts)
    }
}
```

#### 3.3.4 UserService 增强
```go
// internal/services/user_service.go

// CreateUser 管理员创建用户
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
    // 实现管理员创建用户功能
}

// UpdateUserProfile 用户更新自己的profile
func (s *UserService) UpdateUserProfile(ctx context.Context, userID int64, req *models.UpdateUserProfileRequest) (*models.User, error) {
    // 只允许更新: display_name, email, avatar_url
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    if req.DisplayName != nil {
        user.DisplayName = req.DisplayName
    }
    if req.Email != nil {
        user.Email = req.Email
    }
    if req.AvatarURL != nil {
        user.AvatarURL = req.AvatarURL
    }
    
    return s.userRepo.Update(ctx, user)
}

// UpdateUserByAdmin 管理员更新用户信息
func (s *UserService) UpdateUserByAdmin(ctx context.Context, userID int64, req *models.UpdateUserRequest, isSuper bool) (*models.User, error) {
    // 普通管理员只能更新: display_name, email, avatar_url, status
    // 超级管理员可以额外更新: role
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // 允许所有管理员更新的字段
    if req.DisplayName != nil {
        user.DisplayName = req.DisplayName
    }
    if req.Email != nil {
        user.Email = req.Email
    }
    if req.AvatarURL != nil {
        user.AvatarURL = req.AvatarURL
    }
    if req.Status != nil {
        user.Status = *req.Status
    }
    
    // 只有超级管理员可以修改角色
    if isSuper && req.Role != nil {
        user.Role = *req.Role
    }
    
    return s.userRepo.Update(ctx, user)
}

// CheckUserExists power_user检查用户是否存在
func (s *UserService) CheckUserExists(ctx context.Context, username string) (bool, error) {
    user, err := s.userRepo.GetByUsername(ctx, username)
    if err != nil {
        return false, nil
    }
    return user != nil, nil
}

// SearchUsers organization:owner搜索用户（用于添加成员）
func (s *UserService) SearchUsers(ctx context.Context, query string, opts *models.ListOptions) ([]*models.User, int64, error) {
    // 根据用户名、邮箱等搜索用户
    return s.userRepo.SearchByQuery(ctx, query, opts)
}

// DeleteUser 删除用户（仅超级管理员）
func (s *UserService) DeleteUser(ctx context.Context, userID int64) error {
    // 删除用户及其相关资源
    // 1. 删除用户的工作区
    // 2. 删除用户的存储卷
    // 3. 从组织中移除
    // 4. 删除用户
    return s.userRepo.Delete(ctx, userID)
}
```

#### 3.3.5 OrganizationService 增强
```go
// internal/services/organization_service.go

// ListAllOrganizations 管理员列出所有组织
func (s *OrganizationService) ListAllOrganizations(ctx context.Context, opts *models.ListOptions) ([]*models.Organization, int64, error) {
    return s.orgRepo.ListAll(ctx, opts)
}
```

### 3.4 阶段四：Repository层增强

#### 3.4.1 WorkspaceRepository 增强
```go
// internal/repository/workspace_repository.go

// ListAll 列出所有工作区（管理员）
func (r *WorkspaceRepository) ListAll(ctx context.Context, opts *models.ListOptions) ([]*models.Workspace, int64, error)

// ListByOrganizationAll 列出用户所属组织的所有工作区（organization:owner和admin）
func (r *WorkspaceRepository) ListByOrganizationAll(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error)

// ListAccessibleByUser 列出用户可访问的工作区（自己创建的、自己所属的、组织内共享的）
func (r *WorkspaceRepository) ListAccessibleByUser(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error)

// ListByLabel 根据标签列出工作区
func (r *WorkspaceRepository) ListByLabel(ctx context.Context, labels map[string]string, opts *models.ListOptions) ([]*models.Workspace, int64, error)

// UpdateLabels 更新工作区标签
func (r *WorkspaceRepository) UpdateLabels(ctx context.Context, id int64, labels models.ResourceLabels) error
```

#### 3.4.2 VolumeRepository 增强
```go
// internal/repository/volume_repository.go

// 类似 WorkspaceRepository 的增强
func (r *VolumeRepository) ListAll(ctx context.Context, opts *models.ListOptions) ([]*models.Volume, int64, error)
func (r *VolumeRepository) ListAccessibleByUser(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.Volume, int64, error)
func (r *VolumeRepository) ListByLabel(ctx context.Context, labels map[string]string, opts *models.ListOptions) ([]*models.Volume, int64, error)
```

#### 3.4.3 TemplateRepository 增强
```go
// internal/repository/template_repository.go

// ListAll 列出所有模板（管理员）
func (r *TemplateRepository) ListAll(ctx context.Context, opts *models.ListOptions) ([]*models.Template, int64, error)

// ListAccessibleByUser 列出用户可访问的模板
func (r *TemplateRepository) ListAccessibleByUser(ctx context.Context, userID int64, orgIDs []int64, opts *models.ListOptions) ([]*models.Template, int64, error)

// ListByVisibility 根据可见性列出模板
func (r *TemplateRepository) ListByVisibility(ctx context.Context, visibility models.TemplateVisibility, opts *models.ListOptions) ([]*models.Template, int64, error)
```

#### 3.4.4 UserRepository 增强
```go
// internal/repository/user_repository.go

// ListAll 列出所有用户（管理员）
func (r *UserRepository) ListAll(ctx context.Context, opts *models.ListOptions) ([]*models.User, int64, error)

// SearchByQuery 搜索用户（organization:owner用于添加成员）
func (r *UserRepository) SearchByQuery(ctx context.Context, query string, opts *models.ListOptions) ([]*models.User, int64, error)
```

#### 3.4.5 OrganizationRepository 增强
```go
// internal/repository/organization_repository.go

// ListAll 列出所有组织（管理员）
func (r *OrganizationRepository) ListAll(ctx context.Context, opts *models.ListOptions) ([]*models.Organization, int64, error)

// GetUserOrganizations 获取用户所属的组织
func (r *OrganizationRepository) GetUserOrganizations(ctx context.Context, userID int64) ([]*models.Organization, error)

// GetUserOrganizationRole 获取用户在组织中的角色
func (r *OrganizationRepository) GetUserOrganizationRole(ctx context.Context, userID, orgID int64) (models.OrganizationMemberRole, error)
```

### 3.5 阶段五：Handler层重构

#### 3.5.1 WorkspaceHandler 重构
```go
// internal/handlers/workspace_handler.go

// ListWorkspaces 重构
// GET /api/v1/workspaces
// 查询参数（可选）:
//   - organization_id: 过滤特定组织的工作区（需要权限）
//   - page, page_size: 分页
//   - sort_by, sort_order: 排序
func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
    userID := c.GetInt64("user_id")
    userRole := c.GetString("user_role")
    
    // 解析查询参数
    var opts models.ListOptions
    if err := c.ShouldBindQuery(&opts); err != nil {
        // ... 错误处理
    }
    
    orgIDStr := c.Query("organization_id")
    
    var workspaces []*models.Workspace
    var total int64
    var err error
    
    if orgIDStr != "" {
        // 请求特定组织的工作区
        orgID, _ := strconv.ParseInt(orgIDStr, 10, 64)
        
        // 检查权限（通过RBAC中间件或在service层）
        workspaces, total, err = h.workspaceService.ListWorkspacesByOrganization(c.Request.Context(), orgID, &opts)
    } else {
        // 根据用户权限返回工作区
        workspaces, total, err = h.workspaceService.ListWorkspaces(c.Request.Context(), userID, models.UserRole(userRole), &opts)
    }
    
    // ... 返回结果
}
```

#### 3.5.2 VolumeHandler 重构
```go
// internal/handlers/volume_handler.go

// ListVolumes 重构（类似 WorkspaceHandler）
func (h *VolumeHandler) ListVolumes(c *gin.Context)
```

#### 3.5.3 TemplateHandler 重构
```go
// internal/handlers/template_handler.go

// ListTemplates 重构
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
    userID := c.GetInt64("user_id")
    userRole := c.GetString("user_role")
    
    // 获取用户所属的组织
    orgIDs, _ := h.orgService.GetUserOrganizationIDs(c.Request.Context(), userID)
    
    var opts models.ListOptions
    if err := c.ShouldBindQuery(&opts); err != nil {
        // ... 错误处理
    }
    
    templates, total, err := h.templateService.ListTemplates(c.Request.Context(), userID, models.UserRole(userRole), orgIDs, &opts)
    
    // ... 返回结果
}
```

#### 3.5.4 UserHandler 重构
```go
// internal/handlers/user_handler.go

// ListUsers 重构
// GET /api/v1/users
// 只有管理员可以访问
func (h *UserHandler) ListUsers(c *gin.Context) {
    // 已通过RBAC中间件验证权限
    
    var opts models.ListOptions
    if err := c.ShouldBindQuery(&opts); err != nil {
        // ... 错误处理
    }
    
    users, total, err := h.userService.ListUsers(c.Request.Context(), &opts)
    
    // ... 返回结果
}

// CreateUser 管理员创建用户
// POST /api/v1/users
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req models.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // ... 错误处理
    }

    user, err := h.userService.CreateUser(c.Request.Context(), &req)

    // ... 返回结果
}

// UpdateProfile 用户更新自己的profile
// PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
    userID := c.GetInt64("user_id")
    
    var req models.UpdateUserProfileRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Error: &models.APIError{
                Code:    "INVALID_REQUEST",
                Message: "Invalid request body",
                Details: err.Error(),
            },
        })
        return
    }

    user, err := h.userService.UpdateUserProfile(c.Request.Context(), userID, &req)
    if err != nil {
        // ... 错误处理
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Data:    user,
    })
}

// UpdateUser 管理员更新用户
// PUT /api/v1/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
    userRole := models.UserRole(c.GetString("user_role"))
    isSuper := userRole == models.UserRoleSuperAdmin
    
    idParam := c.Param("id")
    id, err := strconv.ParseInt(idParam, 10, 64)
    if err != nil {
        // ... 错误处理
    }
    
    var req models.UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // ... 错误处理
    }

    user, err := h.userService.UpdateUserByAdmin(c.Request.Context(), id, &req, isSuper)
    // ... 返回结果
}

// DeleteUser 删除用户（仅超级管理员）
// DELETE /api/v1/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
    idParam := c.Param("id")
    id, err := strconv.ParseInt(idParam, 10, 64)
    if err != nil {
        // ... 错误处理
    }

    err = h.userService.DeleteUser(c.Request.Context(), id)
    // ... 返回结果
}

// CheckUserExists power_user查询用户是否存在
// GET /api/v1/users/check?username=xxx
func (h *UserHandler) CheckUserExists(c *gin.Context) {
    username := c.Query("username")
    if username == "" {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success: false,
            Error: &models.APIError{
                Code:    "INVALID_REQUEST",
                Message: "username parameter is required",
            },
        })
        return
    }

    exists, err := h.userService.CheckUserExists(c.Request.Context(), username)
    if err != nil {
        // ... 错误处理
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Success: true,
        Data: map[string]interface{}{
            "username": username,
            "exists":   exists,
        },
    })
}
```

#### 3.5.5 OrganizationHandler 重构
```go
// internal/handlers/organization_handler.go

// ListUserOrganizations 重构 - 列出当前用户的组织
// GET /api/v1/organizations
func (h *OrganizationHandler) ListUserOrganizations(c *gin.Context)

// ListAllOrganizations 管理员列出所有组织
// GET /api/v1/organizations?all=true
// 需要管理员权限
func (h *OrganizationHandler) ListAllOrganizations(c *gin.Context) {
    var opts models.ListOptions
    if err := c.ShouldBindQuery(&opts); err != nil {
        // ... 错误处理
    }
    
    orgs, total, err := h.orgService.ListAllOrganizations(c.Request.Context(), &opts)
    
    // ... 返回结果
}

// PromoteToAdmin 提升成员为管理员（仅organization:owner）
// POST /api/v1/organizations/:id/admins/:user_id
func (h *OrganizationHandler) PromoteToAdmin(c *gin.Context)

// DemoteFromAdmin 解除管理员身份（仅organization:owner）
// DELETE /api/v1/organizations/:id/admins/:user_id
func (h *OrganizationHandler) DemoteFromAdmin(c *gin.Context)

// SearchUsers 搜索用户以添加到组织（organization:owner）
// GET /api/v1/organizations/:id/search-users?q=username
func (h *OrganizationHandler) SearchUsers(c *gin.Context)
```

### 3.6 阶段六：API路由重构

#### 3.6.1 统一路由设计
```go
// internal/api/server.go

func (s *Server) SetupRoutes(...) {
    // ... 现有代码
    
    // API v1 routes
    v1 := s.router.Group("/api/v1")
    {
        // Public auth routes
        auth := v1.Group("/auth")
        {
            auth.POST("/login", userHandler.Login)
            auth.POST("/register", userHandler.Register)
        }
        
        // Protected routes with JWT authentication
        protected := v1.Group("")
        protected.Use(middleware.AuthMiddleware(s.jwtManager))
        {
            // User routes
            users := protected.Group("/users")
            {
                users.GET("/me", userHandler.GetProfile)
                users.PUT("/me", userHandler.UpdateProfile)
                users.POST("/me/password", userHandler.ChangePassword)
                
                // power_user可以查询用户是否存在
                users.GET("/check", middleware.RBACCheck(rbacClient, "user", "check"), userHandler.CheckUserExists)
                
                // 需要RBAC权限检查（管理员）
                users.GET("", middleware.RBACCheck(rbacClient, "user", "list"), userHandler.ListUsers)
                users.POST("", middleware.RBACCheck(rbacClient, "user", "create"), userHandler.CreateUser)
                users.GET("/:id", middleware.RBACCheck(rbacClient, "user", "read"), userHandler.GetUser)
                users.PUT("/:id", middleware.RBACCheck(rbacClient, "user", "update"), userHandler.UpdateUser)
                
                // 只有超级管理员可以删除用户
                users.DELETE("/:id", middleware.RequireSuperAdmin(), userHandler.DeleteUser)
            }
            
            // Organization routes
            orgs := protected.Group("/organizations")
            {
                // power_user可以创建组织
                orgs.POST("", middleware.RBACCheck(rbacClient, "organization", "create"), orgHandler.CreateOrganization)
                orgs.GET("", orgHandler.ListOrganizations) // 根据查询参数决定是列出用户的组织还是所有组织
                orgs.GET("/:id", middleware.RBACCheck(rbacClient, "organization", "read"), orgHandler.GetOrganization)
                
                // organization:owner可以更新和删除组织
                orgs.PUT("/:id", middleware.RBACCheck(rbacClient, "organization", "update"), orgHandler.UpdateOrganization)
                orgs.DELETE("/:id", middleware.RBACCheck(rbacClient, "organization", "delete"), orgHandler.DeleteOrganization)
                
                // Organization member management (organization:owner和admin都可以管理普通成员)
                orgs.POST("/:id/members", middleware.RBACCheck(rbacClient, "organization", "manage_members"), orgHandler.AddMember)
                orgs.DELETE("/:id/members/:user_id", middleware.RBACCheck(rbacClient, "organization", "manage_members"), orgHandler.RemoveMember)
                orgs.PUT("/:id/members/:user_id", middleware.RBACCheck(rbacClient, "organization", "manage_members"), orgHandler.UpdateMemberRole)
                
                // Admin management (仅organization:owner)
                orgs.POST("/:id/admins/:user_id", middleware.RBACCheck(rbacClient, "organization", "manage_admins"), orgHandler.PromoteToAdmin)
                orgs.DELETE("/:id/admins/:user_id", middleware.RBACCheck(rbacClient, "organization", "manage_admins"), orgHandler.DemoteFromAdmin)
                
                // User search (organization:owner用于搜索并添加成员)
                orgs.GET("/:id/search-users", middleware.RBACCheck(rbacClient, "user", "search"), orgHandler.SearchUsers)
                
                // 组织资源管理（organization:owner和admin可以查看和管理）
                orgs.GET("/:id/workspaces", middleware.RBACCheck(rbacClient, "organization", "read"), workspaceHandler.ListOrganizationWorkspaces)
                orgs.GET("/:id/volumes", middleware.RBACCheck(rbacClient, "organization", "read"), volumeHandler.ListOrganizationVolumes)
                orgs.GET("/:id/templates", middleware.RBACCheck(rbacClient, "organization", "read"), templateHandler.ListOrganizationTemplates)
            }
            
            // Template routes
            templates := protected.Group("/templates")
            {
                templates.GET("", templateHandler.ListTemplates) // 根据用户权限返回不同范围
                templates.POST("", middleware.RBACCheck(rbacClient, "template", "create"), templateHandler.CreateTemplate)
                templates.GET("/:id", middleware.RBACCheck(rbacClient, "template", "read"), templateHandler.GetTemplate)
                templates.PUT("/:id", middleware.RBACCheck(rbacClient, "template", "update"), templateHandler.UpdateTemplate)
                templates.DELETE("/:id", middleware.RBACCheck(rbacClient, "template", "delete"), templateHandler.DeleteTemplate)
            }
            
            // Workspace routes
            workspaces := protected.Group("/workspaces")
            {
                workspaces.GET("", workspaceHandler.ListWorkspaces) // 根据用户权限返回不同范围
                workspaces.POST("", middleware.RBACCheck(rbacClient, "workspace", "create"), workspaceHandler.CreateWorkspace)
                workspaces.GET("/:id", middleware.RBACCheck(rbacClient, "workspace", "read"), workspaceHandler.GetWorkspace)
                workspaces.PUT("/:id", middleware.RBACCheck(rbacClient, "workspace", "update"), workspaceHandler.UpdateWorkspace)
                workspaces.DELETE("/:id", middleware.RBACCheck(rbacClient, "workspace", "delete"), workspaceHandler.DeleteWorkspace)
                workspaces.POST("/:id/start", middleware.RBACCheck(rbacClient, "workspace", "start"), workspaceHandler.StartWorkspace)
                workspaces.POST("/:id/stop", middleware.RBACCheck(rbacClient, "workspace", "stop"), workspaceHandler.StopWorkspace)
                workspaces.POST("/:id/volumes/:volume_id", middleware.RBACCheck(rbacClient, "workspace", "update"), workspaceHandler.AttachVolume)
                workspaces.DELETE("/:id/volumes/:volume_id", middleware.RBACCheck(rbacClient, "workspace", "update"), workspaceHandler.DetachVolume)
            }
            
            // Volume routes
            volumes := protected.Group("/volumes")
            {
                volumes.GET("", volumeHandler.ListVolumes) // 根据用户权限返回不同范围
                volumes.POST("", middleware.RBACCheck(rbacClient, "volume", "create"), volumeHandler.CreateVolume)
                volumes.GET("/:id", middleware.RBACCheck(rbacClient, "volume", "read"), volumeHandler.GetVolume)
                volumes.PUT("/:id", middleware.RBACCheck(rbacClient, "volume", "update"), volumeHandler.UpdateVolume)
                volumes.DELETE("/:id", middleware.RBACCheck(rbacClient, "volume", "delete"), volumeHandler.DeleteVolume)
            }
        }
    }
}
```

#### 3.6.2 RBAC中间件改进
```go
// internal/middleware/rbac.go

// RBACCheck 创建一个RBAC检查中间件
func RBACCheck(rbacClient *rbac.Client, resourceType, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetInt64("user_id")
        resourceID := c.Param("id")
        
        // 对于list操作，不需要resourceID
        if action == "list" {
            resourceID = ""
        }
        
        allowed, err := rbacClient.CheckPermission(c.Request.Context(), &rbac.CheckPermissionRequest{
            UserID:       userID,
            ResourceType: resourceType,
            ResourceID:   resourceID,
            Action:       action,
        })
        
        if err != nil {
            c.JSON(http.StatusInternalServerError, models.APIResponse{
                Success: false,
                Error: &models.APIError{
                    Code:    "INTERNAL_ERROR",
                    Message: "Failed to check permission",
                },
            })
            c.Abort()
            return
        }
        
        if !allowed {
            c.JSON(http.StatusForbidden, models.APIResponse{
                Success: false,
                Error: &models.APIError{
                    Code:    "PERMISSION_DENIED",
                    Message: "You don't have permission to perform this action",
                },
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 3.7 阶段七：资源标签自动管理

#### 3.7.1 创建时自动添加标签
```go
// internal/services/workspace_service.go

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req *models.CreateWorkspaceRequest) (*models.Workspace, error) {
    // ... 现有代码
    
    // 自动添加标签
    labels := models.ResourceLabels{
        "owner_type": string(req.OwnerType),
        "owner_id":   fmt.Sprintf("%d", req.OwnerID),
    }
    
    // 如果是组织拥有的，添加组织标签
    if req.OwnerType == models.OwnerTypeOrganization {
        labels["organization_id"] = fmt.Sprintf("%d", req.OwnerID)
    }
    
    workspace.Labels = labels
    
    // ... 继续现有逻辑
}
```

#### 3.7.2 在RBAC中注册资源权限
```go
// internal/services/workspace_service.go

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req *models.CreateWorkspaceRequest) (*models.Workspace, error) {
    // ... 创建workspace
    
    // 在RBAC中注册权限
    if s.rbacClient != nil {
        // 赋予创建者完全权限
        resourceID := fmt.Sprintf("workspace:%d", workspace.ID)
        err := s.rbacClient.AssignRoleForResource(ctx, req.CreatedBy, "workspace", resourceID, "owner")
        if err != nil {
            s.logger.Error("Failed to assign RBAC permission",
                zap.Int64("workspace_id", workspace.ID),
                zap.Error(err))
        }
        
        // 如果是组织资源，赋予组织成员权限
        if req.OwnerType == models.OwnerTypeOrganization {
            // 异步处理组织成员权限
            go s.assignOrganizationMemberPermissions(context.Background(), workspace)
        }
    }
    
    return workspace, nil
}
```

### 3.8 阶段八：测试

#### 3.8.1 手动测试验证
使用Postman或curl进行手动API测试，验证：

**用户相关：**
- 普通用户可以编辑自己的profile（昵称、邮箱、头像）
- 普通用户不能修改自己的role
- 管理员可以创建用户
- 管理员不能修改用户role（只有超级管理员可以）
- 超级管理员可以删除用户
- power_user可以查询用户是否存在

**组织相关：**
- user不能创建组织
- power_user可以创建组织
- organization:owner可以管理组织信息和成员
- organization:owner可以指派/解除管理员
- organization:owner可以搜索用户并添加到组织
- organization:owner可以查看和管理组织内所有资源（包括所有工作区）
- organization:owner可以删除组织内成员的工作区和存储卷
- organization:admin可以管理普通成员，但不能修改owner或其他admin
- organization:admin可以查看组织内所有工作区
- organization:admin不能指派/解除管理员
- organization:member只能看到自己的工作区和组织内共享的工作区

**资源相关：**
- 用户可以列出自己的工作区和存储卷
- 管理员可以列出所有资源
- organization:owner可以列出组织内的资源
- RBAC权限正确控制各种操作

## 4. 文件清单

### 4.1 新增文件
```
components/controller/
├── pkg/rbac/
│   ├── client.go                    # RBAC客户端
│   └── types.go                     # RBAC类型定义
└── internal/middleware/
    └── rbac.go                      # RBAC中间件
    │   ├── workspace_test.go
    │   ├── volume_test.go
    │   ├── template_test.go
    │   └── rbac_test.go
    └── api/
        ├── user_test.go
        ├── workspace_test.go
        ├── volume_test.go
        ├── template_test.go
        └── organization_test.go
```

### 4.2 修改文件
```
components/controller/
├── migrations/
│   ├── 000001_init_schema.up.sql   # 添加 labels 和 visibility 字段
│   └── 000001_init_schema.down.sql
├── internal/
│   ├── models/
│   │   ├── common.go               # 添加 ResourceLabels
│   │   ├── user.go                 # 添加 UpdateUserProfileRequest
│   │   ├── workspace.go            # 添加 Labels 字段
│   │   ├── volume.go               # 添加 Labels 字段
│   │   └── template.go             # 添加 Labels 和 Visibility 字段
│   ├── repository/
│   │   ├── workspace_repository.go  # 添加新方法
│   │   ├── volume_repository.go     # 添加新方法
│   │   ├── template_repository.go   # 添加新方法
│   │   ├── user_repository.go       # 添加新方法
│   │   └── organization_repository.go # 添加新方法
│   ├── services/
│   │   ├── workspace_service.go     # 重构List方法，添加RBAC集成
│   │   ├── volume_service.go        # 重构List方法，添加RBAC集成
│   │   ├── template_service.go      # 重构List方法，添加RBAC集成
│   │   ├── user_service.go          # 添加Profile、Delete、Check方法
│   │   └── organization_service.go  # 添加ListAll方法
│   ├── handlers/
│   │   ├── workspace_handler.go     # 重构ListWorkspaces
│   │   ├── volume_handler.go        # 重构ListVolumes
│   │   ├── template_handler.go      # 重构ListTemplates
│   │   ├── user_handler.go          # 添加Profile、Delete、Check handlers
│   │   └── organization_handler.go  # 添加ListAll、资源管理handlers
│   ├── middleware/
│   │   └── auth.go                  # 添加RequireSuperAdmin中间件
│   └── api/
│       └── server.go                # 重构路由，添加RBAC中间件
├── cmd/controller/main.go           # 初始化RBAC客户端
└── internal/config/config.go        # 添加RBAC配置
```

### 4.3 RBAC服务修改
```
components/rbac/
├── configs/
│   ├── model.conf                   # 更新策略模型
│   └── policy.csv                   # 更新策略规则
└── internal/
    └── rbac/
        └── enforcer.go              # 可能需要增强
```

## 5. 实施步骤

### 5.1 准备阶段（已完成）
1. ✅ 分析现有代码
2. ✅ 编写重构计划文档

### 5.2 第一阶段：数据层和RBAC基础（2-3天）
1. ⬜ 修改数据库迁移SQL（添加labels和visibility字段）
2. ⬜ 更新models定义（添加ResourceLabels、UpdateUserProfileRequest等）
3. ⬜ 实现RBAC客户端
4. ⬜ 更新RBAC服务策略模型（添加power_user角色）
5. ⬜ 实现Repository层新方法

### 5.3 第二阶段：Service层（2-3天）
1. ⬜ 实现UserService增强方法（Profile、Delete、Check）
2. ⬜ 实现WorkspaceService List方法重构
3. ⬜ 实现VolumeService List方法重构
4. ⬜ 实现TemplateService List方法重构
5. ⬜ 实现OrganizationService ListAll方法
6. ⬜ 添加资源标签自动管理

### 5.4 第三阶段：Handler和API层（2-3天）
1. ⬜ 实现RBAC中间件和RequireSuperAdmin中间件
2. ⬜ 重构UserHandler（添加Profile、Delete、Check）
3. ⬜ 重构WorkspaceHandler（List方法）
4. ⬜ 重构VolumeHandler（List方法）
5. ⬜ 重构TemplateHandler（List方法）
6. ⬜ 重构OrganizationHandler（添加资源管理路由）
7. ⬜ 更新API路由配置

### 5.5 第四阶段：测试和验证（1-2天）
1. ⬜ 手动API测试（使用Postman或curl）
2. ⬜ 验证用户权限控制
3. ⬜ 验证organization owner权限
4. ⬜ 验证power_user功能
5. ⬜ 修复发现的问题

### 5.6 第五阶段：文档和部署（1天）
1. ⬜ 更新API文档
2. ⬜ 准备部署脚本
3. ⬜ 部署到开发环境

## 6. 注意事项

### 6.1 RBAC策略配置
- 确保power_user角色正确配置
- 确保organization:owner对组织资源有完全控制权
- 测试各角色的权限边界

### 6.2 字段级别权限控制
- 普通用户只能修改自己的profile字段（display_name、email、avatar_url）
- 管理员可以修改用户的status，但不能修改role
- 只有超级管理员可以修改用户的role
- 只有超级管理员可以删除用户

### 6.3 组织资源管理
- organization:owner可以查看和管理组织内所有资源
- organization:owner可以删除组织内任何成员的工作区和存储卷
- 确保组织删除时正确清理相关资源

### 6.4 性能考虑
- RBAC检查可能增加API延迟
- 如需要可在RBAC服务中实现缓存
- 监控关键API的响应时间

## 7. 成功标准

### 7.1 功能完整性
- ✅ 所有API端点都有RBAC保护
- ✅ 用户可以编辑自己的profile（昵称、邮箱、头像）
- ✅ 用户不能修改自己的role和status
- ✅ 管理员可以创建用户，但不能修改role
- ✅ 超级管理员可以修改role和删除用户
- ✅ power_user可以创建组织和查询用户
- ✅ user不能创建组织
- ✅ 普通用户可以列出自己的资源
- ✅ System Admin可以列出所有资源
- ✅ Organization Owner可以查看和管理组织内所有资源
- ✅ 所有资源都有标签

### 7.2 权限验证
- ✅ 各角色的权限边界正确
- ✅ organization:owner对组织资源有完全控制
- ✅ 字段级别的权限控制正确实施

### 7.3 文档完整性
- ✅ API文档更新
- ✅ RBAC策略文档
- ✅ 角色权限说明文档

## 8. 后续优化

### 8.1 性能优化
- 实现RBAC决策缓存
- 优化数据库查询
- 添加Redis缓存层

### 8.2 功能增强
- 支持自定义角色
- 支持资源级别的细粒度权限
- 添加审计日志

### 8.3 监控和告警
- 添加权限检查失败告警
- 监控RBAC服务健康状态
- 性能指标可视化

## 9. 参考资料

- [api.md](../../references/controller/docs/api.md) - 参考API设计
- [Casbin文档](https://casbin.org/docs/zh-CN/overview) - RBAC实现
- [Gin框架文档](https://gin-gonic.com/docs/) - Web框架
- [PostgreSQL JSONB](https://www.postgresql.org/docs/current/datatype-json.html) - 标签存储

---

**文档版本:** 1.0  
**创建日期:** 2026-01-14  
**最后更新:** 2026-01-14  
**维护者:** IDEKube Team
