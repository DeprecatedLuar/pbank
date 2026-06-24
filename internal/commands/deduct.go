package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/DeprecatedLuar/pbank/internal/constants"
	"github.com/DeprecatedLuar/pbank/internal/transactions"
)

const minArgsDeduct = 4

func HandleDeduct(db *sql.DB, args []string) error {
	if len(args) < minArgsDeduct {
		return fmt.Errorf("usage: pbank deduct <fund> <currency> <amount> <title> [%s X] [%s \"...\"]", constants.FlagCategory, constants.FlagNotes)
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
		if args[i] == constants.FlagCategory && i+1 < len(args) {
			category = args[i+1]
			i++
		} else if args[i] == constants.FlagNotes && i+1 < len(args) {
			notes = args[i+1]
			i++
		}
	}

	return transactions.AddMoney(db, fund, currency, -amount, title, category, notes)
}
