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

// ============================================================================
// Main CRUD Operations
// ============================================================================

// List godoc
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
func (h *VolumeHandler) List(c *gin.Context) {
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

// Create godoc
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
func (h *VolumeHandler) Create(c *gin.Context) {
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

// Delete godoc
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
func (h *VolumeHandler) Delete(c *gin.Context) {
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

// ============================================================================
// Sub-resource: Profile
// ============================================================================

// GetProfile godoc
// @Summary 获取存储卷Profile
// @Description 获取指定存储卷的Profile子资源
// @Tags Volume Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Success 200 {object} models.APIResponse{data=models.VolumeProfileResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id}/profile [get]
func (h *VolumeHandler) GetProfile(c *gin.Context) {
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

	profile, err := h.volumeService.GetVolumeProfile(c.Request.Context(), id)
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
		Data:    profile,
	})
}

// UpdateProfile godoc
// @Summary 更新存储卷Profile
// @Description 更新指定存储卷的Profile子资源
// @Tags Volume Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Param request body models.UpdateVolumeProfileRequest true "Profile更新请求"
// @Success 200 {object} models.APIResponse{data=models.VolumeProfileResponse} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id}/profile [put]
func (h *VolumeHandler) UpdateProfile(c *gin.Context) {
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

	var req models.UpdateVolumeProfileRequest
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

	profile, err := h.volumeService.UpdateVolumeProfile(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update volume profile",
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
// Sub-resource: Size MB
// ============================================================================

// GetSizeMB godoc
// @Summary 获取存储卷大小
// @Description 获取指定存储卷的大小（MB）
// @Tags Volume Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Success 200 {object} models.APIResponse{data=models.VolumeSizeMBResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id}/size-mb [get]
func (h *VolumeHandler) GetSizeMB(c *gin.Context) {
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

	sizeMB, err := h.volumeService.GetVolumeSizeMB(c.Request.Context(), id)
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
		Data:    models.VolumeSizeMBResponse{SizeMB: sizeMB},
	})
}

// UpdateSizeMB godoc
// @Summary 更新存储卷大小
// @Description 更新指定存储卷的大小（只能扩容）
// @Tags Volume Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Param request body models.UpdateVolumeSizeRequest true "大小更新请求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id}/size-mb [put]
func (h *VolumeHandler) UpdateSizeMB(c *gin.Context) {
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

	var req models.UpdateVolumeSizeRequest
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

	err = h.volumeService.UpdateVolumeSize(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update volume size",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Volume size updated successfully",
	})
}

// ============================================================================
// Sub-resource: Storage Class
// ============================================================================

// GetStorageClass godoc
// @Summary 获取存储卷存储类
// @Description 获取指定存储卷的存储类
// @Tags Volume Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Success 200 {object} models.APIResponse{data=models.VolumeStorageClassResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id}/storage-class [get]
func (h *VolumeHandler) GetStorageClass(c *gin.Context) {
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

	storageClass, err := h.volumeService.GetVolumeStorageClass(c.Request.Context(), id)
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
		Data:    models.VolumeStorageClassResponse{StorageClass: storageClass},
	})
}

// ============================================================================
// Sub-resource: Access Mode
// ============================================================================

// GetAccessMode godoc
// @Summary 获取存储卷访问模式
// @Description 获取指定存储卷的访问模式
// @Tags Volume Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Success 200 {object} models.APIResponse{data=models.VolumeAccessModeResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id}/access-mode [get]
func (h *VolumeHandler) GetAccessMode(c *gin.Context) {
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

	accessMode, err := h.volumeService.GetVolumeAccessMode(c.Request.Context(), id)
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
		Data:    models.VolumeAccessModeResponse{AccessMode: accessMode},
	})
}

// ============================================================================
// Sub-resource: Owner
// ============================================================================

// GetOwner godoc
// @Summary 获取存储卷所有者
// @Description 获取指定存储卷的所有者信息
// @Tags Volume Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Success 200 {object} models.APIResponse{data=models.VolumeOwnerResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id}/owner [get]
func (h *VolumeHandler) GetOwner(c *gin.Context) {
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

	owner, err := h.volumeService.GetVolumeOwner(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Volume not found",
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
// @Summary 转让存储卷所有权
// @Description 将存储卷所有权转让给其他组织
// @Tags Volume Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Param request body models.TransferVolumeOwnershipRequest true "转让请求"
// @Success 200 {object} models.APIResponse "转让成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id}/owner [put]
func (h *VolumeHandler) TransferOwnership(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

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

	var req models.TransferVolumeOwnershipRequest
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

	err = h.volumeService.TransferVolumeOwnership(c.Request.Context(), id, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TRANSFER_FAILED",
				Message: "Failed to transfer volume ownership",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Volume ownership transferred successfully",
	})
}

// ============================================================================
// Sub-resource: Public
// ============================================================================

// GetPublic godoc
// @Summary 获取存储卷公开状态
// @Description 获取指定存储卷的公开状态
// @Tags Volume Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Success 200 {object} models.APIResponse{data=models.VolumePublicResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id}/public [get]
func (h *VolumeHandler) GetPublic(c *gin.Context) {
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

	isPublic, err := h.volumeService.GetVolumePublic(c.Request.Context(), id)
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
		Data:    models.VolumePublicResponse{IsPublic: isPublic},
	})
}

// SetPublic godoc
// @Summary 更新存储卷公开状态
// @Description 更新指定存储卷的is_public状态
// @Tags Volume Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "存储卷ID"
// @Param request body models.UpdateVolumeIsPublicRequest true "公开状态更新请求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "存储卷不存在"
// @Router /volumes/{id}/public [put]
func (h *VolumeHandler) SetPublic(c *gin.Context) {
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

	var req models.UpdateVolumeIsPublicRequest
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

	err = h.volumeService.UpdateVolumeIsPublic(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update volume is_public",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Volume is_public updated successfully",
	})
}
