package main

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

type RedisKeys struct {
	Queue           string
	ProcessedQueue  string
	JobStatusPrefix string
}

func (k RedisKeys) JobStatusKey(jobID string) string {
	return k.JobStatusPrefix + ":" + jobID
}

type Config struct {
	Port                string
	RedisAddr           string
	Keys                RedisKeys
	JobStatusTTLSeconds int64
}
