package repository

import "github.com/amiraminb/kaizen/internal/model"

type Repository interface {
	LoadConfig() (model.Config, error)
	SaveConfig(model.Config) error
	LoadHabits() ([]model.Habit, error)
	SaveHabits([]model.Habit) error
	LoadEntries() ([]model.Entry, error)
	SaveEntries([]model.Entry) error
	LoadNotes() ([]model.Note, error)
	SaveNotes([]model.Note) error
}

type FileRepository struct {
	dataDir string
}

func NewFileRepository() *FileRepository {
	return &FileRepository{}
}

func NewFileRepositoryAt(dataDir string) *FileRepository {
	return &FileRepository{dataDir: dataDir}
}
