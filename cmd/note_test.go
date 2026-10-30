package cmd

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/repository"
	"github.com/amiraminb/kaizen/internal/service"
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

func TestNoteReadPrintsTheWholeHabitDocument(t *testing.T) {
	oldRepo, oldSvc := repo, svc
	t.Cleanup(func() {
		repo, svc = oldRepo, oldSvc
	})

	repo = repository.NewFileRepositoryAt(t.TempDir())
	svc = service.NewWithClock(repo, func() time.Time { return time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC) })
	seedNoteCommandData(t)
	if err := repo.SaveNotes([]model.Note{{HabitID: "hab_read", Text: "line one\nline two"}}); err != nil {
		t.Fatalf("SaveNotes returned error: %v", err)
	}

	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	if err := printNote(cmd, "read"); err != nil {
		t.Fatalf("printNote returned error: %v", err)
	}
	if output.String() != "line one\nline two\n" {
		t.Errorf("printed note = %q, want the full document", output.String())
	}
}

func TestNotesEditSavesTheHabitDocument(t *testing.T) {
	oldRepo, oldSvc := repo, svc
	t.Cleanup(func() {
		repo, svc = oldRepo, oldSvc
		notesEdit = false
	})

	repo = repository.NewFileRepositoryAt(t.TempDir())
	svc = service.NewWithClock(repo, func() time.Time { return time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC) })
	seedNoteCommandData(t)
	editor := t.TempDir() + "/editor.sh"
	if err := os.WriteFile(editor, []byte("#!/bin/sh\nprintf 'edited document' > \"$1\"\n"), 0o700); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", editor)

	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBuffer(nil))
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	notesEdit = true
	if err := notesCmd.RunE(cmd, []string{"read"}); err != nil {
		t.Fatalf("notes edit returned error: %v", err)
	}

	notes, err := repo.LoadNotes()
	if err != nil {
		t.Fatalf("LoadNotes returned error: %v", err)
	}
	if len(notes) != 1 || notes[0].Text != "edited document" {
		t.Errorf("notes = %+v, want one edited document", notes)
	}
}

func seedNoteCommandData(t *testing.T) {
	t.Helper()
	if err := repo.SaveConfig(model.DefaultConfig()); err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}
	if err := repo.SaveHabits([]model.Habit{{
		ID: "hab_read", Slug: "read", Name: "Read", Schedule: model.Schedule{Kind: model.ScheduleDaily},
		StartDate: "2026-09-01", CreatedAt: "2026-09-01T12:00:00Z", UpdatedAt: "2026-09-01T12:00:00Z",
	}}); err != nil {
		t.Fatalf("SaveHabits returned error: %v", err)
	}
}
