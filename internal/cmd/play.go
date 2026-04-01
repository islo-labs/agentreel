package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/adamgold/agentcast/internal/player"
	"github.com/adamgold/agentcast/internal/storage"
)

var playCmd = &cobra.Command{
	Use:   "play [recording-id]",
	Short: "Replay a recording in the terminal",
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

		speedStr, _ := cmd.Flags().GetString("speed")
		speed := parseSpeed(speedStr)

		f, err := os.Open(rec.Path)
		if err != nil {
			return fmt.Errorf("open recording: %w", err)
		}
		defer f.Close()

		return player.Play(f, player.Options{Speed: speed})
	},
}

func parseSpeed(s string) float64 {
	s = strings.TrimSuffix(strings.TrimSpace(s), "x")
	if s == "" {
		return 1.0
	}
	var speed float64
	fmt.Sscanf(s, "%f", &speed)
	if speed <= 0 {
		return 1.0
	}
	return speed
}

func init() {
	playCmd.Flags().StringP("speed", "s", "1x", "playback speed (e.g. 2x, 0.5x)")
	rootCmd.AddCommand(playCmd)
}
