package pkg

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type StateStore interface {
	Get(locationKey LocationKey, instanceKey InstanceKey, operationName OperationName) (string, error)
	Set(locationKey LocationKey, instanceKey InstanceKey, operationName OperationName, value string) error
	Close() error
}

type SQLiteStateStore struct {
	db *sql.DB
}

func NewSQLiteState(dbPath string) (*SQLiteStateStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS state (
			config_key TEXT,
			location_key TEXT,
			operation_name TEXT,
			value TEXT,
			PRIMARY KEY (config_key, location_key, operation_name)
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &SQLiteStateStore{db: db}, nil
}

func (s *SQLiteStateStore) Get(locationKey LocationKey, instanceKey InstanceKey, operationName OperationName) (string, error) {
	var value string
	err := s.db.QueryRow(
		"SELECT value FROM state WHERE location_key = ? AND instance_key = ? AND operation_name = ?",
		locationKey, instanceKey, operationName,
	).Scan(&value)

	if err == sql.ErrNoRows {
		return "", nil // Return empty string if no value found
	} else if err != nil {
		return "", fmt.Errorf("failed to get state: %w", err)
	}

	return value, nil
}

func (s *SQLiteStateStore) Set(locationKey LocationKey, instanceKey InstanceKey, operationName OperationName, value string) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO state (location_key, instance_key, operation_name, value)
		VALUES (?, ?, ?, ?)`,
		locationKey, instanceKey, operationName, value,
	)
	if err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}
	return nil
}

func (s *SQLiteStateStore) Close() error {
	return s.db.Close()
}
