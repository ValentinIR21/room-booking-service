package service

import (
	"avito-talk/internal/domain"
	"context"

	"github.com/google/uuid"
)

type RoomService struct {
	repo RoomRepo
}

func NewRoomService(repo RoomRepo) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) CreateRoom(ctx context.Context, name string, description *string, capacity *int) (*domain.Room, error) {
	room := &domain.Room{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
	}
	if err := s.repo.Create(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) ListRooms(ctx context.Context) ([]domain.Room, error) {
	return s.repo.List(ctx)
}
