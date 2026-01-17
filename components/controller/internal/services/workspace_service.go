package services

import (
	"context"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/permission"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/pkg/queue"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WorkspaceService handles workspace business logic
type WorkspaceService struct {
	workspaceRepo     *repository.WorkspaceRepository
	templateRepo      *repository.TemplateRepository
	volumeRepo        *repository.VolumeRepository
	eventPublisher    *queue.EventPublisher
	logger            *zap.Logger
	permissionService *permission.ResourcePermissionService

	// Flag to enable direct K8S operations (for backward compatibility during migration)
	enableDirectK8S bool
}

// NewWorkspaceService creates a new workspace service
func NewWorkspaceService(
	workspaceRepo *repository.WorkspaceRepository,
	templateRepo *repository.TemplateRepository,
	volumeRepo *repository.VolumeRepository,
	eventPublisher *queue.EventPublisher,
	logger *zap.Logger,
	permissionService *permission.ResourcePermissionService,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo:     workspaceRepo,
		templateRepo:      templateRepo,
		volumeRepo:        volumeRepo,
		eventPublisher:    eventPublisher,
		logger:            logger,
		permissionService: permissionService,
		enableDirectK8S:   false, // Disable direct K8S operations by default
	}
}

// CreateWorkspace creates a new workspace
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req *models.CreateWorkspaceRequest) (*models.Workspace, error) {
	// Get template
	template, err := s.templateRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Use template defaults if not provided
	cpuMillicores := req.CPUMillicores
	if cpuMillicores == 0 {
		cpuMillicores = template.DefaultCPUMillicores
	}

	memoryMB := req.MemoryMB
	if memoryMB == 0 {
		memoryMB = template.DefaultMemoryMB
	}

	storageMB := req.StorageMB
	if storageMB == 0 {
		storageMB = template.DefaultStorageMB
	}

	// Create labels for RBAC
	labels := models.ResourceLabels{
		"owner_type": string(req.OwnerType),
		"owner_id":   fmt.Sprintf("%d", req.OwnerID),
	}

	// Add organization_id label if it's an org workspace
	var orgID *int64
	if req.OwnerType == models.OwnerTypeOrganization {
		orgID = &req.OwnerID
		labels["organization_id"] = fmt.Sprintf("%d", req.OwnerID)
	}

	workspace := &models.Workspace{
		UUID:           uuid.New(),
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		OwnerType:      req.OwnerType,
		OwnerID:        req.OwnerID,
		TemplateID:     req.TemplateID,
		CPUMillicores:  cpuMillicores,
		MemoryMB:       memoryMB,
		StorageMB:      storageMB,
		CurrentStatus:  models.WorkspaceStatusPending,
		TargetStatus:   models.WorkspaceStatusRunning,
		Labels:         labels,
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Create in database first
	err = s.workspaceRepo.Create(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Grant ownership permissions automatically (if permission service is available)
	if s.permissionService != nil {
		// Determine the actual user who created this workspace
		var creatorUserID int64
		if req.OwnerType == models.OwnerTypeUser {
			creatorUserID = req.OwnerID
		} else {
			// For organization workspaces, the CreatedBy field should be set by the handler
			// For now, we'll grant ownership to the OwnerID (which is the org)
			// TODO: Track actual creator separately and grant them permissions
			creatorUserID = req.OwnerID
		}

		if err := s.permissionService.GrantResourceOwnership(ctx, creatorUserID, "workspace", workspace.ID); err != nil {
			s.logger.Warn("Failed to grant workspace ownership permissions",
				zap.Int64("workspace_id", workspace.ID),
				zap.Int64("creator_id", creatorUserID),
				zap.Error(err))
		}
	}

	// Get attached volumes
	_, err = s.workspaceRepo.ListVolumes(ctx, workspace.ID)
	if err != nil {
		s.logger.Warn("Failed to list volumes for workspace",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
	}

	// Publish workspace creation event to HouseKeeper
	// NOTE: EventPublisher expects Volume, not WorkspaceVolume.
	// In a production scenario, you'd resolve full volumes from workspaceVolumes.
	volumes := []*models.Volume{}
	if err := s.eventPublisher.PublishWorkspaceCreate(ctx, workspace, template, volumes); err != nil {
		s.logger.Error("Failed to publish workspace create event",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
		// Don't rollback - HouseKeeper reconciler will handle it
	}

	s.logger.Info("Workspace created, event published",
		zap.Int64("workspace_id", workspace.ID),
		zap.String("name", workspace.Name),
		zap.String("status", string(workspace.CurrentStatus)))

	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (s *WorkspaceService) GetWorkspace(ctx context.Context, id int64) (*models.Workspace, error) {
	return s.workspaceRepo.GetByID(ctx, id)
}

// ListWorkspacesByOwner lists workspaces owned by a specific owner
func (s *WorkspaceService) ListWorkspacesByOwner(ctx context.Context, ownerType models.OwnerType, ownerID int64) ([]*models.Workspace, error) {
	return s.workspaceRepo.ListByOwner(ctx, ownerType, ownerID)
}

// UpdateWorkspace updates a workspace
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, id int64, req *models.UpdateWorkspaceRequest) (*models.Workspace, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.DisplayName != nil {
		workspace.DisplayName = req.DisplayName
	}
	if req.Description != nil {
		workspace.Description = req.Description
	}
	if req.IsShared != nil {
		workspace.IsShared = *req.IsShared
	}

	// Handle resource changes
	needsUpdate := false
	reason := ""
	if req.CPUMillicores != nil && *req.CPUMillicores != workspace.CPUMillicores {
		workspace.CPUMillicores = *req.CPUMillicores
		needsUpdate = true
		reason += fmt.Sprintf("CPU: %d millicores; ", *req.CPUMillicores)
	}
	if req.MemoryMB != nil && *req.MemoryMB != workspace.MemoryMB {
		workspace.MemoryMB = *req.MemoryMB
		needsUpdate = true
		reason += fmt.Sprintf("Memory: %d MB; ", *req.MemoryMB)
	}

	// Update workspace in database
	err = s.workspaceRepo.Update(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	// Publish update event if resources changed
	if needsUpdate {
		template, _ := s.templateRepo.GetByID(ctx, workspace.TemplateID)

		// NOTE: EventPublisher expects Volume, not WorkspaceVolume.
		// In production you'd resolve full volumes from workspaceVolumes.
		volumes := []*models.Volume{}
		if err := s.eventPublisher.PublishWorkspaceUpdate(ctx, workspace, template, volumes, reason); err != nil {
			s.logger.Error("Failed to publish workspace update event",
				zap.Int64("workspace_id", workspace.ID),
				zap.Error(err))
		}
	}

	return workspace, nil
}

// DeleteWorkspace deletes a workspace
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id int64) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Publish delete event to HouseKeeper
	if err := s.eventPublisher.PublishWorkspaceDelete(ctx, id, workspace); err != nil {
		s.logger.Error("Failed to publish workspace delete event",
			zap.Int64("workspace_id", id),
			zap.Error(err))
		// Continue with deletion anyway
	}

	// Delete from database
	return s.workspaceRepo.Delete(ctx, id)
}

// StartWorkspace starts a stopped workspace
func (s *WorkspaceService) StartWorkspace(ctx context.Context, id int64) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update target status
	workspace.TargetStatus = models.WorkspaceStatusRunning
	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	// Publish start event
	// TODO: Fetch template and volumes for event publishing
	if err := s.eventPublisher.PublishWorkspaceStart(ctx, workspace, nil, nil); err != nil {
		s.logger.Error("Failed to publish workspace start event",
			zap.Int64("workspace_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to publish start event: %w", err)
	}

	return nil
}

// StopWorkspace stops a running workspace
func (s *WorkspaceService) StopWorkspace(ctx context.Context, id int64) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update target status
	workspace.TargetStatus = models.WorkspaceStatusStopped
	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	// Publish stop event
	if err := s.eventPublisher.PublishWorkspaceStop(ctx, workspace); err != nil {
		s.logger.Error("Failed to publish workspace stop event",
			zap.Int64("workspace_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to publish stop event: %w", err)
	}

	return nil
}

// AttachVolume attaches a volume to a workspace
func (s *WorkspaceService) AttachVolume(ctx context.Context, workspaceID, volumeID int64, mountPath string) error {
	// Verify workspace exists
	_, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Verify volume exists
	_, err = s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return err
	}

	wv := &models.WorkspaceVolume{
		WorkspaceID: workspaceID,
		VolumeID:    volumeID,
		MountPath:   mountPath,
	}

	return s.workspaceRepo.AttachVolume(ctx, wv)
}

// DetachVolume detaches a volume from a workspace
func (s *WorkspaceService) DetachVolume(ctx context.Context, workspaceID, volumeID int64) error {
	return s.workspaceRepo.DetachVolume(ctx, workspaceID, volumeID)
}

// ListWorkspaces lists workspaces based on user role and permissions
func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID int64, userRole models.UserRole, orgRole *models.OrganizationMemberRole, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	switch userRole {
	case models.UserRoleSuperAdmin, models.UserRoleAdmin:
		// Admins can see all workspaces
		return s.workspaceRepo.ListAll(ctx, opts)

	case models.UserRolePowerUser, models.UserRoleUser:
		// Regular users and power users see their own workspaces and org workspaces
		return s.workspaceRepo.ListAccessibleByUser(ctx, userID, opts)

	default:
		return nil, 0, fmt.Errorf("unknown user role: %s", userRole)
	}
}

// ListWorkspacesByOrganization lists workspaces in a specific organization
func (s *WorkspaceService) ListWorkspacesByOrganization(ctx context.Context, orgID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	labels := map[string]string{
		"organization_id": fmt.Sprintf("%d", orgID),
	}
	return s.workspaceRepo.ListByLabel(ctx, labels, opts)
}

// ListOrgWorkspacesForAdmin lists workspaces in organizations where user is owner/admin
func (s *WorkspaceService) ListOrgWorkspacesForAdmin(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	return s.workspaceRepo.ListByOrganizationAll(ctx, userID, opts)
}
