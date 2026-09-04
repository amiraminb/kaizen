package repository

import "github.com/amiraminb/kaizen/internal/model"

type habitsDocument struct {
	SchemaVersion int           `json:"schema_version"`
	Habits        []model.Habit `json:"habits"`
}

func (r *FileRepository) LoadHabits() ([]model.Habit, error) {
	path, err := r.DataFilePath(HabitsFileName)
	if err != nil {
		return nil, err
	}

	habits, err := loadDocument(path, func(d habitsDocument) []model.Habit { return d.Habits })
	if err != nil {
		return nil, err
	}
	for _, habit := range habits {
		if err := habit.Validate(); err != nil {
			return nil, err
		}
	}
	return habits, nil
}

func (r *FileRepository) SaveHabits(habits []model.Habit) error {
	path, err := r.DataFilePath(HabitsFileName)
	if err != nil {
		return err
	}
	if habits == nil {
		habits = []model.Habit{}
	}
	return saveDocument(path, habitsDocument{SchemaVersion: model.SchemaVersion, Habits: habits})
}
