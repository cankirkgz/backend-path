package service

import (
	"context"
	"strings"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type UserService struct {
	userRepo interfaces.UserRepository
	hasher   PasswordHasher
}

func NewUserService(userRepo interfaces.UserRepository, hasher PasswordHasher) *UserService {
	return &UserService{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (s *UserService) Register(ctx context.Context, user *domain.User, plainPassword string) error {
	if user == nil {
		return domain.ErrUserNotFound
	}

	plainPassword = strings.TrimSpace(plainPassword)
	if plainPassword == "" {
		return domain.ErrInvalidPassword
	}

	existingUserByEmail, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		return err
	}
	if existingUserByEmail != nil {
		return domain.ErrEmailAlreadyExists
	}

	existingUserByUsername, err := s.userRepo.GetByUsername(ctx, user.Username)
	if err != nil {
		return err
	}
	if existingUserByUsername != nil {
		return domain.ErrUsernameAlreadyExists
	}

	hashedPassword, err := s.hasher.Hash(plainPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword

	if err := user.Validate(); err != nil {
		return err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) Authenticate(ctx context.Context, email, plainPassword string) (*domain.User, error) {
	email = strings.TrimSpace(email)
	plainPassword = strings.TrimSpace(plainPassword)

	if email == "" {
		return nil, domain.ErrInvalidEmail
	}

	if plainPassword == "" {
		return nil, domain.ErrInvalidPassword
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := s.hasher.Compare(user.PasswordHash, plainPassword); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrUserNotFound
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, domain.ErrInvalidEmail
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

func (s *UserService) AuthorizeAdmin(user *domain.User) error {
	if user == nil {
		return domain.ErrUnauthorized
	}

	if !user.IsAdmin() {
		return domain.ErrUnauthorized
	}

	return nil
}

func (s *UserService) List(ctx context.Context) ([]*domain.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	if users == nil {
		return []*domain.User{}, nil
	}

	return users, nil
}

func (s *UserService) Update(ctx context.Context, user *domain.User) error {
	if user == nil {
		return domain.ErrUserNotFound
	}

	if user.ID <= 0 {
		return domain.ErrUserNotFound
	}

	existingUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}

	if existingUser == nil {
		return domain.ErrUserNotFound
	}

	existingByEmail, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		return err
	}
	if existingByEmail != nil && existingByEmail.ID != user.ID {
		return domain.ErrEmailAlreadyExists
	}

	existingByUsername, err := s.userRepo.GetByUsername(ctx, user.Username)
	if err != nil {
		return err
	}
	if existingByUsername != nil && existingByUsername.ID != user.ID {
		return domain.ErrUsernameAlreadyExists
	}

	user.PasswordHash = existingUser.PasswordHash
	user.CreatedAt = existingUser.CreatedAt
	user.UpdatedAt = existingUser.UpdatedAt

	if err := user.Validate(); err != nil {
		return err
	}

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrUserNotFound
	}

	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existingUser == nil {
		return domain.ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, id)
}
