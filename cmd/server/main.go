package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	musicdb "github.com/GoreeCloud/goreecloud-music/internal/database"
	"github.com/GoreeCloud/goreecloud-music/internal/httpapi"
	"github.com/GoreeCloud/goreecloud-music/internal/ingest"
	"github.com/GoreeCloud/goreecloud-music/internal/metadata"
	"github.com/GoreeCloud/goreecloud-music/internal/migrate"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

func main() {
	addr := env("GOREECLOUD_MUSIC_ADDR", ":8080")

	handler := httpapi.NewRouter()
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL != "" {
		connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		db, err := musicdb.OpenPostgres(connectCtx, databaseURL)
		cancel()
		if err != nil {
			log.Fatalf("connect to PostgreSQL: %v", err)
		}
		defer db.Close()

		if envBool("GOREECLOUD_MUSIC_AUTO_MIGRATE", false) {
			migrationCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := migrate.New().Apply(migrationCtx, db)
			cancel()
			if err != nil {
				log.Fatalf("apply database migrations: %v", err)
			}
			log.Print("database migrations applied")
		}

		postgresStore := store.NewPostgres(db)
		handler = httpapi.NewRouterWithDependencies(httpapi.Dependencies{
			Store:     postgresStore,
			ScanStore: postgresStore,
			Scanner:   ingest.New(postgresStore, metadata.Probe),
		})
	} else {
		log.Print("DATABASE_URL is not set; persistence-backed endpoints are unavailable")
	}

	handler = auth.DevelopmentIdentity{
		Enabled: envBool("GOREECLOUD_MUSIC_DEV_IDENTITY", false),
	}.Middleware(handler)

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("GoreeCloud Music API listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		log.Printf("ignoring invalid boolean %s=%q", key, value)
		return fallback
	}
}
