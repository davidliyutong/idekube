package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

// EmailService handles email operations
type EmailService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.OAuthSessionRepository
	smtpHost    string
	smtpPort    int
	smtpUser    string
	smtpPass    string
	fromEmail   string
	fromName    string
	baseURL     string
}

// NewEmailService creates a new email service
func NewEmailService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.OAuthSessionRepository,
	smtpHost string,
	smtpPort int,
	smtpUser string,
	smtpPass string,
	fromEmail string,
	fromName string,
	baseURL string,
) *EmailService {
	return &EmailService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		smtpHost:    smtpHost,
		smtpPort:    smtpPort,
		smtpUser:    smtpUser,
		smtpPass:    smtpPass,
		fromEmail:   fromEmail,
		fromName:    fromName,
		baseURL:     baseURL,
	}
}

// SendVerificationEmail sends email verification link
func (s *EmailService) SendVerificationEmail(ctx context.Context, user *models.User) error {
	// Generate verification token
	token, err := generateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in session
	session := &models.OAuthSession{
		Key:       fmt.Sprintf("email_verify_%s", token),
		Value:     fmt.Sprintf("%d", user.ID),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Generate verification URL
	verifyURL := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", s.baseURL, token)

	// Send email
	subject := "Verify your email address"
	body := fmt.Sprintf(`
<html>
<body>
	<h2>Welcome to IDEKube!</h2>
	<p>Hi %s,</p>
	<p>Please click the link below to verify your email address:</p>
	<p><a href="%s">Verify Email</a></p>
	<p>This link will expire in 24 hours.</p>
	<p>If you didn't create an account, please ignore this email.</p>
</body>
</html>
`, user.Username, verifyURL)

	emailAddr := ""
	if user.Email != nil {
		emailAddr = *user.Email
	}
	return s.sendEmail(emailAddr, subject, body)
}

// VerifyEmail verifies email with token
func (s *EmailService) VerifyEmail(ctx context.Context, token string) error {
	// Get session
	session, err := s.sessionRepo.GetByKey(ctx, fmt.Sprintf("email_verify_%s", token))
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	// Parse user ID
	var userID int64
	_, err = fmt.Sscanf(session.Value, "%d", &userID)
	if err != nil {
		return fmt.Errorf("invalid token")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Update email verified status
	user.EmailVerified = true
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Delete used token
	_ = s.sessionRepo.DeleteByKey(ctx, fmt.Sprintf("email_verify_%s", token))

	return nil
}

// SendPasswordResetEmail sends password reset link
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email string) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists
		return nil
	}

	// Generate reset token
	token, err := generateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in session
	session := &models.OAuthSession{
		Key:       fmt.Sprintf("password_reset_%s", token),
		Value:     fmt.Sprintf("%d", user.ID),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Generate reset URL
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

	// Send email
	subject := "Reset your password"
	body := fmt.Sprintf(`
<html>
<body>
	<h2>Password Reset Request</h2>
	<p>Hi %s,</p>
	<p>We received a request to reset your password. Click the link below to reset it:</p>
	<p><a href="%s">Reset Password</a></p>
	<p>This link will expire in 1 hour.</p>
	<p>If you didn't request this, please ignore this email.</p>
</body>
</html>
`, user.Username, resetURL)

	emailAddr := ""
	if user.Email != nil {
		emailAddr = *user.Email
	}
	return s.sendEmail(emailAddr, subject, body)
}

// ResetPassword resets password with token
func (s *EmailService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Get session
	session, err := s.sessionRepo.GetByKey(ctx, fmt.Sprintf("password_reset_%s", token))
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	// Parse user ID
	var userID int64
	_, err = fmt.Sscanf(session.Value, "%d", &userID)
	if err != nil {
		return fmt.Errorf("invalid token")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete used token
	_ = s.sessionRepo.DeleteByKey(ctx, fmt.Sprintf("password_reset_%s", token))

	return nil
}

// sendEmail sends an email using SMTP
func (s *EmailService) sendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUser, s.smtpPass)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// generateRandomToken generates a random token
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
