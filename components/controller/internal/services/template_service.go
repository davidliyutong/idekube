package services

import (
	"context"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// TemplateService handles template business logic
type TemplateService struct {
	templateRepo *repository.TemplateRepository
	orgRepo      *repository.OrganizationRepository
}

// NewTemplateService creates a new template service
func NewTemplateService(
	templateRepo *repository.TemplateRepository,
	orgRepo *repository.OrganizationRepository,
) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		orgRepo:      orgRepo,
	}
}

// CreateTemplate creates a new template
func (s *TemplateService) CreateTemplate(ctx context.Context, userID int64, userRole models.UserRole, req *models.CreateTemplateRequest) (*models.Template, error) {
	// Validate YAML
	if err := s.validateTemplateYAML(req.TemplateYAML); err != nil {
		return nil, fmt.Errorf("invalid template YAML: %w", err)
	}

	// Set default resources if not provided
	if req.DefaultCPUMillicores == 0 {
		req.DefaultCPUMillicores = 1000
	}
	if req.DefaultMemoryMB == 0 {
		req.DefaultMemoryMB = 2048
	}
	if req.DefaultStorageMB == 0 {
		req.DefaultStorageMB = 10240
	}

	template := &models.Template{
		UUID:                 uuid.New(),
		Name:                 req.Name,
		DisplayName:          req.DisplayName,
		Description:          req.Description,
		ImageRef:             req.ImageRef,
		TemplateYAML:         req.TemplateYAML,
		IconURL:              req.IconURL,
		IsPublic:             req.IsPublic,
		DefaultCPUMillicores: req.DefaultCPUMillicores,
		DefaultMemoryMB:      req.DefaultMemoryMB,
		DefaultStorageMB:     req.DefaultStorageMB,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// System templates (super_admin only)
	if userRole == models.UserRoleSuperAdmin && req.IsPublic {
		// System template: owner_type and owner_id are NULL
		template.OwnerType = nil
		template.OwnerID = nil
	} else {
		// User/Org template
		ownerType := string(models.OwnerTypeUser)
		template.OwnerType = &ownerType
		template.OwnerID = &userID
	}

	err := s.templateRepo.Create(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

// GetTemplate retrieves a template by ID
func (s *TemplateService) GetTemplate(ctx context.Context, id int64) (*models.Template, error) {
	return s.templateRepo.GetByID(ctx, id)
}

// ListAccessibleTemplates lists all templates accessible to a user
func (s *TemplateService) ListAccessibleTemplates(ctx context.Context, userID int64) ([]*models.Template, error) {
	// Get user's organizations
	orgs, err := s.orgRepo.ListUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}

	orgIDs := make([]int64, 0, len(orgs))
	for _, org := range orgs {
		orgIDs = append(orgIDs, org.ID)
	}

	return s.templateRepo.ListAccessible(ctx, userID, orgIDs)
}

// ListPublicTemplates lists all public templates
func (s *TemplateService) ListPublicTemplates(ctx context.Context) ([]*models.Template, error) {
	return s.templateRepo.ListPublic(ctx)
}

// UpdateTemplate updates a template
func (s *TemplateService) UpdateTemplate(ctx context.Context, id int64, req *models.UpdateTemplateRequest) (*models.Template, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.DisplayName != nil {
		template.DisplayName = req.DisplayName
	}
	if req.Description != nil {
		template.Description = req.Description
	}
	if req.ImageRef != nil {
		template.ImageRef = *req.ImageRef
	}
	if req.TemplateYAML != nil {
		// Validate new YAML
		if err := s.validateTemplateYAML(*req.TemplateYAML); err != nil {
			return nil, fmt.Errorf("invalid template YAML: %w", err)
		}
		template.TemplateYAML = *req.TemplateYAML
	}
	if req.IconURL != nil {
		template.IconURL = req.IconURL
	}
	if req.IsPublic != nil {
		template.IsPublic = *req.IsPublic
	}
	if req.DefaultCPUMillicores != nil {
		template.DefaultCPUMillicores = *req.DefaultCPUMillicores
	}
	if req.DefaultMemoryMB != nil {
		template.DefaultMemoryMB = *req.DefaultMemoryMB
	}
	if req.DefaultStorageMB != nil {
		template.DefaultStorageMB = *req.DefaultStorageMB
	}

	err = s.templateRepo.Update(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return template, nil
}

// DeleteTemplate deletes a template
func (s *TemplateService) DeleteTemplate(ctx context.Context, id int64) error {
	return s.templateRepo.Delete(ctx, id)
}

// CheckTemplateAccess checks if a user can access a template
func (s *TemplateService) CheckTemplateAccess(ctx context.Context, templateID, userID int64) (bool, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return false, err
	}

	// Public templates are accessible to all
	if template.IsPublic {
		return true, nil
	}

	// System templates (owner_type = NULL) are public
	if template.OwnerType == nil {
		return true, nil
	}

	// Check if user owns the template
	if *template.OwnerType == string(models.OwnerTypeUser) && *template.OwnerID == userID {
		return true, nil
	}

	// Check if user is member of the organization that owns the template
	if *template.OwnerType == string(models.OwnerTypeOrganization) {
		member, _ := s.orgRepo.GetMember(ctx, *template.OwnerID, userID)
		return member != nil, nil
	}

	return false, nil
}

// validateTemplateYAML validates Kubernetes YAML manifest
func (s *TemplateService) validateTemplateYAML(yamlContent string) error {
	// Basic YAML syntax validation
	var data interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &data); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	// TODO: Add more sophisticated validation
	// - Check for required Kubernetes fields (apiVersion, kind, metadata, spec)
	// - Validate against Kubernetes schema
	// - Check for valid placeholder syntax {{.VariableName}}

	return nil
}

// ListTemplates lists templates based on user role and permissions
func (s *TemplateService) ListTemplates(ctx context.Context, userID int64, userRole models.UserRole, orgIDs []int64, opts *models.ListOptions) ([]*models.Template, int64, error) {
	switch userRole {
	case models.UserRoleSuperAdmin, models.UserRoleAdmin:
		// Admins can see all templates
		return s.templateRepo.ListAll(ctx, opts)

	case models.UserRolePowerUser, models.UserRoleUser:
		// Regular users see public templates, their own templates, and org templates
		return s.templateRepo.ListAccessibleByUser(ctx, userID, orgIDs, opts)

	default:
		return nil, 0, fmt.Errorf("unknown user role: %s", userRole)
	}
}

// ListAllTemplates lists all templates (admin only)
func (s *TemplateService) ListAllTemplates(ctx context.Context, opts *models.ListOptions) ([]*models.Template, int64, error) {
	return s.templateRepo.ListAll(ctx, opts)
}
