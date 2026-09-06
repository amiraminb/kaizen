package cmd

import (
	"fmt"
	"os"

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

const longHelp = `Kaizen tracks daily habits in plain JSON files on your own machine.

GETTING STARTED
  kaizen init                    create the data directory and its files
  kaizen new "Read daily"        add a habit; --start backdates it
  kaizen                         open today's checklist

DAILY USE
  Run kaizen with no arguments to get today's checklist. Space cycles the habit
  under the cursor through done, skipped and cleared; d, s and c set a state
  directly; c also clears something you recorded by mistake. Enter saves only the
  rows you actually changed, and esc discards everything.

  For one habit the express lane is faster:

  kaizen done read               mark it done for today
  kaizen skip read               a rest day you chose, on purpose
  kaizen done read gym           several at once
  kaizen done read -d yesterday  backfill; also -d -3 or -d 2026-08-30
  kaizen done read -n "on train" attach a note

HOW A DAY IS COUNTED
  done      you checked in: the streak grows
  skipped   a rest day you chose: the streak survives and the day is left out of
            your completion rate entirely
  pending   today, nothing recorded yet: never counts against you
  missed    a past day with nothing recorded: this is what breaks a streak
  n/a       before the habit's start date

  You never record a miss. Leave the day alone and it becomes one once the day is
  over. Do not use skip for a day you meant to do and did not, or your completion
  rate will flatter you.

REVIEWING
  kaizen report                  completion and streaks for this month so far
  kaizen report 30d              the last 30 days
  kaizen report lastmonth        a named period
  kaizen report entries read 7d  every check-in with its notes

ADDRESSING HABITS
  Any unambiguous prefix of a slug works, so "kaizen done med" finds meditate.
  Renaming with kaizen edit never changes the slug unless you pass --slug.

WHERE YOUR DATA LIVES
  ~/Documents/.kaizen by default. Set KAIZEN_DATA_DIR to an absolute path to keep
  it somewhere else, such as a synced or version-controlled folder. Every write is
  atomic, and the files are sorted so they diff cleanly in git.`

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
			return printToday(cmd)
		}

		outcome, err := tui.RunChecklist()
		if outcome.Applied > 0 {
			fmt.Fprintf(out, "recorded %d change(s)\n", outcome.Applied)
		}
		if err != nil {
			return err
		}
		if !outcome.Started {
			return printToday(cmd)
		}
		if outcome.Applied == 0 {
			fmt.Fprintln(out, "no changes")
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func printToday(cmd *cobra.Command) error {
	report, err := svc.Today(1)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprint(out, render.New(out).Today(report.Summaries))
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
