package handlers

import (
	"net/http"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// MFAHandler handles MFA-related HTTP requests
type MFAHandler struct {
	mfaService *services.MFAService
}

// NewMFAHandler creates a new MFA handler
func NewMFAHandler(mfaService *services.MFAService) *MFAHandler {
	return &MFAHandler{
		mfaService: mfaService,
	}
}

// EnableMFA initiates MFA setup
// POST /api/v1/users/me/mfa/enable
func (h *MFAHandler) EnableMFA(c *gin.Context) {
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

	setup, err := h.mfaService.EnableMFA(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MFA_ENABLE_FAILED",
				Message: "Failed to enable MFA",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    setup,
		Message: "Scan the QR code with your authenticator app and verify with a code",
	})
}

// VerifyMFASetup verifies and completes MFA setup
// POST /api/v1/users/me/mfa/verify
func (h *MFAHandler) VerifyMFASetup(c *gin.Context) {
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
		Code string `json:"code" binding:"required"`
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

	err = h.mfaService.VerifyAndEnableMFA(c.Request.Context(), userID, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VERIFICATION_FAILED",
				Message: "Failed to verify MFA code",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "MFA enabled successfully",
	})
}

// DisableMFA disables MFA
// POST /api/v1/users/me/mfa/disable
func (h *MFAHandler) DisableMFA(c *gin.Context) {
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
		Password string `json:"password" binding:"required"`
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

	err = h.mfaService.DisableMFA(c.Request.Context(), userID, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MFA_DISABLE_FAILED",
				Message: "Failed to disable MFA",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "MFA disabled successfully",
	})
}

// GenerateBackupCodes generates new backup codes
// POST /api/v1/users/me/mfa/backup-codes
func (h *MFAHandler) GenerateBackupCodes(c *gin.Context) {
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

	codes, err := h.mfaService.GenerateBackupCodes(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "BACKUP_CODES_FAILED",
				Message: "Failed to generate backup codes",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"backup_codes": codes,
		},
		Message: "Save these backup codes securely. They can only be used once.",
	})
}
