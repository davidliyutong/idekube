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
	workspaceService *services.WorkspaceService
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(workspaceService *services.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
	}
}

// CreateWorkspace godoc
// @Summary 创建工作区
// @Description 创建新的工作区实例
// @Tags 工作区
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateWorkspaceRequest true "工作区创建请求"
// @Success 201 {object} models.APIResponse{data=models.Workspace} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /workspaces [post]
func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
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

	workspace, err := h.workspaceService.CreateWorkspace(c.Request.Context(), &req)
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

// GetWorkspace godoc
// @Summary 获取工作区详情
// @Description 根据ID获取工作区的详细信息
// @Tags 工作区
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse{data=models.Workspace} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "工作区不存在"
// @Router /workspaces/{id} [get]
func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
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

	workspace, err := h.workspaceService.GetWorkspace(c.Request.Context(), id)
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
		Data:    workspace,
	})
}

// ListWorkspaces godoc
// @Summary 列出工作区
// @Description 根据用户权限列出可访问的工作区
// @Tags 工作区
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
func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
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

// UpdateWorkspace godoc
// @Summary 更新工作区
// @Description 更新工作区配置信息
// @Tags 工作区
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Param request body models.UpdateWorkspaceRequest true "工作区更新请求"
// @Success 200 {object} models.APIResponse{data=models.Workspace} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /workspaces/{id} [put]
func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
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

	var req models.UpdateWorkspaceRequest
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

	workspace, err := h.workspaceService.UpdateWorkspace(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update workspace",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    workspace,
	})
}

// DeleteWorkspace godoc
// @Summary 删除工作区
// @Description 删除指定的工作区
// @Tags 工作区
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /workspaces/{id} [delete]
func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
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

// StartWorkspace godoc
// @Summary 启动工作区
// @Description 启动已停止的工作区
// @Tags 工作区
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse{data=models.Workspace} "启动成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /workspaces/{id}/start [post]
func (h *WorkspaceHandler) StartWorkspace(c *gin.Context) {
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

	err = h.workspaceService.StartWorkspace(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "START_FAILED",
				Message: "Failed to start workspace",
				Details: err.Error(),
			},
		})
		return
	}

	workspace, _ := h.workspaceService.GetWorkspace(c.Request.Context(), id)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    workspace,
	})
}

// StopWorkspace godoc
// @Summary 停止工作区
// @Description 停止正在运行的工作区
// @Tags 工作区
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Success 200 {object} models.APIResponse{data=models.Workspace} "停止成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /workspaces/{id}/stop [post]
func (h *WorkspaceHandler) StopWorkspace(c *gin.Context) {
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

	err = h.workspaceService.StopWorkspace(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "STOP_FAILED",
				Message: "Failed to stop workspace",
				Details: err.Error(),
			},
		})
		return
	}

	workspace, _ := h.workspaceService.GetWorkspace(c.Request.Context(), id)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    workspace,
	})
}

// AttachVolume godoc
// @Summary 挂载存储卷
// @Description 将存储卷挂载到工作区
// @Tags 工作区
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Param volume_id path int true "存储卷ID"
// @Param request body object{mount_path=string} true "挂载路径"
// @Success 200 {object} models.APIResponse "挂载成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /workspaces/{id}/volumes/{volume_id} [post]
func (h *WorkspaceHandler) AttachVolume(c *gin.Context) {
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

	volumeIDParam := c.Param("volume_id")
	volumeID, err := strconv.ParseInt(volumeIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid volume ID",
			},
		})
		return
	}

	var req struct {
		MountPath string `json:"mount_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "mount_path is required",
				Details: err.Error(),
			},
		})
		return
	}

	err = h.workspaceService.AttachVolume(c.Request.Context(), id, volumeID, req.MountPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ATTACH_FAILED",
				Message: "Failed to attach volume",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Volume attached successfully",
	})
}

// DetachVolume godoc
// @Summary 卸载存储卷
// @Description 从工作区卸载存储卷
// @Tags 工作区
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "工作区ID"
// @Param volume_id path int true "存储卷ID"
// @Success 200 {object} models.APIResponse "卸载成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /workspaces/{id}/volumes/{volume_id} [delete]
func (h *WorkspaceHandler) DetachVolume(c *gin.Context) {
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

	volumeIDParam := c.Param("volume_id")
	volumeID, err := strconv.ParseInt(volumeIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid volume ID",
			},
		})
		return
	}

	err = h.workspaceService.DetachVolume(c.Request.Context(), id, volumeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DETACH_FAILED",
				Message: "Failed to detach volume",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Volume detached successfully",
	})
}
