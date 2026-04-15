package server

import (
	"encoding/json"
	"net/http"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
	"backend-path/internal/service"
)

type AuthHandler struct {
	userService  interfaces.UserService
	tokenService *service.TokenService
}

func NewAuthHandler(userService interfaces.UserService, tokenService *service.TokenService) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		tokenService: tokenService,
	}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authUserResponse struct {
	ID        int64           `json:"id"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	Role      domain.UserRole `json:"role"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

type authMessageResponse struct {
	Message     string           `json:"message"`
	User        authUserResponse `json:"user"`
	AccessToken string           `json:"access_token,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     domain.RoleUser,
	}

	if err := h.userService.Register(r.Context(), user, req.Password); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, authMessageResponse{
		Message: "user registered successfully",
		User:    toAuthUserResponse(user),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	user, err := h.userService.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	accessToken, err := h.tokenService.Generate(user)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, authMessageResponse{
		Message:     "login successful",
		User:        toAuthUserResponse(user),
		AccessToken: accessToken,
	})
}

func toAuthUserResponse(user *domain.User) authUserResponse {
	return authUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format(http.TimeFormat),
		UpdatedAt: user.UpdatedAt.Format(http.TimeFormat),
	}
}
