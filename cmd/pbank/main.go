package main

import (
	"fmt"
	"os"

	"github.com/DeprecatedLuar/pbank/internal/commands"
	"github.com/DeprecatedLuar/pbank/internal/db"
)

func main() {
	if len(os.Args) < 2 {
		commands.HandleHelp([]string{})
		os.Exit(1)
	}

	cmd := os.Args[1]

	if cmd == "help" || cmd == "--help" || cmd == "-h" {
		commands.HandleHelp(os.Args[1:])
		return
	}

	database, err := db.OpenDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := db.EnsureTables(database); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "fund":
		if len(os.Args) < 3 {
			commands.HandleHelp([]string{"help", "fund"})
			os.Exit(1)
		}
		if err := commands.HandleFund(database, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "add":
		if err := commands.HandleAdd(database, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "deduct":
		if err := commands.HandleDeduct(database, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := commands.HandleList(database, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "edit":
		if err := commands.HandleEdit(database, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "balance":
		if err := commands.HandleBalance(database, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "value":
		if err := commands.HandleValue(database, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		commands.HandleHelp([]string{})
		os.Exit(1)
	}
}
