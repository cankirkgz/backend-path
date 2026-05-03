package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type BalanceHandler struct {
	balanceService interfaces.BalanceService
}

func NewBalanceHandler(balanceService interfaces.BalanceService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

func (h *BalanceHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
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

	currency := currencyFromQuery(r)

	amount, err := h.balanceService.GetCurrentAmountByCurrency(r.Context(), userID, currency)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":  userID,
		"amount":   amount,
		"currency": currency,
	})
}

func (h *BalanceHandler) GetHistorical(w http.ResponseWriter, r *http.Request) {
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

	logs, err := h.balanceService.GetHistory(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	type balanceHistoryItem struct {
		Action    string `json:"action"`
		Details   string `json:"details"`
		CreatedAt string `json:"created_at"`
	}

	response := make([]balanceHistoryItem, 0, len(logs))

	for _, log := range logs {
		response = append(response, balanceHistoryItem{
			Action:    log.Action,
			Details:   log.Details,
			CreatedAt: log.CreatedAt.Format(time.RFC1123),
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *BalanceHandler) GetAtTime(w http.ResponseWriter, r *http.Request) {
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

	atStr := r.URL.Query().Get("at")
	if atStr == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "at is required",
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

	atTime, err := time.Parse(time.RFC3339, atStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid at value, use RFC3339 format",
		})
		return
	}

	logs, err := h.balanceService.GetHistory(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	currency := currencyFromQuery(r)

	amount := 0.0

	for _, log := range logs {
		if log.CreatedAt.After(atTime) {
			continue
		}

		parsedAmount, ok := extractNewBalance(log.Details)
		if ok {
			amount = parsedAmount
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":  userID,
		"amount":   amount,
		"currency": currency,
	})
}

func currencyFromQuery(r *http.Request) domain.Currency {
	currency := domain.Currency(r.URL.Query().Get("currency"))
	if currency == "" {
		return domain.CurrencyTRY
	}

	return currency
}

func extractNewBalance(details string) (float64, bool) {
	idx := strings.LastIndex(details, "new_balance ")
	if idx == -1 {
		return 0, false
	}

	value := strings.TrimSpace(details[idx+len("new_balance "):])
	amount, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}

	return amount, true
}

type AuthRefreshHandler struct{}

func NewAuthRefreshHandler() *AuthRefreshHandler {
	return &AuthRefreshHandler{}
}

func (h *AuthRefreshHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	writeJSON(w, http.StatusNotImplemented, errorResponse{
		Error: "refresh token flow is not implemented yet",
	})
}

var _ = domain.RoleUser
