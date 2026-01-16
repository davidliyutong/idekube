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

// EnableMFA godoc
// @Summary 启用MFA
// @Description 发起MFA设置流程，生成二维码
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.MFASetup} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /users/me/mfa/enable [post]
func (h *MFAHandler) EnableMFA(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

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

// VerifyMFASetup godoc
// @Summary 验证MFA设置
// @Description 验证并完成MFA设置
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{code=string} true "MFA验证码"
// @Success 200 {object} models.APIResponse "验证成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /users/me/mfa/verify [post]
func (h *MFAHandler) VerifyMFASetup(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

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

	err := h.mfaService.VerifyAndEnableMFA(c.Request.Context(), userID, req.Code)
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

// DisableMFA godoc
// @Summary 禁用MFA
// @Description 禁用用户的MFA功能
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{password=string} true "用户密码确认"
// @Success 200 {object} models.APIResponse "禁用成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /users/me/mfa/disable [post]
func (h *MFAHandler) DisableMFA(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

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

	err := h.mfaService.DisableMFA(c.Request.Context(), userID, req.Password)
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

// GenerateBackupCodes godoc
// @Summary 生成备份码
// @Description 为MFA生成新的备份码
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=[]string} "生成成功"
// @Failure 400 {object} models.APIResponse "请求失败"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /users/me/mfa/backup-codes [post]
func (h *MFAHandler) GenerateBackupCodes(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

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
