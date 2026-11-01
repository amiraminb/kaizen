package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/stats"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type ReportOptions struct {
	From, To, AsOf time.Time
	// Named year ranges use monthly rates even early in the year.
	Monthly bool
	// A selected habit expands each monthly section into weekly daily marks.
	ExpandMonths bool
}

func (r *Renderer) Report(summaries []stats.Summary, options ReportOptions) string {
	var out strings.Builder
	if len(summaries) == 0 {
		r.reportLine(&out, "", "no habits yet, add one with: kaizen new \"Read daily\"")
		return out.String()
	}

	r.reportLine(&out, "", r.Muted(options.From.Format(model.DateLayout)+" .. "+options.To.Format(model.DateLayout)))
	r.reportLine(&out, "", r.Muted("Completion = done / (done + missed). Skipped and pending days are excluded."))
	r.reportLine(&out, "", r.Muted("Streaks as of "+options.AsOf.Format(model.DateLayout)))
	monthly := options.Monthly || options.To.After(options.From.AddDate(0, 0, 30))

	for _, summary := range summaries {
		out.WriteByte('\n')
		r.reportLine(&out, "", r.Label(summary.Habit.Slug))
		r.reportTokens(&out, "  ", []string{
			completionText(summary.Done, summary.Missed),
			fmt.Sprintf("%d done", summary.Done),
			fmt.Sprintf("%d missed", summary.Missed),
			fmt.Sprintf("%d skipped", summary.Skipped),
		})
		r.reportTokens(&out, "  ", []string{
			fmt.Sprintf("streak today: %d", summary.CurrentStreak),
			fmt.Sprintf("best lifetime: %d", summary.LongestStreak),
		})

		var cells []stats.Cell
		for _, cell := range summary.Cells {
			if cell.Status != model.DayNotApplicable {
				cells = append(cells, cell)
			}
		}
		if len(cells) == 0 {
			r.reportLine(&out, "  ", r.Muted(emptyHabitRange(summary.Habit, options)))
			continue
		}

		r.reportLine(&out, "  ", r.Muted(cells[0].Date+" .. "+cells[len(cells)-1].Date))
		if monthly {
			for _, month := range monthlyReports(cells) {
				r.reportLine(&out, "  ", month.heading)
				if options.ExpandMonths {
					for _, week := range r.weeklyStrips(month.cells) {
						r.reportLine(&out, "    ", week)
					}
				}
			}
		} else {
			for _, week := range r.weeklyStrips(cells) {
				r.reportLine(&out, "  ", week)
			}
		}
	}
	return out.String()
}

func completionText(done, missed int) string {
	if done+missed == 0 {
		return "completion: —"
	}
	return fmt.Sprintf("completion: %.0f%% (%d/%d days)", 100*float64(done)/float64(done+missed), done, done+missed)
}

func emptyHabitRange(habit model.Habit, options ReportOptions) string {
	switch {
	case habit.StartDate > options.AsOf.Format(model.DateLayout):
		return "starts " + habit.StartDate
	case habit.StartDate > options.To.Format(model.DateLayout):
		return "started " + habit.StartDate + "; no days in this range"
	case habit.ArchivedDate() != "" && habit.ArchivedDate() < options.From.Format(model.DateLayout):
		return "archived " + habit.ArchivedDate() + "; no days in this range"
	default:
		return "no scheduled days in this range"
	}
}

// Group by calendar week, starting each partial week at its first displayed day.
func (r *Renderer) weeklyStrips(cells []stats.Cell) []string {
	var tokens []string
	start := 0
	appendWeek := func(end int) {
		first, _ := time.Parse(model.DateLayout, cells[start].Date)
		last, _ := time.Parse(model.DateLayout, cells[end-1].Date)
		label := first.Format("Jan 02 (Mon)")
		if end-start > 1 {
			endFormat := "02"
			if first.Month() != last.Month() || first.Year() != last.Year() {
				endFormat = "Jan 02"
			}
			label = fmt.Sprintf("%s–%s (%s–%s)", first.Format("Jan 02"), last.Format(endFormat), first.Format("Mon"), last.Format("Mon"))
		}
		tokens = append(tokens, r.Muted(label)+"  "+r.Strip(cells[start:end]))
	}
	for i := range cells {
		date, _ := time.Parse(model.DateLayout, cells[i].Date)
		if i > 0 && (date.Weekday() == time.Sunday || date.AddDate(0, 0, -1).Format(model.DateLayout) != cells[i-1].Date) {
			appendWeek(i)
			start = i
		}
	}
	if len(cells) > 0 {
		appendWeek(len(cells))
	}
	return tokens
}

type monthReport struct {
	heading string
	cells   []stats.Cell
}

func monthlyReports(cells []stats.Cell) []monthReport {
	var months []monthReport
	var month string
	start, done, missed := 0, 0, 0
	for i, cell := range cells {
		date, _ := time.Parse(model.DateLayout, cell.Date)
		label := date.Format("Jan 2006")
		if label != month {
			if month != "" {
				months = append(months, monthReport{heading: month + " " + completionText(done, missed), cells: cells[start:i]})
			}
			month, start, done, missed = label, i, 0, 0
		}
		switch cell.Status {
		case model.DayDone:
			done++
		case model.DayMiss:
			missed++
		}
	}
	if month != "" {
		months = append(months, monthReport{heading: month + " " + completionText(done, missed), cells: cells[start:]})
	}
	return months
}

// Keep each summary field together when wrapping.
func (r *Renderer) reportTokens(out *strings.Builder, indent string, tokens []string) {
	line := ""
	for _, token := range tokens {
		if line != "" && lipgloss.Width(indent+line+"  "+token) > r.width {
			r.reportLine(out, indent, line)
			line = ""
		}
		if line != "" {
			line += "  "
		}
		line += token
	}
	if line != "" {
		r.reportLine(out, indent, line)
	}
}

func (r *Renderer) reportLine(out *strings.Builder, indent, text string) {
	indent = indent[:min(len(indent), max(r.width-1, 0))]
	for _, line := range strings.Split(ansi.Wrap(text, max(r.width-len(indent), 1), ""), "\n") {
		fmt.Fprintln(out, indent+line)
	}
}
