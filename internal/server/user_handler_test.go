package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend-path/internal/domain"
)

type fakeUserQueryService struct{}

func (f *fakeUserQueryService) Register(ctx context.Context, user *domain.User, plainPassword string) error {
	return nil
}

func (f *fakeUserQueryService) Authenticate(ctx context.Context, email, plainPassword string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeUserQueryService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if id == 1 {
		return &domain.User{
			ID:        1,
			Username:  "can",
			Email:     "can@example.com",
			Role:      domain.RoleUser,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	return nil, domain.ErrUserNotFound
}

func (f *fakeUserQueryService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeUserQueryService) AuthorizeAdmin(user *domain.User) error {
	return nil
}

func (f *fakeUserQueryService) Update(ctx context.Context, user *domain.User) error {
	return nil
}

func (f *fakeUserQueryService) Delete(ctx context.Context, id int64) error {
	return nil
}

func TestUserHandlerGetByID_Success(t *testing.T) {
	handler := NewUserHandler(&fakeUserQueryService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1", nil)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUserHandlerGetByID_InvalidMethod(t *testing.T) {
	handler := NewUserHandler(&fakeUserQueryService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/1", nil)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestUserHandlerGetByID_InvalidID(t *testing.T) {
	handler := NewUserHandler(&fakeUserQueryService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/abc", nil)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUserHandlerGetByID_NotFound(t *testing.T) {
	handler := NewUserHandler(&fakeUserQueryService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/999", nil)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func (f *fakeUserQueryService) List(ctx context.Context) ([]*domain.User, error) {
	return []*domain.User{
		{
			ID:        1,
			Username:  "can",
			Email:     "can@example.com",
			Role:      domain.RoleUser,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			Username:  "hasne",
			Email:     "hasne@example.com",
			Role:      domain.RoleUser,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}, nil
}

func TestUserHandlerList_Success(t *testing.T) {
	handler := NewUserHandler(&fakeUserQueryService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUserHandlerList_InvalidMethod(t *testing.T) {
	handler := NewUserHandler(&fakeUserQueryService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestUserHandlerUpdate_Success(t *testing.T) {
	handler := NewUserHandler(&fakeUserQueryService{})

	body := `{"username":"can-new","email":"cannew@example.com","role":"user"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/1", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUserHandlerUpdate_InvalidID(t *testing.T) {
	handler := NewUserHandler(&fakeUserQueryService{})

	body := `{"username":"can-new","email":"cannew@example.com","role":"user"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/abc", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUserHandlerDelete_Success(t *testing.T) {
	handler := NewUserHandler(&fakeUserQueryService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/1", nil)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUserHandlerDelete_InvalidID(t *testing.T) {
	handler := NewUserHandler(&fakeUserQueryService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/abc", nil)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
