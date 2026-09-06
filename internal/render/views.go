package render

import (
	"fmt"
	"strings"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/stats"
)

func (r *Renderer) Today(summaries []stats.Summary) string {
	if len(summaries) == 0 {
		return "no habits yet, add one with: kaizen new \"Read daily\"\n"
	}

	slugWidth, nameWidth := 0, 0
	for _, summary := range summaries {
		slugWidth = max(slugWidth, len(summary.Habit.Slug))
		nameWidth = max(nameWidth, len(summary.Habit.Name))
	}

	var out strings.Builder
	counts := map[model.DayStatus]int{}
	for _, summary := range summaries {
		counts[summary.Today]++

		fmt.Fprintf(&out, "%s  %s  %s  %s\n",
			r.Glyph(summary.Today),
			Pad(r.Label(summary.Habit.Slug), slugWidth),
			Pad(summary.Habit.Name, nameWidth),
			r.Muted(fmt.Sprintf("streak %d", summary.CurrentStreak)),
		)
	}

	fmt.Fprintf(&out, "\n%s\n", r.Muted(todayFooter(len(summaries), counts)))
	return out.String()
}

func (r *Renderer) Streaks(summaries []stats.Summary) string {
	if len(summaries) == 0 {
		return "no habits yet, add one with: kaizen new \"Read daily\"\n"
	}

	slugWidth := len("habit")
	for _, summary := range summaries {
		slugWidth = max(slugWidth, len(summary.Habit.Slug))
	}

	var out strings.Builder
	fmt.Fprintf(&out, "%s  %s  %s  %s\n",
		Pad(r.Muted("habit"), slugWidth),
		r.Muted("recent"),
		r.Muted("cur"),
		r.Muted("best"),
	)

	for _, summary := range summaries {
		fmt.Fprintf(&out, "%s  %s  %3d  %4d\n",
			Pad(r.Label(summary.Habit.Slug), slugWidth),
			r.Strip(summary.Cells),
			summary.CurrentStreak,
			summary.LongestStreak,
		)
	}
	return out.String()
}

func (r *Renderer) Report(summaries []stats.Summary, from, to string) string {
	if len(summaries) == 0 {
		return "no habits yet, add one with: kaizen new \"Read daily\"\n"
	}

	slugWidth := len("habit")
	for _, summary := range summaries {
		slugWidth = max(slugWidth, len(summary.Habit.Slug))
	}
	stripWidth := max(len(summaries[0].Cells), len("days"))

	var out strings.Builder
	fmt.Fprintf(&out, "%s\n\n", r.Muted(from+" .. "+to))
	fmt.Fprintf(&out, "%s  %s  %s  %s  %s  %s\n",
		Pad(r.Muted("habit"), slugWidth),
		Pad(r.Muted("days"), stripWidth),
		r.Muted("done"),
		r.Muted("rate"),
		r.Muted("cur"),
		r.Muted("best"),
	)

	for _, summary := range summaries {
		fmt.Fprintf(&out, "%s  %s  %4d  %3.0f%%  %3d  %4d\n",
			Pad(r.Label(summary.Habit.Slug), slugWidth),
			Pad(r.Strip(summary.Cells), stripWidth),
			summary.Done,
			summary.Completion()*100,
			summary.CurrentStreak,
			summary.LongestStreak,
		)
	}
	return out.String()
}

func (r *Renderer) Entries(rows []stats.EntryRow) string {
	if len(rows) == 0 {
		return "no check-ins in that range\n"
	}

	slugWidth, statusWidth := len("habit"), len("status")
	for _, row := range rows {
		slugWidth = max(slugWidth, len(row.Habit.Slug))
		statusWidth = max(statusWidth, len(row.Entry.Status))
	}

	var out strings.Builder
	writeRow := func(date, slug, status, note string) {
		line := fmt.Sprintf("%s  %s  %s  %s", Pad(date, len(model.DateLayout)), Pad(slug, slugWidth), Pad(status, statusWidth), note)
		out.WriteString(strings.TrimRight(line, " ") + "\n")
	}

	writeRow(r.Muted("date"), r.Muted("habit"), r.Muted("status"), r.Muted("note"))
	for _, row := range rows {
		writeRow(row.Entry.Date, r.Label(row.Habit.Slug), r.statusText(row.Entry.Status), row.Entry.Note)
	}
	return out.String()
}

func (r *Renderer) statusText(status string) string {
	if status == model.StatusSkipped {
		return r.skipped.Render(status)
	}
	return r.done.Render(status)
}

func (r *Renderer) Strip(cells []stats.Cell) string {
	var out strings.Builder
	for _, cell := range cells {
		out.WriteString(r.Glyph(cell.Status))
	}
	return out.String()
}

func todayFooter(total int, counts map[model.DayStatus]int) string {
	parts := []string{fmt.Sprintf("%d habits", total)}
	for _, status := range []model.DayStatus{model.DayDone, model.DaySkipped, model.DayPending, model.DayMiss} {
		if count := counts[status]; count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", count, StatusWord(status)))
		}
	}
	return strings.Join(parts, " · ")
}
