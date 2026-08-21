package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/httpapi"
)

func main() {
	addr := env("GOREECLOUD_MUSIC_ADDR", ":8080")

	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

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
