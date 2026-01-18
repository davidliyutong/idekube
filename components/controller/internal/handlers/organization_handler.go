package handlers

import (
	"net/http"
	"strconv"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// OrganizationHandler handles organization-related HTTP requests
type OrganizationHandler struct {
	orgService *services.OrganizationService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(orgService *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
	}
}

// Create godoc
// @Summary 创建组织
// @Description 创建新的组织
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateOrganizationRequest true "组织创建请求"
// @Success 201 {object} models.APIResponse{data=models.Organization} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /organizations [post]
func (h *OrganizationHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req models.CreateOrganizationRequest
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

	org, err := h.orgService.CreateOrganization(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATE_FAILED",
				Message: "Failed to create organization",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    org,
	})
}

// List godoc
// @Summary 列出用户的组织
// @Description 获取当前用户所属的所有组织
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param all query boolean false "管理员是否列出所有组织"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} models.APIResponse{data=[]models.Organization} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /organizations [get]
func (h *OrganizationHandler) List(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	// Check if requesting all organizations (permission already checked by RBAC middleware)
	if c.Query("all") == "true" {
		var opts models.ListOptions
		if err := c.ShouldBindQuery(&opts); err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_REQUEST",
					Message: "Invalid query parameters",
					Details: err.Error(),
				},
			})
			return
		}

		organizations, total, err := h.orgService.ListAllOrganizations(c.Request.Context(), &opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to list organizations",
					Details: err.Error(),
				},
			})
			return
		}

		totalPages := int(total) / opts.PageSize
		if int(total)%opts.PageSize > 0 {
			totalPages++
		}

		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Data: models.PaginatedResponse{
				Items:      organizations,
				Total:      total,
				Page:       opts.Page,
				PageSize:   opts.PageSize,
				TotalPages: totalPages,
			},
		})
		return
	}

	orgs, err := h.orgService.ListUserOrganizations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list organizations",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    orgs,
	})
}

// Delete godoc
// @Summary 删除组织
// @Description 删除指定的组织（仅所有者可操作）
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /organizations/{id} [delete]
func (h *OrganizationHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	userID := middleware.MustGetUserID(c)

	// Check if user is owner
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, userID, models.OrgRoleOwner)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Only organization owner can delete the organization",
			},
		})
		return
	}

	err = h.orgService.DeleteOrganization(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete organization",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Organization deleted successfully",
	})
}

// ============================================================================
// Profile Sub-resource Handlers
// ============================================================================

// GetProfile godoc
// @Summary 获取组织Profile
// @Description 获取指定组织的Profile子资源
// @Tags Organization Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Success 200 {object} models.APIResponse{data=models.OrganizationProfileResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "组织不存在"
// @Router /organizations/{id}/profile [get]
func (h *OrganizationHandler) GetProfile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	profile, err := h.orgService.GetOrganizationProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Organization not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    profile,
	})
}

// UpdateProfile godoc
// @Summary 更新组织Profile
// @Description 更新指定组织的Profile子资源
// @Tags Organization Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param request body models.UpdateOrganizationProfileRequest true "Profile更新请求"
// @Success 200 {object} models.APIResponse{data=models.OrganizationProfileResponse} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "组织不存在"
// @Router /organizations/{id}/profile [put]
func (h *OrganizationHandler) UpdateProfile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	userID := middleware.MustGetUserID(c)

	// Check if user has admin permission
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, userID, models.OrgRoleAdmin)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Insufficient permissions",
			},
		})
		return
	}

	var req models.UpdateOrganizationProfileRequest
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

	profile, err := h.orgService.UpdateOrganizationProfile(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update organization profile",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    profile,
	})
}

// ============================================================================
// Members Sub-resource Handlers
// ============================================================================

// ListMembers godoc
// @Summary 获取组织成员列表
// @Description 获取指定组织的所有成员
// @Tags Organization Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Success 200 {object} models.APIResponse{data=[]models.OrganizationMemberWithUser} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "组织不存在"
// @Router /organizations/{id}/members [get]
func (h *OrganizationHandler) ListMembers(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	members, err := h.orgService.ListOrganizationMembers(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Organization not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    members,
	})
}

// AddMembers godoc
// @Summary 添加组织成员
// @Description 向组织添加新成员（需要管理员权限）
// @Tags Organization Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param request body models.AddOrganizationMemberRequest true "成员添加请求"
// @Success 201 {object} models.APIResponse{data=models.OrganizationMember} "添加成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /organizations/{id}/members [post]
func (h *OrganizationHandler) AddMembers(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	userID := middleware.MustGetUserID(c)

	// Check if user has admin permission
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, userID, models.OrgRoleAdmin)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Insufficient permissions",
			},
		})
		return
	}

	var req models.AddOrganizationMemberRequest
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

	member, err := h.orgService.AddMember(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ADD_MEMBER_FAILED",
				Message: "Failed to add member",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    member,
	})
}

// RemoveMember godoc
// @Summary 移除组织成员
// @Description 从组织中移除成员（需要管理员权限）
// @Tags Organization Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param user_id path int true "用户ID"
// @Success 200 {object} models.APIResponse "移除成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /organizations/{id}/members/{user_id} [delete]
func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	memberIDParam := c.Param("user_id")
	memberID, err := strconv.ParseInt(memberIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	userID := middleware.MustGetUserID(c)

	// Check if user has admin permission
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, userID, models.OrgRoleAdmin)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Insufficient permissions",
			},
		})
		return
	}

	err = h.orgService.RemoveMember(c.Request.Context(), id, memberID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "REMOVE_MEMBER_FAILED",
				Message: "Failed to remove member",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Member removed successfully",
	})
}

