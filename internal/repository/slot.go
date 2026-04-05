package repository

import (
	"avito-talk/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SlotRepository struct {
	db *DB
}

func NewSlotRepository(db *DB) *SlotRepository {
	return &SlotRepository{db: db}
}

// UpsertSlots вставляет слоты батчем, игнорируя дубликаты
func (s *SlotRepository) UpsertSlots(ctx context.Context, slots []domain.Slot) error {
	if len(slots) == 0 {
		return nil
	}

	query := "INSERT INTO slots (id, room_id, start_time, end_time) VALUES "
	args := make([]any, 0, len(slots)*4)
	for i, slot := range slots {
		if i > 0 {
			query += ", "
		}
		base := i * 4
		query += fmt.Sprintf("($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
		args = append(args, slot.ID, slot.RoomID, slot.Start, slot.End)
	}
	query += " ON CONFLICT (room_id, start_time) DO NOTHING"

	_, err := s.db.Pool.Exec(ctx, query, args...)
	return err
}

// GetFreeSlots возвращает свободные слоты для комнаты в заданном диапазоне
func (s *SlotRepository) GetFreeSlots(ctx context.Context, roomID uuid.UUID, start, end time.Time) ([]domain.Slot, error) {
	query := `
	SELECT s.id, s.room_id, s.start_time, s.end_time
	FROM slots s
	LEFT JOIN bookings b ON b.slot_id = s.id AND b.status = 'active'
	WHERE s.room_id = $1 AND s.start_time >= $2 AND s.start_time < $3
	  AND s.start_time > NOW() AND b.id IS NULL
	ORDER BY s.start_time
	`

	rows, err := s.db.Pool.Query(ctx, query, roomID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := make([]domain.Slot, 0)
	for rows.Next() {
		var slot domain.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End); err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return slots, nil
}

// GetByID возвращает слот по ID
func (s *SlotRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Slot, error) {
	var slot domain.Slot
	query := `SELECT id, room_id, start_time, end_time FROM slots WHERE id = $1`
	err := s.db.Pool.QueryRow(ctx, query, id).Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSlotNotFound
		}
		return nil, err
	}
	return &slot, nil
}
