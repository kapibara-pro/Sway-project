package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kapibara-pro/sway-project/backend/internal/api"
	"github.com/kapibara-pro/sway-project/backend/internal/llm"
	"github.com/kapibara-pro/sway-project/backend/internal/store"
)

func main() {
	addr := os.Getenv("SWAY_BACKEND_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	llmTimeout := llmTimeoutFromEnv()
	llmClient := llm.NewClient(llmTimeout)
	llmClient.Debug = boolFromEnv("SWAY_LLM_DEBUG")
	srv := api.NewServer(store.NewMemoryStore(), llmClient)
	log.Printf("Sway backend listening on %s, llm_timeout=%s, llm_debug=%t", addr, llmTimeout, llmClient.Debug)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}

func llmTimeoutFromEnv() time.Duration {
	const defaultTimeout = 30 * time.Second
	raw := os.Getenv("SWAY_LLM_TIMEOUT_SECONDS")
	if raw == "" {
		return defaultTimeout
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		log.Printf("invalid SWAY_LLM_TIMEOUT_SECONDS=%q, using default %s", raw, defaultTimeout)
		return defaultTimeout
	}
	return time.Duration(seconds) * time.Second
}

func boolFromEnv(key string) bool {
	switch os.Getenv(key) {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	default:
		return false
	}
}
