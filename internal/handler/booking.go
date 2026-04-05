package handler

import (
	"avito-talk/internal/domain"
	"avito-talk/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BookingHandler struct {
	bookingService *service.BookingService
}

func NewBookingHandler(bs *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bs}
}

// POST /bookings/create
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(RoleKey).(string)
	if role != domain.RoleUser {
		http.Error(w, `{"error":{"code":"FORBIDDEN","message":"only users can book"}}`, http.StatusForbidden)
		return
	}
	userIDStr := r.Context().Value(UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)

	var req struct {
		SlotID string `json:"slotId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"invalid json"}}`, http.StatusBadRequest)
		return
	}
	slotID, err := uuid.Parse(req.SlotID)
	if err != nil {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"invalid slotId"}}`, http.StatusBadRequest)
		return
	}

	booking, err := h.bookingService.CreateBooking(r.Context(), userID, slotID)
	if err != nil {
		if err.Error() == "slot not found" {
			http.Error(w, `{"error":{"code":"NOT_FOUND","message":"slot not found"}}`, http.StatusNotFound)
		} else if err.Error() == "slot is in the past" {
			http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"slot is in the past"}}`, http.StatusBadRequest)
		} else if err.Error() == "slot already booked" {
			http.Error(w, `{"error":{"code":"SLOT_ALREADY_BOOKED","message":"slot already booked"}}`, http.StatusConflict)
		} else {
			http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"create failed"}}`, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"booking": booking})
}

// GET /bookings/my
func (h *BookingHandler) MyBookings(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(RoleKey).(string)
	if role != domain.RoleUser {
		http.Error(w, `{"error":{"code":"FORBIDDEN","message":"only users can view their bookings"}}`, http.StatusForbidden)
		return
	}
	userIDStr := r.Context().Value(UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)

	bookings, err := h.bookingService.GetUserBookings(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"db error"}}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"bookings": bookings})
}

// GET /bookings/list (admin only)
func (h *BookingHandler) ListAllBookings(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(RoleKey).(string)
	if role != domain.RoleAdmin {
		http.Error(w, `{"error":{"code":"FORBIDDEN","message":"admin only"}}`, http.StatusForbidden)
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	bookings, total, err := h.bookingService.GetAllBookings(r.Context(), page, pageSize)
	if err != nil {
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"db error"}}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bookings": bookings,
		"pagination": map[string]int{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		},
	})
}

// POST /bookings/{bookingId}/cancel
func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(RoleKey).(string)
	if role != domain.RoleUser {
		http.Error(w, `{"error":{"code":"FORBIDDEN","message":"only users can cancel bookings"}}`, http.StatusForbidden)
		return
	}
	userIDStr := r.Context().Value(UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)

	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"invalid booking id"}}`, http.StatusBadRequest)
		return
	}
	booking, err := h.bookingService.CancelBooking(r.Context(), bookingID, userID)
	if err != nil {
		if err.Error() == "booking not found" {
			http.Error(w, `{"error":{"code":"NOT_FOUND","message":"booking not found"}}`, http.StatusNotFound)
		} else if err.Error() == "cannot cancel another user's booking" {
			http.Error(w, `{"error":{"code":"FORBIDDEN","message":"cannot cancel another user's booking"}}`, http.StatusForbidden)
		} else {
			http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"cancel failed"}}`, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"booking": booking})
}
