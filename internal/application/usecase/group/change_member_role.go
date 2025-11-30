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

// ChangeMemberRoleInput represents the input for changing a member's role.
type ChangeMemberRoleInput struct {
	GroupID    uuid.UUID
	MemberID   uuid.UUID
	NewRole    entity.MemberRole
	RequesterID uuid.UUID
}

// ChangeMemberRoleOutput represents the output of changing a member's role.
type ChangeMemberRoleOutput struct {
	Member *entity.GroupMember
}

// ChangeMemberRoleUseCase handles changing member roles.
type ChangeMemberRoleUseCase struct {
	groupRepo adapter.GroupRepository
}

// NewChangeMemberRoleUseCase creates a new ChangeMemberRoleUseCase instance.
func NewChangeMemberRoleUseCase(groupRepo adapter.GroupRepository) *ChangeMemberRoleUseCase {
	return &ChangeMemberRoleUseCase{
		groupRepo: groupRepo,
	}
}

// Execute performs the role change.
func (uc *ChangeMemberRoleUseCase) Execute(ctx context.Context, input ChangeMemberRoleInput) (*ChangeMemberRoleOutput, error) {
	// Validate role
	if !isValidMemberRole(input.NewRole) {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeInvalidMemberRole,
			"role must be 'admin' or 'member'",
			domainerror.ErrInvalidMemberRole,
		)
	}

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
			"only admins can change member roles",
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

	// If demoting from admin to member, check if this would leave the group without admins
	if targetMember.Role == entity.MemberRoleAdmin && input.NewRole == entity.MemberRoleMember {
		adminCount, err := uc.groupRepo.CountAdminsByGroupID(ctx, input.GroupID)
		if err != nil {
			return nil, fmt.Errorf("failed to count admins: %w", err)
		}
		if adminCount <= 1 {
			return nil, domainerror.NewGroupError(
				domainerror.ErrCodeCannotRemoveSoleAdmin,
				"cannot demote: this is the only admin",
				domainerror.ErrCannotRemoveSoleAdmin,
			)
		}
	}

	// Update the role
	targetMember.Role = input.NewRole
	if err := uc.groupRepo.UpdateMember(ctx, targetMember); err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	return &ChangeMemberRoleOutput{
		Member: targetMember,
	}, nil
}

// isValidMemberRole validates the member role.
func isValidMemberRole(role entity.MemberRole) bool {
	return role == entity.MemberRoleAdmin || role == entity.MemberRoleMember
}
