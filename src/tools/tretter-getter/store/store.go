package store

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite" // SQLite driver
)

//go:embed schema.sql
var schema string

// Store wraps the generic sqlc Queries.
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore creates a new Store with the given database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
		db:      db,
	}
}

// InitDB opens a connection to the SQLite database and executes the schema.
func InitDB(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	// Create table if not exists
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	return NewStore(db), nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
