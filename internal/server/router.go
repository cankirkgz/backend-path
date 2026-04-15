package server

import (
	"net/http"

	"backend-path/internal/domain"
	"backend-path/internal/middleware"
	"backend-path/internal/service"
)

type Handlers struct {
	Auth        *AuthHandler
	AuthRefresh *AuthRefreshHandler
	User        *UserHandler
	Transaction *TransactionHandler
	Balance     *BalanceHandler
}

type RouterDependencies struct {
	TokenService *service.TokenService
}

func NewRouter(handlers Handlers, deps RouterDependencies) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte("method not allowed"))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	if handlers.Auth != nil {
		registerHandler := middleware.Chain(
			http.HandlerFunc(handlers.Auth.Register),
			middleware.RequireJSON,
			middleware.MaxBodyBytes(1024*1024),
		)

		loginHandler := middleware.Chain(
			http.HandlerFunc(handlers.Auth.Login),
			middleware.RequireJSON,
			middleware.MaxBodyBytes(1024*1024),
		)

		mux.Handle("/api/v1/auth/register", registerHandler)
		mux.Handle("/api/v1/auth/login", loginHandler)
	}

	if handlers.AuthRefresh != nil {
		refreshHandler := middleware.Chain(
			http.HandlerFunc(handlers.AuthRefresh.Refresh),
			middleware.RequireJSON,
			middleware.MaxBodyBytes(1024*1024),
		)

		mux.Handle("/api/v1/auth/refresh", refreshHandler)
	}

	if handlers.User != nil {
		updateUserHandler := middleware.Chain(
			http.HandlerFunc(handlers.User.Update),
			middleware.RequireJSON,
			middleware.MaxBodyBytes(1024*1024),
		)

		adminOnlyUserListHandler := middleware.Chain(
			http.HandlerFunc(handlers.User.List),
			middleware.Authentication(deps.TokenService),
			middleware.RequireRole(domain.RoleAdmin),
		)

		protectedUserGetHandler := middleware.Chain(
			http.HandlerFunc(handlers.User.GetByID),
			middleware.Authentication(deps.TokenService),
		)

		protectedUserUpdateHandler := middleware.Chain(
			updateUserHandler,
			middleware.Authentication(deps.TokenService),
		)

		adminOnlyUserDeleteHandler := middleware.Chain(
			http.HandlerFunc(handlers.User.Delete),
			middleware.Authentication(deps.TokenService),
			middleware.RequireRole(domain.RoleAdmin),
		)

		mux.Handle("/api/v1/users", adminOnlyUserListHandler)
		mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				protectedUserGetHandler.ServeHTTP(w, r)
			case http.MethodPut:
				protectedUserUpdateHandler.ServeHTTP(w, r)
			case http.MethodDelete:
				adminOnlyUserDeleteHandler.ServeHTTP(w, r)
			default:
				writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
					Error: "method not allowed",
				})
			}
		})
	}

	if handlers.Transaction != nil {
		creditHandler := middleware.Chain(
			http.HandlerFunc(handlers.Transaction.Credit),
			middleware.RequireJSON,
			middleware.MaxBodyBytes(1024*1024),
		)

		debitHandler := middleware.Chain(
			http.HandlerFunc(handlers.Transaction.Debit),
			middleware.RequireJSON,
			middleware.MaxBodyBytes(1024*1024),
		)

		transferHandler := middleware.Chain(
			http.HandlerFunc(handlers.Transaction.Transfer),
			middleware.RequireJSON,
			middleware.MaxBodyBytes(1024*1024),
		)

		protectedCreditHandler := middleware.Chain(
			creditHandler,
			middleware.Authentication(deps.TokenService),
		)

		protectedDebitHandler := middleware.Chain(
			debitHandler,
			middleware.Authentication(deps.TokenService),
		)

		protectedTransferHandler := middleware.Chain(
			transferHandler,
			middleware.Authentication(deps.TokenService),
		)

		protectedTransactionHistoryHandler := middleware.Chain(
			http.HandlerFunc(handlers.Transaction.GetByUserID),
			middleware.Authentication(deps.TokenService),
		)

		protectedTransactionGetByIDHandler := middleware.Chain(
			http.HandlerFunc(handlers.Transaction.GetByID),
			middleware.Authentication(deps.TokenService),
		)

		mux.Handle("/api/v1/transactions/credit", protectedCreditHandler)
		mux.Handle("/api/v1/transactions/debit", protectedDebitHandler)
		mux.Handle("/api/v1/transactions/transfer", protectedTransferHandler)

		mux.Handle("/api/v1/transactions/history", protectedTransactionHistoryHandler)
		mux.Handle("/api/v1/transactions/", protectedTransactionGetByIDHandler)
	}

	if handlers.Balance != nil {
		protectedBalanceCurrentHandler := middleware.Chain(
			http.HandlerFunc(handlers.Balance.GetCurrent),
			middleware.Authentication(deps.TokenService),
		)

		protectedBalanceHistoricalHandler := middleware.Chain(
			http.HandlerFunc(handlers.Balance.GetHistorical),
			middleware.Authentication(deps.TokenService),
		)

		protectedBalanceAtTimeHandler := middleware.Chain(
			http.HandlerFunc(handlers.Balance.GetAtTime),
			middleware.Authentication(deps.TokenService),
		)

		mux.Handle("/api/v1/balances/current", protectedBalanceCurrentHandler)
		mux.Handle("/api/v1/balances/historical", protectedBalanceHistoricalHandler)
		mux.Handle("/api/v1/balances/at-time", protectedBalanceAtTimeHandler)
	}

	return mux
}
