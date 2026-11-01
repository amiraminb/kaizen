package cmd

import (
	"bytes"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/repository"
	"github.com/amiraminb/kaizen/internal/service"
	"github.com/spf13/cobra"
)

func TestReportSelectsHabitAndRange(t *testing.T) {
	oldSvc, oldAll := svc, reportAll
	t.Cleanup(func() { svc, reportAll = oldSvc, oldAll })
	store := repository.NewFileRepositoryAt(t.TempDir())
	svc = service.NewWithClock(store, func() time.Time {
		return time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	})
	slugs := []string{"read", "reading", "gym", "month", "retired"}
	var habits []model.Habit
	for _, slug := range slugs {
		habit := model.Habit{
			ID: slug, Slug: slug, Name: slug, StartDate: "2026-10-01",
			Schedule: model.Schedule{Kind: model.ScheduleDaily},
		}
		if slug == "retired" {
			habit.ArchivedAt = "2026-10-15T12:00:00Z"
		}
		habits = append(habits, habit)
	}
	if err := store.SaveHabits(habits); err != nil {
		t.Fatalf("SaveHabits returned error: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		all     bool
		want    []string
		header  string
		monthly bool
		details bool
		wantErr string
	}{
		{name: "default month for all habits", want: slugs[:4], header: "2026-10-01 .. 2026-10-31"},
		{name: "single range wins over a matching habit name", args: []string{"month"}, want: slugs[:4], header: "2026-10-01 .. 2026-10-31"},
		{name: "week uses sunday through saturday", args: []string{"week"}, want: slugs[:4], header: "2026-10-25 .. 2026-10-31"},
		{name: "lastweek uses sunday through saturday", args: []string{"lastweek"}, want: slugs[:4], header: "2026-10-18 .. 2026-10-24"},
		{name: "single date for all habits", args: []string{"2026-10-30"}, want: slugs[:4], header: "2026-10-30 .. 2026-10-30"},
		{name: "exact habit wins over prefixes", args: []string{"read", "month"}, want: []string{"read"}, header: "2026-10-01 .. 2026-10-31"},
		{name: "habit prefix with a single date", args: []string{"g", "2026-10-30"}, want: []string{"gym"}, header: "2026-10-30 .. 2026-10-30"},
		{name: "year for all habits", args: []string{"year"}, want: slugs[:4], header: "2026-01-01 .. 2026-10-31", monthly: true},
		{name: "habit named month with a year range", args: []string{"month", "year"}, want: []string{"month"}, header: "2026-01-01 .. 2026-10-31", monthly: true, details: true},
		{name: "habit with a range longer than one month", args: []string{"read", "2026-09-30..2026-10-31"}, want: []string{"read"}, header: "2026-09-30 .. 2026-10-31", monthly: true, details: true},
		{name: "case insensitive habit and range", args: []string{"READ", "TODAY"}, want: []string{"read"}, header: "2026-10-31 .. 2026-10-31"},
		{name: "all includes retired habits", args: []string{"month"}, all: true, want: slugs, header: "2026-10-01 .. 2026-10-31"},
		{name: "selected retired habit", args: []string{"retired", "month"}, all: true, want: []string{"retired"}, header: "2026-10-01 .. 2026-10-31"},
		{name: "retired habit excluded by default", args: []string{"retired", "month"}, wantErr: "no habit matches"},
		{name: "unknown habit", args: []string{"missing", "month"}, wantErr: "no habit matches"},
		{name: "ambiguous habit", args: []string{"re", "month"}, wantErr: "ambiguous"},
		{name: "habit selection requires a range", args: []string{"read"}, wantErr: "invalid range"},
		{name: "invalid date for selected habit", args: []string{"read", "2026-02-30"}, wantErr: "invalid range"},
		{name: "too many arguments", args: []string{"read", "month", "extra"}, wantErr: "accepts at most 2 arg(s)"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reportAll = tc.all
			var output bytes.Buffer
			command := &cobra.Command{}
			command.SetOut(&output)
			err := reportCmd.Args(command, tc.args)
			if err == nil {
				err = reportCmd.RunE(command, tc.args)
			}
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("report returned %v, want error containing %q", err, tc.wantErr)
				}
				if output.Len() != 0 {
					t.Errorf("invalid selection should not print a report: %s", output.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("report returned error: %v", err)
			}
			rendered := output.String()
			if !strings.HasPrefix(rendered, tc.header+"\n") {
				t.Errorf("report should start with range %q:\n%s", tc.header, rendered)
			}
			for _, slug := range slugs {
				if got, want := strings.Contains(rendered, "\n"+slug+"\n  completion:"), slices.Contains(tc.want, slug); got != want {
					t.Errorf("habit %q present = %v, want %v:\n%s", slug, got, want, rendered)
				}
			}
			if tc.monthly && !strings.Contains(rendered, "Oct 2026 completion:") {
				t.Errorf("selected-habit year report should use monthly rates:\n%s", rendered)
			}
			if got := strings.Contains(rendered, "\n    Oct 01–03 (Thu–Sat)"); got != tc.details {
				t.Errorf("monthly weekly details present = %v, want %v:\n%s", got, tc.details, rendered)
			}
		})
	}
}
