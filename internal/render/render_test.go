package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/stats"
)

// lastDay is the date of the final cell, so Day() can look a status up by date.
const lastDay = "2026-09-05"

func summary(slug, name string, current, longest int, today model.DayStatus, pattern string) stats.Summary {
	end, err := time.ParseInLocation(model.DateLayout, lastDay, time.UTC)
	if err != nil {
		panic(err)
	}

	var cells []stats.Cell
	for i, symbol := range pattern {
		var status model.DayStatus
		switch symbol {
		case 'v':
			status = model.DayDone
		case 's':
			status = model.DaySkipped
		case 'o':
			status = model.DayPending
		case 'x':
			status = model.DayMiss
		default:
			status = model.DayNotApplicable
		}

		date := end.AddDate(0, 0, i-len(pattern)+1).Format(model.DateLayout)
		cells = append(cells, stats.Cell{Date: date, Status: status})
	}

	built := stats.Summary{
		Habit:         model.Habit{Slug: slug, Name: name},
		Cells:         cells,
		Today:         today,
		CurrentStreak: current,
		LongestStreak: longest,
	}
	for _, cell := range cells {
		switch cell.Status {
		case model.DayDone:
			built.Done++
		case model.DaySkipped:
			built.Skipped++
		case model.DayMiss:
			built.Missed++
		}
	}
	return built
}

func TestPadAccountsForInvisibleEscapes(t *testing.T) {
	styled := New(&bytes.Buffer{}).done.Render("ok")

	if got := len(Pad("ok", 6)); got != 6 {
		t.Errorf("plain padded length = %d, want 6", got)
	}
	if !strings.HasSuffix(Pad(styled, 6), "    ") {
		t.Error("Pad must add padding based on visible width, not byte length")
	}
	if Pad("toolongvalue", 4) != "toolongvalue" {
		t.Error("Pad must not truncate text wider than the column")
	}
}

func TestDayRendersOneRowPerHabit(t *testing.T) {
	out := &bytes.Buffer{}
	rendered := New(out).Day([]stats.Summary{
		summary("read", "Read daily", 7, 9, model.DayDone, "vvv"),
		summary("gym", "Gym session", 0, 3, model.DayPending, "xxo"),
	}, lastDay)

	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("rendered %d lines, want 2 habits, a blank line and a footer:\n%s", len(lines), rendered)
	}
	if !strings.HasPrefix(lines[0], GlyphDone) {
		t.Errorf("first row = %q, want it to start with %q", lines[0], GlyphDone)
	}
	if !strings.HasPrefix(lines[1], GlyphPending) {
		t.Errorf("second row = %q, want it to start with %q", lines[1], GlyphPending)
	}
	if !strings.Contains(lines[0], "streak 7") {
		t.Errorf("first row = %q, want it to carry the current streak", lines[0])
	}
	if want := "2 habits · 1 done · 1 pending"; lines[3] != want {
		t.Errorf("footer = %q, want %q", lines[3], want)
	}
}

func TestDayWithNoHabitsSuggestsTheNextStep(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Day(nil, lastDay)

	if !strings.Contains(rendered, "kaizen new") {
		t.Errorf("empty state = %q, want it to point at kaizen new", rendered)
	}
}

func TestReportAlignsHabitsOfDifferentNameLengths(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Report([]stats.Summary{
		summary("read", "Read daily", 7, 9, model.DayDone, "vvv"),
		summary("meditate", "Meditate", 1, 4, model.DayPending, "vxo"),
	}, "2026-09-03", "2026-09-05")

	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("rendered %d lines, want a range line, a blank, a header and 2 rows:\n%s", len(lines), rendered)
	}

	stripColumn := strings.Index(lines[2], "days")
	for i, line := range lines[3:] {
		if got := strings.IndexAny(line, GlyphDone+GlyphMiss+GlyphPending); got != stripColumn {
			t.Errorf("row %d starts its strip at column %d, want %d:\n%s", i, got, stripColumn, rendered)
		}
	}
}

func TestEntriesAlignsEveryColumnIncludingTheHeader(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Entries([]stats.EntryRow{
		{Habit: model.Habit{Slug: "read"}, Entry: model.Entry{Date: "2026-09-05", Status: model.StatusDone}},
		{Habit: model.Habit{Slug: "meditate"}, Entry: model.Entry{Date: "2026-09-04", Status: model.StatusSkipped, Note: "travel"}},
	})

	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("rendered %d lines, want a header and 2 rows:\n%s", len(lines), rendered)
	}

	habitColumn := strings.Index(lines[0], "habit")
	for i, line := range lines[1:] {
		if got := strings.Index(line, "read"); i == 0 && got != habitColumn {
			t.Errorf("row %d puts the habit at column %d, want %d:\n%s", i, got, habitColumn, rendered)
		}
		if got := strings.Index(line, "meditate"); i == 1 && got != habitColumn {
			t.Errorf("row %d puts the habit at column %d, want %d:\n%s", i, got, habitColumn, rendered)
		}
	}

	for i, line := range lines {
		if strings.HasSuffix(line, " ") {
			t.Errorf("line %d has trailing whitespace: %q", i, line)
		}
	}
}

func TestEntriesWithNoRowsSaysSo(t *testing.T) {
	if rendered := New(&bytes.Buffer{}).Entries(nil); !strings.Contains(rendered, "no check-ins") {
		t.Errorf("empty state = %q, want it to say there are no check-ins", rendered)
	}
}

func TestReportShowsRateAndStreaks(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Report([]stats.Summary{
		summary("read", "Read daily", 13, 13, model.DayDone, "vvvv"),
	}, "2026-09-01", "2026-09-04")

	if !strings.Contains(rendered, "2026-09-01 .. 2026-09-04") {
		t.Errorf("report should state the range it covers:\n%s", rendered)
	}
	if !strings.Contains(rendered, "100%") {
		t.Errorf("four done days out of four should read 100%%:\n%s", rendered)
	}
	if !strings.Contains(rendered, "13") {
		t.Errorf("report should carry the streaks:\n%s", rendered)
	}
}

func TestBufferOutputCarriesNoEscapes(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Report([]stats.Summary{
		summary("read", "Read daily", 7, 9, model.DayDone, "vsxo"),
	}, "2026-09-02", "2026-09-05")

	if strings.Contains(rendered, "\x1b[") {
		t.Errorf("non-terminal output must carry no ANSI escapes, got %q", rendered)
	}
}
