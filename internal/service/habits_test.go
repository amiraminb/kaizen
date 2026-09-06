package service

import (
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/repository"
)

func newTestService(t *testing.T, now time.Time) (*Service, *repository.FileRepository) {
	t.Helper()

	repo := repository.NewFileRepositoryAt(t.TempDir())
	return NewWithClock(repo, func() time.Time { return now }), repo
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Read daily", want: "read-daily"},
		{input: "  Gym  ", want: "gym"},
		{input: "Meditate 10m", want: "meditate-10m"},
		{input: "No-Alcohol!", want: "no-alcohol"},
		{input: "Drink   8   glasses", want: "drink-8-glasses"},
		{input: "!!!", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := Slugify(tc.input); got != tc.want {
				t.Errorf("Slugify(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestCreateHabit(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, _ := newTestService(t, now)

	habit, err := svc.CreateHabit("Read daily", "", "")
	if err != nil {
		t.Fatalf("CreateHabit returned error: %v", err)
	}
	if habit.Slug != "read-daily" {
		t.Errorf("slug = %q, want %q", habit.Slug, "read-daily")
	}
	if habit.StartDate != "2026-09-03" {
		t.Errorf("start date = %q, want %q", habit.StartDate, "2026-09-03")
	}
	if habit.Schedule.Kind != model.ScheduleDaily {
		t.Errorf("schedule kind = %q, want %q", habit.Schedule.Kind, model.ScheduleDaily)
	}
	if habit.Archived() {
		t.Error("a new habit must not be archived")
	}

	if _, err := svc.CreateHabit("Read daily", "", ""); err == nil {
		t.Error("creating a duplicate slug must fail")
	}
}

func TestCreateHabitWithAnExplicitStartDate(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, _ := newTestService(t, now)

	habit, err := svc.CreateHabit("Read daily", "read", "2026-08-01")
	if err != nil {
		t.Fatalf("CreateHabit returned error: %v", err)
	}
	if habit.StartDate != "2026-08-01" {
		t.Errorf("start date = %q, want %q", habit.StartDate, "2026-08-01")
	}

	result, err := svc.CheckIn("read", "2026-08-15", model.StatusDone, "")
	if err != nil {
		t.Fatalf("backfill within the start date returned error: %v", err)
	}
	if result.Entry.Date != "2026-08-15" {
		t.Errorf("entry date = %q, want %q", result.Entry.Date, "2026-08-15")
	}
}

func TestCreateHabitRejectsBadInput(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		habitName string
		slug      string
		start     string
	}{
		{name: "empty name", habitName: "  "},
		{name: "unslugifiable name", habitName: "!!!"},
		{name: "slug with spaces", habitName: "Read", slug: "read daily"},
		{name: "slug with underscore", habitName: "Read", slug: "read_daily"},
		{name: "slug with trailing dash", habitName: "Read", slug: "read-"},
		{name: "unparseable start date", habitName: "Read", start: "last august"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newTestService(t, now)
			if _, err := svc.CreateHabit(tc.habitName, tc.slug, tc.start); err == nil {
				t.Errorf("CreateHabit(%q, %q, %q) succeeded, want error", tc.habitName, tc.slug, tc.start)
			}
		})
	}
}

func TestUpdateHabit(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	t.Run("renaming leaves the slug alone", func(t *testing.T) {
		svc, _ := newTestService(t, now)
		if _, err := svc.CreateHabit("Read daily", "read", ""); err != nil {
			t.Fatalf("CreateHabit returned error: %v", err)
		}

		updated, err := svc.UpdateHabit("read", "Read every day", "")
		if err != nil {
			t.Fatalf("UpdateHabit returned error: %v", err)
		}
		if updated.Name != "Read every day" {
			t.Errorf("name = %q, want the new name", updated.Name)
		}
		if updated.Slug != "read" {
			t.Errorf("slug = %q, want it unchanged by a rename", updated.Slug)
		}
	})

	t.Run("changing the slug keeps entry history reachable", func(t *testing.T) {
		svc, repo := newTestService(t, now)
		created, err := svc.CreateHabit("Read daily", "read", "")
		if err != nil {
			t.Fatalf("CreateHabit returned error: %v", err)
		}
		if _, err := svc.CheckIn("read", "", model.StatusDone, ""); err != nil {
			t.Fatalf("CheckIn returned error: %v", err)
		}

		updated, err := svc.UpdateHabit("read", "", "reading")
		if err != nil {
			t.Fatalf("UpdateHabit returned error: %v", err)
		}
		if updated.ID != created.ID {
			t.Error("a slug change must not change the habit ID")
		}

		entries, err := repo.LoadEntries()
		if err != nil {
			t.Fatalf("LoadEntries returned error: %v", err)
		}
		if len(entries) != 1 || entries[0].HabitID != created.ID {
			t.Errorf("entries = %+v, want the original habit ID untouched", entries)
		}
		if _, err := svc.CheckIn("reading", "", model.StatusDone, ""); err != nil {
			t.Errorf("the new slug should resolve: %v", err)
		}
	})

	t.Run("rejects a slug already in use", func(t *testing.T) {
		svc, _ := newTestService(t, now)
		if _, err := svc.CreateHabit("Read daily", "read", ""); err != nil {
			t.Fatalf("CreateHabit returned error: %v", err)
		}
		if _, err := svc.CreateHabit("Gym session", "gym", ""); err != nil {
			t.Fatalf("CreateHabit returned error: %v", err)
		}

		if _, err := svc.UpdateHabit("read", "", "gym"); err == nil {
			t.Error("taking another habit's slug must fail")
		}
	})

	t.Run("keeping its own slug is allowed", func(t *testing.T) {
		svc, _ := newTestService(t, now)
		if _, err := svc.CreateHabit("Read daily", "read", ""); err != nil {
			t.Fatalf("CreateHabit returned error: %v", err)
		}

		if _, err := svc.UpdateHabit("read", "Read more", "read"); err != nil {
			t.Errorf("reusing a habit's own slug must be allowed: %v", err)
		}
	})

	t.Run("rejects an invalid slug and an unknown habit", func(t *testing.T) {
		svc, _ := newTestService(t, now)
		if _, err := svc.CreateHabit("Read daily", "read", ""); err != nil {
			t.Fatalf("CreateHabit returned error: %v", err)
		}

		if _, err := svc.UpdateHabit("read", "", "not a slug"); err == nil {
			t.Error("an invalid slug must be rejected")
		}
		if _, err := svc.UpdateHabit("nope", "x", ""); err == nil {
			t.Error("an unknown habit must be rejected")
		}
	})

	t.Run("no arguments is a no-op that does not touch UpdatedAt", func(t *testing.T) {
		svc, _ := newTestService(t, now)
		created, err := svc.CreateHabit("Read daily", "read", "")
		if err != nil {
			t.Fatalf("CreateHabit returned error: %v", err)
		}

		unchanged, err := svc.UpdateHabit("read", "", "")
		if err != nil {
			t.Fatalf("UpdateHabit returned error: %v", err)
		}
		if unchanged != created {
			t.Errorf("habit = %+v, want it byte-identical to before", unchanged)
		}
	})
}

func TestResolveHabit(t *testing.T) {
	habits := []model.Habit{
		{ID: "hab_1", Slug: "read"},
		{ID: "hab_2", Slug: "read-papers"},
		{ID: "hab_3", Slug: "meditate"},
		{ID: "hab_4", Slug: "gym", ArchivedAt: "2026-08-01T00:00:00Z"},
	}

	tests := []struct {
		name            string
		input           string
		includeArchived bool
		wantID          string
		wantErr         bool
	}{
		{name: "exact match wins over prefix", input: "read", wantID: "hab_1"},
		{name: "unique prefix", input: "med", wantID: "hab_3"},
		{name: "longer unique prefix", input: "read-p", wantID: "hab_2"},
		{name: "case insensitive", input: "MED", wantID: "hab_3"},
		{name: "ambiguous prefix", input: "re", wantErr: true},
		{name: "no match", input: "zzz", wantErr: true},
		{name: "empty input", input: "", wantErr: true},
		{name: "archived excluded by default", input: "gym", wantErr: true},
		{name: "archived included on request", input: "gym", includeArchived: true, wantID: "hab_4"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			habit, err := ResolveHabit(habits, tc.input, tc.includeArchived)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ResolveHabit(%q) = %q, want error", tc.input, habit.ID)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveHabit(%q) returned error: %v", tc.input, err)
			}
			if habit.ID != tc.wantID {
				t.Errorf("ResolveHabit(%q) = %q, want %q", tc.input, habit.ID, tc.wantID)
			}
		})
	}
}

func TestListHabitsExcludesArchived(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	svc, repo := newTestService(t, now)

	seed := []model.Habit{
		{ID: "hab_1", Slug: "read", Schedule: model.Schedule{Kind: model.ScheduleDaily}, StartDate: "2026-07-01"},
		{ID: "hab_2", Slug: "gym", Schedule: model.Schedule{Kind: model.ScheduleDaily}, StartDate: "2026-07-01", ArchivedAt: "2026-08-01T00:00:00Z"},
	}
	if err := repo.SaveHabits(seed); err != nil {
		t.Fatalf("SaveHabits returned error: %v", err)
	}

	active, err := svc.ListHabits(false)
	if err != nil {
		t.Fatalf("ListHabits returned error: %v", err)
	}
	if len(active) != 1 || active[0].Slug != "read" {
		t.Errorf("ListHabits(false) = %v, want only read", active)
	}

	all, err := svc.ListHabits(true)
	if err != nil {
		t.Fatalf("ListHabits returned error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("ListHabits(true) returned %d habits, want 2", len(all))
	}
}
