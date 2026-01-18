package handlers

import (
	"net/http"
	"strconv"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// WorkspaceHandler handles workspace-related HTTP requests
type WorkspaceHandler struct {
	workspaceService         *services.WorkspaceService
	workspaceTransferService *services.WorkspaceTransferService
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(
	workspaceService *services.WorkspaceService,
	workspaceTransferService *services.WorkspaceTransferService,
) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService:         workspaceService,
		workspaceTransferService: workspaceTransferService,
	}
}

// ============================================================================
// Main CRUD Operations
// ============================================================================

// List godoc
// @Summary 列出工作区
// @Description 根据用户权限列出可访问的工作区
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id query int false "组织ID过滤"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} models.APIResponse{data=models.PaginatedResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /workspaces [get]
func (h *WorkspaceHandler) List(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	userRole := middleware.MustGetUserRole(c)

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

	// Check if filtering by organization
	orgIDStr := c.Query("organization_id")
	if orgIDStr != "" {
		orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_REQUEST",
					Message: "Invalid organization_id",
				},
			})
			return
		}

		workspaces, total, err := h.workspaceService.ListWorkspacesByOrganization(c.Request.Context(), orgID, &opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to list workspaces",
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
				Items:      workspaces,
				Total:      total,
				Page:       opts.Page,
				PageSize:   opts.PageSize,
				TotalPages: totalPages,
			},
		})
		return
	}

	// List workspaces based on user role
	workspaces, total, err := h.workspaceService.ListWorkspaces(
		c.Request.Context(),
		userID,
		userRole,
		nil, // orgRole - can be extended later
		&opts,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list workspaces",
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
			Items:      workspaces,
			Total:      total,
			Page:       opts.Page,
			PageSize:   opts.PageSize,
			TotalPages: totalPages,
		},
	})
}

// Create godoc
// @Summary 创建工作区
// @Description 创建新的工作区实例
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateWorkspaceRequest true "工作区创建请求"
// @Success 201 {object} models.APIResponse{data=models.Workspace} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /workspaces [post]
func (h *WorkspaceHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req models.CreateWorkspaceRequest
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

	workspace, err := h.workspaceService.CreateWorkspace(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATE_FAILED",
				Message: "Failed to create workspace",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    workspace,
	})
}

// Delete godoc
// @Summary 删除工作区
// @Description 删除指定的工作区
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /workspaces/{id} [delete]
func (h *WorkspaceHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	err = h.workspaceService.DeleteWorkspace(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete workspace",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Workspace deleted successfully",
	})
}

// ============================================================================
// Sub-resource: Profile
// ============================================================================

// GetProfile godoc
// @Summary 获取工作区Profile
// @Description 获取指定工作区的Profile子资源
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse{data=models.WorkspaceProfileResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/profile [get]
func (h *WorkspaceHandler) GetProfile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	profile, err := h.workspaceService.GetWorkspaceProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Workspace not found",
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
// @Summary 更新工作区Profile
// @Description 更新指定工作区的Profile子资源
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Param request body models.UpdateWorkspaceProfileRequest true "Profile更新请求"
// @Success 200 {object} models.APIResponse{data=models.WorkspaceProfileResponse} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/profile [put]
func (h *WorkspaceHandler) UpdateProfile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	var req models.UpdateWorkspaceProfileRequest
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

	profile, err := h.workspaceService.UpdateWorkspaceProfile(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update workspace profile",
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
// Sub-resource: Template
// ============================================================================

// GetTemplate godoc
// @Summary 获取工作区模板信息
// @Description 获取指定工作区使用的模板信息
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse{data=models.Template} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/template [get]
func (h *WorkspaceHandler) GetTemplate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	template, err := h.workspaceService.GetWorkspaceTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Workspace or template not found",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    template,
	})
}

// ============================================================================
// Sub-resource: Volume Mounts
// ============================================================================

// ListVolumeMounts godoc
// @Summary 获取工作区的Volume列表
// @Description 获取指定工作区挂载的所有Volume
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse{data=[]models.WorkspaceVolumeResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/volumes [get]
func (h *WorkspaceHandler) ListVolumeMounts(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	volumes, err := h.workspaceService.ListWorkspaceVolumes(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Workspace not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    volumes,
	})
}

