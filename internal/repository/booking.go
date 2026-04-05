package repository

import (
	"avito-talk/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type BookingRepository struct {
	db *DB
}

func NewBookingRepository(db *DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create создаёт бронирование
func (b *BookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	query := `
	INSERT INTO bookings (id, slot_id, user_id, status, conference_link)
	VALUES ($1, $2, $3, $4, $5)
	`
	_, err := b.db.Pool.Exec(ctx, query, booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "idx_unique_active_booking" {
			return domain.ErrSlotAlreadyBooked
		}
		return err
	}
	return nil
}

// Cancel отменяет бронирование
func (b *BookingRepository) Cancel(ctx context.Context, bookingID uuid.UUID) error {
	query := `
	UPDATE bookings SET status = 'cancelled'
	WHERE id = $1 AND status != 'cancelled'
	`
	_, err := b.db.Pool.Exec(ctx, query, bookingID)
	return err
}

// GetByID возвращает бронь по ID
func (b *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	query := `
	SELECT id, slot_id, user_id, status, conference_link, created_at
	FROM bookings
	WHERE id = $1
	`

	var bk domain.Booking
	err := b.db.Pool.QueryRow(ctx, query, id).Scan(&bk.ID, &bk.SlotID, &bk.UserID, &bk.Status, &bk.ConferenceLink, &bk.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, err
	}

	return &bk, nil
}

// GetByUser возвращает брони пользователя (только на будущие слоты)
func (b *BookingRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	rows, err := b.db.Pool.Query(ctx, `
		SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at
		FROM bookings b
		JOIN slots s ON s.id = b.slot_id
		WHERE b.user_id = $1 AND s.start_time >= NOW()
		ORDER BY s.start_time
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := []domain.Booking{}
	for rows.Next() {
		var bk domain.Booking
		if err := rows.Scan(&bk.ID, &bk.SlotID, &bk.UserID, &bk.Status, &bk.ConferenceLink, &bk.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, bk)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}

// GetAllPaginated возвращает все брони с пагинацией
func (b *BookingRepository) GetAllPaginated(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	offset := (page - 1) * pageSize

	var total int
	err := b.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := b.db.Pool.Query(ctx, `
		SELECT id, slot_id, user_id, status, conference_link, created_at
		FROM bookings
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	bookings := []domain.Booking{}
	for rows.Next() {
		var bk domain.Booking
		if err := rows.Scan(&bk.ID, &bk.SlotID, &bk.UserID, &bk.Status, &bk.ConferenceLink, &bk.CreatedAt); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, bk)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return bookings, total, nil
}
