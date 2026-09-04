package model

const (
	StatusDone    = "done"
	StatusSkipped = "skipped"
)

const ScheduleDaily = "daily"

type DayStatus int

const (
	DayNotApplicable DayStatus = iota
	DayDone
	DaySkipped
	DayPending
	DayMiss
)

func ValidStatus(status string) bool {
	return status == StatusDone || status == StatusSkipped
}

func ValidScheduleKind(kind string) bool {
	return kind == ScheduleDaily
}
