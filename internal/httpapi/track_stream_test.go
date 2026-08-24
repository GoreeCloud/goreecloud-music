package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/store"
)

type streamTestStore struct {
	user       domain.User
	library    domain.Library
	membership domain.LibraryMembership
	descriptor domain.StreamDescriptor
}

func (s streamTestStore) UserByID(context.Context, string) (domain.User, error) { return s.user, nil }
func (s streamTestStore) LibrariesForUser(context.Context, string) ([]domain.Library, error) {
	return nil, nil
}
func (s streamTestStore) LibraryMembership(context.Context, string, string) (domain.LibraryMembership, error) {
	return s.membership, nil
}
func (s streamTestStore) LibraryByID(context.Context, string) (domain.Library, error) {
	return s.library, nil
}
func (s streamTestStore) ResolveTrackStream(context.Context, string, string) (domain.StreamDescriptor, error) {
	return s.descriptor, nil
}

func TestTrackStreamRequiresReadableMembership(t *testing.T) {
	backend := streamTestStore{
		user:       domain.User{ID: "user-1", DisplayName: "User", Role: domain.RoleUser},
		library:    domain.Library{ID: "lib-1", Name: "Library", RootPath: "/music", Visibility: domain.LibraryPrivate},
		membership: domain.LibraryMembership{LibraryID: "lib-1", UserID: "user-1", CanRead: false},
	}
	h := trackStreamHandler{store: backend, stream: backend}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/libraries/lib-1/tracks/track-1/stream", nil)
	req.SetPathValue("libraryID", "lib-1")
	req.SetPathValue("trackID", "track-1")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleUser}))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
	}
}

func TestTrackStreamServesByteRange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.mp3")
	if err := os.WriteFile(path, []byte("0123456789"), 0o600); err != nil {
		t.Fatal(err)
	}
	backend := streamTestStore{
		user:       domain.User{ID: "user-1", DisplayName: "User", Role: domain.RoleUser},
		library:    domain.Library{ID: "lib-1", Name: "Library", RootPath: "/music", Visibility: domain.LibraryPrivate},
		membership: domain.LibraryMembership{LibraryID: "lib-1", UserID: "user-1", CanRead: true},
		descriptor: domain.StreamDescriptor{ContentType: "audio/mpeg", ContentLength: 10, AcceptRanges: true, SourcePath: path},
	}
	h := trackStreamHandler{store: backend, stream: backend}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/libraries/lib-1/tracks/track-1/stream", nil)
	req.Header.Set("Range", "bytes=2-5")
	req.SetPathValue("libraryID", "lib-1")
	req.SetPathValue("trackID", "track-1")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleUser}))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusPartialContent)
	}
	if got := res.Header().Get("Content-Range"); got != "bytes 2-5/10" {
		t.Fatalf("content range = %q", got)
	}
	if got := res.Body.String(); got != "2345" {
		t.Fatalf("body = %q", got)
	}
}

var _ store.MusicStore = streamTestStore{}
var _ TrackStreamStore = streamTestStore{}
