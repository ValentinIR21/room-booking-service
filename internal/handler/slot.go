package handler

import (
	"avito-talk/internal/service"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SlotHandler struct {
	slotService *service.SlotService
}

func NewSlotHandler(ss *service.SlotService) *SlotHandler {
	return &SlotHandler{slotService: ss}
}

func (h *SlotHandler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"invalid room id"}}`, http.StatusBadRequest)
		return
	}
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"date is required"}}`, http.StatusBadRequest)
		return
	}
	slots, err := h.slotService.GetAvailableSlots(r.Context(), roomID, date)
	if err != nil {
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"failed to get slots"}}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"slots": slots})
}
