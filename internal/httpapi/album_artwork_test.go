package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/artwork"
	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

type artworkTestStore struct {
	user       domain.User
	library    domain.Library
	membership domain.LibraryMembership
	source     artwork.Source
}

func (s artworkTestStore) UserByID(context.Context, string) (domain.User, error) { return s.user, nil }
func (s artworkTestStore) LibrariesForUser(context.Context, string) ([]domain.Library, error) { return nil, nil }
func (s artworkTestStore) LibraryMembership(context.Context, string, string) (domain.LibraryMembership, error) {
	return s.membership, nil
}
func (s artworkTestStore) LibraryByID(context.Context, string) (domain.Library, error) { return s.library, nil }
func (s artworkTestStore) AlbumArtwork(context.Context, string, string) (artwork.Source, error) { return s.source, nil }

func TestAlbumArtworkRequiresReadableMembership(t *testing.T) {
	backend := artworkTestStore{
		user: domain.User{ID: "user-1", DisplayName: "User", Role: domain.RoleUser},
		library: domain.Library{ID: "lib-1", Name: "Library", RootPath: "/music", Visibility: domain.LibraryPrivate},
		membership: domain.LibraryMembership{LibraryID: "lib-1", UserID: "user-1", CanRead: false},
	}
	h := albumArtworkHandler{store: backend, artwork: backend}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/libraries/lib-1/albums/album-1/artwork", nil)
	req.SetPathValue("libraryID", "lib-1")
	req.SetPathValue("albumID", "album-1")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleUser}))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
	}
}

func TestAlbumArtworkServesSidecar(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cover.jpg")
	if err := os.WriteFile(path, []byte("image-data"), 0o600); err != nil {
		t.Fatal(err)
	}
	backend := artworkTestStore{
		user: domain.User{ID: "user-1", DisplayName: "User", Role: domain.RoleUser},
		library: domain.Library{ID: "lib-1", Name: "Library", RootPath: "/music", Visibility: domain.LibraryPrivate},
		membership: domain.LibraryMembership{LibraryID: "lib-1", UserID: "user-1", CanRead: true},
		source: artwork.Source{Type: artwork.SourceSidecar, Path: path, MIMEType: "image/jpeg"},
	}
	h := albumArtworkHandler{store: backend, artwork: backend}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/libraries/lib-1/albums/album-1/artwork", nil)
	req.SetPathValue("libraryID", "lib-1")
	req.SetPathValue("albumID", "album-1")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleUser}))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := res.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("content type = %q", got)
	}
	if got := res.Body.String(); got != "image-data" {
		t.Fatalf("body = %q", got)
	}
}

var _ store.MusicStore = artworkTestStore{}
var _ AlbumArtworkStore = artworkTestStore{}
