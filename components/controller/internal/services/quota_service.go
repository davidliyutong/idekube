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

// QuotaRequest represents a resource request to check against quota
type QuotaRequest struct {
	CPUMillicores  int `json:"cpu_millicores"`
	MemoryMB       int `json:"memory_mb"`
	StorageMB      int `json:"storage_mb"`
	GPUCount       int `json:"gpu_count"`
	WorkspaceCount int `json:"workspace_count"`
	VolumeCount    int `json:"volume_count"`
}

// GetQuotaUsageByOrganization calculates current resource usage for an organization
func (s *QuotaService) GetQuotaUsageByOrganization(ctx context.Context, orgID int64) (*QuotaUsage, error) {
	usage := &QuotaUsage{}

	// Calculate workspace resource usage
	opts := &models.ListOptions{}
	opts.Page = 1
	opts.PageSize = 10000
	workspaces, _, err := s.workspaceRepo.ListByOrganization(ctx, orgID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}

	for _, ws := range workspaces {
		if ws.Quota.CPUMillicores != nil {
			usage.CPUMillicores += *ws.Quota.CPUMillicores
		}
		if ws.Quota.MemoryMB != nil {
			usage.MemoryMB += *ws.Quota.MemoryMB
		}
		if ws.Quota.StorageMB != nil {
			usage.StorageMB += *ws.Quota.StorageMB
		}
		if ws.Quota.GPU != nil {
			usage.GPUCount += *ws.Quota.GPU
		}
		usage.WorkspaceCount++
	}

	// Calculate volume usage
	volumes, _, err := s.volumeRepo.ListByOrganization(ctx, orgID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get volumes: %w", err)
	}

	for _, vol := range volumes {
		usage.StorageMB += vol.SizeMB
		usage.VolumeCount++
	}

	return usage, nil
}

// CheckQuotaByOrganization verifies if a resource request is within quota limits
func (s *QuotaService) CheckQuotaByOrganization(ctx context.Context, orgID int64, req *QuotaRequest) error {
	// Get quota for organization
	quota, err := s.quotaRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get quota: %w", err)
	}

	// Get current usage
	usage, err := s.GetQuotaUsageByOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	// Check CPU limit (QuotaLimits is embedded in Quota)
	if quota.CPUMillicores != nil {
		if usage.CPUMillicores+req.CPUMillicores > *quota.CPUMillicores {
			return fmt.Errorf("CPU quota exceeded: current %d, requested %d, limit %d",
				usage.CPUMillicores, req.CPUMillicores, *quota.CPUMillicores)
		}
	}

	// Check Memory limit
	if quota.MemoryMB != nil {
		if usage.MemoryMB+req.MemoryMB > *quota.MemoryMB {
			return fmt.Errorf("Memory quota exceeded: current %d, requested %d, limit %d",
				usage.MemoryMB, req.MemoryMB, *quota.MemoryMB)
		}
	}

	// Check Storage limit
	if quota.StorageMB != nil {
		if usage.StorageMB+req.StorageMB > *quota.StorageMB {
			return fmt.Errorf("Storage quota exceeded: current %d, requested %d, limit %d",
				usage.StorageMB, req.StorageMB, *quota.StorageMB)
		}
	}

	// Check GPU limit
	if quota.GPU != nil && req.GPUCount > 0 {
		if usage.GPUCount+req.GPUCount > *quota.GPU {
			return fmt.Errorf("GPU quota exceeded: current %d, requested %d, limit %d",
				usage.GPUCount, req.GPUCount, *quota.GPU)
		}
	}

	// Check workspace count limit
	// FIXME: [nitpick] The field name Workspaces in quota is ambiguous as it could refer to workspace instances or workspace count. Consider renaming to MaxWorkspaces for clarity.
	if quota.Workspaces != nil && req.WorkspaceCount > 0 {
		if usage.WorkspaceCount+req.WorkspaceCount > *quota.Workspaces {
			return fmt.Errorf("Workspace count quota exceeded: current %d, requested %d, limit %d",
				usage.WorkspaceCount, req.WorkspaceCount, *quota.Workspaces)
		}
	}

	// Check volume count limit
	if quota.Volumes != nil && req.VolumeCount > 0 {
		if usage.VolumeCount+req.VolumeCount > *quota.Volumes {
			return fmt.Errorf("Volume count quota exceeded: current %d, requested %d, limit %d",
				usage.VolumeCount, req.VolumeCount, *quota.Volumes)
		}
	}

	return nil
}

// UpdateQuotaByOrganization updates quota limits for an organization
func (s *QuotaService) UpdateQuotaByOrganization(ctx context.Context, orgID int64, req *models.UpdateQuotaRequest) (*models.Quota, error) {
	quota, err := s.quotaRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	// Update fields if provided (embedded QuotaLimits fields)
	if req.CPUMillicores != nil {
		quota.CPUMillicores = req.CPUMillicores
	}
	if req.MemoryMB != nil {
		quota.MemoryMB = req.MemoryMB
	}
	if req.StorageMB != nil {
		quota.StorageMB = req.StorageMB
	}
	if req.GPU != nil {
		quota.GPU = req.GPU
	}
	if req.Workspaces != nil {
		quota.Workspaces = req.Workspaces
	}
	if req.Volumes != nil {
		quota.Volumes = req.Volumes
	}
	if req.TimeoutSeconds != nil {
		quota.TimeoutSeconds = req.TimeoutSeconds
	}

	err = s.quotaRepo.Update(ctx, quota)
	if err != nil {
		return nil, fmt.Errorf("failed to update quota: %w", err)
	}

	return quota, nil
}
