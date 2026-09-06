package daterange

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
)

var names = []string{"today", "yesterday", "week", "lastweek", "month", "lastmonth", "year", "lastyear"}

func Names() []string {
	return slices.Clone(names)
}

// A range never extends past asOf: a report showing future days as anything but
// blank would misreport them, and the stats engine treats them as not applicable.
func Resolve(input string, asOf time.Time) (time.Time, time.Time, error) {
	value := strings.ToLower(strings.TrimSpace(input))
	today := DateOnly(asOf)

	switch value {
	case "", "month":
		start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
		return start, today, nil
	case "today":
		return today, today, nil
	case "yesterday":
		yesterday := today.AddDate(0, 0, -1)
		return yesterday, yesterday, nil
	case "week":
		return startOfWeek(today), today, nil
	case "lastweek":
		start := startOfWeek(today).AddDate(0, 0, -7)
		return start, start.AddDate(0, 0, 6), nil
	case "lastmonth":
		thisMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
		return thisMonth.AddDate(0, -1, 0), thisMonth.AddDate(0, 0, -1), nil
	case "year":
		return time.Date(today.Year(), time.January, 1, 0, 0, 0, 0, today.Location()), today, nil
	case "lastyear":
		year := today.Year() - 1
		start := time.Date(year, time.January, 1, 0, 0, 0, 0, today.Location())
		return start, time.Date(year, time.December, 31, 0, 0, 0, 0, today.Location()), nil
	}

	// "7d" is the documented relative form because a leading dash is consumed by the
	// flag parser before the command ever sees it; "-7" still works after "--".
	if days, ok := relativeDays(value); ok {
		return today.AddDate(0, 0, -days+1), today, nil
	}

	start, end, err := parseExplicitRange(value, today)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

func parseExplicitRange(value string, today time.Time) (time.Time, time.Time, error) {
	parts := strings.Split(value, "..")

	if len(parts) == 1 {
		day, err := time.ParseInLocation(model.DateLayout, strings.TrimSpace(parts[0]), today.Location())
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid range %q, want one of %s, Nd, YYYY-MM-DD or YYYY-MM-DD..YYYY-MM-DD", value, strings.Join(names, ", "))
		}
		return day, day, nil
	}
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid range %q", value)
	}

	start, err := time.ParseInLocation(model.DateLayout, strings.TrimSpace(parts[0]), today.Location())
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date %q", parts[0])
	}
	end, err := time.ParseInLocation(model.DateLayout, strings.TrimSpace(parts[1]), today.Location())
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date %q", parts[1])
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end date %q is before start date %q", parts[1], parts[0])
	}
	return start, end, nil
}

func relativeDays(value string) (int, bool) {
	if trimmed, ok := strings.CutSuffix(value, "d"); ok {
		days, err := strconv.Atoi(trimmed)
		if err == nil && days > 0 {
			return days, true
		}
		return 0, false
	}

	days, err := strconv.Atoi(value)
	if err == nil && days < 0 {
		return -days + 1, true
	}
	return 0, false
}

func DateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1))
}
