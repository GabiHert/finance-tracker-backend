// Package reconciliation contains credit card reconciliation use cases.
package reconciliation

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/valueobject"
)

// TriggerReconciliationInput represents the input for triggering reconciliation.
type TriggerReconciliationInput struct {
	UserID       uuid.UUID
	BillingCycle *string // Optional - if nil, reconcile all pending cycles
}

// AutoLinkedCycleOutput represents a cycle that was automatically linked.
type AutoLinkedCycleOutput struct {
	BillingCycle     string
	BillID           uuid.UUID
	BillDescription  string
	TransactionCount int
	Confidence       valueobject.Confidence
	AmountDifference decimal.Decimal
}

// PendingWithMatchesOutput represents a pending cycle with multiple matches.
type PendingWithMatchesOutput struct {
	BillingCycle   string
	PotentialBills []PotentialBillOutput
}

// NoMatchCycleOutput represents a cycle with no matching bills.
type NoMatchCycleOutput struct {
	BillingCycle     string
	TransactionCount int
	TotalAmount      decimal.Decimal
}

// TriggerReconciliationOutput represents the result of reconciliation.
type TriggerReconciliationOutput struct {
	AutoLinked        []AutoLinkedCycleOutput
	RequiresSelection []PendingWithMatchesOutput
	NoMatch           []NoMatchCycleOutput
	Summary           ReconciliationResultSummary
}

// ReconciliationResultSummary contains counts from reconciliation.
type ReconciliationResultSummary struct {
	AutoLinked        int
	RequiresSelection int
	NoMatch           int
}

// TriggerReconciliationUseCase handles triggering the reconciliation process.
type TriggerReconciliationUseCase struct {
	reconciliationRepo adapter.ReconciliationRepository
	config             valueobject.MatchingConfig
}

// NewTriggerReconciliationUseCase creates a new TriggerReconciliationUseCase instance.
func NewTriggerReconciliationUseCase(reconciliationRepo adapter.ReconciliationRepository) *TriggerReconciliationUseCase {
	return &TriggerReconciliationUseCase{
		reconciliationRepo: reconciliationRepo,
		config:             valueobject.DefaultMatchingConfig(),
	}
}

// Execute triggers the reconciliation process.
func (uc *TriggerReconciliationUseCase) Execute(ctx context.Context, input TriggerReconciliationInput) (*TriggerReconciliationOutput, error) {
	var cyclesToProcess []adapter.PendingCycleData

	if input.BillingCycle != nil {
		// Validate billing cycle format
		if !billingCycleRegex.MatchString(*input.BillingCycle) {
			return &TriggerReconciliationOutput{
				Summary: ReconciliationResultSummary{},
			}, nil
		}

		// Check if the cycle is already linked
		isLinked, _, err := uc.reconciliationRepo.IsCycleLinked(ctx, input.UserID, *input.BillingCycle)
		if err != nil {
			return nil, err
		}
		if isLinked {
			// Already linked, nothing to do
			return &TriggerReconciliationOutput{
				Summary: ReconciliationResultSummary{},
			}, nil
		}

		// Get the specific cycle
		cycles, err := uc.reconciliationRepo.GetPendingBillingCycles(ctx, input.UserID, 100, 0)
		if err != nil {
			return nil, err
		}

		for _, c := range cycles {
			if c.BillingCycle == *input.BillingCycle {
				cyclesToProcess = append(cyclesToProcess, c)
				break
			}
		}
	} else {
		// Get all pending cycles
		var err error
		cyclesToProcess, err = uc.reconciliationRepo.GetPendingBillingCycles(ctx, input.UserID, 100, 0)
		if err != nil {
			return nil, err
		}
	}

	// Process each cycle
	var autoLinked []AutoLinkedCycleOutput
	var requiresSelection []PendingWithMatchesOutput
	var noMatch []NoMatchCycleOutput

	for _, cycle := range cyclesToProcess {
		result := uc.processCycle(ctx, input.UserID, cycle)

		switch result.Type {
		case "auto_linked":
			autoLinked = append(autoLinked, result.AutoLinked)
		case "requires_selection":
			requiresSelection = append(requiresSelection, result.RequiresSelection)
		case "no_match":
			noMatch = append(noMatch, result.NoMatch)
		}
	}

	return &TriggerReconciliationOutput{
		AutoLinked:        autoLinked,
		RequiresSelection: requiresSelection,
		NoMatch:           noMatch,
		Summary: ReconciliationResultSummary{
			AutoLinked:        len(autoLinked),
			RequiresSelection: len(requiresSelection),
			NoMatch:           len(noMatch),
		},
	}, nil
}

