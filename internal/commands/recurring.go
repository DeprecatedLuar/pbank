package commands

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/DeprecatedLuar/pbank/internal/constants"
	"github.com/DeprecatedLuar/pbank/internal/recurring"
	"github.com/DeprecatedLuar/pbank/internal/transactions"
)

const (
	minArgsRecurringAdd    = 5 // fund, currency, amount, name, billing_day
	minArgsRecurringEdit   = 3 // id, field, value
	minArgsRecurringPause  = 1 // id
	minArgsRecurringResume = 1 // id
	minArgsRecurringRm     = 1 // id
	dateFormat             = "2006-01-02"
)

func HandleRecurring(db *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pbank recurring <add|deduct|list|ls|edit|pause|resume|rm>")
	}

	subcmd := args[0]
	switch subcmd {
	case "add":
		if len(args) < minArgsRecurringAdd+1 {
			return fmt.Errorf("usage: pbank recurring add <fund> <currency> <amount> <name> <billing-day> [--category X] [--notes \"...\"]")
		}
		return recurringAdd(db, args[1:], false) // false = don't charge today

	case "deduct":
		if len(args) < minArgsRecurringAdd+1 {
			return fmt.Errorf("usage: pbank recurring deduct <fund> <currency> <amount> <name> <billing-day> [--category X] [--notes \"...\"]")
		}
		return recurringDeduct(db, args[1:], false)

	case "list", "ls":
		return recurringList(db, args[1:])

	case "edit":
		if len(args) < minArgsRecurringEdit+1 {
			return fmt.Errorf("usage: pbank recurring edit <id> <field> <value>")
		}
		return recurringEdit(db, args[1:])

	case "pause":
		if len(args) < minArgsRecurringPause+1 {
			return fmt.Errorf("usage: pbank recurring pause <id>")
		}
		return recurringSetStatus(db, args[1], "paused")

	case "resume":
		if len(args) < minArgsRecurringResume+1 {
			return fmt.Errorf("usage: pbank recurring resume <id>")
		}
		return recurringSetStatus(db, args[1], "active")

	case "rm":
		if len(args) < minArgsRecurringRm+1 {
			return fmt.Errorf("usage: pbank recurring rm <id>")
		}
		return recurringRm(db, args[1])

	default:
		return fmt.Errorf("unknown subcommand: %s", subcmd)
	}
}

func recurringAdd(db *sql.DB, args []string, chargeToday bool) error {
	// Parse positional args
	fundLabel := args[0]
	currency := strings.ToUpper(args[1])
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	name := args[3]
	billingDay, err := strconv.Atoi(args[4])
	if err != nil || billingDay < 1 || billingDay > 31 {
		return fmt.Errorf("billing day must be between 1 and 31")
	}

	// Parse optional flags
	var category, notes string
	for i := 5; i < len(args); i++ {
		if args[i] == constants.FlagCategory && i+1 < len(args) {
			category = args[i+1]
			i++
		} else if args[i] == constants.FlagNotes && i+1 < len(args) {
			notes = args[i+1]
			i++
		}
	}

	// Resolve fund_id
	var fundID int
	if err := db.QueryRow("SELECT id FROM funds WHERE label = ?", fundLabel).Scan(&fundID); err != nil {
		return fmt.Errorf("fund '%s' not found", fundLabel)
	}

	// Calculate initial next_billing
	today := time.Now()
	nextBilling := recurring.CalculateNextBilling(today, billingDay)

	// Insert recurring transaction
	_, err = db.Exec(`
		INSERT INTO recurring_transactions
			(fund_id, currency, name, amount, billing_day, next_billing, status, category, notes)
		VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?)
	`, fundID, currency, name, amount, billingDay, nextBilling, category, notes)

	if err != nil {
		return fmt.Errorf("failed to add recurring transaction: %w", err)
	}

	fmt.Printf("Recurring transaction '%s' added (next billing: %s)\n", name, nextBilling)

	// Optionally charge today
	if chargeToday {
		todayStr := today.Format(dateFormat)
		title := "[Recurring] " + name
		if err := transactions.AddMoney(db, fundLabel, currency, amount, title, category, notes, todayStr); err != nil {
			return fmt.Errorf("failed to create initial transaction: %w", err)
		}
		fmt.Printf("Initial transaction created for today\n")
	}

	return nil
}

func recurringDeduct(db *sql.DB, args []string, chargeToday bool) error {
	// Parse amount and negate it
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	args[2] = strconv.FormatFloat(-amount, 'f', -1, 64) // Negate for deduct

	return recurringAdd(db, args, chargeToday)
}

