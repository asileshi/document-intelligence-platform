package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	serviceName = "ingestion-worker"
	version     = "0.0.0"
)

type healthResponse struct {
	Status string `json:"status"`
}

type versionResponse struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

type ingestionJob struct {
	JobID   string         `json:"job_id"`
	Source  string         `json:"source"`
	Payload map[string]any `json:"payload"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	redisQueue := os.Getenv("REDIS_QUEUE")
	if redisQueue == "" {
		redisQueue = "ingestion:jobs"
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() {
		_ = rdb.Close()
	}()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("redis ping failed", "addr", redisAddr, "err", err)
		os.Exit(1)
	}
	logger.Info("redis connected", "addr", redisAddr, "queue", redisQueue)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	})
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, versionResponse{Service: serviceName, Version: version})
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("http server starting", "service", serviceName, "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped unexpectedly", "err", err)
			os.Exit(1)
		}
	}()

	go func() {
		if err := runWorker(ctx, logger, rdb, redisQueue); err != nil && !errors.Is(err, context.Canceled) {
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

func runWorker(ctx context.Context, logger *slog.Logger, rdb *redis.Client, queue string) error {
	logger.Info("worker loop starting", "queue", queue)
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		// BRPOP response is [queueName, payload]. With timeout, redis.Nil means “no item”.
		result, err := rdb.BRPop(ctx, 5*time.Second, queue).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			return fmt.Errorf("brpop failed: %w", err)
		}
		if len(result) != 2 {
			logger.Warn("unexpected brpop result", "result", result)
			continue
		}

		payload := result[1]
		payloadPreview := payload
		if len(payloadPreview) > 500 {
			payloadPreview = payloadPreview[:500] + "…"
		}
		logger.Info("job received", "queue", queue, "payload", payloadPreview)

		job, err := parseJob(payload)
		if err != nil {
			logger.Warn("job parse failed", "err", err)
			continue
		}

		logger.Info("job processed", "job_id", job.JobID, "source", job.Source)
	}
}

func parseJob(raw string) (ingestionJob, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ingestionJob{}, errors.New("empty payload")
	}

	var job ingestionJob
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		return ingestionJob{}, fmt.Errorf("invalid json: %w", err)
	}
	if job.JobID == "" {
		return ingestionJob{}, errors.New("missing job_id")
	}
	if job.Source == "" {
		job.Source = "unknown"
	}
	if job.Payload == nil {
		job.Payload = map[string]any{}
	}
	return job, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
