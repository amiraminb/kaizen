package repository

import (
	"cmp"
	"slices"

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
	return notes, nil
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
	slices.SortFunc(sorted, func(a, b model.Note) int {
		if byDate := cmp.Compare(a.Date, b.Date); byDate != 0 {
			return byDate
		}
		return cmp.Compare(a.HabitID, b.HabitID)
	})

	return saveDocument(path, notesDocument{SchemaVersion: model.SchemaVersion, Notes: sorted})
}
