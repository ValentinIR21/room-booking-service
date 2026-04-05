package handler

import (
	"avito-talk/internal/service"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

// AuthMiddleware возвращает функцию-обёртку, которая проверяет JWT и кладёт user_id и role в контекст
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Берём заголовок Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"missing token"}}`, http.StatusUnauthorized)
				return
			}

			// Ожидаем формат "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"invalid auth header"}}`, http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			// Валидируем токен
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"invalid or expired token"}}`, http.StatusUnauthorized)
				return
			}

			// Извлекаем user_id и role из claims
			userID, ok := claims["user_id"].(string)
			if !ok {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"user_id missing in token"}}`, http.StatusUnauthorized)
				return
			}
			role, ok := claims["role"].(string)
			if !ok {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"role missing in token"}}`, http.StatusUnauthorized)
				return
			}

			// Кладём эти значения в контекст запроса
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
