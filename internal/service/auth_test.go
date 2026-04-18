package service

import (
	"testing"
)

func TestGenerateAndValidateToken(t *testing.T) {
	auth := NewAuthService("test-secret")

	token, err := auth.GenerateToken("user-123", "admin")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}

	if claims["user_id"] != "user-123" {
		t.Errorf("user_id = %v, want user-123", claims["user_id"])
	}
	if claims["role"] != "admin" {
		t.Errorf("role = %v, want admin", claims["role"])
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	auth := NewAuthService("test-secret")

	_, err := auth.ValidateToken("invalid.token.here")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	auth1 := NewAuthService("secret-1")
	auth2 := NewAuthService("secret-2")

	token, _ := auth1.GenerateToken("user-1", "user")
	_, err := auth2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}
