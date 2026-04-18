package service

import (
	"avito-talk/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ScheduleService struct {
	repo     ScheduleRepo
	roomRepo RoomRepo
}

func NewScheduleService(repo ScheduleRepo, roomRepo RoomRepo) *ScheduleService {
	return &ScheduleService{repo: repo, roomRepo: roomRepo}
}

func (s *ScheduleService) CreateSchedule(
	ctx context.Context,
	roomID uuid.UUID,
	daysOfWeek []int,
	startTime string,
	endTime string,
) (*domain.Schedule, error) {
	// проверка существования комнаты
	if _, err := s.roomRepo.GetByID(ctx, roomID); err != nil {
		return nil, err
	}

	// валидация daysOfWeek
	if len(daysOfWeek) == 0 {
		return nil, fmt.Errorf("%w: daysOfWeek is empty", domain.ErrInvalidSchedule)
	}
	seen := make(map[int]struct{}, len(daysOfWeek))
	for _, d := range daysOfWeek {
		if d < 1 || d > 7 {
			return nil, fmt.Errorf("%w: daysOfWeek values must be 1-7", domain.ErrInvalidSchedule)
		}
		if _, dup := seen[d]; dup {
			return nil, fmt.Errorf("%w: duplicate day %d", domain.ErrInvalidSchedule, d)
		}
		seen[d] = struct{}{}
	}

	// валидация формата времени
	st, err := time.Parse("15:04", startTime)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid startTime format", domain.ErrInvalidSchedule)
	}
	et, err := time.Parse("15:04", endTime)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid endTime format", domain.ErrInvalidSchedule)
	}
	if !et.After(st) {
		return nil, fmt.Errorf("%w: endTime must be after startTime", domain.ErrInvalidSchedule)
	}

	schedule := domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: daysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	if err = s.repo.Create(ctx, &schedule); err != nil {
		return nil, err
	}
	return &schedule, nil
}
