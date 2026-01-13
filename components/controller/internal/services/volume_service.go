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
	eventPublisher *queue.EventPublisher
	logger         *zap.Logger
}

// NewVolumeService creates a new volume service
func NewVolumeService(
	volumeRepo *repository.VolumeRepository,
	eventPublisher *queue.EventPublisher,
	logger *zap.Logger,
) *VolumeService {
	return &VolumeService{
		volumeRepo:     volumeRepo,
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
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
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
	oldSize := ""
	if req.SizeMB != nil && *req.SizeMB != volume.SizeMB {
		if *req.SizeMB < volume.SizeMB {
			return nil, fmt.Errorf("cannot shrink volume size")
		}

		oldSize = fmt.Sprintf("%dMB", volume.SizeMB)
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
		if err := s.eventPublisher.PublishVolumeResize(ctx, volume, oldSize); err != nil {
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

