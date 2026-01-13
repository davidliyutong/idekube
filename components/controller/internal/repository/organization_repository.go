package repository

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"gorm.io/gorm"
)

// OrganizationRepository handles organization data access
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create creates a new organization
func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(ctx context.Context, id int64) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).First(&org, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("organization not found")
	}
	return &org, err
}

// GetByName retrieves an organization by name
func (r *OrganizationRepository) GetByName(ctx context.Context, name string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&org).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("organization not found")
	}
	return &org, err
}

// Update updates an organization
func (r *OrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

// Delete deletes an organization
func (r *OrganizationRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Organization{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("organization not found")
	}
	return nil
}

// ListByOwner lists organizations owned by a user
func (r *OrganizationRepository) ListByOwner(ctx context.Context, ownerID int64) ([]*models.Organization, error) {
	var orgs []*models.Organization
	err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&orgs).Error
	return orgs, err
}

// AddMember adds a member to an organization
func (r *OrganizationRepository) AddMember(ctx context.Context, member *models.OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveMember removes a member from an organization
func (r *OrganizationRepository) RemoveMember(ctx context.Context, organizationID, userID int64) error {
	result := r.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", organizationID, userID).
		Delete(&models.OrganizationMember{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

// UpdateMemberRole updates a member's role
func (r *OrganizationRepository) UpdateMemberRole(ctx context.Context, organizationID, userID int64, role models.OrganizationMemberRole) error {
	result := r.db.WithContext(ctx).Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		Update("role", role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

// ListMembers lists all members of an organization
func (r *OrganizationRepository) ListMembers(ctx context.Context, organizationID int64) ([]*models.OrganizationMember, error) {
	var members []*models.OrganizationMember
	err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("joined_at ASC").Find(&members).Error
	return members, err
}

// GetMember retrieves a specific organization member
func (r *OrganizationRepository) GetMember(ctx context.Context, organizationID, userID int64) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	err := r.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", organizationID, userID).First(&member).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("member not found")
	}
	return &member, err
}

// ListUserOrganizations lists all organizations a user is a member of
func (r *OrganizationRepository) ListUserOrganizations(ctx context.Context, userID int64) ([]*models.Organization, error) {
	var orgs []*models.Organization
	err := r.db.WithContext(ctx).
		Joins("INNER JOIN organization_members ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ?", userID).
		Order("organization_members.joined_at DESC").
		Find(&orgs).Error
	return orgs, err
	return orgs, err
}
