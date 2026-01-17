package handlers

import (
	"net/http"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// EmailHandler handles email-related HTTP requests
type EmailHandler struct {
	emailService *services.EmailService
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(emailService *services.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

// VerifyEmail godoc
// @Summary 验证邮箱
// @Description 使用token验证用户邮箱
// @Tags Auth
// @Accept json
// @Produce json
// @Param token query string true "验证token"
// @Success 200 {object} models.APIResponse "验证成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Router /auth/verify-email [get]
func (h *EmailHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_TOKEN",
				Message: "Token is required",
			},
		})
		return
	}

	err := h.emailService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VERIFICATION_FAILED",
				Message: "Failed to verify email",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Email verified successfully",
	})
}

// RequestPasswordReset godoc
// @Summary 请求密码重置
// @Description 发送密码重置邮件
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body object{email=string} true "密码重置请求"
// @Success 200 {object} models.APIResponse "请求已提交"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /auth/request-password-reset [post]
func (h *EmailHandler) RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
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

	err := h.emailService.SendPasswordResetEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "EMAIL_SEND_FAILED",
				Message: "Failed to send email",
				Details: err.Error(),
			},
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
// @Summary 重置密码
// @Description 使用token重置用户密码
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body object{token=string,new_password=string} true "密码重置请求"
// @Success 200 {object} models.APIResponse "重置成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Router /auth/reset-password [post]
func (h *EmailHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
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
