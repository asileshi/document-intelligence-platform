package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

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
