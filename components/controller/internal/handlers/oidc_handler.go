package handlers

import (
	"net/http"
	"strconv"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// OIDCHandler handles OIDC-related HTTP requests
type OIDCHandler struct {
	oidcService *services.OIDCService
	userService *services.UserService
	jwtManager  *middleware.JWTManager
}

// NewOIDCHandler creates a new OIDC handler
func NewOIDCHandler(
	oidcService *services.OIDCService,
	userService *services.UserService,
	jwtManager *middleware.JWTManager,
) *OIDCHandler {
	return &OIDCHandler{
		oidcService: oidcService,
		userService: userService,
		jwtManager:  jwtManager,
	}
}

// InitiateLogin godoc
// @Summary 发起OIDC登录
// @Description 发起OIDC认证流程，返回认证URL和state
// @Tags Auth
// @Accept json
// @Produce json
// @Param provider path string true "OIDC提供商名称"
// @Success 200 {object} models.APIResponse{data=map[string]string} "成功返回认证URL"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Router /auth/oidc/{provider}/login [get]
func (h *OIDCHandler) InitiateLogin(c *gin.Context) {
	providerName := c.Param("provider")

	authURL, state, err := h.oidcService.GenerateAuthURL(c.Request.Context(), providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "OIDC_ERROR",
				Message: "Failed to initiate OIDC login",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]string{
			"auth_url": authURL,
			"state":    state,
		},
	})
}

// HandleCallback godoc
// @Summary 处理OIDC回调
// @Description 处理OIDC认证提供商的回调，验证用户并返回JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param provider path string true "OIDC提供商名称"
// @Param state query string true "认证state参数"
// @Param code query string true "认证code参数"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}} "认证成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /oidc/{provider}/callback [get]
func (h *OIDCHandler) HandleCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CALLBACK",
				Message: "Missing state or code parameter",
			},
		})
		return
	}

	// Handle callback
	user, provider, err := h.oidcService.HandleCallback(c.Request.Context(), state, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CALLBACK_ERROR",
				Message: "Failed to handle OIDC callback",
				Details: err.Error(),
			},
		})
		return
	}

	// Generate JWT token
	token, _, err := h.jwtManager.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_ERROR",
				Message: "Failed to generate token",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"token":    token,
			"user":     user,
			"provider": provider,
		},
	})
}

// CreateProvider godoc
// @Summary 创建OIDC提供商
// @Description 创建新的OIDC认证提供商配置
// @Tags OIDCManagement
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateOIDCProviderRequest true "OIDC提供商创建请求"
// @Success 201 {object} models.APIResponse{data=models.OIDCProvider} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /oidc/providers [post]
func (h *OIDCHandler) CreateProvider(c *gin.Context) {
	var req models.CreateOIDCProviderRequest
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

	provider, err := h.oidcService.CreateProvider(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATE_FAILED",
				Message: "Failed to create OIDC provider",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    provider,
	})
}

// ListProviders godoc
// @Summary 列出OIDC提供商
// @Description 获取所有OIDC认证提供商的列表
// @Tags OIDCManagement
// @Accept json
// @Produce json
// @Param enabled_only query boolean false "仅显示已启用的提供商"
// @Success 200 {object} models.APIResponse{data=[]models.OIDCProvider} "成功"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /oidc/providers [get]
func (h *OIDCHandler) ListProviders(c *gin.Context) {
	enabledOnly := c.Query("enabled_only") == "true"

	providers, err := h.oidcService.ListProviders(c.Request.Context(), enabledOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "LIST_FAILED",
				Message: "Failed to list OIDC providers",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    providers,
	})
}

// UpdateProvider godoc
// @Summary 更新OIDC提供商
// @Description 更新指定的OIDC认证提供商配置
// @Tags OIDCManagement
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "提供商ID"
// @Param request body models.UpdateOIDCProviderRequest true "OIDC提供商更新请求"
// @Success 200 {object} models.APIResponse{data=models.OIDCProvider} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /oidc/providers/{id} [put]
func (h *OIDCHandler) UpdateProvider(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid provider ID",
			},
		})
		return
	}

	var req models.UpdateOIDCProviderRequest
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

	provider, err := h.oidcService.UpdateProvider(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update OIDC provider",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    provider,
	})
}

// DeleteProvider godoc
// @Summary 删除OIDC提供商
// @Description 删除指定的OIDC认证提供商
// @Tags OIDCManagement
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "提供商ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /auth/oidc/providers/{id} [delete]
func (h *OIDCHandler) DeleteProvider(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid provider ID",
			},
		})
		return
	}

	err = h.oidcService.DeleteProvider(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete OIDC provider",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "OIDC provider deleted successfully",
	})
}
