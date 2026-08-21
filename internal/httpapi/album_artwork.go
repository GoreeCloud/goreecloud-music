package httpapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/GoreeCloud/goreecloud-music/internal/artwork"
	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

type AlbumArtworkStore interface {
	LibraryByID(ctx context.Context, libraryID string) (domain.Library, error)
	LibraryMembership(ctx context.Context, libraryID, userID string) (domain.LibraryMembership, error)
	AlbumArtwork(ctx context.Context, libraryID, albumID string) (artwork.Source, error)
}

type albumArtworkHandler struct {
	store   store.MusicStore
	artwork AlbumArtworkStore
}

func (h albumArtworkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	if h.store == nil || h.artwork == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "album artwork unavailable"})
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
	albumID := strings.TrimSpace(r.PathValue("albumID"))
	if libraryID == "" || albumID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "library id and album id are required"})
		return
	}
	if _, err := h.artwork.LibraryByID(r.Context(), libraryID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "library not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "library lookup failed"})
		return
	}
	if principal.Role != domain.RoleAdmin {
		membership, err := h.artwork.LibraryMembership(r.Context(), libraryID, principal.UserID)
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

	source, err := h.artwork.AlbumArtwork(r.Context(), libraryID, albumID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "album artwork not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "album artwork lookup failed"})
		return
	}

	w.Header().Set("Content-Type", source.MIMEType)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	switch source.Type {
	case artwork.SourceSidecar:
		file, err := os.Open(source.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				http.Error(w, "artwork source missing", http.StatusNotFound)
				return
			}
			http.Error(w, "artwork source unavailable", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		if _, err := io.Copy(w, file); err != nil {
			return
		}
	case artwork.SourceEmbedded:
		data, err := extractEmbeddedArtwork(r.Context(), source.Path)
		if err != nil {
			http.Error(w, "embedded artwork unavailable", http.StatusInternalServerError)
			return
		}
		_, _ = io.Copy(w, bytes.NewReader(data))
	default:
		http.Error(w, "unsupported artwork source", http.StatusInternalServerError)
	}
}

func extractEmbeddedArtwork(ctx context.Context, path string) ([]byte, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("embedded artwork path is required")
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", "-v", "error", "-i", path, "-map", "0:v:0", "-frames:v", "1", "-f", "image2pipe", "pipe:1")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg artwork extraction: %w", err)
	}
	if len(output) == 0 {
		return nil, errors.New("embedded artwork is empty")
	}
	return output, nil
}
