package commands

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
			gohelp.Item("history", "Show/manage transaction history", "pbank history --fund wallet --since 2026-05-01"),
			gohelp.Item("edit", "Edit a transaction field", "pbank edit 42 category groceries"),
			gohelp.Item("balance", "Show current balances for all funds", "pbank balance"),
			gohelp.Item("networth", "Show total net worth in any currency", "pbank networth USD"),
		).
		Section("Global Options",
			gohelp.Item("help", "Show help for any command", "pbank help fund"),
			gohelp.Item("help --all", "Show all help topics at once"),
		).
		Text("All commands support multi-currency tracking. Currencies are stored as 3-letter codes (USD, EUR, BTC, etc.).").
		Text("Data is stored in finances.db (SQLite) in the current directory.").
		Text("For recurring bills, use external cron jobs to call 'pbank deduct' commands.")

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
		Section("Transaction History",
			gohelp.Item("history", "Show latest 15 transactions (newest first)", "pbank history"),
			gohelp.Item("history --limit <N>", "Show latest N transactions", "pbank history --limit 50"),
			gohelp.Item("history --all", "Show all transactions (no limit)", "pbank history --all"),
			gohelp.Item("history --fund <name>", "Filter by fund", "pbank history --fund savings"),
			gohelp.Item("history --currency <code>", "Filter by currency", "pbank history --currency BTC"),
			gohelp.Item("history --since <date>", "Show transactions from date onward (YYYY-MM-DD)", "pbank history --since 2026-05-01"),
			gohelp.Item("history --category <cat>", "Filter by category", "pbank history --category food"),
			gohelp.Item("history undo <id>", "Delete transaction and revert balance changes", "pbank history undo 31"),
		).
		Text("Filters can be combined: pbank history --fund wallet --currency USD --since 2026-01-01 --limit 20")

	balance := gohelp.NewPage("balance", "view current balances").
		Usage("pbank balance").
		Text("Shows current balances for all funds, grouped by fund and currency.").
		Section("Example Output",
			gohelp.Item("Wise:", ""),
			gohelp.Item("  USD: 450.00", ""),
			gohelp.Item("  EUR: 200.00", ""),
			gohelp.Item("OCBC:", ""),
			gohelp.Item("  SGD: 5,000.00", ""),
		).
		Text("Balances are calculated from all transactions. Use 'pbank history' to see transaction history.")

	edit := gohelp.NewPage("edit", "edit transaction fields").
		Usage("pbank edit <id> <field> <value>").
		Section("Fields",
			gohelp.Item("title", "Transaction description", "pbank edit 5 title 'Updated description'"),
			gohelp.Item("amount", "Transaction amount (recalculates fund balance)", "pbank edit 5 amount -75.00"),
			gohelp.Item("category", "Transaction category", "pbank edit 5 category food"),
			gohelp.Item("notes", "Additional notes", "pbank edit 5 notes 'Paid via credit card'"),
			gohelp.Item("date", "Transaction date (YYYY-MM-DD)", "pbank edit 5 date 2026-05-15"),
		).
		Text("When editing amount, fund balance is automatically recalculated.")

	networth := gohelp.NewPage("networth", "show portfolio net worth in any currency").
		Usage("pbank networth <CURRENCY>").
		Text("Converts all balances to the specified target currency using live exchange rates.").
		Text("Crypto tickers use CoinGecko API, fiat currencies use AwesomeAPI.").
		Text("Currency must be specified as a 3-letter code (USD, EUR, BRL, JPY, etc.).").
		Section("Examples",
			gohelp.Item("pbank networth USD", "Show net worth in US Dollars"),
			gohelp.Item("pbank networth EUR", "Show net worth in Euros"),
			gohelp.Item("pbank networth BRL", "Show net worth in Brazilian Real"),
		).
		Section("Example Output (pbank networth USD)",
			gohelp.Item("Fund", "Currency\tAmount\tUSD Value"),
			gohelp.Item("Wise", "EUR\t100.00\t115.00"),
			gohelp.Item("Binance", "BTC\t0.5000\t30,000.00"),
			gohelp.Item("Total", "\t\t30,115.00 USD"),
		).
		Text("N/A is shown for currencies that cannot be converted.")

	return root, []*gohelp.Page{fund, transactions, balance, edit, networth}
}

func HandleHelp(args []string) {
	root, pages := buildHelp()
	gohelp.Run(args, root, pages...)
}
