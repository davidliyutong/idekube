package services

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
)

// QuotaService handles quota checking and enforcement
type QuotaService struct {
	quotaRepo     *repository.QuotaRepository
	workspaceRepo *repository.WorkspaceRepository
	volumeRepo    *repository.VolumeRepository
}

// NewQuotaService creates a new quota service
func NewQuotaService(
	quotaRepo *repository.QuotaRepository,
	workspaceRepo *repository.WorkspaceRepository,
	volumeRepo *repository.VolumeRepository,
) *QuotaService {
	return &QuotaService{
		quotaRepo:     quotaRepo,
		workspaceRepo: workspaceRepo,
		volumeRepo:    volumeRepo,
	}
}

// QuotaUsage represents current resource usage
type QuotaUsage struct {
	CPUMillicores  int `json:"cpu_millicores"`
	MemoryMB       int `json:"memory_mb"`
	StorageMB      int `json:"storage_mb"`
	GPUCount       int `json:"gpu_count"`
	WorkspaceCount int `json:"workspace_count"`
	VolumeCount    int `json:"volume_count"`
}

// GetQuotaUsage calculates current resource usage for an owner
func (s *QuotaService) GetQuotaUsage(ctx context.Context, ownerType models.OwnerType, ownerID int64) (*QuotaUsage, error) {
	usage := &QuotaUsage{}

	// Calculate workspace resource usage
	workspaces, err := s.workspaceRepo.ListByOwner(ctx, ownerType, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}

	for _, ws := range workspaces {
		usage.CPUMillicores += ws.CPUMillicores
		usage.MemoryMB += ws.MemoryMB
		usage.StorageMB += ws.StorageMB
		// Note: GPU support will be added when Workspace model includes GPU field
		usage.WorkspaceCount++
	}

	// Calculate volume usage
	volumes, err := s.volumeRepo.ListByOwner(ctx, ownerType, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get volumes: %w", err)
	}

	for _, vol := range volumes {
		usage.StorageMB += vol.SizeMB
		usage.VolumeCount++
	}

	return usage, nil
}

// CheckQuota verifies if a resource request is within quota limits
func (s *QuotaService) CheckQuota(ctx context.Context, ownerType models.OwnerType, ownerID int64, req *models.QuotaRequest) error {
	// Get quota for owner
	quota, err := s.quotaRepo.GetByOwner(ctx, ownerType, ownerID)
	if err != nil {
		return fmt.Errorf("failed to get quota: %w", err)
	}

	// Get current usage
	usage, err := s.GetQuotaUsage(ctx, ownerType, ownerID)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	// Check CPU limit
	if quota.MaxCPUMillicores != nil {
		if usage.CPUMillicores+req.CPUMillicores > *quota.MaxCPUMillicores {
			return fmt.Errorf("CPU quota exceeded: current %d, requested %d, limit %d",
				usage.CPUMillicores, req.CPUMillicores, *quota.MaxCPUMillicores)
		}
	}

	// Check Memory limit
	if quota.MaxMemoryMB != nil {
		if usage.MemoryMB+req.MemoryMB > *quota.MaxMemoryMB {
			return fmt.Errorf("Memory quota exceeded: current %d, requested %d, limit %d",
				usage.MemoryMB, req.MemoryMB, *quota.MaxMemoryMB)
		}
	}

	// Check Storage limit
	if quota.MaxStorageMB != nil {
		if usage.StorageMB+req.StorageMB > *quota.MaxStorageMB {
			return fmt.Errorf("Storage quota exceeded: current %d, requested %d, limit %d",
				usage.StorageMB, req.StorageMB, *quota.MaxStorageMB)
		}
	}

	// Check GPU limit
	if quota.MaxGPU != nil && req.GPUCount > 0 {
		if usage.GPUCount+req.GPUCount > *quota.MaxGPU {
			return fmt.Errorf("GPU quota exceeded: current %d, requested %d, limit %d",
				usage.GPUCount, req.GPUCount, *quota.MaxGPU)
		}
	}

	// Check workspace count limit
	if quota.MaxWorkspaces != nil && req.WorkspaceCount > 0 {
		if usage.WorkspaceCount+req.WorkspaceCount > *quota.MaxWorkspaces {
			return fmt.Errorf("Workspace count quota exceeded: current %d, requested %d, limit %d",
				usage.WorkspaceCount, req.WorkspaceCount, *quota.MaxWorkspaces)
		}
	}

	// Check volume count limit
	if quota.MaxVolumes != nil && req.VolumeCount > 0 {
		if usage.VolumeCount+req.VolumeCount > *quota.MaxVolumes {
			return fmt.Errorf("Volume count quota exceeded: current %d, requested %d, limit %d",
				usage.VolumeCount, req.VolumeCount, *quota.MaxVolumes)
		}
	}

	return nil
}

// UpdateQuota updates quota limits for an owner
func (s *QuotaService) UpdateQuota(ctx context.Context, ownerType models.OwnerType, ownerID int64, req *models.UpdateQuotaRequest) (*models.Quota, error) {
	quota, err := s.quotaRepo.GetByOwner(ctx, ownerType, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	// Update fields if provided
	if req.MaxCPUMillicores != nil {
		quota.MaxCPUMillicores = req.MaxCPUMillicores
	}
	if req.MaxMemoryMB != nil {
		quota.MaxMemoryMB = req.MaxMemoryMB
	}
	if req.MaxStorageMB != nil {
		quota.MaxStorageMB = req.MaxStorageMB
	}
	if req.MaxGPU != nil {
		quota.MaxGPU = req.MaxGPU
	}
	if req.MaxWorkspaces != nil {
		quota.MaxWorkspaces = req.MaxWorkspaces
	}
	if req.MaxVolumes != nil {
		quota.MaxVolumes = req.MaxVolumes
	}

	err = s.quotaRepo.Update(ctx, quota)
	if err != nil {
		return nil, fmt.Errorf("failed to update quota: %w", err)
	}

	return quota, nil
}
