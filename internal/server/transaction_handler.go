package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type TransactionHandler struct {
	transactionService interfaces.TransactionService
}

func NewTransactionHandler(transactionService interfaces.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

type creditTransactionRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type transferTransactionRequest struct {
	FromUserID int64   `json:"from_user_id"`
	ToUserID   int64   `json:"to_user_id"`
	Amount     float64 `json:"amount"`
}

type transactionResponse struct {
	ID         int64                    `json:"id"`
	FromUserID int64                    `json:"from_user_id"`
	ToUserID   int64                    `json:"to_user_id"`
	Amount     float64                  `json:"amount"`
	Type       domain.TransactionType   `json:"type"`
	Status     domain.TransactionStatus `json:"status"`
	CreatedAt  string                   `json:"created_at"`
}

func (h *TransactionHandler) Credit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	var req creditTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	tx, err := h.transactionService.Credit(r.Context(), req.UserID, req.Amount)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTransactionResponse(tx))
}

func (h *TransactionHandler) Debit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	var req creditTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	tx, err := h.transactionService.Debit(r.Context(), req.UserID, req.Amount)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTransactionResponse(tx))
}

func toTransactionResponse(tx *domain.Transaction) transactionResponse {
	createdAt := ""
	if !tx.CreatedAt.IsZero() {
		createdAt = tx.CreatedAt.Format(time.RFC1123)
	}

	return transactionResponse{
		ID:         tx.ID,
		FromUserID: tx.FromUserID,
		ToUserID:   tx.ToUserID,
		Amount:     tx.Amount,
		Type:       tx.Type,
		Status:     tx.Status,
		CreatedAt:  createdAt,
	}
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	var req transferTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	tx, err := h.transactionService.Transfer(
		r.Context(),
		req.FromUserID,
		req.ToUserID,
		req.Amount,
	)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTransactionResponse(tx))
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	id, err := parseIDFromPath(r.URL.Path, "/api/v1/transactions/")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid transaction id",
		})
		return
	}

	tx, err := h.transactionService.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	if tx == nil {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error: "transaction not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, toTransactionResponse(tx))
}

func (h *TransactionHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "user_id is required",
		})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid user_id",
		})
		return
	}

	txs, err := h.transactionService.GetByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	response := make([]transactionResponse, 0)

	for _, tx := range txs {
		response = append(response, toTransactionResponse(tx))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TransactionHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "user_id is required",
		})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid user_id",
		})
		return
	}

	amount, err := h.transactionService.GetByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	total := 0.0

	for _, tx := range amount {
		if tx.ToUserID == userID {
			total += tx.Amount
		}
		if tx.FromUserID == userID {
			total -= tx.Amount
		}
	}

	writeJSON(w, http.StatusOK, map[string]float64{
		"balance": total,
	})
}
