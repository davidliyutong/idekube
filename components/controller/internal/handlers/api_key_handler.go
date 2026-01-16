package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// APIKeyHandler handles API key-related HTTP requests
type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// CreateAPIKey godoc
// @Summary 创建API密钥
// @Description 为当前用户创建新的API密钥
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{name=string,scopes=[]string,expires_at=int64} true "API密钥创建请求"
// @Success 201 {object} models.APIResponse{data=map[string]interface{}} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /api-keys [post]
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req struct {
		Name      string   `json:"name" binding:"required"`
		Scopes    []string `json:"scopes"`
		ExpiresAt *int64   `json:"expires_at"` // Unix timestamp
	}
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

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t := time.Unix(*req.ExpiresAt, 0)
		expiresAt = &t
	}

	apiKey, plainKey, err := h.apiKeyService.CreateAPIKey(
		c.Request.Context(),
		userID,
		req.Name,
		req.Scopes,
		expiresAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "API_KEY_CREATE_FAILED",
				Message: "Failed to create API key",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":         apiKey.ID,
			"name":       apiKey.Name,
			"key":        plainKey,
			"scopes":     apiKey.Scopes,
			"expires_at": apiKey.ExpiresAt,
			"created_at": apiKey.CreatedAt,
		},
		Message: "API key created successfully. Save it securely - it won't be shown again!",
	})
}

// GetAPIKey godoc
// @Summary 获取API密钥详情
// @Description 根据ID获取API密钥的详细信息
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "API密钥ID"
// @Success 200 {object} models.APIResponse{data=models.APIKey} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "未找到"
// @Router /api-keys/{id} [get]
func (h *APIKeyHandler) GetAPIKey(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid API key ID",
			},
		})
		return
	}

	apiKey, err := h.apiKeyService.GetAPIKey(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "API_KEY_NOT_FOUND",
				Message: "API key not found",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    apiKey,
	})
}

// ListAPIKeys godoc
// @Summary 列出API密钥
// @Description 获取当前用户的所有API密钥
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=map[string]interface{}} "成功"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /api-keys [get]
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	apiKeys, err := h.apiKeyService.ListAPIKeys(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "API_KEY_LIST_FAILED",
				Message: "Failed to list API keys",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"api_keys": apiKeys,
			"total":    len(apiKeys),
		},
	})
}

// RevokeAPIKey godoc
// @Summary 撤销API密钥
// @Description 撤销指定的API密钥
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "API密钥ID"
// @Success 200 {object} models.APIResponse "撤销成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /api-keys/{id} [delete]
func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid API key ID",
			},
		})
		return
	}

	err = h.apiKeyService.RevokeAPIKey(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "API_KEY_REVOKE_FAILED",
				Message: "Failed to revoke API key",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "API key revoked successfully",
	})
}
