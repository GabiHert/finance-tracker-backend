// Package reconciliation contains credit card reconciliation use cases.
package reconciliation

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/valueobject"
)

// GetPendingInput represents the input for getting pending reconciliations.
type GetPendingInput struct {
	UserID uuid.UUID
	Limit  int
	Offset int
}

// PendingCycleOutput represents a pending billing cycle with potential matches.
type PendingCycleOutput struct {
	BillingCycle     string
	DisplayName      string
	TransactionCount int
	TotalAmount      decimal.Decimal
	OldestDate       time.Time
	NewestDate       time.Time
	PotentialBills   []PotentialBillOutput
}

// PotentialBillOutput represents a potential bill match.
type PotentialBillOutput struct {
	BillID                  uuid.UUID
	BillDate                time.Time
	BillDescription         string
	BillAmount              decimal.Decimal
	CategoryName            *string
	Confidence              valueobject.Confidence
	AmountDifference        decimal.Decimal
	AmountDifferencePercent decimal.Decimal
	Score                   float64
}

// GetPendingOutput represents the output for getting pending reconciliations.
type GetPendingOutput struct {
	PendingCycles []PendingCycleOutput
	Summary       ReconciliationSummaryOutput
}

// ReconciliationSummaryOutput contains summary statistics.
type ReconciliationSummaryOutput struct {
	TotalPending  int
	TotalLinked   int
	MonthsCovered int
}

// GetPendingUseCase handles getting pending reconciliations.
type GetPendingUseCase struct {
	reconciliationRepo adapter.ReconciliationRepository
	config             valueobject.MatchingConfig
}

// NewGetPendingUseCase creates a new GetPendingUseCase instance.
func NewGetPendingUseCase(reconciliationRepo adapter.ReconciliationRepository) *GetPendingUseCase {
	return &GetPendingUseCase{
		reconciliationRepo: reconciliationRepo,
		config:             valueobject.DefaultMatchingConfig(),
	}
}

// Execute retrieves pending billing cycles with potential matches.
func (uc *GetPendingUseCase) Execute(ctx context.Context, input GetPendingInput) (*GetPendingOutput, error) {
	// Set defaults
	limit := input.Limit
	if limit <= 0 {
		limit = 12 // Default to 12 months
	}

	// Get pending billing cycles
	pendingCycles, err := uc.reconciliationRepo.GetPendingBillingCycles(ctx, input.UserID, limit, input.Offset)
	if err != nil {
		return nil, err
	}

	// Get summary
	summary, err := uc.reconciliationRepo.GetReconciliationSummary(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Build output with potential matches for each cycle
	outputCycles := make([]PendingCycleOutput, 0, len(pendingCycles))

	for _, cycle := range pendingCycles {
		// Calculate date range for bill matching
		dateRange := uc.calculateDateRange(cycle.BillingCycle)

		// Find potential bills
		ccTotal := decimal.NewFromInt(cycle.TotalAmount)
		potentialBills, err := uc.reconciliationRepo.FindPotentialBills(
			ctx, input.UserID, cycle.BillingCycle, ccTotal, dateRange,
		)
		if err != nil {
			// Continue without potential bills on error
			potentialBills = nil
		}

		// Convert and score potential bills
		potentialBillOutputs := uc.scorePotentialBills(ccTotal, potentialBills)

		outputCycles = append(outputCycles, PendingCycleOutput{
			BillingCycle:     cycle.BillingCycle,
			DisplayName:      valueobject.FormatBillingCycleDisplay(cycle.BillingCycle),
			TransactionCount: cycle.TransactionCount,
			TotalAmount:      ccTotal,
			OldestDate:       cycle.OldestDate,
			NewestDate:       cycle.NewestDate,
			PotentialBills:   potentialBillOutputs,
		})
	}

	return &GetPendingOutput{
		PendingCycles: outputCycles,
		Summary: ReconciliationSummaryOutput{
			TotalPending:  summary.TotalPending,
			TotalLinked:   summary.TotalLinked,
			MonthsCovered: summary.MonthsCovered,
		},
	}, nil
}

// calculateDateRange calculates the date range for finding potential bill matches.
func (uc *GetPendingUseCase) calculateDateRange(billingCycle string) adapter.DateRange {
	year, month := parseBillingCycle(billingCycle)
	toleranceDays := uc.config.DateToleranceDays

	// Start: first day of billing cycle month minus tolerance
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	startDate = startDate.AddDate(0, 0, -toleranceDays)

	// End: last day of next month plus tolerance (bills usually come in the following month)
	nextMonth := time.Month(month + 1)
	nextYear := year
	if month == 12 {
		nextMonth = 1
		nextYear = year + 1
	}
	endDate := time.Date(nextYear, nextMonth+1, 0, 23, 59, 59, 0, time.UTC)
	endDate = endDate.AddDate(0, 0, toleranceDays)

	return adapter.DateRange{
		Start: startDate,
		End:   endDate,
	}
}

// scorePotentialBills calculates confidence and scores for potential bill matches.
func (uc *GetPendingUseCase) scorePotentialBills(ccTotal decimal.Decimal, bills []adapter.BillData) []PotentialBillOutput {
	if len(bills) == 0 {
		return nil
	}

	results := make([]PotentialBillOutput, 0, len(bills))

	for _, bill := range bills {
		billAmount := decimal.NewFromInt(bill.Amount)
		diff := ccTotal.Sub(billAmount)
		diffAbs := diff.Abs()

		// Calculate percentage difference
		var percentDiff decimal.Decimal
		if !billAmount.IsZero() {
			percentDiff = diffAbs.Div(billAmount.Abs())
		}

		// Check if within tolerance
		if !uc.config.IsWithinTolerance(ccTotal, billAmount) {
			continue
		}

		// Calculate confidence
		confidence := valueobject.CalculateConfidence(uc.config, ccTotal, billAmount)

		// Calculate score (higher is better, 0-1 range)
		// Closer amounts get higher scores
		percentDiffFloat, _ := percentDiff.Float64()
		score := 1.0 - percentDiffFloat

		results = append(results, PotentialBillOutput{
			BillID:                  bill.ID,
			BillDate:                bill.Date,
			BillDescription:         bill.Description,
			BillAmount:              billAmount,
			CategoryName:            bill.CategoryName,
			Confidence:              confidence,
			AmountDifference:        diff,
			AmountDifferencePercent: percentDiff,
			Score:                   score,
		})
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// parseBillingCycle parses a billing cycle string (YYYY-MM) into year and month.
func parseBillingCycle(billingCycle string) (int, int) {
	t, _ := time.Parse("2006-01", billingCycle)
	return t.Year(), int(t.Month())
}
