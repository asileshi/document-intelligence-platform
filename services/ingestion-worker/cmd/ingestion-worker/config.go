package main

import (
	"os"
	"strings"
)

func loadConfigFromEnv() Config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "9001"
	}

	redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	queue := strings.TrimSpace(os.Getenv("REDIS_QUEUE"))
	if queue == "" {
		queue = "ingestion:jobs"
	}

	processedQueue := strings.TrimSpace(os.Getenv("REDIS_PROCESSED_QUEUE"))
	if processedQueue == "" {
		processedQueue = "ingestion:processed"
	}

	jobStatusPrefix := strings.TrimSpace(os.Getenv("REDIS_JOB_STATUS_PREFIX"))
	if jobStatusPrefix == "" {
		jobStatusPrefix = "ingestion:job"
	}

	jobStatusTTLSeconds := int64(604800)
	if v := strings.TrimSpace(os.Getenv("JOB_STATUS_TTL_SECONDS")); v != "" {
		if parsed, err := parseInt64(v); err == nil && parsed > 0 {
			jobStatusTTLSeconds = parsed
		}
	}

	qdrantURL := strings.TrimSpace(os.Getenv("QDRANT_URL"))
	if qdrantURL == "" {
		qdrantURL = "http://qdrant:6333"
	}
	qdrantCollection := strings.TrimSpace(os.Getenv("QDRANT_COLLECTION"))
	if qdrantCollection == "" {
		qdrantCollection = "documents"
	}

	embeddingDim := 8
	if v := strings.TrimSpace(os.Getenv("EMBEDDING_DIM")); v != "" {
		if parsed, err := parseInt(v); err == nil && parsed > 0 {
			embeddingDim = parsed
		}
	}

	chunkSize := 800
	if v := strings.TrimSpace(os.Getenv("CHUNK_SIZE")); v != "" {
		if parsed, err := parseInt(v); err == nil && parsed > 0 {
			chunkSize = parsed
		}
	}
	chunkOverlap := 100
	if v := strings.TrimSpace(os.Getenv("CHUNK_OVERLAP")); v != "" {
		if parsed, err := parseInt(v); err == nil && parsed >= 0 {
			chunkOverlap = parsed
		}
	}

	return Config{
		Port:      port,
		RedisAddr: redisAddr,
		Keys: RedisKeys{
			Queue:           queue,
			ProcessedQueue:  processedQueue,
			JobStatusPrefix: jobStatusPrefix,
		},
		JobStatusTTLSeconds: jobStatusTTLSeconds,
		QdrantURL:           qdrantURL,
		QdrantCollection:    qdrantCollection,
		EmbeddingDim:        embeddingDim,
		ChunkSize:           chunkSize,
		ChunkOverlap:        chunkOverlap,
	}
}
