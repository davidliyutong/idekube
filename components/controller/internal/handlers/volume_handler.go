package handlers

import (
	"net/http"
	"strconv"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// VolumeHandler handles volume-related HTTP requests
type VolumeHandler struct {
	volumeService *services.VolumeService
}

// NewVolumeHandler creates a new volume handler
func NewVolumeHandler(volumeService *services.VolumeService) *VolumeHandler {
	return &VolumeHandler{
		volumeService: volumeService,
	}
}

// CreateVolume godoc
// @Summary 创建存储卷
// @Description 创建新的持久化存储卷
// @Tags Volumes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateVolumeRequest true "存储卷创建请求"
// @Success 201 {object} models.APIResponse{data=models.Volume} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /volumes [post]
func (h *VolumeHandler) CreateVolume(c *gin.Context) {
	var req models.CreateVolumeRequest
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

	volume, err := h.volumeService.CreateVolume(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATE_FAILED",
				Message: "Failed to create volume",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    volume,
	})
}

// GetVolume godoc
// @Summary 获取存储卷详情
// @Description 根据ID获取存储卷的详细信息
// @Tags Volumes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Success 200 {object} models.APIResponse{data=models.Volume} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id} [get]
func (h *VolumeHandler) GetVolume(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
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

	volume, err := h.volumeService.GetVolume(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Volume not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    volume,
	})
}

// ListVolumes godoc
// @Summary 列出存储卷
// @Description 根据用户权限列出可访问的存储卷
// @Tags Volumes
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
// @Router /volumes [get]
func (h *VolumeHandler) ListVolumes(c *gin.Context) {
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

		volumes, total, err := h.volumeService.ListVolumesByOrganization(c.Request.Context(), orgID, &opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to list volumes",
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
				Items:      volumes,
				Total:      total,
				Page:       opts.Page,
				PageSize:   opts.PageSize,
				TotalPages: totalPages,
			},
		})
		return
	}

	// List volumes based on user role
	volumes, total, err := h.volumeService.ListVolumes(
		c.Request.Context(),
		userID,
		userRole,
		&opts,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list volumes",
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
			Items:      volumes,
			Total:      total,
			Page:       opts.Page,
			PageSize:   opts.PageSize,
			TotalPages: totalPages,
		},
	})
}

// UpdateVolume godoc
// @Summary 更新存储卷
// @Description 更新存储卷的配置信息
// @Tags Volumes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Param request body models.UpdateVolumeRequest true "存储卷更新请求"
// @Success 200 {object} models.APIResponse{data=models.Volume} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /volumes/{id} [put]
func (h *VolumeHandler) UpdateVolume(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
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

	var req models.UpdateVolumeRequest
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

	volume, err := h.volumeService.UpdateVolume(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update volume",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    volume,
	})
}

// DeleteVolume godoc
// @Summary 删除存储卷
// @Description 删除指定的存储卷
// @Tags Volumes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /volumes/{id} [delete]
func (h *VolumeHandler) DeleteVolume(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
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

	err = h.volumeService.DeleteVolume(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete volume",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Volume deleted successfully",
	})
}

// SyncVolumeStatus godoc
// @Summary 同步存储卷状态
// @Description 从Kubernetes同步存储卷的实际状态
// @Tags Volumes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Success 200 {object} models.APIResponse{data=models.Volume} "同步成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /volumes/{id}/sync [post]
func (h *VolumeHandler) SyncVolumeStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
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

	err = h.volumeService.SyncVolumeStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SYNC_FAILED",
				Message: "Failed to sync volume status",
				Details: err.Error(),
			},
		})
		return
	}

	volume, _ := h.volumeService.GetVolume(c.Request.Context(), id)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    volume,
	})
}
