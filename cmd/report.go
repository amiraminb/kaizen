package cmd

import (
	"fmt"
	"strings"

	"github.com/amiraminb/kaizen/internal/daterange"
	"github.com/amiraminb/kaizen/internal/render"
	"github.com/spf13/cobra"
)

var reportAll bool

var reportCmd = &cobra.Command{
	Use:   "report [habit] [range]",
	Short: "Show per-habit completion and streaks",
	Long: fmt.Sprintf(`Show per-habit completion and streaks over a range.

Ranges: %s, Nd for the last N days, YYYY-MM-DD, or YYYY-MM-DD..YYYY-MM-DD.
With no arguments, shows all active habits for the current month. A single
argument is a date or range. Supply a habit slug (or unambiguous prefix) followed
by a date or range to show only that habit.
Each habit's days begin at its start date or the range start, whichever is later.

Short ranges show one row per Sunday–Saturday calendar week, labeled with its dates
and weekdays. The first and last weeks may be partial. Year ranges and ranges
longer than 31 days show one row per month with its completion rate. Output fits
the terminal width. When selecting a habit, each monthly section also shows that
month's weekly day-by-day rows.
Completion is done / (done + missed), excluding skipped and pending days. Each
percentage shows the completed-day count and denominator for its own period.
Totals cover the selected range; streak today and best lifetime cover the habit's
whole history. A rate of — means there are no done or missed days to score.`, strings.Join(daterange.Names(), ", ")),
	Example: `  kaizen report month
  kaizen report read month
  kaizen report read 2026-10-31
  kaizen report read year`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		asOf, err := svc.AsOf()
		if err != nil {
			return err
		}

		var habitInput, rangeInput string
		switch len(args) {
		case 1:
			rangeInput = args[0]
		case 2:
			habitInput, rangeInput = args[0], args[1]
		}
		rangeInput = strings.ToLower(strings.TrimSpace(rangeInput))
		from, to, err := daterange.Resolve(rangeInput, asOf)
		if err != nil {
			return err
		}

		report, err := svc.Summaries(habitInput, asOf, from, to, reportAll)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprint(out, render.New(out).Report(report.Summaries, render.ReportOptions{
			From: from, To: to, AsOf: asOf,
			Monthly:      rangeInput == "year" || rangeInput == "lastyear",
			ExpandMonths: strings.TrimSpace(habitInput) != "",
		}))
		return nil
	},
}

func init() {
	reportCmd.Flags().BoolVar(&reportAll, "all", false, "include retired habits")
	rootCmd.AddCommand(reportCmd)
}
