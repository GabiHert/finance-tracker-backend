// Package dashboard contains dashboard-related use cases.
package dashboard

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// Granularity represents the time granularity for trends data.
type Granularity string

const (
	GranularityDaily   Granularity = "daily"
	GranularityWeekly  Granularity = "weekly"
	GranularityMonthly Granularity = "monthly"
)

// OthersCategoryID is a constant UUID used to represent the "Others" category.
var OthersCategoryID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

// OthersCategoryColor is the default color for the "Others" category.
const OthersCategoryColor = "#9CA3AF"

// GetCategoryTrendsInput represents the input for getting category trends.
type GetCategoryTrendsInput struct {
	UserID        uuid.UUID
	StartDate     time.Time
	EndDate       time.Time
	Granularity   Granularity
	TopCategories int
}

// GetCategoryTrendsOutput represents the output of getting category trends.
type GetCategoryTrendsOutput struct {
	Period     TrendPeriod
	Categories []CategoryInfo
	Trends     []TrendDataPoint
}

// TrendPeriod represents the time period for the trends data.
type TrendPeriod struct {
	StartDate   time.Time
	EndDate     time.Time
	Granularity Granularity
}

// CategoryInfo represents category information with total amount.
type CategoryInfo struct {
	ID          uuid.UUID
	Name        string
	Color       string
	TotalAmount decimal.Decimal
	IsOthers    bool
}

// TrendDataPoint represents a single trend data point with amounts per category.
type TrendDataPoint struct {
	Date        time.Time
	PeriodLabel string
	Amounts     []CategoryAmount
}

// CategoryAmount represents the amount for a specific category in a period.
type CategoryAmount struct {
	CategoryID uuid.UUID
	Amount     decimal.Decimal
}

// monthAbbreviations maps months to Portuguese abbreviations.
var monthAbbreviations = map[time.Month]string{
	time.January:   "Jan",
	time.February:  "Fev",
	time.March:     "Mar",
	time.April:     "Abr",
	time.May:       "Mai",
	time.June:      "Jun",
	time.July:      "Jul",
	time.August:    "Ago",
	time.September: "Set",
	time.October:   "Out",
	time.November:  "Nov",
	time.December:  "Dez",
}

// GetCategoryTrendsUseCase handles getting category expense trends.
type GetCategoryTrendsUseCase struct {
	transactionRepo adapter.TransactionRepository
}

// NewGetCategoryTrendsUseCase creates a new GetCategoryTrendsUseCase instance.
func NewGetCategoryTrendsUseCase(
	transactionRepo adapter.TransactionRepository,
) *GetCategoryTrendsUseCase {
	return &GetCategoryTrendsUseCase{
		transactionRepo: transactionRepo,
	}
}

