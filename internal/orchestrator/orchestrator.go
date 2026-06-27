package orchestrator

import (
	"fmt"
	"os"

	"github.com/DeprecatedLuar/pbank/internal/commands"
	"github.com/DeprecatedLuar/pbank/internal/db"
	"github.com/DeprecatedLuar/pbank/internal/selfheal"
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
	cmdHelp      = "help"
	cmdFund      = "fund"
	cmdAdd       = "add"
	cmdDeduct    = "deduct"
	cmdList      = "history"
	cmdEdit      = "edit"
	cmdBalance   = "balance"
	cmdNetworth  = "networth"
	cmdRecurring = "recurring"
)

// Help flags
const (
	flagHelp      = "--help"
	flagHelpShort = "-h"
)

// Run orchestrates the application flow and returns an exit code.
// It sequences calls to handlers without containing business logic.
func Run(args []string) int {
	if len(args) < minArgs {
		commands.HandleHelp([]string{})
		return exitError
	}

	cmd := args[cmdArgIdx]

	if cmd == cmdHelp || cmd == flagHelp || cmd == flagHelpShort {
		commands.HandleHelp(args[cmdArgIdx:])
		return exitSuccess
	}

	database, err := db.OpenDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitError
	}
	defer database.Close()

	if err := db.EnsureTables(database); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitError
	}

	// Self-healing: process recurring transactions
	if err := selfheal.Run(database); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: self-heal failed: %v\n", err)
		// Non-fatal - continue to command execution
	}

	switch cmd {
	case cmdFund:
		if len(args) < minArgs+1 {
			commands.HandleHelp([]string{cmdHelp, cmdFund})
			return exitError
		}
		if err := commands.HandleFund(database, args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return exitError
		}
	case cmdAdd:
		if err := commands.HandleAdd(database, args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return exitError
		}
	case cmdDeduct:
		if err := commands.HandleDeduct(database, args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return exitError
		}
	case cmdList:
		if err := commands.HandleHistory(database, args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return exitError
		}
	case cmdEdit:
		if err := commands.HandleEdit(database, args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return exitError
		}
	case cmdBalance:
		if err := commands.HandleBalance(database, args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return exitError
		}
	case cmdNetworth:
		if err := commands.HandleNetworth(database, args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return exitError
		}
	case cmdRecurring:
		if len(args) < minArgs+1 {
			commands.HandleHelp([]string{cmdHelp, cmdRecurring})
			return exitError
		}
		if err := commands.HandleRecurring(database, args[argsOffset:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return exitError
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		commands.HandleHelp([]string{})
		return exitError
	}

	return exitSuccess
}
