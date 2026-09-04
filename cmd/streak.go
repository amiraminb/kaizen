package cmd

import (
	"fmt"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/render"
	"github.com/amiraminb/kaizen/internal/service"
	"github.com/amiraminb/kaizen/internal/stats"
	"github.com/spf13/cobra"
)

var streakWindow int

var streakCmd = &cobra.Command{
	Use:   "streak [habit]",
	Short: "Show current and longest streaks",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := svc.Today(streakWindow)
		if err != nil {
			return err
		}

		summaries := report.Summaries
		if len(args) == 1 {
			summaries, err = filterSummaries(summaries, args[0])
			if err != nil {
				return err
			}
		}

		out := cmd.OutOrStdout()
		fmt.Fprint(out, render.New(out).Streaks(summaries))
		return nil
	},
}

func filterSummaries(summaries []stats.Summary, habitInput string) ([]stats.Summary, error) {
	habits := make([]model.Habit, len(summaries))
	for i, summary := range summaries {
		habits[i] = summary.Habit
	}

	habit, err := service.ResolveHabit(habits, habitInput, false)
	if err != nil {
		return nil, err
	}
	for _, summary := range summaries {
		if summary.Habit.ID == habit.ID {
			return []stats.Summary{summary}, nil
		}
	}
	return nil, fmt.Errorf("no habit matches %q", habitInput)
}

func init() {
	streakCmd.Flags().IntVarP(&streakWindow, "days", "N", 30, "number of recent days to show in the strip")
	rootCmd.AddCommand(streakCmd)
}