// Execute performs the category trends retrieval.
func (uc *GetCategoryTrendsUseCase) Execute(
	ctx context.Context,
	input GetCategoryTrendsInput,
) (*GetCategoryTrendsOutput, error) {
	// 1. Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// 2. Get expense transactions in date range
	expenses, err := uc.transactionRepo.GetExpensesByDateRange(
		ctx, input.UserID, input.StartDate, input.EndDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	// 3. If no expenses, return empty result
	if len(expenses) == 0 {
		return &GetCategoryTrendsOutput{
			Period: TrendPeriod{
				StartDate:   input.StartDate,
				EndDate:     input.EndDate,
				Granularity: input.Granularity,
			},
			Categories: []CategoryInfo{},
			Trends:     []TrendDataPoint{},
		}, nil
	}

	// 4. Calculate totals per category
	categoryTotals := make(map[uuid.UUID]decimal.Decimal)
	categoryInfo := make(map[uuid.UUID]struct {
		Name  string
		Color string
	})

	for _, exp := range expenses {
		current := categoryTotals[exp.CategoryID]
		categoryTotals[exp.CategoryID] = current.Add(exp.Amount)
		categoryInfo[exp.CategoryID] = struct {
			Name  string
			Color string
		}{
			Name:  exp.CategoryName,
			Color: exp.CategoryColor,
		}
	}

	// 5. Determine top N categories and "Others"
	topCategoryIDs, othersTotal := uc.selectTopCategories(categoryTotals, input.TopCategories)

	// 6. Build category infos list
	categories := uc.buildCategoryInfos(topCategoryIDs, categoryTotals, categoryInfo, othersTotal)

	// 7. Aggregate by time period
	trends := uc.aggregateByPeriod(expenses, topCategoryIDs, input.Granularity, input.StartDate, input.EndDate)

	return &GetCategoryTrendsOutput{
		Period: TrendPeriod{
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			Granularity: input.Granularity,
		},
		Categories: categories,
		Trends:     trends,
	}, nil
}

// validateInput validates the input parameters.
func (uc *GetCategoryTrendsUseCase) validateInput(input GetCategoryTrendsInput) error {
	if input.StartDate.IsZero() {
		return domainerror.NewDashboardError(
			domainerror.ErrCodeMissingStartDate,
			"start_date is required",
			domainerror.ErrMissingStartDate,
		)
	}

	if input.EndDate.IsZero() {
		return domainerror.NewDashboardError(
			domainerror.ErrCodeMissingEndDate,
			"end_date is required",
			domainerror.ErrMissingEndDate,
		)
	}

	if input.EndDate.Before(input.StartDate) {
		return domainerror.NewDashboardError(
			domainerror.ErrCodeInvalidDateRange,
			"end_date must be after start_date",
			domainerror.ErrInvalidDateRange,
		)
	}

	if input.Granularity != GranularityDaily &&
		input.Granularity != GranularityWeekly &&
		input.Granularity != GranularityMonthly {
		return domainerror.NewDashboardError(
			domainerror.ErrCodeInvalidGranularity,
			"granularity must be: daily, weekly, or monthly",
			domainerror.ErrInvalidGranularity,
		)
	}

	return nil
}

// selectTopCategories returns top N category IDs and the "others" total.
func (uc *GetCategoryTrendsUseCase) selectTopCategories(
	totals map[uuid.UUID]decimal.Decimal,
	topN int,
) ([]uuid.UUID, decimal.Decimal) {
	// Sort by total descending
	type catTotal struct {
		ID    uuid.UUID
		Total decimal.Decimal
	}
	var sorted []catTotal
	for id, total := range totals {
		sorted = append(sorted, catTotal{ID: id, Total: total})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Total.GreaterThan(sorted[j].Total)
	})

	// Select top N
	topIDs := make([]uuid.UUID, 0, topN)
	othersTotal := decimal.Zero
	for i, ct := range sorted {
		if i < topN {
			topIDs = append(topIDs, ct.ID)
		} else {
			othersTotal = othersTotal.Add(ct.Total)
		}
	}

	return topIDs, othersTotal
}

// buildCategoryInfos builds the category info list including "Others" if needed.
func (uc *GetCategoryTrendsUseCase) buildCategoryInfos(
	topCategoryIDs []uuid.UUID,
	categoryTotals map[uuid.UUID]decimal.Decimal,
	categoryInfo map[uuid.UUID]struct {
		Name  string
		Color string
	},
	othersTotal decimal.Decimal,
) []CategoryInfo {
	categories := make([]CategoryInfo, 0, len(topCategoryIDs)+1)

	for _, catID := range topCategoryIDs {
		info := categoryInfo[catID]
		categories = append(categories, CategoryInfo{
			ID:          catID,
			Name:        info.Name,
			Color:       info.Color,
			TotalAmount: categoryTotals[catID],
			IsOthers:    false,
		})
	}

	// Add "Others" category if there's a remaining total
	if othersTotal.GreaterThan(decimal.Zero) {
		categories = append(categories, CategoryInfo{
			ID:          OthersCategoryID,
			Name:        "Outros",
			Color:       OthersCategoryColor,
			TotalAmount: othersTotal,
			IsOthers:    true,
		})
	}

	return categories
}

// aggregateByPeriod groups expenses by time period.
func (uc *GetCategoryTrendsUseCase) aggregateByPeriod(
	expenses []*entity.ExpenseWithCategory,
	topCategoryIDs []uuid.UUID,
	granularity Granularity,
	startDate, endDate time.Time,
) []TrendDataPoint {
	// Create a set of top category IDs for quick lookup
	topSet := make(map[uuid.UUID]bool)
	for _, id := range topCategoryIDs {
		topSet[id] = true
	}

	// Generate all periods in the range
	periods := uc.generatePeriods(startDate, endDate, granularity)

	// Initialize aggregation map: periodKey -> categoryID -> amount
	aggregation := make(map[string]map[uuid.UUID]decimal.Decimal)
	for _, p := range periods {
		periodKey := p.Date.Format("2006-01-02")
		aggregation[periodKey] = make(map[uuid.UUID]decimal.Decimal)
		// Initialize all categories with zero
		for _, catID := range topCategoryIDs {
			aggregation[periodKey][catID] = decimal.Zero
		}
		aggregation[periodKey][OthersCategoryID] = decimal.Zero
	}

	// Aggregate expenses into periods
	for _, exp := range expenses {
		periodKey := uc.getPeriodKey(exp.Date, granularity)
		if _, exists := aggregation[periodKey]; !exists {
			continue // Skip if outside the generated periods
		}

		catID := exp.CategoryID
		if !topSet[catID] {
			catID = OthersCategoryID
		}

		current := aggregation[periodKey][catID]
		aggregation[periodKey][catID] = current.Add(exp.Amount)
	}

	// Build trend data points
	trends := make([]TrendDataPoint, 0, len(periods))
	for _, p := range periods {
		periodKey := p.Date.Format("2006-01-02")
		amounts := make([]CategoryAmount, 0, len(topCategoryIDs)+1)

		// Add amounts for top categories
		for _, catID := range topCategoryIDs {
			amounts = append(amounts, CategoryAmount{
				CategoryID: catID,
				Amount:     aggregation[periodKey][catID],
			})
		}

		// Add "Others" amount if > 0
		if aggregation[periodKey][OthersCategoryID].GreaterThan(decimal.Zero) {
			amounts = append(amounts, CategoryAmount{
				CategoryID: OthersCategoryID,
				Amount:     aggregation[periodKey][OthersCategoryID],
			})
		}

		trends = append(trends, TrendDataPoint{
			Date:        p.Date,
			PeriodLabel: p.PeriodLabel,
			Amounts:     amounts,
		})
	}

	return trends
}

