package models

import (
	"time"

	"gorm.io/datatypes"
)

// WorkspaceStatus constants (for business logic validation, stored as Base.Status string)
const (
	WorkspaceStatusPending  = "pending"
	WorkspaceStatusStarting = "starting"
	WorkspaceStatusRunning  = "running"
	WorkspaceStatusStopped  = "stopped"
	WorkspaceStatusFailed   = "failed"
)

// Workspace represents a user workspace (IDE instance)
// All Workspaces belong to an Organization (removed OwnerType)
// Embeds Base (ID, UUID, CreatedAt, UpdatedAt, DeletedAt, Labels, Status, ExtraInfo)
// Embeds Profile (Identifier as name, DisplayName, IconURL, Description)
// Embeds K8SResources (managed by housekeeper)
// Embeds QuotaLimits (workspace quota settings)
type Workspace struct {
	Base                                     // Embedded Base fields
	Profile          `gorm:"embedded"`       // Embedded Profile fields (Identifier serves as workspace name)
	TemplateID       int64                   `json:"template_id" gorm:"not null;index"`                // Immutable after creation
	// FIXME: use a string representation of template snapshot
	TemplateSnapshot datatypes.JSONMap       `json:"template_snapshot,omitempty" gorm:"type:jsonb"`    // Immutable - snapshot of template at creation time
	Parameters       datatypes.JSONMap       `json:"parameters,omitempty" gorm:"type:jsonb"`           // Immutable - parameters passed at creation
	K8S              K8SResources            `gorm:"embedded"`                                         // K8S resources (managed by housekeeper)
	Quota            QuotaLimits             `gorm:"embedded"`                                         // Workspace quota settings
	IsPublic         bool                    `json:"is_public" gorm:"default:false"`
	AccessedAt       *time.Time              `json:"accessed_at,omitempty"`                            // Managed by housekeeper
	StartedAt        *time.Time              `json:"started_at,omitempty"`                             // Managed by housekeeper
	OwnerID          int64                   `json:"owner_id" gorm:"not null;index"`                   // Organization ID
	CreatedBy        int64                   `json:"created_by" gorm:"not null"`                       // User ID who created this workspace
	TargetStatus     string                  `json:"target_status" gorm:"type:varchar(50);default:'running'"`
}

// TableName specifies the table name for Workspace
func (Workspace) TableName() string {
	return "workspaces"
}

// GetName returns the workspace name (alias for Identifier for backward compatibility)
func (w *Workspace) GetName() string {
	return w.Identifier
}

// SetName sets the workspace name (alias for Identifier for backward compatibility)
func (w *Workspace) SetName(name string) {
	w.Identifier = name
}

// WorkspaceVolume represents a volume attached to a workspace
type WorkspaceVolume struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	WorkspaceID int64     `json:"workspace_id" gorm:"not null;index"`
	VolumeID    int64     `json:"volume_id" gorm:"not null;index"`
	MountPath   string    `json:"mount_path" gorm:"type:varchar(500);not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for WorkspaceVolume
func (WorkspaceVolume) TableName() string {
	return "workspace_volumes"
}

// WorkspaceWithDetails represents a workspace with additional details
type WorkspaceWithDetails struct {
	Workspace
	Template *Template `json:"template,omitempty"`
	Volumes  []Volume  `json:"volumes,omitempty"`
}

// ============================================================================
// Request/Response Types
// ============================================================================

// CreateWorkspaceRequest represents the request to create a workspace
type CreateWorkspaceRequest struct {
	Name           string            `json:"name" binding:"required,min=3,max=255"`
	DisplayName    *string           `json:"display_name,omitempty"`
	Description    *string           `json:"description,omitempty"`
	IconURL        *string           `json:"icon_url,omitempty"`
	OwnerID        int64             `json:"owner_id" binding:"required"` // Organization ID
	TemplateID     int64             `json:"template_id" binding:"required"`
	Parameters     datatypes.JSONMap `json:"parameters,omitempty"`
	CPUMillicores  *int              `json:"cpu_millicores,omitempty"`
	MemoryMB       *int              `json:"memory_mb,omitempty"`
	StorageMB      *int              `json:"storage_mb,omitempty"`
	GPU            *int              `json:"gpu,omitempty"`
	TimeoutSeconds *int              `json:"timeout_seconds,omitempty"`
	Labels         datatypes.JSONMap `json:"labels,omitempty"`
}

// UpdateWorkspaceRequest represents the request to update a workspace
type UpdateWorkspaceRequest struct {
	DisplayName    *string           `json:"display_name,omitempty"`
	Description    *string           `json:"description,omitempty"`
	IconURL        *string           `json:"icon_url,omitempty"`
	CPUMillicores  *int              `json:"cpu_millicores,omitempty"`
	MemoryMB       *int              `json:"memory_mb,omitempty"`
	StorageMB      *int              `json:"storage_mb,omitempty"`
	GPU            *int              `json:"gpu,omitempty"`
	TimeoutSeconds *int              `json:"timeout_seconds,omitempty"`
	TargetStatus   *string           `json:"target_status,omitempty"`
	IsPublic       *bool             `json:"is_public,omitempty"`
	Labels         datatypes.JSONMap `json:"labels,omitempty"`
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

// WorkspaceProfileResponse represents the workspace profile sub-resource response
type WorkspaceProfileResponse struct {
	Identifier  string  `json:"identifier"`
	DisplayName *string `json:"display_name,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateWorkspaceProfileRequest represents the request to update workspace profile
type UpdateWorkspaceProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// WorkspaceQuotaResponse represents the workspace quota sub-resource response
type WorkspaceQuotaResponse struct {
	CPUMillicores  *int `json:"cpu_millicores,omitempty"`
	MemoryMB       *int `json:"memory_mb,omitempty"`
	StorageMB      *int `json:"storage_mb,omitempty"`
	GPU            *int `json:"gpu,omitempty"`
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
}

// UpdateWorkspaceQuotaRequest represents the request to update workspace quota
type UpdateWorkspaceQuotaRequest struct {
	CPUMillicores  *int `json:"cpu_millicores,omitempty"`
	MemoryMB       *int `json:"memory_mb,omitempty"`
	StorageMB      *int `json:"storage_mb,omitempty"`
	GPU            *int `json:"gpu,omitempty"`
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
}

// UpdateWorkspaceIsPublicRequest represents the request to update workspace is_public status
type UpdateWorkspaceIsPublicRequest struct {
	IsPublic bool `json:"is_public"`
}

// WorkspaceVolumeResponse represents a volume mount in a workspace
type WorkspaceVolumeResponse struct {
	VolumeID  int64  `json:"volume_id"`
	MountPath string `json:"mount_path"`
	Volume    *Volume `json:"volume,omitempty"`
}

// ============================================================================
// Workspace Transfer Types
// ============================================================================

// WorkspaceTransferStatus constants
const (
	WorkspaceTransferStatusPending   = "pending"
	WorkspaceTransferStatusAccepted  = "accepted"
	WorkspaceTransferStatusRejected  = "rejected"
	WorkspaceTransferStatusCancelled = "cancelled"
)

// WorkspaceTransfer represents a workspace ownership transfer request
type WorkspaceTransfer struct {
	ID          int64      `json:"id" gorm:"primaryKey"`
	WorkspaceID int64      `json:"workspace_id" gorm:"not null;index"`
	FromUserID  int64      `json:"from_user_id" gorm:"not null;index"`
	ToUsername  string     `json:"to_username" gorm:"type:varchar(255);not null"`
	ToUserID    *int64     `json:"to_user_id,omitempty" gorm:"index"`
	Status      string     `json:"status" gorm:"type:varchar(50);default:'pending'"`
	Message     *string    `json:"message,omitempty" gorm:"type:text"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// TableName specifies the table name for WorkspaceTransfer
func (WorkspaceTransfer) TableName() string {
	return "workspace_transfers"
}

// WorkspaceTransferWithDetails represents a transfer with workspace and user details
type WorkspaceTransferWithDetails struct {
	WorkspaceTransfer
	Workspace *Workspace `json:"workspace,omitempty"`
	FromUser  *User      `json:"from_user,omitempty"`
	ToUser    *User      `json:"to_user,omitempty"`
}

// CreateWorkspaceTransferRequest represents the request to create a transfer
type CreateWorkspaceTransferRequest struct {
	ToUsername string  `json:"to_username" binding:"required"`
	Message    *string `json:"message,omitempty"`
}

// RespondWorkspaceTransferRequest represents the request to respond to a transfer
type RespondWorkspaceTransferRequest struct {
	Accept  bool    `json:"accept"`
	Message *string `json:"message,omitempty"`
}

// ============================================================================
// Additional Sub-resource Response Types
// ============================================================================

// WorkspacePublicResponse represents the workspace public status response
type WorkspacePublicResponse struct {
	IsPublic bool `json:"is_public"`
}

// WorkspaceOwnerResponse represents the workspace owner information response
type WorkspaceOwnerResponse struct {
	OwnerID   int64         `json:"owner_id"`
	OwnerType string        `json:"owner_type"` // Always "organization"
	Owner     *Organization `json:"owner,omitempty"`
}

// TransferWorkspaceOwnershipRequest represents the request to transfer workspace ownership
type TransferWorkspaceOwnershipRequest struct {
	NewOwnerID int64  `json:"new_owner_id" binding:"required"` // New organization ID
	Message    *string `json:"message,omitempty"`
}

// WorkspaceStateResponse represents the workspace state response
type WorkspaceStateResponse struct {
	CurrentStatus string `json:"current_status"` // Current actual status
	TargetStatus  string `json:"target_status"`  // Desired target status
}

// UpdateWorkspaceStateRequest represents the request to update workspace state
type UpdateWorkspaceStateRequest struct {
	TargetStatus string `json:"target_status" binding:"required,oneof=running stopped"` // Only running or stopped
}

// UpdateVolumeMountsRequest represents the request to update volume mounts
type UpdateVolumeMountsRequest struct {
	Volumes []AttachVolumeRequest `json:"volumes" binding:"required"`
}

