package store

import (
	"fmt"
	"time"
)

// Message represents a message in the database.
type Message struct {
	ID               int64
	ChatID           int64
	FromUserID       int64
	Date             int64
	Text             string
	ReplyToMessageID int
	MediaType        string
	MediaPath        string
	UpdatedAt        int64
}

// InsertMessage inserts a new message.
func (s *Store) InsertMessage(id, chatID, fromUserID int64, date time.Time, text string, replyToID int, mediaType, mediaPath string) error {
	now := time.Now().UTC().Unix()
	dateUnix := date.Unix()

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO messages (id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, chatID, fromUserID, dateUnix, text, replyToID, mediaType, mediaPath, now)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	// Update chat's last message
	_ = s.UpdateChatLastMessage(chatID, id, dateUnix)

	return nil
}

// GetMessage retrieves a message by ID.
func (s *Store) GetMessage(id int64) (*Message, error) {
	row := s.db.QueryRow(`
		SELECT id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path, updated_at
		FROM messages WHERE id = ?
	`, id)

	var msg Message
	err := row.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &msg.Text, &msg.ReplyToMessageID, &msg.MediaType, &msg.MediaPath, &msg.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get message: %w", err)
	}
	return &msg, nil
}

// ListMessages returns messages for a chat.
func (s *Store) ListMessages(chatID int64, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(`
		SELECT id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path, updated_at
		FROM messages
		WHERE chat_id = ?
		ORDER BY date DESC
		LIMIT ?
	`, chatID, limit)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &msg.Text, &msg.ReplyToMessageID, &msg.MediaType, &msg.MediaPath, &msg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// SearchMessages searches messages using FTS5.
func (s *Store) SearchMessages(query string, chatID int64, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 50
	}

	var rows interface{ Close() error }
	var err error

	if chatID != 0 {
		rows, err = s.db.Query(`
			SELECT m.id, m.chat_id, m.from_user_id, m.date, m.text, m.reply_to_message_id, m.media_type, m.media_path, m.updated_at
			FROM messages m
			JOIN messages_fts fts ON m.id = fts.rowid
			WHERE fts.text MATCH ? AND m.chat_id = ?
			ORDER BY m.date DESC
			LIMIT ?
		`, query, chatID, limit)
	} else {
		rows, err = s.db.Query(`
			SELECT m.id, m.chat_id, m.from_user_id, m.date, m.text, m.reply_to_message_id, m.media_type, m.media_path, m.updated_at
			FROM messages m
			JOIN messages_fts fts ON m.id = fts.rowid
			WHERE fts.text MATCH ?
			ORDER BY m.date DESC
			LIMIT ?
		`, query, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	defer rows.Close()

	// Type assert to get Scan method
	sqlRows, ok := rows.(interface {
		Next() bool
		Scan(...interface{}) error
		Err() error
	})
	if !ok {
		return nil, fmt.Errorf("unexpected rows type")
	}

	var messages []Message
	for sqlRows.Next() {
		var msg Message
		if err := sqlRows.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &msg.Text, &msg.ReplyToMessageID, &msg.MediaType, &msg.MediaPath, &msg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, sqlRows.Err()
}
