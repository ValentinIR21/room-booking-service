package service

import (
	"avito-talk/internal/domain"
	"avito-talk/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepos *repository.BookingRepository
	slotRepos    *repository.SlotRepository
}

func NewBookingService(bookingRepo *repository.BookingRepository, slotRepo *repository.SlotRepository) *BookingService {
	return &BookingService{
		bookingRepos: bookingRepo,
		slotRepos:    slotRepo,
	}
}

// CreateBooking создаёт бронь для пользователя на указанный слот
func (s *BookingService) CreateBooking(ctx context.Context, userID uuid.UUID, slotID uuid.UUID) (*domain.Booking, error) {

	// получаем слот из БД
	slot, err := s.slotRepos.GetByID(ctx, slotID)
	if err != nil {
		return nil, errors.New("slot not found")
	}

	// проверка, что слот не в прошлом
	if slot.Start.Before(time.Now().UTC()) {
		return nil, errors.New("slot is in the past")
	}

	// пытаемся создать бронь
	booking := &domain.Booking{
		ID:     uuid.New(),
		SlotID: slotID,
		UserID: userID,
		Status: domain.BookingStatusActive,
	}
	err = s.bookingRepos.Create(ctx, booking)
	if err != nil {
		//слот уже занят
		return nil, errors.New("slot already booked")
	}
	return booking, nil
}

// CancelBooking отменяет бронь
func (s *BookingService) CancelBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) (*domain.Booking, error) {
	booking, err := s.bookingRepos.GetByID(ctx, bookingID)
	if err != nil {
		return nil, errors.New("бронирование не найдено")
	}
	if booking.UserID != userID {
		return nil, errors.New("невозможно отменить бронирование другого пользователя")
	}
	if booking.Status == domain.BookingStatusCancelled {
		return booking, nil
	}
	// отменяем
	if err := s.bookingRepos.Cancel(ctx, bookingID); err != nil {
		return nil, err
	}
	booking.Status = domain.BookingStatusCancelled
	return booking, nil
}

// возвращает будущие брони пользователя
func (s *BookingService) GetUserBookings(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	return s.bookingRepos.GetByUser(ctx, userID)
}

// возвращает все брони с пагинацией (для admina)
func (s *BookingService) GetAllBookings(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	return s.bookingRepos.GetAllPaginated(ctx, page, pageSize)
}
