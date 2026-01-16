package opa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
	"gorm.io/gorm"

	"github.com/davidliyutong/idekube-rbac/pkg/logger"
)

// Enforcer wraps OPA rego evaluation for RBAC decisions.
type Enforcer struct {
	query        rego.PreparedEvalQuery
	db           *gorm.DB
	log          *logger.Logger
	policyModule string
}

// NewEnforcer creates an OPA enforcer from a rego policy file and optional data.
func NewEnforcer(db *gorm.DB, policyPath, dataPath string, log *logger.Logger) (*Enforcer, error) {
	if db == nil {
		return nil, errors.New("gorm db is required")
	}
	if strings.TrimSpace(policyPath) == "" {
		return nil, errors.New("OPA policy path is required")
	}

	// Check if policy file exists
	if _, err := os.Stat(policyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("policy file not found: %s", policyPath)
	}

	cleanedPolicy := filepath.Clean(policyPath)
	policyContent, err := os.ReadFile(cleanedPolicy)
	if err != nil {
		return nil, fmt.Errorf("read OPA policy: %w", err)
	}

	// Load data if provided
	var data map[string]interface{}
	if dataPath != "" {
		cleanedData := filepath.Clean(dataPath)
		if _, err := os.Stat(cleanedData); err == nil {
			dataContent, err := os.ReadFile(cleanedData)
			if err != nil {
				return nil, fmt.Errorf("read OPA data: %w", err)
			}
			if err := json.Unmarshal(dataContent, &data); err != nil {
				return nil, fmt.Errorf("parse OPA data: %w", err)
			}
		} else if log != nil {
			log.Warnf("data file not found: %s, will initialize with empty data", dataPath)
		}
	}

	// Create rego query
	ctx := context.Background()
	regoOpts := []func(*rego.Rego){
		rego.Query("data.idekube.rbac.allow"),
		rego.Module("policy.rego", string(policyContent)),
	}

	// Note: Static data can be loaded here if needed, but we'll primarily
	// use dynamic data from the database during evaluation
	if data != nil {
		log.Infof("Loaded initial OPA data with %d role_bindings and %d policies",
			len(data["role_bindings"].(map[string]interface{})),
			len(data["policies"].([]interface{})))
	}

	r := rego.New(regoOpts...)
	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare OPA query: %w", err)
	}

	return &Enforcer{
		query:        query,
		db:           db,
		log:          log,
		policyModule: string(policyContent),
	}, nil
}

// Enforce evaluates whether a subject can perform an action on an object.
func (e *Enforcer) Enforce(subject, object, action string) (bool, error) {
	// Load dynamic data from database
	dynamicData, err := e.loadDynamicData()
	if err != nil {
		return false, fmt.Errorf("load dynamic data: %w", err)
	}

	input := map[string]interface{}{
		"subject": subject,
		"object":  object,
		"action":  action,
	}

	ctx := context.Background()
	
	// Prepare evaluation with dynamic data
	r := rego.New(
		rego.Query("data.idekube.rbac.allow"),
		rego.Module("policy.rego", e.policyModule),
		rego.Input(input),
		rego.Store(inmem.NewFromObject(dynamicData)),
	)

	results, err := r.Eval(ctx)
	if err != nil {
		return false, fmt.Errorf("evaluate OPA policy: %w", err)
	}

	if len(results) == 0 || len(results[0].Expressions) == 0 {
		return false, nil
	}

	// Check if allow is true
	allowed, ok := results[0].Expressions[0].Value.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected result type from OPA: %T", results[0].Expressions[0].Value)
	}

	return allowed, nil
}

// loadDynamicData loads role assignments and policies from the database.
func (e *Enforcer) loadDynamicData() (map[string]interface{}, error) {
	var roleBindings []RoleBinding
	if err := e.db.Find(&roleBindings).Error; err != nil {
		return nil, fmt.Errorf("query role bindings: %w", err)
	}

	var policies []Policy
	if err := e.db.Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("query policies: %w", err)
	}

	// Convert to OPA data format
	roles := make(map[string][]string)
	for _, rb := range roleBindings {
		roles[rb.Subject] = append(roles[rb.Subject], rb.Role)
	}

	policiesMap := make([]map[string]string, 0, len(policies))
	for _, p := range policies {
		policiesMap = append(policiesMap, map[string]string{
			"subject": p.Subject,
			"object":  p.Object,
			"action":  p.Action,
		})
	}

	return map[string]interface{}{
		"role_bindings": roles,
		"policies":      policiesMap,
	}, nil
}

// AddPolicy adds a new policy rule to the database.
func (e *Enforcer) AddPolicy(subject, object, action string) error {
	policy := Policy{
		Subject: subject,
		Object:  object,
		Action:  action,
	}
	return e.db.Create(&policy).Error
}

// RemovePolicy removes a policy rule from the database.
func (e *Enforcer) RemovePolicy(subject, object, action string) error {
	return e.db.Where("subject = ? AND object = ? AND action = ?", subject, object, action).
		Delete(&Policy{}).Error
}

// AddRoleForUser assigns a role to a user.
func (e *Enforcer) AddRoleForUser(user, role string) error {
	rb := RoleBinding{
		Subject: user,
		Role:    role,
	}
	return e.db.Create(&rb).Error
}

// RemoveRoleForUser removes a role from a user.
func (e *Enforcer) RemoveRoleForUser(user, role string) error {
	return e.db.Where("subject = ? AND role = ?", user, role).
		Delete(&RoleBinding{}).Error
}

// GetRolesForUser returns all roles assigned to a user.
func (e *Enforcer) GetRolesForUser(user string) ([]string, error) {
	var roleBindings []RoleBinding
	if err := e.db.Where("subject = ?", user).Find(&roleBindings).Error; err != nil {
		return nil, err
	}

	roles := make([]string, 0, len(roleBindings))
	for _, rb := range roleBindings {
		roles = append(roles, rb.Role)
	}
	return roles, nil
}

// GetAllPolicies returns all policy rules.
func (e *Enforcer) GetAllPolicies() ([][]string, error) {
	var policies []Policy
	if err := e.db.Find(&policies).Error; err != nil {
		return nil, err
	}

	result := make([][]string, 0, len(policies))
	for _, p := range policies {
		result = append(result, []string{p.Subject, p.Object, p.Action})
	}
	return result, nil
}

// RoleBinding represents a user-to-role assignment in the database.
type RoleBinding struct {
	ID      uint   `gorm:"primaryKey"`
	Subject string `gorm:"index;not null"`
	Role    string `gorm:"not null"`
}

// TableName specifies the table name for RoleBinding.
func (RoleBinding) TableName() string {
	return "opa_role_bindings"
}

// Policy represents a policy rule in the database.
type Policy struct {
	ID      uint   `gorm:"primaryKey"`
	Subject string `gorm:"index;not null"`
	Object  string `gorm:"not null"`
	Action  string `gorm:"not null"`
}

// TableName specifies the table name for Policy.
func (Policy) TableName() string {
	return "opa_policies"
}
