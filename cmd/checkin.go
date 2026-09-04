package cmd

import (
	"fmt"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/spf13/cobra"
)

var (
	checkInDate string
	checkInNote string
	undoDate    string
)

var doneCmd = &cobra.Command{
	Use:   "done <habit>...",
	Short: "Mark habits done",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheckIn(cmd, args, model.StatusDone)
	},
}

var skipCmd = &cobra.Command{
	Use:   "skip <habit>...",
	Short: "Mark habits deliberately skipped, keeping the streak alive",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheckIn(cmd, args, model.StatusSkipped)
	},
}

var undoCmd = &cobra.Command{
	Use:   "undo <habit>...",
	Short: "Remove a check-in",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, habitInput := range args {
			result, err := svc.Undo(habitInput, undoDate)
			if err != nil {
				return err
			}
			if result.Removed {
				fmt.Fprintf(cmd.OutOrStdout(), "removed %s on %s\n", result.Habit.Slug, result.Date)
				continue
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s had nothing recorded on %s\n", result.Habit.Slug, result.Date)
		}
		return nil
	},
}

func runCheckIn(cmd *cobra.Command, habitInputs []string, status string) error {
	for _, habitInput := range habitInputs {
		result, err := svc.CheckIn(habitInput, checkInDate, status, checkInNote)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s on %s\n", result.Habit.Slug, result.Entry.Status, result.Entry.Date)
	}
	return nil
}

func init() {
	for _, command := range []*cobra.Command{doneCmd, skipCmd} {
		command.Flags().StringVarP(&checkInDate, "date", "d", "", "date to record: YYYY-MM-DD, today, yesterday or -N")
		command.Flags().StringVarP(&checkInNote, "note", "n", "", "note to attach to the check-in")
	}
	undoCmd.Flags().StringVarP(&undoDate, "date", "d", "", "date to clear: YYYY-MM-DD, today, yesterday or -N")

	rootCmd.AddCommand(doneCmd, skipCmd, undoCmd)
}
