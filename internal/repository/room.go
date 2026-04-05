package repository

import (
	"avito-talk/internal/domain"
	"context"

	"github.com/google/uuid"
)

type RoomRepository struct {
	db *DB
}

func NewRoomRepository(db *DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// Создать комнату
func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	query := `INSERT INTO rooms (id, name, description, capacity) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Pool.Exec(ctx, query, room.ID, room.Name, room.Description, room.Capacity)
	return err
}

// Получить список всех комнат
func (r *RoomRepository) List(ctx context.Context) ([]domain.Room, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rooms []domain.Room
	for rows.Next() {
		var rm domain.Room
		err := rows.Scan(&rm.ID, &rm.Name, &rm.Description, &rm.Capacity, &rm.CreatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, rm)
	}
	return rooms, nil
}

// Получить комнату по ID
func (r *RoomRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	var rm domain.Room
	query := `SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&rm.ID, &rm.Name, &rm.Description, &rm.Capacity, &rm.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rm, nil
}
