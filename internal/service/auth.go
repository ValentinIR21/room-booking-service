package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Секретный ключ в виде байтов
type AuthService struct {
	secret []byte
}

// Конструктор для AuthService
func NewAuthService(secret string) *AuthService {
	return &AuthService{secret: []byte(secret)}
}

// GenerateToken создаёт новый JWT для пользователя с заданным ID и ролью
func (s *AuthService) GenerateToken(userID, role string) (string, error) {
	// Создаём claims (данные, которые будут внутри токена) в виде мапы
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	// Создаём новый токен с алгоритмом подписи HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// token.SignedString вычисляет подпись и склеивает три части через точки
	return token.SignedString(s.secret)
}

// ValidateToken проверяет токен, возвращает claims map и ошибку, если токен невалиден
func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {

	// Проверка токена на метод подписи и возврат секретной подписи
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неверный метод подписи")
		}
		return s.secret, nil
	}

	// Проверка подписи и целостности токена
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, err
	}

	// Извлечение данных (второй части) из токена в виде мапы
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
