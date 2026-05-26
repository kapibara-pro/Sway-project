package main

import (
	"log"
	"net/http"
	"os"
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
	srv := api.NewServer(store.NewMemoryStore(), llm.NewClient(8*time.Second))
	log.Printf("Sway backend listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
