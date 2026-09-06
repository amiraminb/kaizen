package daterange

import (
	"testing"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
)

func TestResolve(t *testing.T) {
	// 2026-09-17 is a Thursday, so the ISO week starts Monday 2026-09-14.
	asOf := time.Date(2026, 9, 17, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		input     string
		wantFrom  string
		wantTo    string
		wantError bool
	}{
		{name: "empty defaults to this month so far", input: "", wantFrom: "2026-09-01", wantTo: "2026-09-17"},
		{name: "month", input: "month", wantFrom: "2026-09-01", wantTo: "2026-09-17"},
		{name: "today", input: "today", wantFrom: "2026-09-17", wantTo: "2026-09-17"},
		{name: "yesterday", input: "yesterday", wantFrom: "2026-09-16", wantTo: "2026-09-16"},
		{name: "week starts monday and stops today", input: "week", wantFrom: "2026-09-14", wantTo: "2026-09-17"},
		{name: "lastweek is a full monday to sunday", input: "lastweek", wantFrom: "2026-09-07", wantTo: "2026-09-13"},
		{name: "lastmonth is a full month", input: "lastmonth", wantFrom: "2026-08-01", wantTo: "2026-08-31"},
		{name: "year stops today", input: "year", wantFrom: "2026-01-01", wantTo: "2026-09-17"},
		{name: "lastyear is a full year", input: "lastyear", wantFrom: "2025-01-01", wantTo: "2025-12-31"},
		{name: "case insensitive", input: "TODAY", wantFrom: "2026-09-17", wantTo: "2026-09-17"},
		{name: "negative days means N days ago through today", input: "-6", wantFrom: "2026-09-11", wantTo: "2026-09-17"},
		{name: "Nd means the last N days", input: "7d", wantFrom: "2026-09-11", wantTo: "2026-09-17"},
		{name: "single day window", input: "1d", wantFrom: "2026-09-17", wantTo: "2026-09-17"},
		{name: "Nd spanning a month boundary", input: "30d", wantFrom: "2026-08-19", wantTo: "2026-09-17"},
		{name: "zero days is not a window", input: "0d", wantError: true},
		{name: "negative Nd", input: "-7d", wantError: true},
		{name: "explicit range with a future end is clamped", input: "2026-09-01..2099-12-31", wantFrom: "2026-09-01", wantTo: "2026-09-17"},
		{name: "explicit single future date", input: "2099-12-31", wantError: true},
		{name: "explicit range starting in the future", input: "2099-01-01..2099-12-31", wantError: true},
		{name: "single explicit date", input: "2026-08-30", wantFrom: "2026-08-30", wantTo: "2026-08-30"},
		{name: "explicit span", input: "2026-08-01..2026-08-15", wantFrom: "2026-08-01", wantTo: "2026-08-15"},
		{name: "span with spaces", input: "2026-08-01 .. 2026-08-15", wantFrom: "2026-08-01", wantTo: "2026-08-15"},
		{name: "reversed span", input: "2026-08-15..2026-08-01", wantError: true},
		{name: "positive number is not a range", input: "6", wantError: true},
		{name: "garbage", input: "lastquarter", wantError: true},
		{name: "bad start date", input: "nope..2026-08-15", wantError: true},
		{name: "bad end date", input: "2026-08-01..nope", wantError: true},
		{name: "too many parts", input: "2026-08-01..2026-08-15..2026-08-20", wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			from, to, err := Resolve(tc.input, asOf)
			if tc.wantError {
				if err == nil {
					t.Fatalf("Resolve(%q) = %s..%s, want error", tc.input, format(from), format(to))
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve(%q) returned error: %v", tc.input, err)
			}
			if got := format(from); got != tc.wantFrom {
				t.Errorf("from = %q, want %q", got, tc.wantFrom)
			}
			if got := format(to); got != tc.wantTo {
				t.Errorf("to = %q, want %q", got, tc.wantTo)
			}
		})
	}
}

func TestResolveNeverReturnsAFutureEndForOpenRanges(t *testing.T) {
	asOf := time.Date(2026, 9, 17, 14, 30, 0, 0, time.UTC)

	for _, input := range []string{"", "month", "week", "year", "-30", "30d", "2026-09-01..2099-12-31"} {
		_, to, err := Resolve(input, asOf)
		if err != nil {
			t.Fatalf("Resolve(%q) returned error: %v", input, err)
		}
		if to.After(DateOnly(asOf)) {
			t.Errorf("Resolve(%q) ends at %s, which is after today", input, format(to))
		}
	}
}

func TestResolveIgnoresTheTimeOfDay(t *testing.T) {
	morning := time.Date(2026, 9, 17, 0, 1, 0, 0, time.UTC)
	night := time.Date(2026, 9, 17, 23, 59, 0, 0, time.UTC)

	fromMorning, toMorning, _ := Resolve("week", morning)
	fromNight, toNight, _ := Resolve("week", night)

	if !fromMorning.Equal(fromNight) || !toMorning.Equal(toNight) {
		t.Errorf("range shifted with the clock: %s..%s vs %s..%s",
			format(fromMorning), format(toMorning), format(fromNight), format(toNight))
	}
}

func format(t time.Time) string {
	return t.Format(model.DateLayout)
}
