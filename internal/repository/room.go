package repository

import (
	"avito-talk/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RoomRepository struct {
	db *DB
}

func NewRoomRepository(db *DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// Create создаёт комнату
func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	query := `INSERT INTO rooms (id, name, description, capacity) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Pool.Exec(ctx, query, room.ID, room.Name, room.Description, room.Capacity)
	return err
}

// List возвращает список всех комнат
func (r *RoomRepository) List(ctx context.Context) ([]domain.Room, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]domain.Room, 0)
	for rows.Next() {
		var rm domain.Room
		if err := rows.Scan(&rm.ID, &rm.Name, &rm.Description, &rm.Capacity, &rm.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, rm)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

// GetByID возвращает комнату по ID
func (r *RoomRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	var rm domain.Room
	query := `SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&rm.ID, &rm.Name, &rm.Description, &rm.Capacity, &rm.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, err
	}
	return &rm, nil
}
