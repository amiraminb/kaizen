package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var urgeCmd = &cobra.Command{
	Use:   "urge",
	Short: "Pause and reflect before giving in to an urge",
	Long: `Pause and reflect before giving in to an urge.

Answer these questions before you decide what to do:
1. What is this thought about?
2. What are the feelings I have about this thought?
3. What are the physical feelings I have?
4. What are the thoughts about the main thought?
5. What do you want to do about this?`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "Before giving in to an urge, pause and answer:")
		fmt.Fprintln(cmd.OutOrStdout(), "1. What is this thought about?")
		fmt.Fprintln(cmd.OutOrStdout(), "2. What are the feelings I have about this thought?")
		fmt.Fprintln(cmd.OutOrStdout(), "3. What are the physical feelings I have?")
		fmt.Fprintln(cmd.OutOrStdout(), "4. What are the thoughts about the main thought?")
		fmt.Fprintln(cmd.OutOrStdout(), "5. What do you want to do about this?")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Take a note of your answers with: kaizen notes <habit> -e")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(urgeCmd)
}
