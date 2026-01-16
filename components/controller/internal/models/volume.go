package models

import (
	"time"

	"github.com/google/uuid"
)

// VolumeAccessMode represents the access mode of a volume
type VolumeAccessMode string

const (
	VolumeAccessModeReadWriteOnce VolumeAccessMode = "ReadWriteOnce"
	VolumeAccessModeReadWriteMany VolumeAccessMode = "ReadWriteMany"
	VolumeAccessModeReadOnlyMany  VolumeAccessMode = "ReadOnlyMany"
)

// VolumeStatus represents the status of a volume
type VolumeStatus string

const (
	VolumeStatusPending VolumeStatus = "pending"
	VolumeStatusBound   VolumeStatus = "bound"
	VolumeStatusFailed  VolumeStatus = "failed"
)

// Volume represents a persistent volume
type Volume struct {
	ID           int64            `json:"id" db:"id"`
	UUID         uuid.UUID        `json:"uuid" db:"uuid"`
	Name         string           `json:"name" db:"name"`
	DisplayName  *string          `json:"display_name,omitempty" db:"display_name"`
	Description  *string          `json:"description,omitempty" db:"description"`
	OwnerType    OwnerType        `json:"owner_type" db:"owner_type"`
	OwnerID      int64            `json:"owner_id" db:"owner_id"`
	SizeMB       int              `json:"size_mb" db:"size_mb"`
	StorageClass *string          `json:"storage_class,omitempty" db:"storage_class"`
	AccessMode   VolumeAccessMode `json:"access_mode" db:"access_mode"`
	PVCName      *string          `json:"pvc_name,omitempty" db:"pvc_name"`
	Status       VolumeStatus     `json:"status" db:"status"`
	Labels       ResourceLabels   `json:"labels,omitempty" db:"labels"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time       `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateVolumeRequest represents the request to create a volume
type CreateVolumeRequest struct {
	Name         string           `json:"name" binding:"required,min=3,max=255"`
	DisplayName  *string          `json:"display_name,omitempty"`
	Description  *string          `json:"description,omitempty"`
	OwnerType    OwnerType        `json:"owner_type" binding:"required"`
	OwnerID      int64            `json:"owner_id" binding:"required"`
	SizeMB       int              `json:"size_mb" binding:"required,min=1"`
	StorageClass *string          `json:"storage_class,omitempty"`
	AccessMode   VolumeAccessMode `json:"access_mode,omitempty"`
}

// UpdateVolumeRequest represents the request to update a volume
type UpdateVolumeRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	SizeMB      *int    `json:"size_mb,omitempty" binding:"omitempty,min=1"`
}

// TODO: support for more types of volumes
