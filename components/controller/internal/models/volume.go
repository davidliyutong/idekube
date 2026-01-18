package models

import (
	"gorm.io/datatypes"
)

// VolumeAccessMode represents the access mode of a volume
type VolumeAccessMode string

const (
	VolumeAccessModeReadWriteOnce VolumeAccessMode = "ReadWriteOnce"
	VolumeAccessModeReadWriteMany VolumeAccessMode = "ReadWriteMany"
	VolumeAccessModeReadOnlyMany  VolumeAccessMode = "ReadOnlyMany"
)

// VolumeStatus constants (for business logic validation, stored as Base.Status string)
const (
	VolumeStatusPending = "pending"
	VolumeStatusBound   = "bound"
	VolumeStatusFailed  = "failed"
)

// Volume represents a persistent volume
// All Volumes belong to an Organization (removed OwnerType)
// Embeds Base (ID, UUID, CreatedAt, UpdatedAt, DeletedAt, Labels, Status, ExtraInfo)
// Embeds Profile (Identifier as name, DisplayName, IconURL, Description)
// Embeds K8SResources (managed by housekeeper)
type Volume struct {
	Base                             // Embedded Base fields
	Profile      `gorm:"embedded"`   // Embedded Profile fields (Identifier serves as volume name)
	SizeMB       int                 `json:"size_mb" gorm:"not null"`
	StorageClass *string             `json:"storage_class,omitempty" gorm:"type:varchar(255)"` // Immutable after creation
	AccessMode   VolumeAccessMode    `json:"access_mode" gorm:"type:varchar(50);default:'ReadWriteOnce'"` // Immutable after creation
	K8S          K8SResources        `gorm:"embedded"` // K8S resources (managed by housekeeper)
	IsPublic     bool                `json:"is_public" gorm:"default:false"`
	OwnerID      int64               `json:"owner_id" gorm:"not null;index"` // Organization ID
}

// TableName specifies the table name for Volume
func (Volume) TableName() string {
	return "volumes"
}

// GetName returns the volume name (alias for Identifier for backward compatibility)
func (v *Volume) GetName() string {
	return v.Identifier
}

// SetName sets the volume name (alias for Identifier for backward compatibility)
func (v *Volume) SetName(name string) {
	v.Identifier = name
}

// VolumeMount represents a mount point of a volume in a workspace
type VolumeMount struct {
	WorkspaceID   int64  `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name,omitempty"`
	MountPath     string `json:"mount_path"`
}

// ============================================================================
// Request/Response Types
// ============================================================================

// CreateVolumeRequest represents the request to create a volume
type CreateVolumeRequest struct {
	Name         string            `json:"name" binding:"required,min=3,max=255"`
	DisplayName  *string           `json:"display_name,omitempty"`
	Description  *string           `json:"description,omitempty"`
	IconURL      *string           `json:"icon_url,omitempty"`
	OwnerID      int64             `json:"owner_id" binding:"required"` // Organization ID
	SizeMB       int               `json:"size_mb" binding:"required,min=1"`
	StorageClass *string           `json:"storage_class,omitempty"`
	AccessMode   VolumeAccessMode  `json:"access_mode,omitempty"`
	IsPublic     bool              `json:"is_public,omitempty"`
	Labels       datatypes.JSONMap `json:"labels,omitempty"`
}

// UpdateVolumeRequest represents the request to update a volume
type UpdateVolumeRequest struct {
	DisplayName *string           `json:"display_name,omitempty"`
	Description *string           `json:"description,omitempty"`
	IconURL     *string           `json:"icon_url,omitempty"`
	SizeMB      *int              `json:"size_mb,omitempty" binding:"omitempty,min=1"`
	IsPublic    *bool             `json:"is_public,omitempty"`
	Labels      datatypes.JSONMap `json:"labels,omitempty"`
}

// VolumeProfileResponse represents the volume profile sub-resource response
type VolumeProfileResponse struct {
	Identifier  string  `json:"identifier"`
	DisplayName *string `json:"display_name,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateVolumeProfileRequest represents the request to update volume profile
type UpdateVolumeProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// VolumeMountsResponse represents the volume mounts sub-resource response
type VolumeMountsResponse struct {
	Mounts []VolumeMount `json:"mounts"`
}

// UpdateVolumeSizeRequest represents the request to update volume size (expand only)
type UpdateVolumeSizeRequest struct {
	SizeMB int `json:"size_mb" binding:"required,min=1"`
}

// UpdateVolumeIsPublicRequest represents the request to update volume is_public status
type UpdateVolumeIsPublicRequest struct {
	IsPublic bool `json:"is_public"`
}

// ============================================================================
// Additional Sub-resource Response Types
// ============================================================================

// VolumeSizeMBResponse represents the volume size response
type VolumeSizeMBResponse struct {
	SizeMB int `json:"size_mb"`
}

// VolumeStorageClassResponse represents the volume storage class response
type VolumeStorageClassResponse struct {
	StorageClass *string `json:"storage_class"`
}

// VolumeAccessModeResponse represents the volume access mode response
type VolumeAccessModeResponse struct {
	AccessMode VolumeAccessMode `json:"access_mode"`
}

// VolumeOwnerResponse represents the volume owner information response
type VolumeOwnerResponse struct {
	OwnerID   int64         `json:"owner_id"`
	OwnerType string        `json:"owner_type"` // Always "organization"
	Owner     *Organization `json:"owner,omitempty"`
}

// TransferVolumeOwnershipRequest represents the request to transfer volume ownership
type TransferVolumeOwnershipRequest struct {
	NewOwnerID int64   `json:"new_owner_id" binding:"required"` // New organization ID
	Message    *string `json:"message,omitempty"`
}

// VolumePublicResponse represents the volume public status response
type VolumePublicResponse struct {
	IsPublic bool `json:"is_public"`
}

