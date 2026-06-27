package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Database configuration
const (
	dbFile            = "finances.db"
	dataDir           = "pbank"
	defaultDataSubdir = ".local/share"
	dataDirPerm       = 0755
)

// Environment variables
const (
	envPbankHome   = "PBANK_HOME"
	envXDGDataHome = "XDG_DATA_HOME"
)

func getDBPath() string {
	// Override via env var
	if home := os.Getenv(envPbankHome); home != "" {
		return filepath.Join(home, dbFile)
	}
	// XDG data home
	if xdg := os.Getenv(envXDGDataHome); xdg != "" {
		return filepath.Join(xdg, dataDir, dbFile)
	}
	// Default: ~/.local/share/pbank/
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, defaultDataSubdir, dataDir, dbFile)
	}
	// Fallback if home dir lookup fails
	return filepath.Join(defaultDataSubdir, dataDir, dbFile)
}

func OpenDB() (*sql.DB, error) {
	dbPath := getDBPath()
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, dataDirPerm); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

func EnsureTables(db *sql.DB) error {
	// Check if old subscriptions table exists (for migration)
	var tableName string
	err := db.QueryRow(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name='subscriptions'
	`).Scan(&tableName)

	if err == nil {
		// Old table exists, migrate it
		_, err = db.Exec(`ALTER TABLE subscriptions RENAME TO recurring_transactions`)
		if err != nil {
			return fmt.Errorf("failed to rename subscriptions table: %w", err)
		}

		// Add missing columns (ignore errors if they already exist)
		db.Exec(`ALTER TABLE recurring_transactions ADD COLUMN category TEXT`)
		db.Exec(`ALTER TABLE recurring_transactions ADD COLUMN notes TEXT`)
	}

	// Create tables (with CREATE TABLE IF NOT EXISTS)
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

	CREATE TABLE IF NOT EXISTS recurring_transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		fund_id INTEGER NOT NULL,
		currency TEXT NOT NULL,
		name TEXT NOT NULL,
		amount REAL NOT NULL,
		billing_day INTEGER NOT NULL,
		last_charged TEXT,
		next_billing TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		category TEXT,
		notes TEXT,
		FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}
