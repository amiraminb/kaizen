package service

import (
	"errors"
	"slices"
	"strings"

	"github.com/amiraminb/kaizen/internal/clock"
	"github.com/amiraminb/kaizen/internal/model"
)

type NoteResult struct {
	Habit   model.Habit
	Note    model.Note
	Created bool
}

func (s *Service) SaveNote(habitInput, text string) (NoteResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return NoteResult{}, errors.New("note cannot be empty")
	}

	habit, err := s.resolveNoteHabit(habitInput)
	if err != nil {
		return NoteResult{}, err
	}

	timestamp := clock.Timestamp(s.now())
	notes, err := s.repo.LoadNotes()
	if err != nil {
		return NoteResult{}, err
	}
	note := model.Note{
		HabitID: habit.ID, Text: text,
		CreatedAt: timestamp, UpdatedAt: timestamp,
	}
	noteIndex := slices.IndexFunc(notes, func(existing model.Note) bool { return existing.HabitID == habit.ID })
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

func (s *Service) NoteFor(habitInput string) (model.Note, error) {
	habit, err := s.resolveNoteHabit(habitInput)
	if err != nil {
		return model.Note{}, err
	}

	notes, err := s.repo.LoadNotes()
	if err != nil {
		return model.Note{}, err
	}
	for _, note := range notes {
		if note.HabitID == habit.ID {
			return note, nil
		}
	}
	return model.Note{HabitID: habit.ID}, nil
}

func (s *Service) resolveNoteHabit(habitInput string) (model.Habit, error) {
	habits, err := s.repo.LoadHabits()
	if err != nil {
		return model.Habit{}, err
	}
	return ResolveHabit(habits, habitInput, true)
}
