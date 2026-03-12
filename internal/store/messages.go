package store

import (
	"database/sql"
	"fmt"
	"strings"
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

// MaxMessageLength is the maximum allowed message text length.
const MaxMessageLength = 4096

// InsertMessage inserts a new message.
func (s *Store) InsertMessage(id, chatID, fromUserID int64, date time.Time, text string, replyToID int, mediaType, mediaPath string) error {
	// Truncate text if too long (Telegram limit is 4096)
	if len(text) > MaxMessageLength {
		text = text[:MaxMessageLength]
	}

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
	var text, mediaType, mediaPath sql.NullString
	err := row.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &text, &msg.ReplyToMessageID, &mediaType, &mediaPath, &msg.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get message: %w", err)
	}
	msg.Text = text.String
	msg.MediaType = mediaType.String
	msg.MediaPath = mediaPath.String
	return &msg, nil
}

// ListMessages returns messages for a chat.
func (s *Store) ListMessages(chatID int64, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 50
	}
	// Cap limit to prevent abuse
	if limit > 1000 {
		limit = 1000
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
		var text, mediaType, mediaPath sql.NullString
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &text, &msg.ReplyToMessageID, &mediaType, &mediaPath, &msg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msg.Text = text.String
		msg.MediaType = mediaType.String
		msg.MediaPath = mediaPath.String
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// escapeLikePattern escapes special characters in LIKE patterns to prevent injection.
func escapeLikePattern(s string) string {
	// Escape special LIKE characters: % _ \
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// SearchMessages searches messages using LIKE with proper escaping.
func (s *Store) SearchMessages(query string, chatID int64, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 50
	}
	// Cap limit to prevent abuse
	if limit > 1000 {
		limit = 1000
	}

	// Escape LIKE wildcards in user input to prevent LIKE injection
	escapedQuery := escapeLikePattern(query)
	searchPattern := "%" + escapedQuery + "%"

	var rows *sql.Rows
	var err error

	if chatID != 0 {
		rows, err = s.db.Query(`
			SELECT id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path, updated_at
			FROM messages
			WHERE text LIKE ? ESCAPE '\' AND chat_id = ?
			ORDER BY date DESC
			LIMIT ?
		`, searchPattern, chatID, limit)
	} else {
		rows, err = s.db.Query(`
			SELECT id, chat_id, from_user_id, date, text, reply_to_message_id, media_type, media_path, updated_at
			FROM messages
			WHERE text LIKE ? ESCAPE '\'
			ORDER BY date DESC
			LIMIT ?
		`, searchPattern, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var text, mediaType, mediaPath sql.NullString
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.FromUserID, &msg.Date, &text, &msg.ReplyToMessageID, &mediaType, &mediaPath, &msg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msg.Text = text.String
		msg.MediaType = mediaType.String
		msg.MediaPath = mediaPath.String
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}
