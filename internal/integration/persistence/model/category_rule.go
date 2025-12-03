// Package model defines database models for persistence layer.
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// CategoryRuleModel represents the category_rules table in the database.
type CategoryRuleModel struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Pattern    string         `gorm:"type:varchar(255);not null"`
	CategoryID uuid.UUID      `gorm:"type:uuid;not null;index"`
	Priority   int            `gorm:"not null;default:0"`
	IsActive   bool           `gorm:"not null;default:true"`
	OwnerType  string         `gorm:"type:varchar(10);not null"`
	OwnerID    uuid.UUID      `gorm:"type:uuid;not null"`
	CreatedAt  time.Time      `gorm:"not null"`
	UpdatedAt  time.Time      `gorm:"not null"`
	DeletedAt  gorm.DeletedAt `gorm:"index"` // Soft-delete support

	// Relationships (not loaded by default, use Preload)
	Category *CategoryModel `gorm:"foreignKey:CategoryID;references:ID"`
}

// TableName returns the table name for the CategoryRuleModel.
func (CategoryRuleModel) TableName() string {
	return "category_rules"
}

// ToEntity converts a CategoryRuleModel to a domain CategoryRule entity.
func (m *CategoryRuleModel) ToEntity() *entity.CategoryRule {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &entity.CategoryRule{
		ID:         m.ID,
		Pattern:    m.Pattern,
		CategoryID: m.CategoryID,
		Priority:   m.Priority,
		IsActive:   m.IsActive,
		OwnerType:  entity.OwnerType(m.OwnerType),
		OwnerID:    m.OwnerID,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		DeletedAt:  deletedAt,
	}
}

// ToEntityWithCategory converts a CategoryRuleModel with its Category to a CategoryRuleWithCategory entity.
func (m *CategoryRuleModel) ToEntityWithCategory() *entity.CategoryRuleWithCategory {
	result := &entity.CategoryRuleWithCategory{
		Rule: m.ToEntity(),
	}

	if m.Category != nil {
		result.Category = m.Category.ToEntity()
	}

	return result
}

// CategoryRuleFromEntity creates a CategoryRuleModel from a domain CategoryRule entity.
func CategoryRuleFromEntity(rule *entity.CategoryRule) *CategoryRuleModel {
	var deletedAt gorm.DeletedAt
	if rule.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *rule.DeletedAt, Valid: true}
	}

	return &CategoryRuleModel{
		ID:         rule.ID,
		Pattern:    rule.Pattern,
		CategoryID: rule.CategoryID,
		Priority:   rule.Priority,
		IsActive:   rule.IsActive,
		OwnerType:  string(rule.OwnerType),
		OwnerID:    rule.OwnerID,
		CreatedAt:  rule.CreatedAt,
		UpdatedAt:  rule.UpdatedAt,
		DeletedAt:  deletedAt,
	}
}
