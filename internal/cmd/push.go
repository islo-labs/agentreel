package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/adamgold/agentcast/internal/storage"
)

var pushCmd = &cobra.Command{
	Use:   "push [recording-id]",
	Short: "Upload a recording and get a shareable link",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id string
		if len(args) > 0 {
			id = args[0]
		}

		rec, err := storage.Resolve(id)
		if err != nil {
			return err
		}

		// TODO: implement upload client
		fmt.Fprintf(cmd.OutOrStdout(), "Would push: %s (%s, %s)\n",
			rec.ID,
			storage.FormatDuration(rec.Duration),
			storage.FormatSize(rec.Size),
		)
		fmt.Fprintln(cmd.OutOrStdout(), "Upload not yet implemented — coming soon.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
