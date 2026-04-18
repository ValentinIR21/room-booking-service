package handler

import (
	"avito-talk/internal/api"
	"avito-talk/internal/service"
)

// Server реализует api.ServerInterface
type Server struct {
	api.Unimplemented
	authService     *service.AuthService
	roomService     *service.RoomService
	scheduleService *service.ScheduleService
	slotService     *service.SlotService
	bookingService  *service.BookingService
}

func NewServer(
	authService *service.AuthService,
	roomService *service.RoomService,
	scheduleService *service.ScheduleService,
	slotService *service.SlotService,
	bookingService *service.BookingService,
) *Server {
	return &Server{
		authService:     authService,
		roomService:     roomService,
		scheduleService: scheduleService,
		slotService:     slotService,
		bookingService:  bookingService,
	}
}
