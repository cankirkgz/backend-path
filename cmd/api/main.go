package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"backend-path/internal/config"
	"backend-path/internal/logger"
	"backend-path/internal/server"
)

func main() {
	_ = godotenv.Load()

	cfg := config.MustLoad()
	logg := logger.New(cfg)

	srv := server.NewHTTPServer(cfg, logg)

	go func() {
		logg.Info("server starting", "port", cfg.Port, "env", cfg.AppEnv)

		if err := srv.ListenAndServe(); err != nil && err.Error() != "http: Server closed" {
			logg.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logg.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logg.Error("shutdown failed", "err", err)
		os.Exit(1)
	}

	logg.Info("server stopped cleanly")
}
