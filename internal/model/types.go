package model

import "fmt"

const SchemaVersion = 1

const DefaultDayStartHour = 4

type Config struct {
	SchemaVersion int `json:"schema_version"`
	DayStartHour  int `json:"day_start_hour"`
}

func DefaultConfig() Config {
	return Config{SchemaVersion: SchemaVersion, DayStartHour: DefaultDayStartHour}
}

func (c Config) Validate() error {
	if c.DayStartHour < 0 || c.DayStartHour > 23 {
		return fmt.Errorf("day_start_hour must be between 0 and 23, got %d", c.DayStartHour)
	}
	return nil
}

type Schedule struct {
	Kind string `json:"kind"`
}

type Habit struct {
	ID         string   `json:"id"`
	Slug       string   `json:"slug"`
	Name       string   `json:"name"`
	Schedule   Schedule `json:"schedule"`
	StartDate  string   `json:"start_date"`
	ArchivedAt string   `json:"archived_at,omitempty"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

func (h Habit) Archived() bool {
	return h.ArchivedAt != ""
}

func (h Habit) Validate() error {
	if h.Slug == "" {
		return fmt.Errorf("habit %q has no slug", h.Name)
	}
	if !ValidScheduleKind(h.Schedule.Kind) {
		return fmt.Errorf("habit %q has unsupported schedule kind %q", h.Slug, h.Schedule.Kind)
	}
	return nil
}

type Entry struct {
	HabitID   string `json:"habit_id"`
	Date      string `json:"date"`
	Status    string `json:"status"`
	Note      string `json:"note,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
