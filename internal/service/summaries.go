package service

import (
	"strings"
	"time"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/amiraminb/kaizen/internal/stats"
)

type Report struct {
	AsOf      time.Time
	From      time.Time
	To        time.Time
	Summaries []stats.Summary
}

// asOf is resolved once and threaded through, so a single command can never straddle
// the day-start cutoff and report two different todays.
func (s *Service) Summaries(from, to time.Time, includeArchived bool) (Report, error) {
	asOf, err := s.AsOf()
	if err != nil {
		return Report{}, err
	}
	return s.summarizeAsOf(asOf, from, to, includeArchived)
}

func (s *Service) Today(windowDays int) (Report, error) {
	asOf, err := s.AsOf()
	if err != nil {
		return Report{}, err
	}

	today := truncateToDay(asOf)
	from := today.AddDate(0, 0, -max(windowDays-1, 0))
	return s.summarizeAsOf(asOf, from, today, false)
}

func (s *Service) summarizeAsOf(asOf, from, to time.Time, includeArchived bool) (Report, error) {
	habits, err := s.ListHabits(includeArchived)
	if err != nil {
		return Report{}, err
	}
	entries, err := s.repo.LoadEntries()
	if err != nil {
		return Report{}, err
	}

	index := stats.NewIndex(entries)
	report := Report{AsOf: asOf, From: from, To: to}
	for _, habit := range habits {
		report.Summaries = append(report.Summaries, stats.Summarize(habit, index, from, to, asOf))
	}
	return report, nil
}

func (s *Service) EntryRows(habitInput string, from, to time.Time) ([]stats.EntryRow, error) {
	habits, err := s.repo.LoadHabits()
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(habitInput) != "" {
		habit, err := ResolveHabit(habits, habitInput, true)
		if err != nil {
			return nil, err
		}
		habits = []model.Habit{habit}
	}

	entries, err := s.repo.LoadEntries()
	if err != nil {
		return nil, err
	}
	return stats.EntryRows(habits, entries, from.Format(model.DateLayout), to.Format(model.DateLayout)), nil
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
