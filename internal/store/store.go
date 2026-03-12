package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store represents the local message/chat database.
type Store struct {
	db *sql.DB
}

// Open opens or creates the store database.
func Open(storeDir string) (*Store, error) {
	// Create store directory with restricted permissions (owner only)
	if err := os.MkdirAll(storeDir, 0700); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}

	// Ensure directory has correct permissions even if it existed
	if err := os.Chmod(storeDir, 0700); err != nil {
		return nil, fmt.Errorf("set store dir permissions: %w", err)
	}

	dbPath := filepath.Join(storeDir, "tgcli.db")
	
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode and optimizations
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set synchronous: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	s := &Store{db: db}

	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	// Set database file permissions (owner read/write only)
	if err := os.Chmod(dbPath, 0600); err != nil {
		// Non-fatal, but log if we had logging
	}

	return s, nil
}

// Close closes the database.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

type migration struct {
	version int
	name    string
	up      func(*Store) error
}

var migrations = []migration{
	{version: 1, name: "core schema", up: migrateCoreSchema},
}

// migrate creates/updates database schema.
func (s *Store) migrate() error {
	// Create migrations tracking table
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Check which migrations are applied
	applied := map[int]bool{}
	rows, err := s.db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		if err := m.up(s); err != nil {
			return fmt.Errorf("apply migration %03d %s: %w", m.version, m.name, err)
		}
		if _, err := s.db.Exec(
			`INSERT INTO schema_migrations(version, name, applied_at) VALUES(?, ?, ?)`,
			m.version,
			m.name,
			time.Now().UTC().Unix(),
		); err != nil {
			return fmt.Errorf("record migration %03d: %w", m.version, err)
		}
	}

	return nil
}

func migrateCoreSchema(s *Store) error {
	// Create tables one at a time for better error handling
	tables := []string{
		`CREATE TABLE chats (
			id INTEGER PRIMARY KEY,
			type TEXT NOT NULL,
			title TEXT,
			username TEXT,
			last_message_id INTEGER,
			last_message_ts INTEGER,
			unread_count INTEGER DEFAULT 0,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			first_name TEXT,
			last_name TEXT,
			username TEXT,
			phone TEXT,
			is_bot INTEGER DEFAULT 0,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE messages (
			id INTEGER PRIMARY KEY,
			chat_id INTEGER NOT NULL,
			from_user_id INTEGER,
			date INTEGER NOT NULL,
			text TEXT,
			reply_to_message_id INTEGER,
			media_type TEXT,
			media_path TEXT,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE INDEX idx_messages_chat_date ON messages(chat_id, date DESC)`,
		`CREATE INDEX idx_messages_from_user ON messages(from_user_id, date DESC)`,
		`CREATE INDEX idx_messages_text ON messages(text)`,
		`CREATE INDEX idx_chats_last_message ON chats(last_message_ts DESC)`,
	}

	for _, ddl := range tables {
		if _, err := s.db.Exec(ddl); err != nil {
			return fmt.Errorf("execute DDL: %w", err)
		}
	}

	return nil
}
