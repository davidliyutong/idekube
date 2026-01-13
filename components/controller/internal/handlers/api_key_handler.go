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

// CreateAPIKey creates a new API key
// POST /api/v1/api-keys
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

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

// GetAPIKey retrieves an API key
// GET /api/v1/api-keys/:id
func (h *APIKeyHandler) GetAPIKey(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

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

// ListAPIKeys lists all API keys
// GET /api/v1/api-keys
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

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

// RevokeAPIKey revokes an API key
// DELETE /api/v1/api-keys/:id
func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

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
