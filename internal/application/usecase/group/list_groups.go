// Package group contains group-related use cases.
package group

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// ListGroupsInput represents the input for listing groups.
type ListGroupsInput struct {
	UserID uuid.UUID
}

// ListGroupsOutput represents the output of listing groups.
type ListGroupsOutput struct {
	Groups []*entity.GroupListItem
}

// ListGroupsUseCase handles listing user groups.
type ListGroupsUseCase struct {
	groupRepo adapter.GroupRepository
}

// NewListGroupsUseCase creates a new ListGroupsUseCase instance.
func NewListGroupsUseCase(groupRepo adapter.GroupRepository) *ListGroupsUseCase {
	return &ListGroupsUseCase{
		groupRepo: groupRepo,
	}
}

// Execute performs the group listing.
func (uc *ListGroupsUseCase) Execute(ctx context.Context, input ListGroupsInput) (*ListGroupsOutput, error) {
	groups, err := uc.groupRepo.FindGroupsByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	return &ListGroupsOutput{
		Groups: groups,
	}, nil
}
