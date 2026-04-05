package service

import (
	"avito-talk/internal/domain"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func setupBookingTest() (*BookingService, *mockBookingRepo, *mockSlotRepo) {
	bookingRepo := newMockBookingRepo()
	slotRepo := newMockSlotRepo()
	svc := NewBookingService(bookingRepo, slotRepo)
	return svc, bookingRepo, slotRepo
}

func addFutureSlot(slotRepo *mockSlotRepo, roomID uuid.UUID) *domain.Slot {
	slot := &domain.Slot{
		ID:     uuid.New(),
		RoomID: roomID,
		Start:  time.Now().Add(24 * time.Hour),
		End:    time.Now().Add(24*time.Hour + 30*time.Minute),
	}
	slotRepo.Slots[slot.ID] = slot
	return slot
}

func addPastSlot(slotRepo *mockSlotRepo, roomID uuid.UUID) *domain.Slot {
	slot := &domain.Slot{
		ID:     uuid.New(),
		RoomID: roomID,
		Start:  time.Now().Add(-24 * time.Hour),
		End:    time.Now().Add(-24*time.Hour + 30*time.Minute),
	}
	slotRepo.Slots[slot.ID] = slot
	return slot
}

func TestCreateBooking_Success(t *testing.T) {
	svc, _, slotRepo := setupBookingTest()
	roomID := uuid.New()
	slot := addFutureSlot(slotRepo, roomID)
	userID := uuid.New()

	booking, err := svc.CreateBooking(context.Background(), userID, slot.ID)
	if err != nil {
		t.Fatalf("CreateBooking: %v", err)
	}
	if booking.SlotID != slot.ID {
		t.Errorf("SlotID = %v, want %v", booking.SlotID, slot.ID)
	}
	if booking.UserID != userID {
		t.Errorf("UserID = %v, want %v", booking.UserID, userID)
	}
	if booking.Status != domain.BookingStatusActive {
		t.Errorf("Status = %q, want active", booking.Status)
	}
}

func TestCreateBooking_SlotNotFound(t *testing.T) {
	svc, _, _ := setupBookingTest()

	_, err := svc.CreateBooking(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrSlotNotFound) {
		t.Errorf("expected ErrSlotNotFound, got %v", err)
	}
}

func TestCreateBooking_SlotInPast(t *testing.T) {
	svc, _, slotRepo := setupBookingTest()
	slot := addPastSlot(slotRepo, uuid.New())

	_, err := svc.CreateBooking(context.Background(), uuid.New(), slot.ID)
	if !errors.Is(err, domain.ErrSlotInPast) {
		t.Errorf("expected ErrSlotInPast, got %v", err)
	}
}

func TestCreateBooking_DoubleBooking(t *testing.T) {
	svc, _, slotRepo := setupBookingTest()
	slot := addFutureSlot(slotRepo, uuid.New())

	_, err := svc.CreateBooking(context.Background(), uuid.New(), slot.ID)
	if err != nil {
		t.Fatalf("first booking: %v", err)
	}

	_, err = svc.CreateBooking(context.Background(), uuid.New(), slot.ID)
	if !errors.Is(err, domain.ErrSlotAlreadyBooked) {
		t.Errorf("expected ErrSlotAlreadyBooked, got %v", err)
	}
}

func TestCancelBooking_Success(t *testing.T) {
	svc, _, slotRepo := setupBookingTest()
	slot := addFutureSlot(slotRepo, uuid.New())
	userID := uuid.New()

	booking, _ := svc.CreateBooking(context.Background(), userID, slot.ID)

	cancelled, err := svc.CancelBooking(context.Background(), booking.ID, userID)
	if err != nil {
		t.Fatalf("CancelBooking: %v", err)
	}
	if cancelled.Status != domain.BookingStatusCancelled {
		t.Errorf("Status = %q, want cancelled", cancelled.Status)
	}
}

func TestCancelBooking_Idempotent(t *testing.T) {
	svc, _, slotRepo := setupBookingTest()
	slot := addFutureSlot(slotRepo, uuid.New())
	userID := uuid.New()

	booking, _ := svc.CreateBooking(context.Background(), userID, slot.ID)
	svc.CancelBooking(context.Background(), booking.ID, userID)

	// повторная отмена не должна быть ошибкой
	cancelled, err := svc.CancelBooking(context.Background(), booking.ID, userID)
	if err != nil {
		t.Fatalf("idempotent cancel: %v", err)
	}
	if cancelled.Status != domain.BookingStatusCancelled {
		t.Errorf("Status = %q, want cancelled", cancelled.Status)
	}
}

func TestCancelBooking_NotFound(t *testing.T) {
	svc, _, _ := setupBookingTest()

	_, err := svc.CancelBooking(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Errorf("expected ErrBookingNotFound, got %v", err)
	}
}

func TestCancelBooking_NotOwner(t *testing.T) {
	svc, _, slotRepo := setupBookingTest()
	slot := addFutureSlot(slotRepo, uuid.New())
	userID := uuid.New()
	otherUser := uuid.New()

	booking, _ := svc.CreateBooking(context.Background(), userID, slot.ID)

	_, err := svc.CancelBooking(context.Background(), booking.ID, otherUser)
	if !errors.Is(err, domain.ErrNotOwner) {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

func TestCancelBooking_FreesSlot(t *testing.T) {
	svc, _, slotRepo := setupBookingTest()
	slot := addFutureSlot(slotRepo, uuid.New())
	userID := uuid.New()

	booking, _ := svc.CreateBooking(context.Background(), userID, slot.ID)
	svc.CancelBooking(context.Background(), booking.ID, userID)

	// слот должен быть снова доступен для бронирования
	_, err := svc.CreateBooking(context.Background(), uuid.New(), slot.ID)
	if err != nil {
		t.Fatalf("re-booking after cancel should succeed: %v", err)
	}
}

func TestGetUserBookings(t *testing.T) {
	svc, _, slotRepo := setupBookingTest()
	userID := uuid.New()
	roomID := uuid.New()

	slot1 := addFutureSlot(slotRepo, roomID)
	slot2 := addFutureSlot(slotRepo, roomID)

	svc.CreateBooking(context.Background(), userID, slot1.ID)
	svc.CreateBooking(context.Background(), userID, slot2.ID)

	bookings, err := svc.GetUserBookings(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserBookings: %v", err)
	}
	if len(bookings) != 2 {
		t.Errorf("expected 2 bookings, got %d", len(bookings))
	}
}

func TestGetAllBookings_Pagination(t *testing.T) {
	svc, _, slotRepo := setupBookingTest()
	roomID := uuid.New()

	for i := 0; i < 5; i++ {
		slot := addFutureSlot(slotRepo, roomID)
		svc.CreateBooking(context.Background(), uuid.New(), slot.ID)
	}

	bookings, total, err := svc.GetAllBookings(context.Background(), 1, 3)
	if err != nil {
		t.Fatalf("GetAllBookings: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(bookings) != 3 {
		t.Errorf("page size = %d, want 3", len(bookings))
	}
}
