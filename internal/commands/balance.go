package commands

import (
	"database/sql"
	"fmt"
)

func HandleBalance(db *sql.DB, args []string) error {
	rows, err := db.Query(`
		SELECT f.label, fb.currency, fb.amount
		FROM fund_balances fb
		JOIN funds f ON fb.fund_id = f.id
		ORDER BY f.label, fb.currency
	`)
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
