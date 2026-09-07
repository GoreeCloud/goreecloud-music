package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/playlists"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

const maxPlaylistRequestBodyBytes = 16 * 1024

type PlaylistService interface {
	List(ctx context.Context, userID string) ([]playlists.CatalogEntry, error)
	Create(ctx context.Context, userID, playlistID, name string) (playlists.Record, error)
	Load(ctx context.Context, userID, playlistID string) (playlists.Record, playlists.Playlist, error)
	Rename(ctx context.Context, userID, playlistID, name string) (playlists.Record, error)
	Append(ctx context.Context, userID, playlistID, trackID string) (playlists.Record, error)
	Insert(ctx context.Context, userID, playlistID string, index int, trackID string) (playlists.Record, error)
	RemoveAt(ctx context.Context, userID, playlistID string, index int) (playlists.Record, string, error)
	Move(ctx context.Context, userID, playlistID string, from, to int) (playlists.Record, error)
}

type playlistAPIHandler struct {
	store    store.MusicStore
	service  PlaylistService
	operation string
}

type playlistSummaryResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	TrackCount int    `json:"trackCount"`
}

type playlistListResponse struct {
	Playlists []playlistSummaryResponse `json:"playlists"`
}

type playlistResponse struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	TrackIDs []string `json:"trackIds"`
	Revision string   `json:"revision"`
}

type createPlaylistRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type renamePlaylistRequest struct {
	Name string `json:"name"`
}

type playlistTrackRequest struct {
	TrackID string `json:"trackId"`
}

type insertPlaylistTrackRequest struct {
	Index   int    `json:"index"`
	TrackID string `json:"trackId"`
}

type movePlaylistTrackRequest struct {
	From int `json:"from"`
	To   int `json:"to"`
}

func (h playlistAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedPlaylistUser(w, r, h.store)
	if !ok {
		return
	}
	if h.service == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "playlist service unavailable"})
		return
	}

	switch h.operation {
	case "list":
		h.list(w, r, userID)
	case "create":
		h.create(w, r, userID)
	case "get":
		h.get(w, r, userID)
	case "rename":
		h.rename(w, r, userID)
	case "append":
		h.append(w, r, userID)
	case "insert":
		h.insert(w, r, userID)
	case "move":
		h.move(w, r, userID)
	case "remove":
		h.remove(w, r, userID)
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "invalid playlist route"})
	}
}

