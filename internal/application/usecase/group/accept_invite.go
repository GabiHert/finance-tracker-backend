// Package group contains group-related use cases.
package group

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// AcceptInviteInput represents the input for accepting an invitation.
type AcceptInviteInput struct {
	Token  string
	UserID uuid.UUID
}

// AcceptInviteOutput represents the output of accepting an invitation.
type AcceptInviteOutput struct {
	GroupID   uuid.UUID
	GroupName string
}

// AcceptInviteUseCase handles accepting group invitations.
type AcceptInviteUseCase struct {
	groupRepo adapter.GroupRepository
	userRepo  adapter.UserRepository
}

// NewAcceptInviteUseCase creates a new AcceptInviteUseCase instance.
func NewAcceptInviteUseCase(groupRepo adapter.GroupRepository, userRepo adapter.UserRepository) *AcceptInviteUseCase {
	return &AcceptInviteUseCase{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// Execute performs the invitation acceptance.
func (uc *AcceptInviteUseCase) Execute(ctx context.Context, input AcceptInviteInput) (*AcceptInviteOutput, error) {
	// Find the invite by token
	invite, err := uc.groupRepo.FindInviteByToken(ctx, input.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to find invite: %w", err)
	}
	if invite == nil {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeInviteNotFound,
			"invite not found or invalid token",
			domainerror.ErrInviteNotFound,
		)
	}

	// Check if invite is still pending
	if invite.Status != entity.InviteStatusPending {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeInviteNotFound,
			"invite is no longer valid",
			domainerror.ErrInviteNotFound,
		)
	}

	// Check if invite has expired
	if invite.IsExpired() {
		// Update invite status to expired
		invite.Status = entity.InviteStatusExpired
		_ = uc.groupRepo.UpdateInvite(ctx, invite)

		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeInviteExpired,
			"invite has expired",
			domainerror.ErrInviteExpired,
		)
	}

	// Check if user is already a member
	isMember, err := uc.groupRepo.IsUserMemberOfGroup(ctx, invite.GroupID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeUserAlreadyMember,
			"you are already a member of this group",
			domainerror.ErrUserAlreadyMember,
		)
	}

	// Get group info
	group, err := uc.groupRepo.FindGroupByID(ctx, invite.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	if group == nil {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeGroupNotFound,
			"group not found",
			domainerror.ErrGroupNotFound,
		)
	}

	// Get user info for member details
	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create member
	member := entity.NewGroupMember(invite.GroupID, input.UserID, entity.MemberRoleMember)
	member.UserName = user.Name
	member.UserEmail = user.Email

	if err := uc.groupRepo.CreateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Update invite status
	invite.Status = entity.InviteStatusAccepted
	if err := uc.groupRepo.UpdateInvite(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to update invite status: %w", err)
	}

	return &AcceptInviteOutput{
		GroupID:   group.ID,
		GroupName: group.Name,
	}, nil
}
