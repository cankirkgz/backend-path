package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type UserService struct {
	userRepo  interfaces.UserRepository
	hasher    PasswordHasher
	cacheRepo interfaces.CacheRepository
}

func NewUserService(
	userRepo interfaces.UserRepository,
	hasher PasswordHasher,
	cacheRepo interfaces.CacheRepository,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		hasher:    hasher,
		cacheRepo: cacheRepo,
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

	cacheKey := userCacheKey(id)

	if s.cacheRepo != nil {
		var cachedUser domain.User

		if err := s.cacheRepo.Get(ctx, cacheKey, &cachedUser); err == nil && cachedUser.ID != 0 {
			if err := cachedUser.Validate(); err == nil {
				return &cachedUser, nil
			}
		}
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, cacheKey, user, userCacheTTL)
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

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, userCacheKey(user.ID), user, userCacheTTL)
	}

	return nil
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

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	if s.cacheRepo != nil {
		_ = s.cacheRepo.Delete(ctx, userCacheKey(id))
	}

	return nil
}

const userCacheTTL = 10 * time.Minute

func userCacheKey(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}
