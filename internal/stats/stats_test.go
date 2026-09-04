package stats

import (
	"strings"
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
)

const habitID = "hab_read"

func day(date string) time.Time {
	parsed, err := time.ParseInLocation(model.DateLayout, date, time.UTC)
	if err != nil {
		panic(err)
	}
	return parsed
}

func habit(startDate, archivedAt string) model.Habit {
	return model.Habit{
		ID:         habitID,
		Slug:       "read",
		Name:       "Read daily",
		Schedule:   model.Schedule{Kind: model.ScheduleDaily},
		StartDate:  startDate,
		ArchivedAt: archivedAt,
	}
}

// Each rune is one day starting at from: v done, s skipped, . no entry.
func indexFrom(from string, pattern string) Index {
	var entries []model.Entry
	start := day(from)

	for offset, symbol := range pattern {
		date := start.AddDate(0, 0, offset).Format(model.DateLayout)
		switch symbol {
		case 'v':
			entries = append(entries, model.Entry{HabitID: habitID, Date: date, Status: model.StatusDone})
		case 's':
			entries = append(entries, model.Entry{HabitID: habitID, Date: date, Status: model.StatusSkipped})
		}
	}
	return NewIndex(entries)
}

func glyphs(cells []Cell) string {
	var out strings.Builder
	for _, cell := range cells {
		switch cell.Status {
		case model.DayDone:
			out.WriteByte('v')
		case model.DaySkipped:
			out.WriteByte('s')
		case model.DayPending:
			out.WriteByte('o')
		case model.DayMiss:
			out.WriteByte('x')
		default:
			out.WriteByte('.')
		}
	}
	return out.String()
}

func TestResolveFiveStates(t *testing.T) {
	asOf := "2026-09-10"
	index := indexFrom("2026-09-08", "vs")

	tests := []struct {
		name  string
		habit model.Habit
		date  string
		want  model.DayStatus
	}{
		{name: "done", habit: habit("2026-09-01", ""), date: "2026-09-08", want: model.DayDone},
		{name: "skipped", habit: habit("2026-09-01", ""), date: "2026-09-09", want: model.DaySkipped},
		{name: "today with no entry is pending", habit: habit("2026-09-01", ""), date: asOf, want: model.DayPending},
		{name: "past day with no entry is a miss", habit: habit("2026-09-01", ""), date: "2026-09-07", want: model.DayMiss},
		{name: "before the start date", habit: habit("2026-09-08", ""), date: "2026-09-07", want: model.DayNotApplicable},
		{name: "after archiving", habit: habit("2026-09-01", "2026-09-05T10:00:00Z"), date: "2026-09-07", want: model.DayNotApplicable},
		{name: "the archive day itself still counts", habit: habit("2026-09-01", "2026-09-08T10:00:00Z"), date: "2026-09-08", want: model.DayDone},
		{name: "future day", habit: habit("2026-09-01", ""), date: "2026-09-11", want: model.DayNotApplicable},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := Resolve(tc.habit, index, tc.date, asOf)
			if got != tc.want {
				t.Errorf("Resolve(%s) = %v, want %v", tc.date, got, tc.want)
			}
		})
	}
}

func TestGridWalksEveryDayInRange(t *testing.T) {
	index := indexFrom("2026-09-01", "vvxsv")

	cells := Grid(habit("2026-09-01", ""), index, day("2026-09-01"), day("2026-09-06"), "2026-09-06")
	if got, want := glyphs(cells), "vvxsvo"; got != want {
		t.Errorf("grid = %q, want %q", got, want)
	}
	if len(cells) != 6 {
		t.Errorf("grid has %d cells, want 6", len(cells))
	}
}

func TestStreaksIgnoreSkippedAndPending(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		asOf        string
		wantCurrent int
		wantLongest int
	}{
		{name: "unbroken run", pattern: "vvvvv", asOf: "2026-09-05", wantCurrent: 5, wantLongest: 5},
		{name: "skip keeps the run alive", pattern: "vvsvv", asOf: "2026-09-05", wantCurrent: 4, wantLongest: 4},
		{name: "miss breaks the run", pattern: "vv.vv", asOf: "2026-09-05", wantCurrent: 2, wantLongest: 2},
		{name: "longest is behind the current", pattern: "vvvv.v", asOf: "2026-09-06", wantCurrent: 1, wantLongest: 4},
		{name: "pending today does not break the run", pattern: "vvvv.", asOf: "2026-09-05", wantCurrent: 4, wantLongest: 4},
		{name: "pending today does not extend the run", pattern: "vvv..", asOf: "2026-09-05", wantCurrent: 0, wantLongest: 3},
		{name: "nothing recorded at all", pattern: ".....", asOf: "2026-09-05", wantCurrent: 0, wantLongest: 0},
		{name: "only skips", pattern: "sssss", asOf: "2026-09-05", wantCurrent: 0, wantLongest: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			index := indexFrom("2026-09-01", tc.pattern)
			asOf := day(tc.asOf)

			summary := Summarize(habit("2026-09-01", ""), index, day("2026-09-01"), asOf, asOf)
			if summary.CurrentStreak != tc.wantCurrent {
				t.Errorf("current streak = %d, want %d (grid %q)", summary.CurrentStreak, tc.wantCurrent, glyphs(summary.Cells))
			}
			if summary.LongestStreak != tc.wantLongest {
				t.Errorf("longest streak = %d, want %d (grid %q)", summary.LongestStreak, tc.wantLongest, glyphs(summary.Cells))
			}
		})
	}
}

