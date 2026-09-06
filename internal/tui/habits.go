package tui

import (
	"errors"
	"fmt"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/service"
)

func pickHabit(title string) (model.Habit, bool, error) {
	habits, err := Svc.ListHabits(false)
	if err != nil {
		return model.Habit{}, false, err
	}
	if len(habits) == 0 {
		return model.Habit{}, false, errors.New("no habits yet, add one with: kaizen new \"Read daily\"")
	}

	items := make([]selectionItem[string], len(habits))
	for i, habit := range habits {
		items[i] = selectionItem[string]{
			label: fmt.Sprintf("%s  %s", habit.Slug, mutedStyle.Render(habit.Name)),
			value: habit.ID,
		}
	}

	id, chose, err := runSelection(title, items)
	if err != nil || !chose {
		return model.Habit{}, false, err
	}

	for _, habit := range habits {
		if habit.ID == id {
			return habit, true, nil
		}
	}
	return model.Habit{}, false, fmt.Errorf("habit %q disappeared while choosing", id)
}

func RunHabitNew() (model.Habit, bool, error) {
	name, ok, err := runTextPrompt("New habit", "Name: ", "", func(value string) error {
		if value == "" {
			return errors.New("name cannot be empty")
		}
		return nil
	})
	if err != nil || !ok {
		return model.Habit{}, false, err
	}

	slug, ok, err := runTextPrompt("New habit", "Slug: ", service.Slugify(name), nil)
	if err != nil || !ok {
		return model.Habit{}, false, err
	}

	habit, err := Svc.CreateHabit(name, slug, "")
	if err != nil {
		return model.Habit{}, false, err
	}
	return habit, true, nil
}

func RunHabitEdit() (model.Habit, bool, error) {
	target, chose, err := pickHabit("Edit which habit?")
	if err != nil || !chose {
		return model.Habit{}, false, err
	}

	name, ok, err := runTextPrompt("Edit habit", "Name: ", target.Name, func(value string) error {
		if value == "" {
			return errors.New("name cannot be empty")
		}
		return nil
	})
	if err != nil || !ok {
		return model.Habit{}, false, err
	}

	slug, ok, err := runTextPrompt("Edit habit", "Slug: ", target.Slug, nil)
	if err != nil || !ok {
		return model.Habit{}, false, err
	}

	habit, err := Svc.UpdateHabit(target.Slug, name, slug)
	if err != nil {
		return model.Habit{}, false, err
	}
	return habit, true, nil
}
