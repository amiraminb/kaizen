package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/amiraminb/kaizen/internal/clock"
	"github.com/amiraminb/kaizen/internal/model"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Entries reference this ID, so two habits sharing one would merge their histories.
// crypto/rand.Read is documented never to fail, so there is no error to handle.
func newHabitID() string {
	var raw [8]byte
	rand.Read(raw[:])
	return "hab_" + hex.EncodeToString(raw[:])
}

func Slugify(name string) string {
	var builder strings.Builder
	pendingDash := false

	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			if pendingDash && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(r)
			pendingDash = false
		default:
			pendingDash = true
		}
	}
	return builder.String()
}

func ResolveHabit(habits []model.Habit, input string, includeArchived bool) (model.Habit, error) {
	query := strings.ToLower(strings.TrimSpace(input))
	if query == "" {
		return model.Habit{}, errors.New("habit is required")
	}

	var candidates []model.Habit
	for _, habit := range habits {
		if habit.Archived() && !includeArchived {
			continue
		}
		if habit.Slug == query {
			return habit, nil
		}
		if strings.HasPrefix(habit.Slug, query) {
			candidates = append(candidates, habit)
		}
	}

	switch len(candidates) {
	case 1:
		return candidates[0], nil
	case 0:
		return model.Habit{}, fmt.Errorf("no habit matches %q", input)
	default:
		slugs := make([]string, len(candidates))
		for i, habit := range candidates {
			slugs[i] = habit.Slug
		}
		return model.Habit{}, fmt.Errorf("%q is ambiguous, it matches %s", input, strings.Join(slugs, ", "))
	}
}

func (s *Service) ListHabits(includeArchived bool) ([]model.Habit, error) {
	habits, err := s.repo.LoadHabits()
	if err != nil {
		return nil, err
	}
	if includeArchived {
		return habits, nil
	}

	var active []model.Habit
	for _, habit := range habits {
		if !habit.Archived() {
			active = append(active, habit)
		}
	}
	return active, nil
}

func normalizeSlug(name, slug string) (string, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		slug = Slugify(name)
		if slug == "" {
			return "", fmt.Errorf("cannot derive a slug from name %q, pass one explicitly", name)
		}
	}
	if !slugPattern.MatchString(slug) {
		return "", fmt.Errorf("invalid slug %q, use lowercase letters, digits and dashes", slug)
	}
	return slug, nil
}

func (s *Service) CreateHabit(name, slug, startInput string) (model.Habit, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Habit{}, errors.New("habit name cannot be empty")
	}

	slug, err := normalizeSlug(name, slug)
	if err != nil {
		return model.Habit{}, err
	}

	asOf, err := s.AsOf()
	if err != nil {
		return model.Habit{}, err
	}
	startDate, err := clock.ParseDate(startInput, asOf)
	if err != nil {
		return model.Habit{}, err
	}

	habits, err := s.repo.LoadHabits()
	if err != nil {
		return model.Habit{}, err
	}
	if slices.ContainsFunc(habits, func(h model.Habit) bool { return h.Slug == slug }) {
		return model.Habit{}, fmt.Errorf("habit %q already exists", slug)
	}

	now := s.now()
	timestamp := clock.Timestamp(now)
	habit := model.Habit{
		ID:        newHabitID(),
		Slug:      slug,
		Name:      name,
		Schedule:  model.Schedule{Kind: model.ScheduleDaily},
		StartDate: startDate,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}

	if err := s.repo.SaveHabits(append(habits, habit)); err != nil {
		return model.Habit{}, err
	}
	return habit, nil
}

// A rename never re-derives the slug: the slug is how you address the habit every
// day, so it changes only when asked for explicitly.
func (s *Service) UpdateHabit(habitInput, name, slug string) (model.Habit, error) {
	habits, err := s.repo.LoadHabits()
	if err != nil {
		return model.Habit{}, err
	}

	target, err := ResolveHabit(habits, habitInput, true)
	if err != nil {
		return model.Habit{}, err
	}

	updated := target
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		updated.Name = trimmed
	}
	if strings.TrimSpace(slug) != "" {
		normalized, err := normalizeSlug(updated.Name, slug)
		if err != nil {
			return model.Habit{}, err
		}
		taken := slices.ContainsFunc(habits, func(h model.Habit) bool {
			return h.Slug == normalized && h.ID != target.ID
		})
		if taken {
			return model.Habit{}, fmt.Errorf("habit %q already exists", normalized)
		}
		updated.Slug = normalized
	}

	if updated == target {
		return target, nil
	}

	updated.UpdatedAt = clock.Timestamp(s.now())
	index := slices.IndexFunc(habits, func(h model.Habit) bool { return h.ID == target.ID })
	habits[index] = updated

	if err := s.repo.SaveHabits(habits); err != nil {
		return model.Habit{}, err
	}
	return updated, nil
}
