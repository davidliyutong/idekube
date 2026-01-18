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

	now := time.Now()
	template := &models.Template{
		Base: models.Base{
			UUID:      uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Status:    models.TemplateStatusActive,
			Labels:    req.Labels,
		},
		Profile: models.Profile{
			Identifier:  req.Name,
			DisplayName: req.DisplayName,
			Description: req.Description,
			IconURL:     req.IconURL,
		},
		ImageRef:     req.ImageRef,
		TemplateYAML: req.TemplateYAML,
		IsPublic:     req.IsPublic,
		DefaultQuota: models.QuotaLimits{
			CPUMillicores:  req.CPUMillicores,
			MemoryMB:       req.MemoryMB,
			StorageMB:      req.StorageMB,
			GPU:            req.GPU,
			TimeoutSeconds: req.TimeoutSeconds,
		},
	}

	// Set default resources if not provided
	if template.DefaultQuota.CPUMillicores == nil {
		defaultCPU := 1000
		template.DefaultQuota.CPUMillicores = &defaultCPU
	}
	if template.DefaultQuota.MemoryMB == nil {
		defaultMem := 2048
		template.DefaultQuota.MemoryMB = &defaultMem
	}
	if template.DefaultQuota.StorageMB == nil {
		defaultStorage := 10240
		template.DefaultQuota.StorageMB = &defaultStorage
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
	if req.IconURL != nil {
		template.IconURL = req.IconURL
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
	if req.IsPublic != nil {
		template.IsPublic = *req.IsPublic
	}
	if req.CPUMillicores != nil {
		template.DefaultQuota.CPUMillicores = req.CPUMillicores
	}
	if req.MemoryMB != nil {
		template.DefaultQuota.MemoryMB = req.MemoryMB
	}
	if req.StorageMB != nil {
		template.DefaultQuota.StorageMB = req.StorageMB
	}
	if req.GPU != nil {
		template.DefaultQuota.GPU = req.GPU
	}
	if req.TimeoutSeconds != nil {
		template.DefaultQuota.TimeoutSeconds = req.TimeoutSeconds
	}
	if req.Labels != nil {
		template.Labels = req.Labels
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

	// Active templates are accessible to all authenticated users
	if template.Status == models.TemplateStatusActive {
		return true, nil
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
		// Regular users see public templates and active templates
		return s.templateRepo.ListAccessibleByUser(ctx, userID, orgIDs, opts)

	default:
		return nil, 0, fmt.Errorf("unknown user role: %s", userRole)
	}
}

// ListAllTemplates lists all templates (admin only)
func (s *TemplateService) ListAllTemplates(ctx context.Context, opts *models.ListOptions) ([]*models.Template, int64, error) {
	return s.templateRepo.ListAll(ctx, opts)
}

// ============================================================================
// Sub-resource APIs
// ============================================================================

// GetTemplateProfile returns the template's profile sub-resource
func (s *TemplateService) GetTemplateProfile(ctx context.Context, templateID int64) (*models.TemplateProfileResponse, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	return &models.TemplateProfileResponse{
		Identifier:  template.Identifier,
		DisplayName: template.DisplayName,
		IconURL:     template.IconURL,
		Description: template.Description,
	}, nil
}

// UpdateTemplateProfile updates the template's profile sub-resource
func (s *TemplateService) UpdateTemplateProfile(ctx context.Context, templateID int64, req *models.UpdateTemplateProfileRequest) (*models.TemplateProfileResponse, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// Update profile fields
	if req.DisplayName != nil {
		template.DisplayName = req.DisplayName
	}
	if req.IconURL != nil {
		template.IconURL = req.IconURL
	}
	if req.Description != nil {
		template.Description = req.Description
	}

	err = s.templateRepo.Update(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &models.TemplateProfileResponse{
		Identifier:  template.Identifier,
		DisplayName: template.DisplayName,
		IconURL:     template.IconURL,
		Description: template.Description,
	}, nil
}

// GetTemplateQuota returns the template's default quota sub-resource
func (s *TemplateService) GetTemplateQuota(ctx context.Context, templateID int64) (*models.TemplateQuotaResponse, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	return &models.TemplateQuotaResponse{
		CPUMillicores:  template.DefaultQuota.CPUMillicores,
		MemoryMB:       template.DefaultQuota.MemoryMB,
		StorageMB:      template.DefaultQuota.StorageMB,
		GPU:            template.DefaultQuota.GPU,
		TimeoutSeconds: template.DefaultQuota.TimeoutSeconds,
	}, nil
}

// UpdateTemplateQuota updates the template's default quota sub-resource
func (s *TemplateService) UpdateTemplateQuota(ctx context.Context, templateID int64, req *models.UpdateTemplateQuotaRequest) (*models.TemplateQuotaResponse, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// Update quota fields
	if req.CPUMillicores != nil {
		template.DefaultQuota.CPUMillicores = req.CPUMillicores
	}
	if req.MemoryMB != nil {
		template.DefaultQuota.MemoryMB = req.MemoryMB
	}
	if req.StorageMB != nil {
		template.DefaultQuota.StorageMB = req.StorageMB
	}
	if req.GPU != nil {
		template.DefaultQuota.GPU = req.GPU
	}
	if req.TimeoutSeconds != nil {
		template.DefaultQuota.TimeoutSeconds = req.TimeoutSeconds
	}

	err = s.templateRepo.Update(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("failed to update quota: %w", err)
	}

	return &models.TemplateQuotaResponse{
		CPUMillicores:  template.DefaultQuota.CPUMillicores,
		MemoryMB:       template.DefaultQuota.MemoryMB,
		StorageMB:      template.DefaultQuota.StorageMB,
		GPU:            template.DefaultQuota.GPU,
		TimeoutSeconds: template.DefaultQuota.TimeoutSeconds,
	}, nil
}

// UpdateTemplateIsPublic updates the template's is_public status
func (s *TemplateService) UpdateTemplateIsPublic(ctx context.Context, templateID int64, req *models.UpdateTemplateIsPublicRequest) error {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return err
	}

	template.IsPublic = req.IsPublic

	return s.templateRepo.Update(ctx, template)
}

// UpdateTemplateYAML updates the template's YAML
func (s *TemplateService) UpdateTemplateYAML(ctx context.Context, templateID int64, req *models.UpdateTemplateYAMLRequest) error {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return err
	}

	// Validate YAML
	if err := s.validateTemplateYAML(req.TemplateYAML); err != nil {
		return fmt.Errorf("invalid template YAML: %w", err)
	}

	template.TemplateYAML = req.TemplateYAML

	return s.templateRepo.Update(ctx, template)
}

// UpdateTemplateImageRef updates the template's image reference
func (s *TemplateService) UpdateTemplateImageRef(ctx context.Context, templateID int64, req *models.UpdateTemplateImageRefRequest) error {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return err
	}

	template.ImageRef = req.ImageRef

	return s.templateRepo.Update(ctx, template)
}

// GetTemplateImageRef gets the template's image reference
func (s *TemplateService) GetTemplateImageRef(ctx context.Context, templateID int64) (string, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return "", err
	}

	return template.ImageRef, nil
}

// GetTemplateYAML gets the template's YAML content
func (s *TemplateService) GetTemplateYAML(ctx context.Context, templateID int64) (string, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return "", err
	}

	return template.TemplateYAML, nil
}

// GetTemplateIsPublic gets the template's public status
func (s *TemplateService) GetTemplateIsPublic(ctx context.Context, templateID int64) (bool, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return false, err
	}

	return template.IsPublic, nil
}
