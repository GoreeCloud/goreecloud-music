package httpapi

import (
	"context"
	"net/http"

	"github.com/GoreeCloud/goreecloud-music/internal/playlists"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

type PlaylistDeletionService interface {
	Delete(ctx context.Context, userID, playlistID string) (playlists.Record, error)
}

type playlistDeleteHandler struct {
	store   store.MusicStore
	service PlaylistDeletionService
}

func (h playlistDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedPlaylistUser(w, r, h.store)
	if !ok {
		return
	}
	if h.service == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "playlist service unavailable"})
		return
	}
	playlistID := r.PathValue("playlistID")
	deleted, err := h.service.Delete(r.Context(), userID, playlistID)
	if err != nil {
		writePlaylistError(w, err)
		return
	}
	playlist, err := deleted.Restore()
	if err != nil || playlist.UserID() != userID || playlist.ID() != playlistID {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "playlist result invalid"})
		return
	}
	writeJSON(w, http.StatusOK, struct {
		ID       string `json:"id"`
		Revision string `json:"revision"`
		Deleted  bool   `json:"deleted"`
	}{
		ID:       playlist.ID(),
		Revision: formatPlaylistRevisionForAPI(deleted.Revision()),
		Deleted:  true,
	})
}

func formatPlaylistRevisionForAPI(revision playlists.Revision) string {
	return playlistHTTPResponseRevision(revision)
}
