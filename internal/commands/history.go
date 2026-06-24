package commands

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/DeprecatedLuar/pbank/internal/constants"
	"github.com/DeprecatedLuar/pbank/internal/transactions"
)

const (
	defaultHistLimit = 15
	unlimitedHistory = 0
	tabWriterMinPad  = 2
)

func HandleHistory(db *sql.DB, args []string) error {
	if len(args) > 0 && !strings.HasPrefix(args[0], "--") {
		subcmd := args[0]
		switch subcmd {
		case "undo":
			if len(args) < 2 {
				return fmt.Errorf("usage: pbank history undo <id>")
			}
			txID, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid transaction ID: %w", err)
			}
			return transactions.Undo(db, txID)
		default:
			return fmt.Errorf("unknown subcommand: %s", subcmd)
		}
	}

	return listHistory(db, args)
}

func listHistory(db *sql.DB, args []string) error {
	var fundFilter, currencyFilter, sinceFilter, categoryFilter string
	limit := defaultHistLimit

	for i := 0; i < len(args); i++ {
		if args[i] == constants.FlagFund && i+1 < len(args) {
			fundFilter = args[i+1]
			i++
		} else if args[i] == constants.FlagCurrency && i+1 < len(args) {
			currencyFilter = strings.ToUpper(args[i+1])
			i++
		} else if args[i] == constants.FlagSince && i+1 < len(args) {
			sinceFilter = args[i+1]
			i++
		} else if args[i] == constants.FlagCategory && i+1 < len(args) {
			categoryFilter = args[i+1]
			i++
		} else if args[i] == constants.FlagLimit && i+1 < len(args) {
			var err error
			limit, err = strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid limit value: %w", err)
			}
			i++
		} else if args[i] == constants.FlagAll {
			limit = unlimitedHistory
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

	if limit > unlimitedHistory {
		query += " LIMIT ?"
		queryArgs = append(queryArgs, limit)
	}

	rows, err := db.Query(query, queryArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, tabWriterMinPad, ' ', 0)
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
