package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ten00m/golang-test-task/internal/config"
	"github.com/ten00m/golang-test-task/internal/logger"
	"github.com/ten00m/golang-test-task/internal/router"
	"github.com/ten00m/golang-test-task/internal/storage"
)

func main() {
	cfg := config.LoadConfig()

	log := logger.New(os.Stdout)

	defer func() {
		if r := recover(); r != nil {
			log.Error("panic recovered", slog.Any("error", r))
		}
	}()

	log.Info("configuration loaded",
		slog.String("http_addr", cfg.HTTPServer.Address),
	)

	db, err := storage.New(&cfg.PostgreSQL, log)
	if err != nil {
		log.Error("failed to initialize storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("failed to close storage", slog.String("error", err.Error()))
		}
	}()

	r := router.New(log, db)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	errorChan := make(chan struct{})

	go func() {
		log.Info("starting http server", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server is not started with error", err)
			errorChan <- struct{}{}
		}
	}()

	<-errorChan

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("server stopped")
}
