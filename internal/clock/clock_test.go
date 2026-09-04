package clock

import (
	"testing"
	"time"
)

func TestLogicalNow(t *testing.T) {
	tests := []struct {
		name         string
		now          time.Time
		dayStartHour int
		wantDate     string
	}{
		{
			name:         "after cutoff counts as the same day",
			now:          time.Date(2026, 9, 3, 9, 0, 0, 0, time.UTC),
			dayStartHour: 4,
			wantDate:     "2026-09-03",
		},
		{
			name:         "before cutoff counts as the previous day",
			now:          time.Date(2026, 9, 4, 1, 30, 0, 0, time.UTC),
			dayStartHour: 4,
			wantDate:     "2026-09-03",
		},
		{
			name:         "exactly at cutoff starts the new day",
			now:          time.Date(2026, 9, 4, 4, 0, 0, 0, time.UTC),
			dayStartHour: 4,
			wantDate:     "2026-09-04",
		},
		{
			name:         "zero cutoff is a plain calendar day",
			now:          time.Date(2026, 9, 4, 0, 15, 0, 0, time.UTC),
			dayStartHour: 0,
			wantDate:     "2026-09-04",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DateOf(LogicalNow(tc.now, tc.dayStartHour))
			if got != tc.wantDate {
				t.Errorf("DateOf(LogicalNow(%v, %d)) = %q, want %q", tc.now, tc.dayStartHour, got, tc.wantDate)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	asOf := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty means today", input: "", want: "2026-09-03"},
		{name: "today", input: "today", want: "2026-09-03"},
		{name: "today is case insensitive", input: "TODAY", want: "2026-09-03"},
		{name: "yesterday", input: "yesterday", want: "2026-09-02"},
		{name: "relative days", input: "-3", want: "2026-08-31"},
		{name: "explicit date", input: "2026-08-30", want: "2026-08-30"},
		{name: "date crossing a month", input: "-4", want: "2026-08-30"},
		{name: "garbage", input: "soonish", wantErr: true},
		{name: "wrong layout", input: "30-08-2026", wantErr: true},
		{name: "impossible date", input: "2026-02-31", wantErr: true},
		{name: "bad relative", input: "-abc", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDate(tc.input, asOf)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseDate(%q) = %q, want error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDate(%q) returned error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("ParseDate(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestTimestampKeepsOffset(t *testing.T) {
	zone := time.FixedZone("EDT", -4*60*60)
	got := Timestamp(time.Date(2026, 9, 3, 1, 14, 22, 0, zone))
	want := "2026-09-03T01:14:22-04:00"
	if got != want {
		t.Errorf("Timestamp() = %q, want %q", got, want)
	}
}
