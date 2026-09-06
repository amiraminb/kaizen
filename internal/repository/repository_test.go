package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amiraminb/kaizen/internal/model"
)

func TestMissingFilesReadAsEmpty(t *testing.T) {
	repo := NewFileRepositoryAt(t.TempDir())

	habits, err := repo.LoadHabits()
	if err != nil {
		t.Fatalf("LoadHabits returned error: %v", err)
	}
	if len(habits) != 0 {
		t.Errorf("LoadHabits returned %d habits, want 0", len(habits))
	}

	entries, err := repo.LoadEntries()
	if err != nil {
		t.Fatalf("LoadEntries returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("LoadEntries returned %d entries, want 0", len(entries))
	}

	config, err := repo.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if config.DayStartHour != model.DefaultDayStartHour {
		t.Errorf("day start hour = %d, want the default %d", config.DayStartHour, model.DefaultDayStartHour)
	}
}

func TestSaveEntriesSortsByDateThenHabit(t *testing.T) {
	repo := NewFileRepositoryAt(t.TempDir())

	unsorted := []model.Entry{
		{HabitID: "hab_b", Date: "2026-09-03", Status: model.StatusDone},
		{HabitID: "hab_a", Date: "2026-09-03", Status: model.StatusDone},
		{HabitID: "hab_a", Date: "2026-09-01", Status: model.StatusSkipped},
	}
	if err := repo.SaveEntries(unsorted); err != nil {
		t.Fatalf("SaveEntries returned error: %v", err)
	}

	entries, err := repo.LoadEntries()
	if err != nil {
		t.Fatalf("LoadEntries returned error: %v", err)
	}

	want := []struct{ habitID, date string }{
		{"hab_a", "2026-09-01"},
		{"hab_a", "2026-09-03"},
		{"hab_b", "2026-09-03"},
	}
	if len(entries) != len(want) {
		t.Fatalf("loaded %d entries, want %d", len(entries), len(want))
	}
	for i, expected := range want {
		if entries[i].HabitID != expected.habitID || entries[i].Date != expected.date {
			t.Errorf("entry %d = (%s, %s), want (%s, %s)", i, entries[i].HabitID, entries[i].Date, expected.habitID, expected.date)
		}
	}
}

func TestSaveEntriesDoesNotReorderTheCallersSlice(t *testing.T) {
	repo := NewFileRepositoryAt(t.TempDir())

	callers := []model.Entry{
		{HabitID: "hab_b", Date: "2026-09-03"},
		{HabitID: "hab_a", Date: "2026-09-01"},
	}
	if err := repo.SaveEntries(callers); err != nil {
		t.Fatalf("SaveEntries returned error: %v", err)
	}
	if callers[0].HabitID != "hab_b" {
		t.Error("SaveEntries must not sort the slice it was given")
	}
}

func TestHabitRoundTrip(t *testing.T) {
	repo := NewFileRepositoryAt(t.TempDir())

	saved := []model.Habit{{
		ID:        "hab_1",
		Slug:      "read",
		Name:      "Read daily",
		Schedule:  model.Schedule{Kind: model.ScheduleDaily},
		StartDate: "2026-09-03",
		CreatedAt: "2026-09-03T12:00:00Z",
		UpdatedAt: "2026-09-03T12:00:00Z",
	}}
	if err := repo.SaveHabits(saved); err != nil {
		t.Fatalf("SaveHabits returned error: %v", err)
	}

	loaded, err := repo.LoadHabits()
	if err != nil {
		t.Fatalf("LoadHabits returned error: %v", err)
	}
	if len(loaded) != 1 || loaded[0] != saved[0] {
		t.Errorf("LoadHabits = %+v, want %+v", loaded, saved)
	}
}

func TestLoadHabitsRejectsUnknownScheduleKind(t *testing.T) {
	dir := t.TempDir()
	repo := NewFileRepositoryAt(dir)

	document := `{"schema_version":1,"habits":[{"id":"hab_1","slug":"read","schedule":{"kind":"times_per_week"}}]}`
	if err := os.WriteFile(filepath.Join(dir, HabitsFileName), []byte(document), 0o600); err != nil {
		t.Fatalf("writing fixture returned error: %v", err)
	}

	if _, err := repo.LoadHabits(); err == nil {
		t.Error("an unsupported schedule kind must fail loudly rather than default to daily")
	}
}

func TestLoadConfigRejectsAnImpossibleDayStartHour(t *testing.T) {
	dir := t.TempDir()
	repo := NewFileRepositoryAt(dir)

	document := `{"schema_version":1,"day_start_hour":25}`
	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(document), 0o600); err != nil {
		t.Fatalf("writing fixture returned error: %v", err)
	}

	if _, err := repo.LoadConfig(); err == nil {
		t.Error("day_start_hour 25 must be rejected")
	}
}

