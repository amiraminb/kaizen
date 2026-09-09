package service

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/amiraminb/kaizen/internal/clock"
	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/stats"
)

type NoteResult struct {
	Habit   model.Habit
	Note    model.Note
	Created bool
}

func (s *Service) AddNote(habitInput, dateInput, text string) (NoteResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return NoteResult{}, errors.New("note cannot be empty")
	}

	habit, date, err := s.resolveTarget(habitInput, dateInput)
	if err != nil {
		return NoteResult{}, err
	}

	timestamp := clock.Timestamp(s.now())
	entries, err := s.repo.LoadEntries()
	if err != nil {
		return NoteResult{}, err
	}
	entryIndex := slices.IndexFunc(entries, func(entry model.Entry) bool {
		return entry.HabitID == habit.ID && entry.Date == date
	})
	if entryIndex >= 0 {
		entry := entries[entryIndex]
		entry.Note = text
		entry.UpdatedAt = timestamp
		entries[entryIndex] = entry
		if err := s.repo.SaveEntries(entries); err != nil {
			return NoteResult{}, err
		}
		return NoteResult{Habit: habit, Note: model.Note{
			HabitID: entry.HabitID, Date: entry.Date, Text: entry.Note,
			CreatedAt: entry.CreatedAt, UpdatedAt: entry.UpdatedAt,
		}, Created: false}, nil
	}

	notes, err := s.repo.LoadNotes()
	if err != nil {
		return NoteResult{}, err
	}
	note := model.Note{
		HabitID: habit.ID, Date: date, Text: text,
		CreatedAt: timestamp, UpdatedAt: timestamp,
	}
	noteIndex := slices.IndexFunc(notes, func(existing model.Note) bool {
		return existing.HabitID == habit.ID && existing.Date == date
	})
	created := noteIndex < 0
	if created {
		notes = append(notes, note)
	} else {
		note.CreatedAt = notes[noteIndex].CreatedAt
		notes[noteIndex] = note
	}

	if err := s.repo.SaveNotes(notes); err != nil {
		return NoteResult{}, err
	}
	return NoteResult{Habit: habit, Note: note, Created: created}, nil
}

func (s *Service) NoteFor(habitInput, dateInput string) (model.Note, error) {
	habit, date, err := s.resolveTarget(habitInput, dateInput)
	if err != nil {
		return model.Note{}, err
	}

	var current model.Note
	entries, err := s.repo.LoadEntries()
	if err != nil {
		return model.Note{}, err
	}
	for _, entry := range entries {
		if entry.HabitID == habit.ID && entry.Date == date && entry.Note != "" {
			current = model.Note{
				HabitID: entry.HabitID, Date: entry.Date, Text: entry.Note,
				CreatedAt: entry.CreatedAt, UpdatedAt: entry.UpdatedAt,
			}
			break
		}
	}

	notes, err := s.repo.LoadNotes()
	if err != nil {
		return model.Note{}, err
	}
	for _, note := range notes {
		if note.HabitID == habit.ID && note.Date == date && (current.Text == "" || note.UpdatedAt >= current.UpdatedAt) {
			current = note
		}
	}
	if current.Text == "" {
		current.HabitID = habit.ID
		current.Date = date
	}
	return current, nil
}

// Notes includes notes attached to check-ins for one complete view of a habit's history.
func (s *Service) Notes(habitInput string, from, to time.Time) ([]stats.NoteRow, error) {
	habits, err := s.repo.LoadHabits()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(habitInput) != "" {
		habit, err := ResolveHabit(habits, habitInput, true)
		if err != nil {
			return nil, err
		}
		habits = []model.Habit{habit}
	}

	entries, err := s.repo.LoadEntries()
	if err != nil {
		return nil, err
	}
	notes, err := s.repo.LoadNotes()
	if err != nil {
		return nil, err
	}
	return stats.NoteRows(habits, entries, notes, from.Format(model.DateLayout), to.Format(model.DateLayout)), nil
}
