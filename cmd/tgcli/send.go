package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/RandyVentures/tgcli/internal/tg"
)

func newSendCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send messages (text, file)",
	}

	cmd.AddCommand(newSendTextCmd(flags))
	cmd.AddCommand(newSendFileCmd(flags))

	return cmd
}

func newSendTextCmd(flags *rootFlags) *cobra.Command {
	var to int64
	var message string
	var replyTo int

	cmd := &cobra.Command{
		Use:   "text",
		Short: "Send text message",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			client, err := a.Client()
			if err != nil {
				return err
			}

			sent, err := client.SendText(tg.SendTextOptions{
				ChatID:  to,
				Text:    message,
				ReplyTo: replyTo,
			})
			if err != nil {
				return err
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, map[string]interface{}{
					"message_id": sent.MessageID,
					"chat_id":    sent.Chat.ID,
					"text":       sent.Text,
					"date":       sent.Date,
				})
			}

			fmt.Printf("✅ Message sent (ID: %d)\n", sent.MessageID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&to, "to", 0, "recipient chat ID (required)")
	cmd.Flags().StringVar(&message, "message", "", "message text (required)")
	cmd.Flags().IntVar(&replyTo, "reply-to", 0, "reply to message ID (optional)")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("message")

	return cmd
}

func newSendFileCmd(flags *rootFlags) *cobra.Command {
	var to int64
	var file string
	var caption string
	var asPhoto bool

	cmd := &cobra.Command{
		Use:   "file",
		Short: "Send file or photo",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			client, err := a.Client()
			if err != nil {
				return err
			}

			var sent interface{}
			if asPhoto {
				s, err := client.SendPhoto(tg.SendPhotoOptions{
					ChatID:   to,
					FilePath: file,
					Caption:  caption,
				})
				if err != nil {
					return err
				}
				sent = s
			} else {
				s, err := client.SendFile(tg.SendFileOptions{
					ChatID:   to,
					FilePath: file,
					Caption:  caption,
				})
				if err != nil {
					return err
				}
				sent = s
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, sent)
			}

			fmt.Println("✅ File sent successfully")
			return nil
		},
	}

	cmd.Flags().Int64Var(&to, "to", 0, "recipient chat ID (required)")
	cmd.Flags().StringVar(&file, "file", "", "file path (required)")
	cmd.Flags().StringVar(&caption, "caption", "", "file caption (optional)")
	cmd.Flags().BoolVar(&asPhoto, "photo", false, "send as photo instead of document")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