// UpdateVolumeMounts godoc
// @Summary 更新工作区的Volume挂载
// @Description 批量更新指定工作区的Volume挂载列表
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Param request body models.UpdateVolumeMountsRequest true "Volume挂载更新请求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/volumes [put]
func (h *WorkspaceHandler) UpdateVolumeMounts(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	var req models.UpdateVolumeMountsRequest
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

	err = h.workspaceService.UpdateWorkspaceVolumeMounts(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update volume mounts",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Volume mounts updated successfully",
	})
}

// ============================================================================
// Sub-resource: Quota
// ============================================================================

// GetQuota godoc
// @Summary 获取工作区配额
// @Description 获取指定工作区的配额子资源
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse{data=models.WorkspaceQuotaResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/quota [get]
func (h *WorkspaceHandler) GetQuota(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	quota, err := h.workspaceService.GetWorkspaceQuota(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Workspace not found",
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
// @Summary 更新工作区配额
// @Description 更新指定工作区的配额子资源
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Param request body models.UpdateWorkspaceQuotaRequest true "配额更新请求"
// @Success 200 {object} models.APIResponse{data=models.WorkspaceQuotaResponse} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/quota [put]
func (h *WorkspaceHandler) UpdateQuota(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	var req models.UpdateWorkspaceQuotaRequest
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

	quota, err := h.workspaceService.UpdateWorkspaceQuota(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update workspace quota",
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

// ============================================================================
// Sub-resource: Public
// ============================================================================

// GetPublic godoc
// @Summary 获取工作区公开状态
// @Description 获取指定工作区的公开状态
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse{data=models.WorkspacePublicResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/public [get]
func (h *WorkspaceHandler) GetPublic(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	isPublic, err := h.workspaceService.GetWorkspacePublic(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Workspace not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    models.WorkspacePublicResponse{IsPublic: isPublic},
	})
}

// SetPublic godoc
// @Summary 更新工作区公开状态
// @Description 更新指定工作区的is_public状态
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Param request body models.UpdateWorkspaceIsPublicRequest true "公开状态更新请求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/public [put]
func (h *WorkspaceHandler) SetPublic(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	var req models.UpdateWorkspaceIsPublicRequest
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

	err = h.workspaceService.UpdateWorkspaceIsPublic(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update workspace is_public",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Workspace is_public updated successfully",
	})
}

// ============================================================================
// Sub-resource: Owner
// ============================================================================

// GetOwner godoc
// @Summary 获取工作区所有者
// @Description 获取指定工作区的所有者信息
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse{data=models.WorkspaceOwnerResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/owner [get]
func (h *WorkspaceHandler) GetOwner(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	owner, err := h.workspaceService.GetWorkspaceOwner(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Workspace not found",
				Details: err.Error(),
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
// @Summary 转让工作区所有权
// @Description 将工作区所有权转让给其他组织
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Param request body models.TransferWorkspaceOwnershipRequest true "转让请求"
// @Success 200 {object} models.APIResponse "转让成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/owner [put]
func (h *WorkspaceHandler) TransferOwnership(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	var req models.TransferWorkspaceOwnershipRequest
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

	err = h.workspaceService.TransferWorkspaceOwnership(c.Request.Context(), id, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TRANSFER_FAILED",
				Message: "Failed to transfer workspace ownership",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Workspace ownership transferred successfully",
	})
}

// ============================================================================
// Sub-resource: State
// ============================================================================

// GetCurrentState godoc
// @Summary 获取工作区当前状态
// @Description 获取工作区的当前运行状态
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse{data=models.WorkspaceStateResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/state [get]
func (h *WorkspaceHandler) GetCurrentState(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	state, err := h.workspaceService.GetWorkspaceState(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Workspace not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    state,
	})
}

// UpdateTargetState godoc
// @Summary 更新工作区目标状态
// @Description 更新工作区的目标状态(启动/停止)
// @Tags Workspace Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Param request body models.UpdateWorkspaceStateRequest true "状态更新请求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/state [put]
func (h *WorkspaceHandler) UpdateTargetState(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	var req models.UpdateWorkspaceStateRequest
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

	err = h.workspaceService.UpdateWorkspaceState(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update workspace state",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Workspace state updated successfully",
	})
}

// ============================================================================
// Backward Compatibility - Transfer Routes
// ============================================================================

// InitiateTransfer godoc (kept for backward compatibility)
// @Summary 发起工作区转让
// @Description 工作区所有者发起将工作区转让给其他用户的请求
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Param request body models.CreateWorkspaceTransferRequest true "转让请求"
// @Success 201 {object} models.APIResponse{data=models.WorkspaceTransfer} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id}/transfer [post]
func (h *WorkspaceHandler) InitiateTransfer(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	idParam := c.Param("id")
	workspaceID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid workspace ID",
			},
		})
		return
	}

	var req models.CreateWorkspaceTransferRequest
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

	transfer, err := h.workspaceTransferService.CreateTransfer(
		c.Request.Context(),
		workspaceID,
		userID,
		&req,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TRANSFER_FAILED",
				Message: "Failed to create transfer request",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    transfer,
	})
}

