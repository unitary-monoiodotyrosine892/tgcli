package tg

import (
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/RandyVentures/tgcli/internal/store"
)

// Client wraps the Telegram Bot API client.
type Client struct {
	bot      *tgbotapi.BotAPI
	store    *store.Store
	storeDir string
}

// Options for creating a new Client.
type Options struct {
	StoreDir string
	Token    string
	Store    *store.Store
}

// New creates a new Telegram Bot API client.
func New(opts Options) (*Client, error) {
	if opts.StoreDir == "" {
		return nil, fmt.Errorf("store directory is required")
	}
	if opts.Token == "" {
		return nil, fmt.Errorf("TGCLI_BOT_TOKEN environment variable is required")
	}

	bot, err := tgbotapi.NewBotAPI(opts.Token)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	c := &Client{
		bot:      bot,
		store:    opts.Store,
		storeDir: opts.StoreDir,
	}

	return c, nil
}

// GetToken returns bot token from environment.
func GetToken() string {
	return os.Getenv("TGCLI_BOT_TOKEN")
}

// Bot returns the underlying bot API.
func (c *Client) Bot() *tgbotapi.BotAPI {
	return c.bot
}

// Store returns the store.
func (c *Client) Store() *store.Store {
	return c.store
}

// Close stops the client.
func (c *Client) Close() error {
	c.bot.StopReceivingUpdates()
	return nil
}

// GetMe returns bot info.
func (c *Client) GetMe() (*tgbotapi.User, error) {
	return &c.bot.Self, nil
}

// IsAuthed checks if the client is authenticated (has valid token).
func (c *Client) IsAuthed() bool {
	return c.bot != nil && c.bot.Self.ID != 0
}
