package permission

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/davidliyutong/idekube-controller/internal/opa"
	"github.com/davidliyutong/idekube-controller/pkg/logger"
)

// PermissionService wraps OPA enforcer operations and provides typed requests.
type PermissionService struct {
	enforcer *opa.Enforcer
	log      *logger.Logger
}

func NewPermissionService(enforcer *opa.Enforcer, log *logger.Logger) *PermissionService {
	return &PermissionService{enforcer: enforcer, log: log}
}

// CheckPermissionRequest describes a permission check input payload.
type CheckPermissionRequest struct {
	UserID       int64  `json:"user_id" example:"123" binding:"required"`
	ResourceType string `json:"resource_type" example:"workspace" binding:"required"`
	ResourceID   string `json:"resource_id" example:"ws-001"`
	Action       string `json:"action" example:"read" binding:"required"`
}

// CheckPermission returns whether the given user can perform the action on the resource.
func (s *PermissionService) CheckPermission(ctx context.Context, req CheckPermissionRequest) (bool, error) {
	if req.UserID == 0 {
		return false, errors.New("user_id is required")
	}
	if strings.TrimSpace(req.ResourceType) == "" {
		return false, errors.New("resource_type is required")
	}
	if strings.TrimSpace(req.Action) == "" {
		return false, errors.New("action is required")
	}

	subject := fmt.Sprintf("user:%d", req.UserID)
	object := strings.TrimSpace(req.ResourceType)
	resourceID := strings.TrimSpace(req.ResourceID)

	// Handle wildcard explicitly - "*" means check generic resource type permission
	if resourceID != "" && resourceID != "*" {
		object = fmt.Sprintf("%s:%s", req.ResourceType, resourceID)
	}

	allowed, err := s.enforcer.Enforce(subject, object, req.Action)
	if err != nil {
		return false, err
	}

	s.log.Debugf("permission check sub=%s obj=%s act=%s allowed=%t", subject, object, req.Action, allowed)
	return allowed, nil
}

// AssignRole grants a role to a user principal (e.g. user:123 -> role:admin).
func (s *PermissionService) AssignRole(ctx context.Context, userID int64, role string) error {
	if userID == 0 || strings.TrimSpace(role) == "" {
		return errors.New("user_id and role are required")
	}

	subject := fmt.Sprintf("user:%d", userID)
	if err := s.enforcer.AddRoleForUser(subject, role); err != nil {
		return err
	}
	s.log.Infof("Assigned role %s to user %d", role, userID)
	return nil
}

// RemoveRole removes a role from a user.
func (s *PermissionService) RemoveRole(ctx context.Context, userID int64, role string) error {
	if userID == 0 || strings.TrimSpace(role) == "" {
		return errors.New("user_id and role are required")
	}

	subject := fmt.Sprintf("user:%d", userID)
	if err := s.enforcer.RemoveRoleForUser(subject, role); err != nil {
		return err
	}
	s.log.Infof("Removed role %s from user %d", role, userID)
	return nil
}

// GetUserRoles returns all roles assigned to a user.
func (s *PermissionService) GetUserRoles(ctx context.Context, userID int64) ([]string, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}

	subject := fmt.Sprintf("user:%d", userID)
	roles, err := s.enforcer.GetRolesForUser(subject)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// AddPolicy adds a new policy rule.
func (s *PermissionService) AddPolicy(ctx context.Context, subject, object, action string) error {
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(object) == "" || strings.TrimSpace(action) == "" {
		return errors.New("subject, object, and action are required")
	}

	if err := s.enforcer.AddPolicy(subject, object, action); err != nil {
		return err
	}
	s.log.Infof("Added policy: %s -> %s [%s]", subject, object, action)
	return nil
}

// RemovePolicy removes a policy rule.
func (s *PermissionService) RemovePolicy(ctx context.Context, subject, object, action string) error {
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(object) == "" || strings.TrimSpace(action) == "" {
		return errors.New("subject, object, and action are required")
	}

	if err := s.enforcer.RemovePolicy(subject, object, action); err != nil {
		return err
	}
	s.log.Infof("Removed policy: %s -> %s [%s]", subject, object, action)
	return nil
}

// GetAllPolicies returns all policy rules.
func (s *PermissionService) GetAllPolicies(ctx context.Context) ([][]string, error) {
	policies, err := s.enforcer.GetAllPolicies()
	if err != nil {
		return nil, err
	}
	return policies, nil
}
