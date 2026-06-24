package transactions

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	dateFormat     = "2006-01-02"
	initialBalance = 0
)

const (
	fieldTitle    = "title"
	fieldCategory = "category"
	fieldNotes    = "notes"
	fieldDate     = "date"
)

func AddMoney(db *sql.DB, fundLabel, currency string, amount float64, title, category, notes string) error {
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
		VALUES (?, ?, ?)
		ON CONFLICT(fund_id, currency) DO NOTHING
	`, fundID, currency, initialBalance)
	if err != nil {
		return err
	}

	date := time.Now().Format(dateFormat)
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

func UpdateAmount(db *sql.DB, txID int, newAmount float64) error {
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

func UpdateField(db *sql.DB, txID int, field, value string) error {
	var exists bool
	err := db.QueryRow("SELECT 1 FROM transactions WHERE id = ?", txID).Scan(&exists)
	if err == sql.ErrNoRows {
		return fmt.Errorf("transaction %d not found", txID)
	} else if err != nil {
		return err
	}

	updateQueries := map[string]string{
		fieldTitle:    "UPDATE transactions SET title = ? WHERE id = ?",
		fieldCategory: "UPDATE transactions SET category = ? WHERE id = ?",
		fieldNotes:    "UPDATE transactions SET notes = ? WHERE id = ?",
		fieldDate:     "UPDATE transactions SET date = ? WHERE id = ?",
	}
	query, ok := updateQueries[field]
	if !ok {
		return fmt.Errorf("invalid field '%s'", field)
	}

	_, err = db.Exec(query, value, txID)
	if err != nil {
		return fmt.Errorf("failed to update %s: %w", field, err)
	}

	fmt.Printf("Updated transaction %d: %s = %s\n", txID, field, value)
	return nil
}

func Undo(db *sql.DB, txID int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var amount float64
	var fundID int
	var currency, title string
	err = tx.QueryRow("SELECT amount, fund_id, currency, title FROM transactions WHERE id = ?", txID).Scan(&amount, &fundID, &currency, &title)
	if err == sql.ErrNoRows {
		return fmt.Errorf("transaction %d not found", txID)
	} else if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM transactions WHERE id = ?", txID)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE fund_balances
		SET amount = amount - ?
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

	fmt.Printf("Undid transaction %d: '%s' (%.2f %s reverted, new balance: %.2f %s)\n", txID, title, amount, currency, newBalance, currency)
	return nil
}
