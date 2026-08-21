package httpapi

import (
	"errors"
	"net/http"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

type libraryResponse struct {
	ID         string                   `json:"id"`
	Name       string                   `json:"name"`
	Visibility domain.LibraryVisibility `json:"visibility"`
}

type librariesResponse struct {
	Libraries []libraryResponse `json:"libraries"`
}

type librariesHandler struct {
	store store.MusicStore
}

func (h librariesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "music store unavailable"})
		return
	}

	user, err := h.store.UserByID(r.Context(), principal.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "library lookup failed"})
		return
	}
	if user.Role != principal.Role {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "principal mismatch"})
		return
	}

	libraries, err := h.store.LibrariesForUser(r.Context(), principal.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "library lookup failed"})
		return
	}

	response := librariesResponse{Libraries: make([]libraryResponse, 0, len(libraries))}
	for _, library := range libraries {
		response.Libraries = append(response.Libraries, libraryResponse{
			ID:         library.ID,
			Name:       library.Name,
			Visibility: library.Visibility,
		})
	}
	writeJSON(w, http.StatusOK, response)
}
