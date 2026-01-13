package handlers

import (
	"net/http"
	"strconv"

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
// @Tags 存储卷
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
// @Tags 存储卷
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

// ListVolumes lists volumes by owner
// GET /api/v1/volumes?owner_type=user&owner_id=1
func (h *VolumeHandler) ListVolumes(c *gin.Context) {
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

	volumes, err := h.volumeService.ListVolumesByOwner(c.Request.Context(), models.OwnerType(ownerTypeStr), ownerID)
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

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    volumes,
	})
}

// UpdateVolume updates a volume
// PUT /api/v1/volumes/:id
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

// DeleteVolume deletes a volume
// DELETE /api/v1/volumes/:id
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

// SyncVolumeStatus syncs volume status from Kubernetes
// POST /api/v1/volumes/:id/sync
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
