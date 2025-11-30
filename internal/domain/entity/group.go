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
