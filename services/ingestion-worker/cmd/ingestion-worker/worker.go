package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func runWorker(
	ctx context.Context,
	logger *slog.Logger,
	rdb *redis.Client,
	keys RedisKeys,
	jobStatusTTLSeconds int64,
) error {
	logger.Info(
		"worker loop starting",
		"queue", keys.Queue,
		"processed_queue", keys.ProcessedQueue,
		"job_status_prefix", keys.JobStatusPrefix,
	)

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		// BRPOP response is [queueName, payload]. With timeout, redis.Nil means “no item”.
		result, err := rdb.BRPop(ctx, 5*time.Second, keys.Queue).Result()
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
		logger.Info("job received", "queue", keys.Queue, "payload", payloadPreview)

		job, err := parseJob(payload)
		if err != nil {
			// For now we only track successful jobs.
			logger.Warn("job parse failed; dropping", "err", err)
			continue
		}

		logger.Info("job processed", "job_id", job.JobID, "source", job.Source)

		now := time.Now().UTC().Format(time.RFC3339Nano)
		statusKey := keys.JobStatusKey(job.JobID)

		tx := rdb.TxPipeline()
		tx.LPush(ctx, keys.ProcessedQueue, job.JobID)
		tx.HSet(
			ctx,
			statusKey,
			"job_id", job.JobID,
			"status", "processed",
			"processed_at", now,
			"updated_at", now,
		)
		tx.Expire(ctx, statusKey, time.Duration(jobStatusTTLSeconds)*time.Second)
		if _, err := tx.Exec(ctx); err != nil {
			return fmt.Errorf("ack/status transaction failed: %w", err)
		}
	}
}