// cycleProcessResult represents the result of processing a single cycle.
type cycleProcessResult struct {
	Type              string // "auto_linked", "requires_selection", "no_match"
	AutoLinked        AutoLinkedCycleOutput
	RequiresSelection PendingWithMatchesOutput
	NoMatch           NoMatchCycleOutput
}

// processCycle processes a single billing cycle for reconciliation.
func (uc *TriggerReconciliationUseCase) processCycle(
	ctx context.Context,
	userID uuid.UUID,
	cycle adapter.PendingCycleData,
) cycleProcessResult {
	ccTotal := decimal.NewFromInt(cycle.TotalAmount)

	// Calculate date range for bill matching
	dateRange := uc.calculateDateRange(cycle.BillingCycle)

	// Find potential bills
	potentialBills, err := uc.reconciliationRepo.FindPotentialBills(
		ctx, userID, cycle.BillingCycle, ccTotal, dateRange,
	)
	if err != nil || len(potentialBills) == 0 {
		return cycleProcessResult{
			Type: "no_match",
			NoMatch: NoMatchCycleOutput{
				BillingCycle:     cycle.BillingCycle,
				TransactionCount: cycle.TransactionCount,
				TotalAmount:      ccTotal,
			},
		}
	}

	// Score and filter potential bills
	scoredBills := uc.scorePotentialBills(ccTotal, potentialBills)
	if len(scoredBills) == 0 {
		return cycleProcessResult{
			Type: "no_match",
			NoMatch: NoMatchCycleOutput{
				BillingCycle:     cycle.BillingCycle,
				TransactionCount: cycle.TransactionCount,
				TotalAmount:      ccTotal,
			},
		}
	}

	// Check if we can auto-link
	if len(scoredBills) == 1 {
		bill := scoredBills[0]
		// Auto-link if high or medium confidence
		if bill.Confidence == valueobject.ConfidenceHigh || bill.Confidence == valueobject.ConfidenceMedium {
			// Perform the linking
			linkedCount, err := uc.reconciliationRepo.LinkCCTransactionsToBill(
				ctx, userID, cycle.BillingCycle, bill.BillID, bill.BillAmount,
			)
			if err != nil {
				// On error, treat as requires selection
				return cycleProcessResult{
					Type: "requires_selection",
					RequiresSelection: PendingWithMatchesOutput{
						BillingCycle:   cycle.BillingCycle,
						PotentialBills: scoredBills,
					},
				}
			}

			return cycleProcessResult{
				Type: "auto_linked",
				AutoLinked: AutoLinkedCycleOutput{
					BillingCycle:     cycle.BillingCycle,
					BillID:           bill.BillID,
					BillDescription:  bill.BillDescription,
					TransactionCount: linkedCount,
					Confidence:       bill.Confidence,
					AmountDifference: bill.AmountDifference,
				},
			}
		}
	}

	// Multiple matches or low confidence - requires selection
	return cycleProcessResult{
		Type: "requires_selection",
		RequiresSelection: PendingWithMatchesOutput{
			BillingCycle:   cycle.BillingCycle,
			PotentialBills: scoredBills,
		},
	}
}

// calculateDateRange calculates the date range for finding potential bill matches.
func (uc *TriggerReconciliationUseCase) calculateDateRange(billingCycle string) adapter.DateRange {
	year, month := parseBillingCycle(billingCycle)
	toleranceDays := uc.config.DateToleranceDays

	// Start: first day of billing cycle month minus tolerance
	startDate := parseDate(year, month, 1)
	startDate = startDate.AddDate(0, 0, -toleranceDays)

	// End: last day of next month plus tolerance
	nextMonth := month + 1
	nextYear := year
	if month == 12 {
		nextMonth = 1
		nextYear = year + 1
	}
	endDate := parseDate(nextYear, nextMonth+1, 0) // Last day of next month
	endDate = endDate.AddDate(0, 0, toleranceDays)

	return adapter.DateRange{
		Start: startDate,
		End:   endDate,
	}
}

// scorePotentialBills calculates confidence and scores for potential bill matches.
func (uc *TriggerReconciliationUseCase) scorePotentialBills(ccTotal decimal.Decimal, bills []adapter.BillData) []PotentialBillOutput {
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
