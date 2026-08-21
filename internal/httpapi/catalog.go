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

type LibraryCatalogStore interface {
	LibraryByID(ctx context.Context, libraryID string) (domain.Library, error)
	LibraryMembership(ctx context.Context, libraryID, userID string) (domain.LibraryMembership, error)
	AlbumsForLibrary(ctx context.Context, libraryID string) ([]domain.Album, error)
	TracksForLibrary(ctx context.Context, libraryID string) ([]domain.Track, error)
}

type catalogHandler struct {
	store   store.MusicStore
	catalog LibraryCatalogStore
	kind    string
}

type albumsResponse struct {
	Albums []domain.Album `json:"albums"`
}

type tracksResponse struct {
	Tracks []domain.Track `json:"tracks"`
}

func (h catalogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	if h.store == nil || h.catalog == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "library catalog unavailable"})
		return
	}

	user, err := h.store.UserByID(r.Context(), principal.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "identity lookup failed"})
		return
	}
	if user.Role != principal.Role {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "principal mismatch"})
		return
	}

	libraryID := strings.TrimSpace(r.PathValue("libraryID"))
	if libraryID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "library id is required"})
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

	switch h.kind {
	case "albums":
		albums, err := h.catalog.AlbumsForLibrary(r.Context(), libraryID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "album lookup failed"})
			return
		}
		if albums == nil {
			albums = []domain.Album{}
		}
		writeJSON(w, http.StatusOK, albumsResponse{Albums: albums})
	case "tracks":
		tracks, err := h.catalog.TracksForLibrary(r.Context(), libraryID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "track lookup failed"})
			return
		}
		if tracks == nil {
			tracks = []domain.Track{}
		}
		writeJSON(w, http.StatusOK, tracksResponse{Tracks: tracks})
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "catalog resource not found"})
	}
}
