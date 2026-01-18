package models

import (
	"time"

	"gorm.io/datatypes"
)

// OrganizationMemberRole represents organization member role enum
type OrganizationMemberRole string

const (
	OrgRoleOwner  OrganizationMemberRole = "owner"
	OrgRoleAdmin  OrganizationMemberRole = "admin"
	OrgRoleMember OrganizationMemberRole = "member"
)

// OrganizationStatus constants (for business logic validation, stored as Base.Status string)
const (
	OrganizationStatusActive   = "active"
	OrganizationStatusInactive = "inactive"
	OrganizationStatusSuspended = "suspended"
)

// Organization represents an organization
// Embeds Base (ID, UUID, CreatedAt, UpdatedAt, DeletedAt, Labels, Status, ExtraInfo)
// Embeds Profile (Identifier as name, DisplayName, IconURL, Description)
type Organization struct {
	Base                       // Embedded Base fields
	Profile  `gorm:"embedded"` // Embedded Profile fields (Identifier serves as organization name)
	OwnerID  int64             `json:"owner_id" gorm:"not null;index"`
	Settings datatypes.JSONMap `json:"settings,omitempty" gorm:"type:jsonb"`
}

// TableName specifies the table name for Organization
func (Organization) TableName() string {
	return "organizations"
}

// GetName returns the organization name (alias for Identifier for backward compatibility)
func (o *Organization) GetName() string {
	return o.Identifier
}

// SetName sets the organization name (alias for Identifier for backward compatibility)
func (o *Organization) SetName(name string) {
	o.Identifier = name
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	ID             int64                  `json:"id" gorm:"primaryKey"`
	OrganizationID int64                  `json:"organization_id" gorm:"not null;index"`
	UserID         int64                  `json:"user_id" gorm:"not null;index"`
	Role           OrganizationMemberRole `json:"role" gorm:"type:varchar(50);default:'member'"`
	JoinedAt       time.Time              `json:"joined_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for OrganizationMember
func (OrganizationMember) TableName() string {
	return "organization_members"
}

// OrganizationWithMembers represents an organization with its members
type OrganizationWithMembers struct {
	Organization
	Members []OrganizationMemberWithUser `json:"members,omitempty"`
}

// OrganizationMemberWithUser represents an organization member with user details
type OrganizationMemberWithUser struct {
	OrganizationMember
	User *User `json:"user,omitempty"`
}

// ============================================================================
// Request/Response Types
// ============================================================================

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name        string            `json:"name" binding:"required,min=3,max=255"`
	DisplayName *string           `json:"display_name,omitempty"`
	Description *string           `json:"description,omitempty"`
	IconURL     *string           `json:"icon_url,omitempty"`
	Settings    datatypes.JSONMap `json:"settings,omitempty"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	DisplayName *string           `json:"display_name,omitempty"`
	Description *string           `json:"description,omitempty"`
	IconURL     *string           `json:"icon_url,omitempty"`
	Settings    datatypes.JSONMap `json:"settings,omitempty"`
	Labels      datatypes.JSONMap `json:"labels,omitempty"`
}

// AddOrganizationMemberRequest represents the request to add a member to an organization
type AddOrganizationMemberRequest struct {
	Username string                 `json:"username" binding:"required"`
	Role     OrganizationMemberRole `json:"role,omitempty"`
}

// UpdateOrganizationMemberRequest represents the request to update a member's role
type UpdateOrganizationMemberRequest struct {
	Role OrganizationMemberRole `json:"role" binding:"required"`
}

// TransferOwnershipRequest represents the request to transfer organization ownership
type TransferOwnershipRequest struct {
	NewOwnerID int64 `json:"new_owner_id" binding:"required"`
}

// OrganizationProfileResponse represents the organization profile sub-resource response
type OrganizationProfileResponse struct {
	Identifier  string  `json:"identifier"`
	DisplayName *string `json:"display_name,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateOrganizationProfileRequest represents the request to update organization profile
type UpdateOrganizationProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// OrganizationQuotaResponse represents the organization quota sub-resource response
type OrganizationQuotaResponse struct {
	CPUMillicores  *int `json:"cpu_millicores,omitempty"`
	MemoryMB       *int `json:"memory_mb,omitempty"`
	StorageMB      *int `json:"storage_mb,omitempty"`
	GPU            *int `json:"gpu,omitempty"`
	Workspaces     *int `json:"workspaces,omitempty"`
	Volumes        *int `json:"volumes,omitempty"`
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
}

// UpdateOrganizationQuotaRequest represents the request to update organization quota
type UpdateOrganizationQuotaRequest struct {
	CPUMillicores  *int `json:"cpu_millicores,omitempty"`
	MemoryMB       *int `json:"memory_mb,omitempty"`
	StorageMB      *int `json:"storage_mb,omitempty"`
	GPU            *int `json:"gpu,omitempty"`
	Workspaces     *int `json:"workspaces,omitempty"`
	Volumes        *int `json:"volumes,omitempty"`
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
}
