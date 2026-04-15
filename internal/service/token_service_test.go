package service

import (
	"testing"
	"time"

	"backend-path/internal/domain"
)

func TestTokenServiceGenerateAndParse(t *testing.T) {
	tokenService := NewTokenService(1 * time.Hour)

	user := &domain.User{
		ID:   1,
		Role: domain.RoleAdmin,
	}

	token, err := tokenService.Generate(user)
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	payload, err := tokenService.Parse(token)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if payload.UserID != user.ID {
		t.Fatalf("expected user id %d, got %d", user.ID, payload.UserID)
	}

	if payload.Role != user.Role {
		t.Fatalf("expected role %s, got %s", user.Role, payload.Role)
	}
}

func TestTokenServiceParse_InvalidToken(t *testing.T) {
	tokenService := NewTokenService(1 * time.Hour)

	_, err := tokenService.Parse("not-a-valid-token")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}

func TestTokenServiceParse_ExpiredToken(t *testing.T) {
	tokenService := NewTokenService(-1 * time.Hour)

	user := &domain.User{
		ID:   1,
		Role: domain.RoleUser,
	}

	token, err := tokenService.Generate(user)
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	_, err = tokenService.Parse(token)
	if err == nil {
		t.Fatalf("expected expired token error but got nil")
	}
}
