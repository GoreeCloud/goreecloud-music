package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/store"
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

type Dependencies struct {
	Store     store.MusicStore
	ScanStore LibraryScanStore
	Catalog   LibraryCatalogStore
	Artwork   AlbumArtworkStore
	Stream    TrackStreamStore
	Scanner   LibraryScanner
	Playlists PlaylistService
}

func NewRouter() http.Handler {
	return NewRouterWithDependencies(Dependencies{})
}

func NewRouterWithDependencies(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("GET /api/v1/about", about)
	mux.Handle("GET /api/v1/me/libraries", librariesHandler{store: deps.Store})
	mux.Handle("GET /api/v1/me/playlists", playlistAPIHandler{store: deps.Store, service: deps.Playlists, operation: "list"})
	mux.Handle("POST /api/v1/me/playlists", playlistAPIHandler{store: deps.Store, service: deps.Playlists, operation: "create"})
	mux.Handle("GET /api/v1/me/playlists/{playlistID}", playlistAPIHandler{store: deps.Store, service: deps.Playlists, operation: "get"})
	mux.Handle("PATCH /api/v1/me/playlists/{playlistID}", playlistAPIHandler{store: deps.Store, service: deps.Playlists, operation: "rename"})
	mux.Handle("POST /api/v1/me/playlists/{playlistID}/tracks", playlistAPIHandler{store: deps.Store, service: deps.Playlists, operation: "append"})
	mux.Handle("POST /api/v1/me/playlists/{playlistID}/tracks/insert", playlistAPIHandler{store: deps.Store, service: deps.Playlists, operation: "insert"})
	mux.Handle("POST /api/v1/me/playlists/{playlistID}/tracks/move", playlistAPIHandler{store: deps.Store, service: deps.Playlists, operation: "move"})
	mux.Handle("DELETE /api/v1/me/playlists/{playlistID}/tracks/{index}", playlistAPIHandler{store: deps.Store, service: deps.Playlists, operation: "remove"})
	mux.Handle("GET /api/v1/libraries/{libraryID}/albums", catalogHandler{store: deps.Store, catalog: deps.Catalog, kind: "albums"})
	mux.Handle("GET /api/v1/libraries/{libraryID}/albums/{albumID}/artwork", albumArtworkHandler{store: deps.Store, artwork: deps.Artwork})
	mux.Handle("GET /api/v1/libraries/{libraryID}/tracks", catalogHandler{store: deps.Store, catalog: deps.Catalog, kind: "tracks"})
	mux.Handle("GET /api/v1/libraries/{libraryID}/tracks/{trackID}/stream", trackStreamHandler{store: deps.Store, stream: deps.Stream})
	mux.Handle("POST /api/v1/libraries/{libraryID}/scan", libraryScanHandler{
		store:   deps.Store,
		scan:    deps.ScanStore,
		scanner: deps.Scanner,
	})
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
			"persistence-boundary",
			"authorization-boundaries",
			"scanner-reconciliation",
			"metadata-ingestion",
			"artwork-ingestion",
			"artwork-delivery",
			"library-catalog-browse",
			"byte-range-streaming",
			"native-playlists",
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
