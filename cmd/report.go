package cmd

import (
	"fmt"
	"strings"

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
Each habit's days begin at its start date or the range start, whichever is later.`, strings.Join(daterange.Names(), ", ")),
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

		report, err := svc.Summaries(asOf, from, to, reportAll)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprint(out, render.New(out).Report(report.Summaries, from.Format(model.DateLayout), to.Format(model.DateLayout)))
		return nil
	},
}

func firstOrEmpty(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func init() {
	reportCmd.Flags().BoolVar(&reportAll, "all", false, "include retired habits")
	rootCmd.AddCommand(reportCmd)
}
