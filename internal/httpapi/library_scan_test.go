package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/auth"
	"github.com/GoreeCloud/goreecloud-music/internal/domain"
	"github.com/GoreeCloud/goreecloud-music/internal/ingest"
)

type fakeLibraryScanStore struct {
	library    domain.Library
	membership domain.LibraryMembership
}

func (f fakeLibraryScanStore) LibraryByID(_ context.Context, _ string) (domain.Library, error) {
	return f.library, nil
}

func (f fakeLibraryScanStore) LibraryMembership(_ context.Context, _, _ string) (domain.LibraryMembership, error) {
	return f.membership, nil
}

type fakeLibraryScanner struct {
	called  bool
	summary ingest.Summary
}

func (f *fakeLibraryScanner) ScanLibrary(_ context.Context, _ domain.Library) (ingest.Summary, error) {
	f.called = true
	return f.summary, nil
}

func TestLibraryScanRequiresManagePermission(t *testing.T) {
	musicStore := fakeMusicStore{
		user: domain.User{ID: "user-1", DisplayName: "Listener", Role: domain.RoleUser},
	}
	scanStore := fakeLibraryScanStore{
		library: domain.Library{ID: "library-1", Name: "Music", RootPath: "/music", Visibility: domain.LibraryShared},
		membership: domain.LibraryMembership{
			LibraryID: "library-1",
			UserID:    "user-1",
			CanRead:   true,
			CanManage: false,
		},
	}
	scanner := &fakeLibraryScanner{}
	router := NewRouterWithDependencies(Dependencies{Store: musicStore, ScanStore: scanStore, Scanner: scanner})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/libraries/library-1/scan", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleUser}))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", res.Code, res.Body.String())
	}
	if scanner.called {
		t.Fatal("scanner must not run without manage permission")
	}
}

func TestLibraryScanRunsForManager(t *testing.T) {
	musicStore := fakeMusicStore{
		user: domain.User{ID: "user-1", DisplayName: "Manager", Role: domain.RoleUser},
	}
	scanStore := fakeLibraryScanStore{
		library: domain.Library{ID: "library-1", Name: "Music", RootPath: "/music", Visibility: domain.LibraryShared},
		membership: domain.LibraryMembership{
			LibraryID: "library-1",
			UserID:    "user-1",
			CanRead:   true,
			CanManage: true,
		},
	}
	scanner := &fakeLibraryScanner{summary: ingest.Summary{Discovered: 10, Indexed: 10, Removed: 1}}
	router := NewRouterWithDependencies(Dependencies{Store: musicStore, ScanStore: scanStore, Scanner: scanner})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/libraries/library-1/scan", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: "user-1", Role: domain.RoleUser}))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if !scanner.called {
		t.Fatal("expected scanner to run")
	}
}
