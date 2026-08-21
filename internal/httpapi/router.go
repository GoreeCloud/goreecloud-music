package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

type aboutResponse struct {
	Name         string   `json:"name"`
	APIVersion   string   `json:"apiVersion"`
	Status       string   `json:"status"`
	Capabilities []string `json:"capabilities"`
}

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("GET /api/v1/about", about)
	return mux
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Service: "goreecloud-music",
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}

func about(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, aboutResponse{
		Name:       "GoreeCloud Music",
		APIVersion: "v1",
		Status:     "development",
		Capabilities: []string{
			"multi-user-domain-foundation",
			"multiple-libraries",
			"native-api",
			"open-subsonic-planned",
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
