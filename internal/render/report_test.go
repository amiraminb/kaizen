package render

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/stats"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func reportDate(date string) time.Time {
	parsed, err := time.Parse(model.DateLayout, date)
	if err != nil {
		panic(err)
	}
	return parsed
}

func reportOptions(from, to string) ReportOptions {
	return ReportOptions{From: reportDate(from), To: reportDate(to), AsOf: reportDate(lastDay)}
}

func TestReportGroupsDailyMarksByCalendarWeek(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Report([]stats.Summary{
		summary("read", "Read", 16, 16, model.DayDone, strings.Repeat("v", 16)),
	}, reportOptions("2026-08-21", lastDay))

	want := "  Aug 21–22 (Fri–Sat)  ✓✓\n  Aug 23–29 (Sun–Sat)  ✓✓✓✓✓✓✓\n  Aug 30–Sep 05 (Sun–Sat)  ✓✓✓✓✓✓✓\n"
	if !strings.Contains(rendered, want) {
		t.Errorf("each calendar week should have its own dated row, with seven days in a full week and explicit partial-week boundaries:\n%s", rendered)
	}
}

func TestLongReportShowsMonthlyRatesOnlySinceTheHabitStarted(t *testing.T) {
	r := New(&bytes.Buffer{})
	r.width = 160 // Months must stay on separate rows even when several would fit.
	rendered := r.Report([]stats.Summary{
		summary("read", "Read", 0, 15, model.DaySkipped, strings.Repeat("v", 15)+strings.Repeat("x", 10)+strings.Repeat("s", 11)),
	}, reportOptions("2026-01-01", lastDay))

	for _, want := range []string{
		"read\n  completion: 60% (15/25 days)  15 done  10 missed  11 skipped",
		"2026-08-01 .. 2026-09-05",
		"  Aug 2026 completion: 60% (15/25 days)\n  Sep 2026 completion: —\n",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("monthly report should contain %q:\n%s", want, rendered)
		}
	}
	if strings.ContainsAny(rendered, GlyphDone+GlyphMiss+GlyphSkipped) || strings.Contains(rendered, "Jul 2026") {
		t.Errorf("monthly report should omit daily marks and months before the start:\n%s", rendered)
	}
}

func TestNamedYearReportUsesMonthsEvenEarlyInJanuary(t *testing.T) {
	options := reportOptions("2026-01-01", "2026-01-05")
	options.AsOf, options.Monthly = options.To, true
	rendered := New(&bytes.Buffer{}).Report([]stats.Summary{{
		Habit: model.Habit{Slug: "read", StartDate: "2026-01-05"},
		Cells: []stats.Cell{{Date: "2026-01-05", Status: model.DayDone}},
		Done:  1,
	}}, options)
	if !strings.Contains(rendered, "Jan 2026 completion: 100% (1/1 days)") || strings.Contains(rendered, GlyphDone) {
		t.Errorf("a named year range should use monthly rates even with five days:\n%s", rendered)
	}
}

func TestSingleHabitYearReportExpandsMonthsIntoWeeks(t *testing.T) {
	options := reportOptions("2026-01-01", lastDay)
	options.ExpandMonths = true
	rendered := New(&bytes.Buffer{}).Report([]stats.Summary{
		summary("read", "Read", 16, 16, model.DayDone, strings.Repeat("v", 16)),
	}, options)

	want := `  Aug 2026 completion: 100% (11/11 days)
    Aug 21–22 (Fri–Sat)  ✓✓
    Aug 23–29 (Sun–Sat)  ✓✓✓✓✓✓✓
    Aug 30–31 (Sun–Mon)  ✓✓
  Sep 2026 completion: 100% (5/5 days)
    Sep 01–05 (Tue–Sat)  ✓✓✓✓✓
`
	if !strings.Contains(rendered, want) {
		t.Errorf("monthly sections should contain their own calendar weeks without duplicating days across month boundaries:\n%s", rendered)
	}
}

