package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

type fakeMusicStore struct {
	user      domain.User
	libraries []domain.Library
}

func (f fakeMusicStore) UserByID(_ context.Context, userID string) (domain.User, error) {
	if f.user.ID == userID {
		return f.user, nil
	}
	return domain.User{}, errNotFoundForTest
}

func (f fakeMusicStore) LibrariesForUser(_ context.Context, _ string) ([]domain.Library, error) {
	return f.libraries, nil
}

func (f fakeMusicStore) LibraryMembership(_ context.Context, _, _ string) (domain.LibraryMembership, error) {
	return domain.LibraryMembership{}, errNotFoundForTest
}

type testNotFoundError struct{}

func (testNotFoundError) Error() string { return "not found" }

var errNotFoundForTest error = testNotFoundError{}

func TestLibrariesRequiresPrincipal(t *testing.T) {
	t.Parallel()

	router := NewRouterWithDependencies(Dependencies{Store: fakeMusicStore{}})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/me/libraries", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestLibrariesReturnsAuthorizedLibraries(t *testing.T) {
	t.Parallel()

	musicStore := fakeMusicStore{
		user: domain.User{ID: "user-1", DisplayName: "Listener", Role: domain.RoleUser},
		libraries: []domain.Library{
			{ID: "library-1", Name: "Shared Music", Visibility: domain.LibraryShared},
			{ID: "library-2", Name: "Private Music", Visibility: domain.LibraryPrivate},
		},
	}
	router := NewRouterWithDependencies(Dependencies{Store: musicStore})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/libraries", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleUser}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response librariesResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Libraries) != 2 {
		t.Fatalf("expected 2 libraries, got %d", len(response.Libraries))
	}
	if response.Libraries[0].Name != "Shared Music" {
		t.Fatalf("unexpected first library: %+v", response.Libraries[0])
	}
}
