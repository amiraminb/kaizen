package stats

import (
	"cmp"
	"slices"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
)

type Index map[string]map[string]model.Entry

func NewIndex(entries []model.Entry) Index {
	index := make(Index)
	for _, entry := range entries {
		byDate, ok := index[entry.HabitID]
		if !ok {
			byDate = make(map[string]model.Entry)
			index[entry.HabitID] = byDate
		}
		byDate[entry.Date] = entry
	}
	return index
}

func (i Index) Lookup(habitID, date string) (model.Entry, bool) {
	entry, ok := i[habitID][date]
	return entry, ok
}

type Cell struct {
	Date   string
	Status model.DayStatus
	Entry  model.Entry
}

type Summary struct {
	Habit         model.Habit
	Cells         []Cell
	Today         model.DayStatus
	Done          int
	Skipped       int
	Missed        int
	CurrentStreak int
	LongestStreak int
}

func (s Summary) StatusOn(date string) model.DayStatus {
	for _, cell := range s.Cells {
		if cell.Date == date {
			return cell.Status
		}
	}
	return model.DayNotApplicable
}

// Completion deliberately ignores skipped and pending days: a rest day you chose is
// not a failure, and a day still in progress is not yet a data point.
func (s Summary) Completion() float64 {
	scored := s.Done + s.Missed
	if scored == 0 {
		return 0
	}
	return float64(s.Done) / float64(scored)
}

type bounds struct {
	start    string
	archived string
	daily    bool
}

// Built once per grid, not per cell, since the archived date parses RFC3339. An
// unparseable one fails closed, so it cannot silently resurrect a retired habit.
func boundsOf(habit model.Habit) bounds {
	archived := habit.ArchivedDate()
	return bounds{
		start:    habit.StartDate,
		archived: archived,
		daily:    habit.Schedule.Kind == model.ScheduleDaily && (habit.ArchivedAt == "" || archived != ""),
	}
}

func (b bounds) covers(date string) bool {
	if date < b.start {
		return false
	}
	if b.archived != "" && date > b.archived {
		return false
	}
	return b.daily
}

func IsScheduled(habit model.Habit, date string) bool {
	return boundsOf(habit).covers(date)
}

func Resolve(habit model.Habit, index Index, date, asOf string) (model.DayStatus, model.Entry) {
	return resolve(boundsOf(habit), habit.ID, index, date, asOf)
}

func resolve(within bounds, habitID string, index Index, date, asOf string) (model.DayStatus, model.Entry) {
	if !within.covers(date) {
		return model.DayNotApplicable, model.Entry{}
	}

	entry, ok := index.Lookup(habitID, date)
	if ok {
		if entry.Status == model.StatusSkipped {
			return model.DaySkipped, entry
		}
		return model.DayDone, entry
	}

	switch {
	case date == asOf:
		return model.DayPending, model.Entry{}
	case date > asOf:
		return model.DayNotApplicable, model.Entry{}
	default:
		return model.DayMiss, model.Entry{}
	}
}

func Grid(habit model.Habit, index Index, from, to time.Time, asOf string) []Cell {
	within := boundsOf(habit)

	var cells []Cell
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		date := day.Format(model.DateLayout)
		status, entry := resolve(within, habit.ID, index, date, asOf)
		cells = append(cells, Cell{Date: date, Status: status, Entry: entry})
	}
	return cells
}

// Streaks are always measured over the habit's whole history, never over the window
// being displayed, so a 30-day report cannot report a 30-day cap as your best run.
func Summarize(habit model.Habit, index Index, from, to time.Time, asOf time.Time) Summary {
	asOfDate := asOf.Format(model.DateLayout)

	start, err := time.ParseInLocation(model.DateLayout, habit.StartDate, asOf.Location())
	if err != nil {
		start = from
	}
	full := Grid(habit, index, earliest(start, from), latest(asOf, to), asOfDate)

	summary := Summary{
		Habit:         habit,
		Today:         model.DayNotApplicable,
		CurrentStreak: currentStreak(full),
		LongestStreak: longestStreak(full),
	}

	fromDate := from.Format(model.DateLayout)
	toDate := to.Format(model.DateLayout)
	for _, cell := range full {
		if cell.Date == asOfDate {
			summary.Today = cell.Status
		}
		if cell.Date < fromDate || cell.Date > toDate {
			continue
		}

		summary.Cells = append(summary.Cells, cell)
		switch cell.Status {
		case model.DayDone:
			summary.Done++
		case model.DaySkipped:
			summary.Skipped++
		case model.DayMiss:
			summary.Missed++
		}
	}
	return summary
}

type EntryRow struct {
	Habit model.Habit
	Entry model.Entry
}

// Newest first, because a log is read to answer "what did I just do", and archived
// habits are included so retiring one never hides its history.
func EntryRows(habits []model.Habit, entries []model.Entry, from, to string) []EntryRow {
	byID := make(map[string]model.Habit, len(habits))
	for _, habit := range habits {
		byID[habit.ID] = habit
	}

	var rows []EntryRow
	for _, entry := range entries {
		if entry.Date < from || entry.Date > to {
			continue
		}
		habit, ok := byID[entry.HabitID]
		if !ok {
			continue
		}
		rows = append(rows, EntryRow{Habit: habit, Entry: entry})
	}

	slices.SortFunc(rows, func(a, b EntryRow) int {
		if byDate := cmp.Compare(b.Entry.Date, a.Entry.Date); byDate != 0 {
			return byDate
		}
		return cmp.Compare(a.Habit.Slug, b.Habit.Slug)
	})
	return rows
}

func currentStreak(cells []Cell) int {
	streak := 0
	for i := len(cells) - 1; i >= 0; i-- {
		switch cells[i].Status {
		case model.DayDone:
			streak++
		case model.DayMiss:
			return streak
		}
	}
	return streak
}

func longestStreak(cells []Cell) int {
	longest, running := 0, 0
	for _, cell := range cells {
		switch cell.Status {
		case model.DayDone:
			running++
			longest = max(longest, running)
		case model.DayMiss:
			running = 0
		}
	}
	return longest
}

func latest(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// The grid must cover the display window even when it predates the habit, otherwise
// strips come out at different lengths per habit and every column after them shifts.
func earliest(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
