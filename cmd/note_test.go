package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestEditNoteUsesConfiguredEditor(t *testing.T) {
	editor := t.TempDir() + "/editor.sh"
	if err := os.WriteFile(editor, []byte("#!/bin/sh\nprintf 'written by editor' > \"$1\"\n"), 0o700); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", editor)

	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBuffer(nil))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	got, err := editNote(cmd, "old note")
	if err != nil {
		t.Fatalf("editNote returned error: %v", err)
	}
	if got != "written by editor" {
		t.Errorf("edited note = %q, want editor output", got)
	}
}
