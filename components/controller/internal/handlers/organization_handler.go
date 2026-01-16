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

// CreateOrganization godoc
// @Summary 创建组织
// @Description 创建新的组织
// @Tags 组织
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateOrganizationRequest true "组织创建请求"
// @Success 201 {object} models.APIResponse{data=models.Organization} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
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

// GetOrganization godoc
// @Summary 获取组织详情
// @Description 根据ID获取组织的详细信息
// @Tags 组织
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param with_members query boolean false "是否包含成员信息"
// @Success 200 {object} models.APIResponse{data=models.Organization} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "未找到"
// @Router /organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
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

	// Check if user wants detailed view with members
	withMembers := c.Query("with_members") == "true"

	if withMembers {
		orgWithMembers, err := h.orgService.GetOrganizationWithMembers(c.Request.Context(), id)
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
			Data:    orgWithMembers,
		})
	} else {
		org, err := h.orgService.GetOrganization(c.Request.Context(), id)
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
			Data:    org,
		})
	}
}

// ListUserOrganizations godoc
// @Summary 列出用户的组织
// @Description 获取当前用户所属的所有组织
// @Tags 组织
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
func (h *OrganizationHandler) ListUserOrganizations(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	userRole := models.UserRole(c.GetString("user_role"))

	// Check if admin wants to list all organizations
	if (userRole == models.UserRoleAdmin || userRole == models.UserRoleSuperAdmin) && c.Query("all") == "true" {
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

// UpdateOrganization godoc
// @Summary 更新组织
// @Description 更新组织信息
// @Tags 组织
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param request body models.UpdateOrganizationRequest true "组织更新请求"
// @Success 200 {object} models.APIResponse{data=models.Organization} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
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

	var req models.UpdateOrganizationRequest
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

	org, err := h.orgService.UpdateOrganization(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update organization",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    org,
	})
}

// DeleteOrganization godoc
// @Summary 删除组织
// @Description 删除指定的组织（仅所有者可操作）
// @Tags 组织
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
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

// AddMember godoc
// @Summary 添加组织成员
// @Description 向组织添加新成员（需要管理员权限）
// @Tags 组织成员
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
func (h *OrganizationHandler) AddMember(c *gin.Context) {
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
// @Tags 组织成员
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
// @Tags 组织成员
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

// ListAllOrganizations godoc
// @Summary 列出所有组织
// @Description 获取系统中所有组织（仅管理员）
// @Tags 组织管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} models.APIResponse{data=models.PaginatedResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /admin/organizations [get]
func (h *OrganizationHandler) ListAllOrganizations(c *gin.Context) {
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
}

// PromoteToAdmin godoc
// @Summary 提升为管理员
// @Description 将组织成员提升为管理员（仅所有者可操作）
// @Tags 组织成员
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
// @Tags 组织成员
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

// SearchUsers godoc
// @Summary 搜索用户
// @Description 搜索用户以邀请加入组织（需要管理员权限）
// @Tags 组织成员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "组织ID"
// @Param q query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} models.APIResponse{data=models.PaginatedResponse} "搜索成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /organizations/{id}/search-users [get]
func (h *OrganizationHandler) SearchUsers(c *gin.Context) {
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

	currentUserID := middleware.MustGetUserID(c)

	// Check if current user has admin permission
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), orgID, currentUserID, models.OrgRoleAdmin)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Only organization admins can search users",
			},
		})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Query parameter 'q' is required",
			},
		})
		return
	}

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

	// Use the existing user service method through organization service
	// This needs to be implemented in OrganizationService to call UserService
	users, total, err := h.orgService.SearchUsersForInvite(c.Request.Context(), orgID, query, &opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to search users",
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
			Items:      users,
			Total:      total,
			Page:       opts.Page,
			PageSize:   opts.PageSize,
			TotalPages: totalPages,
		},
	})
}
