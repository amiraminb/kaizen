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
	if len(cells) > 0 {
		built.Habit.StartDate = cells[0].Date
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

func TestReportShowsTotalsBeforeEachHabitsDatedTimeline(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Report([]stats.Summary{
		summary("read", "Read daily", 7, 9, model.DayDone, "vvv"),
		summary("meditate", "Meditate", 1, 4, model.DayPending, "vvvsxo"),
	}, reportOptions("2026-08-31", "2026-09-05"))

	for _, want := range []string{
		"Completion = done / (done + missed). Skipped and pending days are excluded.",
		"read\n  completion: 100% (3/3 days)  3 done  0 missed  0 skipped",
		"  2026-09-03 .. 2026-09-05\n  Sep 03–05 (Thu–Sat)  ✓✓✓",
		"meditate\n  completion: 75% (3/4 days)  3 done  1 missed  1 skipped",
		"  2026-08-31 .. 2026-09-05\n  Aug 31–Sep 05 (Mon–Sat)  ✓✓✓~✗◦",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("report should contain %q:\n%s", want, rendered)
		}
	}
	if strings.Index(rendered, "3 done") > strings.Index(rendered, "Sep 03–05 (Thu–Sat)  ✓✓✓") {
		t.Errorf("totals should appear before the timeline:\n%s", rendered)
	}
}

func TestReportShowsRateAndStreaks(t *testing.T) {
	options := reportOptions("2026-09-01", "2026-09-04")
	habit := model.Habit{ID: "read", Slug: "read", StartDate: "2026-08-24", Schedule: model.Schedule{Kind: model.ScheduleDaily}}
	var entries []model.Entry
	for date := reportDate(habit.StartDate); !date.After(options.AsOf); date = date.AddDate(0, 0, 1) {
		entries = append(entries, model.Entry{HabitID: habit.ID, Date: date.Format(model.DateLayout), Status: model.StatusDone})
	}
	s := stats.Summarize(habit, stats.NewIndex(entries), options.From, options.To, options.AsOf)
	rendered := New(&bytes.Buffer{}).Report([]stats.Summary{s}, options)

	if !strings.Contains(rendered, "2026-09-01 .. 2026-09-04") {
		t.Errorf("report should state the range it covers:\n%s", rendered)
	}
	if !strings.Contains(rendered, "100%") {
		t.Errorf("four done days out of four should read 100%%:\n%s", rendered)
	}
	for _, want := range []string{"Streaks as of 2026-09-05", "streak today: 13", "best lifetime: 13"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("report should explicitly identify streak scope with %q:\n%s", want, rendered)
		}
	}
}

func TestBufferOutputCarriesNoEscapes(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Report([]stats.Summary{
		summary("read", "Read daily", 7, 9, model.DayDone, "vsxo"),
	}, reportOptions("2026-09-02", "2026-09-05"))

	if strings.Contains(rendered, "\x1b[") {
		t.Errorf("non-terminal output must carry no ANSI escapes, got %q", rendered)
	}
}
