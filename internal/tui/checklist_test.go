package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/repository"
	"github.com/amiraminb/kaizen/internal/service"
	"github.com/amiraminb/kaizen/internal/stats"
	tea "github.com/charmbracelet/bubbletea"
)

func press(m checklistModel, keys ...string) checklistModel {
	for _, key := range keys {
		var msg tea.KeyMsg
		if key == " " {
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
		} else if len(key) == 1 {
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		} else {
			msg = tea.KeyMsg{Type: keyTypes[key]}
		}

		next, _ := m.Update(msg)
		m = next.(checklistModel)
	}
	return m
}

var keyTypes = map[string]tea.KeyType{
	"up":    tea.KeyUp,
	"down":  tea.KeyDown,
	"enter": tea.KeyEnter,
	"esc":   tea.KeyEscape,
}

const testDate = "2026-09-05"

func newModel(todays ...model.DayStatus) checklistModel {
	summaries := make([]stats.Summary, len(todays))
	for i, today := range todays {
		entry := model.Entry{}
		switch today {
		case model.DayDone:
			entry.Status = model.StatusDone
		case model.DaySkipped:
			entry.Status = model.StatusSkipped
		case model.DayMiss:
			entry.Status = model.StatusMissed
		}
		summaries[i] = stats.Summary{
			Habit:         model.Habit{ID: string(rune('a' + i)), Slug: string(rune('a' + i)), Name: "Habit"},
			Cells:         []stats.Cell{{Date: testDate, Status: today, Entry: entry}},
			Today:         today,
			CurrentStreak: i,
		}
	}
	return newChecklistModel(summaries, testDate)
}

func TestChecklistOutcomeReportsWhetherItStarted(t *testing.T) {
	if (ChecklistOutcome{}).Started {
		t.Error("the zero outcome must report that the TUI never started, so the caller prints instead of blocking")
	}
}

func TestChecklistCarriesTheDisplayedDate(t *testing.T) {
	if got := newModel(model.DayPending).date; got != testDate {
		t.Errorf("date = %q, want the date captured when rows were built", got)
	}
}

func TestChecklistWillNotToggleADayThatDoesNotApply(t *testing.T) {
	m := newModel(model.DayNotApplicable)

	for _, key := range []string{" ", "d", "s", "m", "c"} {
		if got := press(m, key).rows[0].desired; got != intentNone {
			t.Errorf("key %q changed a not-applicable row to %v, want it left alone", key, got)
		}
	}
	if press(m, "d").rows[0].changed() {
		t.Error("a not-applicable row must never report itself as changed")
	}
}

func TestChecklistShowsWhyANotApplicableRowIsLocked(t *testing.T) {
	summaries := []stats.Summary{{
		Habit: model.Habit{ID: "hab_1", Slug: "trip", Name: "Trip prep", StartDate: "2030-01-01"},
		Cells: []stats.Cell{{Date: testDate, Status: model.DayNotApplicable}},
	}}

	view := newChecklistModel(summaries, testDate).View()
	if !strings.Contains(view, "starts 2030-01-01") {
		t.Errorf("view should say why the row is locked:\n%s", view)
	}
}

func TestChecklistSeedsFromTodaysStatus(t *testing.T) {
	m := newModel(model.DayPending, model.DayDone, model.DaySkipped, model.DayMiss)

	want := []intent{intentNone, intentDone, intentSkipped, intentMissed}
	for i, expected := range want {
		if m.rows[i].desired != expected {
			t.Errorf("row %d desired = %v, want %v", i, m.rows[i].desired, expected)
		}
		if m.rows[i].changed() {
			t.Errorf("row %d reports changed before any keypress", i)
		}
	}
}

func TestChecklistSpaceCyclesThroughFourStates(t *testing.T) {
	m := newModel(model.DayPending)

	states := []intent{intentDone, intentSkipped, intentMissed, intentNone, intentDone}
	for step, want := range states {
		m = press(m, " ")
		if got := m.rows[0].desired; got != want {
			t.Fatalf("after %d presses desired = %v, want %v", step+1, got, want)
		}
	}
}

