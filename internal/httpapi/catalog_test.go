package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

type fakeCatalogStore struct {
	library    domain.Library
	membership domain.LibraryMembership
	albums     []domain.Album
	tracks     []domain.Track
}

func (f fakeCatalogStore) LibraryByID(_ context.Context, libraryID string) (domain.Library, error) {
	if f.library.ID == libraryID {
		return f.library, nil
	}
	return domain.Library{}, store.ErrNotFound
}

func (f fakeCatalogStore) LibraryMembership(_ context.Context, libraryID, userID string) (domain.LibraryMembership, error) {
	if f.membership.LibraryID == libraryID && f.membership.UserID == userID {
		return f.membership, nil
	}
	return domain.LibraryMembership{}, store.ErrNotFound
}

func (f fakeCatalogStore) AlbumsForLibrary(_ context.Context, _ string) ([]domain.Album, error) {
	return f.albums, nil
}

func (f fakeCatalogStore) TracksForLibrary(_ context.Context, _ string) ([]domain.Track, error) {
	return f.tracks, nil
}

func TestCatalogRequiresReadableMembership(t *testing.T) {
	t.Parallel()

	musicStore := fakeMusicStore{user: domain.User{ID: "user-1", DisplayName: "Listener", Role: domain.RoleUser}}
	catalog := fakeCatalogStore{
		library:    domain.Library{ID: "library-1", Name: "Shared", RootPath: "/music", Visibility: domain.LibraryShared},
		membership: domain.LibraryMembership{LibraryID: "library-1", UserID: "user-1", CanRead: false},
	}
	router := NewRouterWithDependencies(Dependencies{Store: musicStore, Catalog: catalog})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/libraries/library-1/albums", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleUser}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestCatalogReturnsReadableAlbums(t *testing.T) {
	t.Parallel()

	musicStore := fakeMusicStore{user: domain.User{ID: "user-1", DisplayName: "Listener", Role: domain.RoleUser}}
	catalog := fakeCatalogStore{
		library:    domain.Library{ID: "library-1", Name: "Shared", RootPath: "/music", Visibility: domain.LibraryShared},
		membership: domain.LibraryMembership{LibraryID: "library-1", UserID: "user-1", CanRead: true},
		albums:     []domain.Album{{ID: "album-1", LibraryID: "library-1", Title: "Album"}},
	}
	router := NewRouterWithDependencies(Dependencies{Store: musicStore, Catalog: catalog})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/libraries/library-1/albums", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleUser}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
