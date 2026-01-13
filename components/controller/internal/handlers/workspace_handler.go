package handlers

import (
	"net/http"
	"strconv"

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

// ListWorkspaces lists workspaces by owner
// GET /api/v1/workspaces?owner_type=user&owner_id=1
func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
	ownerTypeStr := c.Query("owner_type")
	ownerIDStr := c.Query("owner_id")

	if ownerTypeStr == "" || ownerIDStr == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "owner_type and owner_id are required",
			},
		})
		return
	}

	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid owner_id",
			},
		})
		return
	}

	workspaces, err := h.workspaceService.ListWorkspacesByOwner(c.Request.Context(), models.OwnerType(ownerTypeStr), ownerID)
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

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    workspaces,
	})
}

// UpdateWorkspace updates a workspace
// PUT /api/v1/workspaces/:id
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

// DeleteWorkspace deletes a workspace
// DELETE /api/v1/workspaces/:id
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

// StartWorkspace starts a stopped workspace
// POST /api/v1/workspaces/:id/start
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

// StopWorkspace stops a running workspace
// POST /api/v1/workspaces/:id/stop
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

// AttachVolume attaches a volume to a workspace
// POST /api/v1/workspaces/:id/volumes/:volume_id
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

// DetachVolume detaches a volume from a workspace
// DELETE /api/v1/workspaces/:id/volumes/:volume_id
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
