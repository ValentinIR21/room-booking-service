package service

import (
	"avito-talk/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepo BookingRepo
	slotRepo    SlotRepo
}

func NewBookingService(bookingRepo BookingRepo, slotRepo SlotRepo) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		slotRepo:    slotRepo,
	}
}

// CreateBooking создаёт бронь для пользователя на указанный слот
func (s *BookingService) CreateBooking(ctx context.Context, userID uuid.UUID, slotID uuid.UUID) (*domain.Booking, error) {
	slot, err := s.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		return nil, err
	}

	if slot.Start.Before(time.Now().UTC()) {
		return nil, domain.ErrSlotInPast
	}

	booking := &domain.Booking{
		ID:     uuid.New(),
		SlotID: slotID,
		UserID: userID,
		Status: domain.BookingStatusActive,
	}
	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

// CancelBooking отменяет бронь
func (s *BookingService) CancelBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) (*domain.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, domain.ErrNotOwner
	}

	if booking.Status == domain.BookingStatusCancelled {
		return booking, nil
	}

	if err := s.bookingRepo.Cancel(ctx, bookingID); err != nil {
		return nil, err
	}

	booking.Status = domain.BookingStatusCancelled

	return booking, nil
}

// GetUserBookings возвращает будущие брони пользователя
func (s *BookingService) GetUserBookings(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	return s.bookingRepo.GetByUser(ctx, userID)
}

// GetAllBookings возвращает все брони с пагинацией
func (s *BookingService) GetAllBookings(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	return s.bookingRepo.GetAllPaginated(ctx, page, pageSize)
}
