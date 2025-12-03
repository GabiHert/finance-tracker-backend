// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// GroupRepository defines the interface for group persistence operations.
type GroupRepository interface {
	// CreateGroup creates a new group in the database.
	CreateGroup(ctx context.Context, group *entity.Group) error

	// FindGroupByID retrieves a group by its ID.
	FindGroupByID(ctx context.Context, id uuid.UUID) (*entity.Group, error)

	// FindGroupsByUserID retrieves all groups a user belongs to.
	FindGroupsByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.GroupListItem, error)

	// UpdateGroup updates an existing group in the database.
	UpdateGroup(ctx context.Context, group *entity.Group) error

	// DeleteGroup removes a group from the database.
	DeleteGroup(ctx context.Context, id uuid.UUID) error

	// CreateMember adds a new member to a group.
	CreateMember(ctx context.Context, member *entity.GroupMember) error

	// FindMemberByID retrieves a group member by their ID.
	FindMemberByID(ctx context.Context, id uuid.UUID) (*entity.GroupMember, error)

	// FindMemberByGroupAndUser retrieves a member by group and user ID.
	FindMemberByGroupAndUser(ctx context.Context, groupID, userID uuid.UUID) (*entity.GroupMember, error)

	// FindMembersByGroupID retrieves all members of a group.
	FindMembersByGroupID(ctx context.Context, groupID uuid.UUID) ([]*entity.GroupMember, error)

	// UpdateMember updates a group member.
	UpdateMember(ctx context.Context, member *entity.GroupMember) error

	// DeleteMember removes a member from a group.
	DeleteMember(ctx context.Context, id uuid.UUID) error

	// CountAdminsByGroupID counts the number of admins in a group.
	CountAdminsByGroupID(ctx context.Context, groupID uuid.UUID) (int, error)

	// CreateInvite creates a new group invitation.
	CreateInvite(ctx context.Context, invite *entity.GroupInvite) error

	// FindInviteByToken retrieves an invitation by its token.
	FindInviteByToken(ctx context.Context, token string) (*entity.GroupInvite, error)

	// FindPendingInviteByGroupAndEmail retrieves a pending invite by group and email.
	FindPendingInviteByGroupAndEmail(ctx context.Context, groupID uuid.UUID, email string) (*entity.GroupInvite, error)

	// FindPendingInvitesByGroupID retrieves all pending invites for a group.
	FindPendingInvitesByGroupID(ctx context.Context, groupID uuid.UUID) ([]*entity.GroupInvite, error)

	// UpdateInvite updates an invitation.
	UpdateInvite(ctx context.Context, invite *entity.GroupInvite) error

	// IsUserMemberOfGroup checks if a user is a member of a group.
	IsUserMemberOfGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error)

	// GetGroupWithMembers retrieves a group with its members and pending invites.
	GetGroupWithMembers(ctx context.Context, groupID uuid.UUID) (*entity.GroupWithMembers, error)

	// GetGroupDashboard retrieves comprehensive dashboard data for a group.
	GetGroupDashboard(ctx context.Context, groupID uuid.UUID, startDate, endDate time.Time) (*entity.GroupDashboardData, error)

	// GetGroupDashboardPreviousPeriod retrieves dashboard totals for comparison period.
	GetGroupDashboardPreviousPeriod(ctx context.Context, groupID uuid.UUID, startDate, endDate time.Time) (totalExpenses, totalIncome float64, err error)
}
