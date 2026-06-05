package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
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

	cmd := os.Args[1]

	switch cmd {
	case "fund":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: pbank fund <add|rm|list>")
			os.Exit(1)
		}
		handleFund(db, os.Args[2:])
	case "add":
		handleAdd(db, os.Args[2:])
	case "deduct":
		handleDeduct(db, os.Args[2:])
	case "tx":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: pbank tx <list|edit>")
			os.Exit(1)
		}
		handleTransaction(db, os.Args[2:])
	case "sub":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: pbank sub <add|list|pause|resume|cancel>")
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
			fmt.Fprintln(os.Stderr, "Usage: pbank report <monthly|networth>")
			os.Exit(1)
		}
		handleReport(db, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: pbank <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  fund <add|rm|list>    Manage funds")
	fmt.Println("  add <fund> <curr> <amt> <title> [--category X] [--notes \"...\"]")
	fmt.Println("  deduct <fund> <curr> <amt> <title> [--category X] [--notes \"...\"]")
	fmt.Println("  tx <list|edit>        View/edit transactions")
	fmt.Println("  sub <add|list|...>    Manage subscriptions")
	fmt.Println("  cron daily            Run daily subscription billing")
	fmt.Println("  report <monthly|networth>")
}

func handleFund(db interface{}, args []string) {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	os.Exit(1)
}

func handleAdd(db interface{}, args []string) {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	os.Exit(1)
}

func handleDeduct(db interface{}, args []string) {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	os.Exit(1)
}

func handleTransaction(db interface{}, args []string) {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	os.Exit(1)
}

func handleSubscription(db interface{}, args []string) {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	os.Exit(1)
}

func handleCron(db interface{}, args []string) {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	os.Exit(1)
}

func handleReport(db interface{}, args []string) {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	os.Exit(1)
}
