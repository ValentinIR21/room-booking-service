package repository

import (
	"avito-talk/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
)

type SlotRepository struct {
	db *DB
}

func NewSlotRepository(db *DB) *SlotRepository {
	return &SlotRepository{db: db}
}

// Сохраняет слот
func (s *SlotRepository) Create(ctx context.Context, slot *domain.Slot) error {

	query := `
	INSERT INTO slots (id, room_id, start_time, end_time) 
	VALUES ($1, $2, $3, $4)
	`

	_, err := s.db.Pool.Exec(ctx, query, slot.ID, slot.RoomID, slot.Start, slot.End)
	return err
}

// Проверка существования слота, для данной комнаты и начала времени
func (s *SlotRepository) Exists(ctx context.Context, roomID uuid.UUID, start time.Time) (bool, error) {
	var exists bool
	query := `
	SELECT EXISTS(SELECT 1 FROM slots 
	WHERE room_id = $1 AND start_time = $2)
	`
	err := s.db.Pool.QueryRow(ctx, query, roomID, start).Scan(&exists)
	return exists, err
}

func (s *SlotRepository) GetFreeSlots(ctx context.Context, roomID uuid.UUID, start, end time.Time) ([]domain.Slot, error) {

	query := `
	SELECT s.id, s.room_id, s.start_time, s.end_time
    FROM slots s
	LEFT JOIN bookings b ON b.slot_id = s.id AND b.status = 'active'
	WHERE s.room_id = $1 AND s.start_time >= $2 AND s.start_time < $3 AND b.id IS NULL
	ORDER BY s.start_time
	`

	rows, err := s.db.Pool.Query(ctx, query, roomID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []domain.Slot

	for rows.Next() {
		var slot domain.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End); err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	return slots, nil
}

// Возвараает слот по ID
func (r *SlotRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Slot, error) {
	var slot domain.Slot
	query := `SELECT id, room_id, start_time, end_time FROM slots WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End)
	if err != nil {
		return nil, err
	}
	return &slot, nil
}
