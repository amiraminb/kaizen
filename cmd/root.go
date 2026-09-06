package cmd

import (
	"fmt"
	"os"

	"github.com/amiraminb/kaizen/internal/repository"
	"github.com/amiraminb/kaizen/internal/service"
	"github.com/amiraminb/kaizen/internal/tui"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	fileRepo                       = repository.NewFileRepository()
	repo     repository.Repository = fileRepo
	svc                            = service.New(repo)
)

func init() {
	tui.Svc = svc
}

var version = "dev"

// main prints the error itself, so cobra's duplicate error line and its usage dump
// are suppressed to keep a failure message from being buried in a help wall.
var rootCmd = &cobra.Command{
	Use:           "kaizen",
	Short:         "Local-first habit tracker",
	Long:          "Local-first habit tracker.\n\nRun with no arguments to open today's checklist.",
	Version:       version,
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if !interactive() {
			return printToday(cmd)
		}

		applied, err := tui.RunChecklist()
		if err != nil {
			return err
		}
		if applied == 0 {
			fmt.Fprintln(out, "no changes")
			return nil
		}
		fmt.Fprintf(out, "recorded %d change(s)\n", applied)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

// Without a terminal the checklist cannot run at all, and bubbletea's own failure
// reads as a crash, so a piped or scripted invocation gets today's status instead.
func interactive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}