func (h playlistAPIHandler) list(w http.ResponseWriter, r *http.Request, userID string) {
	entries, err := h.service.List(r.Context(), userID)
	if err != nil {
		writePlaylistError(w, err)
		return
	}
	response := playlistListResponse{Playlists: make([]playlistSummaryResponse, 0, len(entries))}
	for _, entry := range entries {
		response.Playlists = append(response.Playlists, playlistSummaryResponse{
			ID: entry.ID,
			Name: entry.Name,
			TrackCount: entry.TrackCount,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h playlistAPIHandler) create(w http.ResponseWriter, r *http.Request, userID string) {
	var body createPlaylistRequest
	if !decodePlaylistJSON(w, r, &body) {
		return
	}
	record, err := h.service.Create(r.Context(), userID, body.ID, body.Name)
	if err != nil {
		writePlaylistError(w, err)
		return
	}
	writePlaylistRecord(w, http.StatusCreated, record)
}

func (h playlistAPIHandler) get(w http.ResponseWriter, r *http.Request, userID string) {
	record, playlist, err := h.service.Load(r.Context(), userID, r.PathValue("playlistID"))
	if err != nil {
		writePlaylistError(w, err)
		return
	}
	writePlaylist(w, http.StatusOK, record, playlist)
}

func (h playlistAPIHandler) rename(w http.ResponseWriter, r *http.Request, userID string) {
	var body renamePlaylistRequest
	if !decodePlaylistJSON(w, r, &body) {
		return
	}
	record, err := h.service.Rename(r.Context(), userID, r.PathValue("playlistID"), body.Name)
	if err != nil {
		writePlaylistError(w, err)
		return
	}
	writePlaylistRecord(w, http.StatusOK, record)
}

func (h playlistAPIHandler) append(w http.ResponseWriter, r *http.Request, userID string) {
	var body playlistTrackRequest
	if !decodePlaylistJSON(w, r, &body) {
		return
	}
	record, err := h.service.Append(r.Context(), userID, r.PathValue("playlistID"), body.TrackID)
	if err != nil {
		writePlaylistError(w, err)
		return
	}
	writePlaylistRecord(w, http.StatusOK, record)
}

func (h playlistAPIHandler) insert(w http.ResponseWriter, r *http.Request, userID string) {
	var body insertPlaylistTrackRequest
	if !decodePlaylistJSON(w, r, &body) {
		return
	}
	record, err := h.service.Insert(r.Context(), userID, r.PathValue("playlistID"), body.Index, body.TrackID)
	if err != nil {
		writePlaylistError(w, err)
		return
	}
	writePlaylistRecord(w, http.StatusOK, record)
}

func (h playlistAPIHandler) move(w http.ResponseWriter, r *http.Request, userID string) {
	var body movePlaylistTrackRequest
	if !decodePlaylistJSON(w, r, &body) {
		return
	}
	record, err := h.service.Move(r.Context(), userID, r.PathValue("playlistID"), body.From, body.To)
	if err != nil {
		writePlaylistError(w, err)
		return
	}
	writePlaylistRecord(w, http.StatusOK, record)
}

func (h playlistAPIHandler) remove(w http.ResponseWriter, r *http.Request, userID string) {
	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil || index < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid playlist track index"})
		return
	}
	record, removed, err := h.service.RemoveAt(r.Context(), userID, r.PathValue("playlistID"), index)
	if err != nil {
		writePlaylistError(w, err)
		return
	}
	playlist, restoreErr := record.Restore()
	if restoreErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "playlist result invalid"})
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Playlist playlistResponse `json:"playlist"`
		RemovedTrackID string `json:"removedTrackId"`
	}{
		Playlist: playlistHTTPResponse(record, playlist),
		RemovedTrackID: removed,
	})
}

func authenticatedPlaylistUser(w http.ResponseWriter, r *http.Request, musicStore store.MusicStore) (string, bool) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return "", false
	}
	if musicStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "music store unavailable"})
		return "", false
	}
	user, err := musicStore.UserByID(r.Context(), principal.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
			return "", false
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "user lookup failed"})
		return "", false
	}
	if user.ID != principal.UserID || user.Role != principal.Role {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "principal mismatch"})
		return "", false
	}
	return principal.UserID, true
}

func decodePlaylistJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxPlaylistRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid playlist request"})
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid playlist request"})
		return false
	}
	return true
}

func writePlaylistRecord(w http.ResponseWriter, status int, record playlists.Record) {
	playlist, err := record.Restore()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "playlist result invalid"})
		return
	}
	writePlaylist(w, status, record, playlist)
}

func writePlaylist(w http.ResponseWriter, status int, record playlists.Record, playlist playlists.Playlist) {
	writeJSON(w, status, playlistHTTPResponse(record, playlist))
}

func playlistHTTPResponse(record playlists.Record, playlist playlists.Playlist) playlistResponse {
	return playlistResponse{
		ID: playlist.ID(),
		Name: playlist.Name(),
		TrackIDs: playlist.TrackIDs(),
		Revision: strconv.FormatUint(uint64(record.Revision()), 10),
	}
}

func writePlaylistError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, playlists.ErrRecordNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "playlist not found"})
	case errors.Is(err, store.ErrPlaylistAlreadyExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "playlist already exists"})
	case errors.Is(err, playlists.ErrStaleRevision):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "playlist changed concurrently"})
	case errors.Is(err, playlists.ErrInvalidPlaylistState), errors.Is(err, playlists.ErrInvalidService):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid playlist request"})
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "playlist request canceled"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "playlist operation failed"})
	}
}
