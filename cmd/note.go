package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/amiraminb/kaizen/internal/clock"
	"github.com/spf13/cobra"
)

var (
	noteDate string
	noteText string
)

var noteCmd = &cobra.Command{
	Use:   "note <habit> [date] [text...]",
	Short: "Save a note for a habit on a day",
	Long: `Save a note for a habit without changing its check-in status.

The date defaults to today. A date can be YYYY-MM-DD, today, yesterday or -N.
With no text, the default editor opens with the existing note, if any. The editor
is selected from VISUAL, then EDITOR, and falls back to vi.
Pass the note as trailing text or with --text/-n for non-interactive use.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		asOf, err := svc.AsOf()
		if err != nil {
			return err
		}

		dateInput := noteDate
		textParts := args[1:]
		if dateInput == "" && len(textParts) > 0 {
			if _, err := clock.ParseDate(textParts[0], asOf); err == nil {
				dateInput = textParts[0]
				textParts = textParts[1:]
			}
		}
		if strings.TrimSpace(noteText) != "" && len(textParts) > 0 {
			return fmt.Errorf("note text was provided both as an argument and with --text")
		}
		text := noteText
		if strings.TrimSpace(text) == "" {
			text = strings.Join(textParts, " ")
		}
		if len(textParts) == 0 && strings.TrimSpace(noteText) == "" {
			existing, err := svc.NoteFor(args[0], dateInput)
			if err != nil {
				return err
			}
			text, err = editNote(cmd, existing.Text)
			if err != nil {
				return err
			}
		}

		result, err := svc.AddNote(args[0], dateInput, text)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "noted %s on %s\n", result.Habit.Slug, result.Note.Date)
		return nil
	},
}

func init() {
	noteCmd.Flags().StringVarP(&noteDate, "date", "d", "", "date to note: YYYY-MM-DD, today, yesterday or -N")
	noteCmd.Flags().StringVarP(&noteText, "text", "n", "", "note text")
	rootCmd.AddCommand(noteCmd)
}

func editNote(cmd *cobra.Command, initial string) (string, error) {
	temp, err := os.CreateTemp("", "kaizen-note-*.txt")
	if err != nil {
		return "", err
	}
	path := temp.Name()
	defer os.Remove(path)

	if _, err := temp.WriteString(initial); err != nil {
		temp.Close()
		return "", err
	}
	if err := temp.Close(); err != nil {
		return "", err
	}

	editorName := strings.TrimSpace(os.Getenv("VISUAL"))
	if editorName == "" {
		editorName = strings.TrimSpace(os.Getenv("EDITOR"))
	}
	if editorName == "" {
		editorName = "vi"
	}
	editorParts := strings.Fields(editorName)
	editor := exec.Command(editorParts[0], append(editorParts[1:], path)...)
	editor.Stdin = cmd.InOrStdin()
	editor.Stdout = cmd.OutOrStdout()
	editor.Stderr = cmd.ErrOrStderr()
	if err := editor.Run(); err != nil {
		return "", fmt.Errorf("editor %q failed: %w", editorName, err)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(contents)), nil
}
