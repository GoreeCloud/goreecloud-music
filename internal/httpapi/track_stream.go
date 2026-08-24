package httpapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
	"github.com/GoreeCloud/goreecloud-music/internal/streaming"
)

type TrackStreamStore interface {
	LibraryByID(ctx context.Context, libraryID string) (domain.Library, error)
	LibraryMembership(ctx context.Context, libraryID, userID string) (domain.LibraryMembership, error)
	ResolveTrackStream(ctx context.Context, libraryID, trackID string) (domain.StreamDescriptor, error)
}

type trackStreamHandler struct {
	store  store.MusicStore
	stream TrackStreamStore
}

func (h trackStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	if h.store == nil || h.stream == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "track streaming unavailable"})
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
	trackID := strings.TrimSpace(r.PathValue("trackID"))
	if libraryID == "" || trackID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "library id and track id are required"})
		return
	}
	if _, err := h.stream.LibraryByID(r.Context(), libraryID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "library not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "library lookup failed"})
		return
	}
	if principal.Role != domain.RoleAdmin {
		membership, err := h.stream.LibraryMembership(r.Context(), libraryID, principal.UserID)
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

	descriptor, err := h.stream.ResolveTrackStream(r.Context(), libraryID, trackID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "track not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "track stream lookup failed"})
		return
	}

	file, err := os.Open(descriptor.SourcePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "track source missing", http.StatusNotFound)
			return
		}
		http.Error(w, "track source unavailable", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.Error(w, "track source unavailable", http.StatusInternalServerError)
		return
	}
	size := info.Size()
	byteRange, err := streaming.ParseSingleRange(r.Header.Get("Range"), size)
	if err != nil {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		http.Error(w, "range not satisfiable", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	w.Header().Set("Content-Type", descriptor.ContentType)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.FormatInt(byteRange.Length(), 10))
	status := http.StatusOK
	if r.Header.Get("Range") != "" {
		status = http.StatusPartialContent
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", byteRange.Start, byteRange.End, size))
	}
	w.WriteHeader(status)
	_, _ = io.CopyN(w, io.NewSectionReader(file, byteRange.Start, byteRange.Length()), byteRange.Length())
}
