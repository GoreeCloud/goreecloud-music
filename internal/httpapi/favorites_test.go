package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

type fakeTrackFavoriteStore struct {
	track       domain.Track
	tracks      []domain.Track
	trackErr    error
	listErr     error
	setErr      error
	setCalls    int
	setUserID   string
	setTrackID  string
	setFavorite bool
	lookupCalls int
	listUserID  string
	listCalls   int
}

func (f *fakeTrackFavoriteStore) TrackForLibrary(_ context.Context, _, _ string) (domain.Track, error) {
	f.lookupCalls++
	return f.track, f.trackErr
}

func (f *fakeTrackFavoriteStore) TrackFavoritesForUser(_ context.Context, userID string) ([]domain.Track, error) {
	f.listCalls++
	f.listUserID = userID
	return append([]domain.Track(nil), f.tracks...), f.listErr
}

func (f *fakeTrackFavoriteStore) SetTrackFavorite(_ context.Context, userID, trackID string, favorite bool) error {
	f.setCalls++
	f.setUserID = userID
	f.setTrackID = trackID
	f.setFavorite = favorite
	return f.setErr
}

type favoriteCatalogStore struct {
	libraries   map[string]domain.Library
	memberships map[string]domain.LibraryMembership
	err         error
}

func (f favoriteCatalogStore) LibraryByID(_ context.Context, libraryID string) (domain.Library, error) {
	if f.err != nil {
		return domain.Library{}, f.err
	}
	library, ok := f.libraries[libraryID]
	if !ok {
		return domain.Library{}, store.ErrNotFound
	}
	return library, nil
}

func (f favoriteCatalogStore) LibraryMembership(_ context.Context, libraryID, userID string) (domain.LibraryMembership, error) {
	if f.err != nil {
		return domain.LibraryMembership{}, f.err
	}
	membership, ok := f.memberships[libraryID+"/"+userID]
	if !ok {
		return domain.LibraryMembership{}, store.ErrNotFound
	}
	return membership, nil
}

func (favoriteCatalogStore) AlbumsForLibrary(context.Context, string) ([]domain.Album, error) {
	return nil, nil
}

func (favoriteCatalogStore) TracksForLibrary(context.Context, string) ([]domain.Track, error) {
	return nil, nil
}

func TestTrackFavoriteRequiresMatchingAuthenticatedPrincipal(t *testing.T) {
	favorites := &fakeTrackFavoriteStore{}
	catalog := readableFavoriteCatalog()
	router := NewRouterWithDependencies(Dependencies{
		Store:     fakeMusicStore{user: domain.User{ID: "user-1", Role: domain.RoleUser}},
		Catalog:   catalog,
		Favorites: favorites,
	})

	unauthenticated := httptest.NewRecorder()
	router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/api/v1/me/favorites/tracks", nil))
	if unauthenticated.Code != http.StatusUnauthorized || favorites.listCalls != 0 {
		t.Fatalf("status=%d calls=%d", unauthenticated.Code, favorites.listCalls)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/favorites/tracks", nil)
	request = request.WithContext(auth.WithPrincipal(request.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleAdmin}))
	mismatch := httptest.NewRecorder()
	router.ServeHTTP(mismatch, request)
	if mismatch.Code != http.StatusUnauthorized || favorites.listCalls != 0 {
		t.Fatalf("status=%d calls=%d", mismatch.Code, favorites.listCalls)
	}
}

func TestTrackFavoriteSetAndUnsetUsePrincipalUserAndReadableTrack(t *testing.T) {
	favorites := &fakeTrackFavoriteStore{track: domain.Track{ID: "track-1", LibraryID: "library-1", Title: "Song"}}
	router := favoriteTestRouter(favorites, readableFavoriteCatalog(), domain.RoleUser)

	setRecorder := httptest.NewRecorder()
	router.ServeHTTP(setRecorder, favoriteRequest(http.MethodPut, "/api/v1/libraries/library-1/tracks/track-1/favorite", domain.RoleUser))
	if setRecorder.Code != http.StatusOK || favorites.setCalls != 1 || favorites.setUserID != "user-1" || favorites.setTrackID != "track-1" || !favorites.setFavorite {
		t.Fatalf("status=%d calls=%d user=%q track=%q favorite=%v", setRecorder.Code, favorites.setCalls, favorites.setUserID, favorites.setTrackID, favorites.setFavorite)
	}

	unsetRecorder := httptest.NewRecorder()
	router.ServeHTTP(unsetRecorder, favoriteRequest(http.MethodDelete, "/api/v1/libraries/library-1/tracks/track-1/favorite", domain.RoleUser))
	if unsetRecorder.Code != http.StatusOK || favorites.setCalls != 2 || favorites.setFavorite {
		t.Fatalf("status=%d calls=%d favorite=%v", unsetRecorder.Code, favorites.setCalls, favorites.setFavorite)
	}
}

