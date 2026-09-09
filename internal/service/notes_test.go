package service

import (
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
)

func TestAddNoteDefaultsToTodayWithoutCheckingIn(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	result, err := svc.AddNote("read", "", "  felt focused  ")
	if err != nil {
		t.Fatalf("AddNote returned error: %v", err)
	}
	if result.Note.Date != "2026-09-03" {
		t.Errorf("note date = %q, want today", result.Note.Date)
	}
	if result.Note.Text != "felt focused" {
		t.Errorf("note text = %q, want trimmed text", result.Note.Text)
	}

	entries, err := repo.LoadEntries()
	if err != nil {
		t.Fatalf("LoadEntries returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("AddNote created %d check-ins, want 0", len(entries))
	}
	notes, err := repo.LoadNotes()
	if err != nil {
		t.Fatalf("LoadNotes returned error: %v", err)
	}
	if len(notes) != 1 || notes[0].Text != "felt focused" {
		t.Errorf("notes = %+v, want one standalone note", notes)
	}
}

func TestAddNoteUpdatesAnExistingCheckInWithoutChangingItsStatus(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	if _, err := svc.CheckIn("read", "", model.StatusSkipped, "old"); err != nil {
		t.Fatalf("CheckIn returned error: %v", err)
	}
	result, err := svc.AddNote("read", "", "new")
	if err != nil {
		t.Fatalf("AddNote returned error: %v", err)
	}
	if result.Created {
		t.Error("updating a check-in note should not report Created")
	}

	entries, err := repo.LoadEntries()
	if err != nil {
		t.Fatalf("LoadEntries returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("stored %d entries, want 1", len(entries))
	}
	if entries[0].Status != model.StatusSkipped || entries[0].Note != "new" {
		t.Errorf("entry = %+v, want skipped with the new note", entries[0])
	}
}

func TestAddNoteRejectsEmptyText(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	if _, err := svc.AddNote("read", "", "  "); err == nil {
		t.Error("an empty note must be rejected")
	}
}

func TestNoteForReturnsTheExistingNote(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	if _, err := svc.AddNote("read", "", "existing"); err != nil {
		t.Fatalf("AddNote returned error: %v", err)
	}
	note, err := svc.NoteFor("read", "")
	if err != nil {
		t.Fatalf("NoteFor returned error: %v", err)
	}
	if note.Date != "2026-09-03" || note.Text != "existing" {
		t.Errorf("note = %+v, want today's existing note", note)
	}
}
