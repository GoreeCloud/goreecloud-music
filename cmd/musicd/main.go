package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	sqlitestore "github.com/GoreeCloud/goreecloud-music/internal/storage/sqlite"
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
	dbPath := os.Getenv("GOREECLOUD_MUSIC_DB")
	if dbPath == "" {
		dbPath = filepath.Join("data", "goreecloud-music.db")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		log.Fatalf("create application-state directory: %v", err)
	}

	startupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	store, err := sqlitestore.Open(startupCtx, dbPath)
	cancel()
	if err != nil {
		log.Fatalf("open application-state database: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close application-state database: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		payload := map[string]string{"status": "ok", "storage": "ok"}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := store.Ping(ctx); err != nil {
			status = http.StatusServiceUnavailable
			payload["status"] = "degraded"
			payload["storage"] = "unavailable"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(payload)
	})
	mux.HandleFunc("GET /api/v1/system/info", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(systemInfo{Name: "GoreeCloud Music", Version: version.Version, Lifecycle: version.Lifecycle, Milestone: "1-native-multi-user-library"})
	})

	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("GoreeCloud Music %s development service listening on %s", version.Version, addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
