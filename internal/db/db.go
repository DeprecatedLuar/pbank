package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const dbFile = "finances.db"

func getDBPath() string {
	// Override via env var
	if home := os.Getenv("PBANK_HOME"); home != "" {
		return filepath.Join(home, dbFile)
	}
	// XDG data home
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "pbank", dbFile)
	}
	// Default: ~/.local/share/pbank/
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "share", "pbank", dbFile)
	}
	// Fallback if home dir lookup fails
	return filepath.Join(".local", "share", "pbank", dbFile)
}

func OpenDB() (*sql.DB, error) {
	dbPath := getDBPath()
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

func EnsureTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS funds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		label TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS fund_balances (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		fund_id INTEGER NOT NULL,
		currency TEXT NOT NULL,
		amount REAL NOT NULL DEFAULT 0,
		UNIQUE(fund_id, currency),
		FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		fund_id INTEGER NOT NULL,
		currency TEXT NOT NULL,
		date TEXT NOT NULL,
		title TEXT NOT NULL,
		amount REAL NOT NULL,
		category TEXT,
		notes TEXT,
		FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		fund_id INTEGER NOT NULL,
		currency TEXT NOT NULL,
		name TEXT NOT NULL,
		amount REAL NOT NULL,
		billing_day INTEGER NOT NULL,
		last_charged TEXT,
		next_billing TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}