// UpdateMemberRole godoc
// @Summary 更新成员角色
// @Description 更新组织成员的角色（仅所有者可操作）
// @Tags Organization Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param user_id path int true "用户ID"
// @Param request body models.UpdateOrganizationMemberRequest true "角色更新请求"
// @Success 200 {object} models.APIResponse{data=models.OrganizationMember} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /organizations/{id}/members/{user_id} [put]
func (h *OrganizationHandler) UpdateMemberRole(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	memberIDParam := c.Param("user_id")
	memberID, err := strconv.ParseInt(memberIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	userID := middleware.MustGetUserID(c)

	// Check if user has owner permission (only owner can change roles)
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, userID, models.OrgRoleOwner)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Only organization owner can change member roles",
			},
		})
		return
	}

	var req models.UpdateOrganizationMemberRequest
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

	member, err := h.orgService.UpdateMemberRole(c.Request.Context(), id, memberID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_ROLE_FAILED",
				Message: "Failed to update member role",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    member,
	})
}

// ============================================================================
// Admins Sub-resource Handlers
// ============================================================================

// ListAdmins godoc
// @Summary 获取组织管理员列表
// @Description 获取指定组织的所有管理员
// @Tags Organization Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Success 200 {object} models.APIResponse{data=[]models.OrganizationMemberWithUser} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "组织不存在"
// @Router /organizations/{id}/admins [get]
func (h *OrganizationHandler) ListAdmins(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	admins, err := h.orgService.ListOrganizationAdmins(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Organization not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    admins,
	})
}

// PromoteToAdmin godoc
// @Summary 提升为管理员
// @Description 将组织成员提升为管理员（仅所有者可操作）
// @Tags Organization Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param user_id path int true "用户ID"
// @Success 200 {object} models.APIResponse "提升成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /organizations/{id}/admins/{user_id} [post]
func (h *OrganizationHandler) PromoteToAdmin(c *gin.Context) {
	idParam := c.Param("id")
	orgID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	userIDParam := c.Param("user_id")
	targetUserID, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	currentUserID := middleware.MustGetUserID(c)

	// Check if current user is owner
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), orgID, currentUserID, models.OrgRoleOwner)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Only organization owner can promote members to admin",
			},
		})
		return
	}

	err = h.orgService.PromoteToAdmin(c.Request.Context(), orgID, targetUserID, currentUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "PROMOTION_FAILED",
				Message: "Failed to promote member to admin",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Member promoted to admin successfully",
	})
}

// DemoteFromAdmin godoc
// @Summary 降级管理员
// @Description 将管理员降级为普通成员（仅所有者可操作）
// @Tags Organization Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param user_id path int true "用户ID"
// @Success 200 {object} models.APIResponse "降级成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /organizations/{id}/admins/{user_id} [delete]
func (h *OrganizationHandler) DemoteFromAdmin(c *gin.Context) {
	idParam := c.Param("id")
	orgID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	userIDParam := c.Param("user_id")
	targetUserID, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	currentUserID := middleware.MustGetUserID(c)

	// Check if current user is owner
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), orgID, currentUserID, models.OrgRoleOwner)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Only organization owner can demote admins",
			},
		})
		return
	}

	err = h.orgService.DemoteFromAdmin(c.Request.Context(), orgID, targetUserID, currentUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DEMOTION_FAILED",
				Message: "Failed to demote admin",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Admin demoted successfully",
	})
}

// ============================================================================
// Owner Sub-resource Handlers
// ============================================================================

// GetOwner godoc
// @Summary 获取组织所有者
// @Description 获取指定组织的所有者信息
// @Tags Organization Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Success 200 {object} models.APIResponse{data=models.User} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "组织不存在"
// @Router /organizations/{id}/owner [get]
func (h *OrganizationHandler) GetOwner(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	owner, err := h.orgService.GetOrganizationOwner(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Organization or owner not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    owner,
	})
}

// TransferOwnership godoc
// @Summary 转移组织所有权
// @Description 将组织所有权转移给另一个用户（仅所有者可操作）
// @Tags Organization Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param request body models.TransferOwnershipRequest true "转移请求"
// @Success 200 {object} models.APIResponse "转移成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "组织不存在"
// @Router /organizations/{id}/owner [put]
func (h *OrganizationHandler) TransferOwnership(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	currentUserID := middleware.MustGetUserID(c)

	// Check if user is owner
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, currentUserID, models.OrgRoleOwner)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Only organization owner can transfer ownership",
			},
		})
		return
	}

	var req models.TransferOwnershipRequest
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

	err = h.orgService.TransferOwnership(c.Request.Context(), id, req.NewOwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TRANSFER_FAILED",
				Message: "Failed to transfer ownership",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Ownership transferred successfully",
	})
}

// ============================================================================
// Quota Sub-resource Handlers
// ============================================================================

// GetQuota godoc
// @Summary 获取组织配额
// @Description 获取指定组织的配额子资源
// @Tags Organization Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Success 200 {object} models.APIResponse{data=models.OrganizationQuotaResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "组织不存在"
// @Router /organizations/{id}/quota [get]
func (h *OrganizationHandler) GetQuota(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	quota, err := h.orgService.GetOrganizationQuota(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Organization or quota not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    quota,
	})
}

// UpdateQuota godoc
// @Summary 更新组织配额
// @Description 更新指定组织的配额子资源
// @Tags Organization Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param request body models.UpdateOrganizationQuotaRequest true "配额更新请求"
// @Success 200 {object} models.APIResponse{data=models.OrganizationQuotaResponse} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "组织不存在"
// @Router /organizations/{id}/quota [put]
func (h *OrganizationHandler) UpdateQuota(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	var req models.UpdateOrganizationQuotaRequest
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

	quota, err := h.orgService.UpdateOrganizationQuota(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update organization quota",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    quota,
	})
}
