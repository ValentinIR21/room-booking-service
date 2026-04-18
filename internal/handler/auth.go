package handler

import (
	"avito-talk/internal/api"
	"encoding/json"
	"net/http"
)

// PostDummyLogin POST /dummyLogin
func (s *Server) PostDummyLogin(w http.ResponseWriter, r *http.Request) {
	var req api.PostDummyLoginJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, api.INVALIDREQUEST, "invalid json")
		return
	}

	if !req.Role.Valid() {
		writeError(w, api.INVALIDREQUEST, "role must be admin or user")
		return
	}

	var userID string
	if req.Role == api.PostDummyLoginJSONBodyRoleAdmin {
		userID = "11111111-1111-1111-1111-111111111111"
	} else {
		userID = "22222222-2222-2222-2222-222222222222"
	}

	token, err := s.authService.GenerateToken(userID, string(req.Role))
	if err != nil {
		writeError(w, api.INTERNALERROR, "token generation failed")
		return
	}

	respondJSON(w, http.StatusOK, api.Token{Token: token})
}
