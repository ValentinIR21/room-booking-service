package job

import (
	"avito-talk/internal/domain"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// --- mocks ---

type mockRoomRepo struct {
	rooms []domain.Room
}

func (m *mockRoomRepo) List(_ context.Context) ([]domain.Room, error) {
	return m.rooms, nil
}

type mockScheduleRepo struct {
	schedules map[uuid.UUID]*domain.Schedule
}

func (m *mockScheduleRepo) GetByRoomID(_ context.Context, roomID uuid.UUID) (*domain.Schedule, error) {
	s, ok := m.schedules[roomID]
	if !ok {
		return nil, nil
	}
	return s, nil
}

type mockSlotRepo struct {
	slots        []domain.Slot
	lastSlotDate map[uuid.UUID]*time.Time
}

func (m *mockSlotRepo) UpsertSlots(_ context.Context, slots []domain.Slot) error {
	m.slots = append(m.slots, slots...)
	return nil
}

func (m *mockSlotRepo) LastSlotDate(_ context.Context, roomID uuid.UUID) (*time.Time, error) {
	if m.lastSlotDate == nil {
		return nil, nil
	}
	return m.lastSlotDate[roomID], nil
}

// --- tests ---

func TestGenerateAll_NoRooms(t *testing.T) {
	gen := NewSlotGenerator(&mockRoomRepo{}, &mockScheduleRepo{schedules: map[uuid.UUID]*domain.Schedule{}}, &mockSlotRepo{})
	if err := gen.GenerateAll(context.Background()); err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}
}

func TestGenerateAll_RoomWithoutSchedule(t *testing.T) {
	roomID := uuid.New()
	gen := NewSlotGenerator(
		&mockRoomRepo{rooms: []domain.Room{{ID: roomID, Name: "R1"}}},
		&mockScheduleRepo{schedules: map[uuid.UUID]*domain.Schedule{}},
		&mockSlotRepo{},
	)
	if err := gen.GenerateAll(context.Background()); err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}
}

func TestGenerateAll_GeneratesSlots(t *testing.T) {
	roomID := uuid.New()
	slotRepo := &mockSlotRepo{}
	gen := NewSlotGenerator(
		&mockRoomRepo{rooms: []domain.Room{{ID: roomID, Name: "R1"}}},
		&mockScheduleRepo{schedules: map[uuid.UUID]*domain.Schedule{
			roomID: {
				ID:         uuid.New(),
				RoomID:     roomID,
				DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7}, // все дни
				StartTime:  "09:00",
				EndTime:    "10:00",
			},
		}},
		slotRepo,
	)

	if err := gen.GenerateAll(context.Background()); err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}

	if len(slotRepo.slots) == 0 {
		t.Fatal("expected slots to be generated")
	}

	// 1 час / 30 мин = 2 слота в день
	slotsPerDay := 2
	for _, s := range slotRepo.slots {
		if s.RoomID != roomID {
			t.Errorf("slot room = %v, want %v", s.RoomID, roomID)
		}
		if s.End.Sub(s.Start) != 30*time.Minute {
			t.Errorf("slot duration = %v, want 30m", s.End.Sub(s.Start))
		}
	}
	// должно быть generateDays * slotsPerDay
	expected := generateDays * slotsPerDay
	if len(slotRepo.slots) != expected {
		t.Errorf("got %d slots, want %d", len(slotRepo.slots), expected)
	}
}

func TestGenerateAll_SkipsExistingDays(t *testing.T) {
	roomID := uuid.New()
	now := time.Now().UTC()
	// Допустим, слоты уже есть до +15 дней от сегодня
	lastDate := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.UTC).AddDate(0, 0, 15)

	slotRepo := &mockSlotRepo{
		lastSlotDate: map[uuid.UUID]*time.Time{roomID: &lastDate},
	}
	gen := NewSlotGenerator(
		&mockRoomRepo{rooms: []domain.Room{{ID: roomID, Name: "R1"}}},
		&mockScheduleRepo{schedules: map[uuid.UUID]*domain.Schedule{
			roomID: {
				ID:         uuid.New(),
				RoomID:     roomID,
				DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7},
				StartTime:  "09:00",
				EndTime:    "10:00",
			},
		}},
		slotRepo,
	)

	if err := gen.GenerateAll(context.Background()); err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}

	// Должны быть сгенерированы только слоты за дни после lastDate
	for _, s := range slotRepo.slots {
		slotDay := time.Date(s.Start.Year(), s.Start.Month(), s.Start.Day(), 0, 0, 0, 0, time.UTC)
		lastDay := time.Date(lastDate.Year(), lastDate.Month(), lastDate.Day(), 0, 0, 0, 0, time.UTC)
		if !slotDay.After(lastDay) {
			t.Errorf("slot date %v should be after last slot date %v", slotDay, lastDay)
		}
	}
}

func TestGenerateAll_RespectsWeekdays(t *testing.T) {
	roomID := uuid.New()
	slotRepo := &mockSlotRepo{}
	// Только понедельник
	gen := NewSlotGenerator(
		&mockRoomRepo{rooms: []domain.Room{{ID: roomID, Name: "R1"}}},
		&mockScheduleRepo{schedules: map[uuid.UUID]*domain.Schedule{
			roomID: {
				ID:         uuid.New(),
				RoomID:     roomID,
				DaysOfWeek: []int{1}, // только Пн
				StartTime:  "10:00",
				EndTime:    "11:00",
			},
		}},
		slotRepo,
	)

	if err := gen.GenerateAll(context.Background()); err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}

	for _, s := range slotRepo.slots {
		wd := int(s.Start.Weekday())
		if wd == 0 {
			wd = 7
		}
		if wd != 1 {
			t.Errorf("slot on weekday %d, want 1 (Monday)", wd)
		}
	}
}

func TestParseTimeStr(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		hour    int
		min     int
	}{
		{"09:00", false, 9, 0},
		{"15:30", false, 15, 30},
		{"23:59:59", false, 23, 59},
		{"bad", true, 0, 0},
	}
	for _, tt := range tests {
		parsed, err := parseTimeStr(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseTimeStr(%q) want error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseTimeStr(%q): %v", tt.input, err)
			continue
		}
		if parsed.Hour() != tt.hour || parsed.Minute() != tt.min {
			t.Errorf("parseTimeStr(%q) = %d:%d, want %d:%d", tt.input, parsed.Hour(), parsed.Minute(), tt.hour, tt.min)
		}
	}
}
