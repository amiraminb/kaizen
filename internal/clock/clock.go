package clock

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

// A check-in typed at 01:30 belongs to the previous day, so shifting now backwards
// by the cutoff lets every downstream date calculation stay an ordinary calendar one.
func LogicalNow(now time.Time, dayStartHour int) time.Time {
	return now.Add(-time.Duration(dayStartHour) * time.Hour)
}

func DateOf(t time.Time) string {
	return t.Format(DateLayout)
}

func Timestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

func ParseDate(input string, asOf time.Time) (string, error) {
	value := strings.ToLower(strings.TrimSpace(input))

	switch value {
	case "", "today":
		return DateOf(asOf), nil
	case "yesterday":
		return DateOf(asOf.AddDate(0, 0, -1)), nil
	}

	if strings.HasPrefix(value, "-") {
		days, err := strconv.Atoi(value)
		if err != nil {
			return "", fmt.Errorf("invalid relative date %q, want -N days", input)
		}
		return DateOf(asOf.AddDate(0, 0, days)), nil
	}

	parsed, err := time.ParseInLocation(DateLayout, value, asOf.Location())
	if err != nil {
		return "", fmt.Errorf("invalid date %q, want YYYY-MM-DD, today, yesterday or -N", input)
	}
	return DateOf(parsed), nil
}
