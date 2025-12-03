// Package entity defines the core business entities for the domain layer.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// MemberRole represents the role of a member in a group.
type MemberRole string

const (
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
)

// InviteStatus represents the status of a group invitation.
type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusDeclined InviteStatus = "declined"
	InviteStatusExpired  InviteStatus = "expired"
)

// Group represents a collaborative group in the Finance Tracker system.
type Group struct {
	ID        uuid.UUID
	Name      string
	CreatedBy uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewGroup creates a new Group entity.
func NewGroup(name string, createdBy uuid.UUID) *Group {
	now := time.Now().UTC()

	return &Group{
		ID:        uuid.New(),
		Name:      name,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GroupMember represents a member of a group.
type GroupMember struct {
	ID       uuid.UUID
	GroupID  uuid.UUID
	UserID   uuid.UUID
	Role     MemberRole
	JoinedAt time.Time
	// User information (populated when needed)
	UserName  string
	UserEmail string
}

// NewGroupMember creates a new GroupMember entity.
func NewGroupMember(groupID, userID uuid.UUID, role MemberRole) *GroupMember {
	return &GroupMember{
		ID:       uuid.New(),
		GroupID:  groupID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now().UTC(),
	}
}

// GroupInvite represents an invitation to join a group.
type GroupInvite struct {
	ID        uuid.UUID
	GroupID   uuid.UUID
	Email     string
	Token     string
	InvitedBy uuid.UUID
	Status    InviteStatus
	ExpiresAt time.Time
	CreatedAt time.Time
}

// NewGroupInvite creates a new GroupInvite entity.
func NewGroupInvite(groupID uuid.UUID, email, token string, invitedBy uuid.UUID, expiresAt time.Time) *GroupInvite {
	return &GroupInvite{
		ID:        uuid.New(),
		GroupID:   groupID,
		Email:     email,
		Token:     token,
		InvitedBy: invitedBy,
		Status:    InviteStatusPending,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}
}

// IsExpired checks if the invitation has expired.
func (i *GroupInvite) IsExpired() bool {
	return time.Now().UTC().After(i.ExpiresAt)
}

// GroupWithMembers represents a group with its members.
type GroupWithMembers struct {
	Group          *Group
	Members        []*GroupMember
	PendingInvites []*GroupInvite
	MemberCount    int
	UserRole       MemberRole
}

// GroupListItem represents a group in a list view.
type GroupListItem struct {
	ID          uuid.UUID
	Name        string
	MemberCount int
	Role        MemberRole
	CreatedAt   time.Time
}

// GroupDashboardData represents comprehensive dashboard data for a group.
type GroupDashboardData struct {
	Summary            *GroupDashboardSummary
	CategoryBreakdown  []*GroupCategoryBreakdown
	MemberBreakdown    []*GroupMemberBreakdown
	Trends             []*GroupTrendPoint
	RecentTransactions []*GroupDashboardTransaction
}

// GroupDashboardSummary represents the summary section of the group dashboard.
type GroupDashboardSummary struct {
	TotalExpenses  float64
	TotalIncome    float64
	NetBalance     float64
	MemberCount    int
	ExpensesChange float64
	IncomeChange   float64
}

// GroupCategoryBreakdown represents a category's contribution to group expenses/income.
type GroupCategoryBreakdown struct {
	CategoryID    uuid.UUID
	CategoryName  string
	CategoryColor string
	Amount        float64
	Percentage    float64
}

// GroupMemberBreakdown represents a member's contribution to group expenses/income.
type GroupMemberBreakdown struct {
	MemberID         uuid.UUID
	MemberName       string
	AvatarURL        string
	Total            float64
	Percentage       float64
	TransactionCount int
}

// GroupTrendPoint represents a single data point in the trends chart.
type GroupTrendPoint struct {
	Date     time.Time
	Income   float64
	Expenses float64
}

// GroupDashboardTransaction represents a transaction in the group dashboard.
type GroupDashboardTransaction struct {
	ID              uuid.UUID
	Description     string
	Amount          float64
	Date            time.Time
	CategoryName    string
	CategoryColor   string
	MemberID        uuid.UUID
	MemberName      string
	MemberAvatarURL string
}
