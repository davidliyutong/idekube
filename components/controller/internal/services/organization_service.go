package services

import (
	"context"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/pkg/queue"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OrganizationService handles organization business logic
type OrganizationService struct {
	orgRepo        *repository.OrganizationRepository
	userRepo       *repository.UserRepository
	userService    *UserService
	eventPublisher *queue.EventPublisher
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(
	orgRepo *repository.OrganizationRepository,
	userRepo *repository.UserRepository,
	userService *UserService,
	eventPublisher *queue.EventPublisher,
) *OrganizationService {
	return &OrganizationService{
		orgRepo:        orgRepo,
		userRepo:       userRepo,
		userService:    userService,
		eventPublisher: eventPublisher,
	}
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(ctx context.Context, ownerID int64, req *models.CreateOrganizationRequest) (*models.Organization, error) {
	// Verify owner exists
	_, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("owner user not found")
	}

	org := &models.Organization{
		UUID:        uuid.New(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		OwnerID:     ownerID,
		Settings:    req.Settings,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.orgRepo.Create(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Add owner as organization member with owner role
	member := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         ownerID,
		Role:           models.OrgRoleOwner,
		JoinedAt:       time.Now(),
	}

	err = s.orgRepo.AddMember(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("failed to add owner as member: %w", err)
	}

	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationService) GetOrganization(ctx context.Context, id int64) (*models.Organization, error) {
	return s.orgRepo.GetByID(ctx, id)
}

// GetOrganizationWithMembers retrieves organization with member details
func (s *OrganizationService) GetOrganizationWithMembers(ctx context.Context, id int64) (*models.OrganizationWithMembers, error) {
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	members, err := s.orgRepo.ListMembers(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load user details for each member
	membersWithUser := make([]models.OrganizationMemberWithUser, 0, len(members))
	for _, m := range members {
		user, err := s.userRepo.GetByID(ctx, m.UserID)
		if err != nil {
			// Skip if user not found (shouldn't happen with foreign keys)
			continue
		}
		membersWithUser = append(membersWithUser, models.OrganizationMemberWithUser{
			OrganizationMember: *m,
			User:               user,
		})
	}

	return &models.OrganizationWithMembers{
		Organization: *org,
		Members:      membersWithUser,
	}, nil
}

// ListUserOrganizations lists all organizations a user is a member of
func (s *OrganizationService) ListUserOrganizations(ctx context.Context, userID int64) ([]*models.Organization, error) {
	return s.orgRepo.ListUserOrganizations(ctx, userID)
}

// UpdateOrganization updates an organization
func (s *OrganizationService) UpdateOrganization(ctx context.Context, id int64, req *models.UpdateOrganizationRequest) (*models.Organization, error) {
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.DisplayName != nil {
		org.DisplayName = req.DisplayName
	}
	if req.Description != nil {
		org.Description = req.Description
	}
	if req.AvatarURL != nil {
		org.AvatarURL = req.AvatarURL
	}
	if req.Settings != nil {
		org.Settings = req.Settings
	}

	err = s.orgRepo.Update(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return org, nil
}

// DeleteOrganization deletes an organization and publishes event for K8S cleanup
func (s *OrganizationService) DeleteOrganization(ctx context.Context, id int64) error {
	// Get organization info before deletion
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete the organization
	err = s.orgRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Publish delete event to HouseKeeper for K8S resource cleanup
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishOrganizationDelete(ctx, id, org.Name); err != nil {
			// Log error but don't fail the operation
			// HouseKeeper reconciler will handle cleanup eventually
			zap.L().Error("Failed to publish organization delete event",
				zap.Int64("organization_id", id),
				zap.Error(err))
		}
	}

	return nil
}

// AddMember adds a member to an organization
func (s *OrganizationService) AddMember(ctx context.Context, orgID int64, req *models.AddOrganizationMemberRequest) (*models.OrganizationMember, error) {
	// Verify user exists
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if already a member
	existingMember, _ := s.orgRepo.GetMember(ctx, orgID, user.ID)
	if existingMember != nil {
		return nil, fmt.Errorf("user is already a member of this organization")
	}

	// Default role is member
	role := req.Role
	if role == "" {
		role = models.OrgRoleMember
	}

	member := &models.OrganizationMember{
		OrganizationID: orgID,
		UserID:         user.ID,
		Role:           role,
		JoinedAt:       time.Now(),
	}

	err = s.orgRepo.AddMember(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	return member, nil
}

// RemoveMember removes a member from an organization
func (s *OrganizationService) RemoveMember(ctx context.Context, orgID, userID int64) error {
	// Check if user is the owner
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return err
	}

	if org.OwnerID == userID {
		return fmt.Errorf("cannot remove the owner from the organization")
	}

	return s.orgRepo.RemoveMember(ctx, orgID, userID)
}

// UpdateMemberRole updates a member's role
func (s *OrganizationService) UpdateMemberRole(ctx context.Context, orgID, userID int64, req *models.UpdateOrganizationMemberRequest) (*models.OrganizationMember, error) {
	// Check if user is the owner
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if org.OwnerID == userID {
		return nil, fmt.Errorf("cannot change the owner's role")
	}

	err = s.orgRepo.UpdateMemberRole(ctx, orgID, userID, req.Role)
	if err != nil {
		return nil, err
	}

	return s.orgRepo.GetMember(ctx, orgID, userID)
}

// CheckMemberPermission checks if a user has a specific role or higher in an organization
func (s *OrganizationService) CheckMemberPermission(ctx context.Context, orgID, userID int64, requiredRole models.OrganizationMemberRole) (bool, error) {
	member, err := s.orgRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		return false, nil // Not a member
	}

	// Role hierarchy: owner > admin > member
	roleLevel := map[models.OrganizationMemberRole]int{
		models.OrgRoleMember: 1,
		models.OrgRoleAdmin:  2,
		models.OrgRoleOwner:  3,
	}

	return roleLevel[member.Role] >= roleLevel[requiredRole], nil
}

// ListAllOrganizations lists all organizations (admin only)
func (s *OrganizationService) ListAllOrganizations(ctx context.Context, opts *models.ListOptions) ([]*models.Organization, int64, error) {
	return s.orgRepo.ListAll(ctx, opts)
}

// PromoteToAdmin promotes a member to admin role (owner only)
func (s *OrganizationService) PromoteToAdmin(ctx context.Context, orgID, targetUserID, actorUserID int64) error {
	// Verify actor is owner
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return err
	}

	if org.OwnerID != actorUserID {
		return fmt.Errorf("only organization owner can promote members to admin")
	}

	// Check target user is not already owner
	if org.OwnerID == targetUserID {
		return fmt.Errorf("user is already the owner")
	}

	// Update role
	return s.orgRepo.UpdateMemberRole(ctx, orgID, targetUserID, models.OrgRoleAdmin)
}

// DemoteFromAdmin demotes an admin to member role (owner only)
func (s *OrganizationService) DemoteFromAdmin(ctx context.Context, orgID, targetUserID, actorUserID int64) error {
	// Verify actor is owner
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return err
	}

	if org.OwnerID != actorUserID {
		return fmt.Errorf("only organization owner can demote admins")
	}

	// Check target user is not owner
	if org.OwnerID == targetUserID {
		return fmt.Errorf("cannot demote the owner")
	}

	// Update role
	return s.orgRepo.UpdateMemberRole(ctx, orgID, targetUserID, models.OrgRoleMember)
}

// GetUserOrganizationRole gets user's role in an organization
func (s *OrganizationService) GetUserOrganizationRole(ctx context.Context, userID, orgID int64) (models.OrganizationMemberRole, error) {
	return s.orgRepo.GetUserOrganizationRole(ctx, userID, orgID)
}

// SearchUsersForInvite searches users for organization invitation
func (s *OrganizationService) SearchUsersForInvite(ctx context.Context, orgID int64, query string, opts *models.ListOptions) ([]*models.User, int64, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, 0, fmt.Errorf("organization not found: %w", err)
	}

	// Search users using UserService
	if s.userService == nil {
		return nil, 0, fmt.Errorf("user service not available")
	}

	return s.userService.SearchUsers(ctx, query, opts)
}
