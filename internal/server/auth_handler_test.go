package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/service"
)

func newTestTokenService() *service.TokenService {
	return service.NewTokenService(1 * time.Hour)
}

type fakeUserService struct {
	registerFn     func(ctx context.Context, user *domain.User, plainPassword string) error
	authenticateFn func(ctx context.Context, email, plainPassword string) (*domain.User, error)
}

func (f *fakeUserService) Register(ctx context.Context, user *domain.User, plainPassword string) error {
	return f.registerFn(ctx, user, plainPassword)
}

func (f *fakeUserService) Authenticate(ctx context.Context, email, plainPassword string) (*domain.User, error) {
	return f.authenticateFn(ctx, email, plainPassword)
}

func (f *fakeUserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return nil, nil
}

func (f *fakeUserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeUserService) AuthorizeAdmin(user *domain.User) error {
	return nil
}

func (f *fakeUserService) List(ctx context.Context) ([]*domain.User, error) {
	return []*domain.User{}, nil
}

func (f *fakeUserService) Update(ctx context.Context, user *domain.User) error {
	return nil
}

func (f *fakeUserService) Delete(ctx context.Context, id int64) error {
	return nil
}

func TestAuthHandlerRegister_Success(t *testing.T) {
	handler := NewAuthHandler(&fakeUserService{
		registerFn: func(ctx context.Context, user *domain.User, plainPassword string) error {
			user.ID = 1
			user.CreatedAt = time.Now()
			user.UpdatedAt = time.Now()
			return nil
		},
	}, newTestTokenService())

	body := registerRequest{
		Username: "can",
		Email:    "can@example.com",
		Password: "123456",
	}

	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(payload))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestAuthHandlerRegister_InvalidMethod(t *testing.T) {
	handler := NewAuthHandler(&fakeUserService{}, newTestTokenService())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/register", nil)
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestAuthHandlerLogin_Success(t *testing.T) {
	handler := NewAuthHandler(&fakeUserService{
		authenticateFn: func(ctx context.Context, email, plainPassword string) (*domain.User, error) {
			return &domain.User{
				ID:        1,
				Username:  "can",
				Email:     "can@example.com",
				Role:      domain.RoleUser,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}, newTestTokenService())

	body := loginRequest{
		Email:    "can@example.com",
		Password: "123456",
	}

	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(payload))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response authMessageResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.AccessToken == "" {
		t.Fatalf("expected access token to be set")
	}
}

func TestAuthHandlerLogin_InvalidCredentials(t *testing.T) {
	handler := NewAuthHandler(&fakeUserService{
		authenticateFn: func(ctx context.Context, email, plainPassword string) (*domain.User, error) {
			return nil, domain.ErrInvalidCredentials
		},
	}, newTestTokenService())

	body := loginRequest{
		Email:    "can@example.com",
		Password: "wrong-password",
	}

	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(payload))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}
