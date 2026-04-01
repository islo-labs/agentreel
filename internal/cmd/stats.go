package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/adamgold/agentcast/internal/render"
	"github.com/adamgold/agentcast/internal/storage"
)

var statsCmd = &cobra.Command{
	Use:   "stats [recording-id]",
	Short: "Show stats for a recording",
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

		f, err := os.Open(rec.Path)
		if err != nil {
			return err
		}
		defer f.Close()

		st, err := render.ExtractStats(f)
		if err != nil {
			return err
		}

		title := st.Title
		if title == "" {
			title = rec.ID
		}

		fmt.Printf("\x1b[1m%s\x1b[0m\n", title)
		fmt.Printf("  Duration:   %s\n", storage.FormatDuration(st.Duration))
		fmt.Printf("  Terminal:   %dx%d\n", st.Width, st.Height)
		fmt.Printf("  Events:     %d\n", st.Events)
		fmt.Printf("  Output:     %s\n", storage.FormatSize(int64(st.OutputBytes)))
		if st.Commands > 0 {
			fmt.Printf("  Commands:   ~%d\n", st.Commands)
		}
		fmt.Printf("  File:       %s (%s)\n", rec.Path, storage.FormatSize(rec.Size))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