func TestConfigRoundTripKeepsAZeroCutoff(t *testing.T) {
	repo := NewFileRepositoryAt(t.TempDir())

	if err := repo.SaveConfig(model.Config{SchemaVersion: model.SchemaVersion, DayStartHour: 0}); err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}

	config, err := repo.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if config.DayStartHour != 0 {
		t.Errorf("day start hour = %d, want 0 to survive the round trip", config.DayStartHour)
	}
}

func TestDataDirRejectsARelativeEnvOverride(t *testing.T) {
	t.Setenv(DataDirEnv, "relative/path")

	if _, err := NewFileRepository().DataDir(); err == nil {
		t.Errorf("a relative %s must be rejected rather than silently used", DataDirEnv)
	}
}

// A blank or unusable override must fail loudly; falling back to the default would put
// the history somewhere the user would not think to look.
func TestDataDirRejectsAnUnusableEnvOverride(t *testing.T) {
	file := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("writing fixture returned error: %v", err)
	}

	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace only", value: "   "},
		{name: "tab only", value: "\t"},
		{name: "points at a regular file", value: file},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(DataDirEnv, tc.value)

			dir, err := NewFileRepository().DataDir()
			if err == nil {
				t.Errorf("%s=%q resolved to %q, want an error instead of a silent fallback", DataDirEnv, tc.value, dir)
			}
		})
	}
}

func TestLoadEntriesRejectsAnUnsupportedStatus(t *testing.T) {
	dir := t.TempDir()
	repo := NewFileRepositoryAt(dir)

	document := `{"schema_version":1,"entries":[{"habit_id":"hab_1","date":"2026-09-01","status":"maybe"}]}`
	if err := os.WriteFile(filepath.Join(dir, EntriesFileName), []byte(document), 0o600); err != nil {
		t.Fatalf("writing fixture returned error: %v", err)
	}

	if _, err := repo.LoadEntries(); err == nil {
		t.Error("an unsupported status must fail loudly rather than silently count as done")
	}
}

func TestLoadEntriesRejectsMalformedRecords(t *testing.T) {
	tests := []struct {
		name     string
		document string
	}{
		{name: "no habit id", document: `{"entries":[{"habit_id":"","date":"2026-09-01","status":"done"}]}`},
		{name: "invalid date", document: `{"entries":[{"habit_id":"hab_1","date":"01-09-2026","status":"done"}]}`},
		{name: "empty status", document: `{"entries":[{"habit_id":"hab_1","date":"2026-09-01","status":""}]}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, EntriesFileName), []byte(tc.document), 0o600); err != nil {
				t.Fatalf("writing fixture returned error: %v", err)
			}

			if _, err := NewFileRepositoryAt(dir).LoadEntries(); err == nil {
				t.Error("a malformed entry must be rejected on load")
			}
		})
	}
}

func TestDataDirUsesTheEnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(DataDirEnv, dir)

	got, err := NewFileRepository().DataDir()
	if err != nil {
		t.Fatalf("DataDir returned error: %v", err)
	}
	if got != dir {
		t.Errorf("DataDir() = %q, want %q", got, dir)
	}
}
