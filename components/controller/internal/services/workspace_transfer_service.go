package services

import (
	"context"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
)

// WorkspaceTransferService handles workspace transfer business logic
type WorkspaceTransferService struct {
	transferRepo  *repository.WorkspaceTransferRepository
	workspaceRepo *repository.WorkspaceRepository
	userRepo      *repository.UserRepository
}

// NewWorkspaceTransferService creates a new workspace transfer service
func NewWorkspaceTransferService(
	transferRepo *repository.WorkspaceTransferRepository,
	workspaceRepo *repository.WorkspaceRepository,
	userRepo *repository.UserRepository,
) *WorkspaceTransferService {
	return &WorkspaceTransferService{
		transferRepo:  transferRepo,
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

// CreateTransfer creates a new workspace transfer request
func (s *WorkspaceTransferService) CreateTransfer(
	ctx context.Context,
	workspaceID int64,
	fromUserID int64,
	req *models.CreateWorkspaceTransferRequest,
) (*models.WorkspaceTransfer, error) {
	// Get workspace to verify ownership
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}

	// Verify the requester is the owner
	if workspace.OwnerType != models.OwnerTypeUser || workspace.OwnerID != fromUserID {
		return nil, fmt.Errorf("only workspace owner can transfer ownership")
	}

	// Check if there's already a pending transfer
	hasPending, err := s.transferRepo.HasPendingTransfer(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check pending transfers: %w", err)
	}
	if hasPending {
		return nil, fmt.Errorf("workspace already has a pending transfer request")
	}

	// Verify target user exists
	toUser, err := s.userRepo.GetByUsername(ctx, req.ToUsername)
	if err != nil {
		return nil, fmt.Errorf("target user '%s' not found", req.ToUsername)
	}

	// Prevent self-transfer
	if toUser.ID == fromUserID {
		return nil, fmt.Errorf("cannot transfer workspace to yourself")
	}

	// Create transfer request
	transfer := &models.WorkspaceTransfer{
		WorkspaceID: workspaceID,
		FromUserID:  fromUserID,
		ToUsername:  req.ToUsername,
		ToUserID:    &toUser.ID,
		Status:      models.WorkspaceTransferStatusPending,
		Message:     req.Message,
	}

	if err := s.transferRepo.Create(ctx, transfer); err != nil {
		return nil, fmt.Errorf("failed to create transfer request: %w", err)
	}

	return transfer, nil
}

// RespondToTransfer allows a user to accept or reject a transfer request
func (s *WorkspaceTransferService) RespondToTransfer(
	ctx context.Context,
	transferID int64,
	userID int64,
	req *models.RespondWorkspaceTransferRequest,
) (*models.WorkspaceTransfer, error) {
	// Get transfer request
	transfer, err := s.transferRepo.GetByID(ctx, transferID)
	if err != nil {
		return nil, fmt.Errorf("transfer request not found: %w", err)
	}

	// Verify the user is the recipient
	if transfer.ToUserID == nil || *transfer.ToUserID != userID {
		return nil, fmt.Errorf("only the recipient can respond to this transfer request")
	}

	// Check if transfer is still pending
	if transfer.Status != models.WorkspaceTransferStatusPending {
		return nil, fmt.Errorf("transfer request is no longer pending")
	}

	now := time.Now()
	transfer.RespondedAt = &now
	transfer.Message = req.Message

	if req.Accept {
		// Accept: change workspace ownership
		transfer.Status = models.WorkspaceTransferStatusAccepted

		workspace, err := s.workspaceRepo.GetByID(ctx, transfer.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("workspace not found: %w", err)
		}

		// Update workspace owner
		workspace.OwnerType = models.OwnerTypeUser
		workspace.OwnerID = userID

		if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
			return nil, fmt.Errorf("failed to transfer workspace ownership: %w", err)
		}
	} else {
		// Reject
		transfer.Status = models.WorkspaceTransferStatusRejected
	}

	// Update transfer status
	if err := s.transferRepo.Update(ctx, transfer); err != nil {
		return nil, fmt.Errorf("failed to update transfer request: %w", err)
	}

	return transfer, nil
}

// CancelTransfer allows the initiator to cancel a pending transfer
func (s *WorkspaceTransferService) CancelTransfer(
	ctx context.Context,
	transferID int64,
	userID int64,
) error {
	// Get transfer request
	transfer, err := s.transferRepo.GetByID(ctx, transferID)
	if err != nil {
		return fmt.Errorf("transfer request not found: %w", err)
	}

	// Verify the user is the initiator
	if transfer.FromUserID != userID {
		return fmt.Errorf("only the initiator can cancel this transfer request")
	}

	// Check if transfer is still pending
	if transfer.Status != models.WorkspaceTransferStatusPending {
		return fmt.Errorf("transfer request is no longer pending")
	}

	// Cancel transfer
	now := time.Now()
	transfer.Status = models.WorkspaceTransferStatusCancelled
	transfer.RespondedAt = &now

	if err := s.transferRepo.Update(ctx, transfer); err != nil {
		return fmt.Errorf("failed to cancel transfer request: %w", err)
	}

	return nil
}

// ListPendingTransfersForUser lists all pending transfers for a user (as recipient)
func (s *WorkspaceTransferService) ListPendingTransfersForUser(ctx context.Context, userID int64) ([]*models.WorkspaceTransfer, error) {
	return s.transferRepo.ListPendingForUser(ctx, userID)
}

// GetTransfer retrieves a specific transfer request
func (s *WorkspaceTransferService) GetTransfer(ctx context.Context, transferID int64) (*models.WorkspaceTransfer, error) {
	return s.transferRepo.GetByID(ctx, transferID)
}

// ListTransfersByWorkspace lists all transfer history for a workspace
func (s *WorkspaceTransferService) ListTransfersByWorkspace(ctx context.Context, workspaceID int64) ([]*models.WorkspaceTransfer, error) {
	return s.transferRepo.ListByWorkspace(ctx, workspaceID)
}