func recurringList(db *sql.DB, args []string) error {
	// Parse optional filters
	var fundFilter, statusFilter string
	for i := 0; i < len(args); i++ {
		if args[i] == constants.FlagFund && i+1 < len(args) {
			fundFilter = args[i+1]
			i++
		} else if args[i] == "--status" && i+1 < len(args) {
			statusFilter = args[i+1]
			i++
		}
	}

	// Build query
	query := `
		SELECT r.id, f.label, r.currency, r.name, r.amount, r.billing_day,
		       r.last_charged, r.next_billing, r.status
		FROM recurring_transactions r
		JOIN funds f ON r.fund_id = f.id
		WHERE 1=1
	`
	args_query := []interface{}{}

	if fundFilter != "" {
		query += " AND f.label = ?"
		args_query = append(args_query, fundFilter)
	}
	if statusFilter != "" {
		query += " AND r.status = ?"
		args_query = append(args_query, statusFilter)
	}

	query += " ORDER BY r.next_billing ASC, r.name ASC"

	rows, err := db.Query(query, args_query...)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Format output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tFund\tCurrency\tName\tAmount\tDay\tLast Charged\tNext Billing\tStatus")

	hasRows := false
	for rows.Next() {
		hasRows = true
		var (
			id                       int
			fund, currency, name     string
			amount                   float64
			billingDay               int
			lastCharged              sql.NullString
			nextBilling, status      string
		)

		if err := rows.Scan(&id, &fund, &currency, &name, &amount, &billingDay,
			&lastCharged, &nextBilling, &status); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		lastChargedStr := lastCharged.String
		if !lastCharged.Valid {
			lastChargedStr = "never"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%.2f\t%d\t%s\t%s\t%s\n",
			id, fund, currency, name, amount, billingDay, lastChargedStr, nextBilling, status)
	}

	w.Flush()

	if !hasRows {
		fmt.Println("No recurring transactions found")
	}

	return nil
}

func recurringEdit(db *sql.DB, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	field := args[1]
	value := args[2]

	// Validate field
	allowedFields := map[string]bool{
		"name": true, "amount": true, "billing_day": true,
		"category": true, "notes": true,
	}

	if !allowedFields[field] {
		return fmt.Errorf("invalid field: %s (allowed: name, amount, billing_day, category, notes)", field)
	}

	// Special handling for billing_day change (recalculate next_billing)
	if field == "billing_day" {
		newDay, err := strconv.Atoi(value)
		if err != nil || newDay < 1 || newDay > 31 {
			return fmt.Errorf("billing day must be between 1 and 31")
		}

		// Get current next_billing to calculate new one
		var currentNextBilling string
		if err := db.QueryRow("SELECT next_billing FROM recurring_transactions WHERE id = ?", id).
			Scan(&currentNextBilling); err != nil {
			return fmt.Errorf("recurring transaction not found")
		}

		currentDate, _ := time.Parse(dateFormat, currentNextBilling)
		newNextBilling := recurring.CalculateNextBilling(currentDate, newDay)

		_, err = db.Exec(`
			UPDATE recurring_transactions
			SET billing_day = ?, next_billing = ?
			WHERE id = ?
		`, newDay, newNextBilling, id)

		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		fmt.Printf("Updated billing_day to %d (next billing: %s)\n", newDay, newNextBilling)
		return nil
	}

	// Standard field update
	query := fmt.Sprintf("UPDATE recurring_transactions SET %s = ? WHERE id = ?", field)
	result, err := db.Exec(query, value, id)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("recurring transaction not found")
	}

	fmt.Printf("Updated %s to '%s'\n", field, value)
	return nil
}

func recurringSetStatus(db *sql.DB, idStr, status string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	result, err := db.Exec("UPDATE recurring_transactions SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("recurring transaction not found")
	}

	fmt.Printf("Recurring transaction %d set to '%s'\n", id, status)
	return nil
}

func recurringRm(db *sql.DB, idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	result, err := db.Exec("DELETE FROM recurring_transactions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("recurring transaction not found")
	}

	fmt.Printf("Recurring transaction %d deleted\n", id)
	return nil
}

// RecurringAddWithCharge is called from HandleAdd when --recurring flag is present
// (charge today + add recurring)
func RecurringAddWithCharge(db *sql.DB, args []string) error {
	return recurringAdd(db, args, true) // true = charge today
}

// RecurringDeductWithCharge is called from HandleDeduct when --recurring flag is present
func RecurringDeductWithCharge(db *sql.DB, args []string) error {
	return recurringDeduct(db, args, true)
}
