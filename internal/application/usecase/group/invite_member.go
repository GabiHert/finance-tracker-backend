// Package group contains group-related use cases.
package group

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

const (
	// InviteTokenLength is the length of the invite token in bytes.
	InviteTokenLength = 32
	// InviteExpirationDays is the number of days until an invite expires.
	InviteExpirationDays = 7
)

// emailRegex is compiled once at package level for performance.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// InviteMemberInput represents the input for inviting a member.
type InviteMemberInput struct {
	GroupID   uuid.UUID
	Email     string
	InviterID uuid.UUID
}

// InviteMemberOutput represents the output of inviting a member.
type InviteMemberOutput struct {
	Invite *entity.GroupInvite
}

// InviteMemberUseCase handles inviting members to a group.
type InviteMemberUseCase struct {
	groupRepo adapter.GroupRepository
	userRepo  adapter.UserRepository
}

// NewInviteMemberUseCase creates a new InviteMemberUseCase instance.
func NewInviteMemberUseCase(groupRepo adapter.GroupRepository, userRepo adapter.UserRepository) *InviteMemberUseCase {
	return &InviteMemberUseCase{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// Execute performs the member invitation.
func (uc *InviteMemberUseCase) Execute(ctx context.Context, input InviteMemberInput) (*InviteMemberOutput, error) {
	// Normalize email
	email := strings.ToLower(strings.TrimSpace(input.Email))

	// Validate email format
	if !isValidEmail(email) {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeInvalidGroupEmail,
			"invalid email address",
			domainerror.ErrInvalidGroupEmail,
		)
	}

	// Check if inviter is an admin of the group
	inviterMember, err := uc.groupRepo.FindMemberByGroupAndUser(ctx, input.GroupID, input.InviterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check inviter membership: %w", err)
	}
	if inviterMember == nil {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeNotGroupMember,
			"you are not a member of this group",
			domainerror.ErrNotGroupMember,
		)
	}
	if inviterMember.Role != entity.MemberRoleAdmin {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeNotGroupAdmin,
			"only admins can invite members",
			domainerror.ErrNotGroupAdmin,
		)
	}

	// Get inviter info to check if they're trying to invite themselves
	inviter, err := uc.userRepo.FindByID(ctx, input.InviterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inviter info: %w", err)
	}
	if strings.EqualFold(inviter.Email, email) {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeCannotInviteSelf,
			"you cannot invite yourself",
			domainerror.ErrCannotInviteSelf,
		)
	}

	// Check if user is already a member (by email)
	existingUser, err := uc.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		isMember, err := uc.groupRepo.IsUserMemberOfGroup(ctx, input.GroupID, existingUser.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing membership: %w", err)
		}
		if isMember {
			return nil, domainerror.NewGroupError(
				domainerror.ErrCodeUserAlreadyMember,
				"user is already a member of this group",
				domainerror.ErrUserAlreadyMember,
			)
		}
	}

	// Check if there's already a pending invite for this email
	existingInvite, err := uc.groupRepo.FindPendingInviteByGroupAndEmail(ctx, input.GroupID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing invites: %w", err)
	}
	if existingInvite != nil {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeInviteAlreadyExists,
			"an invite already exists for this email",
			domainerror.ErrInviteAlreadyExists,
		)
	}

	// Generate invite token
	token, err := generateInviteToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite token: %w", err)
	}

	// Create invite
	expiresAt := time.Now().UTC().AddDate(0, 0, InviteExpirationDays)
	invite := entity.NewGroupInvite(input.GroupID, email, token, input.InviterID, expiresAt)

	if err := uc.groupRepo.CreateInvite(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create invite: %w", err)
	}

	return &InviteMemberOutput{
		Invite: invite,
	}, nil
}

// isValidEmail validates email format.
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// generateInviteToken generates a secure random token for invites.
func generateInviteToken() (string, error) {
	bytes := make([]byte, InviteTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
