package cmd

import (
	"fmt"
	"os"

	"github.com/amiraminb/kaizen/internal/clock"
	"github.com/amiraminb/kaizen/internal/render"
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

const longHelp = `Kaizen is a local-first habit tracker.

Usage:
  kaizen                         open today's checklist
  kaizen new <name>              create a habit
  kaizen done <habit>...         record completed habits
  kaizen skip <habit>...         record a deliberate rest day
  kaizen note <habit> [date]    write a note without changing completion
  kaizen notes [habit] [range]   review saved notes
  kaizen report [range]          view completion and streaks
  kaizen report entries ...      view check-ins and their notes
  kaizen edit [habit]            rename a habit or change its slug
  kaizen init                    initialize the data directory

Run kaizen <command> --help for command-specific options. Data is stored locally
in ~/Documents/.kaizen, or in KAIZEN_DATA_DIR when configured.`

// main prints the error itself, so cobra's duplicate error line and its usage dump
// are suppressed to keep a failure message from being buried in a help wall.
var rootCmd = &cobra.Command{
	Use:           "kaizen",
	Short:         "Local-first habit tracker",
	Long:          longHelp,
	Version:       version,
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if !interactive() {
			return printDay(cmd)
		}

		outcome, err := tui.RunChecklist(checklistDate)
		if outcome.Applied > 0 {
			fmt.Fprintf(out, "recorded %d change(s)\n", outcome.Applied)
		}
		if err != nil {
			return err
		}
		if !outcome.Started {
			return printDay(cmd)
		}
		if outcome.Applied == 0 {
			fmt.Fprintln(out, "no changes")
		}
		return nil
	},
}

var checklistDate string

func init() {
	rootCmd.Flags().StringVarP(&checklistDate, "date", "d", "", "day to open: YYYY-MM-DD, today, yesterday or -N")
}

func Execute() error {
	return rootCmd.Execute()
}

func printDay(cmd *cobra.Command) error {
	report, err := svc.Day(checklistDate)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprint(out, render.New(out).Day(report.Summaries, clock.DateOf(report.From)))
	return nil
}

// Both ends must be a terminal: stdin alone is not enough, because `kaizen > file`
// leaves stdin a TTY and would render the whole TUI into the file.
func interactive() bool {
	return isTerminal(os.Stdin) && isTerminal(os.Stdout)
}

func isTerminal(file *os.File) bool {
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}
