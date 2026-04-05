package service

import (
	"context"
	"testing"
)

func TestCreateRoom(t *testing.T) {
	repo := newMockRoomRepo()
	svc := NewRoomService(repo)

	desc := "test room"
	cap := 10
	room, err := svc.CreateRoom(context.Background(), "Room A", &desc, &cap)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	if room.Name != "Room A" {
		t.Errorf("Name = %q, want Room A", room.Name)
	}
	if *room.Description != "test room" {
		t.Errorf("Description = %q, want test room", *room.Description)
	}
	if *room.Capacity != 10 {
		t.Errorf("Capacity = %d, want 10", *room.Capacity)
	}
}

func TestCreateRoom_NilOptionalFields(t *testing.T) {
	repo := newMockRoomRepo()
	svc := NewRoomService(repo)

	room, err := svc.CreateRoom(context.Background(), "Room B", nil, nil)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	if room.Description != nil {
		t.Errorf("Description should be nil")
	}
	if room.Capacity != nil {
		t.Errorf("Capacity should be nil")
	}
}

func TestListRooms_Empty(t *testing.T) {
	repo := newMockRoomRepo()
	svc := NewRoomService(repo)

	rooms, err := svc.ListRooms(context.Background())
	if err != nil {
		t.Fatalf("ListRooms: %v", err)
	}
	if len(rooms) != 0 {
		t.Errorf("expected 0 rooms, got %d", len(rooms))
	}
}

func TestListRooms_AfterCreate(t *testing.T) {
	repo := newMockRoomRepo()
	svc := NewRoomService(repo)

	svc.CreateRoom(context.Background(), "Room 1", nil, nil)
	svc.CreateRoom(context.Background(), "Room 2", nil, nil)

	rooms, err := svc.ListRooms(context.Background())
	if err != nil {
		t.Fatalf("ListRooms: %v", err)
	}
	if len(rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(rooms))
	}
}
