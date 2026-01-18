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