func TestMonthlyReportLabelsYearsAcrossYearBoundaries(t *testing.T) {
	rendered := New(&bytes.Buffer{}).Report([]stats.Summary{{
		Habit: model.Habit{Slug: "read", StartDate: "2025-12-31"},
		Cells: []stats.Cell{
			{Date: "2025-12-31", Status: model.DayDone},
			{Date: "2026-01-01", Status: model.DayMiss},
			{Date: "2026-01-02", Status: model.DaySkipped},
		},
		Done: 1, Missed: 1, Skipped: 1,
	}}, reportOptions("2025-12-01", "2026-01-02"))
	for _, want := range []string{"read\n  completion: 50% (1/2 days)", "Dec 2025 completion: 100% (1/1 days)", "Jan 2026 completion: 0% (0/1 days)"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("overall and monthly completion should each show their own numerator and denominator in %q:\n%s", want, rendered)
		}
	}
}

func TestReportDistinguishesUnscoredDaysFromFailure(t *testing.T) {
	tests := []struct{ name, pattern, rate string }{
		{"pending", "o", "—"},
		{"skipped", "sss", "—"},
		{"missed", "x", "0%"},
		{"completed", "v", "100%"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rendered := New(&bytes.Buffer{}).Report([]stats.Summary{
				summary("read", "Read", 0, 0, model.DayPending, tc.pattern),
			}, reportOptions("2026-09-01", lastDay))
			if !strings.Contains(rendered, "read\n  completion: "+tc.rate+" ") {
				t.Errorf("rate should be %s:\n%s", tc.rate, rendered)
			}
		})
	}
}

func TestReportExplainsHabitsOutsideTheSelectedRange(t *testing.T) {
	tests := []struct {
		name    string
		habit   model.Habit
		options ReportOptions
		want    string
	}{
		{"future", model.Habit{Slug: "read", StartDate: "2026-09-06"}, reportOptions("2026-09-01", lastDay), "starts 2026-09-06"},
		{"started after historical range", model.Habit{Slug: "read", StartDate: "2026-09-01"}, reportOptions("2026-08-01", "2026-08-31"), "started 2026-09-01; no days in this range"},
		{"archived", model.Habit{Slug: "read", StartDate: "2026-08-01", ArchivedAt: "2026-08-31T12:00:00Z"}, reportOptions("2026-09-01", lastDay), "archived 2026-08-31; no days in this range"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rendered := New(&bytes.Buffer{}).Report([]stats.Summary{{
				Habit: tc.habit,
				Cells: []stats.Cell{{Date: tc.options.To.Format(model.DateLayout), Status: model.DayNotApplicable}},
			}}, tc.options)
			if !strings.Contains(rendered, tc.want) || !strings.Contains(rendered, "read\n  completion: —") {
				t.Errorf("empty range should have an unscored rate and explanation %q:\n%s", tc.want, rendered)
			}
			if strings.Contains(rendered, GlyphNotApplicable) {
				t.Errorf("empty ranges should not render placeholder days:\n%s", rendered)
			}
		})
	}
}

func TestReportFitsNarrowAndWideTerminals(t *testing.T) {
	for _, width := range []int{24, 40, 80, 120} {
		for _, colored := range []bool{false, true} {
			for _, days := range []int{31, 365} {
				t.Run(fmt.Sprintf("width=%d/color=%v/days=%d", width, colored, days), func(t *testing.T) {
					r := New(&bytes.Buffer{})
					r.width = width
					if colored {
						styles := lipgloss.NewRenderer(&bytes.Buffer{})
						styles.SetColorProfile(termenv.ANSI256)
						r.done = styles.NewStyle().Foreground(lipgloss.Color("2"))
						r.label = styles.NewStyle().Bold(true)
						r.muted = styles.NewStyle().Foreground(lipgloss.Color("8"))
					}
					s := summary("a-very-long-habit-name-that-exceeds-a-narrow-terminal", "Read", days, days, model.DayDone, strings.Repeat("v", days))
					options := reportOptions(s.Cells[0].Date, lastDay)
					options.ExpandMonths = true
					rendered := r.Report([]stats.Summary{s}, options)
					for _, line := range strings.Split(rendered, "\n") {
						if got := lipgloss.Width(line); got > width {
							t.Errorf("line width %d exceeds terminal width %d: %q", got, width, line)
						}
						if strings.HasSuffix(line, " ") {
							t.Errorf("line has trailing padding: %q", line)
						}
					}
				})
			}
		}
	}
}
