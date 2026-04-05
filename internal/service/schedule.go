package service

import (
	"avito-talk/internal/domain"
	"avito-talk/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
)

type ScheduleService struct {
	repos *repository.ScheduleRepository
}

func NewScheduleService(repos *repository.ScheduleRepository) *ScheduleService {
	return &ScheduleService{repos: repos}
}

func (s *ScheduleService) CreateSchedule(
	ctx context.Context,
	roomID uuid.UUID,
	daysOfWeek []int,
	startTime string,
	endTime string,
) (*domain.Schedule, error) {

	existing, err := s.repos.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("Расписание уже есть")
	}

	schedule := domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: daysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	}
	if err := s.repos.Create(ctx, &schedule); err != nil {
		return nil, err
	}
	return &schedule, nil
}
