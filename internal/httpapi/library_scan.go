package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/ingest"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

type LibraryScanStore interface {
	LibraryByID(ctx context.Context, libraryID string) (domain.Library, error)
	LibraryMembership(ctx context.Context, libraryID, userID string) (domain.LibraryMembership, error)
}

type LibraryScanner interface {
	ScanLibrary(ctx context.Context, library domain.Library) (ingest.Summary, error)
}

type libraryScanHandler struct {
	store   store.MusicStore
	scan    LibraryScanStore
	scanner LibraryScanner
}

type libraryScanResponse struct {
	Status  string         `json:"status"`
	Library string         `json:"libraryId"`
	Summary ingest.Summary `json:"summary"`
}

func (h libraryScanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	if h.store == nil || h.scan == nil || h.scanner == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "library scanner unavailable"})
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
	library, err := h.scan.LibraryByID(r.Context(), libraryID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "library not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "library lookup failed"})
		return
	}

	if principal.Role != domain.RoleAdmin {
		membership, err := h.scan.LibraryMembership(r.Context(), libraryID, principal.UserID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "library management denied"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "library authorization failed"})
			return
		}
		if !auth.CanManageLibrary(principal, membership) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "library management denied"})
			return
		}
	}

	summary, err := h.scanner.ScanLibrary(r.Context(), library)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, libraryScanResponse{
			Status:  "incomplete",
			Library: library.ID,
			Summary: summary,
		})
		return
	}
	writeJSON(w, http.StatusOK, libraryScanResponse{
		Status:  "complete",
		Library: library.ID,
		Summary: summary,
	})
}
