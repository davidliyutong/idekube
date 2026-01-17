package permission

import (
	"context"
	"fmt"
)

// ResourcePermissionService handles resource-level permissions
type ResourcePermissionService struct {
	permService *PermissionService
}

// NewResourcePermissionService creates a new resource permission service
func NewResourcePermissionService(permService *PermissionService) *ResourcePermissionService {
	return &ResourcePermissionService{
		permService: permService,
	}
}

// GrantResourceOwnership grants full ownership permissions for a resource to a user
// This is used when a user creates a resource (workspace, organization, etc.)
func (s *ResourcePermissionService) GrantResourceOwnership(ctx context.Context, userID int64, resourceType string, resourceID int64) error {
	// Create a special role for this resource instance
	// Format: "role:owner:{resourceType}:{resourceID}"
	ownerRole := fmt.Sprintf("role:owner:%s:%d", resourceType, resourceID)

	// Assign the owner role to the user
	if err := s.permService.AssignRole(ctx, userID, ownerRole); err != nil {
		return fmt.Errorf("failed to assign owner role: %w", err)
	}

	// Add policies for this owner role to have full access
	resourceObject := fmt.Sprintf("%s:%d", resourceType, resourceID)

	actions := []string{"read", "update", "delete", "manage"}
	for _, action := range actions {
		if err := s.permService.AddPolicy(ctx, ownerRole, resourceObject, action); err != nil {
			s.permService.log.Warnf("Failed to add policy for %s: %v", action, err)
		}
	}

	s.permService.log.Infof("Granted ownership of %s:%d to user %d", resourceType, resourceID, userID)
	return nil
}

// GrantOrganizationMembership grants appropriate permissions based on organization role
func (s *ResourcePermissionService) GrantOrganizationMembership(ctx context.Context, userID int64, orgID int64, role string) error {
	// Map organization roles to permission roles
	var permRole string
	switch role {
	case "owner":
		return s.GrantResourceOwnership(ctx, userID, "organization", orgID)
	case "admin":
		permRole = fmt.Sprintf("role:admin:organization:%d", orgID)
	case "member":
		permRole = fmt.Sprintf("role:member:organization:%d", orgID)
	default:
		return fmt.Errorf("unknown organization role: %s", role)
	}

	// Assign the role
	if err := s.permService.AssignRole(ctx, userID, permRole); err != nil {
		return fmt.Errorf("failed to assign org role: %w", err)
	}

	// Add appropriate policies
	orgObject := fmt.Sprintf("organization:%d", orgID)

	var actions []string
	if role == "admin" {
		actions = []string{"read", "update", "manage_members"}
	} else {
		actions = []string{"read"}
	}

	for _, action := range actions {
		if err := s.permService.AddPolicy(ctx, permRole, orgObject, action); err != nil {
			s.permService.log.Warnf("Failed to add org policy for %s: %v", action, err)
		}
	}

	s.permService.log.Infof("Granted %s role on organization %d to user %d", role, orgID, userID)
	return nil
}

// RevokeResourcePermissions removes all permissions for a user on a specific resource
func (s *ResourcePermissionService) RevokeResourcePermissions(ctx context.Context, userID int64, resourceType string, resourceID int64) error {
	// Remove owner role
	ownerRole := fmt.Sprintf("role:owner:%s:%d", resourceType, resourceID)
	if err := s.permService.RemoveRole(ctx, userID, ownerRole); err != nil {
		s.permService.log.Warnf("Failed to remove owner role: %v", err)
	}

	// Remove other resource-specific roles
	roles := []string{
		fmt.Sprintf("role:admin:%s:%d", resourceType, resourceID),
		fmt.Sprintf("role:member:%s:%d", resourceType, resourceID),
	}

	for _, role := range roles {
		if err := s.permService.RemoveRole(ctx, userID, role); err != nil {
			s.permService.log.Warnf("Failed to remove role %s: %v", role, err)
		}
	}

	s.permService.log.Infof("Revoked permissions on %s:%d from user %d", resourceType, resourceID, userID)
	return nil
}

// GrantOrganizationResourceAccess grants access to all workspaces/volumes in an organization
func (s *ResourcePermissionService) GrantOrganizationResourceAccess(ctx context.Context, userID int64, orgID int64, resourceType string) error {
	// Create a role that grants access to all resources of a type within an org
	// Format: "role:org_member:{orgID}:{resourceType}"
	role := fmt.Sprintf("role:org_member:%d:%s", orgID, resourceType)

	// Assign the role
	if err := s.permService.AssignRole(ctx, userID, role); err != nil {
		return fmt.Errorf("failed to assign org resource role: %w", err)
	}

	// Add policy to read resources in this organization
	// Pattern: workspace:org:{orgID}:* means all workspaces in this org
	orgResourcePattern := fmt.Sprintf("%s:org:%d:*", resourceType, orgID)

	if err := s.permService.AddPolicy(ctx, role, orgResourcePattern, "read"); err != nil {
		return fmt.Errorf("failed to add org resource policy: %w", err)
	}

	s.permService.log.Infof("Granted %s access in organization %d to user %d", resourceType, orgID, userID)
	return nil
}
