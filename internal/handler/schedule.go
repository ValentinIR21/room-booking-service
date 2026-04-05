package handler

import (
	"avito-talk/internal/domain"
	"avito-talk/internal/service"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ScheduleHandler struct {
	scheduleService *service.ScheduleService
}

func NewScheduleHandler(ss *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: ss}
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(RoleKey).(string)
	if role != domain.RoleAdmin {
		http.Error(w, `{"error":{"code":"FORBIDDEN","message":"admin only"}}`, http.StatusForbidden)
		return
	}
	roomIDStr := chi.URLParam(r, "roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"invalid room id"}}`, http.StatusBadRequest)
		return
	}

	var req struct {
		DaysOfWeek []int  `json:"daysOfWeek"`
		StartTime  string `json:"startTime"`
		EndTime    string `json:"endTime"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"invalid json"}}`, http.StatusBadRequest)
		return
	}

	schedule, err := h.scheduleService.CreateSchedule(r.Context(), roomID, req.DaysOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		if err.Error() == "schedule already exists" {
			http.Error(w, `{"error":{"code":"SCHEDULE_EXISTS","message":"schedule already exists"}}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"create failed"}}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"schedule": schedule})
}
