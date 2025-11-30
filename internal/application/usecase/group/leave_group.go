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

// LeaveGroupInput represents the input for leaving a group.
type LeaveGroupInput struct {
	GroupID uuid.UUID
	UserID  uuid.UUID
}

// LeaveGroupOutput represents the output of leaving a group.
type LeaveGroupOutput struct {
	Success bool
}

// LeaveGroupUseCase handles leaving a group.
type LeaveGroupUseCase struct {
	groupRepo adapter.GroupRepository
}

// NewLeaveGroupUseCase creates a new LeaveGroupUseCase instance.
func NewLeaveGroupUseCase(groupRepo adapter.GroupRepository) *LeaveGroupUseCase {
	return &LeaveGroupUseCase{
		groupRepo: groupRepo,
	}
}

// Execute performs the group leave operation.
func (uc *LeaveGroupUseCase) Execute(ctx context.Context, input LeaveGroupInput) (*LeaveGroupOutput, error) {
	// Find user's membership
	member, err := uc.groupRepo.FindMemberByGroupAndUser(ctx, input.GroupID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find membership: %w", err)
	}
	if member == nil {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeNotGroupMember,
			"you are not a member of this group",
			domainerror.ErrNotGroupMember,
		)
	}

	// If user is an admin, check if they're the only admin
	if member.Role == entity.MemberRoleAdmin {
		adminCount, err := uc.groupRepo.CountAdminsByGroupID(ctx, input.GroupID)
		if err != nil {
			return nil, fmt.Errorf("failed to count admins: %w", err)
		}
		if adminCount <= 1 {
			// Check if there are other members who could become admin
			members, err := uc.groupRepo.FindMembersByGroupID(ctx, input.GroupID)
			if err != nil {
				return nil, fmt.Errorf("failed to find members: %w", err)
			}
			// If user is the only member, they can leave (group will be empty)
			// Otherwise, they need to promote someone else first
			if len(members) > 1 {
				return nil, domainerror.NewGroupError(
					domainerror.ErrCodeCannotRemoveSoleAdmin,
					"cannot leave: you are the only admin. Promote another member first.",
					domainerror.ErrCannotRemoveSoleAdmin,
				)
			}
		}
	}

	// Delete the membership
	if err := uc.groupRepo.DeleteMember(ctx, member.ID); err != nil {
		return nil, fmt.Errorf("failed to leave group: %w", err)
	}

	return &LeaveGroupOutput{
		Success: true,
	}, nil
}
