package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"backend-path/internal/domain"
)

var ErrInvalidToken = errors.New("invalid token")

type TokenPayload struct {
	UserID    int64           `json:"user_id"`
	Role      domain.UserRole `json:"role"`
	ExpiresAt int64           `json:"expires_at"`
}

type TokenService struct {
	tokenTTL time.Duration
}

func NewTokenService(tokenTTL time.Duration) *TokenService {
	return &TokenService{
		tokenTTL: tokenTTL,
	}
}

func (s *TokenService) Generate(user *domain.User) (string, error) {
	if user == nil {
		return "", domain.ErrUnauthorized
	}

	payload := TokenPayload{
		UserID:    user.ID,
		Role:      user.Role,
		ExpiresAt: time.Now().Add(s.tokenTTL).Unix(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	token := base64.StdEncoding.EncodeToString(data)
	return token, nil
}

func (s *TokenService) Parse(token string) (*TokenPayload, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrInvalidToken
	}

	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var payload TokenPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, ErrInvalidToken
	}

	if payload.UserID <= 0 {
		return nil, ErrInvalidToken
	}

	if payload.Role != domain.RoleUser && payload.Role != domain.RoleAdmin {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > payload.ExpiresAt {
		return nil, ErrInvalidToken
	}

	return &payload, nil
}
