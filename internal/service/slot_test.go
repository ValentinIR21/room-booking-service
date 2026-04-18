package service

import (
	"avito-talk/internal/domain"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func setupSlotTest() (*SlotService, *mockSlotRepo, *mockRoomRepo) {
	slotRepo := newMockSlotRepo()
	roomRepo := newMockRoomRepo()
	svc := NewSlotService(slotRepo, roomRepo)
	return svc, slotRepo, roomRepo
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestGetAvailableSlots_RoomNotFound(t *testing.T) {
	svc, _, _ := setupSlotTest()

	_, err := svc.GetAvailableSlots(context.Background(), uuid.New(), parseDate("2026-04-07"))
	if !errors.Is(err, domain.ErrRoomNotFound) {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestGetAvailableSlots_NoSlots(t *testing.T) {
	svc, _, roomRepo := setupSlotTest()

	roomID := uuid.New()
	roomRepo.Rooms[roomID] = &domain.Room{ID: roomID, Name: "Room"}

	slots, err := svc.GetAvailableSlots(context.Background(), roomID, parseDate("2026-04-07"))
	if err != nil {
		t.Fatalf("GetAvailableSlots: %v", err)
	}
	if len(slots) != 0 {
		t.Errorf("expected 0 slots, got %d", len(slots))
	}
}

func TestGetAvailableSlots_ReturnsOnlyRequestedDate(t *testing.T) {
	svc, slotRepo, roomRepo := setupSlotTest()

	roomID := uuid.New()
	roomRepo.Rooms[roomID] = &domain.Room{ID: roomID, Name: "Room"}

	tomorrow := time.Now().AddDate(0, 0, 1)
	tomorrowDate := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)
	afterTomorrow := tomorrowDate.AddDate(0, 0, 1)

	// слоты на завтра
	slot1 := domain.Slot{ID: uuid.New(), RoomID: roomID, Start: tomorrowDate.Add(9 * time.Hour), End: tomorrowDate.Add(9*time.Hour + 30*time.Minute)}
	slot2 := domain.Slot{ID: uuid.New(), RoomID: roomID, Start: tomorrowDate.Add(10 * time.Hour), End: tomorrowDate.Add(10*time.Hour + 30*time.Minute)}
	// слот на послезавтра — не должен попасть в результат
	slot3 := domain.Slot{ID: uuid.New(), RoomID: roomID, Start: afterTomorrow.Add(9 * time.Hour), End: afterTomorrow.Add(9*time.Hour + 30*time.Minute)}

	slotRepo.Slots[slot1.ID] = &slot1
	slotRepo.Slots[slot2.ID] = &slot2
	slotRepo.Slots[slot3.ID] = &slot3

	slots, err := svc.GetAvailableSlots(context.Background(), roomID, tomorrowDate)
	if err != nil {
		t.Fatalf("GetAvailableSlots: %v", err)
	}
	if len(slots) != 2 {
		t.Errorf("expected 2 slots for requested date, got %d", len(slots))
	}
}

func TestGetAvailableSlots_ExcludesBookedSlots(t *testing.T) {
	svc, slotRepo, roomRepo := setupSlotTest()

	roomID := uuid.New()
	roomRepo.Rooms[roomID] = &domain.Room{ID: roomID, Name: "Room"}

	tomorrow := time.Now().AddDate(0, 0, 1)
	tomorrowDate := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)

	slot1 := domain.Slot{ID: uuid.New(), RoomID: roomID, Start: tomorrowDate.Add(9 * time.Hour), End: tomorrowDate.Add(9*time.Hour + 30*time.Minute)}
	slot2 := domain.Slot{ID: uuid.New(), RoomID: roomID, Start: tomorrowDate.Add(10 * time.Hour), End: tomorrowDate.Add(10*time.Hour + 30*time.Minute)}

	slotRepo.Slots[slot1.ID] = &slot1
	slotRepo.Slots[slot2.ID] = &slot2
	slotRepo.Bookings[slot1.ID] = true // slot1 забронирован

	slots, err := svc.GetAvailableSlots(context.Background(), roomID, tomorrowDate)
	if err != nil {
		t.Fatalf("GetAvailableSlots: %v", err)
	}
	if len(slots) != 1 {
		t.Errorf("expected 1 free slot, got %d", len(slots))
	}
}
