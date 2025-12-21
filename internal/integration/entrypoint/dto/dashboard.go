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

// DataRangeResponse represents the response for data range API.
type DataRangeResponse struct {
	Data DataRangeData `json:"data"`
}

// DataRangeData represents the data section of data range response.
type DataRangeData struct {
	OldestDate        *string `json:"oldest_date"`
	NewestDate        *string `json:"newest_date"`
	TotalTransactions int     `json:"total_transactions"`
	HasData           bool    `json:"has_data"`
}

// ToDataRangeResponse converts a GetDataRangeOutput to DataRangeResponse DTO.
func ToDataRangeResponse(output *dashboard.GetDataRangeOutput) DataRangeResponse {
	var oldestDate, newestDate *string
	if output.OldestDate != nil {
		s := output.OldestDate.Format("2006-01-02")
		oldestDate = &s
	}
	if output.NewestDate != nil {
		s := output.NewestDate.Format("2006-01-02")
		newestDate = &s
	}

	return DataRangeResponse{
		Data: DataRangeData{
			OldestDate:        oldestDate,
			NewestDate:        newestDate,
			TotalTransactions: output.TotalTransactions,
			HasData:           output.HasData,
		},
	}
}

// TrendsResponse represents the response for trends API.
type TrendsResponse struct {
	Data TrendsData `json:"data"`
}

// TrendsData represents the data section of trends response.
type TrendsData struct {
	Period TrendsPeriodResponse `json:"period"`
	Trends []TrendPointResponse `json:"trends"`
}

// TrendsPeriodResponse represents the period information in trends response.
type TrendsPeriodResponse struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Granularity string `json:"granularity"`
}

// TrendPointResponse represents a single trend data point in the response.
type TrendPointResponse struct {
	Date             string  `json:"date"`
	PeriodLabel      string  `json:"period_label"`
	Income           float64 `json:"income"`
	Expenses         float64 `json:"expenses"`
	Balance          float64 `json:"balance"`
	TransactionCount int     `json:"transaction_count"`
}

// ToTrendsResponse converts a GetTrendsOutput to TrendsResponse DTO.
func ToTrendsResponse(output *dashboard.GetTrendsOutput) TrendsResponse {
	trends := make([]TrendPointResponse, len(output.Trends))
	for i, t := range output.Trends {
		income, _ := t.Income.Float64()
		expenses, _ := t.Expenses.Float64()
		balance, _ := t.Balance.Float64()
		trends[i] = TrendPointResponse{
			Date:             t.Date.Format("2006-01-02"),
			PeriodLabel:      t.PeriodLabel,
			Income:           income,
			Expenses:         expenses,
			Balance:          balance,
			TransactionCount: t.TransactionCount,
		}
	}

	return TrendsResponse{
		Data: TrendsData{
			Period: TrendsPeriodResponse{
				StartDate:   output.Period.StartDate.Format("2006-01-02"),
				EndDate:     output.Period.EndDate.Format("2006-01-02"),
				Granularity: string(output.Period.Granularity),
			},
			Trends: trends,
		},
	}
}

// CategoryBreakdownResponse represents the response for category breakdown API.
type CategoryBreakdownResponse struct {
	Data CategoryBreakdownData `json:"data"`
}

// CategoryBreakdownData represents the data section of category breakdown response.
type CategoryBreakdownData struct {
	Period        BreakdownPeriodResponse       `json:"period"`
	TotalExpenses float64                       `json:"total_expenses"`
	Categories    []CategoryBreakdownItemResponse `json:"categories"`
}

// BreakdownPeriodResponse represents the period information in category breakdown response.
type BreakdownPeriodResponse struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	PeriodLabel string `json:"period_label"`
}

// CategoryBreakdownItemResponse represents a single category in the breakdown response.
type CategoryBreakdownItemResponse struct {
	CategoryID       string  `json:"category_id"`
	CategoryName     string  `json:"category_name"`
	CategoryColor    string  `json:"category_color"`
	CategoryIcon     string  `json:"category_icon"`
	Amount           float64 `json:"amount"`
	Percentage       float64 `json:"percentage"`
	TransactionCount int     `json:"transaction_count"`
}

