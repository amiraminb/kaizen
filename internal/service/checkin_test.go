package service

import (
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/repository"
)

func seedHabit(t *testing.T, repo *repository.FileRepository, slug, startDate string) model.Habit {
	t.Helper()

	habit := model.Habit{
		ID:        "hab_" + slug,
		Slug:      slug,
		Name:      slug,
		Schedule:  model.Schedule{Kind: model.ScheduleDaily},
		StartDate: startDate,
	}
	if err := repo.SaveHabits([]model.Habit{habit}); err != nil {
		t.Fatalf("SaveHabits returned error: %v", err)
	}
	return habit
}

func TestCheckInCreatesAnEntry(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	result, err := svc.CheckIn("read", "", model.StatusDone, "")
	if err != nil {
		t.Fatalf("CheckIn returned error: %v", err)
	}
	if !result.Created {
		t.Error("first check-in should report Created")
	}
	if result.Entry.Date != "2026-09-03" {
		t.Errorf("entry date = %q, want %q", result.Entry.Date, "2026-09-03")
	}
	if result.Entry.Status != model.StatusDone {
		t.Errorf("entry status = %q, want %q", result.Entry.Status, model.StatusDone)
	}

	entries, err := repo.LoadEntries()
	if err != nil {
		t.Fatalf("LoadEntries returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("stored %d entries, want 1", len(entries))
	}
}

func TestCheckInIsIdempotent(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	if _, err := svc.CheckIn("read", "", model.StatusDone, "first"); err != nil {
		t.Fatalf("first CheckIn returned error: %v", err)
	}
	result, err := svc.CheckIn("read", "", model.StatusSkipped, "")
	if err != nil {
		t.Fatalf("second CheckIn returned error: %v", err)
	}
	if result.Created {
		t.Error("second check-in on the same day should not report Created")
	}
	if result.Entry.Status != model.StatusSkipped {
		t.Errorf("status = %q, want %q", result.Entry.Status, model.StatusSkipped)
	}
	if result.Entry.Note != "first" {
		t.Errorf("note = %q, want the original note to survive an empty note", result.Entry.Note)
	}

	entries, err := repo.LoadEntries()
	if err != nil {
		t.Fatalf("LoadEntries returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("stored %d entries, want 1", len(entries))
	}
}

func TestCheckInUsesTheDayStartCutoff(t *testing.T) {
	now := time.Date(2026, 9, 4, 1, 30, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	result, err := svc.CheckIn("read", "", model.StatusDone, "")
	if err != nil {
		t.Fatalf("CheckIn returned error: %v", err)
	}
	if result.Entry.Date != "2026-09-03" {
		t.Errorf("entry date = %q, want the previous day %q", result.Entry.Date, "2026-09-03")
	}
}

func TestCheckInBackfillsAndRejectsInvalidDates(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		date     string
		wantDate string
		wantErr  bool
	}{
		{name: "yesterday", date: "yesterday", wantDate: "2026-09-02"},
		{name: "relative days", date: "-10", wantDate: "2026-08-24"},
		{name: "explicit past date", date: "2026-08-15", wantDate: "2026-08-15"},
		{name: "start date boundary", date: "2026-08-01", wantDate: "2026-08-01"},
		{name: "before the habit existed", date: "2026-07-31", wantErr: true},
		{name: "future date", date: "2026-09-04", wantErr: true},
		{name: "unparseable date", date: "next tuesday", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newTestService(t, now)
			seedHabit(t, repo, "read", "2026-08-01")

			result, err := svc.CheckIn("read", tc.date, model.StatusDone, "")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("CheckIn(date=%q) = %q, want error", tc.date, result.Entry.Date)
				}
				return
			}
			if err != nil {
				t.Fatalf("CheckIn(date=%q) returned error: %v", tc.date, err)
			}
			if result.Entry.Date != tc.wantDate {
				t.Errorf("entry date = %q, want %q", result.Entry.Date, tc.wantDate)
			}
		})
	}
}

func TestCheckInRejectsBadStatusAndHabit(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	if _, err := svc.CheckIn("read", "", "maybe", ""); err == nil {
		t.Error("an unknown status must be rejected")
	}
	if _, err := svc.CheckIn("nope", "", model.StatusDone, ""); err == nil {
		t.Error("an unknown habit must be rejected")
	}
}

// The checklist opens a past day through Day() so a wrongly recorded check-in can be
// cleared there; the status it shows must be that day's, not today's.
func TestDayReportsThePastDaysOwnStatus(t *testing.T) {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-09-01")

	if _, err := svc.CheckIn("read", "2026-09-03", model.StatusDone, ""); err != nil {
		t.Fatalf("CheckIn returned error: %v", err)
	}

	tests := []struct {
		name  string
		input string
		want  model.DayStatus
	}{
		{name: "the recorded day", input: "2026-09-03", want: model.DayDone},
		{name: "an untouched past day", input: "2026-09-02", want: model.DayMiss},
		{name: "today", input: "", want: model.DayPending},
		{name: "before the habit started", input: "2026-08-30", want: model.DayNotApplicable},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report, err := svc.Day(tc.input)
			if err != nil {
				t.Fatalf("Day(%q) returned error: %v", tc.input, err)
			}
			if len(report.Summaries) != 1 {
				t.Fatalf("got %d summaries, want 1", len(report.Summaries))
			}

			date := report.From.Format(model.DateLayout)
			if got := report.Summaries[0].StatusOn(date); got != tc.want {
				t.Errorf("status on %s = %v, want %v", date, got, tc.want)
			}
		})
	}
}

func TestDayRejectsAFutureDate(t *testing.T) {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-09-01")

	if _, err := svc.Day("2026-09-06"); err == nil {
		t.Error("opening a future day must be rejected")
	}
}

func TestClearingAPastDayRemovesTheEntry(t *testing.T) {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-09-01")

	if _, err := svc.CheckIn("read", "2026-09-03", model.StatusDone, ""); err != nil {
		t.Fatalf("CheckIn returned error: %v", err)
	}

	result, err := svc.Undo("read", "2026-09-03")
	if err != nil {
		t.Fatalf("Undo returned error: %v", err)
	}
	if !result.Removed {
		t.Fatal("Undo should have removed the past entry")
	}

	report, err := svc.Day("2026-09-03")
	if err != nil {
		t.Fatalf("Day returned error: %v", err)
	}
	if got := report.Summaries[0].StatusOn("2026-09-03"); got != model.DayMiss {
		t.Errorf("status = %v, want the day to fall back to a miss", got)
	}
}

func TestUndo(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)
	seedHabit(t, repo, "read", "2026-08-01")

	if _, err := svc.CheckIn("read", "", model.StatusDone, ""); err != nil {
		t.Fatalf("CheckIn returned error: %v", err)
	}

	result, err := svc.Undo("read", "")
	if err != nil {
		t.Fatalf("Undo returned error: %v", err)
	}
	if !result.Removed {
		t.Error("Undo should report Removed for an existing entry")
	}

	entries, err := repo.LoadEntries()
	if err != nil {
		t.Fatalf("LoadEntries returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("stored %d entries after undo, want 0", len(entries))
	}

	again, err := svc.Undo("read", "")
	if err != nil {
		t.Fatalf("second Undo returned error: %v", err)
	}
	if again.Removed {
		t.Error("Undo on a day with no entry should report Removed false, not an error")
	}
}
