package models

import (
	"time"
)

// Quota represents resource quotas for an organization
// Quota only belongs to Organization (removed OwnerType)
type Quota struct {
	ID             int64       `json:"id" gorm:"primaryKey"`
	OrganizationID int64       `json:"organization_id" gorm:"uniqueIndex;not null"`
	QuotaLimits    `gorm:"embedded"` // Embedded QuotaLimits fields
	CreatedAt      time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Quota
func (Quota) TableName() string {
	return "quotas"
}

// ============================================================================
// Request/Response Types
// ============================================================================

// CreateQuotaRequest represents the request to create a quota
type CreateQuotaRequest struct {
	OrganizationID int64 `json:"organization_id" binding:"required"`
	CPUMillicores  *int  `json:"cpu_millicores,omitempty"`
	MemoryMB       *int  `json:"memory_mb,omitempty"`
	StorageMB      *int  `json:"storage_mb,omitempty"`
	GPU            *int  `json:"gpu,omitempty"`
	Workspaces     *int  `json:"workspaces,omitempty"`
	Volumes        *int  `json:"volumes,omitempty"`
	TimeoutSeconds *int  `json:"timeout_seconds,omitempty"`
}

// UpdateQuotaRequest represents the request to update a quota
type UpdateQuotaRequest struct {
	CPUMillicores  *int `json:"cpu_millicores,omitempty" binding:"omitempty,min=0"`
	MemoryMB       *int `json:"memory_mb,omitempty" binding:"omitempty,min=0"`
	StorageMB      *int `json:"storage_mb,omitempty" binding:"omitempty,min=0"`
	GPU            *int `json:"gpu,omitempty" binding:"omitempty,min=0"`
	Workspaces     *int `json:"workspaces,omitempty" binding:"omitempty,min=1"`
	Volumes        *int `json:"volumes,omitempty" binding:"omitempty,min=1"`
	TimeoutSeconds *int `json:"timeout_seconds,omitempty" binding:"omitempty,min=0"`
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

// QuotaCheckRequest represents a resource request for quota checking
type QuotaCheckRequest struct {
	CPUMillicores  int `json:"cpu_millicores"`
	MemoryMB       int `json:"memory_mb"`
	StorageMB      int `json:"storage_mb"`
	GPUCount       int `json:"gpu_count"`
	WorkspaceCount int `json:"workspace_count"`
	VolumeCount    int `json:"volume_count"`
}