// RespondToTransfer godoc
// @Summary 响应工作区转让请求
// @Description 接收方接受或拒绝工作区转让请求
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transfer_id path int true "转让请求ID"
// @Param request body models.RespondWorkspaceTransferRequest true "响应请求"
// @Success 200 {object} models.APIResponse{data=models.WorkspaceTransfer} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "转让请求不存在"
// @Router /workspace-transfers/{transfer_id}/respond [post]
func (h *WorkspaceHandler) RespondToTransfer(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	transferIDParam := c.Param("transfer_id")
	transferID, err := strconv.ParseInt(transferIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid transfer ID",
			},
		})
		return
	}

	var req models.RespondWorkspaceTransferRequest
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

	transfer, err := h.workspaceTransferService.RespondToTransfer(
		c.Request.Context(),
		transferID,
		userID,
		&req,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "RESPOND_FAILED",
				Message: "Failed to respond to transfer request",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    transfer,
	})
}

// CancelTransfer godoc
// @Summary 取消工作区转让请求
// @Description 发起方取消待处理的工作区转让请求
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transfer_id path int true "转让请求ID"
// @Success 200 {object} models.APIResponse "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "转让请求不存在"
// @Router /workspace-transfers/{transfer_id}/cancel [post]
func (h *WorkspaceHandler) CancelTransfer(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	transferIDParam := c.Param("transfer_id")
	transferID, err := strconv.ParseInt(transferIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid transfer ID",
			},
		})
		return
	}

	err = h.workspaceTransferService.CancelTransfer(
		c.Request.Context(),
		transferID,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CANCEL_FAILED",
				Message: "Failed to cancel transfer request",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Transfer request cancelled successfully",
	})
}

// ListPendingTransfers godoc
// @Summary 列出待处理的工作区转让请求
// @Description 获取当前用户收到的所有待处理转让请求
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=[]models.WorkspaceTransfer} "成功"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /workspace-transfers/pending [get]
func (h *WorkspaceHandler) ListPendingTransfers(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	transfers, err := h.workspaceTransferService.ListPendingTransfersForUser(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "LIST_FAILED",
				Message: "Failed to list pending transfers",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    transfers,
	})
}

// GetTransfer godoc
// @Summary 获取转让请求详情
// @Description 获取特定转让请求的详细信息
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transfer_id path int true "转让请求ID"
// @Success 200 {object} models.APIResponse{data=models.WorkspaceTransfer} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "转让请求不存在"
// @Router /workspace-transfers/{transfer_id} [get]
func (h *WorkspaceHandler) GetTransfer(c *gin.Context) {
	transferIDParam := c.Param("transfer_id")
	transferID, err := strconv.ParseInt(transferIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid transfer ID",
			},
		})
		return
	}

	transfer, err := h.workspaceTransferService.GetTransfer(
		c.Request.Context(),
		transferID,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Transfer request not found",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    transfer,
	})
}
