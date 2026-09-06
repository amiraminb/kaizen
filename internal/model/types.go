package model

import (
	"fmt"
	"time"
)

const DateLayout = "2006-01-02"

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

// Grid resolution slices this date without re-parsing, which Validate makes safe by
// rejecting an unparseable timestamp at load time.
func (h Habit) ArchivedDate() string {
	if h.ArchivedAt == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, h.ArchivedAt)
	if err != nil {
		return ""
	}
	return parsed.Format(DateLayout)
}

func (h Habit) Validate() error {
	if h.Slug == "" {
		return fmt.Errorf("habit %q has no slug", h.Name)
	}
	if !ValidScheduleKind(h.Schedule.Kind) {
		return fmt.Errorf("habit %q has unsupported schedule kind %q", h.Slug, h.Schedule.Kind)
	}
	if _, err := time.Parse(DateLayout, h.StartDate); err != nil {
		return fmt.Errorf("habit %q has invalid start date %q", h.Slug, h.StartDate)
	}
	if h.ArchivedAt != "" {
		if _, err := time.Parse(time.RFC3339, h.ArchivedAt); err != nil {
			return fmt.Errorf("habit %q has invalid archived_at %q", h.Slug, h.ArchivedAt)
		}
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

// The data directory is meant to be kept in git or a syncing folder, so a hand edit or
// a bad merge is an expected failure mode rather than a hypothetical one.
func (e Entry) Validate() error {
	if e.HabitID == "" {
		return fmt.Errorf("entry on %q has no habit id", e.Date)
	}
	if _, err := time.Parse(DateLayout, e.Date); err != nil {
		return fmt.Errorf("entry for habit %s has invalid date %q", e.HabitID, e.Date)
	}
	if !ValidStatus(e.Status) {
		return fmt.Errorf("entry for habit %s on %s has unsupported status %q", e.HabitID, e.Date, e.Status)
	}
	return nil
}
