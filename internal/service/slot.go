package service

import (
	"avito-talk/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
)

type SlotService struct {
	slotRepo SlotRepo
	roomRepo RoomRepo
}

func NewSlotService(slotRepo SlotRepo, roomRepo RoomRepo) *SlotService {
	return &SlotService{
		slotRepo: slotRepo,
		roomRepo: roomRepo,
	}
}

// GetAvailableSlots возвращает свободные слоты для комнаты на указанную дату.
func (s *SlotService) GetAvailableSlots(ctx context.Context, roomID uuid.UUID, dateStr string) ([]domain.Slot, error) {
	if _, err := s.roomRepo.GetByID(ctx, roomID); err != nil {
		return nil, err
	}

	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, domain.ErrInvalidDateFormat
	}
	targetDate = targetDate.UTC()

	dayEnd := targetDate.AddDate(0, 0, 1)
	return s.slotRepo.GetFreeSlots(ctx, roomID, targetDate, dayEnd)
}
