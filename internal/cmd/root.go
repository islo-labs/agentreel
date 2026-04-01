package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/adamgold/agentcast/internal/diff"
	"github.com/adamgold/agentcast/internal/render"
	"github.com/adamgold/agentcast/internal/storage"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "cast",
	Short: "Generate a shareable GIF from your git diff",
	Long:  "Run after your agent is done. Reads git diff, generates an animated GIF of the changes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		base, _ := cmd.Flags().GetString("base")

		var summary diff.Summary
		var err error
		if base != "" {
			summary, err = diff.FromRef(base)
		} else {
			summary, err = diff.FromGit()
		}
		if err != nil {
			return fmt.Errorf("read git state: %w", err)
		}
		if summary.Stats.FilesChanged == 0 {
			fmt.Fprintln(os.Stderr, "No changes detected. Make some changes first, then run cast.")
			return nil
		}

		if output == "" {
			output = "cast.gif"
		}

		f, err := os.Create(output)
		if err != nil {
			return err
		}
		defer f.Close()

		fmt.Fprintf(os.Stderr, "Generating: %s (%d files, +%d -%d)\n",
			summary.Title,
			summary.Stats.FilesChanged,
			summary.Stats.Additions,
			summary.Stats.Deletions,
		)

		if err := render.RenderDiffGIF(summary, f); err != nil {
			os.Remove(output)
			return err
		}

		info, _ := f.Stat()
		size := "?"
		if info != nil {
			size = storage.FormatSize(info.Size())
		}
		fmt.Fprintf(os.Stderr, "Saved: %s (%s)\n", output, size)
		return nil
	},
}

func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func init() {
	rootCmd.Version = version
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.Flags().StringP("output", "o", "", "output file (default: cast.gif)")
	rootCmd.Flags().StringP("base", "b", "", "base ref to diff against (e.g. main)")
}
