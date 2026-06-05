package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const dbFile = "finances.db"

func getDBPath() string {
	exe, err := os.Executable()
	if err != nil {
		return dbFile
	}
	return filepath.Join(filepath.Dir(exe), dbFile)
}

func openDB() (*sql.DB, error) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

func ensureTables(db *sql.DB) error {
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
