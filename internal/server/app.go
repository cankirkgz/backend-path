package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"backend-path/internal/config"
	"backend-path/internal/domain"
	"backend-path/internal/middleware"
	"backend-path/internal/repository/memory"
	"backend-path/internal/service"
)

func NewHTTPServer(cfg config.Config, logg *slog.Logger) *http.Server {
	userRepo := memory.NewUserRepository()
	passwordHasher := service.NewBcryptPasswordHasher()
	userService := service.NewUserService(userRepo, passwordHasher)
	tokenService := service.NewTokenService(1 * time.Hour)

	adminUser := &domain.User{
		Username: "admin",
		Email:    "admin@example.com",
		Role:     domain.RoleAdmin,
	}

	if err := userService.Register(context.Background(), adminUser, "admin123"); err != nil {
		logg.Error("failed to seed admin user", "err", err)
	}

	balanceRepo := memory.NewBalanceRepository()
	auditLogRepo := memory.NewAuditLogRepository()
	balanceService := service.NewBalanceService(balanceRepo, auditLogRepo)

	transactionRepo := memory.NewTransactionRepository()
	transactionService := service.NewTransactionService(transactionRepo, balanceService)

	authHandler := NewAuthHandler(userService, tokenService)
	authRefreshHandler := NewAuthRefreshHandler()
	userHandler := NewUserHandler(userService)
	transactionHandler := NewTransactionHandler(transactionService)
	balanceHandler := NewBalanceHandler(balanceService)

	router := NewRouter(
		Handlers{
			Auth:        authHandler,
			AuthRefresh: authRefreshHandler,
			User:        userHandler,
			Transaction: transactionHandler,
			Balance:     balanceHandler,
		},
		RouterDependencies{
			TokenService: tokenService,
		},
	)

	rateLimiter := middleware.NewRateLimiter(20, time.Minute)

	handler := middleware.Chain(
		router,
		middleware.RequestID,
		middleware.Recovery(logg),
		middleware.Logging(logg),
		middleware.Performance,
		middleware.Security,
		middleware.CORS,
		rateLimiter.Middleware,
	)

	logg.Info("router initialized")

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
