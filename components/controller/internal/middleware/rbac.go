package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/permission"
	"github.com/gin-gonic/gin"
)

// RBACMiddleware creates middleware for RBAC permission checking
func RBACMiddleware(permService *permission.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is a base middleware that just ensures permission service is available
		// Actual permission checks are done in route-specific middleware
		c.Set("permission_service", permService)
		c.Next()
	}
}

// RBACCheck creates a middleware that checks specific permissions
func RBACCheck(permService *permission.PermissionService, resourceType, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		// Get resource ID from URL parameter or use wildcard
		resourceID := "*"
		if idParam := c.Param("id"); idParam != "" {
			resourceID = idParam
		}

		// Check permission with permission service
		allowed, err := permService.CheckPermission(c.Request.Context(), permission.CheckPermissionRequest{
			UserID:       userID.(int64),
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Action:       action,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "RBAC_ERROR",
					Message: fmt.Sprintf("Failed to check permission: %v", err),
				},
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "FORBIDDEN",
					Message: fmt.Sprintf("You don't have permission to %s %s", action, resourceType),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RBACCheckEndpoint creates middleware that checks API endpoint-level permissions
// This is the new unified approach that replaces role-based middleware
func RBACCheckEndpoint(permService *permission.PermissionService, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		// Derive action from HTTP method
		action := httpMethodToAction(c.Request.Method)

		// Get resource ID from URL parameter or use wildcard for list operations
		resourceID := "*"
		if idParam := c.Param("id"); idParam != "" {
			resourceID = idParam
		} else if userIDParam := c.Param("user_id"); userIDParam != "" {
			// For nested resources like /organizations/:id/members/:user_id
			resourceID = userIDParam
		}

		// Check permission with permission service
		allowed, err := permService.CheckPermission(c.Request.Context(), permission.CheckPermissionRequest{
			UserID:       userID.(int64),
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Action:       action,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "RBAC_ERROR",
					Message: fmt.Sprintf("Failed to check permission: %v", err),
				},
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "FORBIDDEN",
					Message: fmt.Sprintf("You don't have permission to %s %s", action, resourceType),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// httpMethodToAction converts HTTP method to RBAC action
func httpMethodToAction(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "read"
	}
}

// RequireSuperAdmin is a middleware that requires super admin role
// DEPRECATED: Use RBACCheckEndpoint instead
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		if userRole != string(models.UserRoleSuperAdmin) {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "FORBIDDEN",
					Message: "This action requires super admin privileges",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdminOrAbove is a middleware that requires admin or super admin role
// DEPRECATED: Use RBACCheckEndpoint instead
func RequireAdminOrAbove() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		role := models.UserRole(userRole.(string))
		if role != models.UserRoleSuperAdmin && role != models.UserRoleAdmin {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "FORBIDDEN",
					Message: "This action requires admin privileges",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePowerUserOrAbove is a middleware that requires power user or higher role
// DEPRECATED: Use RBACCheckEndpoint instead
func RequirePowerUserOrAbove() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		role := models.UserRole(userRole.(string))
		if role != models.UserRoleSuperAdmin &&
			role != models.UserRoleAdmin &&
			role != models.UserRolePowerUser {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "FORBIDDEN",
					Message: "This action requires power user privileges",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckResourceOwnership checks if user owns the resource or has org-level access
func CheckResourceOwnership(resourceIDParam string, ownerTypeGetter, ownerIDGetter func(*gin.Context) (models.OwnerType, int64, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		userRole, _ := c.Get("user_role")

		// Super admin and admin can access all resources
		role := models.UserRole(userRole.(string))
		if role == models.UserRoleSuperAdmin || role == models.UserRoleAdmin {
			c.Next()
			return
		}

		// Get resource ownership info
		ownerType, ownerID, err := ownerTypeGetter(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to check ownership",
				},
			})
			c.Abort()
			return
		}

		// Check if user owns the resource
		if ownerType == models.OwnerTypeUser && ownerID == userID.(int64) {
			c.Next()
			return
		}

		// Check if user has org-level access (implemented in service layer)
		// For now, deny access
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "You don't have permission to access this resource",
			},
		})
		c.Abort()
	}
}

// ExtractResourceID extracts resource ID from URL parameter
func ExtractResourceID(c *gin.Context, paramName string) (int64, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, fmt.Errorf("resource ID not found")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid resource ID")
	}

	return id, nil
}
