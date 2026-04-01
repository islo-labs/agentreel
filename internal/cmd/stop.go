package cmd

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/adamgold/agentcast/internal/storage"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the active recording",
	RunE: func(cmd *cobra.Command, args []string) error {
		lock, err := storage.ReadLock()
		if err != nil {
			return fmt.Errorf("no active recording found")
		}

		if err := syscall.Kill(lock.PID, syscall.SIGTERM); err != nil {
			return fmt.Errorf("failed to stop recording (PID %d): %w", lock.PID, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Stopped recording: %s\n", lock.CastID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
