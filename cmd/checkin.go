package cmd

import (
	"fmt"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/spf13/cobra"
)

var (
	checkInDate string
	checkInNote string
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

var missCmd = &cobra.Command{
	Use:   "miss <habit>...",
	Short: "Mark habits missed, breaking the streak",
	Long: `Record an explicit missed day for one or more habits, including today.

A miss counts against completion and breaks the streak. Past days without a
check-in already count as missed automatically. Use done or skip to correct a miss.`,
	Example: `  kaizen miss read gym
  kaizen miss read -d yesterday -n "ran out of time"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheckIn(cmd, args, model.StatusMissed)
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
	for _, command := range []*cobra.Command{doneCmd, skipCmd, missCmd} {
		command.Flags().StringVarP(&checkInDate, "date", "d", "", "date to record: YYYY-MM-DD, today, yesterday or -N")
		command.Flags().StringVarP(&checkInNote, "note", "n", "", "note to attach to the check-in")
	}
	rootCmd.AddCommand(doneCmd, skipCmd, missCmd)
}
