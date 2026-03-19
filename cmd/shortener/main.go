package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"shortener/internal/config"
	"shortener/internal/delivery/handler"
	"shortener/internal/repository/memory"
	"shortener/internal/repository/postgres"
	"shortener/internal/usecase"
	"shortener/pkg/generator"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("starting with storage", slog.String("storage", cfg.StorageType))

	var storage usecase.URLStorage

	switch cfg.StorageType {
	case config.StoragePostgres:
		db, err := sql.Open("postgres", cfg.DatabaseURL)
		if err != nil {
			logger.Error("db open error", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer db.Close()

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetConnMaxIdleTime(5 * time.Minute)

		pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := db.PingContext(pingCtx); err != nil {
			logger.Error("db ping error", slog.String("error", err.Error()))
			os.Exit(1)
		}

		storage = postgres.NewPostgresRepo(db)
		logger.Info("postgres connected")

	case config.StorageMemory:
		storage = memory.NewMemoryRepo()
		logger.Info("memory storage ready")

	default:
		logger.Error("unknown storage type", slog.String("type", cfg.StorageType))
		os.Exit(1)
	}

	gen := generator.New()
	uc := usecase.NewURLUseCase(storage, gen, cfg.MaxRetries)
	h := handler.NewHandler(uc, logger)
	router := handler.NewRouter(h)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		logger.Info("http server started", slog.String("addr", ":"+cfg.Port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("received stop signal")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("shutdown complete")
}
