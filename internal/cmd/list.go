package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/adamgold/agentcast/internal/storage"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List local recordings",
	RunE: func(cmd *cobra.Command, args []string) error {
		recs, err := storage.List()
		if err != nil {
			return err
		}
		if len(recs) == 0 {
			fmt.Println("No recordings found. Run `cast rec` to create one.")
			return nil
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "ID\tTITLE\tDURATION\tSIZE\tDATE")
		for _, r := range recs {
			title := r.Header.Title
			if title == "" {
				title = "-"
			}
			date := time.Unix(r.Header.Timestamp, 0).Format("2006-01-02 15:04")
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
				r.ID,
				title,
				storage.FormatDuration(r.Duration),
				storage.FormatSize(r.Size),
				date,
			)
		}
		return tw.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
