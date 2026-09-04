package service

import (
	"time"

	"github.com/amiraminb/kaizen/internal/clock"
	"github.com/amiraminb/kaizen/internal/repository"
)

type Service struct {
	repo repository.Repository
	now  func() time.Time
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func NewWithClock(repo repository.Repository, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

func (s *Service) AsOf() (time.Time, error) {
	config, err := s.repo.LoadConfig()
	if err != nil {
		return time.Time{}, err
	}
	return clock.LogicalNow(s.now(), config.DayStartHour), nil
}
