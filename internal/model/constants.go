package model

const (
	StatusDone    = "done"
	StatusSkipped = "skipped"
	StatusMissed  = "missed"
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
	return status == StatusDone || status == StatusSkipped || status == StatusMissed
}

func ValidScheduleKind(kind string) bool {
	return kind == ScheduleDaily
}
