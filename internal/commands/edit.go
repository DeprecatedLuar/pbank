package commands

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DeprecatedLuar/pbank/internal/transactions"
)

const minArgsEdit = 3

const (
	fieldTitle    = "title"
	fieldAmount   = "amount"
	fieldCategory = "category"
	fieldNotes    = "notes"
	fieldDate     = "date"
)

func HandleEdit(db *sql.DB, args []string) error {
	if len(args) < minArgsEdit {
		return fmt.Errorf("usage: pbank edit <id> <field> <value>\nFields: %s, %s, %s, %s, %s",
			fieldTitle, fieldAmount, fieldCategory, fieldNotes, fieldDate)
	}

	txID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid transaction ID: %w", err)
	}

	field := args[1]
	value := args[2]

	allowedFields := map[string]bool{
		fieldTitle:    true,
		fieldAmount:   true,
		fieldCategory: true,
		fieldNotes:    true,
		fieldDate:     true,
	}

	if !allowedFields[field] {
		return fmt.Errorf("invalid field '%s'. Allowed: %s, %s, %s, %s, %s", field,
			fieldTitle, fieldAmount, fieldCategory, fieldNotes, fieldDate)
	}

	if field == fieldAmount {
		amount, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}
		return transactions.UpdateAmount(db, txID, amount)
	}

	return transactions.UpdateField(db, txID, field, value)
}
