package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DeprecatedLuar/pbank/internal/constants"
	"github.com/DeprecatedLuar/pbank/internal/transactions"
)

const (
	minArgsAdd = 4
)

func HandleAdd(db *sql.DB, args []string) error {
	if len(args) < minArgsAdd {
		return fmt.Errorf("usage: pbank add <fund> <currency> <amount> <title> [%s X] [%s \"...\"] [--recurring <day>]", constants.FlagCategory, constants.FlagNotes)
	}

	fund := args[0]
	currency := strings.ToUpper(args[1])
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	title := args[3]

	var category, notes string
	var recurringDay int
	hasRecurring := false

	for i := 4; i < len(args); i++ {
		if args[i] == constants.FlagCategory && i+1 < len(args) {
			category = args[i+1]
			i++
		} else if args[i] == constants.FlagNotes && i+1 < len(args) {
			notes = args[i+1]
			i++
		} else if args[i] == "--recurring" && i+1 < len(args) {
			day, err := strconv.Atoi(args[i+1])
			if err != nil || day < 1 || day > 31 {
				return fmt.Errorf("--recurring day must be between 1 and 31")
			}
			recurringDay = day
			hasRecurring = true
			i++
		}
	}

	// If --recurring flag present, create recurring transaction + charge today
	if hasRecurring {
		// Build args for RecurringAddWithCharge
		recurringArgs := []string{
			fund, currency, strconv.FormatFloat(amount, 'f', -1, 64),
			title, strconv.Itoa(recurringDay),
		}
		if category != "" {
			recurringArgs = append(recurringArgs, constants.FlagCategory, category)
		}
		if notes != "" {
			recurringArgs = append(recurringArgs, constants.FlagNotes, notes)
		}

		return RecurringAddWithCharge(db, recurringArgs)
	}

	// Standard add (no recurring)
	date := time.Now().Format("2006-01-02")
	return transactions.AddMoney(db, fund, currency, amount, title, category, notes, date)
}
