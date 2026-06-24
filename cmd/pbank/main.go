package main

import (
	"fmt"
	"os"

	"github.com/DeprecatedLuar/pbank/internal/commands"
	"github.com/DeprecatedLuar/pbank/internal/db"
)

// Exit codes
const (
	exitSuccess = 0
	exitError   = 1
)

// Argument indices
const (
	minArgs    = 2
	cmdArgIdx  = 1
	argsOffset = 2
)

// Command names
const (
	cmdHelp     = "help"
	cmdFund     = "fund"
	cmdAdd      = "add"
	cmdDeduct   = "deduct"
	cmdList     = "history"
	cmdEdit     = "edit"
	cmdBalance  = "balance"
	cmdNetworth = "networth"
)

// Help flags
const (
	flagHelp      = "--help"
	flagHelpShort = "-h"
)

func main() {
	if len(os.Args) < minArgs {
		commands.HandleHelp([]string{})
		os.Exit(exitError)
	}

	cmd := os.Args[cmdArgIdx]

	if cmd == cmdHelp || cmd == flagHelp || cmd == flagHelpShort {
		commands.HandleHelp(os.Args[cmdArgIdx:])
		return
	}

	database, err := db.OpenDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitError)
	}
	defer database.Close()

	if err := db.EnsureTables(database); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitError)
	}

	switch cmd {
	case cmdFund:
		if len(os.Args) < minArgs+1 {
			commands.HandleHelp([]string{cmdHelp, cmdFund})
			os.Exit(exitError)
		}
		if err := commands.HandleFund(database, os.Args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(exitError)
		}
	case cmdAdd:
		if err := commands.HandleAdd(database, os.Args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(exitError)
		}
	case cmdDeduct:
		if err := commands.HandleDeduct(database, os.Args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(exitError)
		}
	case cmdList:
		if err := commands.HandleHistory(database, os.Args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(exitError)
		}
	case cmdEdit:
		if err := commands.HandleEdit(database, os.Args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(exitError)
		}
	case cmdBalance:
		if err := commands.HandleBalance(database, os.Args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(exitError)
		}
	case cmdNetworth:
		if err := commands.HandleNetworth(database, os.Args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(exitError)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		commands.HandleHelp([]string{})
		os.Exit(exitError)
	}
}
