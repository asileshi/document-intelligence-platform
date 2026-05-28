package main

import (
	"encoding/json"
	"errors"
	"testing"
)

type deadletterEntryForTest struct {
	Error string `json:"error"`
	Raw   string `json:"raw"`
	TS    string `json:"ts"`
}

func TestMakeDeadletterPayload(t *testing.T) {
	payload, err := makeDeadletterPayload("not-json", errors.New("invalid json"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var entry deadletterEntryForTest
	if err := json.Unmarshal([]byte(payload), &entry); err != nil {
		t.Fatalf("expected valid json, got %v", err)
	}
	if entry.Raw != "not-json" {
		t.Fatalf("expected raw not-json, got %q", entry.Raw)
	}
	if entry.Error != "invalid json" {
		t.Fatalf("expected error invalid json, got %q", entry.Error)
	}
	if entry.TS == "" {
		t.Fatalf("expected ts to be set")
	}
}
