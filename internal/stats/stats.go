package stats

import (
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

// Completion deliberately ignores skipped and pending days: a rest day you chose is
// not a failure, and a day still in progress is not yet a data point.
func (s Summary) Completion() float64 {
	scored := s.Done + s.Missed
	if scored == 0 {
		return 0
	}
	return float64(s.Done) / float64(scored)
}

func IsScheduled(habit model.Habit, date string) bool {
	if date < habit.StartDate {
		return false
	}
	if archived := habit.ArchivedDate(); archived != "" && date > archived {
		return false
	}
	return habit.Schedule.Kind == model.ScheduleDaily
}

func Resolve(habit model.Habit, index Index, date, asOf string) (model.DayStatus, model.Entry) {
	if !IsScheduled(habit, date) {
		return model.DayNotApplicable, model.Entry{}
	}

	entry, ok := index.Lookup(habit.ID, date)
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
	var cells []Cell
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		date := day.Format(model.DateLayout)
		status, entry := Resolve(habit, index, date, asOf)
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