func TestChecklistExplicitKeysSetStateDirectly(t *testing.T) {
	tests := []struct {
		key  string
		want intent
	}{
		{key: "d", want: intentDone},
		{key: "s", want: intentSkipped},
		{key: "m", want: intentMissed},
		{key: "c", want: intentNone},
		{key: "x", want: intentNone},
	}

	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			m := press(newModel(model.DayPending), tc.key)
			if got := m.rows[0].desired; got != tc.want {
				t.Errorf("key %q gave %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestChecklistCursorStaysInBounds(t *testing.T) {
	m := newModel(model.DayPending, model.DayPending)

	if got := press(m, "up", "up").cursor; got != 0 {
		t.Errorf("cursor = %d after moving up past the top, want 0", got)
	}
	if got := press(m, "down", "down", "down").cursor; got != 1 {
		t.Errorf("cursor = %d after moving down past the end, want 1", got)
	}
}

func TestChecklistTogglesOnlyTheRowUnderTheCursor(t *testing.T) {
	m := press(newModel(model.DayPending, model.DayPending), "down", "d")

	if m.rows[0].desired != intentNone {
		t.Errorf("row 0 desired = %v, want it untouched", m.rows[0].desired)
	}
	if m.rows[1].desired != intentDone {
		t.Errorf("row 1 desired = %v, want done", m.rows[1].desired)
	}
}

func TestChecklistEnterConfirmsAndEscapeCancels(t *testing.T) {
	confirmed := press(newModel(model.DayPending), "d", "enter")
	if !confirmed.confirmed || confirmed.quit {
		t.Errorf("enter should confirm, got confirmed=%v quit=%v", confirmed.confirmed, confirmed.quit)
	}

	cancelled := press(newModel(model.DayPending), "d", "esc")
	if cancelled.confirmed || !cancelled.quit {
		t.Errorf("esc should cancel, got confirmed=%v quit=%v", cancelled.confirmed, cancelled.quit)
	}
}

func TestChecklistOnlyChangedRowsAreTreatedAsWork(t *testing.T) {
	m := press(newModel(model.DayDone, model.DayPending), "d", "down", "d")

	if m.rows[0].changed() {
		t.Error("re-marking an already done habit is not a change")
	}
	if !m.rows[1].changed() {
		t.Error("marking a pending habit done is a change")
	}
}

func TestChecklistClearingADoneDayIsAChange(t *testing.T) {
	m := press(newModel(model.DayDone), "c")

	if !m.rows[0].changed() {
		t.Error("clearing a recorded day must count as a change so Undo runs")
	}
	if m.rows[0].desired != intentNone {
		t.Errorf("desired = %v, want none", m.rows[0].desired)
	}
}

func TestChecklistKeysOnAnEmptyListDoNotPanic(t *testing.T) {
	m := press(newModel(), " ", "d", "s", "m", "c", "up", "down")

	if len(m.rows) != 0 {
		t.Errorf("rows = %d, want 0", len(m.rows))
	}
	if !strings.Contains(m.View(), "kaizen new") {
		t.Error("empty checklist should point at kaizen new")
	}
}

func TestChecklistViewMarksChangedRows(t *testing.T) {
	view := press(newModel(model.DayPending), "d").View()

	if !strings.Contains(view, "changed") {
		t.Errorf("view should mark a modified row as changed:\n%s", view)
	}
}

func TestChecklistSavesAndClearsAnExplicitMiss(t *testing.T) {
	oldSvc := Svc
	t.Cleanup(func() { Svc = oldSvc })
	Svc = service.NewWithClock(repository.NewFileRepositoryAt(t.TempDir()), func() time.Time {
		return time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	})
	if _, err := Svc.CreateHabit("Read", "read", "2026-09-04"); err != nil {
		t.Fatalf("CreateHabit returned error: %v", err)
	}
	load := func(date string) checklistModel {
		t.Helper()
		report, err := Svc.Day(date)
		if err != nil {
			t.Fatalf("Day returned error: %v", err)
		}
		return newChecklistModel(report.Summaries, date)
	}
	apply := func(m checklistModel) {
		t.Helper()
		if applied, err := applyChecklist(m.rows, m.date); err != nil || applied != 1 {
			t.Fatalf("applyChecklist = (%d, %v), want one saved change", applied, err)
		}
	}

	apply(press(load(testDate), "m"))
	reloaded := load(testDate)
	if reloaded.rows[0].changed() || !strings.Contains(reloaded.View(), "[✗]") {
		t.Errorf("saved miss should show an unchanged checked miss:\n%s", reloaded.View())
	}
	apply(press(reloaded, "c"))
	if got := load(testDate).rows[0].original; got != model.DayPending {
		t.Errorf("cleared status = %v, want pending", got)
	}

	// Clearing an automatic past miss must not manufacture a recorded change.
	past := press(load("2026-09-04"), "c")
	if past.rows[0].changed() {
		t.Error("clearing an unrecorded past miss should not count as a change")
	}
}
