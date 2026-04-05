package handler

import (
	"avito-talk/internal/api"
	"avito-talk/internal/service"
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

// AuthMiddleware проверяет JWT и кладёт user_id и role в контекст.
// Применяется только к группам роутов, требующих авторизации.
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, api.UNAUTHORIZED, "missing token")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeError(w, api.UNAUTHORIZED, "invalid auth header")
				return
			}
			tokenString := parts[1]

			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				writeError(w, api.UNAUTHORIZED, "invalid or expired token")
				return
			}

			userID, ok := claims["user_id"].(string)
			if !ok {
				writeError(w, api.UNAUTHORIZED, "user_id missing in token")
				return
			}
			role, ok := claims["role"].(string)
			if !ok {
				writeError(w, api.UNAUTHORIZED, "role missing in token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole возвращает middleware, который проверяет, что роль пользователя входит в список допустимых.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actual, ok := r.Context().Value(RoleKey).(string)
			if !ok {
				writeError(w, api.FORBIDDEN, "insufficient permissions")
				return
			}
			if _, found := allowed[actual]; !found {
				writeError(w, api.FORBIDDEN, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// userIDFromContext извлекает и парсит UUID пользователя из контекста. Возвращает false и пишет 401 при ошибке.
func userIDFromContext(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userIDStr, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		writeError(w, api.UNAUTHORIZED, "user id missing in token")
		return uuid.UUID{}, false
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, api.UNAUTHORIZED, "invalid user id in token")
		return uuid.UUID{}, false
	}
	return userID, true
}
