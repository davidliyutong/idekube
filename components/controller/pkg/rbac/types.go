package rbac

// CheckPermissionRequest represents a permission check request
type CheckPermissionRequest struct {
	UserID       int64  `json:"user_id"`
	ResourceType string `json:"resource_type"` // user, workspace, volume, template, organization
	ResourceID   string `json:"resource_id"`   // resource identifier
	Action       string `json:"action"`        // create, read, update, delete, list
}

// CheckPermissionResponse represents a permission check response
type CheckPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// AssignRoleRequest represents a role assignment request
type AssignRoleRequest struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"` // system role like admin, power_user, user
}

// AssignResourceRoleRequest represents a resource-level role assignment
type AssignResourceRoleRequest struct {
	UserID       int64  `json:"user_id"`
	ResourceType string `json:"resource_type"` // workspace, volume, template, organization
	ResourceID   string `json:"resource_id"`
	Role         string `json:"role"` // resource-specific role like owner, admin, member
}

// RevokeResourceRoleRequest represents a resource-level role revocation
type RevokeResourceRoleRequest struct {
	UserID       int64  `json:"user_id"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Role         string `json:"role"`
}
