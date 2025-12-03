// Package model defines database models for persistence layer.
package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// GroupModel represents the groups table in the database.
type GroupModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:varchar(100);not null"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName returns the table name for the GroupModel.
func (GroupModel) TableName() string {
	return "groups"
}

// ToEntity converts a GroupModel to a domain Group entity.
func (m *GroupModel) ToEntity() *entity.Group {
	return &entity.Group{
		ID:        m.ID,
		Name:      m.Name,
		CreatedBy: m.CreatedBy,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// GroupFromEntity creates a GroupModel from a domain Group entity.
func GroupFromEntity(group *entity.Group) *GroupModel {
	return &GroupModel{
		ID:        group.ID,
		Name:      group.Name,
		CreatedBy: group.CreatedBy,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}
}

// GroupMemberModel represents the group_members table in the database.
type GroupMemberModel struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	GroupID  uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Role     string    `gorm:"type:varchar(20);not null"`
	JoinedAt time.Time `gorm:"not null"`
	// User information (joined from users table)
	UserName  string `gorm:"-"`
	UserEmail string `gorm:"-"`
}

// TableName returns the table name for the GroupMemberModel.
func (GroupMemberModel) TableName() string {
	return "group_members"
}

// ToEntity converts a GroupMemberModel to a domain GroupMember entity.
func (m *GroupMemberModel) ToEntity() *entity.GroupMember {
	return &entity.GroupMember{
		ID:        m.ID,
		GroupID:   m.GroupID,
		UserID:    m.UserID,
		Role:      entity.MemberRole(m.Role),
		JoinedAt:  m.JoinedAt,
		UserName:  m.UserName,
		UserEmail: m.UserEmail,
	}
}

// GroupMemberFromEntity creates a GroupMemberModel from a domain GroupMember entity.
func GroupMemberFromEntity(member *entity.GroupMember) *GroupMemberModel {
	return &GroupMemberModel{
		ID:       member.ID,
		GroupID:  member.GroupID,
		UserID:   member.UserID,
		Role:     string(member.Role),
		JoinedAt: member.JoinedAt,
	}
}

// GroupInviteModel represents the group_invites table in the database.
type GroupInviteModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	GroupID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Email     string    `gorm:"type:varchar(255);not null"`
	Token     string    `gorm:"type:varchar(64);not null;uniqueIndex"`
	InvitedBy uuid.UUID `gorm:"type:uuid;not null"`
	Status    string    `gorm:"type:varchar(20);not null;default:'pending'"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

// TableName returns the table name for the GroupInviteModel.
func (GroupInviteModel) TableName() string {
	return "group_invites"
}

// ToEntity converts a GroupInviteModel to a domain GroupInvite entity.
func (m *GroupInviteModel) ToEntity() *entity.GroupInvite {
	return &entity.GroupInvite{
		ID:        m.ID,
		GroupID:   m.GroupID,
		Email:     m.Email,
		Token:     m.Token,
		InvitedBy: m.InvitedBy,
		Status:    entity.InviteStatus(m.Status),
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
	}
}

// GroupInviteFromEntity creates a GroupInviteModel from a domain GroupInvite entity.
func GroupInviteFromEntity(invite *entity.GroupInvite) *GroupInviteModel {
	return &GroupInviteModel{
		ID:        invite.ID,
		GroupID:   invite.GroupID,
		Email:     invite.Email,
		Token:     invite.Token,
		InvitedBy: invite.InvitedBy,
		Status:    string(invite.Status),
		ExpiresAt: invite.ExpiresAt,
		CreatedAt: invite.CreatedAt,
	}
}
