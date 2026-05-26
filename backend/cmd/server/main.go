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
	srv := api.NewServer(store.NewMemoryStore(), llm.NewClient(llmTimeout))
	log.Printf("Sway backend listening on %s, llm_timeout=%s", addr, llmTimeout)
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
