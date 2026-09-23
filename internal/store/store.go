// Package store persists validation runs in a local SQLite file.
package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver
)

// Run is one recorded validation run.
type Run struct {
	ID          int64
	Time        time.Time
	Files       []string
	Valid       bool
	Diagnostics int
}

// Store is a handle on a run-history database.
type Store struct {
	db *sql.DB
}

const schema = `CREATE TABLE IF NOT EXISTS runs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	ran_at TEXT NOT NULL,
	files TEXT NOT NULL,
	valid INTEGER NOT NULL,
	diagnostics INTEGER NOT NULL
)`

// Open opens (creating if needed) the database at path. Use ":memory:" for a throwaway store.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// An in-memory database exists per connection, so keep a single one.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("init %s: %w", path, err)
	}
	return &Store{db: db}, nil
}

// Close releases the database.
func (s *Store) Close() error { return s.db.Close() }

// Add stores a run and returns its assigned id. A zero Time is replaced by now.
func (s *Store) Add(r Run) (int64, error) {
	if r.Time.IsZero() {
		r.Time = time.Now()
	}
	valid := 0
	if r.Valid {
		valid = 1
	}
	res, err := s.db.Exec(`INSERT INTO runs (ran_at, files, valid, diagnostics) VALUES (?, ?, ?, ?)`,
		r.Time.UTC().Format(time.RFC3339Nano), strings.Join(r.Files, "\n"), valid, r.Diagnostics)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// List returns the most recent runs, newest first. limit <= 0 means all.
func (s *Store) List(limit int) ([]Run, error) {
	q := `SELECT id, ran_at, files, valid, diagnostics FROM runs ORDER BY id DESC`
	args := []any{}
	if limit > 0 {
		q += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	runs := []Run{}
	for rows.Next() {
		var (
			r     Run
			at    string
			files string
			valid int
		)
		if err := rows.Scan(&r.ID, &at, &files, &valid, &r.Diagnostics); err != nil {
			return nil, err
		}
		if r.Time, err = time.Parse(time.RFC3339Nano, at); err != nil {
			return nil, err
		}
		if files != "" {
			r.Files = strings.Split(files, "\n")
		}
		r.Valid = valid == 1
		runs = append(runs, r)
	}
	return runs, rows.Err()
}
