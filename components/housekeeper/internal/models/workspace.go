package models

import (
	"time"
)

// WorkspaceStatus represents the status of a workspace
type WorkspaceStatus string

const (
	WorkspaceStatusPending  WorkspaceStatus = "pending"
	WorkspaceStatusStarting WorkspaceStatus = "starting"
	WorkspaceStatusRunning  WorkspaceStatus = "running"
	WorkspaceStatusStopped  WorkspaceStatus = "stopped"
	WorkspaceStatusFailed   WorkspaceStatus = "failed"
)

// OwnerType represents the type of owner
type OwnerType string

const (
	OwnerTypeUser         OwnerType = "user"
	OwnerTypeOrganization OwnerType = "organization"
)

// Workspace represents a development workspace
type Workspace struct {
	ID                int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID              string          `gorm:"type:uuid;unique;not null" json:"uuid"`
	Name              string          `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	DisplayName       *string         `gorm:"type:varchar(255)" json:"display_name,omitempty"`
	Description       *string         `gorm:"type:text" json:"description,omitempty"`
	OwnerType         OwnerType       `gorm:"type:varchar(50);not null" json:"owner_type"`
	OwnerID           int64           `gorm:"not null;index:idx_owner" json:"owner_id"`
	TemplateID        int64           `gorm:"not null" json:"template_id"`
	CPUMillicores     int64           `gorm:"not null" json:"cpu_millicores"`
	MemoryMB          int64           `gorm:"not null" json:"memory_mb"`
	StorageMB         int64           `gorm:"not null" json:"storage_mb"`
	CurrentStatus     WorkspaceStatus `gorm:"type:varchar(50);not null;index" json:"current_status"`
	TargetStatus      WorkspaceStatus `gorm:"type:varchar(50);not null" json:"target_status"`
	K8sNamespace      *string         `gorm:"type:varchar(255)" json:"k8s_namespace,omitempty"`
	K8sDeploymentName *string         `gorm:"type:varchar(255)" json:"k8s_deployment_name,omitempty"`
	K8sServiceName    *string         `gorm:"type:varchar(255)" json:"k8s_service_name,omitempty"`
	TimeoutSeconds    *int64          `gorm:"default:3600" json:"timeout_seconds,omitempty"`
	AccessedAt        *time.Time      `gorm:"type:timestamp with time zone" json:"accessed_at,omitempty"`
	CreatedAt         time.Time       `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"type:timestamp with time zone;not null;default:now()" json:"updated_at"`
	DeletedAt         *time.Time      `gorm:"type:timestamp with time zone;index" json:"deleted_at,omitempty"`
}

// Template represents a workspace template
type Template struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID                 string    `gorm:"type:uuid;unique;not null" json:"uuid"`
	Name                 string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	DisplayName          *string   `gorm:"type:varchar(255)" json:"display_name,omitempty"`
	Description          *string   `gorm:"type:text" json:"description,omitempty"`
	Image                string    `gorm:"type:varchar(500);not null" json:"image"`
	Command              *string   `gorm:"type:text" json:"command,omitempty"`
	Args                 *string   `gorm:"type:text" json:"args,omitempty"`
	WorkingDir           *string   `gorm:"type:varchar(500)" json:"working_dir,omitempty"`
	DefaultCPUMillicores int64     `gorm:"not null;default:1000" json:"default_cpu_millicores"`
	DefaultMemoryMB      int64     `gorm:"not null;default:2048" json:"default_memory_mb"`
	DefaultStorageMB     int64     `gorm:"not null;default:10240" json:"default_storage_mb"`
	MinCPUMillicores     int64     `gorm:"not null;default:100" json:"min_cpu_millicores"`
	MinMemoryMB          int64     `gorm:"not null;default:256" json:"min_memory_mb"`
	MaxCPUMillicores     *int64    `json:"max_cpu_millicores,omitempty"`
	MaxMemoryMB          *int64    `json:"max_memory_mb,omitempty"`
	IsPublic             bool      `gorm:"not null;default:false;index" json:"is_public"`
	CreatorID            *int64    `json:"creator_id,omitempty"`
	OrganizationID       *int64    `json:"organization_id,omitempty"`
	CreatedAt            time.Time `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	UpdatedAt            time.Time `gorm:"type:timestamp with time zone;not null;default:now()" json:"updated_at"`
}

// Volume represents a persistent volume
type Volume struct {
	ID           int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID         string       `gorm:"type:uuid;unique;not null" json:"uuid"`
	Name         string       `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	DisplayName  *string      `gorm:"type:varchar(255)" json:"display_name,omitempty"`
	Description  *string      `gorm:"type:text" json:"description,omitempty"`
	OwnerType    OwnerType    `gorm:"type:varchar(50);not null" json:"owner_type"`
	OwnerID      int64        `gorm:"not null;index:idx_volume_owner" json:"owner_id"`
	SizeMB       int64        `gorm:"not null" json:"size_mb"`
	StorageClass *string      `gorm:"type:varchar(255)" json:"storage_class,omitempty"`
	AccessMode   string       `gorm:"type:varchar(50);not null;default:'ReadWriteOnce'" json:"access_mode"`
	Status       VolumeStatus `gorm:"type:varchar(50);not null;index" json:"status"`
	PVCName      *string      `gorm:"type:varchar(255)" json:"pvc_name,omitempty"`
	CreatedAt    time.Time    `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time    `gorm:"type:timestamp with time zone;not null;default:now()" json:"updated_at"`
}

// VolumeStatus represents the status of a volume
type VolumeStatus string

const (
	VolumeStatusPending VolumeStatus = "pending"
	VolumeStatusBound   VolumeStatus = "bound"
	VolumeStatusFailed  VolumeStatus = "failed"
)
