package service

import (
	"context"
	"testing"

	"backend-path/internal/domain"
)

type fakeUserRepository struct {
	data   map[int64]*domain.User
	nextID int64
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		data:   make(map[int64]*domain.User),
		nextID: 1,
	}
}

func (r *fakeUserRepository) Create(ctx context.Context, user *domain.User) error {
	user.ID = r.nextID
	r.nextID++
	r.data[user.ID] = user
	return nil
}

func (r *fakeUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user, ok := r.data[id]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (r *fakeUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, user := range r.data {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (r *fakeUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, user := range r.data {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}

func (r *fakeUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.data[user.ID] = user
	return nil
}

func (r *fakeUserRepository) Delete(ctx context.Context, id int64) error {
	delete(r.data, id)
	return nil
}

type fakePasswordHasher struct{}

func (h *fakePasswordHasher) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (h *fakePasswordHasher) Compare(hashedPassword, plainPassword string) error {
	if hashedPassword != "hashed:"+plainPassword {
		return domain.ErrInvalidCredentials
	}
	return nil
}

func TestUserServiceRegister_Success(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	service := NewUserService(userRepo, hasher)

	user := &domain.User{
		Username: "can",
		Email:    "can@example.com",
		Role:     domain.RoleUser,
	}

	err := service.Register(context.Background(), user, "123456")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID == 0 {
		t.Fatalf("expected user ID to be set, got %d", user.ID)
	}

	if user.PasswordHash != "hashed:123456" {
		t.Fatalf("expected password hash to be set, got %s", user.PasswordHash)
	}

	storedUser, ok := userRepo.data[user.ID]
	if !ok {
		t.Fatal("expected user to be stored in repository")
	}

	if storedUser.Email != "can@example.com" {
		t.Fatalf("expected email=can@example.com, got %s", storedUser.Email)
	}

	if storedUser.Username != "can" {
		t.Fatalf("expected username=can, got %s", storedUser.Username)
	}
}

func TestUserServiceRegister_EmailAlreadyExists(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	existingUser := &domain.User{
		ID:           1,
		Username:     "existing-user",
		Email:        "can@example.com",
		PasswordHash: "hashed:oldpass",
		Role:         domain.RoleUser,
	}
	userRepo.data[existingUser.ID] = existingUser
	userRepo.nextID = 2

	service := NewUserService(userRepo, hasher)

	user := &domain.User{
		Username: "new-user",
		Email:    "can@example.com",
		Role:     domain.RoleUser,
	}

	err := service.Register(context.Background(), user, "123456")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrEmailAlreadyExists {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}

	if len(userRepo.data) != 1 {
		t.Fatalf("expected repository size=1, got %d", len(userRepo.data))
	}
}

func TestUserServiceRegister_UsernameAlreadyExists(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	existingUser := &domain.User{
		ID:           1,
		Username:     "can",
		Email:        "old@example.com",
		PasswordHash: "hashed:oldpass",
		Role:         domain.RoleUser,
	}
	userRepo.data[existingUser.ID] = existingUser
	userRepo.nextID = 2

	service := NewUserService(userRepo, hasher)

	user := &domain.User{
		Username: "can",
		Email:    "new@example.com",
		Role:     domain.RoleUser,
	}

	err := service.Register(context.Background(), user, "123456")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrUsernameAlreadyExists {
		t.Fatalf("expected ErrUsernameAlreadyExists, got %v", err)
	}

	if len(userRepo.data) != 1 {
		t.Fatalf("expected repository size=1, got %d", len(userRepo.data))
	}
}

func TestUserServiceRegister_InvalidPassword(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	service := NewUserService(userRepo, hasher)

	user := &domain.User{
		Username: "can",
		Email:    "can@example.com",
		Role:     domain.RoleUser,
	}

	err := service.Register(context.Background(), user, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}

	if len(userRepo.data) != 0 {
		t.Fatalf("expected repository size=0, got %d", len(userRepo.data))
	}
}

func TestUserServiceAuthenticate_Success(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	existingUser := &domain.User{
		ID:           1,
		Username:     "can",
		Email:        "can@example.com",
		PasswordHash: "hashed:123456",
		Role:         domain.RoleUser,
	}
	userRepo.data[existingUser.ID] = existingUser
	userRepo.nextID = 2

	service := NewUserService(userRepo, hasher)

	user, err := service.Authenticate(context.Background(), "can@example.com", "123456")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("expected user, got nil")
	}

	if user.ID != 1 {
		t.Fatalf("expected user ID=1, got %d", user.ID)
	}

	if user.Email != "can@example.com" {
		t.Fatalf("expected email=can@example.com, got %s", user.Email)
	}
}

func TestUserServiceAuthenticate_UserNotFound(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	service := NewUserService(userRepo, hasher)

	user, err := service.Authenticate(context.Background(), "missing@example.com", "123456")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}

func TestUserServiceAuthenticate_WrongPassword(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	existingUser := &domain.User{
		ID:           1,
		Username:     "can",
		Email:        "can@example.com",
		PasswordHash: "hashed:123456",
		Role:         domain.RoleUser,
	}
	userRepo.data[existingUser.ID] = existingUser
	userRepo.nextID = 2

	service := NewUserService(userRepo, hasher)

	user, err := service.Authenticate(context.Background(), "can@example.com", "wrongpass")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}

func TestUserServiceAuthenticate_InvalidPassword(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	service := NewUserService(userRepo, hasher)

	user, err := service.Authenticate(context.Background(), "can@example.com", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}

	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}

func TestUserServiceAuthorizeAdmin_Success(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}
	service := NewUserService(userRepo, hasher)

	user := &domain.User{
		ID:           1,
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: "hashed:123456",
		Role:         domain.RoleAdmin,
	}

	err := service.AuthorizeAdmin(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUserServiceAuthorizeAdmin_Unauthorized(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}
	service := NewUserService(userRepo, hasher)

	user := &domain.User{
		ID:           1,
		Username:     "can",
		Email:        "can@example.com",
		PasswordHash: "hashed:123456",
		Role:         domain.RoleUser,
	}

	err := service.AuthorizeAdmin(user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func (r *fakeUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	users := make([]*domain.User, 0, len(r.data))

	for _, user := range r.data {
		users = append(users, user)
	}

	return users, nil
}

func TestUserServiceList_Success(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	userRepo.data[1] = &domain.User{
		ID:           1,
		Username:     "can",
		Email:        "can@example.com",
		PasswordHash: "hashed:123456",
		Role:         domain.RoleUser,
	}

	userRepo.data[2] = &domain.User{
		ID:           2,
		Username:     "hasne",
		Email:        "hasne@example.com",
		PasswordHash: "hashed:654321",
		Role:         domain.RoleUser,
	}

	service := NewUserService(userRepo, hasher)

	users, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestUserServiceUpdate_Success(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	existingUser := &domain.User{
		ID:           1,
		Username:     "can",
		Email:        "can@example.com",
		PasswordHash: "hashed:123456",
		Role:         domain.RoleUser,
	}
	userRepo.data[1] = existingUser
	userRepo.nextID = 2

	service := NewUserService(userRepo, hasher)

	updatedUser := &domain.User{
		ID:       1,
		Username: "can-new",
		Email:    "cannew@example.com",
		Role:     domain.RoleUser,
	}

	err := service.Update(context.Background(), updatedUser)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	storedUser, ok := userRepo.data[1]
	if !ok {
		t.Fatal("expected user to exist in repository")
	}

	if storedUser.Username != "can-new" {
		t.Fatalf("expected username=can-new, got %s", storedUser.Username)
	}

	if storedUser.Email != "cannew@example.com" {
		t.Fatalf("expected email=cannew@example.com, got %s", storedUser.Email)
	}

	if storedUser.PasswordHash != "hashed:123456" {
		t.Fatalf("expected password hash to stay same, got %s", storedUser.PasswordHash)
	}
}

func TestUserServiceUpdate_UserNotFound(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	service := NewUserService(userRepo, hasher)

	updatedUser := &domain.User{
		ID:       99,
		Username: "ghost",
		Email:    "ghost@example.com",
		Role:     domain.RoleUser,
	}

	err := service.Update(context.Background(), updatedUser)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserServiceDelete_Success(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	userRepo.data[1] = &domain.User{
		ID:           1,
		Username:     "can",
		Email:        "can@example.com",
		PasswordHash: "hashed:123456",
		Role:         domain.RoleUser,
	}

	service := NewUserService(userRepo, hasher)

	err := service.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, ok := userRepo.data[1]; ok {
		t.Fatal("expected user to be deleted")
	}
}

func TestUserServiceDelete_UserNotFound(t *testing.T) {
	userRepo := newFakeUserRepository()
	hasher := &fakePasswordHasher{}

	service := NewUserService(userRepo, hasher)

	err := service.Delete(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
