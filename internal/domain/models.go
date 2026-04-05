package domain

import (
	"time"

	"github.com/google/uuid"
)

// Роли и статусы
const (
	RoleAdmin              = "admin"
	RoleUser               = "user"
	BookingStatusActive    = "active"
	BookingStatusCancelled = "cancelled"
)

type User struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	CreatedAt *time.Time `json:"createdAt"`
}

type Room struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Capacity    *int       `json:"capacity"`
	CreatedAt   *time.Time `json:"createdAt"`
}

type Schedule struct {
	ID         uuid.UUID `json:"id"`
	RoomID     uuid.UUID `json:"roomId"`
	DaysOfWeek []int     `json:"daysOfWeek"`
	StartTime  string    `json:"startTime"`
	EndTime    string    `json:"endTime"`
}

type Slot struct {
	ID     uuid.UUID `json:"id"`
	RoomID uuid.UUID `json:"roomId"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type Booking struct {
	ID             uuid.UUID  `json:"id"`
	SlotID         uuid.UUID  `json:"slotId"`
	UserID         uuid.UUID  `json:"userId"`
	Status         string     `json:"status"` // active или cancelled
	ConferenceLink *string    `json:"conferenceLink"`
	CreatedAt      *time.Time `json:"createdAt"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}
