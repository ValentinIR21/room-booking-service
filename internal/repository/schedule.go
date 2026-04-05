package repository

import (
	"avito-talk/internal/domain"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ScheduleRepository struct {
	db *DB
}

func NewScheduleRepository(db *DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// Создание расписания
func (s *ScheduleRepository) Create(ctx context.Context, sinfo *domain.Schedule) error {

	query := `INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time) VALUES ($1, $2, $3, $4, $5)`

	_, err := s.db.Pool.Exec(
		ctx,
		query,
		sinfo.ID,
		sinfo.RoomID,
		sinfo.DaysOfWeek,
		sinfo.StartTime,
		sinfo.EndTime,
	)
	return err
}

// Получить расписание комнаты
func (s *ScheduleRepository) GetByRoomID(ctx context.Context, roomID uuid.UUID) (*domain.Schedule, error) {
	query := `SELECT id, room_id, days_of_week, start_time, end_time FROM schedules WHERE room_id = $1`

	var sched domain.Schedule

	err := s.db.Pool.QueryRow(ctx, query, roomID).Scan(&sched.ID,
		&sched.RoomID, &sched.DaysOfWeek, &sched.StartTime, &sched.EndTime)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &sched, nil
}
