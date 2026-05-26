package main

import (
	"testing"
	"time"
)

func TestLLMTimeoutFromEnv(t *testing.T) {
	t.Setenv("SWAY_LLM_TIMEOUT_SECONDS", "")
	if got := llmTimeoutFromEnv(); got != 30*time.Second {
		t.Fatalf("default timeout=%s, want 30s", got)
	}

	t.Setenv("SWAY_LLM_TIMEOUT_SECONDS", "45")
	if got := llmTimeoutFromEnv(); got != 45*time.Second {
		t.Fatalf("configured timeout=%s, want 45s", got)
	}

	t.Setenv("SWAY_LLM_TIMEOUT_SECONDS", "bad")
	if got := llmTimeoutFromEnv(); got != 30*time.Second {
		t.Fatalf("invalid timeout=%s, want default 30s", got)
	}

	t.Setenv("SWAY_LLM_TIMEOUT_SECONDS", "0")
	if got := llmTimeoutFromEnv(); got != 30*time.Second {
		t.Fatalf("zero timeout=%s, want default 30s", got)
	}
}
