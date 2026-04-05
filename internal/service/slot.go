package service

import (
	"avito-talk/internal/domain"
	"avito-talk/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type SlotService struct {
	slotRepo     *repository.SlotRepository
	scheduleRepo *repository.ScheduleRepository
}

func NewSlotService(slotRepo *repository.SlotRepository, scheduleRepo *repository.ScheduleRepository) *SlotService {
	return &SlotService{
		slotRepo:     slotRepo,
		scheduleRepo: scheduleRepo,
	}
}

// Возвращает свободные слоты для комнаты на указанную дату и следующие 7 дней
func (s *SlotService) GetAvailableSlots(ctx context.Context, roomID uuid.UUID, dateStr string) ([]domain.Slot, error) {

	// парсим дату
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date format")
	}
	targetDate = targetDate.UTC()

	// получаем расписание комнаты
	schedule, err := s.scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return []domain.Slot{}, nil
	}

	startDate := targetDate
	endDate := targetDate.AddDate(0, 0, 7)

	// парсим время начала и окончания из строк
	startTime, _ := time.Parse("15:04:05", schedule.StartTime)
	endTime, _ := time.Parse("15:04:05", schedule.EndTime)

	// генерируем слоты
	newSlots := []domain.Slot{}
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {

		weekday := int(d.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		if !contains(schedule.DaysOfWeek, weekday) {
			continue
		}
		slotStart := time.Date(d.Year(), d.Month(), d.Day(), startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
		slotEnd := time.Date(d.Year(), d.Month(), d.Day(), endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

		for ts := slotStart; ts.Add(30*time.Minute).Before(slotEnd) || ts.Add(30*time.Minute).Equal(slotEnd); ts = ts.Add(30 * time.Minute) {
			newSlots = append(newSlots, domain.Slot{
				ID:     uuid.New(),
				RoomID: roomID,
				Start:  ts,
				End:    ts.Add(30 * time.Minute),
			})
		}
	}

	// сохраняем новые слоты
	for _, slot := range newSlots {
		exists, err := s.slotRepo.Exists(ctx, slot.RoomID, slot.Start)
		if err != nil {
			return nil, err
		}
		if !exists {
			if err := s.slotRepo.Create(ctx, &slot); err != nil {
				return nil, err
			}
		}
	}

	// возвращаем свободные слоты
	return s.slotRepo.GetFreeSlots(ctx, roomID, startDate, endDate)
}

func contains(s []int, e int) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
