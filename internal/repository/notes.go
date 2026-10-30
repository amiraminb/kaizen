package repository

import (
	"cmp"
	"slices"
	"strings"

	"github.com/amiraminb/kaizen/internal/model"
)

type notesDocument struct {
	SchemaVersion int          `json:"schema_version"`
	Notes         []model.Note `json:"notes"`
}

func (r *FileRepository) LoadNotes() ([]model.Note, error) {
	path, err := r.DataFilePath(NotesFileName)
	if err != nil {
		return nil, err
	}
	notes, err := loadDocument(path, func(d notesDocument) []model.Note { return d.Notes })
	if err != nil {
		return nil, err
	}
	for _, note := range notes {
		if err := note.Validate(); err != nil {
			return nil, err
		}
	}
	return oneNotePerHabit(notes), nil
}

func (r *FileRepository) SaveNotes(notes []model.Note) error {
	path, err := r.DataFilePath(NotesFileName)
	if err != nil {
		return err
	}

	sorted := slices.Clone(notes)
	if sorted == nil {
		sorted = []model.Note{}
	}
	sorted = oneNotePerHabit(sorted)
	slices.SortFunc(sorted, func(a, b model.Note) int {
		return cmp.Compare(a.HabitID, b.HabitID)
	})

	return saveDocument(path, notesDocument{SchemaVersion: model.SchemaVersion, Notes: sorted})
}

func oneNotePerHabit(notes []model.Note) []model.Note {
	byHabit := make(map[string][]model.Note, len(notes))
	for _, note := range notes {
		byHabit[note.HabitID] = append(byHabit[note.HabitID], note)
	}

	result := make([]model.Note, 0, len(byHabit))
	for _, notes := range byHabit {
		slices.SortStableFunc(notes, func(a, b model.Note) int {
			if byDate := cmp.Compare(a.Date, b.Date); byDate != 0 {
				return byDate
			}
			return cmp.Compare(a.UpdatedAt, b.UpdatedAt)
		})

		merged := notes[len(notes)-1]
		legacy := slices.ContainsFunc(notes, func(note model.Note) bool { return note.Date != "" })
		if legacy {
			parts := make([]string, 0, len(notes))
			for _, note := range notes {
				if note.Date != "" {
					parts = append(parts, note.Date+"\n"+note.Text)
				} else {
					parts = append(parts, note.Text)
				}
			}
			merged.Text = strings.Join(parts, "\n\n")
			merged.Date = ""
		}
		for _, note := range notes {
			if note.CreatedAt != "" && (merged.CreatedAt == "" || note.CreatedAt < merged.CreatedAt) {
				merged.CreatedAt = note.CreatedAt
			}
		}
		result = append(result, merged)
	}
	return result
}