// periodInfo holds period information for aggregation.
type periodInfo struct {
	Date        time.Time
	PeriodLabel string
}

// generatePeriods generates all periods in the date range.
func (uc *GetCategoryTrendsUseCase) generatePeriods(
	startDate, endDate time.Time,
	granularity Granularity,
) []periodInfo {
	var periods []periodInfo
	current := startDate

	switch granularity {
	case GranularityDaily:
		for !current.After(endDate) {
			periods = append(periods, periodInfo{
				Date:        current,
				PeriodLabel: uc.getDailyLabel(current),
			})
			current = current.AddDate(0, 0, 1)
		}

	case GranularityWeekly:
		// Start from the Monday of the week containing startDate
		current = uc.getWeekStart(startDate)
		for !current.After(endDate) {
			weekEnd := current.AddDate(0, 0, 6)
			if weekEnd.After(endDate) {
				weekEnd = endDate
			}
			periods = append(periods, periodInfo{
				Date:        current,
				PeriodLabel: uc.getWeeklyLabel(current, weekEnd),
			})
			current = current.AddDate(0, 0, 7)
		}

	case GranularityMonthly:
		// Start from the first of the month containing startDate
		current = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
		for !current.After(endDate) {
			periods = append(periods, periodInfo{
				Date:        current,
				PeriodLabel: uc.getMonthlyLabel(current),
			})
			current = current.AddDate(0, 1, 0)
		}
	}

	return periods
}

// getPeriodKey returns the period key for a date based on granularity.
func (uc *GetCategoryTrendsUseCase) getPeriodKey(date time.Time, granularity Granularity) string {
	switch granularity {
	case GranularityDaily:
		return date.Format("2006-01-02")
	case GranularityWeekly:
		return uc.getWeekStart(date).Format("2006-01-02")
	case GranularityMonthly:
		return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).Format("2006-01-02")
	}
	return date.Format("2006-01-02")
}

// getWeekStart returns the Monday of the week containing the given date.
func (uc *GetCategoryTrendsUseCase) getWeekStart(date time.Time) time.Time {
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 7
	}
	daysFromMonday := weekday - 1
	return time.Date(date.Year(), date.Month(), date.Day()-daysFromMonday, 0, 0, 0, 0, date.Location())
}

// getDailyLabel returns the label for a daily period (e.g., "15 Nov").
func (uc *GetCategoryTrendsUseCase) getDailyLabel(date time.Time) string {
	return fmt.Sprintf("%d %s", date.Day(), monthAbbreviations[date.Month()])
}

// getWeeklyLabel returns the label for a weekly period (e.g., "1-7 Nov" or "29 Nov - 5 Dez").
func (uc *GetCategoryTrendsUseCase) getWeeklyLabel(weekStart, weekEnd time.Time) string {
	if weekStart.Month() == weekEnd.Month() {
		return fmt.Sprintf("%d-%d %s", weekStart.Day(), weekEnd.Day(), monthAbbreviations[weekStart.Month()])
	}
	return fmt.Sprintf("%d %s - %d %s",
		weekStart.Day(), monthAbbreviations[weekStart.Month()],
		weekEnd.Day(), monthAbbreviations[weekEnd.Month()])
}

// getMonthlyLabel returns the label for a monthly period (e.g., "Nov").
func (uc *GetCategoryTrendsUseCase) getMonthlyLabel(date time.Time) string {
	return monthAbbreviations[date.Month()]
}
