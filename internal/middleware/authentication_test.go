package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/service"
)

func TestAuthentication_AllowsValidToken(t *testing.T) {
	tokenService := service.NewTokenService(1 * time.Hour)

	token, err := tokenService.Generate(&domain.User{
		ID:   1,
		Role: domain.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("unexpected token generate error: %v", err)
	}

	handler := Authentication(tokenService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetAuthenticatedUserID(r)
		role := GetAuthenticatedUserRole(r)

		if userID != 1 {
			t.Fatalf("expected user id 1, got %d", userID)
		}

		if role != domain.RoleAdmin {
			t.Fatalf("expected role %s, got %s", domain.RoleAdmin, role)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAuthentication_RejectsMissingHeader(t *testing.T) {
	tokenService := service.NewTokenService(1 * time.Hour)

	handler := Authentication(tokenService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthentication_RejectsInvalidToken(t *testing.T) {
	tokenService := service.NewTokenService(1 * time.Hour)

	handler := Authentication(tokenService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}
