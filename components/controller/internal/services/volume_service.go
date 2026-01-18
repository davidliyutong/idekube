package services

import (
	"context"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/pkg/queue"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// VolumeService handles volume business logic
type VolumeService struct {
	volumeRepo     *repository.VolumeRepository
	workspaceRepo  *repository.WorkspaceRepository
	eventPublisher *queue.EventPublisher
	logger         *zap.Logger
}

// NewVolumeService creates a new volume service
func NewVolumeService(
	volumeRepo *repository.VolumeRepository,
	workspaceRepo *repository.WorkspaceRepository,
	eventPublisher *queue.EventPublisher,
	logger *zap.Logger,
) *VolumeService {
	return &VolumeService{
		volumeRepo:     volumeRepo,
		workspaceRepo:  workspaceRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// CreateVolume creates a new volume
func (s *VolumeService) CreateVolume(ctx context.Context, req *models.CreateVolumeRequest) (*models.Volume, error) {
	// Set default access mode if not provided
	accessMode := req.AccessMode
	if accessMode == "" {
		accessMode = models.VolumeAccessModeReadWriteOnce
	}

	now := time.Now()
	volume := &models.Volume{
		Base: models.Base{
			UUID:      uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Status:    models.VolumeStatusPending,
			Labels:    req.Labels,
		},
		Profile: models.Profile{
			Identifier:  req.Name,
			DisplayName: req.DisplayName,
			Description: req.Description,
			IconURL:     req.IconURL,
		},
		SizeMB:       req.SizeMB,
		StorageClass: req.StorageClass,
		AccessMode:   accessMode,
		IsPublic:     req.IsPublic,
		OwnerID:      req.OwnerID, // Organization ID
	}

	// Create in database first
	err := s.volumeRepo.Create(ctx, volume)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
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
		zap.String("name", volume.Identifier))

	return volume, nil
}

// GetVolume retrieves a volume by ID
func (s *VolumeService) GetVolume(ctx context.Context, id int64) (*models.Volume, error) {
	return s.volumeRepo.GetByID(ctx, id)
}

// ListVolumesByOrganization lists volumes by organization
func (s *VolumeService) ListVolumesByOrganization(ctx context.Context, orgID int64, opts *models.ListOptions) ([]*models.Volume, int64, error) {
	return s.volumeRepo.ListByOrganization(ctx, orgID, opts)
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
	if req.IconURL != nil {
		volume.IconURL = req.IconURL
	}
	if req.IsPublic != nil {
		volume.IsPublic = *req.IsPublic
	}
	if req.Labels != nil {
		volume.Labels = req.Labels
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
		zap.String("status", volume.Status))

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

// ============================================================================
// Sub-resource APIs
// ============================================================================

// GetVolumeProfile returns the volume's profile sub-resource
func (s *VolumeService) GetVolumeProfile(ctx context.Context, volumeID int64) (*models.VolumeProfileResponse, error) {
	volume, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return nil, err
	}

	return &models.VolumeProfileResponse{
		Identifier:  volume.Identifier,
		DisplayName: volume.DisplayName,
		IconURL:     volume.IconURL,
		Description: volume.Description,
	}, nil
}

// UpdateVolumeProfile updates the volume's profile sub-resource
func (s *VolumeService) UpdateVolumeProfile(ctx context.Context, volumeID int64, req *models.UpdateVolumeProfileRequest) (*models.VolumeProfileResponse, error) {
	volume, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return nil, err
	}

	// Update profile fields
	if req.DisplayName != nil {
		volume.DisplayName = req.DisplayName
	}
	if req.IconURL != nil {
		volume.IconURL = req.IconURL
	}
	if req.Description != nil {
		volume.Description = req.Description
	}

	err = s.volumeRepo.Update(ctx, volume)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &models.VolumeProfileResponse{
		Identifier:  volume.Identifier,
		DisplayName: volume.DisplayName,
		IconURL:     volume.IconURL,
		Description: volume.Description,
	}, nil
}

// GetVolumeMounts returns the volume's mounts sub-resource
func (s *VolumeService) GetVolumeMounts(ctx context.Context, volumeID int64) (*models.VolumeMountsResponse, error) {
	// Get all workspaces that have this volume attached
	if s.workspaceRepo == nil {
		return nil, fmt.Errorf("workspace repository not available")
	}

	mounts, err := s.volumeRepo.GetVolumeMounts(ctx, volumeID)
	if err != nil {
		return nil, err
	}

	return &models.VolumeMountsResponse{
		Mounts: mounts,
	}, nil
}

// UpdateVolumeSize updates the volume's size (expand only).
// Note: Expanding a volume may require workspace restarts or downtime depending
// on the underlying storage implementation, and can have cloud provider cost
// implications due to increased provisioned capacity.
func (s *VolumeService) UpdateVolumeSize(ctx context.Context, volumeID int64, req *models.UpdateVolumeSizeRequest) error {
	volume, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return err
	}

	// Check that new size is larger
	if req.SizeMB < volume.SizeMB {
		return fmt.Errorf("cannot shrink volume size: current %d MB, requested %d MB", volume.SizeMB, req.SizeMB)
	}

	if req.SizeMB == volume.SizeMB {
		return nil // No change needed
	}

	oldSizeMB := int64(volume.SizeMB)
	volume.SizeMB = req.SizeMB

	err = s.volumeRepo.Update(ctx, volume)
	if err != nil {
		return fmt.Errorf("failed to update volume size: %w", err)
	}

	// Publish resize event
	if err := s.eventPublisher.PublishVolumeResize(ctx, volume, oldSizeMB, int64(volume.SizeMB)); err != nil {
		s.logger.Error("Failed to publish volume resize event",
			zap.Int64("volume_id", volume.ID),
			zap.Error(err))
	}

	return nil
}

