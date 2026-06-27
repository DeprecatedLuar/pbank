package recurring

import "time"

// Note: dateFormat is also defined in transactions/operations.go
// This duplication is intentional per project architecture (file-specific constants)
const dateFormat = "2006-01-02"

// CalculateNextBilling calculates the next billing date from a given date
// using the Stripe-style anchor day algorithm.
//
// Algorithm:
// - Add one month to currentDate
// - Use billingDay or last day of month (whichever is smaller)
//
// Examples:
//   billingDay=31, current=2026-01-31 → 2026-02-28 (Feb has only 28 days)
//   billingDay=31, current=2026-02-28 → 2026-03-31 (back to 31st)
//   billingDay=15, current=2026-01-15 → 2026-02-15 (always 15th)
func CalculateNextBilling(currentDate time.Time, billingDay int) string {
	// Move to next month
	year, month, _ := currentDate.Date()
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)

	// Get last day of next month
	lastDayOfMonth := time.Date(nextMonth.Year(), nextMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

	// Use billing day or last day (whichever is smaller)
	actualDay := billingDay
	if billingDay > lastDayOfMonth {
		actualDay = lastDayOfMonth
	}

	return time.Date(nextMonth.Year(), nextMonth.Month(), actualDay, 0, 0, 0, 0, time.UTC).Format(dateFormat)
}
