package service

import (
	"fmt"
	"slices"

	"github.com/amiraminb/kaizen/internal/clock"
	"github.com/amiraminb/kaizen/internal/model"
)

type CheckInResult struct {
	Habit   model.Habit
	Entry   model.Entry
	Created bool
}

type UndoResult struct {
	Habit   model.Habit
	Date    string
	Removed bool
}

func (s *Service) CheckIn(habitInput, dateInput, status, note string) (CheckInResult, error) {
	if !model.ValidStatus(status) {
		return CheckInResult{}, fmt.Errorf("invalid status %q, want %s or %s", status, model.StatusDone, model.StatusSkipped)
	}

	habit, date, err := s.resolveTarget(habitInput, dateInput)
	if err != nil {
		return CheckInResult{}, err
	}

	entries, err := s.repo.LoadEntries()
	if err != nil {
		return CheckInResult{}, err
	}

	timestamp := clock.Timestamp(s.now())
	index := slices.IndexFunc(entries, func(e model.Entry) bool {
		return e.HabitID == habit.ID && e.Date == date
	})

	entry := model.Entry{
		HabitID:   habit.ID,
		Date:      date,
		Status:    status,
		Note:      note,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}
	created := index < 0

	if created {
		entries = append(entries, entry)
	} else {
		entry.CreatedAt = entries[index].CreatedAt
		if note == "" {
			entry.Note = entries[index].Note
		}
		entries[index] = entry
	}

	if err := s.repo.SaveEntries(entries); err != nil {
		return CheckInResult{}, err
	}
	return CheckInResult{Habit: habit, Entry: entry, Created: created}, nil
}

func (s *Service) Undo(habitInput, dateInput string) (UndoResult, error) {
	habit, date, err := s.resolveTarget(habitInput, dateInput)
	if err != nil {
		return UndoResult{}, err
	}

	entries, err := s.repo.LoadEntries()
	if err != nil {
		return UndoResult{}, err
	}

	index := slices.IndexFunc(entries, func(e model.Entry) bool {
		return e.HabitID == habit.ID && e.Date == date
	})
	if index < 0 {
		return UndoResult{Habit: habit, Date: date}, nil
	}

	if err := s.repo.SaveEntries(slices.Delete(entries, index, index+1)); err != nil {
		return UndoResult{}, err
	}
	return UndoResult{Habit: habit, Date: date, Removed: true}, nil
}

func (s *Service) resolveTarget(habitInput, dateInput string) (model.Habit, string, error) {
	asOf, err := s.AsOf()
	if err != nil {
		return model.Habit{}, "", err
	}

	date, err := clock.ParseDate(dateInput, asOf)
	if err != nil {
		return model.Habit{}, "", err
	}

	habits, err := s.repo.LoadHabits()
	if err != nil {
		return model.Habit{}, "", err
	}
	habit, err := ResolveHabit(habits, habitInput, false)
	if err != nil {
		return model.Habit{}, "", err
	}

	if date > clock.DateOf(asOf) {
		return model.Habit{}, "", fmt.Errorf("cannot check in for a future date (%s)", date)
	}
	if date < habit.StartDate {
		return model.Habit{}, "", fmt.Errorf("%q did not exist on %s, it starts %s", habit.Slug, date, habit.StartDate)
	}
	return habit, date, nil
}
