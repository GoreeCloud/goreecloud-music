package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/playlists"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

type fakePlaylistService struct {
	listEntries []playlists.CatalogEntry
	listErr     error
	record      playlists.Record
	playlist    playlists.Playlist
	err         error
	removed     string
	calls       []string
	userIDs     []string
	playlistIDs []string
}

func (f *fakePlaylistService) remember(operation, userID, playlistID string) {
	f.calls = append(f.calls, operation)
	f.userIDs = append(f.userIDs, userID)
	f.playlistIDs = append(f.playlistIDs, playlistID)
}

func (f *fakePlaylistService) List(_ context.Context, userID string) ([]playlists.CatalogEntry, error) {
	f.remember("list", userID, "")
	return append([]playlists.CatalogEntry(nil), f.listEntries...), f.listErr
}

func (f *fakePlaylistService) Create(_ context.Context, userID, playlistID, _ string) (playlists.Record, error) {
	f.remember("create", userID, playlistID)
	return f.record, f.err
}

func (f *fakePlaylistService) Load(_ context.Context, userID, playlistID string) (playlists.Record, playlists.Playlist, error) {
	f.remember("load", userID, playlistID)
	return f.record, f.playlist, f.err
}

func (f *fakePlaylistService) Rename(_ context.Context, userID, playlistID, _ string) (playlists.Record, error) {
	f.remember("rename", userID, playlistID)
	return f.record, f.err
}

func (f *fakePlaylistService) Append(_ context.Context, userID, playlistID, _ string) (playlists.Record, error) {
	f.remember("append", userID, playlistID)
	return f.record, f.err
}

func (f *fakePlaylistService) Insert(_ context.Context, userID, playlistID string, _ int, _ string) (playlists.Record, error) {
	f.remember("insert", userID, playlistID)
	return f.record, f.err
}

func (f *fakePlaylistService) RemoveAt(_ context.Context, userID, playlistID string, _ int) (playlists.Record, string, error) {
	f.remember("remove", userID, playlistID)
	return f.record, f.removed, f.err
}

func (f *fakePlaylistService) Move(_ context.Context, userID, playlistID string, _, _ int) (playlists.Record, error) {
	f.remember("move", userID, playlistID)
	return f.record, f.err
}

func TestPlaylistAPIRequiresExistingMatchingPrincipal(t *testing.T) {
	service := &fakePlaylistService{}
	musicStore := fakeMusicStore{user: domain.User{ID: "user-1", Role: domain.RoleUser}}
	router := NewRouterWithDependencies(Dependencies{Store: musicStore, Playlists: service})

	unauthenticated := httptest.NewRecorder()
	router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/api/v1/me/playlists", nil))
	if unauthenticated.Code != http.StatusUnauthorized || len(service.calls) != 0 {
		t.Fatalf("status=%d calls=%v", unauthenticated.Code, service.calls)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/playlists", nil)
	request = request.WithContext(auth.WithPrincipal(request.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleAdmin}))
	mismatch := httptest.NewRecorder()
	router.ServeHTTP(mismatch, request)
	if mismatch.Code != http.StatusUnauthorized || len(service.calls) != 0 {
		t.Fatalf("status=%d calls=%v", mismatch.Code, service.calls)
	}
}

func TestPlaylistAPIListsOnlyPrincipalUserEntries(t *testing.T) {
	service := &fakePlaylistService{listEntries: []playlists.CatalogEntry{
		{ID: "a", Name: "Alpha", TrackCount: 1},
		{ID: "b", Name: "Beta", TrackCount: 2},
	}}
	router := playlistTestRouter(service)
	request := authenticatedPlaylistRequest(http.MethodGet, "/api/v1/me/playlists", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response playlistListResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if len(response.Playlists) != 2 || response.Playlists[0].ID != "a" || response.Playlists[1].TrackCount != 2 {
		t.Fatalf("response=%+v", response)
	}
	if len(service.userIDs) != 1 || service.userIDs[0] != "user-1" {
		t.Fatalf("users=%v", service.userIDs)
	}
}

func TestPlaylistAPICreateAndGetReturnCanonicalRevisionStrings(t *testing.T) {
	playlist, _ := playlists.New("user-1", "playlist-1", "Road Trip")
	_ = playlist.Append("track-1")
	record, _ := playlists.NewRecord(playlist)
	service := &fakePlaylistService{record: record, playlist: playlist}
	router := playlistTestRouter(service)

	create := authenticatedPlaylistRequest(http.MethodPost, "/api/v1/me/playlists", strings.NewReader(`{"id":"playlist-1","name":"Road Trip"}`))
	createRecorder := httptest.NewRecorder()
	router.ServeHTTP(createRecorder, create)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRecorder.Code, createRecorder.Body.String())
	}
	var created playlistResponse
	if err := json.NewDecoder(createRecorder.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID != "playlist-1" || created.Revision != "0" || len(created.TrackIDs) != 1 {
		t.Fatalf("created=%+v", created)
	}

	get := authenticatedPlaylistRequest(http.MethodGet, "/api/v1/me/playlists/playlist-1", nil)
	getRecorder := httptest.NewRecorder()
	router.ServeHTTP(getRecorder, get)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", getRecorder.Code, getRecorder.Body.String())
	}
	if service.playlistIDs[len(service.playlistIDs)-1] != "playlist-1" {
		t.Fatalf("playlist IDs=%v", service.playlistIDs)
	}
}

