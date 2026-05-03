package server

import (
	"encoding/json"
	"net/http"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type ScheduledTransactionHandler struct {
	scheduledService interfaces.ScheduledTransactionService
}

func NewScheduledTransactionHandler(scheduledService interfaces.ScheduledTransactionService) *ScheduledTransactionHandler {
	return &ScheduledTransactionHandler{
		scheduledService: scheduledService,
	}
}

type scheduleTransactionRequest struct {
	FromUserID int64                  `json:"from_user_id"`
	ToUserID   int64                  `json:"to_user_id"`
	Amount     float64                `json:"amount"`
	Type       domain.TransactionType `json:"type"`
	RunAt      string                 `json:"run_at"`
}

func (h *ScheduledTransactionHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	var req scheduleTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	runAt, err := time.Parse(time.RFC3339, req.RunAt)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "run_at must be a valid RFC3339 datetime",
		})
		return
	}

	scheduledTx := &domain.ScheduledTransaction{
		FromUserID: req.FromUserID,
		ToUserID:   req.ToUserID,
		Amount:     req.Amount,
		Type:       req.Type,
		RunAt:      runAt,
	}

	if err := h.scheduledService.Schedule(r.Context(), scheduledTx); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, scheduledTx)
}

func (h *ScheduledTransactionHandler) ProcessDue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	if err := h.scheduledService.ProcessDue(r.Context(), time.Now()); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "processed due scheduled transactions",
	})
}
