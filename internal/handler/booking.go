package handler

import (
	"avito-talk/internal/api"
	"avito-talk/internal/domain"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// PostBookingsCreate POST /bookings/create
func (s *Server) PostBookingsCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(w, r)
	if !ok {
		return
	}

	var req api.PostBookingsCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, api.INVALIDREQUEST, "invalid json")
		return
	}

	booking, err := s.bookingService.CreateBooking(r.Context(), userID, req.SlotId)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSlotNotFound):
			writeError(w, api.SLOTNOTFOUND, "slot not found")
		case errors.Is(err, domain.ErrSlotInPast):
			writeError(w, api.INVALIDREQUEST, "slot is in the past")
		case errors.Is(err, domain.ErrSlotAlreadyBooked):
			writeError(w, api.SLOTALREADYBOOKED, "slot already booked")
		default:
			log.Printf("CreateBooking error: %v", err)
			writeError(w, api.INTERNALERROR, "create failed")
		}
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"booking": booking})
}

// GetBookingsMy GET /bookings/my
func (s *Server) GetBookingsMy(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(w, r)
	if !ok {
		return
	}

	bookings, err := s.bookingService.GetUserBookings(r.Context(), userID)
	if err != nil {
		log.Printf("GetUserBookings error: %v", err)
		writeError(w, api.INTERNALERROR, "db error")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"bookings": bookings})
}

// GetBookingsList GET /bookings/list
func (s *Server) GetBookingsList(w http.ResponseWriter, r *http.Request, params api.GetBookingsListParams) {
	page := 1
	if params.Page != nil {
		page = *params.Page
	}
	if page < 1 {
		page = 1
	}

	pageSize := 20
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	bookings, total, err := s.bookingService.GetAllBookings(r.Context(), page, pageSize)
	if err != nil {
		log.Printf("GetAllBookings error: %v", err)
		writeError(w, api.INTERNALERROR, "db error")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"bookings": bookings,
		"pagination": api.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// PostBookingsBookingIdCancel POST /bookings/{bookingId}/cancel
func (s *Server) PostBookingsBookingIdCancel(w http.ResponseWriter, r *http.Request, bookingId api.BookingIdPath) {
	userID, ok := userIDFromContext(w, r)
	if !ok {
		return
	}

	booking, err := s.bookingService.CancelBooking(r.Context(), bookingId, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBookingNotFound):
			writeError(w, api.BOOKINGNOTFOUND, "booking not found")
		case errors.Is(err, domain.ErrNotOwner):
			writeError(w, api.FORBIDDEN, "cannot cancel another user's booking")
		default:
			log.Printf("CancelBooking error: %v", err)
			writeError(w, api.INTERNALERROR, "cancel failed")
		}
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"booking": booking})
}
