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

// DeleteGroupInput represents the input for deleting a group.
type DeleteGroupInput struct {
	GroupID     uuid.UUID
	RequesterID uuid.UUID
}

// DeleteGroupOutput represents the output of deleting a group.
type DeleteGroupOutput struct {
	Success bool
}

// DeleteGroupUseCase handles deleting a group.
type DeleteGroupUseCase struct {
	groupRepo adapter.GroupRepository
}

// NewDeleteGroupUseCase creates a new DeleteGroupUseCase instance.
func NewDeleteGroupUseCase(groupRepo adapter.GroupRepository) *DeleteGroupUseCase {
	return &DeleteGroupUseCase{
		groupRepo: groupRepo,
	}
}

// Execute performs the group deletion operation.
func (uc *DeleteGroupUseCase) Execute(ctx context.Context, input DeleteGroupInput) (*DeleteGroupOutput, error) {
	// Verify requester is a member of the group
	member, err := uc.groupRepo.FindMemberByGroupAndUser(ctx, input.GroupID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeNotGroupMember,
			"you are not a member of this group",
			domainerror.ErrNotGroupMember,
		)
	}

	// Only admins can delete groups
	if member.Role != entity.MemberRoleAdmin {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeNotGroupAdmin,
			"only group admins can delete groups",
			domainerror.ErrNotGroupAdmin,
		)
	}

	// Delete the group (this should cascade to members)
	if err := uc.groupRepo.DeleteGroup(ctx, input.GroupID); err != nil {
		return nil, fmt.Errorf("failed to delete group: %w", err)
	}

	return &DeleteGroupOutput{
		Success: true,
	}, nil
}