// A window shorter than the habit's history must not cap the streak at the window.
func TestStreaksSpanTheWholeHistoryNotTheWindow(t *testing.T) {
	index := indexFrom("2026-08-01", strings.Repeat("v", 40))
	asOf := day("2026-09-09")

	summary := Summarize(habit("2026-08-01", ""), index, asOf.AddDate(0, 0, -4), asOf, asOf)
	if len(summary.Cells) != 5 {
		t.Errorf("displayed %d cells, want 5", len(summary.Cells))
	}
	if summary.CurrentStreak != 40 {
		t.Errorf("current streak = %d, want 40 despite a 5 day window", summary.CurrentStreak)
	}
	if summary.LongestStreak != 40 {
		t.Errorf("longest streak = %d, want 40 despite a 5 day window", summary.LongestStreak)
	}
}

// Habits of different ages must produce equal-length strips, or the columns after
// the strip shift per row.
func TestWindowWiderThanTheHabitStillFillsEveryCell(t *testing.T) {
	asOf := day("2026-09-10")
	from := asOf.AddDate(0, 0, -9)

	young := Summarize(habit("2026-09-08", ""), indexFrom("2026-09-08", "vv"), from, asOf, asOf)
	old := Summarize(habit("2026-08-01", ""), indexFrom("2026-09-01", "vvvvvvvvvv"), from, asOf, asOf)

	if len(young.Cells) != 10 || len(old.Cells) != 10 {
		t.Fatalf("cell counts = %d and %d, want 10 each", len(young.Cells), len(old.Cells))
	}
	if got, want := glyphs(young.Cells), ".......vvo"; got != want {
		t.Errorf("young grid = %q, want %q", got, want)
	}
	if young.Missed != 0 {
		t.Errorf("missed = %d, want 0 for days before the habit existed", young.Missed)
	}
}

func TestSummarizeCountsAndCompletion(t *testing.T) {
	index := indexFrom("2026-09-01", "vvs..v")
	asOf := day("2026-09-06")

	summary := Summarize(habit("2026-09-01", ""), index, day("2026-09-01"), asOf, asOf)
	if got, want := glyphs(summary.Cells), "vvsxxv"; got != want {
		t.Fatalf("grid = %q, want %q", got, want)
	}
	if summary.Done != 3 {
		t.Errorf("done = %d, want 3", summary.Done)
	}
	if summary.Skipped != 1 {
		t.Errorf("skipped = %d, want 1", summary.Skipped)
	}
	if summary.Missed != 2 {
		t.Errorf("missed = %d, want 2", summary.Missed)
	}
	if got := summary.Completion(); got < 0.59 || got > 0.61 {
		t.Errorf("completion = %.3f, want 0.60 from 3 done out of 5 scored days", got)
	}
}

func TestCompletionIgnoresSkippedDays(t *testing.T) {
	index := indexFrom("2026-09-01", "vsss")
	asOf := day("2026-09-04")

	summary := Summarize(habit("2026-09-01", ""), index, day("2026-09-01"), asOf, asOf)
	if got := summary.Completion(); got != 1 {
		t.Errorf("completion = %.3f, want 1 because skipped days are not failures", got)
	}
}

func TestSummarizeReportsTodaysStatus(t *testing.T) {
	asOf := day("2026-09-05")

	pending := Summarize(habit("2026-09-01", ""), indexFrom("2026-09-01", "vvvv"), day("2026-09-01"), asOf, asOf)
	if pending.Today != model.DayPending {
		t.Errorf("today = %v, want pending", pending.Today)
	}

	done := Summarize(habit("2026-09-01", ""), indexFrom("2026-09-01", "vvvvv"), day("2026-09-01"), asOf, asOf)
	if done.Today != model.DayDone {
		t.Errorf("today = %v, want done", done.Today)
	}
}

func TestArchivedHabitStopsGeneratingMisses(t *testing.T) {
	index := indexFrom("2026-09-01", "vvv")
	asOf := day("2026-09-10")

	summary := Summarize(habit("2026-09-01", "2026-09-03T18:00:00Z"), index, day("2026-09-01"), asOf, asOf)
	if summary.Missed != 0 {
		t.Errorf("missed = %d, want 0 after archiving", summary.Missed)
	}
	if summary.CurrentStreak != 3 {
		t.Errorf("current streak = %d, want the run at archive time, 3", summary.CurrentStreak)
	}
	if summary.Today != model.DayNotApplicable {
		t.Errorf("today = %v, want n/a for an archived habit", summary.Today)
	}
}

func TestNewIndexKeepsTheLastEntryPerDay(t *testing.T) {
	index := NewIndex([]model.Entry{
		{HabitID: habitID, Date: "2026-09-01", Status: model.StatusDone},
		{HabitID: habitID, Date: "2026-09-01", Status: model.StatusSkipped},
	})

	entry, ok := index.Lookup(habitID, "2026-09-01")
	if !ok {
		t.Fatal("Lookup did not find the entry")
	}
	if entry.Status != model.StatusSkipped {
		t.Errorf("status = %q, want the last entry to win", entry.Status)
	}
}
