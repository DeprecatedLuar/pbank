package recurring

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/DeprecatedLuar/pbank/internal/transactions"
)

const (
	recurringPrefix  = "[Recurring] "
	statusActive     = "active"
	warningThreshold = 12 // Warn if processing > 12 cycles
)

// ProcessDueTransactions finds and processes all recurring transactions
// where next_billing <= today.
func ProcessDueTransactions(db *sql.DB) error {
	today := time.Now()
	todayStr := today.Format(dateFormat)

	// Query active recurring transactions due for processing
	query := `
		SELECT id, fund_id, currency, name, amount, billing_day,
		       last_charged, next_billing, category, notes
		FROM recurring_transactions
		WHERE status = ? AND next_billing <= ?
		  AND (last_charged IS NULL OR last_charged != ?)
	`

	rows, err := db.Query(query, statusActive, todayStr, todayStr)
	if err != nil {
		return fmt.Errorf("query recurring transactions: %w", err)
	}
	defer rows.Close()

	// Track if any errors occurred (non-fatal)
	var processingErrors []error

	for rows.Next() {
		var (
			id                         int
			fundID                     int
			currency, name             string
			amount                     float64
			billingDay                 int
			lastCharged, nextBilling   string
			category, notes            sql.NullString
		)

		if err := rows.Scan(&id, &fundID, &currency, &name, &amount, &billingDay,
			&lastCharged, &nextBilling, &category, &notes); err != nil {
			processingErrors = append(processingErrors,
				fmt.Errorf("scan recurring transaction %d: %w", id, err))
			continue
		}

		// Get fund label for transaction creation
		var fundLabel string
		if err := db.QueryRow("SELECT label FROM funds WHERE id = ?", fundID).Scan(&fundLabel); err != nil {
			processingErrors = append(processingErrors,
				fmt.Errorf("recurring transaction %d: fund not found: %w", id, err))
			continue
		}

		// Parse next_billing date
		currentBilling, err := time.Parse(dateFormat, nextBilling)
		if err != nil {
			processingErrors = append(processingErrors,
				fmt.Errorf("recurring transaction %d: invalid next_billing: %w", id, err))
			continue
		}

		// Calculate total cycles first (for warning detection)
		totalCycles := countCycles(currentBilling, today, billingDay)

		// Prepare warning suffix if needed
		warningSuffix := ""
		if totalCycles > warningThreshold {
			warningSuffix = fmt.Sprintf(" [Caught up %d missed cycles on %s]", totalCycles, todayStr)
			fmt.Fprintf(os.Stderr, "Warning: Processing %d missed cycles for '%s'\n", totalCycles, name)
		}

		// Process all missed cycles
		processedCount := 0
		var lastProcessedDate time.Time

		for currentBilling.Compare(today) <= 0 {
			// Create transaction dated on billing date (not today!)
			title := recurringPrefix + name
			txNotes := notes.String

			// Append warning to first transaction's notes if needed
			if processedCount == 0 && warningSuffix != "" {
				if txNotes != "" {
					txNotes = txNotes + warningSuffix
				} else {
					txNotes = warningSuffix
				}
			}

			// Create transaction via existing AddMoney helper
			if err := transactions.AddMoney(
				db,
				fundLabel,
				currency,
				amount,
				title,
				category.String,
				txNotes,
				currentBilling.Format(dateFormat), // Use billing date, not today!
			); err != nil {
				processingErrors = append(processingErrors,
					fmt.Errorf("recurring transaction %d (%s): create transaction failed: %w",
						id, name, err))
				break // Stop processing this recurring item
			}

			lastProcessedDate = currentBilling
			processedCount++

			// Move to next billing cycle
			currentBilling, _ = time.Parse(dateFormat, CalculateNextBilling(currentBilling, billingDay))
		}

		if processedCount == 0 {
			continue // Nothing processed, skip update
		}

		// Update recurring transaction metadata
		_, err = db.Exec(`
			UPDATE recurring_transactions
			SET last_charged = ?, next_billing = ?
			WHERE id = ?
		`, lastProcessedDate.Format(dateFormat), currentBilling.Format(dateFormat), id)

		if err != nil {
			processingErrors = append(processingErrors,
				fmt.Errorf("recurring transaction %d: update metadata failed: %w", id, err))
		}
	}

	// Return combined errors (non-fatal)
	if len(processingErrors) > 0 {
		return fmt.Errorf("processed with %d errors: %v", len(processingErrors), processingErrors)
	}

	return nil
}

// countCycles counts how many billing cycles exist between start and end dates.
// Helper for determining if we should add warning to transaction notes.
func countCycles(start, end time.Time, billingDay int) int {
	count := 0
	current := start

	for current.Compare(end) <= 0 {
		count++
		next, _ := time.Parse(dateFormat, CalculateNextBilling(current, billingDay))
		current = next

		if count > 100 { // Safety limit for this helper
			break
		}
	}

	return count
}
