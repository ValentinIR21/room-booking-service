package repository

import (
	"avito-talk/internal/domain"
	"context"

	"github.com/google/uuid"
)

type BookingRepository struct {
	db *DB
}

func NewBookingRepository(db *DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Создание бронирования
func (b *BookingRepository) Create(ctx context.Context, booking *domain.Booking) error {

	query := `
	INSERT INTO bookings (id, slot_id, user_id, status) 
	VALUES ($1, $2, $3, $4)
	`

	_, err := b.db.Pool.Exec(ctx, query, booking.ID, booking.SlotID, booking.UserID, booking.Status)
	return err
}

// Отмена бронирования
func (b *BookingRepository) Cancel(ctx context.Context, bookingID uuid.UUID) error {
	query := `
	UPDATE bookings SET status = 'cancelled' 
	WHERE id = $1 AND status != 'cancelled'
	`

	_, err := b.db.Pool.Exec(ctx, query, bookingID)
	return err
}

// Возвращает бронь по ID
func (b *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	query := `
	SELECT id, slot_id, user_id, status, created_at 
	FROM bookings 
	WHERE id = $1
	`
	var bok domain.Booking
	err := b.db.Pool.QueryRow(ctx, query, id).Scan(&bok.ID, &bok.SlotID, &bok.UserID, &bok.Status, &bok.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &bok, nil
}

// Возвращает брони пользователя (только на будущие слоты)
func (r *BookingRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	rows, err := r.db.Pool.Query(ctx, `
        SELECT b.id, b.slot_id, b.user_id, b.status, b.created_at
        FROM bookings b
        JOIN slots s ON s.id = b.slot_id
        WHERE b.user_id = $1 AND s.start_time >= NOW()
        ORDER BY s.start_time
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookings []domain.Booking
	for rows.Next() {
		var bk domain.Booking
		if err := rows.Scan(&bk.ID, &bk.SlotID, &bk.UserID, &bk.Status, &bk.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, bk)
	}
	return bookings, nil
}

// Возврат всех броней (пагинация)
func (r *BookingRepository) GetAllPaginated(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	offset := (page - 1) * pageSize

	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Pool.Query(ctx, `
        SELECT id, slot_id, user_id, status, created_at
        FROM bookings
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var bookings []domain.Booking
	for rows.Next() {
		var bk domain.Booking
		if err := rows.Scan(&bk.ID, &bk.SlotID, &bk.UserID, &bk.Status, &bk.CreatedAt); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, bk)
	}
	return bookings, total, nil
}
