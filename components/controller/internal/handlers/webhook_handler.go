package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// WebhookHandler handles webhook-related HTTP requests
type WebhookHandler struct {
	webhookService *services.WebhookService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(webhookService *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// CreateWebhook godoc
// @Summary 创建Webhook
// @Description 创建新的Webhook订阅
// @Tags Webhook
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Webhook true "Webhook创建请求"
// @Success 201 {object} models.APIResponse{data=models.Webhook} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /webhooks [post]
func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var webhook models.Webhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
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

	err := h.webhookService.CreateWebhook(c.Request.Context(), userID, &webhook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "WEBHOOK_CREATE_FAILED",
				Message: "Failed to create webhook",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    webhook,
		Message: "Webhook created successfully",
	})
}

// GetWebhook godoc
// @Summary 获取Webhook详情
// @Description 根据ID获取Webhook的详细信息
// @Tags Webhook
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "WebhookID"
// @Success 200 {object} models.APIResponse{data=models.Webhook} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "未找到"
// @Router /webhooks/{id} [get]
func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid webhook ID",
			},
		})
		return
	}

	webhook, err := h.webhookService.GetWebhook(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "WEBHOOK_NOT_FOUND",
				Message: "Webhook not found",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    webhook,
	})
}

// ListWebhooks godoc
// @Summary 列出Webhook
// @Description 获取用户的所有Webhook
// @Tags Webhook
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param events query string false "事件类型过滤，逗号分隔"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}} "成功"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /webhooks [get]
func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	eventsStr := c.Query("events")
	var events []string
	if eventsStr != "" {
		events = strings.Split(eventsStr, ",")
	}

	webhooks, err := h.webhookService.ListWebhooks(c.Request.Context(), userID, events)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "WEBHOOK_LIST_FAILED",
				Message: "Failed to list webhooks",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"webhooks": webhooks,
			"total":    len(webhooks),
		},
	})
}

// UpdateWebhook godoc
// @Summary 更新Webhook
// @Description 更新指定的Webhook配置
// @Tags Webhook
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "WebhookID"
// @Param request body map[string]interface{} true "Webhook更新请求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /webhooks/{id} [put]
func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid webhook ID",
			},
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
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

	err = h.webhookService.UpdateWebhook(c.Request.Context(), id, userID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "WEBHOOK_UPDATE_FAILED",
				Message: "Failed to update webhook",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Webhook updated successfully",
	})
}

// DeleteWebhook godoc
// @Summary 删除Webhook
// @Description 删除指定的Webhook
// @Tags Webhook
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "WebhookID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /webhooks/{id} [delete]
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid webhook ID",
			},
		})
		return
	}

	err = h.webhookService.DeleteWebhook(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "WEBHOOK_DELETE_FAILED",
				Message: "Failed to delete webhook",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Webhook deleted successfully",
	})
}

// TestWebhook godoc
// @Summary 测试Webhook
// @Description 发送测试事件到Webhook
// @Tags Webhook
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "WebhookID"
// @Success 200 {object} models.APIResponse "测试成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /webhooks/{id}/test [post]
func (h *WebhookHandler) TestWebhook(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid webhook ID",
			},
		})
		return
	}

	err = h.webhookService.TestWebhook(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "WEBHOOK_TEST_FAILED",
				Message: "Failed to test webhook",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Test webhook sent successfully",
	})
}
