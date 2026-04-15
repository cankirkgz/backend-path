package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type UserHandler struct {
	userService interfaces.UserService
}

type updateUserRequest struct {
	Username string          `json:"username"`
	Email    string          `json:"email"`
	Role     domain.UserRole `json:"role"`
}

func NewUserHandler(userService interfaces.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	id, err := parseIDFromPath(r.URL.Path, "/api/v1/users/")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid user id",
		})
		return
	}

	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAuthUserResponse(user))
}

func parseIDFromPath(path string, prefix string) (int64, error) {
	idPart := strings.TrimPrefix(path, prefix)
	return strconv.ParseInt(idPart, 10, 64)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	users, err := h.userService.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	response := make([]authUserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, toAuthUserResponse(user))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	id, err := parseIDFromPath(r.URL.Path, "/api/v1/users/")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid user id",
		})
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	user := &domain.User{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
	}

	if err := h.userService.Update(r.Context(), user); err != nil {
		writeError(w, err)
		return
	}

	updatedUser, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAuthUserResponse(updatedUser))
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	id, err := parseIDFromPath(r.URL.Path, "/api/v1/users/")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid user id",
		})
		return
	}

	if err := h.userService.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "user deleted successfully",
	})
}
