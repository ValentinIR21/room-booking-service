package handler

import (
	"avito-talk/internal/api"
	"avito-talk/internal/domain"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// PostRoomsRoomIdScheduleCreate POST /rooms/{roomId}/schedule/create
func (s *Server) PostRoomsRoomIdScheduleCreate(w http.ResponseWriter, r *http.Request, roomId api.RoomIdPath) {
	var req api.PostRoomsRoomIdScheduleCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, api.INVALIDREQUEST, "invalid json")
		return
	}

	schedule, err := s.scheduleService.CreateSchedule(r.Context(), roomId, req.DaysOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRoomNotFound):
			writeError(w, api.ROOMNOTFOUND, "room not found")
		case errors.Is(err, domain.ErrScheduleExists):
			writeError(w, api.SCHEDULEEXISTS, "schedule already exists")
		case errors.Is(err, domain.ErrInvalidSchedule):
			writeError(w, api.INVALIDREQUEST, err.Error())
		default:
			log.Printf("CreateSchedule error: %v", err)
			writeError(w, api.INTERNALERROR, "create failed")
		}
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"schedule": schedule})
}
