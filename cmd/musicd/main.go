package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/version"
)

type systemInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Lifecycle string `json:"lifecycle"`
	Milestone string `json:"milestone"`
}

func main() {
	addr := os.Getenv("GOREECLOUD_MUSIC_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/v1/system/info", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(systemInfo{Name: "GoreeCloud Music", Version: version.Version, Lifecycle: version.Lifecycle, Milestone: "0-architecture-foundation"})
	})

	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("GoreeCloud Music %s development service listening on %s", version.Version, addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
