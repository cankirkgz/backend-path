package domain

import (
	"strings"
	"time"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return ErrInvalidUsername
	}

	if strings.TrimSpace(u.Email) == "" || !strings.Contains(u.Email, "@") {
		return ErrInvalidEmail
	}

	if strings.TrimSpace(u.PasswordHash) == "" {
		return ErrInvalidPasswordHash
	}

	if u.Role != RoleUser && u.Role != RoleAdmin {
		return ErrInvalidRole
	}

	return nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func NewUser(username, email, passwordHash string, role UserRole) (*User, error) {
	if role == "" {
		role = RoleUser
	}

	now := time.Now()

	user := &User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}
