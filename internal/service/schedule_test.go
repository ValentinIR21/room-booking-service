package service

import (
	"avito-talk/internal/domain"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func setupScheduleTest() (*ScheduleService, *mockRoomRepo, *mockScheduleRepo) {
	roomRepo := newMockRoomRepo()
	scheduleRepo := newMockScheduleRepo()
	svc := NewScheduleService(scheduleRepo, roomRepo)
	return svc, roomRepo, scheduleRepo
}

func TestCreateSchedule_Success(t *testing.T) {
	svc, roomRepo, _ := setupScheduleTest()

	roomID := uuid.New()
	roomRepo.Rooms[roomID] = &domain.Room{ID: roomID, Name: "Room A"}

	sched, err := svc.CreateSchedule(context.Background(), roomID, []int{1, 2, 3}, "09:00", "18:00")
	if err != nil {
		t.Fatalf("CreateSchedule: %v", err)
	}
	if sched.RoomID != roomID {
		t.Errorf("RoomID = %v, want %v", sched.RoomID, roomID)
	}
	if sched.StartTime != "09:00" {
		t.Errorf("StartTime = %q, want 09:00", sched.StartTime)
	}
}

func TestCreateSchedule_RoomNotFound(t *testing.T) {
	svc, _, _ := setupScheduleTest()

	_, err := svc.CreateSchedule(context.Background(), uuid.New(), []int{1}, "09:00", "18:00")
	if !errors.Is(err, domain.ErrRoomNotFound) {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestCreateSchedule_AlreadyExists(t *testing.T) {
	svc, roomRepo, _ := setupScheduleTest()

	roomID := uuid.New()
	roomRepo.Rooms[roomID] = &domain.Room{ID: roomID, Name: "Room A"}

	_, err := svc.CreateSchedule(context.Background(), roomID, []int{1}, "09:00", "18:00")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	_, err = svc.CreateSchedule(context.Background(), roomID, []int{2}, "10:00", "17:00")
	if !errors.Is(err, domain.ErrScheduleExists) {
		t.Errorf("expected ErrScheduleExists, got %v", err)
	}
}

func TestCreateSchedule_InvalidDaysOfWeek_Empty(t *testing.T) {
	svc, roomRepo, _ := setupScheduleTest()

	roomID := uuid.New()
	roomRepo.Rooms[roomID] = &domain.Room{ID: roomID, Name: "Room A"}

	_, err := svc.CreateSchedule(context.Background(), roomID, []int{}, "09:00", "18:00")
	if !errors.Is(err, domain.ErrInvalidSchedule) {
		t.Errorf("expected ErrInvalidSchedule, got %v", err)
	}
}

func TestCreateSchedule_InvalidDaysOfWeek_OutOfRange(t *testing.T) {
	svc, roomRepo, _ := setupScheduleTest()

	roomID := uuid.New()
	roomRepo.Rooms[roomID] = &domain.Room{ID: roomID, Name: "Room A"}

	_, err := svc.CreateSchedule(context.Background(), roomID, []int{0, 8}, "09:00", "18:00")
	if !errors.Is(err, domain.ErrInvalidSchedule) {
		t.Errorf("expected ErrInvalidSchedule, got %v", err)
	}
}

func TestCreateSchedule_InvalidTimeFormat(t *testing.T) {
	svc, roomRepo, _ := setupScheduleTest()

	roomID := uuid.New()
	roomRepo.Rooms[roomID] = &domain.Room{ID: roomID, Name: "Room A"}

	_, err := svc.CreateSchedule(context.Background(), roomID, []int{1}, "invalid", "18:00")
	if !errors.Is(err, domain.ErrInvalidSchedule) {
		t.Errorf("expected ErrInvalidSchedule for bad startTime, got %v", err)
	}

	_, err = svc.CreateSchedule(context.Background(), roomID, []int{1}, "09:00", "invalid")
	if !errors.Is(err, domain.ErrInvalidSchedule) {
		t.Errorf("expected ErrInvalidSchedule for bad endTime, got %v", err)
	}
}

func TestCreateSchedule_EndBeforeStart(t *testing.T) {
	svc, roomRepo, _ := setupScheduleTest()

	roomID := uuid.New()
	roomRepo.Rooms[roomID] = &domain.Room{ID: roomID, Name: "Room A"}

	_, err := svc.CreateSchedule(context.Background(), roomID, []int{1}, "18:00", "09:00")
	if !errors.Is(err, domain.ErrInvalidSchedule) {
		t.Errorf("expected ErrInvalidSchedule, got %v", err)
	}
}
