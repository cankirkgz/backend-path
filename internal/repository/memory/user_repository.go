package memory

import (
	"context"
	"sync"
	"time"

	"backend-path/internal/domain"
)

type UserRepository struct {
	mu     sync.RWMutex
	data   map[int64]*domain.User
	nextID int64
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		data:   make(map[int64]*domain.User),
		nextID: 1,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	user.ID = r.nextID
	user.CreatedAt = now
	user.UpdatedAt = now
	r.nextID++

	copied := *user
	r.data[user.ID] = &copied

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.data[id]
	if !ok {
		return nil, nil
	}

	copied := *user
	return &copied, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.data {
		if user.Email == email {
			copied := *user
			return &copied, nil
		}
	}

	return nil, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.data {
		if user.Username == username {
			copied := *user
			return &copied, nil
		}
	}

	return nil, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.UpdatedAt = time.Now()

	copied := *user
	r.data[user.ID] = &copied

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.data, id)
	return nil
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*domain.User, 0, len(r.data))

	for _, user := range r.data {
		copied := *user
		users = append(users, &copied)
	}

	return users, nil
}
