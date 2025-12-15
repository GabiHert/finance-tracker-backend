// Package valueobject contains domain value objects for the Finance Tracker system.
package valueobject

import "github.com/shopspring/decimal"

// MatchingConfig contains the configuration for bill-to-CC matching.
type MatchingConfig struct {
	// Amount tolerance: whichever is greater
	AmountTolerancePercent  decimal.Decimal // 0.02 = 2%
	AmountToleranceAbsolute int64           // 2000 = R$ 20.00 in cents

	// Date tolerance
	DateToleranceDays int // 15 days

	// Confidence thresholds
	HighConfidencePercent  decimal.Decimal // 0.005 = 0.5%
	HighConfidenceAbsolute int64           // 500 = R$ 5.00
	MedConfidencePercent   decimal.Decimal // 0.02 = 2%
	MedConfidenceAbsolute  int64           // 2000 = R$ 20.00
}

// DefaultMatchingConfig returns the default matching configuration.
func DefaultMatchingConfig() MatchingConfig {
	return MatchingConfig{
		AmountTolerancePercent:  decimal.NewFromFloat(0.02),
		AmountToleranceAbsolute: 2000,
		DateToleranceDays:       15,
		HighConfidencePercent:   decimal.NewFromFloat(0.005),
		HighConfidenceAbsolute:  500,
		MedConfidencePercent:    decimal.NewFromFloat(0.02),
		MedConfidenceAbsolute:   2000,
	}
}

// IsWithinTolerance checks if the amount difference is within acceptable tolerance.
func (c MatchingConfig) IsWithinTolerance(ccTotal, billAmount decimal.Decimal) bool {
	diff := ccTotal.Sub(billAmount).Abs()

	// Check absolute tolerance first (for small amounts)
	absoluteTolerance := decimal.NewFromInt(c.AmountToleranceAbsolute)
	if diff.LessThanOrEqual(absoluteTolerance) {
		return true
	}

	// Check percentage tolerance
	if billAmount.IsZero() {
		return false
	}
	percentDiff := diff.Div(billAmount.Abs())
	return percentDiff.LessThanOrEqual(c.AmountTolerancePercent)
}
