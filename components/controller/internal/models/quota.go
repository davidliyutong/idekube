package models

import (
	"time"
)

// OwnerType represents the type of resource owner
type OwnerType string

const (
	OwnerTypeUser         OwnerType = "user"
	OwnerTypeOrganization OwnerType = "organization"
)

// Quota represents resource quotas
type Quota struct {
	ID                int64      `json:"id" gorm:"primaryKey"`
	OwnerType         OwnerType  `json:"owner_type" gorm:"type:varchar(50);not null"`
	OwnerID           int64      `json:"owner_id" gorm:"not null"`
	MaxCPUMillicores  *int       `json:"max_cpu_millicores" gorm:"column:cpu_millicores;default:8000"`
	MaxMemoryMB       *int       `json:"max_memory_mb" gorm:"column:memory_mb;default:16384"`
	MaxStorageMB      *int       `json:"max_storage_mb" gorm:"column:storage_mb;default:51200"`
	MaxGPU            *int       `json:"max_gpu" gorm:"column:gpu_count;default:0"`
	MaxWorkspaces     *int       `json:"max_workspaces" gorm:"default:10"`
	MaxVolumes        *int       `json:"max_volumes" gorm:"default:20"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Quota) TableName() string {
	return "quotas"
}

// CreateQuotaRequest represents the request to create a quota
type CreateQuotaRequest struct {
	OwnerType        OwnerType `json:"owner_type" binding:"required"`
	OwnerID          int64     `json:"owner_id" binding:"required"`
	MaxCPUMillicores int       `json:"max_cpu_millicores" binding:"required,min=0"`
	MaxMemoryMB      int       `json:"max_memory_mb" binding:"required,min=0"`
	MaxStorageMB     int       `json:"max_storage_mb" binding:"required,min=0"`
	MaxGPU           int       `json:"max_gpu,omitempty"`
	MaxWorkspaces    int       `json:"max_workspaces" binding:"required,min=1"`
	MaxVolumes       int       `json:"max_volumes" binding:"required,min=1"`
}

// UpdateQuotaRequest represents the request to update a quota
type UpdateQuotaRequest struct {
	MaxCPUMillicores *int `json:"max_cpu_millicores,omitempty" binding:"omitempty,min=0"`
	MaxMemoryMB      *int `json:"max_memory_mb,omitempty" binding:"omitempty,min=0"`
	MaxStorageMB     *int `json:"max_storage_mb,omitempty" binding:"omitempty,min=0"`
	MaxGPU           *int `json:"max_gpu,omitempty"`
	MaxWorkspaces    *int `json:"max_workspaces,omitempty" binding:"omitempty,min=1"`
	MaxVolumes       *int `json:"max_volumes,omitempty" binding:"omitempty,min=1"`
}

// QuotaUsage represents current resource usage against quotas
type QuotaUsage struct {
	Quota             Quota `json:"quota"`
	UsedCPUMillicores int   `json:"used_cpu_millicores"`
	UsedMemoryMB      int   `json:"used_memory_mb"`
	UsedStorageMB     int   `json:"used_storage_mb"`
	UsedGPUCount      int   `json:"used_gpu_count"`
	UsedWorkspaces    int   `json:"used_workspaces"`
	UsedVolumes       int   `json:"used_volumes"`
}

// QuotaRequest represents a resource request for quota checking
type QuotaRequest struct {
	CPUMillicores  int `json:"cpu_millicores"`
	MemoryMB       int `json:"memory_mb"`
	StorageMB      int `json:"storage_mb"`
	GPUCount       int `json:"gpu_count"`
	WorkspaceCount int `json:"workspace_count"`
	VolumeCount    int `json:"volume_count"`
}
