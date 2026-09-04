package cmd

import (
	"fmt"
	"io"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/repository"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the kaizen data directory and its data files",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		dir, err := fileRepo.EnsureDataDir()
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "data dir ready: %s\n", dir)

		seeds := []struct {
			fileName string
			save     func() error
		}{
			{repository.ConfigFileName, func() error { return repo.SaveConfig(model.DefaultConfig()) }},
			{repository.HabitsFileName, func() error { return repo.SaveHabits(nil) }},
			{repository.EntriesFileName, func() error { return repo.SaveEntries(nil) }},
		}

		for _, seed := range seeds {
			if err := seedFile(out, seed.fileName, seed.save); err != nil {
				return err
			}
		}
		return nil
	},
}

func seedFile(out io.Writer, fileName string, save func() error) error {
	exists, err := fileRepo.Exists(fileName)
	if err != nil {
		return err
	}
	if exists {
		fmt.Fprintf(out, "%s exists\n", fileName)
		return nil
	}
	if err := save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "created %s\n", fileName)
	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
