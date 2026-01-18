package handlers

import (
	"net/http"
	"strconv"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// TemplateHandler handles template-related HTTP requests
type TemplateHandler struct {
	templateService *services.TemplateService
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(templateService *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
	}
}

// CreateTemplate godoc
// @Summary 创建模板
// @Description 创建新的工作区模板
// @Tags Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateTemplateRequest true "模板创建请求"
// @Success 201 {object} models.APIResponse{data=models.Template} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /templates [post]
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	userRole := middleware.MustGetUserRole(c)

	var req models.CreateTemplateRequest
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

	template, err := h.templateService.CreateTemplate(c.Request.Context(), userID, userRole, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATE_FAILED",
				Message: "Failed to create template",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    template,
	})
}

// GetTemplate godoc
// @Summary 获取模板详情
// @Description 根据ID获取模板的详细信息
// @Tags Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} models.APIResponse{data=models.Template} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id} [get]
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	userID := middleware.MustGetUserID(c)

	// Check access
	hasAccess, err := h.templateService.CheckTemplateAccess(c.Request.Context(), id, userID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Access denied to this template",
			},
		})
		return
	}

	template, err := h.templateService.GetTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Template not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    template,
	})
}

// List godoc
// @Summary 列出模板
// @Description 根据用户权限列出可访问的模板
// @Tags Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id query int false "组织ID过滤"
// @Param all query boolean false "管理员是否列出所有模板"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} models.APIResponse{data=models.PaginatedResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /templates [get]
func (h *TemplateHandler) List(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	userRole := models.UserRole(c.GetString("user_role"))

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

	// Check if requesting all templates (permission already checked by RBAC middleware)
	if c.Query("all") == "true" {
		templates, total, err := h.templateService.ListAllTemplates(c.Request.Context(), &opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to list templates",
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
				Items:      templates,
				Total:      total,
				Page:       opts.Page,
				PageSize:   opts.PageSize,
				TotalPages: totalPages,
			},
		})
		return
	}

	// Parse optional organization_id
	var orgIDs []int64
	orgIDStr := c.Query("organization_id")
	if orgIDStr != "" {
		parsed, err := strconv.ParseInt(orgIDStr, 10, 64)
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
		orgIDs = []int64{parsed}
	}

	// List templates based on user permissions
	templates, total, err := h.templateService.ListTemplates(
		c.Request.Context(),
		userID,
		userRole,
		orgIDs,
		&opts,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list templates",
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
			Items:      templates,
			Total:      total,
			Page:       opts.Page,
			PageSize:   opts.PageSize,
			TotalPages: totalPages,
		},
	})
}

// UpdateTemplate godoc
// @Summary 更新模板
// @Description 更新模板配置信息
// @Tags Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body models.UpdateTemplateRequest true "模板更新请求"
// @Success 200 {object} models.APIResponse{data=models.Template} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id} [put]
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	// Note: Permission to update is already checked by RBACCheckEndpoint middleware
	// This handler only needs to perform the update operation

	var req models.UpdateTemplateRequest
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

	updatedTemplate, err := h.templateService.UpdateTemplate(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update template",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    updatedTemplate,
	})
}

// DeleteTemplate godoc
// @Summary 删除模板
// @Description 删除指定的模板
// @Tags Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id} [delete]
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	// Note: Permission to delete is already checked by RBACCheckEndpoint middleware
	// This handler only needs to perform the delete operation

	err = h.templateService.DeleteTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete template",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Template deleted successfully",
	})
}

// ============================================================================
// Sub-resource Handlers
// ============================================================================

