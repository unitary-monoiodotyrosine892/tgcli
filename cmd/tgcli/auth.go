package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/RandyVentures/tgcli/internal/tg"
)

func newAuthCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Validate bot token and show bot info",
		Long: `Validates the TGCLI_BOT_TOKEN environment variable and displays bot information.

To get a bot token:
1. Open Telegram and message @BotFather
2. Send /newbot and follow the prompts
3. Copy the token and set: export TGCLI_BOT_TOKEN="your_token"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			token := tg.GetToken()
			if token == "" {
				return fmt.Errorf("TGCLI_BOT_TOKEN environment variable is not set.\n\nTo get a token:\n1. Message @BotFather on Telegram\n2. Send /newbot\n3. Copy the token\n4. Run: export TGCLI_BOT_TOKEN=\"your_token\"")
			}

			// Create client to validate token
			a, lk, err := newApp(cmd.Context(), flags, false, true)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			client, err := a.Client()
			if err != nil {
				return fmt.Errorf("invalid token: %w", err)
			}

			me, err := client.GetMe()
			if err != nil {
				return fmt.Errorf("get bot info: %w", err)
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, map[string]interface{}{
					"id":       me.ID,
					"username": me.UserName,
					"name":     me.FirstName,
					"is_bot":   me.IsBot,
				})
			}

			fmt.Println("✅ Bot authenticated successfully!")
			fmt.Printf("   ID:       %d\n", me.ID)
			fmt.Printf("   Username: @%s\n", me.UserName)
			fmt.Printf("   Name:     %s\n", me.FirstName)
			return nil
		},
	}

	return cmd
}
