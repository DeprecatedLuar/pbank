package main

import (
	"database/sql"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		showHelp([]string{})
		os.Exit(1)
	}

	cmd := os.Args[1]

	// Handle help command
	if cmd == "help" || cmd == "--help" || cmd == "-h" {
		showHelp(os.Args[1:])
		return
	}

	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := ensureTables(db); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "fund":
		if len(os.Args) < 3 {
			showHelp([]string{"help", "fund"})
			os.Exit(1)
		}
		handleFund(db, os.Args[2:])
	case "add":
		handleAdd(db, os.Args[2:])
	case "deduct":
		handleDeduct(db, os.Args[2:])
	case "list":
		txList(db, os.Args[2:])
	case "edit":
		txEdit(db, os.Args[2:])
	case "sub":
		if len(os.Args) < 3 {
			showHelp([]string{"help", "sub"})
			os.Exit(1)
		}
		handleSubscription(db, os.Args[2:])
	case "cron":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: pbank cron daily")
			os.Exit(1)
		}
		handleCron(db, os.Args[2:])
	case "report":
		if len(os.Args) < 3 {
			showHelp([]string{"help", "report"})
			os.Exit(1)
		}
		handleReport(db, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		showHelp([]string{})
		os.Exit(1)
	}
}

func handleSubscription(db *sql.DB, args []string) {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	os.Exit(1)
}

func handleCron(db *sql.DB, args []string) {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	os.Exit(1)
}

func handleReport(db *sql.DB, args []string) {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	os.Exit(1)
}
