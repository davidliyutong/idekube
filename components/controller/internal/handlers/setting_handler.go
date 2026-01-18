package handlers

import (
	"net/http"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// SettingHandler handles setting-related HTTP requests
type SettingHandler struct {
	settingService *services.SettingService
}

// NewSettingHandler creates a new setting handler
func NewSettingHandler(settingService *services.SettingService) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
	}
}

// GetAllSettings godoc
// @Summary 获取所有系统设置
// @Description 获取系统的所有配置项（需要admin权限）
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.GetSettingsResponse} "获取成功"
// @Failure 401 {object} models.APIResponse "未授权"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 500 {object} models.APIResponse "服务器错误"
// @Router /settings [get]
func (h *SettingHandler) GetAllSettings(c *gin.Context) {
	settings, err := h.settingService.GetAllSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "QUERY_FAILED",
				Message: "Failed to retrieve settings",
				Details: err.Error(),
			},
		})
		return
	}

	// Convert to response format
	responses := make([]models.SettingResponse, len(settings))
	for i, setting := range settings {
		responses[i] = setting.ToResponse()
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.GetSettingsResponse{
			Settings: responses,
		},
	})
}

// GetSetting godoc
// @Summary 获取单个系统设置
// @Description 根据key获取指定的配置项（需要admin权限）
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "设置的key"
// @Success 200 {object} models.APIResponse{data=models.SettingResponse} "获取成功"
// @Failure 401 {object} models.APIResponse "未授权"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "配置项不存在"
// @Failure 500 {object} models.APIResponse "服务器错误"
// @Router /settings/kv/{key} [get]
func (h *SettingHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")

	setting, err := h.settingService.GetSetting(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Setting not found",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    setting.ToResponse(),
	})
}

// UpdateSetting godoc
// @Summary 更新单个系统设置
// @Description 更新指定key的配置项值（需要admin权限）
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "设置的key"
// @Param request body models.UpdateSettingRequest true "更新请求"
// @Success 200 {object} models.APIResponse{data=models.SettingResponse} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未授权"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "配置项不存在"
// @Failure 500 {object} models.APIResponse "服务器错误"
// @Router /settings/kv/{key} [put]
func (h *SettingHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")

	var req models.UpdateSettingRequest
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

	if err := h.settingService.UpdateSetting(c.Request.Context(), key, req.Value); err != nil {
		// Determine appropriate status code
		statusCode := http.StatusInternalServerError
		errorCode := "UPDATE_FAILED"
		if err.Error() == "setting not found" {
			statusCode = http.StatusNotFound
			errorCode = "NOT_FOUND"
		} else if err.Error()[:13] == "invalid value" {
			statusCode = http.StatusBadRequest
			errorCode = "INVALID_VALUE"
		}

		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    errorCode,
				Message: "Failed to update setting",
				Details: err.Error(),
			},
		})
		return
	}

	// Retrieve updated setting
	setting, err := h.settingService.GetSetting(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "QUERY_FAILED",
				Message: "Setting updated but failed to retrieve",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    setting.ToResponse(),
		Message: "Setting updated successfully",
	})
}

// BatchUpdateSettings godoc
// @Summary 批量更新系统设置
// @Description 批量更新多个配置项（需要admin权限）
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.BatchUpdateSettingsRequest true "批量更新请求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未授权"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 500 {object} models.APIResponse "服务器错误"
// @Router /settings [put]
func (h *SettingHandler) BatchUpdateSettings(c *gin.Context) {
	var req models.BatchUpdateSettingsRequest
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

	if len(req.Settings) == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "No settings provided",
			},
		})
		return
	}

	if err := h.settingService.BatchUpdateSettings(c.Request.Context(), req.Settings); err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "UPDATE_FAILED"
		if err.Error()[:13] == "invalid value" || err.Error()[:7] == "setting" {
			statusCode = http.StatusBadRequest
			errorCode = "INVALID_REQUEST"
		}

		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    errorCode,
				Message: "Failed to update settings",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Settings updated successfully",
	})
}
