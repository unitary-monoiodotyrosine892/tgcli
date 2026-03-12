package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func newMessagesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Message operations (list, search)",
	}

	cmd.AddCommand(newMessagesListCmd(flags))
	cmd.AddCommand(newMessagesSearchCmd(flags))

	return cmd
}

func newMessagesListCmd(flags *rootFlags) *cobra.Command {
	var chatID int64
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List messages in a chat",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, true)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			messages, err := a.Store().ListMessages(chatID, limit)
			if err != nil {
				return fmt.Errorf("list messages: %w", err)
			}

			if len(messages) == 0 {
				if flags.asJSON {
					return writeJSON(os.Stdout, []interface{}{})
				}
				fmt.Println("No messages found for this chat.")
				return nil
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, messages)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tFROM\tDATE\tTEXT")
			for _, msg := range messages {
				text := msg.Text
				if len(text) > 50 {
					text = text[:47] + "..."
				}
				date := time.Unix(msg.Date, 0).Format("01/02 15:04")
				fmt.Fprintf(w, "%d\t%d\t%s\t%s\n", msg.ID, msg.FromUserID, date, text)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max messages to show")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}

func newMessagesSearchCmd(flags *rootFlags) *cobra.Command {
	var chatID int64
	var limit int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search messages (FTS)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, true)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			messages, err := a.Store().SearchMessages(query, chatID, limit)
			if err != nil {
				return fmt.Errorf("search messages: %w", err)
			}

			if len(messages) == 0 {
				if flags.asJSON {
					return writeJSON(os.Stdout, []interface{}{})
				}
				fmt.Printf("No messages found matching '%s'\n", query)
				return nil
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, messages)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tCHAT\tDATE\tTEXT")
			for _, msg := range messages {
				text := msg.Text
				if len(text) > 50 {
					text = text[:47] + "..."
				}
				date := time.Unix(msg.Date, 0).Format("01/02 15:04")
				fmt.Fprintf(w, "%d\t%d\t%s\t%s\n", msg.ID, msg.ChatID, date, text)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "search within specific chat (optional)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max results")

	return cmd
}
