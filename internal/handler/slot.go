package handler

import (
	"avito-talk/internal/api"
	"avito-talk/internal/domain"
	"errors"
	"log"
	"net/http"
)

// GetRoomsRoomIdSlotsList GET /rooms/{roomId}/slots/list
func (s *Server) GetRoomsRoomIdSlotsList(w http.ResponseWriter, r *http.Request, roomId api.RoomIdPath, params api.GetRoomsRoomIdSlotsListParams) {
	date := params.Date.Time.Format("2006-01-02")

	slots, err := s.slotService.GetAvailableSlots(r.Context(), roomId, date)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRoomNotFound):
			writeError(w, api.ROOMNOTFOUND, "room not found")
		case errors.Is(err, domain.ErrInvalidDateFormat):
			writeError(w, api.INVALIDREQUEST, "invalid date format")
		default:
			log.Printf("GetAvailableSlots error: %v", err)
			writeError(w, api.INTERNALERROR, "failed to get slots")
		}
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"slots": slots})
}