func TestPlaylistAPIRejectsUnknownTrailingAndOversizedJSON(t *testing.T) {
	service := &fakePlaylistService{}
	router := playlistTestRouter(service)
	for name, body := range map[string]string{
		"unknown":   `{"id":"playlist-1","name":"Mix","extra":true}`,
		"trailing":  `{"id":"playlist-1","name":"Mix"}{}`,
		"oversized": `{"id":"playlist-1","name":"` + strings.Repeat("x", maxPlaylistRequestBodyBytes) + `"}`,
	} {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, authenticatedPlaylistRequest(http.MethodPost, "/api/v1/me/playlists", strings.NewReader(body)))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
	if len(service.calls) != 0 {
		t.Fatalf("unexpected service calls=%v", service.calls)
	}
}

func TestPlaylistAPIMutationRoutesAndRemoveResponse(t *testing.T) {
	playlist, _ := playlists.New("user-1", "playlist-1", "Mix")
	_ = playlist.Append("track-b")
	record, _ := playlists.NewRecord(playlist)
	service := &fakePlaylistService{record: record, playlist: playlist, removed: "track-a"}
	router := playlistTestRouter(service)

	cases := []struct {
		method string
		path   string
		body   string
		call   string
	}{
		{http.MethodPatch, "/api/v1/me/playlists/playlist-1", `{"name":"New"}`, "rename"},
		{http.MethodPost, "/api/v1/me/playlists/playlist-1/tracks", `{"trackId":"track-a"}`, "append"},
		{http.MethodPost, "/api/v1/me/playlists/playlist-1/tracks/insert", `{"index":0,"trackId":"track-a"}`, "insert"},
		{http.MethodPost, "/api/v1/me/playlists/playlist-1/tracks/move", `{"from":0,"to":1}`, "move"},
	}
	for _, test := range cases {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, authenticatedPlaylistRequest(test.method, test.path, strings.NewReader(test.body)))
		if recorder.Code != http.StatusOK || service.calls[len(service.calls)-1] != test.call {
			t.Fatalf("%s %s status=%d calls=%v body=%s", test.method, test.path, recorder.Code, service.calls, recorder.Body.String())
		}
	}

	removeRecorder := httptest.NewRecorder()
	router.ServeHTTP(removeRecorder, authenticatedPlaylistRequest(http.MethodDelete, "/api/v1/me/playlists/playlist-1/tracks/0", nil))
	if removeRecorder.Code != http.StatusOK || service.calls[len(service.calls)-1] != "remove" {
		t.Fatalf("status=%d calls=%v body=%s", removeRecorder.Code, service.calls, removeRecorder.Body.String())
	}
	var removed struct {
		RemovedTrackID string `json:"removedTrackId"`
	}
	if err := json.NewDecoder(removeRecorder.Body).Decode(&removed); err != nil || removed.RemovedTrackID != "track-a" {
		t.Fatalf("removed=%+v err=%v", removed, err)
	}
}

func TestPlaylistAPIMapsDomainAndConcurrencyErrors(t *testing.T) {
	playlist, _ := playlists.New("user-1", "playlist-1", "Mix")
	record, _ := playlists.NewRecord(playlist)
	service := &fakePlaylistService{record: record, playlist: playlist}
	router := playlistTestRouter(service)

	cases := []struct {
		err    error
		status int
	}{
		{playlists.ErrRecordNotFound, http.StatusNotFound},
		{store.ErrPlaylistAlreadyExists, http.StatusConflict},
		{playlists.ErrStaleRevision, http.StatusConflict},
		{playlists.ErrInvalidPlaylistState, http.StatusBadRequest},
		{errors.New("database failed"), http.StatusInternalServerError},
	}
	for _, test := range cases {
		service.err = test.err
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, authenticatedPlaylistRequest(http.MethodGet, "/api/v1/me/playlists/playlist-1", nil))
		if recorder.Code != test.status {
			t.Fatalf("error=%v status=%d want=%d body=%s", test.err, recorder.Code, test.status, recorder.Body.String())
		}
	}
}

func TestPlaylistAPIRemoveRejectsInvalidIndexBeforeService(t *testing.T) {
	service := &fakePlaylistService{}
	router := playlistTestRouter(service)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, authenticatedPlaylistRequest(http.MethodDelete, "/api/v1/me/playlists/playlist-1/tracks/not-a-number", nil))
	if recorder.Code != http.StatusBadRequest || len(service.calls) != 0 {
		t.Fatalf("status=%d calls=%v", recorder.Code, service.calls)
	}
}

func playlistTestRouter(service PlaylistService) http.Handler {
	return NewRouterWithDependencies(Dependencies{
		Store:     fakeMusicStore{user: domain.User{ID: "user-1", Role: domain.RoleUser}},
		Playlists: service,
	})
}

func authenticatedPlaylistRequest(method, path string, body io.Reader) *http.Request {
	request := httptest.NewRequest(method, path, body)
	return request.WithContext(auth.WithPrincipal(request.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleUser}))
}
