package storage

import (
	"Leaderboard/config"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	cfg := config.LoadConfig()
	connStr := cfg.GetDBConnectionString()

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return createTables()
}

func createTables() error {
	usersTable := `
    CREATE TABLE IF NOT EXISTS users (
        user_id SERIAL PRIMARY KEY,
        username VARCHAR(255) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`

	leaderboardTable := `
    CREATE TABLE IF NOT EXISTS leaderboard (
        id SERIAL PRIMARY KEY,
        username VARCHAR(255) NOT NULL,
        score INTEGER NOT NULL,
        submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (username) REFERENCES users(username)
    )`

	if _, err := DB.Exec(usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	if _, err := DB.Exec(leaderboardTable); err != nil {
		return fmt.Errorf("failed to create leaderboard table: %w", err)
	}

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
