package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/adamgold/agentcast/internal/render"
	"github.com/adamgold/agentcast/internal/storage"
)

var gifCmd = &cobra.Command{
	Use:   "gif [recording-id or .cast file]",
	Short: "Export a recording as an animated GIF",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		speed, _ := cmd.Flags().GetFloat64("speed")
		maxCols, _ := cmd.Flags().GetInt("cols")
		maxRows, _ := cmd.Flags().GetInt("rows")
		output, _ := cmd.Flags().GetString("output")

		// Resolve input: could be a file path or a recording ID
		var castPath, label string
		if len(args) > 0 && (strings.HasSuffix(args[0], ".cast") || strings.Contains(args[0], "/")) {
			castPath = args[0]
			label = filepath.Base(castPath)
		} else {
			var id string
			if len(args) > 0 {
				id = args[0]
			}
			rec, err := storage.Resolve(id)
			if err != nil {
				return err
			}
			castPath = rec.Path
			label = rec.ID
		}

		if output == "" {
			output = strings.TrimSuffix(filepath.Base(castPath), ".cast") + ".gif"
		}

		castFile, err := os.Open(castPath)
		if err != nil {
			return err
		}
		defer castFile.Close()

		outFile, err := os.Create(output)
		if err != nil {
			return err
		}
		defer outFile.Close()

		fmt.Fprintf(os.Stderr, "Rendering GIF from %s...\n", label)

		err = render.RenderGIF(castFile, outFile, render.GIFOptions{
			MaxWidth:  maxCols,
			MaxHeight: maxRows,
			IdleLimit: 2.0,
			FrameRate: 10,
			SpeedUp:   speed,
			ShowStats: true,
		})
		if err != nil {
			os.Remove(output)
			return fmt.Errorf("render: %w", err)
		}

		info, _ := outFile.Stat()
		size := "?"
		if info != nil {
			size = storage.FormatSize(info.Size())
		}
		fmt.Fprintf(os.Stderr, "Saved: %s (%s)\n", output, size)
		return nil
	},
}

func init() {
	gifCmd.Flags().Float64P("speed", "s", 1.0, "speed multiplier (e.g. 2.0)")
	gifCmd.Flags().IntP("cols", "c", 0, "max columns (0 = use recording width)")
	gifCmd.Flags().IntP("rows", "r", 0, "max rows (0 = use recording height)")
	gifCmd.Flags().StringP("output", "o", "", "output file (default: <id>.gif)")
	rootCmd.AddCommand(gifCmd)
}
