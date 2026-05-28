package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, cancel := notifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := loadConfigFromEnv()

	rdb := newRedisClient(cfg.RedisAddr)
	defer func() {
		_ = rdb.Close()
	}()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("redis ping failed", "addr", cfg.RedisAddr, "err", err)
		os.Exit(1)
	}
	logger.Info(
		"redis connected",
		"addr", cfg.RedisAddr,
		"queue", cfg.Keys.Queue,
		"processed_queue", cfg.Keys.ProcessedQueue,
		"job_status_prefix", cfg.Keys.JobStatusPrefix,
		"job_status_ttl_seconds", cfg.JobStatusTTLSeconds,
	)

	srv := newHTTPServer(cfg.Port)
	startHTTPServer(logger, srv, cfg.Port)

	go func() {
		if err := runWorker(ctx, logger, rdb, cfg.Keys, cfg.JobStatusTTLSeconds); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("worker stopped unexpectedly", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received", "err", context.Cause(ctx))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("shutdown complete")
}
