package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"backend-path/internal/cache"
	"backend-path/internal/config"
	"backend-path/internal/domain"
	"backend-path/internal/middleware"
	"backend-path/internal/repository/memory"
	"backend-path/internal/service"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewHTTPServer(cfg config.Config, logg *slog.Logger) *http.Server {
	ctx := context.Background()
	redisClient := cache.NewRedisClient(ctx, cfg, logg)
	cacheRepo := cache.NewRedisRepository(redisClient)

	userRepo := memory.NewUserRepository()
	passwordHasher := service.NewBcryptPasswordHasher()
	userService := service.NewUserService(userRepo, passwordHasher, cacheRepo)
	tokenService := service.NewTokenService(1 * time.Hour)

	adminUser := &domain.User{
		Username: "admin",
		Email:    "admin@example.com",
		Role:     domain.RoleAdmin,
	}

	if err := userService.Register(ctx, adminUser, "admin123"); err != nil {
		logg.Error("failed to seed admin user", "err", err)
	}

	balanceRepo := memory.NewBalanceRepository()
	auditLogRepo := memory.NewAuditLogRepository()
	balanceService := service.NewBalanceService(balanceRepo, auditLogRepo, cacheRepo)

	transactionRepo := memory.NewTransactionRepository()
	eventStore := memory.NewEventStoreRepository()
	eventCacheService := service.NewEventCacheService(eventStore, cacheRepo)
	if err := eventCacheService.SyncBalanceCache(ctx); err != nil {
		logg.Error("failed to warm up balance cache", "err", err)
	}
	transactionService := service.NewTransactionService(transactionRepo, balanceService, eventStore)

	scheduledTransactionRepo := memory.NewScheduledTransactionRepository()
	scheduledTransactionService := service.NewScheduledTransactionService(scheduledTransactionRepo, transactionService)

	authHandler := NewAuthHandler(userService, tokenService)
	authRefreshHandler := NewAuthRefreshHandler()
	userHandler := NewUserHandler(userService)
	transactionHandler := NewTransactionHandler(transactionService)
	balanceHandler := NewBalanceHandler(balanceService)

	scheduledTransactionHandler := NewScheduledTransactionHandler(scheduledTransactionService)

	router := NewRouter(
		Handlers{
			Auth:                 authHandler,
			AuthRefresh:          authRefreshHandler,
			User:                 userHandler,
			Transaction:          transactionHandler,
			Balance:              balanceHandler,
			ScheduledTransaction: scheduledTransactionHandler,
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
		middleware.Metrics,
		middleware.Security,
		middleware.CORS,
		rateLimiter.Middleware,
	)

	handler = otelhttp.NewHandler(handler, "backend-http-server")

	logg.Info("router initialized")

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
