package models

import "time"

// SettingValueType represents the type of setting value
type SettingValueType string

const (
	SettingValueTypeString SettingValueType = "string"
	SettingValueTypeInt    SettingValueType = "int"
	SettingValueTypeBool   SettingValueType = "bool"
)

// SettingCategory represents the category of a setting
type SettingCategory string

const (
	SettingCategoryGeneral  SettingCategory = "general"
	SettingCategoryAuth     SettingCategory = "auth"
	SettingCategorySecurity SettingCategory = "security"
)

// Setting represents a system configuration setting
type Setting struct {
	Key         string           `json:"key" db:"key"`
	Value       string           `json:"value" db:"value"`
	ValueType   SettingValueType `json:"value_type" db:"value_type"`
	Description *string          `json:"description,omitempty" db:"description"`
	Category    SettingCategory  `json:"category" db:"category"`
	IsPublic    bool             `json:"is_public" db:"is_public"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

// SettingResponse represents a setting in API responses
type SettingResponse struct {
	Key         string           `json:"key"`
	Value       string           `json:"value"`
	ValueType   SettingValueType `json:"value_type"`
	Description *string          `json:"description,omitempty"`
	Category    SettingCategory  `json:"category"`
	IsPublic    bool             `json:"is_public"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// UpdateSettingRequest represents the request to update a setting
type UpdateSettingRequest struct {
	Value string `json:"value" binding:"required"`
}

// BatchUpdateSettingsRequest represents the request to update multiple settings
type BatchUpdateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// GetSettingsResponse represents the response for getting all settings
type GetSettingsResponse struct {
	Settings []SettingResponse `json:"settings"`
}

// ToResponse converts a Setting to SettingResponse
func (s *Setting) ToResponse() SettingResponse {
	return SettingResponse{
		Key:         s.Key,
		Value:       s.Value,
		ValueType:   s.ValueType,
		Description: s.Description,
		Category:    s.Category,
		IsPublic:    s.IsPublic,
		UpdatedAt:   s.UpdatedAt,
	}
}
