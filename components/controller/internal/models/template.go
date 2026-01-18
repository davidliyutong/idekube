package models

import (
	"gorm.io/datatypes"
)

// TemplateStatus constants (for business logic validation, stored as Base.Status string)
const (
	TemplateStatusActive   = "active"
	TemplateStatusInactive = "inactive"
	TemplateStatusArchived = "archived"
)

// Template represents a workspace template
// Template is a system-level concept (removed OwnerType, OwnerID, Visibility)
// Embeds Base (ID, UUID, CreatedAt, UpdatedAt, DeletedAt, Labels, Status, ExtraInfo)
// Embeds Profile (Identifier as name, DisplayName, IconURL, Description)
type Template struct {
	Base                                   // Embedded Base fields
	Profile      `gorm:"embedded"`         // Embedded Profile fields (Identifier serves as template name)
	ImageRef     string                    `json:"image_ref" gorm:"type:varchar(500);not null"`
	TemplateYAML string                    `json:"template_yaml" gorm:"type:text;not null"`
	IsPublic     bool                      `json:"is_public" gorm:"default:false"`
	DefaultQuota QuotaLimits               `gorm:"embedded;embeddedPrefix:default_"` // Default quota limits for workspaces created from this template
}

// TableName specifies the table name for Template
func (Template) TableName() string {
	return "templates"
}

// GetName returns the template name (alias for Identifier for backward compatibility)
func (t *Template) GetName() string {
	return t.Identifier
}

// SetName sets the template name (alias for Identifier for backward compatibility)
func (t *Template) SetName(name string) {
	t.Identifier = name
}

// ============================================================================
// Request/Response Types
// ============================================================================

// CreateTemplateRequest represents the request to create a template
type CreateTemplateRequest struct {
	Name           string            `json:"name" binding:"required,min=3,max=255"`
	DisplayName    *string           `json:"display_name,omitempty"`
	Description    *string           `json:"description,omitempty"`
	IconURL        *string           `json:"icon_url,omitempty"`
	ImageRef       string            `json:"image_ref" binding:"required"`
	TemplateYAML   string            `json:"template_yaml" binding:"required"`
	IsPublic       bool              `json:"is_public,omitempty"`
	CPUMillicores  *int              `json:"cpu_millicores,omitempty"`
	MemoryMB       *int              `json:"memory_mb,omitempty"`
	StorageMB      *int              `json:"storage_mb,omitempty"`
	GPU            *int              `json:"gpu,omitempty"`
	TimeoutSeconds *int              `json:"timeout_seconds,omitempty"`
	Labels         datatypes.JSONMap `json:"labels,omitempty"`
}

// UpdateTemplateRequest represents the request to update a template
type UpdateTemplateRequest struct {
	DisplayName    *string           `json:"display_name,omitempty"`
	Description    *string           `json:"description,omitempty"`
	IconURL        *string           `json:"icon_url,omitempty"`
	ImageRef       *string           `json:"image_ref,omitempty"`
	TemplateYAML   *string           `json:"template_yaml,omitempty"`
	IsPublic       *bool             `json:"is_public,omitempty"`
	CPUMillicores  *int              `json:"cpu_millicores,omitempty"`
	MemoryMB       *int              `json:"memory_mb,omitempty"`
	StorageMB      *int              `json:"storage_mb,omitempty"`
	GPU            *int              `json:"gpu,omitempty"`
	TimeoutSeconds *int              `json:"timeout_seconds,omitempty"`
	Labels         datatypes.JSONMap `json:"labels,omitempty"`
}

// TemplateProfileResponse represents the template profile sub-resource response
type TemplateProfileResponse struct {
	Identifier  string  `json:"identifier"`
	DisplayName *string `json:"display_name,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateTemplateProfileRequest represents the request to update template profile
type UpdateTemplateProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// TemplateQuotaResponse represents the template default quota sub-resource response
type TemplateQuotaResponse struct {
	CPUMillicores  *int `json:"cpu_millicores,omitempty"`
	MemoryMB       *int `json:"memory_mb,omitempty"`
	StorageMB      *int `json:"storage_mb,omitempty"`
	GPU            *int `json:"gpu,omitempty"`
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
}

// UpdateTemplateQuotaRequest represents the request to update template default quota
type UpdateTemplateQuotaRequest struct {
	CPUMillicores  *int `json:"cpu_millicores,omitempty"`
	MemoryMB       *int `json:"memory_mb,omitempty"`
	StorageMB      *int `json:"storage_mb,omitempty"`
	GPU            *int `json:"gpu,omitempty"`
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
}

// UpdateTemplateIsPublicRequest represents the request to update template is_public status
type UpdateTemplateIsPublicRequest struct {
	IsPublic bool `json:"is_public"`
}

// UpdateTemplateYAMLRequest represents the request to update template YAML
type UpdateTemplateYAMLRequest struct {
	TemplateYAML string `json:"template_yaml" binding:"required"`
}

// UpdateTemplateImageRefRequest represents the request to update template image reference
type UpdateTemplateImageRefRequest struct {
	ImageRef string `json:"image_ref" binding:"required"`
}
