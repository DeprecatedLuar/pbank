package commands

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
)

func HandleFund(db *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pbank fund <add|rm|list>")
	}

	subcmd := args[0]
	switch subcmd {
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: pbank fund add <label>")
		}
		return fundAdd(db, args[1])
	case "rm":
		if len(args) < 2 {
			return fmt.Errorf("usage: pbank fund rm <label> [--force]")
		}
		force := len(args) > 2 && args[2] == "--force"
		return fundRm(db, args[1], force)
	case "list":
		return fundList(db)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcmd)
	}
}

func fundAdd(db *sql.DB, label string) error {
	_, err := db.Exec("INSERT INTO funds (label) VALUES (?)", label)
	if err != nil {
		return fmt.Errorf("failed to add fund: %w", err)
	}
	fmt.Printf("Fund '%s' added\n", label)
	return nil
}

func fundRm(db *sql.DB, label string, force bool) error {
	var fundID int
	err := db.QueryRow("SELECT id FROM funds WHERE label = ?", label).Scan(&fundID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("fund '%s' not found", label)
	} else if err != nil {
		return err
	}

	if !force {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM transactions WHERE fund_id = ?", fundID).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("fund '%s' has %d transactions. Use --force to delete", label, count)
		}
	}

	_, err = db.Exec("DELETE FROM funds WHERE id = ?", fundID)
	if err != nil {
		return fmt.Errorf("failed to delete fund: %w", err)
	}

	if force {
		fmt.Printf("Fund '%s' deleted (with all balances and transactions)\n", label)
	} else {
		fmt.Printf("Fund '%s' deleted\n", label)
	}
	return nil
}

func fundList(db *sql.DB) error {
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
		return err
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
			return err
		}
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\n", id, label, currencies, activeBalances)
		count++
	}
	w.Flush()

	if count == 0 {
		fmt.Println("No funds found")
	}
	return nil
}
