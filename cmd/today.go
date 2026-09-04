package cmd

import (
	"fmt"

	"github.com/amiraminb/kaizen/internal/render"
	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's status for every habit",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := svc.Today(1)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprint(out, render.New(out).Today(report.Summaries))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}
