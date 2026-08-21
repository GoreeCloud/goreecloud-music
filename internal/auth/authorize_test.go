package auth

import (
	"testing"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

func TestLibraryAuthorization(t *testing.T) {
	t.Parallel()

	membership := domain.LibraryMembership{LibraryID: "library-1", UserID: "user-1", CanRead: true, CanManage: false}
	user := Principal{UserID: "user-1", Role: domain.RoleUser}
	other := Principal{UserID: "user-2", Role: domain.RoleUser}
	admin := Principal{UserID: "admin-1", Role: domain.RoleAdmin}

	if !CanReadLibrary(user, membership) {
		t.Fatal("expected member to read library")
	}
	if CanManageLibrary(user, membership) {
		t.Fatal("did not expect read-only member to manage library")
	}
	if CanReadLibrary(other, membership) {
		t.Fatal("did not expect another user to inherit membership")
	}
	if !CanManageLibrary(admin, membership) {
		t.Fatal("expected administrator to manage library")
	}
}
