package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/pkg/queue"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// WorkspaceService handles workspace business logic
type WorkspaceService struct {
	workspaceRepo  *repository.WorkspaceRepository
	templateRepo   *repository.TemplateRepository
	volumeRepo     *repository.VolumeRepository
	eventPublisher *queue.EventPublisher
	logger         *zap.Logger

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
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo:   workspaceRepo,
		templateRepo:    templateRepo,
		volumeRepo:      volumeRepo,
		eventPublisher:  eventPublisher,
		logger:          logger,
		enableDirectK8S: false, // Disable direct K8S operations by default
	}
}

// CreateWorkspace creates a new workspace
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req *models.CreateWorkspaceRequest, createdBy int64) (*models.Workspace, error) {
	// Get template
	template, err := s.templateRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Create template snapshot (immutable)
	templateSnapshot, _ := json.Marshal(template)

	// Use template defaults if not provided
	cpuMillicores := req.CPUMillicores
	// FIXME: When nil pointer values are dereferenced in template.DefaultQuota (line 61), this will cause a panic if DefaultQuota fields are nil. Add nil checks before dereferencing: if template.DefaultQuota.CPUMillicores != nil { cpuMillicores = template.DefaultQuota.CPUMillicores }
	if cpuMillicores == nil {
		cpuMillicores = template.DefaultQuota.CPUMillicores
	}

	memoryMB := req.MemoryMB
	if memoryMB == nil {
		memoryMB = template.DefaultQuota.MemoryMB
	}

	storageMB := req.StorageMB
	if storageMB == nil {
		storageMB = template.DefaultQuota.StorageMB
	}

	gpu := req.GPU
	if gpu == nil {
		gpu = template.DefaultQuota.GPU
	}

	timeoutSeconds := req.TimeoutSeconds
	if timeoutSeconds == nil {
		timeoutSeconds = template.DefaultQuota.TimeoutSeconds
	}

	now := time.Now()
	workspace := &models.Workspace{
		Base: models.Base{
			UUID:      uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Status:    models.WorkspaceStatusPending,
			ExtraInfo: nil,
			Labels:    req.Labels,
		},
		Profile: models.Profile{
			Identifier:  req.Name,
			DisplayName: req.DisplayName,
			Description: req.Description,
			IconURL:     req.IconURL,
		},
		TemplateID:       req.TemplateID,
		TemplateSnapshot: datatypes.JSONMap(map[string]interface{}{"data": string(templateSnapshot)}),
		Parameters:       req.Parameters,
		Quota: models.QuotaLimits{
			CPUMillicores:  cpuMillicores,
			MemoryMB:       memoryMB,
			StorageMB:      storageMB,
			GPU:            gpu,
			TimeoutSeconds: timeoutSeconds,
		},
		IsPublic:     false,
		OwnerID:      req.OwnerID, // Organization ID
		CreatedBy:    createdBy,
		TargetStatus: models.WorkspaceStatusRunning,
	}

	// Create in database first
	err = s.workspaceRepo.Create(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Get attached volumes
	_, err = s.workspaceRepo.ListVolumes(ctx, workspace.ID)
	if err != nil {
		s.logger.Warn("Failed to list volumes for workspace",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
	}

	// Publish workspace creation event to HouseKeeper
	volumes := []*models.Volume{}
	if err := s.eventPublisher.PublishWorkspaceCreate(ctx, workspace, template, volumes); err != nil {
		s.logger.Error("Failed to publish workspace create event",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
		// Don't rollback - HouseKeeper reconciler will handle it
	}

	s.logger.Info("Workspace created, event published",
		zap.Int64("workspace_id", workspace.ID),
		zap.String("name", workspace.Identifier),
		zap.String("status", workspace.Status))

	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (s *WorkspaceService) GetWorkspace(ctx context.Context, id int64) (*models.Workspace, error) {
	return s.workspaceRepo.GetByID(ctx, id)
}

// ListWorkspacesByOrganization lists workspaces by organization
func (s *WorkspaceService) ListWorkspacesByOrganization(ctx context.Context, orgID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	return s.workspaceRepo.ListByOrganization(ctx, orgID, opts)
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
	if req.IconURL != nil {
		workspace.IconURL = req.IconURL
	}
	if req.IsPublic != nil {
		workspace.IsPublic = *req.IsPublic
	}
	if req.Labels != nil {
		workspace.Labels = req.Labels
	}

	// Handle resource changes
	needsUpdate := false
	reason := ""
	// FIXME: Dereferencing workspace.Quota.CPUMillicores without checking if it's nil first will cause a panic. The condition should be restructured: if req.CPUMillicores != nil && (workspace.Quota.CPUMillicores == nil || *req.CPUMillicores != *workspace.Quota.CPUMillicores)
	if req.CPUMillicores != nil && (workspace.Quota.CPUMillicores == nil || *req.CPUMillicores != *workspace.Quota.CPUMillicores) {
		workspace.Quota.CPUMillicores = req.CPUMillicores
		needsUpdate = true
		reason += fmt.Sprintf("CPU: %d millicores; ", *req.CPUMillicores)
	}
	if req.MemoryMB != nil && (workspace.Quota.MemoryMB == nil || *req.MemoryMB != *workspace.Quota.MemoryMB) {
		workspace.Quota.MemoryMB = req.MemoryMB
		needsUpdate = true
		reason += fmt.Sprintf("Memory: %d MB; ", *req.MemoryMB)
	}
	if req.StorageMB != nil && (workspace.Quota.StorageMB == nil || *req.StorageMB != *workspace.Quota.StorageMB) {
		workspace.Quota.StorageMB = req.StorageMB
		needsUpdate = true
		reason += fmt.Sprintf("Storage: %d MB; ", *req.StorageMB)
	}
	if req.GPU != nil && (workspace.Quota.GPU == nil || *req.GPU != *workspace.Quota.GPU) {
		workspace.Quota.GPU = req.GPU
		needsUpdate = true
		reason += fmt.Sprintf("GPU: %d; ", *req.GPU)
	}
	if req.TimeoutSeconds != nil && (workspace.Quota.TimeoutSeconds == nil || *req.TimeoutSeconds != *workspace.Quota.TimeoutSeconds) {
		workspace.Quota.TimeoutSeconds = req.TimeoutSeconds
		needsUpdate = true
		reason += fmt.Sprintf("Timeout: %d seconds; ", *req.TimeoutSeconds)
	}

	// Update workspace in database
	err = s.workspaceRepo.Update(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	// Publish update event if resources changed
	if needsUpdate {
		template, _ := s.templateRepo.GetByID(ctx, workspace.TemplateID)
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

// ListOrgWorkspacesForAdmin lists workspaces in organizations where user is owner/admin
func (s *WorkspaceService) ListOrgWorkspacesForAdmin(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	return s.workspaceRepo.ListByOrganizationAll(ctx, userID, opts)
}

// ============================================================================
// Sub-resource APIs
// ============================================================================

// GetWorkspaceProfile returns the workspace's profile sub-resource
func (s *WorkspaceService) GetWorkspaceProfile(ctx context.Context, workspaceID int64) (*models.WorkspaceProfileResponse, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return &models.WorkspaceProfileResponse{
		Identifier:  workspace.Identifier,
		DisplayName: workspace.DisplayName,
		IconURL:     workspace.IconURL,
		Description: workspace.Description,
	}, nil
}

// UpdateWorkspaceProfile updates the workspace's profile sub-resource
func (s *WorkspaceService) UpdateWorkspaceProfile(ctx context.Context, workspaceID int64, req *models.UpdateWorkspaceProfileRequest) (*models.WorkspaceProfileResponse, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Update profile fields
	if req.DisplayName != nil {
		workspace.DisplayName = req.DisplayName
	}
	if req.IconURL != nil {
		workspace.IconURL = req.IconURL
	}
	if req.Description != nil {
		workspace.Description = req.Description
	}

	err = s.workspaceRepo.Update(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &models.WorkspaceProfileResponse{
		Identifier:  workspace.Identifier,
		DisplayName: workspace.DisplayName,
		IconURL:     workspace.IconURL,
		Description: workspace.Description,
	}, nil
}

// GetWorkspaceQuota returns the workspace's quota sub-resource
func (s *WorkspaceService) GetWorkspaceQuota(ctx context.Context, workspaceID int64) (*models.WorkspaceQuotaResponse, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return &models.WorkspaceQuotaResponse{
		CPUMillicores:  workspace.Quota.CPUMillicores,
		MemoryMB:       workspace.Quota.MemoryMB,
		StorageMB:      workspace.Quota.StorageMB,
		GPU:            workspace.Quota.GPU,
		TimeoutSeconds: workspace.Quota.TimeoutSeconds,
	}, nil
}

// UpdateWorkspaceQuota updates the workspace's quota sub-resource
func (s *WorkspaceService) UpdateWorkspaceQuota(ctx context.Context, workspaceID int64, req *models.UpdateWorkspaceQuotaRequest) (*models.WorkspaceQuotaResponse, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Update quota fields
	needsUpdate := false
	reason := ""

	if req.CPUMillicores != nil {
		workspace.Quota.CPUMillicores = req.CPUMillicores
		needsUpdate = true
		reason += fmt.Sprintf("CPU: %d millicores; ", *req.CPUMillicores)
	}
	if req.MemoryMB != nil {
		workspace.Quota.MemoryMB = req.MemoryMB
		needsUpdate = true
		reason += fmt.Sprintf("Memory: %d MB; ", *req.MemoryMB)
	}
	if req.StorageMB != nil {
		workspace.Quota.StorageMB = req.StorageMB
		needsUpdate = true
		reason += fmt.Sprintf("Storage: %d MB; ", *req.StorageMB)
	}
	if req.GPU != nil {
		workspace.Quota.GPU = req.GPU
		needsUpdate = true
		reason += fmt.Sprintf("GPU: %d; ", *req.GPU)
	}
	if req.TimeoutSeconds != nil {
		workspace.Quota.TimeoutSeconds = req.TimeoutSeconds
		needsUpdate = true
		reason += fmt.Sprintf("Timeout: %d seconds; ", *req.TimeoutSeconds)
	}

	err = s.workspaceRepo.Update(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to update quota: %w", err)
	}

	// Publish update event if resources changed
	if needsUpdate {
		template, _ := s.templateRepo.GetByID(ctx, workspace.TemplateID)
		volumes := []*models.Volume{}
		if err := s.eventPublisher.PublishWorkspaceUpdate(ctx, workspace, template, volumes, reason); err != nil {
			s.logger.Error("Failed to publish workspace update event",
				zap.Int64("workspace_id", workspace.ID),
				zap.Error(err))
		}
	}

	return &models.WorkspaceQuotaResponse{
		CPUMillicores:  workspace.Quota.CPUMillicores,
		MemoryMB:       workspace.Quota.MemoryMB,
		StorageMB:      workspace.Quota.StorageMB,
		GPU:            workspace.Quota.GPU,
		TimeoutSeconds: workspace.Quota.TimeoutSeconds,
	}, nil
}

// ListWorkspaceVolumes lists all volumes attached to a workspace
func (s *WorkspaceService) ListWorkspaceVolumes(ctx context.Context, workspaceID int64) ([]models.WorkspaceVolumeResponse, error) {
	workspaceVolumes, err := s.workspaceRepo.ListVolumes(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	result := make([]models.WorkspaceVolumeResponse, 0, len(workspaceVolumes))
	for _, wv := range workspaceVolumes {
		volume, err := s.volumeRepo.GetByID(ctx, wv.VolumeID)
		if err != nil {
			continue
		}
		result = append(result, models.WorkspaceVolumeResponse{
			VolumeID:  wv.VolumeID,
			MountPath: wv.MountPath,
			Volume:    volume,
		})
	}

	return result, nil
}

// AddWorkspaceVolume attaches a volume to a workspace
func (s *WorkspaceService) AddWorkspaceVolume(ctx context.Context, workspaceID int64, req *models.AttachVolumeRequest) error {
	return s.AttachVolume(ctx, workspaceID, req.VolumeID, req.MountPath)
}

// RemoveWorkspaceVolume detaches a volume from a workspace
func (s *WorkspaceService) RemoveWorkspaceVolume(ctx context.Context, workspaceID, volumeID int64) error {
	return s.DetachVolume(ctx, workspaceID, volumeID)
}

// UpdateWorkspaceIsPublic updates the workspace's is_public status
func (s *WorkspaceService) UpdateWorkspaceIsPublic(ctx context.Context, workspaceID int64, req *models.UpdateWorkspaceIsPublicRequest) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	workspace.IsPublic = req.IsPublic

	return s.workspaceRepo.Update(ctx, workspace)
}

// ============================================================================
// Additional Sub-resource APIs
// ============================================================================

// GetWorkspaceTemplate returns the template used by this workspace
func (s *WorkspaceService) GetWorkspaceTemplate(ctx context.Context, workspaceID int64) (*models.Template, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return s.templateRepo.GetByID(ctx, workspace.TemplateID)
}

// GetWorkspacePublic returns the workspace's public status
func (s *WorkspaceService) GetWorkspacePublic(ctx context.Context, workspaceID int64) (bool, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	return workspace.IsPublic, nil
}

// GetWorkspaceOwner returns the workspace's owner information
func (s *WorkspaceService) GetWorkspaceOwner(ctx context.Context, workspaceID int64) (*models.WorkspaceOwnerResponse, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Note: In a full implementation, you would fetch the Organization details
	// For now, we return the basic info
	return &models.WorkspaceOwnerResponse{
		OwnerID:   workspace.OwnerID,
		OwnerType: "organization",
		Owner:     nil, // TODO: Fetch organization details if needed
	}, nil
}

// TransferWorkspaceOwnership transfers workspace ownership to another organization
func (s *WorkspaceService) TransferWorkspaceOwnership(ctx context.Context, workspaceID, userID int64, req *models.TransferWorkspaceOwnershipRequest) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	// TODO: Add permission checks - user should be owner/admin of both orgs

	// Update ownership
	workspace.OwnerID = req.NewOwnerID

	return s.workspaceRepo.Update(ctx, workspace)
}

// GetWorkspaceState returns the workspace's current and target state
func (s *WorkspaceService) GetWorkspaceState(ctx context.Context, workspaceID int64) (*models.WorkspaceStateResponse, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return &models.WorkspaceStateResponse{
		CurrentStatus: workspace.Status,
		TargetStatus:  workspace.TargetStatus,
	}, nil
}

// UpdateWorkspaceState updates the workspace's target state
func (s *WorkspaceService) UpdateWorkspaceState(ctx context.Context, workspaceID int64, req *models.UpdateWorkspaceStateRequest) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Update target status
	workspace.TargetStatus = req.TargetStatus

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update target status: %w", err)
	}

	// Publish appropriate event based on target status
	switch req.TargetStatus {
	case models.WorkspaceStatusRunning:
		if err := s.eventPublisher.PublishWorkspaceStart(ctx, workspace, nil, nil); err != nil {
			s.logger.Error("Failed to publish workspace start event",
				zap.Int64("workspace_id", workspace.ID),
				zap.Error(err))
		}
	case models.WorkspaceStatusStopped:
		if err := s.eventPublisher.PublishWorkspaceStop(ctx, workspace); err != nil {
			s.logger.Error("Failed to publish workspace stop event",
				zap.Int64("workspace_id", workspace.ID),
				zap.Error(err))
		}
	}

	return nil
}

// UpdateWorkspaceVolumeMounts updates all volume mounts for a workspace
func (s *WorkspaceService) UpdateWorkspaceVolumeMounts(ctx context.Context, workspaceID int64, req *models.UpdateVolumeMountsRequest) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Get current volumes
	currentVolumes, err := s.workspaceRepo.ListVolumes(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list current volumes: %w", err)
	}

	// Create a map of current volume IDs for comparison
	currentVolumeMap := make(map[int64]bool)
	for _, wv := range currentVolumes {
		currentVolumeMap[wv.VolumeID] = true
	}

	// Create a map of new volume IDs
	newVolumeMap := make(map[int64]string) // volumeID -> mountPath
	for _, v := range req.Volumes {
		newVolumeMap[v.VolumeID] = v.MountPath
	}

	// Detach volumes that are not in the new list
	for volumeID := range currentVolumeMap {
		if _, exists := newVolumeMap[volumeID]; !exists {
			if err := s.DetachVolume(ctx, workspaceID, volumeID); err != nil {
				s.logger.Warn("Failed to detach volume",
					zap.Int64("workspace_id", workspaceID),
					zap.Int64("volume_id", volumeID),
					zap.Error(err))
			}
		}
	}

	// Attach new volumes or update mount paths
	for volumeID, mountPath := range newVolumeMap {
		if !currentVolumeMap[volumeID] {
			// New volume - attach it
			if err := s.AttachVolume(ctx, workspaceID, volumeID, mountPath); err != nil {
				return fmt.Errorf("failed to attach volume %d: %w", volumeID, err)
			}
		} else {
			// Existing volume - update mount path if needed
			// Note: This is a simplified approach. In production, you might need
			// to check if mount path changed and update accordingly
			for _, wv := range currentVolumes {
				if wv.VolumeID == volumeID && wv.MountPath != mountPath {
					// Detach and reattach with new mount path
					if err := s.DetachVolume(ctx, workspaceID, volumeID); err != nil {
						return fmt.Errorf("failed to detach volume for update: %w", err)
					}
					if err := s.AttachVolume(ctx, workspaceID, volumeID, mountPath); err != nil {
						return fmt.Errorf("failed to reattach volume: %w", err)
					}
					break
				}
			}
		}
	}

	// Publish update event to notify housekeeper
	template, _ := s.templateRepo.GetByID(ctx, workspace.TemplateID)
	volumes := []*models.Volume{}
	for volumeID := range newVolumeMap {
		if vol, err := s.volumeRepo.GetByID(ctx, volumeID); err == nil {
			volumes = append(volumes, vol)
		}
	}
	
	if err := s.eventPublisher.PublishWorkspaceUpdate(ctx, workspace, template, volumes, "Volume mounts updated"); err != nil {
		s.logger.Error("Failed to publish workspace update event",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
	}

	return nil
}

