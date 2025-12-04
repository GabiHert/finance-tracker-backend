// Package model defines database models for persistence layer.
package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// TransactionModel represents the transactions table in the database.
type TransactionModel struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID       `gorm:"type:uuid;not null;index"`
	Date        time.Time       `gorm:"type:date;not null;index"`
	Description string          `gorm:"type:varchar(255);not null"`
	Amount      decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	Type        string          `gorm:"type:varchar(10);not null;index"`
	CategoryID  *uuid.UUID      `gorm:"type:uuid;index"`
	Notes       string          `gorm:"type:text"`
	IsRecurring bool            `gorm:"default:false"`
	UploadedAt  *time.Time      `gorm:"type:timestamp"`
	CreatedAt   time.Time       `gorm:"not null"`
	UpdatedAt   time.Time       `gorm:"not null"`
	DeletedAt   gorm.DeletedAt  `gorm:"index"` // Soft-delete support

	// Credit card import fields
	CreditCardPaymentID *uuid.UUID      `gorm:"type:uuid;index"`
	BillingCycle        string          `gorm:"type:varchar(7)"`
	OriginalAmount      *decimal.Decimal `gorm:"type:decimal(12,2)"`
	IsCreditCardPayment bool            `gorm:"default:false"`
	ExpandedAt          *time.Time      `gorm:"type:timestamp"`
	InstallmentCurrent  *int            `gorm:"type:integer"`
	InstallmentTotal    *int            `gorm:"type:integer"`
	IsHidden            bool            `gorm:"default:false"`

	// Relationships (not loaded by default, use Preload)
	Category          *CategoryModel     `gorm:"foreignKey:CategoryID;references:ID"`
	User              *UserModel         `gorm:"foreignKey:UserID;references:ID"`
	CreditCardPayment *TransactionModel  `gorm:"foreignKey:CreditCardPaymentID;references:ID"`
}

// TableName returns the table name for the TransactionModel.
func (TransactionModel) TableName() string {
	return "transactions"
}

// ToEntity converts a TransactionModel to a domain Transaction entity.
func (m *TransactionModel) ToEntity() *entity.Transaction {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &entity.Transaction{
		ID:          m.ID,
		UserID:      m.UserID,
		Date:        m.Date,
		Description: m.Description,
		Amount:      m.Amount,
		Type:        entity.TransactionType(m.Type),
		CategoryID:  m.CategoryID,
		Notes:       m.Notes,
		IsRecurring: m.IsRecurring,
		UploadedAt:  m.UploadedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   deletedAt,
		// Credit card fields
		CreditCardPaymentID: m.CreditCardPaymentID,
		BillingCycle:        m.BillingCycle,
		OriginalAmount:      m.OriginalAmount,
		IsCreditCardPayment: m.IsCreditCardPayment,
		ExpandedAt:          m.ExpandedAt,
		InstallmentCurrent:  m.InstallmentCurrent,
		InstallmentTotal:    m.InstallmentTotal,
		IsHidden:            m.IsHidden,
	}
}

// ToEntityWithCategory converts a TransactionModel with its Category to a TransactionWithCategory entity.
func (m *TransactionModel) ToEntityWithCategory() *entity.TransactionWithCategory {
	result := &entity.TransactionWithCategory{
		Transaction: m.ToEntity(),
	}

	if m.Category != nil {
		result.Category = m.Category.ToEntity()
	}

	return result
}

// TransactionFromEntity creates a TransactionModel from a domain Transaction entity.
func TransactionFromEntity(transaction *entity.Transaction) *TransactionModel {
	var deletedAt gorm.DeletedAt
	if transaction.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *transaction.DeletedAt, Valid: true}
	}

	return &TransactionModel{
		ID:          transaction.ID,
		UserID:      transaction.UserID,
		Date:        transaction.Date,
		Description: transaction.Description,
		Amount:      transaction.Amount,
		Type:        string(transaction.Type),
		CategoryID:  transaction.CategoryID,
		Notes:       transaction.Notes,
		IsRecurring: transaction.IsRecurring,
		UploadedAt:  transaction.UploadedAt,
		CreatedAt:   transaction.CreatedAt,
		UpdatedAt:   transaction.UpdatedAt,
		DeletedAt:   deletedAt,
		// Credit card fields
		CreditCardPaymentID: transaction.CreditCardPaymentID,
		BillingCycle:        transaction.BillingCycle,
		OriginalAmount:      transaction.OriginalAmount,
		IsCreditCardPayment: transaction.IsCreditCardPayment,
		ExpandedAt:          transaction.ExpandedAt,
		InstallmentCurrent:  transaction.InstallmentCurrent,
		InstallmentTotal:    transaction.InstallmentTotal,
		IsHidden:            transaction.IsHidden,
	}
}
