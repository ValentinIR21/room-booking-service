package handler

import (
	"avito-talk/internal/domain"
	"avito-talk/internal/service"
	"encoding/json"
	"net/http"
)

type RoomHandler struct {
	roomService *service.RoomService
}

func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

// GET /rooms/list
func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.roomService.ListRooms(r.Context())
	if err != nil {
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"db error"}}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rooms": rooms})
}

// POST /rooms/create (только admin)
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	// Проверяем роль из контекста (устанавливается middleware)
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != domain.RoleAdmin {
		http.Error(w, `{"error":{"code":"FORBIDDEN","message":"admin only"}}`, http.StatusForbidden)
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Capacity    *int    `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"invalid json"}}`, http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"name required"}}`, http.StatusBadRequest)
		return
	}

	room, err := h.roomService.CreateRoom(r.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"create failed"}}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"room": room})
}
