package cmd

import (
	"fmt"

	"github.com/amiraminb/kaizen/internal/daterange"
	"github.com/amiraminb/kaizen/internal/render"
	"github.com/spf13/cobra"
)

var notesCmd = &cobra.Command{
	Use:     "notes [habit] [range]",
	Aliases: []string{"journal"},
	Short:   "Show saved habit notes",
	Args:    cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		asOf, err := svc.AsOf()
		if err != nil {
			return err
		}

		habitInput, rangeInput := splitEntriesArgs(args, asOf)
		from, to, err := daterange.Resolve(rangeInput, asOf)
		if err != nil {
			return err
		}

		rows, err := svc.Notes(habitInput, from, to)
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), render.New(cmd.OutOrStdout()).Notes(rows))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(notesCmd)
}
