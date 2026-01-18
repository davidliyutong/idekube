package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ============================================================================
// Base Embedded Structures
// ============================================================================

// Base is the base struct for all models
// All derived models use this Status field (string type)
// Each model defines its own status constants (e.g., UserStatusActive = "active") for business validation
type Base struct {
	ID        int64             `json:"id" gorm:"primaryKey"`
	UUID      uuid.UUID         `json:"uuid" gorm:"type:uuid;default:uuid_generate_v4();uniqueIndex"`
	CreatedAt time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty" gorm:"index"`
	Labels    datatypes.JSONMap `json:"labels,omitempty" gorm:"type:jsonb"`
	Status    string            `json:"status" gorm:"type:varchar(50);default:'active'"`
	ExtraInfo datatypes.JSONMap `json:"extra_info,omitempty" gorm:"type:jsonb"`
}

// Profile represents resource profile/descriptive information
type Profile struct {
	Identifier  string  `json:"identifier" gorm:"column:identifier;type:varchar(255);uniqueIndex;not null"`
	DisplayName *string `json:"display_name,omitempty" gorm:"type:varchar(255)"`
	IconURL     *string `json:"icon_url,omitempty" gorm:"type:text"`
	Description *string `json:"description,omitempty" gorm:"type:text"`
}

// Security represents security-related fields for users
type Security struct {
	PasswordHash   string     `json:"-" gorm:"type:varchar(255);not null"`
	MFAEnabled     bool       `json:"mfa_enabled" gorm:"default:false"`
	MFASecret      *string    `json:"-" gorm:"type:text"`
	MFABackupCodes []string   `json:"-" gorm:"type:text[]"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
}

// K8SResources represents K8S resource information (managed by housekeeper)
type K8SResources struct {
	Namespace      *string `json:"k8s_namespace,omitempty" gorm:"column:k8s_namespace;type:varchar(255)"`
	DeploymentName *string `json:"k8s_deployment_name,omitempty" gorm:"column:k8s_deployment_name;type:varchar(255)"`
	ServiceName    *string `json:"k8s_service_name,omitempty" gorm:"column:k8s_service_name;type:varchar(255)"`
	IngressName    *string `json:"k8s_ingress_name,omitempty" gorm:"column:k8s_ingress_name;type:varchar(255)"`
	PVCName        *string `json:"k8s_pvc_name,omitempty" gorm:"column:k8s_pvc_name;type:varchar(255)"`
}

// QuotaLimits represents quota/resource limits
type QuotaLimits struct {
	CPUMillicores  *int `json:"cpu_millicores,omitempty" gorm:"column:cpu_millicores;default:8000"`
	MemoryMB       *int `json:"memory_mb,omitempty" gorm:"column:memory_mb;default:16384"`
	StorageMB      *int `json:"storage_mb,omitempty" gorm:"column:storage_mb;default:51200"`
	GPU            *int `json:"gpu,omitempty" gorm:"column:gpu_count;default:0"`
	Workspaces     *int `json:"workspaces,omitempty" gorm:"column:max_workspaces;default:10"`
	Volumes        *int `json:"volumes,omitempty" gorm:"column:max_volumes;default:20"`
	TimeoutSeconds *int `json:"timeout_seconds,omitempty" gorm:"column:timeout_seconds;default:0"`
}

// ============================================================================
// API Response Types
// ============================================================================

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents a simple error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize int `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// ListOptions represents common list query options
type ListOptions struct {
	PaginationRequest
	SortBy    string `form:"sort_by,default=created_at"`
	SortOrder string `form:"sort_order,default=desc" binding:"omitempty,oneof=asc desc"`
	Search    string `form:"search"`
}

// ResourceLabels represents labels attached to resources for organization and RBAC
// Deprecated: Use Base.Labels instead
type ResourceLabels map[string]string