// UpdateVolumeIsPublic updates the volume's is_public status
func (s *VolumeService) UpdateVolumeIsPublic(ctx context.Context, volumeID int64, req *models.UpdateVolumeIsPublicRequest) error {
	volume, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return err
	}

	volume.IsPublic = req.IsPublic

	return s.volumeRepo.Update(ctx, volume)
}

// ============================================================================
// Additional Sub-resource APIs
// ============================================================================

// GetVolumeSizeMB returns the volume's size in MB
func (s *VolumeService) GetVolumeSizeMB(ctx context.Context, volumeID int64) (int, error) {
	volume, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return 0, err
	}

	return volume.SizeMB, nil
}

// GetVolumeStorageClass returns the volume's storage class
func (s *VolumeService) GetVolumeStorageClass(ctx context.Context, volumeID int64) (*string, error) {
	volume, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return nil, err
	}

	return volume.StorageClass, nil
}

// GetVolumeAccessMode returns the volume's access mode
func (s *VolumeService) GetVolumeAccessMode(ctx context.Context, volumeID int64) (models.VolumeAccessMode, error) {
	volume, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return "", err
	}

	return volume.AccessMode, nil
}

// GetVolumeOwner returns the volume's owner information
func (s *VolumeService) GetVolumeOwner(ctx context.Context, volumeID int64) (*models.VolumeOwnerResponse, error) {
	volume, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return nil, err
	}

	// Note: In a full implementation, you would fetch the Organization details
	// For now, we return the basic info
	return &models.VolumeOwnerResponse{
		OwnerID:   volume.OwnerID,
		OwnerType: "organization",
		Owner:     nil, // TODO: Fetch organization details if needed
	}, nil
}

// TransferVolumeOwnership transfers volume ownership to another organization
func (s *VolumeService) TransferVolumeOwnership(ctx context.Context, volumeID, userID int64, req *models.TransferVolumeOwnershipRequest) error {
	volume, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return err
	}

	// TODO: Add permission checks - user should be owner/admin of both orgs

	// Update ownership
	volume.OwnerID = req.NewOwnerID

	return s.volumeRepo.Update(ctx, volume)
}

// GetVolumePublic returns the volume's public status
func (s *VolumeService) GetVolumePublic(ctx context.Context, volumeID int64) (bool, error) {
	volume, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return false, err
	}

	return volume.IsPublic, nil
}

