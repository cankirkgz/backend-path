package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend-path/internal/domain"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, err error) {
	if err == nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "unknown error",
		})
		return
	}

	statusCode := http.StatusInternalServerError

	switch {
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidPassword),
		errors.Is(err, domain.ErrInvalidUsername):
		statusCode = http.StatusBadRequest

	case errors.Is(err, domain.ErrInsufficientBalance),
		errors.Is(err, domain.ErrInvalidTransactionAmount),
		errors.Is(err, domain.ErrInvalidTransactionType),
		errors.Is(err, domain.ErrInvalidTransactionUsers),
		errors.Is(err, domain.ErrSameSenderReceiver),
		errors.Is(err, domain.ErrInvalidBalanceAmount),
		errors.Is(err, domain.ErrInvalidBalanceUserID):
		statusCode = http.StatusBadRequest

	case errors.Is(err, domain.ErrEmailAlreadyExists),
		errors.Is(err, domain.ErrUsernameAlreadyExists):
		statusCode = http.StatusConflict

	case errors.Is(err, domain.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized

	case errors.Is(err, domain.ErrUserNotFound):
		statusCode = http.StatusNotFound
	}

	writeJSON(w, statusCode, errorResponse{
		Error: err.Error(),
	})
}
