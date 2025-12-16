// Package reconciliation contains credit card reconciliation use cases.
package reconciliation

import "time"

// parseDate creates a time.Time from year, month, and day.
func parseDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
