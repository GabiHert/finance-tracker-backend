// Package model defines database models for persistence layer.
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// CategoryModel represents the categories table in the database.
type CategoryModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name      string         `gorm:"type:varchar(50);not null"`
	Color     string         `gorm:"type:varchar(7);default:'#6366F1'"`
	Icon      string         `gorm:"type:varchar(50);default:'tag'"`
	OwnerType string         `gorm:"type:varchar(10);not null"`
	OwnerID   uuid.UUID      `gorm:"type:uuid;not null;index"`
	Type      string         `gorm:"type:varchar(10);not null"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft-delete support
}

// TableName returns the table name for the CategoryModel.
func (CategoryModel) TableName() string {
	return "categories"
}

// ToEntity converts a CategoryModel to a domain Category entity.
func (m *CategoryModel) ToEntity() *entity.Category {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &entity.Category{
		ID:        m.ID,
		Name:      m.Name,
		Color:     m.Color,
		Icon:      m.Icon,
		OwnerType: entity.OwnerType(m.OwnerType),
		OwnerID:   m.OwnerID,
		Type:      entity.CategoryType(m.Type),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// CategoryFromEntity creates a CategoryModel from a domain Category entity.
func CategoryFromEntity(category *entity.Category) *CategoryModel {
	var deletedAt gorm.DeletedAt
	if category.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *category.DeletedAt, Valid: true}
	}

	return &CategoryModel{
		ID:        category.ID,
		Name:      category.Name,
		Color:     category.Color,
		Icon:      category.Icon,
		OwnerType: string(category.OwnerType),
		OwnerID:   category.OwnerID,
		Type:      string(category.Type),
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
