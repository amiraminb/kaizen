package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/stats"
)

func summary(slug, name string, current, longest int, today model.DayStatus, pattern string) stats.Summary {
	var cells []stats.Cell
	for _, symbol := range pattern {
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
		cells = append(cells, stats.Cell{Status: status})
	}

	return stats.Summary{
		Habit:         model.Habit{Slug: slug, Name: name},
		Cells:         cells,
		Today:         today,
		CurrentStreak: current,
		LongestStreak: longest,
	}
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

func TestTodayRendersOneRowPerHabit(t *testing.T) {
	out := &bytes.Buffer{}
	rendered := New(out).Today([]stats.Summary{
		summary("read", "Read daily", 7, 9, model.DayDone, "vvv"),
		summary("gym", "Gym session", 0, 3, model.DayPending, "xxo"),
	})

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

func TestTodayWithNoHabitsSuggestsTheNextStep(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Today(nil)

	if !strings.Contains(rendered, "kaizen new") {
		t.Errorf("empty state = %q, want it to point at kaizen new", rendered)
	}
}

func TestStreaksAlignsHabitsOfDifferentNameLengths(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Streaks([]stats.Summary{
		summary("read", "Read daily", 7, 9, model.DayDone, "vvv"),
		summary("meditate", "Meditate", 1, 4, model.DayPending, "vxo"),
	})

	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("rendered %d lines, want a header and 2 rows:\n%s", len(lines), rendered)
	}

	stripColumn := strings.Index(lines[0], "recent")
	for i, line := range lines[1:] {
		if got := strings.IndexAny(line, GlyphDone+GlyphMiss+GlyphPending); got != stripColumn {
			t.Errorf("row %d starts its strip at column %d, want %d:\n%s", i, got, stripColumn, rendered)
		}
	}
}

func TestBufferOutputCarriesNoEscapes(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Streaks([]stats.Summary{
		summary("read", "Read daily", 7, 9, model.DayDone, "vsxo"),
	})

	if strings.Contains(rendered, "\x1b[") {
		t.Errorf("non-terminal output must carry no ANSI escapes, got %q", rendered)
	}
}
