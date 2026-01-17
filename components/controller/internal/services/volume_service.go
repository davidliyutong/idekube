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

// VolumeService handles volume business logic
type VolumeService struct {
	volumeRepo        *repository.VolumeRepository
	eventPublisher    *queue.EventPublisher
	logger            *zap.Logger
	permissionService *permission.ResourcePermissionService
}

// NewVolumeService creates a new volume service
func NewVolumeService(
	volumeRepo *repository.VolumeRepository,
	eventPublisher *queue.EventPublisher,
	logger *zap.Logger,
	permissionService *permission.ResourcePermissionService,
) *VolumeService {
	return &VolumeService{
		volumeRepo:        volumeRepo,
		eventPublisher:    eventPublisher,
		logger:            logger,
		permissionService: permissionService,
	}
}

// CreateVolume creates a new volume
func (s *VolumeService) CreateVolume(ctx context.Context, req *models.CreateVolumeRequest) (*models.Volume, error) {
	// Set default access mode if not provided
	accessMode := req.AccessMode
	if accessMode == "" {
		accessMode = models.VolumeAccessModeReadWriteOnce
	}

	// Create labels for RBAC
	labels := models.ResourceLabels{
		"owner_type": string(req.OwnerType),
		"owner_id":   fmt.Sprintf("%d", req.OwnerID),
	}

	// Add organization_id label if it's an org volume
	if req.OwnerType == models.OwnerTypeOrganization {
		labels["organization_id"] = fmt.Sprintf("%d", req.OwnerID)
	}

	volume := &models.Volume{
		UUID:         uuid.New(),
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		OwnerType:    req.OwnerType,
		OwnerID:      req.OwnerID,
		SizeMB:       req.SizeMB,
		StorageClass: req.StorageClass,
		AccessMode:   accessMode,
		Status:       models.VolumeStatusPending,
		Labels:       labels,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create in database first
	err := s.volumeRepo.Create(ctx, volume)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	// Grant ownership permissions automatically (if permission service is available)
	if s.permissionService != nil {
		// Determine the actual user who created this volume
		var creatorUserID int64
		if req.OwnerType == models.OwnerTypeUser {
			creatorUserID = req.OwnerID
		} else {
			// For organization volumes, the CreatedBy field should be set by the handler
			// For now, we'll grant ownership to the OwnerID (which is the org)
			// TODO: Track actual creator separately and grant them permissions
			creatorUserID = req.OwnerID
		}

		if err := s.permissionService.GrantResourceOwnership(ctx, creatorUserID, "volume", volume.ID); err != nil {
			s.logger.Warn("Failed to grant volume ownership permissions",
				zap.Int64("volume_id", volume.ID),
				zap.Int64("creator_id", creatorUserID),
				zap.Error(err))
		}
	}

	// Publish volume creation event to HouseKeeper
	if err := s.eventPublisher.PublishVolumeCreate(ctx, volume); err != nil {
		s.logger.Error("Failed to publish volume create event",
			zap.Int64("volume_id", volume.ID),
			zap.Error(err))
		// Don't rollback - HouseKeeper reconciler will handle it
	}

	s.logger.Info("Volume created, event published",
		zap.Int64("volume_id", volume.ID),
		zap.String("name", volume.Name))

	return volume, nil
}

// GetVolume retrieves a volume by ID
func (s *VolumeService) GetVolume(ctx context.Context, id int64) (*models.Volume, error) {
	return s.volumeRepo.GetByID(ctx, id)
}

// ListVolumesByOwner lists volumes owned by a specific owner
func (s *VolumeService) ListVolumesByOwner(ctx context.Context, ownerType models.OwnerType, ownerID int64) ([]*models.Volume, error) {
	return s.volumeRepo.ListByOwner(ctx, ownerType, ownerID)
}

// UpdateVolume updates a volume
func (s *VolumeService) UpdateVolume(ctx context.Context, id int64, req *models.UpdateVolumeRequest) (*models.Volume, error) {
	volume, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.DisplayName != nil {
		volume.DisplayName = req.DisplayName
	}
	if req.Description != nil {
		volume.Description = req.Description
	}

	// Handle size change (requires PVC resize)
	needsResize := false
	var oldSizeMB int64
	if req.SizeMB != nil && *req.SizeMB != volume.SizeMB {
		if *req.SizeMB < volume.SizeMB {
			return nil, fmt.Errorf("cannot shrink volume size")
		}

		oldSizeMB = int64(volume.SizeMB)
		volume.SizeMB = *req.SizeMB
		needsResize = true
	}

	// Update volume in database
	err = s.volumeRepo.Update(ctx, volume)
	if err != nil {
		return nil, fmt.Errorf("failed to update volume: %w", err)
	}

	// Publish resize event if needed
	if needsResize {
		if err := s.eventPublisher.PublishVolumeResize(ctx, volume, oldSizeMB, int64(volume.SizeMB)); err != nil {
			s.logger.Error("Failed to publish volume resize event",
				zap.Int64("volume_id", volume.ID),
				zap.Error(err))
		}
	}

	return volume, nil
}

// DeleteVolume deletes a volume
func (s *VolumeService) DeleteVolume(ctx context.Context, id int64) error {
	volume, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Publish delete event to HouseKeeper
	if err := s.eventPublisher.PublishVolumeDelete(ctx, id, volume); err != nil {
		s.logger.Error("Failed to publish volume delete event",
			zap.Int64("volume_id", id),
			zap.Error(err))
		// Continue with deletion anyway
	}

	// Delete from database
	return s.volumeRepo.Delete(ctx, id)
}

// SyncVolumeStatus syncs volume status from database (status updated by HouseKeeper)
func (s *VolumeService) SyncVolumeStatus(ctx context.Context, id int64) error {
	// This method can be used to force a refresh from the database
	volume, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	s.logger.Info("Volume status synced",
		zap.Int64("volume_id", volume.ID),
		zap.String("status", string(volume.Status)))

	return nil
}

// ListVolumes lists volumes based on user role and permissions
func (s *VolumeService) ListVolumes(ctx context.Context, userID int64, userRole models.UserRole, opts *models.ListOptions) ([]*models.Volume, int64, error) {
	switch userRole {
	case models.UserRoleSuperAdmin, models.UserRoleAdmin:
		// Admins can see all volumes
		return s.volumeRepo.ListAll(ctx, opts)

	case models.UserRolePowerUser, models.UserRoleUser:
		// Regular users and power users see their own volumes
		return s.volumeRepo.ListAccessibleByUser(ctx, userID, opts)

	default:
		return nil, 0, fmt.Errorf("unknown user role: %s", userRole)
	}
}

// ListVolumesByOrganization lists volumes in a specific organization
func (s *VolumeService) ListVolumesByOrganization(ctx context.Context, orgID int64, opts *models.ListOptions) ([]*models.Volume, int64, error) {
	labels := map[string]string{
		"organization_id": fmt.Sprintf("%d", orgID),
	}
	return s.volumeRepo.ListByLabel(ctx, labels, opts)
}
