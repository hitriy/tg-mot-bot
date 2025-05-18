package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Logger struct {
	db *sql.DB
}

type RequestLog struct {
	ID        int64
	Timestamp time.Time
	UserID    int64
	Username  string
	CarPlate  string
	Response  string
}

type Stats struct {
	LastDay   int
	LastMonth int
	AllTime   int
}

func NewLogger(dbPath string) (*Logger, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTable(db); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Logger{db: db}, nil
}

func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS request_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		user_id INTEGER NOT NULL,
		username TEXT NOT NULL,
		car_plate TEXT NOT NULL,
		response TEXT NOT NULL
	)`

	_, err := db.Exec(query)
	return err
}

func (l *Logger) LogRequest(userID int64, username, carPlate, response string) error {
	query := `
	INSERT INTO request_logs (timestamp, user_id, username, car_plate, response)
	VALUES (?, ?, ?, ?, ?)`

	_, err := l.db.Exec(query, time.Now().UTC(), userID, username, carPlate, response)
	if err != nil {
		return fmt.Errorf("failed to log request: %w", err)
	}

	return nil
}

func (l *Logger) GetStats() (*Stats, error) {
	now := time.Now().UTC()
	dayAgo := now.AddDate(0, 0, -1)
	monthAgo := now.AddDate(0, -1, 0)

	stats := &Stats{}

	// Get last day stats
	query := `SELECT COUNT(*) FROM request_logs WHERE timestamp >= ?`
	if err := l.db.QueryRow(query, dayAgo).Scan(&stats.LastDay); err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}

	// Get last month stats
	if err := l.db.QueryRow(query, monthAgo).Scan(&stats.LastMonth); err != nil {
		return nil, fmt.Errorf("failed to get monthly stats: %w", err)
	}

	// Get all time stats
	query = `SELECT COUNT(*) FROM request_logs`
	if err := l.db.QueryRow(query).Scan(&stats.AllTime); err != nil {
		return nil, fmt.Errorf("failed to get all time stats: %w", err)
	}

	return stats, nil
}

func (l *Logger) Close() error {
	return l.db.Close()
}
