package commands

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/pbank/internal/constants"
)

func HandleBalance(db *sql.DB, args []string) error {
	var fundFilter, currencyFilter string

	// Parse flags
	for i := 0; i < len(args); i++ {
		if args[i] == constants.FlagFund && i+1 < len(args) {
			fundFilter = args[i+1]
			i++
		} else if args[i] == constants.FlagCurrency && i+1 < len(args) {
			currencyFilter = strings.ToUpper(args[i+1])
			i++
		}
	}

	// Build query with filters
	query := `
		SELECT f.label, fb.currency, fb.amount
		FROM fund_balances fb
		JOIN funds f ON fb.fund_id = f.id
		WHERE 1=1  -- Base condition for dynamic filter building
	`
	var queryArgs []interface{}

	if fundFilter != "" {
		query += " AND f.label = ?"
		queryArgs = append(queryArgs, fundFilter)
	}
	if currencyFilter != "" {
		query += " AND fb.currency = ?"
		queryArgs = append(queryArgs, currencyFilter)
	}

	query += " ORDER BY f.label, fb.currency"

	rows, err := db.Query(query, queryArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	currentFund := ""
	hasBalances := false

	for rows.Next() {
		var fund, currency string
		var amount float64

		if err := rows.Scan(&fund, &currency, &amount); err != nil {
			return err
		}

		hasBalances = true

		if fund != currentFund {
			if currentFund != "" {
				fmt.Println()
			}
			fmt.Printf("%s:\n", fund)
			currentFund = fund
		}

		fmt.Printf("  %s: %v\n", currency, amount)
	}

	if !hasBalances {
		fmt.Println("No balances found")
	}
	return nil
}
