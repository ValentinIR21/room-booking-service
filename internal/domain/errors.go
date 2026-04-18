package domain

import "errors"

var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrScheduleExists    = errors.New("schedule already exists")
	ErrInvalidSchedule = errors.New("invalid schedule parameters")
	ErrSlotNotFound    = errors.New("slot not found")
	ErrSlotInPast        = errors.New("slot is in the past")
	ErrSlotAlreadyBooked = errors.New("slot already booked")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrNotOwner          = errors.New("cannot cancel another user's booking")
)
