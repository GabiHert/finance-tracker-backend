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

// GetGroupInput represents the input for getting group details.
type GetGroupInput struct {
	GroupID uuid.UUID
	UserID  uuid.UUID
}

// GetGroupOutput represents the output of getting group details.
type GetGroupOutput struct {
	Group          *entity.Group
	Members        []*entity.GroupMember
	PendingInvites []*entity.GroupInvite
	UserRole       entity.MemberRole
}

// GetGroupUseCase handles getting group details.
type GetGroupUseCase struct {
	groupRepo adapter.GroupRepository
}

// NewGetGroupUseCase creates a new GetGroupUseCase instance.
func NewGetGroupUseCase(groupRepo adapter.GroupRepository) *GetGroupUseCase {
	return &GetGroupUseCase{
		groupRepo: groupRepo,
	}
}

// Execute performs the group retrieval.
func (uc *GetGroupUseCase) Execute(ctx context.Context, input GetGroupInput) (*GetGroupOutput, error) {
	// Check if user is a member of the group
	member, err := uc.groupRepo.FindMemberByGroupAndUser(ctx, input.GroupID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeGroupNotFound,
			"group not found",
			domainerror.ErrGroupNotFound,
		)
	}

	// Get group with members
	groupWithMembers, err := uc.groupRepo.GetGroupWithMembers(ctx, input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	if groupWithMembers == nil {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeGroupNotFound,
			"group not found",
			domainerror.ErrGroupNotFound,
		)
	}

	return &GetGroupOutput{
		Group:          groupWithMembers.Group,
		Members:        groupWithMembers.Members,
		PendingInvites: groupWithMembers.PendingInvites,
		UserRole:       member.Role,
	}, nil
}