// GetProfile godoc
// @Summary 获取模板Profile
// @Description 获取指定模板的Profile子资源
// @Tags Template Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} models.APIResponse{data=models.TemplateProfileResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id}/profile [get]
func (h *TemplateHandler) GetProfile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	profile, err := h.templateService.GetTemplateProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Template not found",
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
// @Summary 更新模板Profile
// @Description 更新指定模板的Profile子资源
// @Tags Template Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body models.UpdateTemplateProfileRequest true "Profile更新请求"
// @Success 200 {object} models.APIResponse{data=models.TemplateProfileResponse} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id}/profile [put]
func (h *TemplateHandler) UpdateProfile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	var req models.UpdateTemplateProfileRequest
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

	profile, err := h.templateService.UpdateTemplateProfile(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update template profile",
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

// GetQuota godoc
// @Summary 获取模板默认配额
// @Description 获取指定模板的默认配额子资源
// @Tags Template Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} models.APIResponse{data=models.TemplateQuotaResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id}/quota [get]
func (h *TemplateHandler) GetQuota(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	quota, err := h.templateService.GetTemplateQuota(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Template not found",
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
// @Summary 更新模板默认配额
// @Description 更新指定模板的默认配额子资源
// @Tags Template Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body models.UpdateTemplateQuotaRequest true "配额更新请求"
// @Success 200 {object} models.APIResponse{data=models.TemplateQuotaResponse} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id}/quota [put]
func (h *TemplateHandler) UpdateQuota(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	var req models.UpdateTemplateQuotaRequest
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

	quota, err := h.templateService.UpdateTemplateQuota(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update template quota",
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

// SetPublic godoc
// @Summary 更新模板公开状态
// @Description 更新指定模板的is_public状态
// @Tags Template Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body models.UpdateTemplateIsPublicRequest true "公开状态更新请求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id}/public [put]
func (h *TemplateHandler) SetPublic(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	var req models.UpdateTemplateIsPublicRequest
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

	err = h.templateService.UpdateTemplateIsPublic(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update template is_public",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Template is_public updated successfully",
	})
}

// UpdateTemplateYAML godoc
// @Summary 更新模板YAML
// @Description 更新指定模板的YAML内容
// @Tags Template Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body models.UpdateTemplateYAMLRequest true "YAML更新请求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id}/template-yaml [put]
func (h *TemplateHandler) UpdateTemplateYAML(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	var req models.UpdateTemplateYAMLRequest
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

	err = h.templateService.UpdateTemplateYAML(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update template YAML",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Template YAML updated successfully",
	})
}

// UpdateImageRef godoc
// @Summary 更新模板镜像引用
// @Description 更新指定模板的镜像引用
// @Tags Template Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body models.UpdateTemplateImageRefRequest true "镜像引用更新请求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id}/image-ref [put]
func (h *TemplateHandler) UpdateImageRef(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	var req models.UpdateTemplateImageRefRequest
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

	err = h.templateService.UpdateTemplateImageRef(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update template image_ref",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Template image_ref updated successfully",
	})
}

// GetImageRef godoc
// @Summary 获取模板镜像引用
// @Description 获取指定模板的镜像引用
// @Tags Template Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} models.APIResponse{data=string} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id}/image-ref [get]
func (h *TemplateHandler) GetImageRef(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	imageRef, err := h.templateService.GetTemplateImageRef(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Template not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    imageRef,
	})
}

// GetTemplateYAML godoc
// @Summary 获取模板YAML
// @Description 获取指定模板的YAML内容
// @Tags Template Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} models.APIResponse{data=string} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id}/template-yaml [get]
func (h *TemplateHandler) GetTemplateYAML(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	templateYAML, err := h.templateService.GetTemplateYAML(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Template not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    templateYAML,
	})
}

// GetPublic godoc
// @Summary 获取模板公开状态
// @Description 获取指定模板的公开状态
// @Tags Template Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} models.APIResponse{data=bool} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "模板不存在"
// @Router /templates/{id}/public [get]
func (h *TemplateHandler) GetPublic(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid template ID",
			},
		})
		return
	}

	isPublic, err := h.templateService.GetTemplateIsPublic(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Template not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    isPublic,
	})
}
