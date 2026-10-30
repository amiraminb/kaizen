package cmd

import (
	"github.com/spf13/cobra"
)

var notesEdit bool

var notesCmd = &cobra.Command{
	Use:     "notes <habit>",
	Aliases: []string{"journal"},
	Short:   "Read or edit a habit note",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if notesEdit {
			return editHabitNote(cmd, args[0])
		}
		return printNote(cmd, args[0])
	},
}

func init() {
	notesCmd.Flags().BoolVarP(&notesEdit, "edit", "e", false, "open the habit note in the editor")
	rootCmd.AddCommand(notesCmd)
}
