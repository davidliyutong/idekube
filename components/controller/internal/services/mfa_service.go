package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/pquerna/otp/totp"
)

// MFAService handles multi-factor authentication
type MFAService struct {
	userRepo *repository.UserRepository
}

// NewMFAService creates a new MFA service
func NewMFAService(userRepo *repository.UserRepository) *MFAService {
	return &MFAService{
		userRepo: userRepo,
	}
}

// EnableMFA enables MFA for a user
func (s *MFAService) EnableMFA(ctx context.Context, userID int64) (*models.MFASetup, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.MFAEnabled {
		return nil, fmt.Errorf("MFA is already enabled")
	}

	// Get account name
	accountName := user.Username
	if user.Email != nil {
		accountName = *user.Email
	}

	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "IDEKube",
		AccountName: accountName,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Generate backup codes
	backupCodes := make([]string, 10)
	for i := 0; i < 10; i++ {
		code, err := generateBackupCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		backupCodes[i] = code
	}

	// Store encrypted secret (in production, encrypt this!)
	secret := key.Secret()
	user.MFASecret = &secret
	user.MFABackupCodes = backupCodes

	// Don't enable yet - user needs to verify first
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &models.MFASetup{
		Secret:      key.Secret(),
		QRCode:      key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// VerifyAndEnableMFA verifies TOTP code and enables MFA
func (s *MFAService) VerifyAndEnableMFA(ctx context.Context, userID int64, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.MFASecret == nil {
		return fmt.Errorf("MFA setup not initiated")
	}

	// Verify TOTP code
	valid := totp.Validate(code, *user.MFASecret)
	if !valid {
		return fmt.Errorf("invalid verification code")
	}

	// Enable MFA
	user.MFAEnabled = true
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to enable MFA: %w", err)
	}

	return nil
}

// DisableMFA disables MFA for a user
func (s *MFAService) DisableMFA(ctx context.Context, userID int64, password string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !user.MFAEnabled {
		return fmt.Errorf("MFA is not enabled")
	}

	// Verify password (implement password verification)
	// ... password verification logic ...

	// Disable MFA
	user.MFAEnabled = false
	user.MFASecret = nil
	user.MFABackupCodes = nil

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	return nil
}

// VerifyMFACode verifies an MFA code
func (s *MFAService) VerifyMFACode(ctx context.Context, userID int64, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !user.MFAEnabled || user.MFASecret == nil {
		return fmt.Errorf("MFA is not enabled")
	}

	// Try TOTP code first
	valid := totp.Validate(code, *user.MFASecret)
	if valid {
		return nil
	}

	// Try backup codes
	if user.MFABackupCodes != nil {
		for i, backupCode := range user.MFABackupCodes {
			if backupCode == code {
				// Remove used backup code
				user.MFABackupCodes = append(user.MFABackupCodes[:i], user.MFABackupCodes[i+1:]...)
				_ = s.userRepo.Update(ctx, user)
				return nil
			}
		}
	}

	return fmt.Errorf("invalid MFA code")
}

// GenerateBackupCodes generates new backup codes
func (s *MFAService) GenerateBackupCodes(ctx context.Context, userID int64) ([]string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.MFAEnabled {
		return nil, fmt.Errorf("MFA is not enabled")
	}

	// Generate new backup codes
	backupCodes := make([]string, 10)
	for i := 0; i < 10; i++ {
		code, err := generateBackupCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		backupCodes[i] = code
	}

	// Update user
	user.MFABackupCodes = backupCodes
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update backup codes: %w", err)
	}

	return backupCodes, nil
}

// generateBackupCode generates a random backup code
func generateBackupCode() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:12], nil
}
