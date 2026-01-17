package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// OrganizationMemberRole represents organization member role enum
type OrganizationMemberRole string

const (
	OrgRoleOwner  OrganizationMemberRole = "owner"
	OrgRoleAdmin  OrganizationMemberRole = "admin"
	OrgRoleMember OrganizationMemberRole = "member"
)

// Organization represents an organization
type Organization struct {
	ID          int64             `json:"id" db:"id"`
	UUID        uuid.UUID         `json:"uuid" db:"uuid"`
	Name        string            `json:"name" db:"name"`
	DisplayName *string           `json:"display_name,omitempty" db:"display_name"`
	Description *string           `json:"description,omitempty" db:"description"`
	AvatarURL   *string           `json:"avatar_url,omitempty" db:"avatar_url"`
	OwnerID     int64             `json:"owner_id" db:"owner_id"`
	Settings    datatypes.JSONMap `json:"settings,omitempty" db:"settings"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	ID             int64                  `json:"id" db:"id"`
	OrganizationID int64                  `json:"organization_id" db:"organization_id"`
	UserID         int64                  `json:"user_id" db:"user_id"`
	Role           OrganizationMemberRole `json:"role" db:"role"`
	JoinedAt       time.Time              `json:"joined_at" db:"joined_at"`
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

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name        string            `json:"name" binding:"required,min=3,max=255"`
	DisplayName *string           `json:"display_name,omitempty"`
	Description *string           `json:"description,omitempty"`
	Settings    datatypes.JSONMap `json:"settings,omitempty"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	DisplayName *string           `json:"display_name,omitempty"`
	Description *string           `json:"description,omitempty"`
	AvatarURL   *string           `json:"avatar_url,omitempty"`
	Settings    datatypes.JSONMap `json:"settings,omitempty"`
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
