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

	return Config{
		Port:      port,
		RedisAddr: redisAddr,
		Keys: RedisKeys{
			Queue:           queue,
			ProcessedQueue:  processedQueue,
			JobStatusPrefix: jobStatusPrefix,
		},
		JobStatusTTLSeconds: jobStatusTTLSeconds,
	}
}
