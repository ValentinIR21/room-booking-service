package testutil

import (
	"avito-talk/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
)

// MockRoomRepo — мок репозитория комнат
type MockRoomRepo struct {
	Rooms map[uuid.UUID]*domain.Room
}

func NewMockRoomRepo() *MockRoomRepo {
	return &MockRoomRepo{Rooms: make(map[uuid.UUID]*domain.Room)}
}

func (m *MockRoomRepo) Create(_ context.Context, room *domain.Room) error {
	m.Rooms[room.ID] = room
	return nil
}

func (m *MockRoomRepo) List(_ context.Context) ([]domain.Room, error) {
	rooms := []domain.Room{}
	for _, r := range m.Rooms {
		rooms = append(rooms, *r)
	}
	return rooms, nil
}

func (m *MockRoomRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Room, error) {
	r, ok := m.Rooms[id]
	if !ok {
		return nil, domain.ErrRoomNotFound
	}
	return r, nil
}

// MockScheduleRepo — мок репозитория расписаний
type MockScheduleRepo struct {
	Schedules map[uuid.UUID]*domain.Schedule // key = roomID
}

func NewMockScheduleRepo() *MockScheduleRepo {
	return &MockScheduleRepo{Schedules: make(map[uuid.UUID]*domain.Schedule)}
}

func (m *MockScheduleRepo) Create(_ context.Context, s *domain.Schedule) error {
	m.Schedules[s.RoomID] = s
	return nil
}

func (m *MockScheduleRepo) GetByRoomID(_ context.Context, roomID uuid.UUID) (*domain.Schedule, error) {
	s, ok := m.Schedules[roomID]
	if !ok {
		return nil, nil
	}
	return s, nil
}

// MockSlotRepo — мок репозитория слотов
type MockSlotRepo struct {
	Slots    map[uuid.UUID]*domain.Slot
	Bookings map[uuid.UUID]bool // slotID → booked
}

func NewMockSlotRepo() *MockSlotRepo {
	return &MockSlotRepo{
		Slots:    make(map[uuid.UUID]*domain.Slot),
		Bookings: make(map[uuid.UUID]bool),
	}
}

func (m *MockSlotRepo) UpsertSlots(_ context.Context, slots []domain.Slot) error {
	for i := range slots {
		exists := false
		for _, s := range m.Slots {
			if s.RoomID == slots[i].RoomID && s.Start.Equal(slots[i].Start) {
				exists = true
				break
			}
		}
		if !exists {
			m.Slots[slots[i].ID] = &slots[i]
		}
	}
	return nil
}

func (m *MockSlotRepo) GetFreeSlots(_ context.Context, roomID uuid.UUID, start, end time.Time) ([]domain.Slot, error) {
	now := time.Now().UTC()
	result := []domain.Slot{}
	for _, s := range m.Slots {
		if s.RoomID == roomID && !s.Start.Before(start) && s.Start.Before(end) && s.Start.After(now) && !m.Bookings[s.ID] {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (m *MockSlotRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Slot, error) {
	s, ok := m.Slots[id]
	if !ok {
		return nil, domain.ErrSlotNotFound
	}
	return s, nil
}

// MockBookingRepo — мок репозитория бронирований
type MockBookingRepo struct {
	Bookings   map[uuid.UUID]*domain.Booking
	ActiveSlot map[uuid.UUID]bool // slotID → has active booking
}

func NewMockBookingRepo() *MockBookingRepo {
	return &MockBookingRepo{
		Bookings:   make(map[uuid.UUID]*domain.Booking),
		ActiveSlot: make(map[uuid.UUID]bool),
	}
}

func (m *MockBookingRepo) Create(_ context.Context, b *domain.Booking) error {
	if m.ActiveSlot[b.SlotID] {
		return domain.ErrSlotAlreadyBooked
	}
	m.Bookings[b.ID] = b
	m.ActiveSlot[b.SlotID] = true
	return nil
}

func (m *MockBookingRepo) Cancel(_ context.Context, id uuid.UUID) error {
	b, ok := m.Bookings[id]
	if !ok {
		return domain.ErrBookingNotFound
	}
	b.Status = domain.BookingStatusCancelled
	delete(m.ActiveSlot, b.SlotID)
	return nil
}

func (m *MockBookingRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Booking, error) {
	b, ok := m.Bookings[id]
	if !ok {
		return nil, domain.ErrBookingNotFound
	}
	return b, nil
}

func (m *MockBookingRepo) GetByUser(_ context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	result := []domain.Booking{}
	for _, b := range m.Bookings {
		if b.UserID == userID {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (m *MockBookingRepo) GetAllPaginated(_ context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	all := []domain.Booking{}
	for _, b := range m.Bookings {
		all = append(all, *b)
	}
	total := len(all)
	start := (page - 1) * pageSize
	if start >= total {
		return []domain.Booking{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}
