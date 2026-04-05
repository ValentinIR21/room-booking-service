package handler

import (
	"avito-talk/internal/api"
	"encoding/json"
	"log"
	"net/http"
)

// GetRoomsList GET /rooms/list
func (s *Server) GetRoomsList(w http.ResponseWriter, r *http.Request) {
	rooms, err := s.roomService.ListRooms(r.Context())
	if err != nil {
		log.Printf("ListRooms error: %v", err)
		writeError(w, api.INTERNALERROR, "db error")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"rooms": rooms})
}

// PostRoomsCreate POST /rooms/create
func (s *Server) PostRoomsCreate(w http.ResponseWriter, r *http.Request) {
	var req api.PostRoomsCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, api.INVALIDREQUEST, "invalid json")
		return
	}
	if req.Name == "" {
		writeError(w, api.INVALIDREQUEST, "name required")
		return
	}

	room, err := s.roomService.CreateRoom(r.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		log.Printf("CreateRoom error: %v", err)
		writeError(w, api.INTERNALERROR, "create failed")
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"room": room})
}
