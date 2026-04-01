package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/adamgold/agentcast/internal/recorder"
	"github.com/adamgold/agentcast/internal/storage"
)

var recCmd = &cobra.Command{
	Use:   "rec",
	Short: "Start recording a terminal session",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")

		fmt.Fprintf(cmd.ErrOrStderr(), "\x1b[1mcast:\x1b[0m Recording started. Exit the shell or run \x1b[1mcast stop\x1b[0m to finish.\n")

		result, err := recorder.Record(recorder.Options{
			Title: title,
		})
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "\n\x1b[1mcast:\x1b[0m Recording saved: %s (%s, %s)\n",
			result.ID,
			storage.FormatDuration(result.Duration),
			storage.FormatSize(result.Size),
		)
		return nil
	},
}

func init() {
	recCmd.Flags().StringP("title", "t", "", "recording title")
	rootCmd.AddCommand(recCmd)
}
