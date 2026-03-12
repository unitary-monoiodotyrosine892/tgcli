package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newMediaCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "Media operations",
	}

	cmd.AddCommand(newMediaDownloadCmd(flags))

	return cmd
}

func newMediaDownloadCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download media from message (not yet implemented)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Media download not yet implemented.")
			fmt.Println("Bot API media download requires file_id from received messages.")
			return nil
		},
	}

	return cmd
}
