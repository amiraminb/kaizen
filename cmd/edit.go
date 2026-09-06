package cmd

import (
	"fmt"

	"github.com/amiraminb/kaizen/internal/tui"
	"github.com/spf13/cobra"
)

var (
	editName string
	editSlug string
)

var editCmd = &cobra.Command{
	Use:   "edit [habit]",
	Short: "Rename a habit or change its slug",
	Long: `Rename a habit or change its slug.

With no arguments it opens an interactive picker and form. With a habit and at
least one flag it runs without prompting.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) == 1 && (editName != "" || editSlug != "") {
			habit, err := svc.UpdateHabit(args[0], editName, editSlug)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "updated %s (%s)\n", habit.Slug, habit.Name)
			return nil
		}

		habit, changed, err := tui.RunHabitEdit()
		if err != nil {
			return err
		}
		if !changed {
			fmt.Fprintln(out, "cancelled")
			return nil
		}
		fmt.Fprintf(out, "updated %s (%s)\n", habit.Slug, habit.Name)
		return nil
	},
}

func init() {
	editCmd.Flags().StringVarP(&editName, "name", "n", "", "new display name")
	editCmd.Flags().StringVarP(&editSlug, "slug", "s", "", "new slug used on the command line")
	rootCmd.AddCommand(editCmd)
}
