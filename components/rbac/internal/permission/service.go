package permission

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"

	"github.com/davidliyutong/idekube-rbac/pkg/logger"
)

// PermissionService wraps Casbin enforcer operations and provides typed requests.
type PermissionService struct {
	enforcer *casbin.Enforcer
	log      *logger.Logger
}

func NewPermissionService(enforcer *casbin.Enforcer, log *logger.Logger) *PermissionService {
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
	if strings.TrimSpace(req.ResourceID) != "" {
		object = fmt.Sprintf("%s:%s", req.ResourceType, req.ResourceID)
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

	if _, err := s.enforcer.AddRoleForUser(fmt.Sprintf("user:%d", userID), role); err != nil {
		return err
	}
	return s.enforcer.SavePolicy()
}
