package models

import (
	"time"

	"github.com/google/uuid"
)

// Template represents a workspace template
type Template struct {
	ID                    int64     `json:"id" db:"id"`
	UUID                  uuid.UUID `json:"uuid" db:"uuid"`
	Name                  string    `json:"name" db:"name"`
	DisplayName           *string   `json:"display_name,omitempty" db:"display_name"`
	Description           *string   `json:"description,omitempty" db:"description"`
	ImageRef              string    `json:"image_ref" db:"image_ref"`
	TemplateYAML          string    `json:"template_yaml" db:"template_yaml"`
	IconURL               *string   `json:"icon_url,omitempty" db:"icon_url"`
	IsPublic              bool      `json:"is_public" db:"is_public"`
	OwnerType             *string   `json:"owner_type,omitempty" db:"owner_type"` // NULL for system templates
	OwnerID               *int64    `json:"owner_id,omitempty" db:"owner_id"`     // NULL for system templates
	DefaultCPUMillicores  int       `json:"default_cpu_millicores" db:"default_cpu_millicores"`
	DefaultMemoryMB       int       `json:"default_memory_mb" db:"default_memory_mb"`
	DefaultStorageMB      int       `json:"default_storage_mb" db:"default_storage_mb"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// CreateTemplateRequest represents the request to create a template
type CreateTemplateRequest struct {
	Name                 string  `json:"name" binding:"required,min=3,max=255"`
	DisplayName          *string `json:"display_name,omitempty"`
	Description          *string `json:"description,omitempty"`
	ImageRef             string  `json:"image_ref" binding:"required"`
	TemplateYAML         string  `json:"template_yaml" binding:"required"`
	IconURL              *string `json:"icon_url,omitempty"`
	IsPublic             bool    `json:"is_public,omitempty"`
	DefaultCPUMillicores int     `json:"default_cpu_millicores,omitempty"`
	DefaultMemoryMB      int     `json:"default_memory_mb,omitempty"`
	DefaultStorageMB     int     `json:"default_storage_mb,omitempty"`
}

// UpdateTemplateRequest represents the request to update a template
type UpdateTemplateRequest struct {
	DisplayName          *string `json:"display_name,omitempty"`
	Description          *string `json:"description,omitempty"`
	ImageRef             *string `json:"image_ref,omitempty"`
	TemplateYAML         *string `json:"template_yaml,omitempty"`
	IconURL              *string `json:"icon_url,omitempty"`
	IsPublic             *bool   `json:"is_public,omitempty"`
	DefaultCPUMillicores *int    `json:"default_cpu_millicores,omitempty"`
	DefaultMemoryMB      *int    `json:"default_memory_mb,omitempty"`
	DefaultStorageMB     *int    `json:"default_storage_mb,omitempty"`
}
