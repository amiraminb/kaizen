package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var noteRead bool

var noteCmd = &cobra.Command{
	Use:   "note <habit>",
	Short: "Edit or read a habit note",
	Long: `Edit one note document for a habit without changing its check-in status.

	With no flags, the default editor opens with the habit's existing note, if any.
Use --read/-r to print the note instead. The editor is selected from VISUAL, then
EDITOR, and falls back to vi.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if noteRead {
			return printNote(cmd, args[0])
		}
		return editHabitNote(cmd, args[0])
	},
}

func init() {
	noteCmd.Flags().BoolVarP(&noteRead, "read", "r", false, "print the habit note instead of opening the editor")
	rootCmd.AddCommand(noteCmd)
}

func editHabitNote(cmd *cobra.Command, habitInput string) error {
	existing, err := svc.NoteFor(habitInput)
	if err != nil {
		return err
	}
	text, err := editNote(cmd, existing.Text)
	if err != nil {
		return err
	}
	result, err := svc.SaveNote(habitInput, text)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "saved note for %s\n", result.Habit.Slug)
	return nil
}

func printNote(cmd *cobra.Command, habitInput string) error {
	note, err := svc.NoteFor(habitInput)
	if err != nil {
		return err
	}
	if note.Text == "" {
		return fmt.Errorf("no note for %q", habitInput)
	}
	fmt.Fprintln(cmd.OutOrStdout(), note.Text)
	return nil
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
