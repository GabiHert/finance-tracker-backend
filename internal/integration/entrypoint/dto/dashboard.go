// Package dto defines data transfer objects for API requests and responses.
package dto

import (
	"github.com/finance-tracker/backend/internal/application/usecase/dashboard"
)

// CategoryTrendsResponse represents the response for category trends API.
type CategoryTrendsResponse struct {
	Data CategoryTrendsData `json:"data"`
}

// CategoryTrendsData represents the data section of category trends response.
type CategoryTrendsData struct {
	Period     TrendPeriodResponse     `json:"period"`
	Categories []CategoryInfoResponse  `json:"categories"`
	Trends     []TrendDataPointResponse `json:"trends"`
}

// TrendPeriodResponse represents the period information in the response.
type TrendPeriodResponse struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Granularity string `json:"granularity"`
}

// CategoryInfoResponse represents category information in the response.
type CategoryInfoResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Color       string  `json:"color"`
	TotalAmount float64 `json:"total_amount"`
	IsOthers    bool    `json:"is_others"`
}

// TrendDataPointResponse represents a single trend data point in the response.
type TrendDataPointResponse struct {
	Date        string                   `json:"date"`
	PeriodLabel string                   `json:"period_label"`
	Amounts     []CategoryAmountResponse `json:"amounts"`
}

// CategoryAmountResponse represents the amount for a category in a period.
type CategoryAmountResponse struct {
	CategoryID string  `json:"category_id"`
	Amount     float64 `json:"amount"`
}

// ToCategoryTrendsResponse converts a GetCategoryTrendsOutput to CategoryTrendsResponse DTO.
func ToCategoryTrendsResponse(output *dashboard.GetCategoryTrendsOutput) CategoryTrendsResponse {
	// Convert categories
	categories := make([]CategoryInfoResponse, len(output.Categories))
	for i, cat := range output.Categories {
		totalAmount, _ := cat.TotalAmount.Float64()
		categories[i] = CategoryInfoResponse{
			ID:          cat.ID.String(),
			Name:        cat.Name,
			Color:       cat.Color,
			TotalAmount: totalAmount,
			IsOthers:    cat.IsOthers,
		}
	}

	// Convert trends
	trends := make([]TrendDataPointResponse, len(output.Trends))
	for i, trend := range output.Trends {
		amounts := make([]CategoryAmountResponse, len(trend.Amounts))
		for j, amt := range trend.Amounts {
			amount, _ := amt.Amount.Float64()
			amounts[j] = CategoryAmountResponse{
				CategoryID: amt.CategoryID.String(),
				Amount:     amount,
			}
		}
		trends[i] = TrendDataPointResponse{
			Date:        trend.Date.Format("2006-01-02"),
			PeriodLabel: trend.PeriodLabel,
			Amounts:     amounts,
		}
	}

	return CategoryTrendsResponse{
		Data: CategoryTrendsData{
			Period: TrendPeriodResponse{
				StartDate:   output.Period.StartDate.Format("2006-01-02"),
				EndDate:     output.Period.EndDate.Format("2006-01-02"),
				Granularity: string(output.Period.Granularity),
			},
			Categories: categories,
			Trends:     trends,
		},
	}
}
