package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/permission"
	"github.com/gin-gonic/gin"
)

// PermissionHandler handles permission-related HTTP requests
type PermissionHandler struct {
	permService *permission.PermissionService
}

// NewPermissionHandler creates a new PermissionHandler
func NewPermissionHandler(permService *permission.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permService: permService,
	}
}

// CheckPermissionRequest represents the request body for permission checks
type CheckPermissionRequest struct {
	UserID       int64  `json:"user_id" example:"123" binding:"required"`
	ResourceType string `json:"resource_type" example:"workspace" binding:"required"`
	ResourceID   string `json:"resource_id" example:"ws-001"`
	Action       string `json:"action" example:"read" binding:"required"`
}

// CheckPermissionResponse represents the response for permission checks
type CheckPermissionResponse struct {
	Allowed bool `json:"allowed" example:"true"`
}

// CheckPermission godoc
// @Summary Check permission
// @Description Check if a user has permission to perform an action on a resource
// @Tags permissions
// @Accept json
// @Produce json
// @Param request body CheckPermissionRequest true "Permission check request"
// @Success 200 {object} models.APIResponse{data=CheckPermissionResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /permissions/check [post]
// @Security BearerAuth
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	var req CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), permission.CheckPermissionRequest{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Action:       req.Action,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "PERMISSION_CHECK_ERROR",
				Message: fmt.Sprintf("Failed to check permission: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    CheckPermissionResponse{Allowed: allowed},
	})
}

// AssignRoleRequest represents the request body for role assignment
type AssignRoleRequest struct {
	Role string `json:"role" example:"role:admin" binding:"required"`
}

// AssignRole godoc
// @Summary Assign role to user
// @Description Assign a role to a specific user
// @Tags permissions
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Param request body AssignRoleRequest true "Role assignment request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /users/{user_id}/roles [post]
// @Security BearerAuth
func (h *PermissionHandler) AssignRole(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_USER_ID",
				Message: "Invalid user ID",
			},
		})
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	if err := h.permService.AssignRole(c.Request.Context(), userID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ROLE_ASSIGNMENT_ERROR",
				Message: fmt.Sprintf("Failed to assign role: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Role assigned successfully",
	})
}

// RemoveRole godoc
// @Summary Remove role from user
// @Description Remove a role from a specific user
// @Tags permissions
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Param role query string true "Role to remove"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /users/{user_id}/roles [delete]
// @Security BearerAuth
func (h *PermissionHandler) RemoveRole(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_USER_ID",
				Message: "Invalid user ID",
			},
		})
		return
	}

	role := c.Query("role")
	if role == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_ROLE",
				Message: "Role parameter is required",
			},
		})
		return
	}

	if err := h.permService.RemoveRole(c.Request.Context(), userID, role); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ROLE_REMOVAL_ERROR",
				Message: fmt.Sprintf("Failed to remove role: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Role removed successfully",
	})
}

// GetUserRolesResponse represents the response for getting user roles
type GetUserRolesResponse struct {
	Roles []string `json:"roles" example:"role:admin,role:user"`
}

// GetUserRoles godoc
// @Summary Get user roles
// @Description Get all roles assigned to a specific user
// @Tags permissions
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {object} models.APIResponse{data=GetUserRolesResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /users/{user_id}/roles [get]
// @Security BearerAuth
func (h *PermissionHandler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_USER_ID",
				Message: "Invalid user ID",
			},
		})
		return
	}

	roles, err := h.permService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "GET_ROLES_ERROR",
				Message: fmt.Sprintf("Failed to get user roles: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    GetUserRolesResponse{Roles: roles},
	})
}

// AddPolicyRequest represents the request body for adding a policy
type AddPolicyRequest struct {
	Subject string `json:"subject" example:"role:admin" binding:"required"`
	Object  string `json:"object" example:"workspace" binding:"required"`
	Action  string `json:"action" example:"read" binding:"required"`
}

// AddPolicy godoc
// @Summary Add policy rule
// @Description Add a new RBAC policy rule
// @Tags permissions
// @Accept json
// @Produce json
// @Param request body AddPolicyRequest true "Policy rule"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /policies [post]
// @Security BearerAuth
func (h *PermissionHandler) AddPolicy(c *gin.Context) {
	var req AddPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	if err := h.permService.AddPolicy(c.Request.Context(), req.Subject, req.Object, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ADD_POLICY_ERROR",
				Message: fmt.Sprintf("Failed to add policy: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Policy added successfully",
	})
}

// RemovePolicy godoc
// @Summary Remove policy rule
// @Description Remove an RBAC policy rule
// @Tags permissions
// @Accept json
// @Produce json
// @Param request body AddPolicyRequest true "Policy rule to remove"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /policies [delete]
// @Security BearerAuth
func (h *PermissionHandler) RemovePolicy(c *gin.Context) {
	var req AddPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	if err := h.permService.RemovePolicy(c.Request.Context(), req.Subject, req.Object, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "REMOVE_POLICY_ERROR",
				Message: fmt.Sprintf("Failed to remove policy: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Policy removed successfully",
	})
}

// GetAllPoliciesResponse represents the response for getting all policies
type GetAllPoliciesResponse struct {
	Policies [][]string `json:"policies" example:"[[\"role:admin\",\"workspace\",\"read\"]]"`
}

// GetAllPolicies godoc
// @Summary Get all policies
// @Description Get all RBAC policy rules
// @Tags permissions
// @Produce json
// @Success 200 {object} models.APIResponse{data=GetAllPoliciesResponse}
// @Failure 500 {object} models.APIResponse
// @Router /policies [get]
// @Security BearerAuth
func (h *PermissionHandler) GetAllPolicies(c *gin.Context) {
	policies, err := h.permService.GetAllPolicies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "GET_POLICIES_ERROR",
				Message: fmt.Sprintf("Failed to get policies: %v", err),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    GetAllPoliciesResponse{Policies: policies},
	})
}
