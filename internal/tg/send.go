package tg

import (
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendTextOptions for sending text messages.
type SendTextOptions struct {
	ChatID  int64
	Text    string
	ReplyTo int
}

// SendText sends a text message.
func (c *Client) SendText(opts SendTextOptions) (*tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(opts.ChatID, opts.Text)
	if opts.ReplyTo != 0 {
		msg.ReplyToMessageID = opts.ReplyTo
	}

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	return &sent, nil
}

// SendFileOptions for sending files.
type SendFileOptions struct {
	ChatID   int64
	FilePath string
	Caption  string
	ReplyTo  int
}

// SendFile sends a file (document).
func (c *Client) SendFile(opts SendFileOptions) (*tgbotapi.Message, error) {
	// Check file exists
	if _, err := os.Stat(opts.FilePath); err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	doc := tgbotapi.NewDocument(opts.ChatID, tgbotapi.FilePath(opts.FilePath))
	if opts.Caption != "" {
		doc.Caption = opts.Caption
	}
	if opts.ReplyTo != 0 {
		doc.ReplyToMessageID = opts.ReplyTo
	}

	sent, err := c.bot.Send(doc)
	if err != nil {
		return nil, fmt.Errorf("send file: %w", err)
	}

	return &sent, nil
}

// SendPhotoOptions for sending photos.
type SendPhotoOptions struct {
	ChatID   int64
	FilePath string
	Caption  string
	ReplyTo  int
}

// SendPhoto sends a photo.
func (c *Client) SendPhoto(opts SendPhotoOptions) (*tgbotapi.Message, error) {
	if _, err := os.Stat(opts.FilePath); err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	photo := tgbotapi.NewPhoto(opts.ChatID, tgbotapi.FilePath(opts.FilePath))
	if opts.Caption != "" {
		photo.Caption = opts.Caption
	}
	if opts.ReplyTo != 0 {
		photo.ReplyToMessageID = opts.ReplyTo
	}

	sent, err := c.bot.Send(photo)
	if err != nil {
		return nil, fmt.Errorf("send photo: %w", err)
	}

	return &sent, nil
}

// SetReactionOptions for setting reactions.
type SetReactionOptions struct {
	ChatID    int64
	MessageID int
	Emoji     string
}

// SetReaction sets a reaction on a message.
func (c *Client) SetReaction(opts SetReactionOptions) error {
	// Note: Bot API v5.5 doesn't have native reaction support
	// Reactions require Bot API 7.0+ - would need to use raw API call
	// For now, return not supported error
	return fmt.Errorf("reactions require Bot API 7.0+ (upgrade telegram-bot-api library)")
}
