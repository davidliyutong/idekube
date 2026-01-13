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

// InitiateLogin initiates OIDC login flow
// GET /api/v1/auth/oidc/:provider/login
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

// HandleCallback handles OIDC callback
// GET /api/v1/auth/oidc/:provider/callback
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

// CreateProvider creates a new OIDC provider
// POST /api/v1/auth/oidc/providers
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

// ListProviders lists OIDC providers
// GET /api/v1/auth/oidc/providers
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

// UpdateProvider updates an OIDC provider
// PUT /api/v1/auth/oidc/providers/:id
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

// DeleteProvider deletes an OIDC provider
// DELETE /api/v1/auth/oidc/providers/:id
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
