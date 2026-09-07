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
	"github.com/GoreeCloud/goreecloud-music/internal/playlists"
)

type fakePlaylistDeleteService struct {
	*fakePlaylistService
	deleted     playlists.Record
	deleteErr   error
	deleteCalls int
	deleteUser  string
	deleteID    string
}

func (f *fakePlaylistDeleteService) Delete(_ context.Context, userID, playlistID string) (playlists.Record, error) {
	f.deleteCalls++
	f.deleteUser = userID
	f.deleteID = playlistID
	return f.deleted, f.deleteErr
}

func TestPlaylistDeleteRequiresMatchingPrincipal(t *testing.T) {
	service := &fakePlaylistDeleteService{fakePlaylistService: &fakePlaylistService{}}
	router := NewRouterWithDependencies(Dependencies{
		Store:     fakeMusicStore{user: domain.User{ID: "user-1", Role: domain.RoleUser}},
		Playlists: service,
	})

	unauthenticated := httptest.NewRecorder()
	router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodDelete, "/api/v1/me/playlists/playlist-1", nil))
	if unauthenticated.Code != http.StatusUnauthorized || service.deleteCalls != 0 {
		t.Fatalf("status=%d calls=%d", unauthenticated.Code, service.deleteCalls)
	}

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/me/playlists/playlist-1", nil)
	request = request.WithContext(auth.WithPrincipal(request.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleAdmin}))
	mismatch := httptest.NewRecorder()
	router.ServeHTTP(mismatch, request)
	if mismatch.Code != http.StatusUnauthorized || service.deleteCalls != 0 {
		t.Fatalf("status=%d calls=%d", mismatch.Code, service.deleteCalls)
	}
}

func TestPlaylistDeleteReturnsCanonicalDeletedRecord(t *testing.T) {
	playlist, _ := playlists.New("user-1", "playlist-1", "Mix")
	record, _ := playlists.NewRecord(playlist)
	service := &fakePlaylistDeleteService{
		fakePlaylistService: &fakePlaylistService{},
		deleted:             record,
	}
	router := playlistDeleteTestRouter(service)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, authenticatedPlaylistRequest(http.MethodDelete, "/api/v1/me/playlists/playlist-1", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		ID       string `json:"id"`
		Revision string `json:"revision"`
		Deleted  bool   `json:"deleted"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.ID != "playlist-1" || response.Revision != "0" || !response.Deleted || service.deleteUser != "user-1" || service.deleteID != "playlist-1" {
		t.Fatalf("response=%+v user=%q id=%q", response, service.deleteUser, service.deleteID)
	}
}

func TestPlaylistDeleteMapsNotFoundAndConcurrentChange(t *testing.T) {
	service := &fakePlaylistDeleteService{fakePlaylistService: &fakePlaylistService{}}
	router := playlistDeleteTestRouter(service)
	for _, test := range []struct {
		err    error
		status int
	}{
		{playlists.ErrRecordNotFound, http.StatusNotFound},
		{playlists.ErrStaleRevision, http.StatusConflict},
		{errors.New("database failed"), http.StatusInternalServerError},
	} {
		service.deleteErr = test.err
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, authenticatedPlaylistRequest(http.MethodDelete, "/api/v1/me/playlists/playlist-1", nil))
		if recorder.Code != test.status {
			t.Fatalf("error=%v status=%d want=%d body=%s", test.err, recorder.Code, test.status, recorder.Body.String())
		}
	}
}

func TestPlaylistDeleteRejectsWrongReturnedIdentity(t *testing.T) {
	wrong, _ := playlists.New("user-2", "playlist-2", "Wrong")
	wrongRecord, _ := playlists.NewRecord(wrong)
	service := &fakePlaylistDeleteService{
		fakePlaylistService: &fakePlaylistService{},
		deleted:             wrongRecord,
	}
	router := playlistDeleteTestRouter(service)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, authenticatedPlaylistRequest(http.MethodDelete, "/api/v1/me/playlists/playlist-1", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func playlistDeleteTestRouter(service *fakePlaylistDeleteService) http.Handler {
	return NewRouterWithDependencies(Dependencies{
		Store:     fakeMusicStore{user: domain.User{ID: "user-1", Role: domain.RoleUser}},
		Playlists: service,
	})
}
