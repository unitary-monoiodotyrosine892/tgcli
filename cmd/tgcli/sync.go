package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func newSyncCmd(flags *rootFlags) *cobra.Command {
	var follow bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Listen for incoming messages",
		Long: `Listens for incoming messages and stores them locally.

Use --follow to continuously listen. Without --follow, exits after receiving
the first batch of updates.

Note: Bot API only receives messages sent TO the bot, not all messages in a chat.
For group chats, the bot must be a member and have privacy mode disabled
(talk to @BotFather: /setprivacy -> Disable).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			// Handle interrupt
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigCh
				fmt.Println("\nStopping...")
				cancel()
			}()

			a, lk, err := newApp(ctx, flags, true, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			client, err := a.Client()
			if err != nil {
				return err
			}

			return client.Sync(ctx, struct {
				Follow  bool
				Timeout int
			}{
				Follow:  follow,
				Timeout: 60,
			})
		},
	}

	cmd.Flags().BoolVar(&follow, "follow", false, "continuously listen for messages")
	return cmd
}
