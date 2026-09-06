package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/amiraminb/kaizen/internal/daterange"
	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/render"
	"github.com/spf13/cobra"
)

var reportAll bool

var reportCmd = &cobra.Command{
	Use:   "report [range]",
	Short: "Show per-habit completion and streaks",
	Long: fmt.Sprintf(`Show per-habit completion and streaks over a range.

Ranges: %s, Nd for the last N days, YYYY-MM-DD, or YYYY-MM-DD..YYYY-MM-DD.
Defaults to the current month.

Subcommands:
  entries [habit] [range]   one row per check-in, with notes`, strings.Join(daterange.Names(), ", ")),
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		asOf, err := svc.AsOf()
		if err != nil {
			return err
		}

		from, to, err := daterange.Resolve(firstOrEmpty(args), asOf)
		if err != nil {
			return err
		}

		report, err := svc.Summaries(from, to, reportAll)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprint(out, render.New(out).Report(report.Summaries, from.Format(model.DateLayout), to.Format(model.DateLayout)))
		return nil
	},
}

var reportEntriesCmd = &cobra.Command{
	Use:     "entries [habit] [range]",
	Aliases: []string{"log"},
	Short:   "Show one row per check-in",
	Args:    cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		asOf, err := svc.AsOf()
		if err != nil {
			return err
		}

		habitInput, rangeInput := splitEntriesArgs(args, asOf)

		from, to, err := daterange.Resolve(rangeInput, asOf)
		if err != nil {
			return err
		}

		rows, err := svc.EntryRows(habitInput, from, to)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprint(out, render.New(out).Entries(rows))
		return nil
	},
}

func firstOrEmpty(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

// A lone argument is read as a range when it parses as one, so both
// "report entries read" and "report entries 7d" work without a placeholder.
func splitEntriesArgs(args []string, asOf time.Time) (habit string, dateRange string) {
	switch len(args) {
	case 1:
		if _, _, err := daterange.Resolve(args[0], asOf); err == nil {
			return "", args[0]
		}
		return args[0], ""
	case 2:
		return args[0], args[1]
	default:
		return "", ""
	}
}

func init() {
	reportCmd.Flags().BoolVar(&reportAll, "all", false, "include retired habits")
	reportCmd.AddCommand(reportEntriesCmd)
	rootCmd.AddCommand(reportCmd)
}
