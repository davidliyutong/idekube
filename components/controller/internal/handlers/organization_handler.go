package handlers

import (
	"net/http"
	"strconv"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// OrganizationHandler handles organization-related HTTP requests
type OrganizationHandler struct {
	orgService *services.OrganizationService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(orgService *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
	}
}

// CreateOrganization creates a new organization
// POST /api/v1/organizations
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	var req models.CreateOrganizationRequest
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

	org, err := h.orgService.CreateOrganization(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATE_FAILED",
				Message: "Failed to create organization",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    org,
	})
}

// GetOrganization retrieves an organization by ID
// GET /api/v1/organizations/:id
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	// Check if user wants detailed view with members
	withMembers := c.Query("with_members") == "true"

	if withMembers {
		orgWithMembers, err := h.orgService.GetOrganizationWithMembers(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "NOT_FOUND",
					Message: "Organization not found",
				},
			})
			return
		}
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Data:    orgWithMembers,
		})
	} else {
		org, err := h.orgService.GetOrganization(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "NOT_FOUND",
					Message: "Organization not found",
				},
			})
			return
		}
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Data:    org,
		})
	}
}

// ListUserOrganizations lists all organizations the current user belongs to
// GET /api/v1/organizations
func (h *OrganizationHandler) ListUserOrganizations(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	orgs, err := h.orgService.ListUserOrganizations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list organizations",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    orgs,
	})
}

// UpdateOrganization updates an organization
// PUT /api/v1/organizations/:id
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	// Check if user has admin permission
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, userID, models.OrgRoleAdmin)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Insufficient permissions",
			},
		})
		return
	}

	var req models.UpdateOrganizationRequest
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

	org, err := h.orgService.UpdateOrganization(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update organization",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    org,
	})
}

// DeleteOrganization deletes an organization
// DELETE /api/v1/organizations/:id
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	// Check if user is owner
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, userID, models.OrgRoleOwner)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Only organization owner can delete the organization",
			},
		})
		return
	}

	err = h.orgService.DeleteOrganization(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete organization",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Organization deleted successfully",
	})
}

// AddMember adds a member to an organization
// POST /api/v1/organizations/:id/members
func (h *OrganizationHandler) AddMember(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	// Check if user has admin permission
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, userID, models.OrgRoleAdmin)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Insufficient permissions",
			},
		})
		return
	}

	var req models.AddOrganizationMemberRequest
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

	member, err := h.orgService.AddMember(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ADD_MEMBER_FAILED",
				Message: "Failed to add member",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    member,
	})
}

// RemoveMember removes a member from an organization
// DELETE /api/v1/organizations/:id/members/:user_id
func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	memberIDParam := c.Param("user_id")
	memberID, err := strconv.ParseInt(memberIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	// Check if user has admin permission
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, userID, models.OrgRoleAdmin)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Insufficient permissions",
			},
		})
		return
	}

	err = h.orgService.RemoveMember(c.Request.Context(), id, memberID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "REMOVE_MEMBER_FAILED",
				Message: "Failed to remove member",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Member removed successfully",
	})
}

// UpdateMemberRole updates a member's role
// PUT /api/v1/organizations/:id/members/:user_id
func (h *OrganizationHandler) UpdateMemberRole(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid organization ID",
			},
		})
		return
	}

	memberIDParam := c.Param("user_id")
	memberID, err := strconv.ParseInt(memberIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	// Check if user has owner permission (only owner can change roles)
	hasPermission, err := h.orgService.CheckMemberPermission(c.Request.Context(), id, userID, models.OrgRoleOwner)
	if err != nil || !hasPermission {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Only organization owner can change member roles",
			},
		})
		return
	}

	var req models.UpdateOrganizationMemberRequest
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

	member, err := h.orgService.UpdateMemberRole(c.Request.Context(), id, memberID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_ROLE_FAILED",
				Message: "Failed to update member role",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    member,
	})
}