// ToCategoryBreakdownResponse converts a GetCategoryBreakdownOutput to CategoryBreakdownResponse DTO.
func ToCategoryBreakdownResponse(output *dashboard.GetCategoryBreakdownOutput) CategoryBreakdownResponse {
	totalExpenses, _ := output.TotalExpenses.Float64()

	categories := make([]CategoryBreakdownItemResponse, len(output.Categories))
	for i, c := range output.Categories {
		amount, _ := c.Amount.Float64()
		categories[i] = CategoryBreakdownItemResponse{
			CategoryID:       c.CategoryID,
			CategoryName:     c.CategoryName,
			CategoryColor:    c.CategoryColor,
			CategoryIcon:     c.CategoryIcon,
			Amount:           amount,
			Percentage:       c.Percentage,
			TransactionCount: c.TransactionCount,
		}
	}

	return CategoryBreakdownResponse{
		Data: CategoryBreakdownData{
			Period: BreakdownPeriodResponse{
				StartDate:   output.Period.StartDate.Format("2006-01-02"),
				EndDate:     output.Period.EndDate.Format("2006-01-02"),
				PeriodLabel: output.Period.PeriodLabel,
			},
			TotalExpenses: totalExpenses,
			Categories:    categories,
		},
	}
}

// PeriodTransactionsResponse represents the response for period transactions API.
type PeriodTransactionsResponse struct {
	Data PeriodTransactionsData `json:"data"`
}

// PeriodTransactionsData represents the data section of period transactions response.
type PeriodTransactionsData struct {
	Period       TransactionsPeriodResponse      `json:"period"`
	Summary      TransactionSummaryResponse      `json:"summary"`
	Transactions []PeriodTransactionItemResponse `json:"transactions"`
	Pagination   DashboardPaginationResponse     `json:"pagination"`
}

// TransactionsPeriodResponse represents the period information in period transactions response.
type TransactionsPeriodResponse struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	PeriodLabel string `json:"period_label"`
}

// TransactionSummaryResponse represents summary totals for transactions.
type TransactionSummaryResponse struct {
	TotalIncome      float64 `json:"total_income"`
	TotalExpenses    float64 `json:"total_expenses"`
	Balance          float64 `json:"balance"`
	TransactionCount int     `json:"transaction_count"`
}

// PeriodTransactionItemResponse represents a transaction within a period.
type PeriodTransactionItemResponse struct {
	ID            string  `json:"id"`
	Description   string  `json:"description"`
	Amount        float64 `json:"amount"`
	Date          string  `json:"date"`
	CategoryID    *string `json:"category_id,omitempty"`
	CategoryName  *string `json:"category_name,omitempty"`
	CategoryColor *string `json:"category_color,omitempty"`
	CategoryIcon  *string `json:"category_icon,omitempty"`
}

// DashboardPaginationResponse represents pagination information for dashboard endpoints.
type DashboardPaginationResponse struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"has_more"`
}

// ToPeriodTransactionsResponse converts a GetPeriodTransactionsOutput to PeriodTransactionsResponse DTO.
func ToPeriodTransactionsResponse(output *dashboard.GetPeriodTransactionsOutput) PeriodTransactionsResponse {
	totalIncome, _ := output.Summary.TotalIncome.Float64()
	totalExpenses, _ := output.Summary.TotalExpenses.Float64()
	balance, _ := output.Summary.Balance.Float64()

	transactions := make([]PeriodTransactionItemResponse, len(output.Transactions))
	for i, t := range output.Transactions {
		amount, _ := t.Amount.Float64()
		transactions[i] = PeriodTransactionItemResponse{
			ID:            t.ID,
			Description:   t.Description,
			Amount:        amount,
			Date:          t.Date.Format("2006-01-02"),
			CategoryID:    t.CategoryID,
			CategoryName:  t.CategoryName,
			CategoryColor: t.CategoryColor,
			CategoryIcon:  t.CategoryIcon,
		}
	}

	return PeriodTransactionsResponse{
		Data: PeriodTransactionsData{
			Period: TransactionsPeriodResponse{
				StartDate:   output.Period.StartDate.Format("2006-01-02"),
				EndDate:     output.Period.EndDate.Format("2006-01-02"),
				PeriodLabel: output.Period.PeriodLabel,
			},
			Summary: TransactionSummaryResponse{
				TotalIncome:      totalIncome,
				TotalExpenses:    totalExpenses,
				Balance:          balance,
				TransactionCount: output.Summary.TransactionCount,
			},
			Transactions: transactions,
			Pagination: DashboardPaginationResponse{
				Total:   output.Pagination.Total,
				Limit:   output.Pagination.Limit,
				Offset:  output.Pagination.Offset,
				HasMore: output.Pagination.HasMore,
			},
		},
	}
}
