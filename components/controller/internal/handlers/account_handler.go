package handlers

import (
	"net/http"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// AccountHandler handles account-related HTTP requests
type AccountHandler struct {
	emailService       *services.EmailService
	authService        *services.AuthService
	enableRegistration bool
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(emailService *services.EmailService, authService *services.AuthService, enableRegistration bool) *AccountHandler {
	return &AccountHandler{
		emailService:       emailService,
		authService:        authService,
		enableRegistration: enableRegistration,
	}
}

// Register godoc
// @Summary Register new user
// @Description Create a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration details"
// @Success 201 {object} models.APIResponse{data=models.User}
// @Failure 400 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Router /auth/register [post]
func (h *AccountHandler) Register(c *gin.Context) {
	if !h.enableRegistration {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "REGISTRATION_DISABLED",
				Message: "User registration is disabled",
			},
		})
		return
	}

	var req models.RegisterRequest
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

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		// Check if it's a conflict error (username/email already exists)
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "CONFLICT",
					Message: err.Error(),
				},
			})
			return
		}
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "REGISTRATION_FAILED",
				Message: "Failed to register user",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RequestPasswordResetRequest true "Password reset request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Router /auth/password/request-reset [get]
func (h *AccountHandler) RequestPasswordReset(c *gin.Context) {
	var req models.RequestPasswordResetRequest
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

	err := h.emailService.SendPasswordResetEmail(c.Request.Context(), req.Email)
	if err != nil {
		// Log error but don't expose details to prevent email enumeration
		// Always return success to prevent email enumeration
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "If the email exists, a password reset link will be sent",
		})
		return
	}

	// Always return success to prevent email enumeration
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "If the email exists, a password reset link will be sent",
	})
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset user password using token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Password reset details"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Router /auth/password/reset [post]
func (h *AccountHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
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

	err := h.emailService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "RESET_FAILED",
				Message: "Failed to reset password",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password reset successfully",
	})
}
