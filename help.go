package main

import (
	gohelp "github.com/DeprecatedLuar/gohelp-luar"
)

func buildHelp() (*gohelp.Page, []*gohelp.Page) {
	root := gohelp.NewPage("pbank", "personal finance tracker with multi-currency support").
		Usage("pbank <command> [args]").
		Section("Commands",
			gohelp.Item("fund", "Manage funds (accounts, wallets, etc.)", "pbank fund list"),
			gohelp.Item("add", "Add money to a fund", "pbank add wallet USD 100 'Salary' --category income"),
			gohelp.Item("deduct", "Deduct money from a fund", "pbank deduct wallet USD 25 'Groceries' --category food"),
			gohelp.Item("list", "List transactions with optional filters", "pbank list --fund wallet --since 2026-05-01"),
			gohelp.Item("edit", "Edit a transaction field", "pbank edit 42 category groceries"),
			gohelp.Item("sub", "Manage recurring subscriptions", "pbank sub list"),
			gohelp.Item("cron", "Run scheduled tasks (e.g., subscription billing)", "pbank cron daily"),
			gohelp.Item("report", "Generate financial reports", "pbank report monthly"),
		).
		Section("Global Options",
			gohelp.Item("help", "Show help for any command", "pbank help fund"),
			gohelp.Item("help --all", "Show all help topics at once"),
		).
		Text("All commands support multi-currency tracking. Currencies are stored as 3-letter codes (USD, EUR, BTC, etc.).").
		Text("Data is stored in finances.db (SQLite) in the current directory.")

	fund := gohelp.NewPage("fund", "manage funds (accounts, wallets, savings, etc.)").
		Usage("pbank fund <subcommand> [args]").
		Section("Subcommands",
			gohelp.Item("add <label>", "Create a new fund", "pbank fund add 'savings'"),
			gohelp.Item("rm <label>", "Delete a fund (fails if it has transactions)", "pbank fund rm 'old-wallet'"),
			gohelp.Item("rm <label> --force", "Delete a fund and all its transactions and balances", "pbank fund rm 'temp' --force"),
			gohelp.Item("list", "Show all funds with balance summary", "pbank fund list"),
		).
		Text("Funds are containers for money. Think of them as accounts, wallets, or budget categories.").
		Text("Each fund can hold balances in multiple currencies independently.")

	transactions := gohelp.NewPage("transactions", "track money movements").
		Text("Transactions record all additions and deductions to your funds.").
		Section("Adding Money",
			gohelp.Item("add <fund> <curr> <amt> <title>", "Add money to a fund", "pbank add wallet USD 1000 'Paycheck'"),
			gohelp.Item("add ... --category <cat>", "Tag with a category", "pbank add wallet USD 50 'Freelance' --category income"),
			gohelp.Item("add ... --notes <text>", "Add notes", "pbank add wallet USD 100 'Gift' --notes 'Birthday from mom'"),
		).
		Section("Deducting Money",
			gohelp.Item("deduct <fund> <curr> <amt> <title>", "Deduct money from a fund", "pbank deduct wallet USD 45 'Dinner'"),
			gohelp.Item("deduct ... --category <cat>", "Tag with a category", "pbank deduct wallet USD 12 'Coffee' --category food"),
		).
		Section("Listing Transactions",
			gohelp.Item("list", "Show all transactions (newest first)", "pbank list"),
			gohelp.Item("list --fund <name>", "Filter by fund", "pbank list --fund savings"),
			gohelp.Item("list --currency <code>", "Filter by currency", "pbank list --currency BTC"),
			gohelp.Item("list --since <date>", "Show transactions from date onward (YYYY-MM-DD)", "pbank list --since 2026-05-01"),
			gohelp.Item("list --category <cat>", "Filter by category", "pbank list --category food"),
		).
		Text("Filters can be combined: pbank list --fund wallet --currency USD --since 2026-01-01")

	sub := gohelp.NewPage("sub", "manage recurring subscriptions").
		Usage("pbank sub <subcommand> [args]").
		Section("Subcommands",
			gohelp.Item("add", "Add a new subscription (not implemented yet)"),
			gohelp.Item("list", "List all active subscriptions (not implemented yet)"),
			gohelp.Item("pause", "Temporarily pause a subscription (not implemented yet)"),
			gohelp.Item("resume", "Resume a paused subscription (not implemented yet)"),
			gohelp.Item("cancel", "Cancel a subscription permanently (not implemented yet)"),
		).
		Text("Subscriptions automatically create deduction transactions on their billing cycle.").
		Text("Run 'pbank cron daily' in a cron job to process subscriptions.")

	report := gohelp.NewPage("report", "generate financial reports").
		Usage("pbank report <type> [options]").
		Section("Report Types",
			gohelp.Item("monthly", "Show income/expense breakdown by month (not implemented yet)"),
			gohelp.Item("networth", "Calculate total net worth across all funds and currencies (not implemented yet)"),
		).
		Text("Reports aggregate your transaction data to provide financial insights.")

	return root, []*gohelp.Page{fund, transactions, sub, report}
}

func showHelp(args []string) {
	root, pages := buildHelp()
	gohelp.Run(args, root, pages...)
}
