package handler

import (
	"avito-talk/internal/service"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// DummyLogin /dummyLogin
func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	// Парсим тело запроса (JSON)
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"invalid json"}}`, http.StatusBadRequest)
		return
	}

	// Проверяем, что роль допустима
	if req.Role != "admin" && req.Role != "user" {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"role must be admin or user"}}`, http.StatusBadRequest)
		return
	}

	// Фиксированные UUID из init.sql
	var userID string
	if req.Role == "admin" {
		userID = "11111111-1111-1111-1111-111111111111"
	} else {
		userID = "22222222-2222-2222-2222-222222222222"
	}

	// Генерируем токен
	token, err := h.authService.GenerateToken(userID, req.Role)
	if err != nil {
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"token generation failed"}}`, http.StatusInternalServerError)
		return
	}

	// Отдаём JSON с токеном
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
