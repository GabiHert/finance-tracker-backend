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

// RemoveMemberInput represents the input for removing a member.
type RemoveMemberInput struct {
	GroupID     uuid.UUID
	MemberID    uuid.UUID
	RequesterID uuid.UUID
}

// RemoveMemberOutput represents the output of removing a member.
type RemoveMemberOutput struct {
	Success bool
}

// RemoveMemberUseCase handles removing members from a group.
type RemoveMemberUseCase struct {
	groupRepo adapter.GroupRepository
}

// NewRemoveMemberUseCase creates a new RemoveMemberUseCase instance.
func NewRemoveMemberUseCase(groupRepo adapter.GroupRepository) *RemoveMemberUseCase {
	return &RemoveMemberUseCase{
		groupRepo: groupRepo,
	}
}

// Execute performs the member removal.
func (uc *RemoveMemberUseCase) Execute(ctx context.Context, input RemoveMemberInput) (*RemoveMemberOutput, error) {
	// Check if requester is an admin
	requesterMember, err := uc.groupRepo.FindMemberByGroupAndUser(ctx, input.GroupID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check requester membership: %w", err)
	}
	if requesterMember == nil {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeNotGroupMember,
			"you are not a member of this group",
			domainerror.ErrNotGroupMember,
		)
	}
	if requesterMember.Role != entity.MemberRoleAdmin {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeNotGroupAdmin,
			"only admins can remove members",
			domainerror.ErrNotGroupAdmin,
		)
	}

	// Find the target member
	targetMember, err := uc.groupRepo.FindMemberByID(ctx, input.MemberID)
	if err != nil {
		return nil, fmt.Errorf("failed to find member: %w", err)
	}
	if targetMember == nil || targetMember.GroupID != input.GroupID {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeMemberNotFound,
			"member not found in this group",
			domainerror.ErrMemberNotFound,
		)
	}

	// Cannot remove self via this endpoint (use LeaveGroup instead)
	if targetMember.UserID == input.RequesterID {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeCannotRemoveSoleAdmin,
			"use leave group endpoint to remove yourself",
			domainerror.ErrCannotRemoveSoleAdmin,
		)
	}

	// Delete the member
	if err := uc.groupRepo.DeleteMember(ctx, targetMember.ID); err != nil {
		return nil, fmt.Errorf("failed to remove member: %w", err)
	}

	return &RemoveMemberOutput{
		Success: true,
	}, nil
}
