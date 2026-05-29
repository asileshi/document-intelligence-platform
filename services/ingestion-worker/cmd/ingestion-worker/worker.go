package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func runWorker(
	ctx context.Context,
	logger *slog.Logger,
	rdb *redis.Client,
	keys RedisKeys,
	jobStatusTTLSeconds int64,
	qdrant *QdrantClient,
	embedder Embedder,
	chunkSize int,
	chunkOverlap int,
) error {
	logger.Info(
		"worker loop starting",
		"queue", keys.Queue,
		"processed_queue", keys.ProcessedQueue,
		"job_status_prefix", keys.JobStatusPrefix,
	)

	if qdrant == nil {
		return fmt.Errorf("qdrant client is nil")
	}
	if embedder == nil {
		return fmt.Errorf("embedder is nil")
	}

	// Qdrant doesn't have a healthcheck in compose, so do a small startup wait.
	var lastEnsureErr error
	for i := 0; i < 30; i++ {
		if err := qdrant.EnsureCollection(ctx); err == nil {
			lastEnsureErr = nil
			break
		} else {
			lastEnsureErr = err
			logger.Warn("qdrant not ready yet", "err", err)
			select {
			case <-time.After(2 * time.Second):
			case <-ctx.Done():
				return context.Canceled
			}
		}
	}
	if lastEnsureErr != nil {
		return fmt.Errorf("qdrant ensure collection failed: %w", lastEnsureErr)
	}

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

		textAny, ok := job.Payload["text"]
		if !ok {
			logger.Warn("job missing payload.text; dropping", "job_id", job.JobID)
			continue
		}
		text, ok := textAny.(string)
		if !ok {
			logger.Warn("job payload.text is not a string; dropping", "job_id", job.JobID)
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			logger.Warn("job payload.text is empty; dropping", "job_id", job.JobID)
			continue
		}

		chunks := chunkText(text, chunkSize, chunkOverlap)
		if len(chunks) == 0 {
			logger.Warn("chunker produced no chunks; dropping", "job_id", job.JobID)
			continue
		}

		points := make([]QdrantPoint, 0, len(chunks))
		nowRFC3339 := time.Now().UTC().Format(time.RFC3339Nano)
		for i, chunk := range chunks {
			vec, err := embedder.Embed(chunk)
			if err != nil {
				logger.Warn("embed failed; dropping job", "job_id", job.JobID, "chunk_index", i, "err", err)
				points = nil
				break
			}
			points = append(points, QdrantPoint{
				ID:     deterministicPointID(job.JobID, i),
				Vector: vec,
				Payload: map[string]any{
					"job_id":      job.JobID,
					"source":      job.Source,
					"chunk_index": i,
					"text":        chunk,
					"created_at":  nowRFC3339,
				},
			})
		}
		if points == nil {
			continue
		}

		if err := qdrant.UpsertPoints(ctx, points); err != nil {
			logger.Warn("qdrant upsert failed; dropping job", "job_id", job.JobID, "err", err)
			continue
		}

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
