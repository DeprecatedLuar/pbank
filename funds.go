package main

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
)

func handleFund(db *sql.DB, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: pbank fund <add|rm|list>")
		os.Exit(1)
	}

	subcmd := args[0]
	switch subcmd {
	case "add":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: pbank fund add <label>")
			os.Exit(1)
		}
		fundAdd(db, args[1])
	case "rm":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: pbank fund rm <label> [--force]")
			os.Exit(1)
		}
		force := len(args) > 2 && args[2] == "--force"
		fundRm(db, args[1], force)
	case "list":
		fundList(db)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcmd)
		os.Exit(1)
	}
}

func fundAdd(db *sql.DB, label string) {
	_, err := db.Exec("INSERT INTO funds (label) VALUES (?)", label)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to add fund: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Fund '%s' added\n", label)
}

func fundRm(db *sql.DB, label string, force bool) {
	var fundID int
	err := db.QueryRow("SELECT id FROM funds WHERE label = ?", label).Scan(&fundID)
	if err == sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "Error: fund '%s' not found\n", label)
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !force {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM transactions WHERE fund_id = ?", fundID).Scan(&count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if count > 0 {
			fmt.Fprintf(os.Stderr, "Error: fund '%s' has %d transactions. Use --force to delete.\n", label, count)
			os.Exit(1)
		}
	}

	_, err = db.Exec("DELETE FROM funds WHERE id = ?", fundID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to delete fund: %v\n", err)
		os.Exit(1)
	}

	if force {
		fmt.Printf("Fund '%s' deleted (with all balances and transactions)\n", label)
	} else {
		fmt.Printf("Fund '%s' deleted\n", label)
	}
}

func fundList(db *sql.DB) {
	rows, err := db.Query(`
		SELECT f.id, f.label,
		       COUNT(DISTINCT fb.currency) as currencies,
		       COALESCE(SUM(CASE WHEN fb.amount != 0 THEN 1 ELSE 0 END), 0) as active_balances
		FROM funds f
		LEFT JOIN fund_balances fb ON f.id = fb.fund_id
		GROUP BY f.id, f.label
		ORDER BY f.label
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tFund\tCurrencies\tActive Balances")
	fmt.Fprintln(w, "--\t----\t----------\t---------------")

	count := 0
	for rows.Next() {
		var id, currencies, activeBalances int
		var label string
		if err := rows.Scan(&id, &label, &currencies, &activeBalances); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\n", id, label, currencies, activeBalances)
		count++
	}
	w.Flush()

	if count == 0 {
		fmt.Println("No funds found")
	}
}

func handleBalance(db *sql.DB, args []string) {
	rows, err := db.Query(`
		SELECT f.label, fb.currency, fb.amount
		FROM fund_balances fb
		JOIN funds f ON fb.fund_id = f.id
		ORDER BY f.label, fb.currency
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	currentFund := ""
	hasBalances := false

	for rows.Next() {
		var fund, currency string
		var amount float64

		if err := rows.Scan(&fund, &currency, &amount); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		hasBalances = true

		if fund != currentFund {
			if currentFund != "" {
				fmt.Println()
			}
			fmt.Printf("%s:\n", fund)
			currentFund = fund
		}

		fmt.Printf("  %s: %.2f\n", currency, amount)
	}

	if !hasBalances {
		fmt.Println("No balances found")
	}
}
