package service

import (
	"avito-talk/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
)

type RoomRepo interface {
	Create(ctx context.Context, room *domain.Room) error
	List(ctx context.Context) ([]domain.Room, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error)
}

type ScheduleRepo interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	Create(ctx context.Context, schedule *domain.Schedule) error
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (*domain.Schedule, error)
}

type SlotRepo interface {
	UpsertSlots(ctx context.Context, slots []domain.Slot) error
	GetFreeSlots(ctx context.Context, roomID uuid.UUID, start, end time.Time) ([]domain.Slot, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Slot, error)
}

type BookingRepo interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	HasActiveBooking(ctx context.Context, slotID uuid.UUID) (bool, error)
	Create(ctx context.Context, booking *domain.Booking) error
	Cancel(ctx context.Context, bookingID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Booking, error)
	GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error)
	GetAllPaginated(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
}
