package tui

import (
	"fmt"
	"strings"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/render"
	"github.com/amiraminb/kaizen/internal/stats"
	tea "github.com/charmbracelet/bubbletea"
)

type intent int

const (
	intentNone intent = iota
	intentDone
	intentSkipped
)

type checklistRow struct {
	habit    model.Habit
	streak   int
	original model.DayStatus
	desired  intent
}

func (r checklistRow) changed() bool {
	return r.desired != intentOf(r.original)
}

func intentOf(status model.DayStatus) intent {
	switch status {
	case model.DayDone:
		return intentDone
	case model.DaySkipped:
		return intentSkipped
	default:
		return intentNone
	}
}

type checklistModel struct {
	rows      []checklistRow
	cursor    int
	confirmed bool
	quit      bool
}

func newChecklistModel(summaries []stats.Summary) checklistModel {
	rows := make([]checklistRow, len(summaries))
	for i, summary := range summaries {
		rows[i] = checklistRow{
			habit:    summary.Habit,
			streak:   summary.CurrentStreak,
			original: summary.Today,
			desired:  intentOf(summary.Today),
		}
	}
	return checklistModel{rows: rows}
}

func (m checklistModel) Init() tea.Cmd { return nil }

func (m checklistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "ctrl+c", "q", "esc":
		m.quit = true
		return m, tea.Quit
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, len(m.rows)-1)
	case " ":
		m.setDesired(cycle(m.currentDesired()))
	case "d":
		m.setDesired(intentDone)
	case "s":
		m.setDesired(intentSkipped)
	case "c", "x":
		m.setDesired(intentNone)
	case "enter":
		m.confirmed = true
		return m, tea.Quit
	}
	return m, nil
}

func (m checklistModel) currentDesired() intent {
	if len(m.rows) == 0 {
		return intentNone
	}
	return m.rows[m.cursor].desired
}

func (m *checklistModel) setDesired(next intent) {
	if len(m.rows) == 0 {
		return
	}
	m.rows[m.cursor].desired = next
}

func cycle(current intent) intent {
	switch current {
	case intentNone:
		return intentDone
	case intentDone:
		return intentSkipped
	default:
		return intentNone
	}
}

func (m checklistModel) View() string {
	if len(m.rows) == 0 {
		return "no habits yet, add one with: kaizen new \"Read daily\"\n"
	}

	slugWidth := 0
	for _, row := range m.rows {
		slugWidth = max(slugWidth, len(row.habit.Slug))
	}

	var out strings.Builder
	out.WriteString("Today\n\n")

	for i, row := range m.rows {
		marker := "  "
		slug := render.Pad(row.habit.Slug, slugWidth)
		if i == m.cursor {
			marker = focusStyle.Render("> ")
			slug = focusStyle.Render(slug)
		}

		trailing := mutedStyle.Render(fmt.Sprintf("streak %d", row.streak))
		if row.changed() {
			trailing = warnStyle.Render("changed")
		}

		fmt.Fprintf(&out, "%s%s  %s  %s  %s\n", marker, checkbox(row.desired), slug, valueStyle.Render(row.habit.Name), trailing)
	}

	out.WriteString("\n" + mutedStyle.Render("(space cycle, d done, s skip, c clear, enter save, esc cancel)") + "\n")
	return out.String()
}

func checkbox(desired intent) string {
	switch desired {
	case intentDone:
		return doneStyle.Render("[" + render.GlyphDone + "]")
	case intentSkipped:
		return skipStyle.Render("[" + render.GlyphSkipped + "]")
	default:
		return emptyStyle.Render("[ ]")
	}
}

func RunChecklist() (int, error) {
	report, err := Svc.Today(1)
	if err != nil {
		return 0, err
	}

	final, err := tea.NewProgram(newChecklistModel(report.Summaries)).Run()
	if err != nil {
		return 0, err
	}

	result := final.(checklistModel)
	if !result.confirmed {
		return 0, nil
	}
	return applyChecklist(result.rows)
}

func applyChecklist(rows []checklistRow) (int, error) {
	applied := 0
	for _, row := range rows {
		if !row.changed() {
			continue
		}

		var err error
		switch row.desired {
		case intentDone:
			_, err = Svc.CheckIn(row.habit.Slug, "", model.StatusDone, "")
		case intentSkipped:
			_, err = Svc.CheckIn(row.habit.Slug, "", model.StatusSkipped, "")
		default:
			_, err = Svc.Undo(row.habit.Slug, "")
		}
		if err != nil {
			return applied, err
		}
		applied++
	}
	return applied, nil
}
