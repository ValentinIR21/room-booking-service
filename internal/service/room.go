package service

import (
	"avito-talk/internal/domain"
	"avito-talk/internal/repository"
	"context"

	"github.com/google/uuid"
)

type RoomService struct {
	repos *repository.RoomRepository
}

func NewRoomService(repos *repository.RoomRepository) *RoomService {
	return &RoomService{repos: repos}
}

func (s *RoomService) CreateRoom(ctx context.Context, name string, description *string, capacity *int) (*domain.Room, error) {
	room := &domain.Room{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
	}
	if err := s.repos.Create(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) ListRooms(ctx context.Context) ([]domain.Room, error) {
	return s.repos.List(ctx)
}
