package service

import (
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
)

func TestSaveNoteCreatesOneDocumentWithoutCheckingIn(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	result, err := svc.SaveNote("read", "  felt focused  ")
	if err != nil {
		t.Fatalf("SaveNote returned error: %v", err)
	}
	if result.Note.Text != "felt focused" {
		t.Errorf("note text = %q, want trimmed text", result.Note.Text)
	}

	entries, err := repo.LoadEntries()
	if err != nil {
		t.Fatalf("LoadEntries returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("SaveNote created %d check-ins, want 0", len(entries))
	}
	notes, err := repo.LoadNotes()
	if err != nil {
		t.Fatalf("LoadNotes returned error: %v", err)
	}
	if len(notes) != 1 || notes[0].HabitID != result.Habit.ID || notes[0].Text != "felt focused" {
		t.Errorf("notes = %+v, want one habit note document", notes)
	}
}

func TestSaveNoteUpdatesTheExistingDocumentWithoutChangingCheckIns(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	if _, err := svc.CheckIn("read", "", model.StatusSkipped, "old"); err != nil {
		t.Fatalf("CheckIn returned error: %v", err)
	}
	result, err := svc.SaveNote("read", "new")
	if err != nil {
		t.Fatalf("SaveNote returned error: %v", err)
	}
	if !result.Created {
		t.Error("saving a habit note without an existing document should report Created")
	}

	entries, err := repo.LoadEntries()
	if err != nil {
		t.Fatalf("LoadEntries returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("stored %d entries, want 1", len(entries))
	}
	if entries[0].Status != model.StatusSkipped || entries[0].Note != "old" {
		t.Errorf("entry = %+v, want skipped with its original note", entries[0])
	}

	if _, err := svc.SaveNote("read", "newer"); err != nil {
		t.Fatalf("second SaveNote returned error: %v", err)
	}
	notes, err := repo.LoadNotes()
	if err != nil {
		t.Fatalf("LoadNotes returned error: %v", err)
	}
	if len(notes) != 1 || notes[0].Text != "newer" {
		t.Errorf("notes = %+v, want one updated document", notes)
	}
}

func TestSaveNoteRejectsEmptyText(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	if _, err := svc.SaveNote("read", "  "); err == nil {
		t.Error("an empty note must be rejected")
	}
}

func TestNoteForReturnsTheExistingNote(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	if _, err := svc.SaveNote("read", "existing"); err != nil {
		t.Fatalf("SaveNote returned error: %v", err)
	}
	note, err := svc.NoteFor("read")
	if err != nil {
		t.Fatalf("NoteFor returned error: %v", err)
	}
	if note.Text != "existing" {
		t.Errorf("note = %+v, want the existing habit note", note)
	}
}
