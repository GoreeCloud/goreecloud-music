package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

type TrackFavoriteStore interface {
	TrackForLibrary(ctx context.Context, libraryID, trackID string) (domain.Track, error)
	TrackFavoritesForUser(ctx context.Context, userID string) ([]domain.Track, error)
	SetTrackFavorite(ctx context.Context, userID, trackID string, favorite bool) error
}

type trackFavoriteHandler struct {
	store     store.MusicStore
	catalog   LibraryCatalogStore
	favorites TrackFavoriteStore
	operation string
}

type trackFavoriteResponse struct {
	TrackID  string `json:"trackId"`
	Favorite bool   `json:"favorite"`
}

func (h trackFavoriteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	principal, ok := authenticatedFavoritePrincipal(w, r, h.store)
	if !ok {
		return
	}
	if h.catalog == nil || h.favorites == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "favorites unavailable"})
		return
	}

	switch h.operation {
	case "list":
		h.list(w, r, principal)
	case "set":
		h.set(w, r, principal, true)
	case "unset":
		h.set(w, r, principal, false)
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "invalid favorites route"})
	}
}

func (h trackFavoriteHandler) list(w http.ResponseWriter, r *http.Request, principal auth.Principal) {
	tracks, err := h.favorites.TrackFavoritesForUser(r.Context(), principal.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "favorites lookup failed"})
		return
	}
	visible := make([]domain.Track, 0, len(tracks))
	readableLibraries := make(map[string]bool)
	checkedLibraries := make(map[string]bool)
	for _, track := range tracks {
		if !validFavoritePathID(track.ID) || !validFavoritePathID(track.LibraryID) {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "favorites result invalid"})
			return
		}
		if principal.Role == domain.RoleAdmin {
			visible = append(visible, track)
			continue
		}
		if !checkedLibraries[track.LibraryID] {
			membership, membershipErr := h.catalog.LibraryMembership(r.Context(), track.LibraryID, principal.UserID)
			switch {
			case membershipErr == nil:
				readableLibraries[track.LibraryID] = membership.CanRead
			case errors.Is(membershipErr, store.ErrNotFound):
				readableLibraries[track.LibraryID] = false
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "favorites authorization failed"})
				return
			}
			checkedLibraries[track.LibraryID] = true
		}
		if readableLibraries[track.LibraryID] {
			visible = append(visible, track)
		}
	}
	writeJSON(w, http.StatusOK, tracksResponse{Tracks: visible})
}

func (h trackFavoriteHandler) set(w http.ResponseWriter, r *http.Request, principal auth.Principal, favorite bool) {
	libraryID := r.PathValue("libraryID")
	trackID := r.PathValue("trackID")
	if !validFavoritePathID(libraryID) || !validFavoritePathID(trackID) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid favorite target"})
		return
	}
	if _, err := h.catalog.LibraryByID(r.Context(), libraryID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "library not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "library lookup failed"})
		return
	}
	if principal.Role != domain.RoleAdmin {
		membership, err := h.catalog.LibraryMembership(r.Context(), libraryID, principal.UserID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "library read denied"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "library authorization failed"})
			return
		}
		if !membership.CanRead {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "library read denied"})
			return
		}
	}
	track, err := h.favorites.TrackForLibrary(r.Context(), libraryID, trackID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "track not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "track lookup failed"})
		return
	}
	if track.ID != trackID || track.LibraryID != libraryID {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "track result invalid"})
		return
	}
	if err := h.favorites.SetTrackFavorite(r.Context(), principal.UserID, trackID, favorite); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "favorite update failed"})
		return
	}
	writeJSON(w, http.StatusOK, trackFavoriteResponse{TrackID: trackID, Favorite: favorite})
}

func authenticatedFavoritePrincipal(w http.ResponseWriter, r *http.Request, musicStore store.MusicStore) (auth.Principal, bool) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return auth.Principal{}, false
	}
	if musicStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "music store unavailable"})
		return auth.Principal{}, false
	}
	user, err := musicStore.UserByID(r.Context(), principal.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
			return auth.Principal{}, false
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "identity lookup failed"})
		return auth.Principal{}, false
	}
	if user.ID != principal.UserID || user.Role != principal.Role {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "principal mismatch"})
		return auth.Principal{}, false
	}
	return principal, true
}

func validFavoritePathID(value string) bool {
	return value != "" && strings.TrimSpace(value) == value
}
