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

const (
	// MaxGroupNameLength is the maximum allowed length for group names.
	MaxGroupNameLength = 100
)

// CreateGroupInput represents the input for group creation.
type CreateGroupInput struct {
	Name   string
	UserID uuid.UUID
}

// CreateGroupOutput represents the output of group creation.
type CreateGroupOutput struct {
	Group   *entity.Group
	Members []*entity.GroupMember
}

// CreateGroupUseCase handles group creation logic.
type CreateGroupUseCase struct {
	groupRepo adapter.GroupRepository
	userRepo  adapter.UserRepository
}

// NewCreateGroupUseCase creates a new CreateGroupUseCase instance.
func NewCreateGroupUseCase(groupRepo adapter.GroupRepository, userRepo adapter.UserRepository) *CreateGroupUseCase {
	return &CreateGroupUseCase{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// Execute performs the group creation.
func (uc *CreateGroupUseCase) Execute(ctx context.Context, input CreateGroupInput) (*CreateGroupOutput, error) {
	// Validate name is not empty
	if input.Name == "" {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeGroupNameRequired,
			"group name is required",
			domainerror.ErrGroupNameRequired,
		)
	}

	// Validate name length
	if len(input.Name) > MaxGroupNameLength {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeGroupNameTooLong,
			fmt.Sprintf("group name must not exceed %d characters", MaxGroupNameLength),
			domainerror.ErrGroupNameTooLong,
		)
	}

	// Get user info for member details
	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create group entity
	group := entity.NewGroup(input.Name, input.UserID)

	// Save group to database
	if err := uc.groupRepo.CreateGroup(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	// Add creator as admin member
	member := entity.NewGroupMember(group.ID, input.UserID, entity.MemberRoleAdmin)
	member.UserName = user.Name
	member.UserEmail = user.Email

	if err := uc.groupRepo.CreateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	return &CreateGroupOutput{
		Group:   group,
		Members: []*entity.GroupMember{member},
	}, nil
}
