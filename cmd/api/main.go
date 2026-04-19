package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fistream/fistream/internal/config"
	repo "github.com/fistream/fistream/internal/repository/postgres"
	"github.com/fistream/fistream/internal/service"
	httptransport "github.com/fistream/fistream/internal/transport/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel()}))

	if err := validateConfig(cfg); err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("ping postgres", "error", err)
		os.Exit(1)
	}

	store := repo.New(pool)
	rooms := service.NewRoomService(store, service.Config{
		ServiceAccessPassword: cfg.ServiceAccessPassword,
		RoomTTL:               cfg.RoomTTL,
		JitsiDomain:           cfg.JitsiDomain,
		JitsiAppID:            cfg.JitsiAppID,
		JitsiAppSecret:        cfg.JitsiAppSecret,
		JitsiAudience:         cfg.JitsiAudience,
		JitsiSubject:          cfg.JitsiSubject,
		JitsiTokenTTL:         cfg.JitsiTokenTTL,
		APITokenSecret:        cfg.APITokenSecret,
		APITokenTTL:           cfg.APITokenTTL,
	})

	go runCleanupLoop(ctx, logger, rooms, cfg.CleanupInterval)

	handler := httptransport.NewServer(cfg, rooms)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("api server started", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("api server stopped")
}

func runCleanupLoop(ctx context.Context, logger *slog.Logger, rooms *service.RoomService, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			closed, err := rooms.CloseExpiredRooms(ctx)
			if err != nil {
				logger.Warn("cleanup failed", "error", err)
				continue
			}
			if closed > 0 {
				logger.Info("expired rooms closed", "count", closed)
			}
		}
	}
}

func validateConfig(cfg config.Config) error {
	if cfg.JitsiAppSecret == "" {
		return fmt.Errorf("JITSI_APP_SECRET is required")
	}
	if cfg.APITokenSecret == "" {
		return fmt.Errorf("API_TOKEN_SECRET is required")
	}
	return nil
}

