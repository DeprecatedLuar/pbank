package commands

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

func HandleAdd(db *sql.DB, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: pbank add <fund> <currency> <amount> <title> [--category X] [--notes \"...\"]")
	}

	fund := args[0]
	currency := strings.ToUpper(args[1])
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	title := args[3]

	var category, notes string
	for i := 4; i < len(args); i++ {
		if args[i] == "--category" && i+1 < len(args) {
			category = args[i+1]
			i++
		} else if args[i] == "--notes" && i+1 < len(args) {
			notes = args[i+1]
			i++
		}
	}

	return addMoney(db, fund, currency, amount, title, category, notes)
}

func HandleDeduct(db *sql.DB, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: pbank deduct <fund> <currency> <amount> <title> [--category X] [--notes \"...\"]")
	}

	fund := args[0]
	currency := strings.ToUpper(args[1])
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	title := args[3]

	var category, notes string
	for i := 4; i < len(args); i++ {
		if args[i] == "--category" && i+1 < len(args) {
			category = args[i+1]
			i++
		} else if args[i] == "--notes" && i+1 < len(args) {
			notes = args[i+1]
			i++
		}
	}

	return addMoney(db, fund, currency, -amount, title, category, notes)
}

func addMoney(db *sql.DB, fundLabel, currency string, amount float64, title, category, notes string) error {
	var fundID int
	err := db.QueryRow("SELECT id FROM funds WHERE label = ?", fundLabel).Scan(&fundID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("fund '%s' not found", fundLabel)
	} else if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO fund_balances (fund_id, currency, amount)
		VALUES (?, ?, 0)
		ON CONFLICT(fund_id, currency) DO NOTHING
	`, fundID, currency)
	if err != nil {
		return err
	}

	date := time.Now().Format("2006-01-02")
	_, err = tx.Exec(`
		INSERT INTO transactions (fund_id, currency, date, title, amount, category, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, fundID, currency, date, title, amount, category, notes)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE fund_balances
		SET amount = amount + ?
		WHERE fund_id = ? AND currency = ?
	`, amount, fundID, currency)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	var newBalance float64
	db.QueryRow("SELECT amount FROM fund_balances WHERE fund_id = ? AND currency = ?", fundID, currency).Scan(&newBalance)

	if amount > 0 {
		fmt.Printf("Added %.2f %s to %s (balance: %.2f %s)\n", amount, currency, fundLabel, newBalance, currency)
	} else {
		fmt.Printf("Deducted %.2f %s from %s (balance: %.2f %s)\n", -amount, currency, fundLabel, newBalance, currency)
	}
	return nil
}

func HandleEdit(db *sql.DB, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: pbank edit <id> <field> <value>\nFields: title, amount, category, notes, date")
	}

	txID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid transaction ID: %w", err)
	}

	field := args[1]
	value := args[2]

	allowedFields := map[string]bool{
		"title":    true,
		"amount":   true,
		"category": true,
		"notes":    true,
		"date":     true,
	}

	if !allowedFields[field] {
		return fmt.Errorf("invalid field '%s'. Allowed: title, amount, category, notes, date", field)
	}

	if field == "amount" {
		return updateAmount(db, txID, value)
	}
	return updateField(db, txID, field, value)
}

func updateAmount(db *sql.DB, txID int, newAmountStr string) error {
	newAmount, err := strconv.ParseFloat(newAmountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var oldAmount float64
	var fundID int
	var currency string
	err = tx.QueryRow("SELECT amount, fund_id, currency FROM transactions WHERE id = ?", txID).Scan(&oldAmount, &fundID, &currency)
	if err == sql.ErrNoRows {
		return fmt.Errorf("transaction %d not found", txID)
	} else if err != nil {
		return err
	}

	diff := newAmount - oldAmount

	_, err = tx.Exec("UPDATE transactions SET amount = ? WHERE id = ?", newAmount, txID)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE fund_balances
		SET amount = amount + ?
		WHERE fund_id = ? AND currency = ?
	`, diff, fundID, currency)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("Updated transaction %d: amount %.2f → %.2f (balance adjusted by %.2f %s)\n", txID, oldAmount, newAmount, diff, currency)
	return nil
}

func updateField(db *sql.DB, txID int, field, value string) error {
	var exists bool
	err := db.QueryRow("SELECT 1 FROM transactions WHERE id = ?", txID).Scan(&exists)
	if err == sql.ErrNoRows {
		return fmt.Errorf("transaction %d not found", txID)
	} else if err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE transactions SET %s = ? WHERE id = ?", field)
	_, err = db.Exec(query, value, txID)
	if err != nil {
		return fmt.Errorf("failed to update %s: %w", field, err)
	}

	fmt.Printf("Updated transaction %d: %s = %s\n", txID, field, value)
	return nil
}

func HandleList(db *sql.DB, args []string) error {
	var fundFilter, currencyFilter, sinceFilter, categoryFilter string

	for i := 0; i < len(args); i++ {
		if args[i] == "--fund" && i+1 < len(args) {
			fundFilter = args[i+1]
			i++
		} else if args[i] == "--currency" && i+1 < len(args) {
			currencyFilter = strings.ToUpper(args[i+1])
			i++
		} else if args[i] == "--since" && i+1 < len(args) {
			sinceFilter = args[i+1]
			i++
		} else if args[i] == "--category" && i+1 < len(args) {
			categoryFilter = args[i+1]
			i++
		}
	}

	query := `
		SELECT t.id, f.label, t.currency, t.date, t.title, t.amount, t.category
		FROM transactions t
		JOIN funds f ON t.fund_id = f.id
		WHERE 1=1
	`
	var queryArgs []interface{}

	if fundFilter != "" {
		query += " AND f.label = ?"
		queryArgs = append(queryArgs, fundFilter)
	}
	if currencyFilter != "" {
		query += " AND t.currency = ?"
		queryArgs = append(queryArgs, currencyFilter)
	}
	if sinceFilter != "" {
		query += " AND t.date >= ?"
		queryArgs = append(queryArgs, sinceFilter)
	}
	if categoryFilter != "" {
		query += " AND t.category = ?"
		queryArgs = append(queryArgs, categoryFilter)
	}

	query += " ORDER BY t.date DESC, t.id DESC"

	rows, err := db.Query(query, queryArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDate\tFund\tCurrency\tAmount\tTitle\tCategory")
	fmt.Fprintln(w, "--\t----\t----\t--------\t------\t-----\t--------")

	count := 0
	for rows.Next() {
		var id int
		var fund, currency, date, title string
		var amount float64
		var category sql.NullString

		if err := rows.Scan(&id, &fund, &currency, &date, &title, &amount, &category); err != nil {
			return err
		}

		cat := "-"
		if category.Valid {
			cat = category.String
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%.2f\t%s\t%s\n", id, date, fund, currency, amount, title, cat)
		count++
	}
	w.Flush()

	if count == 0 {
		fmt.Println("No transactions found")
	}
	return nil
}
