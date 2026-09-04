package repository

import (
	"cmp"
	"slices"

	"github.com/amiraminb/kaizen/internal/model"
)

type entriesDocument struct {
	SchemaVersion int           `json:"schema_version"`
	Entries       []model.Entry `json:"entries"`
}

func (r *FileRepository) LoadEntries() ([]model.Entry, error) {
	path, err := r.DataFilePath(EntriesFileName)
	if err != nil {
		return nil, err
	}
	return loadDocument(path, func(d entriesDocument) []model.Entry { return d.Entries })
}

// Entries are sorted on save so the file stays readable and produces small diffs
// when the data directory is kept in git or a syncing folder.
func (r *FileRepository) SaveEntries(entries []model.Entry) error {
	path, err := r.DataFilePath(EntriesFileName)
	if err != nil {
		return err
	}

	sorted := slices.Clone(entries)
	if sorted == nil {
		sorted = []model.Entry{}
	}
	slices.SortFunc(sorted, func(a, b model.Entry) int {
		if byDate := cmp.Compare(a.Date, b.Date); byDate != 0 {
			return byDate
		}
		return cmp.Compare(a.HabitID, b.HabitID)
	})

	return saveDocument(path, entriesDocument{SchemaVersion: model.SchemaVersion, Entries: sorted})
}
