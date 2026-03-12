package store

import (
	"fmt"
	"time"
)

// User represents a user in the database.
type User struct {
	ID        int64
	FirstName string
	LastName  string
	Username  string
	Phone     string
	IsBot     bool
	UpdatedAt int64
}

// UpsertUser inserts or updates a user.
func (s *Store) UpsertUser(id int64, firstName, lastName, username string, isBot bool) error {
	now := time.Now().UTC().Unix()
	isBotInt := 0
	if isBot {
		isBotInt = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO users (id, first_name, last_name, username, is_bot, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			first_name = excluded.first_name,
			last_name = excluded.last_name,
			username = excluded.username,
			is_bot = excluded.is_bot,
			updated_at = excluded.updated_at
	`, id, firstName, lastName, username, isBotInt, now)
	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}
	return nil
}

// GetUser retrieves a user by ID.
func (s *Store) GetUser(id int64) (*User, error) {
	row := s.db.QueryRow(`
		SELECT id, first_name, last_name, username, phone, is_bot, updated_at
		FROM users WHERE id = ?
	`, id)

	var user User
	var phone *string
	var isBotInt int
	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &phone, &isBotInt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if phone != nil {
		user.Phone = *phone
	}
	user.IsBot = isBotInt == 1
	return &user, nil
}

// ListUsers returns all users.
func (s *Store) ListUsers(limit int) ([]User, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.Query(`
		SELECT id, first_name, last_name, username, phone, is_bot, updated_at
		FROM users
		ORDER BY updated_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var phone *string
		var isBotInt int
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &phone, &isBotInt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		if phone != nil {
			user.Phone = *phone
		}
		user.IsBot = isBotInt == 1
		users = append(users, user)
	}

	return users, rows.Err()
}
