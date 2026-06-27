package selfheal

import (
	"database/sql"
	"fmt"

	"github.com/DeprecatedLuar/pbank/internal/recurring"
)

// Run performs startup checks on every binary invocation.
// Returns non-fatal errors (logged but don't block execution).
func Run(db *sql.DB) error {
	// Process due recurring transactions
	if err := recurring.ProcessDueTransactions(db); err != nil {
		return fmt.Errorf("recurring transaction processing: %w", err)
	}

	// Future: Add other self-healing checks here
	// - Currency cache refresh
	// - Orphaned record cleanup
	// - Data integrity checks

	return nil
}
