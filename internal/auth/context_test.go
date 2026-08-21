package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

func TestDevelopmentIdentityDisabledPassesThrough(t *testing.T) {
	t.Parallel()

	handler := DevelopmentIdentity{Enabled: false}.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := PrincipalFromContext(r.Context()); ok {
			t.Fatal("did not expect principal when development identity is disabled")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.Code)
	}
}

func TestDevelopmentIdentityRequiresUser(t *testing.T) {
	t.Parallel()

	handler := DevelopmentIdentity{Enabled: true}.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestDevelopmentIdentityCreatesPrincipal(t *testing.T) {
	t.Parallel()

	handler := DevelopmentIdentity{Enabled: true}.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("expected principal")
		}
		if principal.UserID != "user-1" || principal.Role != domain.RoleAdmin {
			t.Fatalf("unexpected principal: %#v", principal)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-GoreeCloud-User-ID", "user-1")
	req.Header.Set("X-GoreeCloud-Role", "admin")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.Code)
	}
}
