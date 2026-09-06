package cmd

import (
	"fmt"
	"strings"

	"github.com/amiraminb/kaizen/internal/tui"
	"github.com/spf13/cobra"
)

var (
	newSlug  string
	newStart string
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a habit",
	Long: `Create a habit.

With no arguments it prompts for the name and slug.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) == 0 {
			habit, created, err := tui.RunHabitNew()
			if err != nil {
				return err
			}
			if !created {
				fmt.Fprintln(out, "cancelled")
				return nil
			}
			fmt.Fprintf(out, "created %s (%s), starting %s\n", habit.Slug, habit.Name, habit.StartDate)
			return nil
		}

		habit, err := svc.CreateHabit(strings.Join(args, " "), newSlug, newStart)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "created %s (%s), starting %s\n", habit.Slug, habit.Name, habit.StartDate)
		return nil
	},
}

func init() {
	newCmd.Flags().StringVarP(&newSlug, "slug", "s", "", "short name used on the command line (derived from the name by default)")
	newCmd.Flags().StringVar(&newStart, "start", "", "first day the habit counts: YYYY-MM-DD, yesterday or -N (today by default)")
	rootCmd.AddCommand(newCmd)
}