func TestTrackFavoriteMutationDeniesUnreadableLibraryBeforeTrackLookup(t *testing.T) {
	favorites := &fakeTrackFavoriteStore{track: domain.Track{ID: "track-1", LibraryID: "library-1"}}
	catalog := readableFavoriteCatalog()
	catalog.memberships["library-1/user-1"] = domain.LibraryMembership{LibraryID: "library-1", UserID: "user-1", CanRead: false}
	router := favoriteTestRouter(favorites, catalog, domain.RoleUser)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, favoriteRequest(http.MethodPut, "/api/v1/libraries/library-1/tracks/track-1/favorite", domain.RoleUser))
	if recorder.Code != http.StatusForbidden || favorites.lookupCalls != 0 || favorites.setCalls != 0 {
		t.Fatalf("status=%d lookups=%d sets=%d", recorder.Code, favorites.lookupCalls, favorites.setCalls)
	}
}

func TestTrackFavoriteMutationRejectsWrongTrackIdentity(t *testing.T) {
	favorites := &fakeTrackFavoriteStore{track: domain.Track{ID: "track-other", LibraryID: "library-1"}}
	router := favoriteTestRouter(favorites, readableFavoriteCatalog(), domain.RoleUser)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, favoriteRequest(http.MethodPut, "/api/v1/libraries/library-1/tracks/track-1/favorite", domain.RoleUser))
	if recorder.Code != http.StatusInternalServerError || favorites.setCalls != 0 {
		t.Fatalf("status=%d sets=%d", recorder.Code, favorites.setCalls)
	}
}

func TestTrackFavoriteListFiltersRevokedLibrariesForUser(t *testing.T) {
	favorites := &fakeTrackFavoriteStore{tracks: []domain.Track{
		{ID: "track-readable", LibraryID: "library-1", Title: "Readable"},
		{ID: "track-revoked", LibraryID: "library-2", Title: "Revoked"},
	}}
	catalog := favoriteCatalogStore{
		libraries: map[string]domain.Library{
			"library-1": {ID: "library-1"},
			"library-2": {ID: "library-2"},
		},
		memberships: map[string]domain.LibraryMembership{
			"library-1/user-1": {LibraryID: "library-1", UserID: "user-1", CanRead: true},
		},
	}
	router := favoriteTestRouter(favorites, catalog, domain.RoleUser)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, favoriteRequest(http.MethodGet, "/api/v1/me/favorites/tracks", domain.RoleUser))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response tracksResponse
	if err := decodeJSONBody(recorder, &response); err != nil {
		t.Fatal(err)
	}
	if favorites.listUserID != "user-1" || len(response.Tracks) != 1 || response.Tracks[0].ID != "track-readable" {
		t.Fatalf("user=%q tracks=%+v", favorites.listUserID, response.Tracks)
	}
}

func TestTrackFavoriteAdminListsOwnFavoritesWithoutMembershipFiltering(t *testing.T) {
	favorites := &fakeTrackFavoriteStore{tracks: []domain.Track{{ID: "track-1", LibraryID: "library-1"}}}
	catalog := favoriteCatalogStore{libraries: map[string]domain.Library{}, memberships: map[string]domain.LibraryMembership{}}
	router := favoriteTestRouter(favorites, catalog, domain.RoleAdmin)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, favoriteRequest(http.MethodGet, "/api/v1/me/favorites/tracks", domain.RoleAdmin))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response tracksResponse
	if err := decodeJSONBody(recorder, &response); err != nil || len(response.Tracks) != 1 {
		t.Fatalf("tracks=%+v err=%v", response.Tracks, err)
	}
}

func TestTrackFavoriteMapsMissingTrackAndStoreFailures(t *testing.T) {
	favorites := &fakeTrackFavoriteStore{trackErr: store.ErrNotFound}
	router := favoriteTestRouter(favorites, readableFavoriteCatalog(), domain.RoleUser)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, favoriteRequest(http.MethodPut, "/api/v1/libraries/library-1/tracks/track-1/favorite", domain.RoleUser))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status=%d", recorder.Code)
	}

	favorites.trackErr = nil
	favorites.track = domain.Track{ID: "track-1", LibraryID: "library-1"}
	favorites.setErr = errors.New("write failed")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, favoriteRequest(http.MethodPut, "/api/v1/libraries/library-1/tracks/track-1/favorite", domain.RoleUser))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", recorder.Code)
	}
}

func readableFavoriteCatalog() favoriteCatalogStore {
	return favoriteCatalogStore{
		libraries: map[string]domain.Library{
			"library-1": {ID: "library-1", Name: "Library", RootPath: "/music", Visibility: domain.LibraryPrivate},
		},
		memberships: map[string]domain.LibraryMembership{
			"library-1/user-1": {LibraryID: "library-1", UserID: "user-1", CanRead: true},
		},
	}
}

func favoriteTestRouter(favorites TrackFavoriteStore, catalog LibraryCatalogStore, role domain.Role) http.Handler {
	return NewRouterWithDependencies(Dependencies{
		Store:     fakeMusicStore{user: domain.User{ID: "user-1", Role: role}},
		Catalog:   catalog,
		Favorites: favorites,
	})
}

func favoriteRequest(method, path string, role domain.Role) *http.Request {
	request := httptest.NewRequest(method, path, nil)
	return request.WithContext(auth.WithPrincipal(request.Context(), auth.Principal{UserID: "user-1", Role: role}))
}

func decodeJSONBody(recorder *httptest.ResponseRecorder, destination any) error {
	return json.NewDecoder(recorder.Body).Decode(destination)
}
