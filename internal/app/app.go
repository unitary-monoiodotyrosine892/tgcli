package app

import (
	"fmt"
	"os"

	"github.com/RandyVentures/tgcli/internal/store"
	"github.com/RandyVentures/tgcli/internal/tg"
)

// Options for creating a new App.
type Options struct {
	StoreDir      string
	Version       string
	JSON          bool
	AllowUnauthed bool
}

// App represents the main application state.
type App struct {
	storeDir      string
	version       string
	json          bool
	allowUnauthed bool
	store         *store.Store
	client        *tg.Client
}

// New creates a new App instance.
func New(opts Options) (*App, error) {
	if opts.StoreDir == "" {
		return nil, fmt.Errorf("store directory is required")
	}

	// Ensure store directory exists
	if err := os.MkdirAll(opts.StoreDir, 0755); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}

	// Open store
	st, err := store.Open(opts.StoreDir)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	a := &App{
		storeDir:      opts.StoreDir,
		version:       opts.Version,
		json:          opts.JSON,
		allowUnauthed: opts.AllowUnauthed,
		store:         st,
	}

	return a, nil
}

// Client returns or creates the Telegram client.
func (a *App) Client() (*tg.Client, error) {
	if a.client != nil {
		return a.client, nil
	}

	token := tg.GetToken()
	if token == "" && !a.allowUnauthed {
		return nil, fmt.Errorf("TGCLI_BOT_TOKEN not set. Run 'tgcli auth' for help")
	}
	if token == "" {
		return nil, fmt.Errorf("TGCLI_BOT_TOKEN environment variable is required")
	}

	client, err := tg.New(tg.Options{
		StoreDir: a.storeDir,
		Token:    token,
		Store:    a.store,
	})
	if err != nil {
		return nil, err
	}

	a.client = client
	return client, nil
}

// Store returns the store.
func (a *App) Store() *store.Store {
	return a.store
}

// Close cleans up app resources.
func (a *App) Close() {
	if a.client != nil {
		a.client.Close()
	}
	if a.store != nil {
		a.store.Close()
	}
}

// StoreDir returns the store directory path.
func (a *App) StoreDir() string {
	return a.storeDir
}

// Version returns the app version.
func (a *App) Version() string {
	return a.version
}

// JSON returns whether JSON output is enabled.
func (a *App) JSON() bool {
	return a.json
}
