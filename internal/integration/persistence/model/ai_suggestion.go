// Package model defines database models for persistence layer.
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// SuggestedCategoryNewJSON represents the JSONB structure for new category suggestions.
type SuggestedCategoryNewJSON struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// Value implements the driver.Valuer interface.
func (s SuggestedCategoryNewJSON) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface.
func (s *SuggestedCategoryNewJSON) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, s)
}

// AISuggestionModel represents the ai_categorization_suggestions table in the database.
type AISuggestionModel struct {
	ID                     uuid.UUID                 `gorm:"type:uuid;primaryKey"`
	UserID                 uuid.UUID                 `gorm:"type:uuid;not null;index"`
	TransactionID          uuid.UUID                 `gorm:"type:uuid;not null;index"`
	SuggestedCategoryID    *uuid.UUID                `gorm:"type:uuid;index"`
	SuggestedCategoryNew   *SuggestedCategoryNewJSON `gorm:"type:jsonb"`
	MatchType              string                    `gorm:"type:varchar(20);not null"`
	MatchKeyword           string                    `gorm:"type:varchar(255);not null"`
	AffectedTransactionIDs pq.StringArray            `gorm:"type:uuid[]"`
	Status                 string                    `gorm:"type:varchar(20);not null;default:'pending';index"`
	PreviousSuggestion     *string                   `gorm:"type:jsonb"`
	RetryReason            *string                   `gorm:"type:text"`
	CreatedAt              time.Time                 `gorm:"not null"`
	UpdatedAt              time.Time                 `gorm:"not null"`
	DeletedAt              gorm.DeletedAt            `gorm:"index"`

	// Relationships (not loaded by default, use Preload)
	Transaction *TransactionModel `gorm:"foreignKey:TransactionID;references:ID"`
	Category    *CategoryModel    `gorm:"foreignKey:SuggestedCategoryID;references:ID"`
}

// TableName returns the table name for the AISuggestionModel.
func (AISuggestionModel) TableName() string {
	return "ai_categorization_suggestions"
}

// ToEntity converts an AISuggestionModel to a domain AISuggestion entity.
func (m *AISuggestionModel) ToEntity() *entity.AISuggestion {
	suggestion := &entity.AISuggestion{
		ID:                  m.ID,
		UserID:              m.UserID,
		TransactionID:       m.TransactionID,
		SuggestedCategoryID: m.SuggestedCategoryID,
		MatchType:           entity.MatchType(m.MatchType),
		MatchKeyword:        m.MatchKeyword,
		Status:              entity.SuggestionStatus(m.Status),
		PreviousSuggestion:  m.PreviousSuggestion,
		RetryReason:         m.RetryReason,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}

	// Convert suggested category new
	if m.SuggestedCategoryNew != nil {
		suggestion.SuggestedCategoryNew = &entity.SuggestedCategoryNew{
			Name:  m.SuggestedCategoryNew.Name,
			Icon:  m.SuggestedCategoryNew.Icon,
			Color: m.SuggestedCategoryNew.Color,
		}
	}

	// Convert affected transaction IDs
	suggestion.AffectedTransactionIDs = make([]uuid.UUID, 0, len(m.AffectedTransactionIDs))
	for _, idStr := range m.AffectedTransactionIDs {
		if id, err := uuid.Parse(idStr); err == nil {
			suggestion.AffectedTransactionIDs = append(suggestion.AffectedTransactionIDs, id)
		}
	}

	return suggestion
}

// ToEntityWithDetails converts an AISuggestionModel with relationships to a domain entity with details.
func (m *AISuggestionModel) ToEntityWithDetails() *entity.AISuggestionWithDetails {
	result := &entity.AISuggestionWithDetails{
		Suggestion:              m.ToEntity(),
		AffectedTransactionCount: len(m.AffectedTransactionIDs),
	}

	if m.Transaction != nil {
		result.Transaction = m.Transaction.ToEntity()
	}

	if m.Category != nil {
		result.Category = m.Category.ToEntity()
	}

	return result
}

// AISuggestionFromEntity creates an AISuggestionModel from a domain AISuggestion entity.
func AISuggestionFromEntity(suggestion *entity.AISuggestion) *AISuggestionModel {
	model := &AISuggestionModel{
		ID:                  suggestion.ID,
		UserID:              suggestion.UserID,
		TransactionID:       suggestion.TransactionID,
		SuggestedCategoryID: suggestion.SuggestedCategoryID,
		MatchType:           string(suggestion.MatchType),
		MatchKeyword:        suggestion.MatchKeyword,
		Status:              string(suggestion.Status),
		PreviousSuggestion:  suggestion.PreviousSuggestion,
		RetryReason:         suggestion.RetryReason,
		CreatedAt:           suggestion.CreatedAt,
		UpdatedAt:           suggestion.UpdatedAt,
	}

	// Convert suggested category new
	if suggestion.SuggestedCategoryNew != nil {
		model.SuggestedCategoryNew = &SuggestedCategoryNewJSON{
			Name:  suggestion.SuggestedCategoryNew.Name,
			Icon:  suggestion.SuggestedCategoryNew.Icon,
			Color: suggestion.SuggestedCategoryNew.Color,
		}
	}

	// Convert affected transaction IDs
	model.AffectedTransactionIDs = make(pq.StringArray, len(suggestion.AffectedTransactionIDs))
	for i, id := range suggestion.AffectedTransactionIDs {
		model.AffectedTransactionIDs[i] = id.String()
	}

	return model
}
