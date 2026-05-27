package main

import "testing"

func TestParseJob_Valid(t *testing.T) {
	job, err := parseJob(`{"job_id":"job-1","source":"manual","payload":{"text":"hello"}}`)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if job.JobID != "job-1" {
		t.Fatalf("expected job_id job-1, got %q", job.JobID)
	}
	if job.Source != "manual" {
		t.Fatalf("expected source manual, got %q", job.Source)
	}
	if job.Payload == nil {
		t.Fatalf("expected payload not nil")
	}
	if got := job.Payload["text"]; got != "hello" {
		t.Fatalf("expected payload.text hello, got %#v", got)
	}
}

func TestParseJob_Defaults(t *testing.T) {
	job, err := parseJob(`{"job_id":"job-2"}`)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if job.Source != "unknown" {
		t.Fatalf("expected default source unknown, got %q", job.Source)
	}
	if job.Payload == nil {
		t.Fatalf("expected default payload map, got nil")
	}
}

func TestParseJob_EmptyPayload(t *testing.T) {
	_, err := parseJob("   ")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseJob_InvalidJSON(t *testing.T) {
	_, err := parseJob("not-json")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseJob_MissingJobID(t *testing.T) {
	_, err := parseJob(`{"source":"manual"}`)
	if err == nil {
		t.Fatalf("expected error")
	}
}
