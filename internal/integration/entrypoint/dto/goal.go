// Package dto defines data transfer objects for API requests and responses.
package dto

import (
	"time"

	"github.com/finance-tracker/backend/internal/application/usecase/goal"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// CreateGoalRequest represents the request body for goal creation.
type CreateGoalRequest struct {
	CategoryID    string   `json:"category_id" binding:"required,uuid"`
	LimitAmount   float64  `json:"limit_amount" binding:"required,gt=0"`
	AlertOnExceed *bool    `json:"alert_on_exceed,omitempty"`
	Period        *string  `json:"period,omitempty" binding:"omitempty,oneof=monthly weekly yearly"`
}

// UpdateGoalRequest represents the request body for goal update.
type UpdateGoalRequest struct {
	LimitAmount   *float64 `json:"limit_amount,omitempty" binding:"omitempty,gt=0"`
	AlertOnExceed *bool    `json:"alert_on_exceed,omitempty"`
	Period        *string  `json:"period,omitempty" binding:"omitempty,oneof=monthly weekly yearly"`
}

// GoalResponse represents a single goal in API responses.
type GoalResponse struct {
	ID            string            `json:"id"`
	UserID        string            `json:"user_id"`
	CategoryID    string            `json:"category_id"`
	Category      *CategoryResponse `json:"category,omitempty"`
	LimitAmount   float64           `json:"limit_amount"`
	CurrentAmount float64           `json:"current_amount"`
	AlertOnExceed bool              `json:"alert_on_exceed"`
	Period        string            `json:"period"`
	StartDate     *string           `json:"start_date,omitempty"`
	EndDate       *string           `json:"end_date,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// GoalListResponse represents the response for listing goals.
type GoalListResponse struct {
	Goals []GoalResponse `json:"goals"`
}

// ToGoalResponse converts a domain Goal entity to a GoalResponse DTO.
func ToGoalResponse(g *entity.Goal) GoalResponse {
	response := GoalResponse{
		ID:            g.ID.String(),
		UserID:        g.UserID.String(),
		CategoryID:    g.CategoryID.String(),
		LimitAmount:   g.LimitAmount,
		CurrentAmount: 0,
		AlertOnExceed: g.AlertOnExceed,
		Period:        string(g.Period),
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}

	if g.StartDate != nil {
		dateStr := g.StartDate.Format("2006-01-02")
		response.StartDate = &dateStr
	}

	if g.EndDate != nil {
		dateStr := g.EndDate.Format("2006-01-02")
		response.EndDate = &dateStr
	}

	return response
}

// ToGoalResponseWithCategory converts a GoalOutput to a GoalResponse DTO with category.
func ToGoalResponseWithCategory(output *goal.GoalOutput) GoalResponse {
	response := GoalResponse{
		ID:            output.ID.String(),
		UserID:        output.UserID.String(),
		CategoryID:    output.CategoryID.String(),
		LimitAmount:   output.LimitAmount,
		CurrentAmount: output.CurrentAmount,
		AlertOnExceed: output.AlertOnExceed,
		Period:        string(output.Period),
		CreatedAt:     output.CreatedAt,
		UpdatedAt:     output.UpdatedAt,
	}

	if output.StartDate != nil {
		dateStr := output.StartDate.Format("2006-01-02")
		response.StartDate = &dateStr
	}

	if output.EndDate != nil {
		dateStr := output.EndDate.Format("2006-01-02")
		response.EndDate = &dateStr
	}

	if output.Category != nil {
		catResponse := ToCategoryResponse(output.Category)
		response.Category = &catResponse
	}

	return response
}

// ToGoalListResponse converts a list of GoalOutput to GoalListResponse.
func ToGoalListResponse(outputs []*goal.GoalOutput) GoalListResponse {
	goals := make([]GoalResponse, len(outputs))
	for i, output := range outputs {
		goals[i] = ToGoalResponseWithCategory(output)
	}
	return GoalListResponse{
		Goals: goals,
	}
}
