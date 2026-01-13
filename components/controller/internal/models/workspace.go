package models

import (
	"time"

	"github.com/google/uuid"
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

// Workspace represents a user workspace (IDE instance)
type Workspace struct {
	ID                 int64           `json:"id" db:"id"`
	UUID               uuid.UUID       `json:"uuid" db:"uuid"`
	Name               string          `json:"name" db:"name"`
	DisplayName        *string         `json:"display_name,omitempty" db:"display_name"`
	Description        *string         `json:"description,omitempty" db:"description"`
	OwnerType          OwnerType       `json:"owner_type" db:"owner_type"`
	OwnerID            int64           `json:"owner_id" db:"owner_id"`
	TemplateID         int64           `json:"template_id" db:"template_id"`
	CPUMillicores      int             `json:"cpu_millicores" db:"cpu_millicores"`
	MemoryMB           int             `json:"memory_mb" db:"memory_mb"`
	StorageMB          int             `json:"storage_mb" db:"storage_mb"`
	CurrentStatus      WorkspaceStatus `json:"current_status" db:"current_status"`
	TargetStatus       WorkspaceStatus `json:"target_status" db:"target_status"`
	K8sNamespace       *string         `json:"k8s_namespace,omitempty" db:"k8s_namespace"`
	K8sDeploymentName  *string         `json:"k8s_deployment_name,omitempty" db:"k8s_deployment_name"`
	K8sServiceName     *string         `json:"k8s_service_name,omitempty" db:"k8s_service_name"`
	CreatedBy          int64           `json:"created_by" db:"created_by"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
	StartedAt          *time.Time      `json:"started_at,omitempty" db:"started_at"`
	AccessedAt         *time.Time      `json:"accessed_at,omitempty" db:"accessed_at"`
	TimeoutSeconds     int             `json:"timeout_seconds" db:"timeout_seconds"`
}

// WorkspaceVolume represents a volume attached to a workspace
type WorkspaceVolume struct {
	ID          int64     `json:"id" db:"id"`
	WorkspaceID int64     `json:"workspace_id" db:"workspace_id"`
	VolumeID    int64     `json:"volume_id" db:"volume_id"`
	MountPath   string    `json:"mount_path" db:"mount_path"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// WorkspaceWithDetails represents a workspace with additional details
type WorkspaceWithDetails struct {
	Workspace
	Template *Template `json:"template,omitempty"`
	Volumes  []Volume  `json:"volumes,omitempty"`
}

// CreateWorkspaceRequest represents the request to create a workspace
type CreateWorkspaceRequest struct {
	Name           string    `json:"name" binding:"required,min=3,max=255"`
	DisplayName    *string   `json:"display_name,omitempty"`
	Description    *string   `json:"description,omitempty"`
	OwnerType      OwnerType `json:"owner_type" binding:"required"`
	OwnerID        int64     `json:"owner_id" binding:"required"`
	TemplateID     int64     `json:"template_id" binding:"required"`
	CPUMillicores  int       `json:"cpu_millicores,omitempty"`
	MemoryMB       int       `json:"memory_mb,omitempty"`
	StorageMB      int       `json:"storage_mb,omitempty"`
	TimeoutSeconds int       `json:"timeout_seconds,omitempty"`
}

// UpdateWorkspaceRequest represents the request to update a workspace
type UpdateWorkspaceRequest struct {
	DisplayName    *string          `json:"display_name,omitempty"`
	Description    *string          `json:"description,omitempty"`
	CPUMillicores  *int             `json:"cpu_millicores,omitempty"`
	MemoryMB       *int             `json:"memory_mb,omitempty"`
	StorageMB      *int             `json:"storage_mb,omitempty"`
	TargetStatus   *WorkspaceStatus `json:"target_status,omitempty"`
	TimeoutSeconds *int             `json:"timeout_seconds,omitempty"`
}

// AttachVolumeRequest represents the request to attach a volume to a workspace
type AttachVolumeRequest struct {
	VolumeID  int64  `json:"volume_id" binding:"required"`
	MountPath string `json:"mount_path" binding:"required"`
}

// WorkspaceActionRequest represents an action to perform on a workspace
type WorkspaceActionRequest struct {
	Action string `json:"action" binding:"required"` // start, stop, restart
}
