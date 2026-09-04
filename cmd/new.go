package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	newSlug  string
	newStart string
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a habit",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		habit, err := svc.CreateHabit(strings.Join(args, " "), newSlug, newStart)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "created %s (%s), starting %s\n", habit.Slug, habit.Name, habit.StartDate)
		return nil
	},
}

func init() {
	newCmd.Flags().StringVarP(&newSlug, "slug", "s", "", "short name used on the command line (derived from the name by default)")
	newCmd.Flags().StringVar(&newStart, "start", "", "first day the habit counts: YYYY-MM-DD, yesterday or -N (today by default)")
	rootCmd.AddCommand(newCmd)
}
