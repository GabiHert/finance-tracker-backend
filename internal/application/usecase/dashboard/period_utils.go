// Package dashboard contains dashboard-related use cases.
package dashboard

import (
	"fmt"
	"time"
)

// GranularityQuarterly represents quarterly granularity for interactive charts.
const GranularityQuarterly Granularity = "quarterly"

// GeneratePeriodLabel generates a human-readable label for a period based on granularity.
// Formats:
// - Weekly: "S{week} {year}" (e.g., "S12 2025")
// - Monthly: "{month_abbr} {year}" (e.g., "Mar 2025")
// - Quarterly: "T{quarter} {year}" (e.g., "T1 2025")
func GeneratePeriodLabel(date time.Time, granularity Granularity) string {
	switch granularity {
	case GranularityWeekly:
		_, week := date.ISOWeek()
		return fmt.Sprintf("S%d %d", week, date.Year())
	case GranularityMonthly:
		return fmt.Sprintf("%s %d", monthAbbreviations[date.Month()], date.Year())
	case GranularityQuarterly:
		quarter := (int(date.Month())-1)/3 + 1
		return fmt.Sprintf("T%d %d", quarter, date.Year())
	default:
		return date.Format("02/01/2006")
	}
}

// GetPeriodBounds returns the start and end dates for a period containing the given date.
func GetPeriodBounds(date time.Time, granularity Granularity) (start, end time.Time) {
	loc := date.Location()

	switch granularity {
	case GranularityWeekly:
		// Week starts on Monday
		weekday := int(date.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = time.Date(date.Year(), date.Month(), date.Day()-(weekday-1), 0, 0, 0, 0, loc)
		end = start.AddDate(0, 0, 6)
	case GranularityMonthly:
		start = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, loc)
		end = start.AddDate(0, 1, -1)
	case GranularityQuarterly:
		quarter := (int(date.Month()) - 1) / 3
		start = time.Date(date.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, loc)
		end = start.AddDate(0, 3, -1)
	default:
		start = date
		end = date
	}
	return start.Truncate(24 * time.Hour), end.Truncate(24 * time.Hour)
}

// GeneratePeriodSeries generates all periods between startDate and endDate for the given granularity.
// This ensures continuous data for chart rendering with no gaps.
func GeneratePeriodSeries(startDate, endDate time.Time, granularity Granularity) []PeriodInfo {
	var periods []PeriodInfo
	loc := startDate.Location()

	switch granularity {
	case GranularityWeekly:
		// Start from the Monday of the week containing startDate
		current := getWeekStartDate(startDate)
		for !current.After(endDate) {
			weekEnd := current.AddDate(0, 0, 6)
			if weekEnd.After(endDate) {
				weekEnd = endDate
			}
			periods = append(periods, PeriodInfo{
				Date:        current,
				PeriodStart: current,
				PeriodEnd:   weekEnd,
				PeriodLabel: GeneratePeriodLabel(current, GranularityWeekly),
			})
			current = current.AddDate(0, 0, 7)
		}

	case GranularityMonthly:
		// Start from the first of the month containing startDate
		current := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, loc)
		for !current.After(endDate) {
			monthEnd := current.AddDate(0, 1, -1)
			periods = append(periods, PeriodInfo{
				Date:        current,
				PeriodStart: current,
				PeriodEnd:   monthEnd,
				PeriodLabel: GeneratePeriodLabel(current, GranularityMonthly),
			})
			current = current.AddDate(0, 1, 0)
		}

	case GranularityQuarterly:
		// Start from the first of the quarter containing startDate
		quarter := (int(startDate.Month()) - 1) / 3
		current := time.Date(startDate.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, loc)
		for !current.After(endDate) {
			quarterEnd := current.AddDate(0, 3, -1)
			periods = append(periods, PeriodInfo{
				Date:        current,
				PeriodStart: current,
				PeriodEnd:   quarterEnd,
				PeriodLabel: GeneratePeriodLabel(current, GranularityQuarterly),
			})
			current = current.AddDate(0, 3, 0)
		}
	}

	return periods
}

// PeriodInfo holds information about a single period.
type PeriodInfo struct {
	Date        time.Time
	PeriodStart time.Time
	PeriodEnd   time.Time
	PeriodLabel string
}

// getWeekStartDate returns the Monday of the week containing the given date.
func getWeekStartDate(date time.Time) time.Time {
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 7
	}
	daysFromMonday := weekday - 1
	return time.Date(date.Year(), date.Month(), date.Day()-daysFromMonday, 0, 0, 0, 0, date.Location())
}

// GetPeriodKeyForDate returns a unique key for the period containing the given date.
func GetPeriodKeyForDate(date time.Time, granularity Granularity) string {
	switch granularity {
	case GranularityWeekly:
		return getWeekStartDate(date).Format("2006-01-02")
	case GranularityMonthly:
		return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).Format("2006-01-02")
	case GranularityQuarterly:
		quarter := (int(date.Month()) - 1) / 3
		return time.Date(date.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, date.Location()).Format("2006-01-02")
	default:
		return date.Format("2006-01-02")
	}
}
